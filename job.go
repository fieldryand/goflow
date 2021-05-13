package goflow

import (
	"fmt"
	"log"
	"time"
)

// A Job is a workflow consisting of independent and dependent tasks
// organized into a graph.
type Job struct {
	Name     string
	Tasks    map[string]*Task
	Dag      Dag
	Params   JobParams
	jobState *jobState
}

// NewJob returns a new job.
func NewJob(name string, p JobParams) *Job {
	j := Job{
		Name:     name,
		Dag:      make(Dag),
		Tasks:    make(map[string]*Task),
		Params:   p,
		jobState: newJobState()}
	return &j
}

// JobParams define optional job parameters.
type JobParams struct {
}

// Jobs and tasks are stateful.
type state string

const (
	none       state = "None"
	running    state = "Running"
	upForRetry state = "UpForRetry"
	skipped    state = "Skipped"
	failed     state = "Failed"
	successful state = "Successful"
)

type jobState struct {
	State     state            `json:"state"`
	TaskState map[string]state `json:"taskState"`
}

type triggerRule string

const (
	allDone       triggerRule = "allDone"
	allSuccessful triggerRule = "allSuccessful"
)

func newJobState() *jobState {
	js := jobState{none, make(map[string]state)}
	return &js
}

type writeOp struct {
	key  string
	val  state
	resp chan bool
}

type readOp struct {
	resp chan *jobState
}

// TaskParams define optional task parameters.
type TaskParams struct {
	TriggerRule triggerRule
	Retries     int
}

// AddTask adds a task to a job.
func (j *Job) AddTask(name string, op Operator, p TaskParams) *Job {
	if !(p.TriggerRule == allDone || p.TriggerRule == allSuccessful) {
		p.TriggerRule = allSuccessful
	}

	t := &Task{
		Name:              name,
		Operator:          op,
		Params:            p,
		attemptsRemaining: p.Retries,
	}

	j.Tasks[t.Name] = t
	j.Dag.addNode(t.Name)
	j.jobState.TaskState[t.Name] = none
	return j
}

// Task getter
func (j *Job) Task(name string) *Task {
	return j.Tasks[name]
}

// SetDownstream sets a dependency relationship between two tasks in the job.
// The dependent task is downstream of the independent task and
// waits for the independent task to finish before starting
// execution.
func (j *Job) SetDownstream(ind, dep *Task) *Job {
	j.Dag.setDownstream(ind.Name, dep.Name)
	return j
}

func (j *Job) run(reads chan readOp) error {
	if !j.Dag.validate() {
		return fmt.Errorf("Invalid Dag for job %s", j.Name)
	}

	writes := make(chan writeOp)
	taskState := j.jobState.TaskState

	for {
		for t, task := range j.Tasks {
			// Start the independent tasks
			if taskState[t] == none && !j.Dag.isDownstream(t) {
				taskState[t] = running
				go task.run(writes)
			}

			// Start the tasks that need to be re-tried
			if taskState[t] == upForRetry {
				taskState[t] = running
				go task.run(writes)
			}

			// If dependencies are done, start the dependent tasks
			if taskState[t] == none && j.Dag.isDownstream(t) {
				upstreamDone := true
				upstreamSuccessful := false
				for _, us := range j.Dag.dependencies(t) {
					if taskState[us] == none || taskState[us] == running || taskState[us] == upForRetry {
						upstreamDone = false
					}
					if taskState[us] == successful {
						upstreamSuccessful = true
					}
				}

				if upstreamDone && task.Params.TriggerRule == "allDone" {
					taskState[t] = running
					go task.run(writes)
				}

				if upstreamSuccessful && task.Params.TriggerRule == "allSuccessful" {
					taskState[t] = running
					go task.run(writes)
				}

				if upstreamDone && !upstreamSuccessful && task.Params.TriggerRule == "allSuccessful" {
					taskState[t] = skipped
				}
			}
		}

		select {
		// Respond to requests for job state
		case read := <-reads:
			read.resp <- j.jobState
		// Receive updates on task state
		case write := <-writes:
			taskState[write.key] = write.val
			// Acknowledge the update
			write.resp <- true
		}

		j.updateJobState()

		if j.allDone() {
			break
		}
	}

	return nil
}

func (j *Job) updateJobState() {
	if j.isRunning() {
		j.jobState.State = running
	}
	if j.allSuccessful() {
		j.jobState.State = successful
	}
	if j.anyFailed() {
		j.jobState.State = failed
	}
}

func (j *Job) allDone() bool {
	for _, v := range j.jobState.TaskState {
		if v == none || v == running {
			return false
		}
	}
	return true
}

func (j *Job) allSuccessful() bool {
	for _, v := range j.jobState.TaskState {
		if v != successful {
			return false
		}
	}
	return true
}

func (j *Job) isRunning() bool {
	for _, v := range j.jobState.TaskState {
		if v == running || v == upForRetry {
			return true
		}
	}
	return false
}

func (j *Job) anyFailed() bool {
	for _, v := range j.jobState.TaskState {
		if v == failed {
			return true
		}
	}
	return false
}

// A Task is the unit of work that makes up a job. Whenever a task is executed, it
// calls its associated operator.
type Task struct {
	Name              string
	Operator          Operator
	Params            TaskParams
	attemptsRemaining int
}

func (t *Task) run(writes chan writeOp) error {
	res, err := t.Operator.Run()
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	logMsg := "task %v reached state %v - %v attempt(s) remaining - result %v"

	if err != nil && t.attemptsRemaining > 0 {
		log.Printf(logMsg, t.Name, "failure", t.attemptsRemaining, err)
		t.attemptsRemaining = t.attemptsRemaining - 1
		write := writeOp{t.Name, upForRetry, make(chan bool)}
		writes <- write
		<-write.resp
		return nil
	}

	if err != nil && t.attemptsRemaining <= 0 {
		log.Printf(logMsg, t.Name, "failure", t.attemptsRemaining, err)
		write := writeOp{t.Name, failed, make(chan bool)}
		writes <- write
		<-write.resp
		return err
	}

	log.Printf(logMsg, t.Name, "success", t.attemptsRemaining, res)
	write := writeOp{t.Name, successful, make(chan bool)}
	writes <- write
	<-write.resp
	return nil
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format(time.RFC3339) + " [GOFLOW] - " + string(bytes))
}

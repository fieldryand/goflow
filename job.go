package goflow

import (
	"fmt"
	"log"
	"time"

	"github.com/fieldryand/goflow/op"
)

// A Job is a workflow consisting of independent and dependent tasks
// organized into a graph.
type Job struct {
	Name     string
	Tasks    map[string]*Task
	jobState *jobState
	Dag      Dag
}

// NewJob returns a new job.
func NewJob(name string) *Job {
	j := Job{
		Name:     name,
		Dag:      make(Dag),
		Tasks:    make(map[string]*Task),
		jobState: newJobState()}
	return &j
}

// Jobs and tasks are stateful.
type state string

const (
	none       state = "None"
	running          = "Running"
	upForRetry       = "UpForRetry"
	failed           = "Failed"
	successful       = "Successful"
)

type jobState struct {
	State     state            `json:"state"`
	TaskState map[string]state `json:"taskState"`
}

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

// Task adds a task to a job.
func (j *Job) Task(name string, op op.Operator) *Job {
	t := newTask(name, op)
	j.Tasks[t.Name] = t
	j.Dag.addNode(t.Name)
	j.jobState.TaskState[t.Name] = none
	return j
}

// SetDownstream sets a dependency relationship between two tasks in the job.
// The dependent task is downstream of the independent task and
// waits for the independent task to finish before starting
// execution.
func (j *Job) SetDownstream(ind, dep string) *Job {
	j.Dag.setDownstream(ind, dep)
	return j
}

func (j *Job) run(reads chan readOp) error {
	if !j.Dag.validate() {
		return fmt.Errorf("Invalid Dag for job %s", j.Name)
	}

	writes := make(chan writeOp)

	for {
		for t, task := range j.Tasks {
			// Start the independent tasks
			if j.jobState.TaskState[t] == none && !j.Dag.isDownstream(t) {
				j.jobState.TaskState[t] = running
				go task.run(writes)
			}

			// If dependencies are done, start the dependent tasks
			if j.jobState.TaskState[t] == none && j.Dag.isDownstream(t) {
				upstreamDone := true
				for _, us := range j.Dag.dependencies(t) {
					if j.jobState.TaskState[us] == none || j.jobState.TaskState[us] == running {
						upstreamDone = false
					}
				}

				if upstreamDone {
					j.jobState.TaskState[t] = running
					go task.run(writes)
				}
			}
		}

		select {
		// Respond to requests for job state
		case read := <-reads:
			read.resp <- j.jobState
		// Receive updates on task state
		case write := <-writes:
			j.jobState.TaskState[write.key] = write.val
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
	Name     string
	operator op.Operator
}

func newTask(name string, op op.Operator) *Task {
	t := Task{name, op}
	return &t
}

func (t *Task) run(writes chan writeOp) error {
	res, err := t.operator.Run()
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	logMsg := "task %v reached state %v with result %v"

	if err != nil {
		log.Printf(logMsg, t.Name, "failure", err)
		write := writeOp{t.Name, failed, make(chan bool)}
		writes <- write
		<-write.resp
		return err
	}

	log.Printf(logMsg, t.Name, "success", res)
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

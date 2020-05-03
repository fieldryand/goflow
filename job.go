package goflow

import (
	"fmt"
	"log"

	"github.com/fieldryand/goflow/operator"
)

// A job is a workflow consisting of independent and dependent tasks
// organized into a graph.
type Job struct {
	Name     string
	Tasks    map[string]*Task
	jobState *jobState
	Dag      *dag
}

// Returns a new job.
func NewJob(name string) *Job {
	j := Job{
		Name:     name,
		Dag:      newDag(),
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

// Adds a task to a job.
func (j *Job) AddTask(t *Task) *Job {
	j.Tasks[t.Name] = t
	j.Dag.addNode(t.Name)
	j.jobState.TaskState[t.Name] = none
	return j
}

// Sets a dependency relationship between two tasks in the job.
// The dependent task is downstream of the independent task and
// waits for the independent task to finish before starting
// execution.
func (j *Job) SetDownstream(ind, dep *Task) *Job {
	j.Dag.setDownstream(ind.Name, dep.Name)
	return j
}

func (j *Job) allDone() bool {
	done := true
	for _, v := range j.jobState.TaskState {
		if v == none || v == running {
			done = false
		}
	}
	return done
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

func (j *Job) isDownstream(taskName string) bool {
	ind := j.Dag.independentNodes()

	for _, name := range ind {
		if taskName == name {
			return false
		}
	}

	return true
}

func (j *Job) run(reads chan readOp) error {
	if !j.Dag.validate() {
		return &invalidDagError{}
	}

	ind := j.Dag.independentNodes()

	writes := make(chan writeOp)

	// Start the independent tasks
	for _, name := range ind {
		go j.Tasks[name].run(writes)
	}

	// Run downstream tasks
	for {
		select {
		case read := <-reads:
			read.resp <- j.jobState
		case write := <-writes:
			j.jobState.TaskState[write.key] = write.val
			if j.isRunning() {
				j.jobState.State = running
			}
			if j.allSuccessful() {
				j.jobState.State = successful
			}
			if write.val == failed {
				return fmt.Errorf("Job failed on task %s", write.key)
			}
			write.resp <- true
		}
		if j.allDone() {
			break
		} else {
			// for each task
			for _, t := range j.Tasks {
				if j.jobState.TaskState[t.Name] == none && j.isDownstream(t.Name) {
					upstream_done := true
					// iterate over the dependencies
					for _, us := range j.Dag.dependencies(t.Name) {
						// if any upstream task is not done, set the flag to false
						if j.jobState.TaskState[us] == none || j.jobState.TaskState[us] == running {
							upstream_done = false
						}
					}

					if upstream_done {
						j.jobState.TaskState[t.Name] = running
						go t.run(writes)
					}
				}
			}
		}
	}

	return nil
}

// Tasks are the units of work that make up a job. Whenever a task is executed, it
// calls its associated operator.
type Task struct {
	Name     string
	operator operator.Operator
}

// Returns a Task.
func NewTask(name string, op operator.Operator) *Task {
	t := Task{name, op}
	return &t
}

func (t *Task) run(writes chan writeOp) error {
	res, err := t.operator.Run()

	if err != nil {
		log.Printf("| Task %-16v | failed | %9v", t.Name, err)
		write := writeOp{t.Name, failed, make(chan bool)}
		writes <- write
		<-write.resp
		return err
	} else {
		log.Printf("| Task %-16v | success | %9v", t.Name, res)
		write := writeOp{t.Name, successful, make(chan bool)}
		writes <- write
		<-write.resp
		return nil
	}
}

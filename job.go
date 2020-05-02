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
	dag      *dag
}

// Returns a new job.
func NewJob(name string) *Job {
	j := Job{
		Name:     name,
		dag:      newDag(),
		Tasks:    make(map[string]*Task),
		jobState: newJobState()}
	return &j
}

// Jobs and tasks are stateful.
type state string

const (
	None       state = "None"
	Running          = "Running"
	UpForRetry       = "UpForRetry"
	Failed           = "Failed"
	Successful       = "Successful"
)

type jobState struct {
	State     state            `json:"state"`
	TaskState map[string]state `json:"taskState"`
}

func newJobState() *jobState {
	js := jobState{None, make(map[string]state)}
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
	j.dag.addNode(t.Name)
	j.jobState.TaskState[t.Name] = None
	return j
}

// Sets a dependency relationship between two tasks in the job.
// The dependent task is downstream of the independent task and
// waits for the independent task to finish before starting
// execution.
func (j *Job) SetDownstream(ind, dep *Task) *Job {
	j.dag.setDownstream(ind.Name, dep.Name)
	return j
}

func (j *Job) allDone() bool {
	done := true
	for _, v := range j.jobState.TaskState {
		if v == None || v == Running {
			done = false
		}
	}
	return done
}

func (j *Job) allSuccessful() bool {
	for _, v := range j.jobState.TaskState {
		if v != Successful {
			return false
		}
	}
	return true
}

func (j *Job) isRunning() bool {
	for _, v := range j.jobState.TaskState {
		if v == Running || v == UpForRetry {
			return true
		}
	}
	return false
}

func (j *Job) isDownstream(taskName string) bool {
	ind := j.dag.independentNodes()

	for _, name := range ind {
		if taskName == name {
			return false
		}
	}

	return true
}

func (j *Job) run(reads chan readOp) error {
	if !j.dag.validate() {
		return &invalidDagError{}
	}

	ind := j.dag.independentNodes()

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
				j.jobState.State = Running
			}
			if j.allSuccessful() {
				j.jobState.State = Successful
			}
			if write.val == Failed {
				return fmt.Errorf("Job failed on task %s", write.key)
			}
			write.resp <- true
		}
		if j.allDone() {
			break
		} else {
			// for each task
			for _, t := range j.Tasks {
				if j.jobState.TaskState[t.Name] == None && j.isDownstream(t.Name) {
					upstream_done := true
					// iterate over the dependencies
					for _, us := range j.dag.dependencies(t.Name) {
						// if any upstream task is not done, set the flag to false
						if j.jobState.TaskState[us] == None || j.jobState.TaskState[us] == Running {
							upstream_done = false
						}
					}

					if upstream_done {
						j.jobState.TaskState[t.Name] = Running
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
		write := writeOp{t.Name, Failed, make(chan bool)}
		writes <- write
		<-write.resp
		return err
	} else {
		log.Printf("| Task %-16v | success | %9v", t.Name, res)
		write := writeOp{t.Name, Successful, make(chan bool)}
		writes <- write
		<-write.resp
		return nil
	}
}

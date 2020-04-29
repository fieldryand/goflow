package goflow

import (
	"fmt"
	"log"
	"os"

	"github.com/fieldryand/goflow/operator"
)

// A job is a workflow consisting of independent and dependent tasks
// organized into a graph.
type Job struct {
	Name      string
	Tasks     map[string]*Task
	logger    *log.Logger
	dag       *dag
	TaskState map[string]string
}

// Returns a new job.
func NewJob(name string) *Job {
	j := Job{
		Name:      name,
		logger:    log.New(os.Stdout, "jobLogger:", log.Lshortfile),
		dag:       newDag(),
		Tasks:     make(map[string]*Task),
		TaskState: make(map[string]string)}
	return &j
}

type jobError struct {
	task string
}

type writeOp struct {
	key  string
	val  string
	resp chan bool
}

type readOp struct {
	Resp chan map[string]string
}

func (e *jobError) Error() string {
	return fmt.Sprintf("Job failed on task %s", e.task)
}

// Adds a task to a job.
func (j *Job) AddTask(t *Task) *Job {
	j.Tasks[t.name] = t
	j.dag.addNode(t.name)
	j.TaskState[t.name] = "None"
	return j
}

// Sets a dependency relationship between two tasks in the job.
// The dependent task is downstream of the independent task and
// waits for the independent task to finish before starting
// execution.
func (j *Job) SetDownstream(ind, dep *Task) *Job {
	j.dag.setDownstream(ind.name, dep.name)
	return j
}

func (j *Job) allDone() bool {
	done := true
	for _, v := range j.TaskState {
		if v == "None" || v == "Running" {
			done = false
		}
	}
	return done
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
			read.Resp <- j.TaskState
		case write := <-writes:
			j.TaskState[write.key] = write.val
			if write.val == "Failure" {
				return &jobError{write.key}
			}
			write.resp <- true
		}
		if j.allDone() {
			break
		} else {
			// for each task
			for _, t := range j.Tasks {
				if j.TaskState[t.name] == "None" && j.isDownstream(t.name) {
					upstream_done := true
					// iterate over the dependencies
					for _, us := range j.dag.dependencies(t.name) {
						// if any upstream task is not done, set the flag to false
						if j.TaskState[us] == "None" || j.TaskState[us] == "Running" {
							upstream_done = false
						}
					}

					if upstream_done {
						j.TaskState[t.name] = "Running"
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
	name     string
	logger   *log.Logger
	operator operator.Operator
}

// Returns a Task.
func NewTask(name string, op operator.Operator) *Task {
	l := log.New(os.Stdout, "taskLogger:", log.Lshortfile)
	t := Task{name, l, op}
	return &t
}

func (t *Task) run(writes chan writeOp) error {
	res, err := t.operator.Run()

	if err != nil {
		t.logger.Println("Task", t.name, "failed:", err)
		write := writeOp{t.name, "Failure", make(chan bool)}
		writes <- write
		<-write.resp
		return err
	} else {
		t.logger.Println("Task", t.name, "succeeded with result", res)
		write := writeOp{t.name, "Success", make(chan bool)}
		writes <- write
		<-write.resp
		return nil
	}
}

package core

import (
	"fmt"
	"log"
	"os"
)

type Job struct {
	name      string
	logger    *log.Logger
	dag       *Dag
	tasks     map[string]*task
	TaskState map[string]string
}

func NewJob(name string) *Job {
	j := Job{
		name:      name,
		logger:    log.New(os.Stdout, "jobLogger:", log.Lshortfile),
		dag:       NewDag(),
		tasks:     make(map[string]*task),
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

type ReadOp struct {
	Resp chan map[string]string
}

func (e *jobError) Error() string {
	return fmt.Sprintf("Job failed on task %s", e.task)
}

func (j *Job) AddTask(t *task) *Job {
	j.tasks[t.name] = t
	j.dag.addNode(t.name)
	j.TaskState[t.name] = "None"
	return j
}

func (j *Job) SetDownstream(ind, dep *task) *Job {
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

func (j *Job) Run(reads chan ReadOp) error {
	if !j.dag.validate() {
		return &InvalidDagError{}
	}

	ind := j.dag.independentNodes()

	writes := make(chan writeOp)

	// Start the independent tasks
	for _, name := range ind {
		go j.tasks[name].run(writes)
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
			for _, t := range j.tasks {
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

type task struct {
	name     string
	logger   *log.Logger
	operator operator
}

func Task(name string, op operator) *task {
	l := log.New(os.Stdout, "taskLogger:", log.Lshortfile)
	t := task{name, l, op}
	return &t
}

func (t *task) run(writes chan writeOp) error {
	res, err := t.operator.run()

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

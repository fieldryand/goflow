package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type job struct {
	name   string
	logger *log.Logger
	dag    *Dag
	tasks  map[string]*task
}

func Job(name string) *job {
	d := NewDag()
	l := log.New(os.Stdout, "jobLogger:", log.Lshortfile)
	j := job{name, l, d, make(map[string]*task)}
	return &j
}

type jobError struct {
	task string
}

func (e *jobError) Error() string {
	return fmt.Sprintf("Job failed on task %s", e.task)
}

func (j *job) AddTask(t *task) *job {
	j.tasks[t.Name] = t
	j.dag.addNode(t.Name)
	return j
}

func (j *job) SetDownstream(ind, dep *task) *job {
	j.dag.setDownstream(ind.Name, dep.Name)
	return j
}

func (j *job) Status() map[string]string {
	s := make(map[string]string)
	for k, v := range j.tasks {
		s[k] = v.Status
	}
	encoded, _ := json.Marshal(s)
	j.logger.Println("Task status:", string(encoded))
	return s
}

func (j *job) Run(stat chan string) error {
	if !j.dag.validate() {
		stat <- "Failed"
		return &InvalidDagError{}
	} else {
		err := j.run_tasks()
		if err != nil {
			j.logger.Println(err)
			stat <- "Failed"
			return err
		} else {
			stat <- "Success"
			return nil
		}
	}
}

func (j *job) run_tasks() error {
	var wg sync.WaitGroup

	total := len(j.tasks)
	done := 0
	ind := j.dag.independentNodes()

	// Run the independent tasks
	for _, name := range ind {
		wg.Add(1)
		done += 1
		go j.tasks[name].run(&wg)
	}

	wg.Wait()

	// Run downstream tasks
	for {
		if done == total {
			break
		} else {
			// for each task
			for _, t := range j.tasks {
				if !t.isDone() {
					upstream_done := true
					// iterate over the dependencies
					for _, us := range j.dag.dependencies(t.Name) {
						// if any upstream task is not done, set the flag to false
						if !j.tasks[us].isDone() {
							upstream_done = false
						}
					}

					if upstream_done {
						wg.Add(1)
						done += 1
						go t.run(&wg)
					}
				}
			}
		}

		wg.Wait()
	}

	for taskName, t := range j.tasks {
		if t.Status == "Failed" {
			return &jobError{taskName}
		}
	}

	return nil
}

type task struct {
	Name     string
	Status   string
	logger   *log.Logger
	operator operator
}

func Task(name string, op operator) *task {
	l := log.New(os.Stdout, "taskLogger:", log.Lshortfile)
	t := task{name, "None", l, op}
	return &t
}

func (t *task) isDone() bool {
	if t.Status == "Success" || t.Status == "Failed" {
		return true
	} else {
		return false
	}
}

func (t *task) run(wg *sync.WaitGroup) error {
	defer wg.Done()

	res, err := t.operator.run()

	if err != nil {
		t.logger.Println("Task", t.Name, "failed:", err)
		t.Status = "Failed"
		return err
	} else {
		t.logger.Println("Task", t.Name, "succeeded with result", res)
		t.Status = "Success"
		return nil
	}
}

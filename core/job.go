package core

import (
	"log"
	"os"
	"sync"
)

type Job struct {
	name  string
	dag   *Dag
	tasks []*Task
}

func NewJob(name string) *Job {
	d := NewDag()
	j := Job{name, d, make([]*Task, 0)}
	return &j
}

func (j *Job) addTask(task *Task) {
	j.tasks = append(j.tasks, task)
	j.dag.addNode(task.name)
}

func (j *Job) setDownstream(ind, dep *Task) {
	j.dag.setDownstream(ind.name, dep.name)
}

func (j *Job) run() error {
	if valid := j.dag.validate(); valid != true {
		return &InvalidDagError{}
	} else {
		err := j.run_tasks()
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (j *Job) run_tasks() error {
	var wg sync.WaitGroup

	total := len(j.tasks)
	done := 0
	ind := j.dag.independentNodes()

	// Run the independent tasks
	for _, name := range ind {
		for _, t := range j.tasks {
			if name == t.name {
				wg.Add(1)
				done += 1
				go t.run(&wg)
			}
		}
	}

	wg.Wait()

	// Run downstream tasks
	for {
		if done == total {
			break
		} else {
			// for each task
			for _, t := range j.tasks {
				// if the status is None
				if t.status == "None" {
					upstream_done := true
					// iterate over the dependencies
					for _, us := range j.dag.dependencies(t.name) {
						for _, tsk := range j.tasks {
							// if the status of an upstream tsk is None, then upstream dependencies are not done
							if tsk.status == "None" {
								if tsk.name == us {
									upstream_done = false
								}
							}
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

	return nil
}

type Task struct {
	name     string
	status   string
	operator Operator
}

func NewTask(name string, op Operator) *Task {
	t := Task{name, "None", op}
	return &t
}

func (t *Task) run(wg *sync.WaitGroup) error {
	defer wg.Done()
	var (
		logger = log.New(os.Stdout, "taskLogger:", log.Lshortfile)
	)

	res, err := t.operator.run()

	if err != nil {
		logger.Println("Task failed")
		t.status = "Failed"
		return err
	} else {
		logger.Println("Task", t.name, "succeeded with result", res)
		t.status = "Success"
		return nil
	}
}

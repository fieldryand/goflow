package core

import (
	"log"
	"os"
	"sync"
)

type job struct {
	name  string
	dag   *Dag
	tasks []*task
}

func Job(name string) *job {
	d := NewDag()
	j := job{name, d, make([]*task, 0)}
	return &j
}

func (j *job) addTask(t *task) {
	j.tasks = append(j.tasks, t)
	j.dag.addNode(t.name)
}

func (j *job) setDownstream(ind, dep *task) {
	j.dag.setDownstream(ind.name, dep.name)
}

func (j *job) run() error {
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

func (j *job) run_tasks() error {
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

type task struct {
	name     string
	status   string
	operator operator
}

func Task(name string, op operator) *task {
	t := task{name, "None", op}
	return &t
}

func (t *task) run(wg *sync.WaitGroup) error {
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

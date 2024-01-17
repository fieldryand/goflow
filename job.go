package goflow

import (
	"errors"
	"log"
	"sync"
)

// A Job is a workflow consisting of independent and dependent tasks
// organized into a graph.
type Job struct {
	Name     string
	Tasks    map[string]*Task
	Schedule string
	Dag      dag
	Active   bool
	state    state
	sync.RWMutex
}

// Jobs and tasks are stateful.
type state string

const (
	none       state = "notstarted"
	running    state = "running"
	upForRetry state = "upforretry"
	skipped    state = "skipped"
	failed     state = "failed"
	successful state = "successful"
)

func (j *Job) loadState() state {
	if !j.allDone() {
		j.storeState(running)
	}
	if j.allSuccessful() {
		j.storeState(successful)
	}
	if j.allDone() && j.anyFailed() {
		j.storeState(failed)
	}
	return j.state
}

func (j *Job) loadTaskState(task string) state {
	j.RLock()
	result := j.Tasks[task].state
	j.RUnlock()
	return result
}

func (j *Job) storeState(value state) {
	j.Lock()
	j.state = value
	j.Unlock()
}

func (j *Job) storeTaskState(task string, value state) {
	j.Lock()
	j.Tasks[task].state = value
	j.Unlock()
}

func (j *Job) rangeOverTasks(f func(key string, value *Task) bool) {
	j.Lock()
	for k, v := range j.Tasks {
		if !f(k, v) {
			break
		}
	}
	j.Unlock()
}

type writeOp struct {
	key  string
	val  state
	resp chan bool
}

// Initialize a job.
func (j *Job) initialize() *Job {
	j.Dag = make(dag)
	j.Tasks = make(map[string]*Task)
	j.storeState(none)
	return j
}

// Add a task to a job.
func (j *Job) Add(t *Task) error {
	if t.Name == "" {
		return errors.New("\"\" is not a valid job name")
	}

	if j.Dag == nil {
		j.initialize()
	}

	if !(t.TriggerRule == allDone || t.TriggerRule == allSuccessful) {
		t.TriggerRule = allSuccessful
	}

	t.attemptsRemaining = t.Retries
	t.state = none

	j.Tasks[t.Name] = t
	j.Dag.addNode(t.Name)
	j.storeTaskState(t.Name, none)

	return nil
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

func (j *Job) run() error {

	log.Printf("starting job <%v>", j.Name)

	writes := make(chan writeOp)

	for {
		for t, task := range j.Tasks {
			// Start the independent tasks
			v := j.loadTaskState(t)
			if v == none && !j.Dag.isDownstream(t) {
				j.storeTaskState(t, running)
				go task.run(writes)
			}

			// Start the tasks that need to be re-tried
			if v == upForRetry {
				task.RetryDelay.wait(task.Name, task.Retries-task.attemptsRemaining)
				task.attemptsRemaining = task.attemptsRemaining - 1
				j.storeTaskState(t, running)
				go task.run(writes)
			}

			// If dependencies are done, start the dependent tasks
			if v == none && j.Dag.isDownstream(t) {
				upstreamDone := true
				upstreamSuccessful := true
				for _, us := range j.Dag.dependencies(t) {
					w := j.loadTaskState(us)
					if w == none || w == running || w == upForRetry {
						upstreamDone = false
					}
					if w != successful {
						upstreamSuccessful = false
					}
				}

				if upstreamDone && task.TriggerRule == allDone {
					j.storeTaskState(t, running)
					go task.run(writes)
				}

				if upstreamSuccessful && task.TriggerRule == allSuccessful {
					j.storeTaskState(t, running)
					go task.run(writes)
				}

				if upstreamDone && !upstreamSuccessful && task.TriggerRule == allSuccessful {
					j.storeTaskState(t, skipped)
					go task.skip(writes)
				}

			}
		}

		// Receive updates on task state
		write := <-writes
		j.storeTaskState(write.key, write.val)

		// Acknowledge the update
		write.resp <- true

		if j.allDone() {
			break
		}
	}

	return nil
}

func (j *Job) allDone() bool {
	out := true
	j.rangeOverTasks(func(k string, v *Task) bool {
		if v.state == none || v.state == running || v.state == upForRetry {
			out = false
		}
		return out
	})
	return out
}

func (j *Job) allSuccessful() bool {
	out := true
	j.rangeOverTasks(func(k string, v *Task) bool {
		if v.state != successful {
			out = false
		}
		return out
	})
	return out
}

func (j *Job) anyFailed() bool {
	out := false
	j.rangeOverTasks(func(k string, v *Task) bool {
		if v.state == failed {
			out = true
		}
		return out
	})
	return out
}

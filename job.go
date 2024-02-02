package goflow

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/philippgille/gokv"
)

// A Job is a workflow consisting of independent and dependent tasks
// organized into a graph.
type Job struct {
	Name     string
	Tasks    []*Task
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
	result := none
	for _, t := range j.Tasks {
		if t.Name == task {
			result = t.state
			break
		}
	}
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
	for _, t := range j.Tasks {
		if t.Name == task {
			t.state = value
		}
	}
	j.Unlock()
}

type writeOp struct {
	key string
	val state
	res pipe
}

// Initialize a job.
func (j *Job) initialize() *Job {
	j.Dag = make(dag)
	j.Tasks = make([]*Task, 0)
	j.storeState(none)
	return j
}

// Add a task to a job.
func (j *Job) Add(t *Task) error {
	if t.Name == "" {
		return errors.New("\"\" is not a valid task name")
	}

	if j.Dag == nil {
		j.initialize()
	}

	if !(t.TriggerRule == allDone || t.TriggerRule == allSuccessful) {
		t.TriggerRule = allSuccessful
	}

	t.attempts = t.Retries
	t.state = none

	j.Tasks = append(j.Tasks, t)
	j.Dag.addNode(t.Name)
	j.storeTaskState(t.Name, none)

	return nil
}

// SetDownstream sets a dependency relationship between two tasks in the job.
// The dependent task is downstream of the independent task and
// waits for the independent task to finish before starting
// execution.
func (j *Job) SetDownstream(ind, dep string) error {
	indExists := false
	depExists := false
	for _, t := range j.Tasks {
		if ind == t.Name {
			indExists = true
		}
		if dep == t.Name {
			depExists = true
		}
	}
	if !indExists {
		return fmt.Errorf("Job does not contain task %s", ind)
	}
	if !depExists {
		return fmt.Errorf("Job does not contain task %s", dep)
	}
	j.Dag.setDownstream(ind, dep)
	return nil
}

func (j *Job) run(store gokv.Store, e *Execution) error {

	log.Printf("starting job: name=%v, ID=%v", j.Name, e.ID)

	res := make(pipe)
	// defalut key value
	res["job_id"] = e.ID
	writes := make(chan writeOp)

	for {
		for _, task := range j.Tasks {

			// Start the independent tasks
			v := j.loadTaskState(task.Name)
			if v == none && !j.Dag.isDownstream(task.Name) {
				j.storeTaskState(task.Name, running)
				if task.PipeOperator != nil {
					go task.runWithPipe(res, writes)
				} else {
					go task.run(res, writes)
				}
			}

			// Start the tasks that need to be re-tried
			if v == upForRetry {
				task.RetryDelay.wait(task.Name, task.Retries-task.attempts)
				task.attempts = task.attempts - 1
				j.storeTaskState(task.Name, running)
				if task.PipeOperator != nil {
					go task.runWithPipe(res, writes)
				} else {
					go task.run(res, writes)
				}
			}

			// If dependencies are done, start the dependent tasks
			if v == none && j.Dag.isDownstream(task.Name) {
				upstreamDone := true
				upstreamSuccessful := true
				for _, us := range j.Dag.dependencies(task.Name) {
					w := j.loadTaskState(us)
					if w == none || w == running || w == upForRetry {
						upstreamDone = false
					}
					if w != successful {
						upstreamSuccessful = false
					}
				}

				if upstreamDone && task.TriggerRule == allDone {
					j.storeTaskState(task.Name, running)
					if task.PipeOperator != nil {
						go task.runWithPipe(res, writes)
					} else {
						go task.run(res, writes)
					}
				}

				if upstreamSuccessful && task.TriggerRule == allSuccessful {
					j.storeTaskState(task.Name, running)
					if task.PipeOperator != nil {
						go task.runWithPipe(res, writes)
					} else {
						go task.run(res, writes)
					}
				}

				if upstreamDone && !upstreamSuccessful && task.TriggerRule == allSuccessful {
					j.storeTaskState(task.Name, skipped)
					go task.skip(res, writes)
				}
			}
		}

		// Receive updates on task state
		write := <-writes
		log.Printf("task update: job=%v, ID=%v, task=%v, state=%v", j.Name, e.ID, write.key, write.val)
		j.storeTaskState(write.key, write.val)

		if write.val == successful {
			res = write.res
		}

		// Sync to store
		e.State = j.loadState()
		e.ElapsedSeconds = time.Since(e.StartTimestamp).Seconds()
		syncStateToStore(store, e, write.key, write.val)
		//log.Printf("task update: job=%v, ID=%v, task=%v, state=%v", j.Name, e.ID, write.key, write.val)

		if j.allDone() {
			break
		}
	}

	log.Printf("job done: name=%v, ID=%v, state=%v", j.Name, e.ID, j.loadState())

	return nil
}

func (j *Job) allDone() bool {
	j.RLock()
	out := true
	for _, t := range j.Tasks {
		if t.state == none || t.state == running || t.state == upForRetry {
			out = false
		}
	}
	j.RUnlock()
	return out
}

func (j *Job) allSuccessful() bool {
	j.RLock()
	out := true
	for _, t := range j.Tasks {
		if t.state != successful {
			out = false
		}
	}
	j.RUnlock()
	return out
}

func (j *Job) anyFailed() bool {
	j.RLock()
	out := false
	for _, t := range j.Tasks {
		if t.state == failed {
			out = true
		}
	}
	j.RUnlock()
	return out
}

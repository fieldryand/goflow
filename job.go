package goflow

import (
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
	Tasks    map[string]*Task
	Schedule string
	Dag      dag
	Active   bool
	state    state
	tasks    []string
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
}

// Initialize a job.
func (j *Job) initialize() *Job {
	j.Dag = make(dag)
	j.Tasks = make(map[string]*Task)
	j.tasks = make([]string, 0)
	j.storeState(none)
	return j
}

// Add a task to a job.
func (j *Job) Add(t *Task) *Job {
	if j.Dag == nil {
		j.initialize()
	}

	if !(t.TriggerRule == allDone || t.TriggerRule == allSuccessful) {
		t.TriggerRule = allSuccessful
	}

	t.remaining = t.Retries

	j.Tasks[t.Name] = t
	j.tasks = append(j.tasks, t.Name)
	j.Dag.addNode(t.Name)
	j.storeTaskState(t.Name, none)
	return j
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

func (j *Job) run(store gokv.Store, e *execution) error {

	if !j.Dag.validate() {
		return fmt.Errorf("Invalid Dag for job %s", j.Name)
	}

	log.Printf("jobID=%v, jobname=%v, msg=starting", e.ID, j.Name)

	writes := make(chan writeOp)

	for {
		for _, task := range j.Tasks {

			// Start the independent tasks
			v := j.loadTaskState(task.Name)
			if v == none && !j.Dag.isDownstream(task.Name) {
				j.storeTaskState(task.Name, running)
				log.Printf("jobID=%v, job=%v, task=%v, msg=starting", e.ID, j.Name, task.Name)
				go task.run(writes)
			}

			// Start the tasks that need to be re-tried
			if v == upForRetry {
				task.RetryDelay.wait(task.Name, task.Retries-task.remaining)
				task.remaining = task.remaining - 1
				j.storeTaskState(task.Name, running)
				log.Printf("jobID=%v, job=%v, task=%v, msg=starting", e.ID, j.Name, task.Name)
				go task.run(writes)
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
					log.Printf("jobID=%v, job=%v, task=%v, msg=starting", e.ID, j.Name, task.Name)
					go task.run(writes)
				}

				if upstreamSuccessful && task.TriggerRule == allSuccessful {
					j.storeTaskState(task.Name, running)
					log.Printf("jobID=%v, job=%v, task=%v, msg=starting", e.ID, j.Name, task.Name)
					go task.run(writes)
				}

				if upstreamDone && !upstreamSuccessful && task.TriggerRule == allSuccessful {
					j.storeTaskState(task.Name, skipped)
					log.Printf("jobID=%v, job=%v, task=%v, msg=skipping", e.ID, j.Name, task.Name)
					go task.skip(writes)
				}

			}
		}

		// Receive updates on task state
		write := <-writes
		j.storeTaskState(write.key, write.val)
		log.Printf("jobID=%v, job=%v, task=%v, msg=%v", e.ID, j.Name, write.key, write.val)

		// Sync to store
		e.State = j.loadState()
		e.ModifiedTimestamp = time.Now().UTC().Format(time.RFC3339Nano)
		syncStateToStore(store, e, write.key, write.val)

		if j.allDone() {
			break
		}
	}

	log.Printf("jobID=%v, job=%v, msg=%v", e.ID, j.Name, j.loadState())

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

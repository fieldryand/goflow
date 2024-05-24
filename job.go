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

// These constants represent the possible states of a job or task.
//
// They are exported so that users can program with them within
// contextual operators.
const (
	None       state = "notstarted"
	Running    state = "running"
	UpForRetry state = "upforretry"
	Skipped    state = "skipped"
	Failed     state = "failed"
	Successful state = "successful"
)

func (j *Job) loadState() state {
	if !j.allDone() {
		j.storeState(Running)
	}
	if j.allSuccessful() {
		j.storeState(Successful)
	}
	if j.allDone() && j.anyFailed() {
		j.storeState(Failed)
	}
	return j.state
}

func (j *Job) loadTaskState(task string) state {
	j.RLock()
	result := None
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

func (j *Job) storeTaskState(task string, value state, result any, err error) {
	j.Lock()
	for _, t := range j.Tasks {
		if t.Name == task {
			t.state = value
			t.result = result
			t.err = err
		}
	}
	j.Unlock()
}

type writeOp struct {
	key    string
	val    state
	result string
	err    error
}

// Initialize a job.
func (j *Job) initialize() *Job {
	j.Dag = make(dag)
	j.Tasks = make(map[string]*Task)
	j.tasks = make([]string, 0)
	j.storeState(None)
	return j
}

// AddTask adds a task to a job.
func (j *Job) AddTask(t *Task) error {

	if t.Name == "" {
		return errors.New("\"\" is not a valid task name")
	}

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
	j.storeTaskState(t.Name, None, nil, nil)
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

	if !j.Dag.validate() {
		return fmt.Errorf("Invalid Dag for job %s", j.Name)
	}

	return nil
}

func (j *Job) run(store gokv.Store, e *Execution) error {

	log.Printf("jobID=%v, jobname=%v, msg=starting", e.ID, j.Name)

	writes := make(chan writeOp)

	for {
		for _, task := range j.Tasks {

			// Start the independent tasks
			v := j.loadTaskState(task.Name)
			if v == None && !j.Dag.isDownstream(task.Name) {
				j.storeTaskState(task.Name, Running, nil, nil)
				e = syncStartTsToExecution(e, task.Name)
				store.Set(e.ID.String(), e)
				log.Printf("jobID=%v, job=%v, task=%v, msg=starting", e.ID, j.Name, task.Name)
				go task.run(e, writes)
			}

			// Start the tasks that need to be re-tried
			if v == UpForRetry {
				task.RetryDelay.wait(task.Name, task.Retries-task.remaining)
				task.remaining = task.remaining - 1
				j.storeTaskState(task.Name, Running, nil, nil)
				log.Printf("jobID=%v, job=%v, task=%v, msg=starting", e.ID, j.Name, task.Name)
				go task.run(e, writes)
			}

			// If dependencies are done, start the dependent tasks
			if v == None && j.Dag.isDownstream(task.Name) {
				upstreamDone := true
				upstreamSuccessful := true
				for _, us := range j.Dag.dependencies(task.Name) {
					w := j.loadTaskState(us)
					if w == None || w == Running || w == UpForRetry {
						upstreamDone = false
					}
					if w != Successful {
						upstreamSuccessful = false
					}
				}

				if upstreamDone && task.TriggerRule == allDone {
					j.storeTaskState(task.Name, Running, nil, nil)
					e = syncStartTsToExecution(e, task.Name)
					store.Set(e.ID.String(), e)
					log.Printf("jobID=%v, job=%v, task=%v, msg=starting", e.ID, j.Name, task.Name)
					go task.run(e, writes)
				}

				if upstreamSuccessful && task.TriggerRule == allSuccessful {
					j.storeTaskState(task.Name, Running, nil, nil)
					e = syncStartTsToExecution(e, task.Name)
					store.Set(e.ID.String(), e)
					log.Printf("jobID=%v, job=%v, task=%v, msg=starting", e.ID, j.Name, task.Name)
					go task.run(e, writes)
				}

				if upstreamDone && !upstreamSuccessful && task.TriggerRule == allSuccessful {
					j.storeTaskState(task.Name, Skipped, nil, nil)
					log.Printf("jobID=%v, job=%v, task=%v, msg=skipping", e.ID, j.Name, task.Name)
					go task.skip(writes)
				}

			}
		}

		// Receive updates on task state
		write := <-writes
		j.storeTaskState(write.key, write.val, write.result, write.err)
		log.Printf("jobID=%v, job=%v, task=%v, msg=%v", e.ID, j.Name, write.key, write.val)

		// Sync to store
		e.State = j.loadState()
		e.ModifiedTs = time.Now().UTC()
		e = syncResultToExecution(e, write.key, write.val, write.result, write.err)
		store.Set(e.ID.String(), e)

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
		if t.state == None || t.state == Running || t.state == UpForRetry {
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
		if t.state != Successful {
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
		if t.state == Failed {
			out = true
		}
	}
	j.RUnlock()
	return out
}

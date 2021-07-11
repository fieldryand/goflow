package goflow

import (
	"fmt"
)

// A Job is a workflow consisting of independent and dependent tasks
// organized into a graph.
type Job struct {
	Name            string
	Tasks           map[string]*Task
	Schedule        string
	Dag             dag
	ActiveByDefault bool
	jobState        *jobState
}

// Jobs and tasks are stateful.
type state string

const (
	none       state = "None"
	running    state = "Running"
	upForRetry state = "UpForRetry"
	skipped    state = "Skipped"
	failed     state = "Failed"
	successful state = "Successful"
)

type jobState struct {
	State     state           `json:"state"`
	TaskState *stringStateMap `json:"taskState"`
}

func newJobState() *jobState {
	js := jobState{none, newStringStateMap()}
	return &js
}

type writeOp struct {
	key  string
	val  state
	resp chan bool
}

type readOp struct {
	resp    chan *jobState
	allDone bool
}

// Initialize a job.
func (j *Job) initialize() *Job {
	j.Dag = make(dag)
	j.Tasks = make(map[string]*Task)
	j.jobState = newJobState()
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

	t.attemptsRemaining = t.Retries

	j.Tasks[t.Name] = t
	j.Dag.addNode(t.Name)
	j.jobState.TaskState.Store(t.Name, none)
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

func (j *Job) run(reads chan readOp) error {
	if !j.Dag.validate() {
		return fmt.Errorf("Invalid Dag for job %s", j.Name)
	}

	writes := make(chan writeOp)
	taskState := j.jobState.TaskState

	for {
		for t, task := range j.Tasks {
			// Start the independent tasks
			v, _ := taskState.Load(t)
			if v == none && !j.Dag.isDownstream(t) {
				taskState.Store(t, running)
				go task.run(writes)
			}

			// Start the tasks that need to be re-tried
			if v == upForRetry {
				task.RetryDelay.wait(task.Name, task.Retries-task.attemptsRemaining)
				task.attemptsRemaining = task.attemptsRemaining - 1
				taskState.Store(t, running)
				go task.run(writes)
			}

			// If dependencies are done, start the dependent tasks
			if v == none && j.Dag.isDownstream(t) {
				upstreamDone := true
				upstreamSuccessful := true
				for _, us := range j.Dag.dependencies(t) {
					w, _ := taskState.Load(us)
					if w == none || w == running || w == upForRetry {
						upstreamDone = false
					}
					if w != successful {
						upstreamSuccessful = false
					}
				}

				if upstreamDone && task.TriggerRule == allDone {
					taskState.Store(t, running)
					go task.run(writes)
				}

				if upstreamSuccessful && task.TriggerRule == allSuccessful {
					taskState.Store(t, running)
					go task.run(writes)
				}

				if upstreamDone && !upstreamSuccessful && task.TriggerRule == allSuccessful {
					taskState.Store(t, skipped)
					go task.skip(writes)
				}

			}
		}

		allDone := false

		select {
		// Respond to requests for job state
		case read := <-reads:
			allDone = read.allDone
			read.resp <- j.jobState
		// Receive updates on task state
		case write := <-writes:
			taskState.Store(write.key, write.val)
			j.updateJobState()
			// Acknowledge the update
			write.resp <- true
		}

		if allDone {
			break
		}
	}

	return nil
}

func (j *Job) updateJobState() {
	if !j.allDone() {
		j.jobState.State = running
	}
	if j.allSuccessful() {
		j.jobState.State = successful
	}
	if j.anyFailed() {
		j.jobState.State = failed
	}
}

func (j *Job) allDone() bool {
	out := true
	j.jobState.TaskState.Range(func(k string, v state) bool {
		if v == none || v == running || v == upForRetry {
			out = false
		}
		return out
	})
	return out
}

func (j *Job) allSuccessful() bool {
	out := true
	j.jobState.TaskState.Range(func(k string, v state) bool {
		if v != successful {
			out = false
		}
		return out
	})
	return out
}

func (j *Job) anyFailed() bool {
	out := false
	j.jobState.TaskState.Range(func(k string, v state) bool {
		if v == failed {
			out = true
		}
		return out
	})
	return out
}

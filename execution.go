package goflow

import (
	"time"

	"github.com/google/uuid"
	"github.com/philippgille/gokv"
)

// Execution of a job.
type Execution struct {
	ID                uuid.UUID       `json:"id"`
	JobName           string          `json:"job"`
	StartedAt         string          `json:"submitted"`
	ModifiedTimestamp string          `json:"modifiedTimestamp"`
	State             state           `json:"state"`
	Tasks             []TaskExecution `json:"tasks"`
}

// TaskExecution represents the execution of a task.
type TaskExecution struct {
	Name     string `json:"name"`
	State    state  `json:"state"`
	Result   any    `json:"result"`
	Error    string `json:"error"`
	Operator any    `json:"operator"`
}

func (j *Job) newExecution() *Execution {
	taskExecutions := make([]TaskExecution, 0)
	for _, task := range j.Tasks {
		taskrun := TaskExecution{task.Name, None, nil, "", task.Operator}
		taskExecutions = append(taskExecutions, taskrun)
	}
	return &Execution{
		ID:                uuid.New(),
		JobName:           j.Name,
		StartedAt:         time.Now().UTC().Format(time.RFC3339Nano),
		ModifiedTimestamp: time.Now().UTC().Format(time.RFC3339Nano),
		State:             None,
		Tasks:             taskExecutions}
}

// Persist a new execution.
func persistNewExecution(s gokv.Store, e *Execution) error {
	key := e.ID
	return s.Set(key.String(), e)
}

type executionIndex struct {
	ExecutionIDs []string `json:"executions"`
}

// Index the job runs
func indexExecutions(s gokv.Store, e *Execution) error {

	// get the job from the execution
	j := e.JobName

	// retrieve the list of executions of that job
	i := executionIndex{}
	s.Get(j, &i)

	// append to the list
	i.ExecutionIDs = append(i.ExecutionIDs, e.ID.String())
	return s.Set(e.JobName, i)
}

// Read all the persisted executions for a given job.
func readExecutions(s gokv.Store, j string) ([]*Execution, error) {

	// retrieve the list of executions of the job
	i := executionIndex{}
	s.Get(j, &i)

	// return the list
	executions := make([]*Execution, 0)
	for _, key := range i.ExecutionIDs {
		val := Execution{}
		s.Get(key, &val)
		executions = append(executions, &val)
	}

	return executions, nil
}

// Sync the task result and state to an execution
func syncResultToExecution(e *Execution, task string, s state, r any, err error) *Execution {

	// if there is an error, convert it to a string
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	for ix, t := range e.Tasks {
		if t.Name == task {
			e.Tasks[ix].State = s
			e.Tasks[ix].Result = r
			e.Tasks[ix].Error = errString
		}
	}

	return e
}

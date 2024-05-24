package goflow

import (
	"time"

	"github.com/google/uuid"
	"github.com/philippgille/gokv"
)

// Execution of a job.
type Execution struct {
	ID         uuid.UUID       `json:"id"`
	JobName    string          `json:"job"`
	StartTs    time.Time       `json:"startTs"`
	ModifiedTs time.Time       `json:"modifiedTs"`
	State      state           `json:"state"`
	Tasks      []TaskExecution `json:"tasks"`
}

// TaskExecution represents the execution of a task.
type TaskExecution struct {
	Name     string    `json:"name"`
	State    state     `json:"state"`
	StartTs  time.Time `json:"startTs"`
	Result   string    `json:"result"`
	Error    string    `json:"error"`
	Operator any       `json:"operator"`
}

func (j *Job) newExecution() *Execution {
	taskExecutions := make([]TaskExecution, 0)
	for _, task := range j.Tasks {
		taskrun := TaskExecution{
			Name:     task.Name,
			State:    None,
			StartTs:  time.Time{},
			Result:   "",
			Error:    "",
			Operator: task.Operator}
		taskExecutions = append(taskExecutions, taskrun)
	}
	return &Execution{
		ID:         uuid.New(),
		JobName:    j.Name,
		StartTs:    time.Now().UTC(),
		ModifiedTs: time.Now().UTC(),
		State:      None,
		Tasks:      taskExecutions}
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
func syncResultToExecution(e *Execution, task string, s state, r string, err error) *Execution {

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

// Sync the starting timestamp of a task to an execution
func syncStartTsToExecution(e *Execution, task string) *Execution {

	for ix, t := range e.Tasks {
		if t.Name == task {
			e.Tasks[ix].StartTs = time.Now().UTC()
		}
	}

	return e
}

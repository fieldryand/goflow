package goflow

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/philippgille/gokv"
)

// Execution of a job.
type Execution struct {
	ID             string          `json:"id"`
	JobName        string          `json:"job"`
	StartTimestamp string          `json:"startTimestamp"`
	EndTimestamp   string          `json:"endTimestamp"`
	State          state           `json:"state"`
	TaskExecutions []taskExecution `json:"tasks"`
}

type taskExecution struct {
	Name  string `json:"name"`
	State state  `json:"state"`
}

func (j *Job) newExecution() *Execution {
	taskExecutions := make([]taskExecution, 0)
	for _, task := range j.Tasks {
		taskrun := taskExecution{task.Name, none}
		taskExecutions = append(taskExecutions, taskrun)
	}
	return &Execution{
		ID:             uuid.New().String(),
		JobName:        j.Name,
		StartTimestamp: time.Now().UTC().Format(time.RFC3339),
		EndTimestamp:   "",
		State:          none,
		TaskExecutions: taskExecutions}
}

// Persist a new execution.
func persistNewExecution(s gokv.Store, e *Execution) error {
	key := e.ID
	err := s.Set(key, e)
	if err != nil {
		fmt.Println(err)
	}
	return err
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
	i.ExecutionIDs = append(i.ExecutionIDs, e.ID)
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

// Sync the current state to the persisted execution.
func syncStateToStore(s gokv.Store, e *Execution, jobState state, taskName string, taskState state) error {
	key := e.ID
	e.State = jobState
	for ix, task := range e.TaskExecutions {
		if task.Name == taskName {
			e.TaskExecutions[ix].State = taskState
		}
	}
	return s.Set(key, e)
}
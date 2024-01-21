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
func persistNewExecution(store gokv.Store, execution *Execution) error {
	key := execution.ID
	err := store.Set(key, execution)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

type executionIndex struct {
	ExecutionIDs []string `json:"executions"`
}

// Index the job runs
func indexExecutions(store gokv.Store, execution *Execution) error {
	index := executionIndex{}
	store.Get(execution.JobName, &index)

	// add the execution ID to the index
	index.ExecutionIDs = append(index.ExecutionIDs, execution.ID)
	return store.Set(execution.JobName, index)
}

// Read all the persisted executions for a given job.
func readExecutions(store gokv.Store, jobName string) ([]*Execution, error) {
	index := executionIndex{}
	store.Get(jobName, &index)

	executions := make([]*Execution, 0)
	for _, key := range index.ExecutionIDs {
		value := Execution{}
		store.Get(key, &value)
		executions = append(executions, &value)
	}

	return executions, nil
}

// Sync the current state to the persisted execution.
func syncStateToStore(store gokv.Store, execution *Execution, jobState state, taskName string, taskState state) error {
	key := execution.ID
	execution.State = jobState
	for ix, task := range execution.TaskExecutions {
		if task.Name == taskName {
			execution.TaskExecutions[ix].State = taskState
		}
	}
	return store.Set(key, execution)
}

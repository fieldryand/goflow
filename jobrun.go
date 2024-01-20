package goflow

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/philippgille/gokv"
)

// Execution of a job.
type Execution struct {
	ID        string    `json:"id"`
	JobName   string    `json:"job"`
	StartedAt string    `json:"submitted"`
	State     state     `json:"state"`
	TaskRuns  []taskRun `json:"tasks"`
}

type executionIndex struct {
	ExecutionIDs []string `json:"executions"`
}

type taskRun struct {
	Name  string `json:"name"`
	State state  `json:"state"`
}

func (j *Job) newExecution() *Execution {
	taskRuns := make([]taskRun, 0)
	for _, task := range j.Tasks {
		taskrun := taskRun{task.Name, none}
		taskRuns = append(taskRuns, taskrun)
	}
	return &Execution{
		ID:        uuid.New().String(),
		JobName:   j.Name,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		State:     none,
		TaskRuns:  taskRuns}
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
	for ix, task := range execution.TaskRuns {
		if task.Name == taskName {
			execution.TaskRuns[ix].State = taskState
		}
	}
	return store.Set(key, execution)
}

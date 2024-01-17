package goflow

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/philippgille/gokv"
)

type jobRun struct {
	ID        string    `json:"id"`
	JobName   string    `json:"job"`
	StartedAt string    `json:"submitted"`
	State     state     `json:"state"`
	TaskRuns  []taskRun `json:"tasks"`
}

type jobRunIndex struct {
	JobRunIDs []string `json:"jobRuns"`
}

type taskRun struct {
	Name  string `json:"name"`
	State state  `json:"state"`
}

func (j *Job) newJobRun() *jobRun {
	taskRuns := make([]taskRun, 0)
	for taskName := range j.Tasks {
		taskrun := taskRun{taskName, none}
		taskRuns = append(taskRuns, taskrun)
	}
	return &jobRun{
		ID:        uuid.New().String(),
		JobName:   j.Name,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		State:     none,
		TaskRuns:  taskRuns}
}

// Persist a new jobrun.
func persistNewJobRun(store gokv.Store, jobrun *jobRun) error {
	key := jobrun.ID
	err := store.Set(key, jobrun)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// Index the job runs
func indexJobRuns(store gokv.Store, jobrun *jobRun) error {
	index := jobRunIndex{}
	store.Get(jobrun.JobName, &index)

	// add the jobrun ID to the index
	index.JobRunIDs = append(index.JobRunIDs, jobrun.ID)
	return store.Set(jobrun.JobName, index)
}

// Read all the persisted jobruns for a given job.
func readJobRuns(store gokv.Store, jobName string) ([]*jobRun, error) {
	index := jobRunIndex{}
	store.Get(jobName, &index)

	jobRuns := make([]*jobRun, 0)
	for _, key := range index.JobRunIDs {
		value := jobRun{}
		store.Get(key, &value)
		jobRuns = append(jobRuns, &value)
	}

	return jobRuns, nil
}

// Sync the current jobstate to the persisted jobrun.
func updateJobState(store gokv.Store, jobrun *jobRun, jobState state) error {
	key := jobrun.ID
	jobrun.State = jobState
	return store.Set(key, jobrun)
}

// Sync the current taskstate to the persisted jobrun.
func updateTaskState(store gokv.Store, jobrun *jobRun, taskName string, taskState state) error {
	key := jobrun.ID
	for ix, task := range jobrun.TaskRuns {
		if task.Name == taskName {
			jobrun.TaskRuns[ix].State = taskState
		}
	}
	return store.Set(key, jobrun)
}

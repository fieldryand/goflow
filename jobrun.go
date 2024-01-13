package goflow

import (
	"fmt"
	"time"

	"github.com/philippgille/gokv"
)

type jobRun struct {
	JobName   string    `json:"job"`
	StartedAt string    `json:"submitted"`
	JobState  *jobState `json:"state"`
}

type jobRunIndex struct {
	JobRunIDs []string
}

func (j *Job) newJobRun() *jobRun {
	return &jobRun{
		JobName:   j.Name,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		JobState:  j.jobState}
}

// Persist a new jobrun.
func persistNewJobRun(store gokv.Store, jobrun *jobRun) error {
	key := jobrun.JobName + "/" + jobrun.StartedAt
	err := store.Set(key, jobrun)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// Index the job runs
func indexJobRuns(store gokv.Store, jobrun *jobRun) error {

	index := jobRunIndex{}

	// Skip error check for same reason as above.
	// TODO: guarantee JobName is not "".
	store.Get(jobrun.JobName+"/index", &index)

	// add the jobrun ID to the index
	index.JobRunIDs = append(index.JobRunIDs, jobrun.StartedAt)
	return store.Set(jobrun.JobName+"/index", index)
}

// Read all the persisted jobruns for a given job.
func readJobRuns(store gokv.Store, jobName string) ([]*jobRun, error) {

	index := jobRunIndex{}

	// Skip error check for same reason as above.
	// TODO: guarantee JobName is not "".
	store.Get(jobName+"/index", &index)

	jobRuns := make([]*jobRun, 0)
	for _, key := range index.JobRunIDs {
		value := jobRun{}
		store.Get(jobName+"/"+key, &value)
		jobRuns = append(jobRuns, &value)
	}

	return jobRuns, nil
}

// Sync the current jobstate to the persisted jobrun.
func updateJobState(store gokv.Store, jobrun *jobRun, jobstate *jobState) error {

	// Get the key
	key := jobrun.JobName + "/" + jobrun.StartedAt

	// Get the lock
	jobstate.TaskState.RLock()

	// Update the jobrun state
	jobrun.JobState = jobstate

	// Persist it
	err := store.Set(key, jobrun)

	// Release lock
	jobstate.TaskState.RUnlock()

	return err
}

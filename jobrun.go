package goflow

import (
	"strconv"
	"time"

	"github.com/philippgille/gokv"
)

type jobRun struct {
	ID        int       `json:"id"`
	JobName   string    `json:"job"`
	StartedAt string    `json:"submitted"`
	JobState  *jobState `json:"state"`
}

type nextID struct{ ID int }

type jobRunIndex struct {
	JobRunIDs []int
}

func (j *Job) newJobRun() *jobRun {
	return &jobRun{
		JobName:   j.Name,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		JobState:  j.jobState}
}

// Persist a new jobrun.
func persistNewJobRun(store gokv.Store, jobrun *jobRun) error {

	jobRunID := nextID{}

	// Get the next available key. No need to check for errors,
	// because store.Get(k, v) only returns an error if k == ""
	// or v == nil.
	store.Get("nextID", &jobRunID)

	// Assign that key to the jobrun as its ID
	jobrun.ID = jobRunID.ID

	// Increment the next available key. Skip error check for
	// same reason as above.
	increment := jobRunID
	increment.ID++
	store.Set("nextID", increment)

	// Persist the jobrun
	key := strconv.Itoa(jobRunID.ID)
	return store.Set(key, jobrun)
}

// Index the job runs
func indexJobRuns(store gokv.Store, jobrun *jobRun) error {

	index := jobRunIndex{}

	// Skip error check for same reason as above.
	// TODO: guarantee JobName is not "".
	store.Get(jobrun.JobName, &index)

	// add the jobrun ID to the index
	index.JobRunIDs = append(index.JobRunIDs, jobrun.ID)
	return store.Set(jobrun.JobName, index)
}

// Read all the persisted jobruns for a given job.
func readJobRuns(store gokv.Store, jobName string) ([]*jobRun, error) {

	index := jobRunIndex{}

	// Skip error check for same reason as above.
	// TODO: guarantee JobName is not "".
	store.Get(jobName, &index)

	jobRuns := make([]*jobRun, 0)
	for _, i := range index.JobRunIDs {
		value := jobRun{}
		key := strconv.Itoa(i)
		store.Get(key, &value)
		jobRuns = append(jobRuns, &value)
	}

	return jobRuns, nil
}

// Sync the current jobstate to the persisted jobrun.
func updateJobState(store gokv.Store, jobrun *jobRun, jobstate *jobState) error {

	// Get the key
	key := strconv.Itoa(jobrun.ID)

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

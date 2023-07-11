package goflow

import (
	"fmt"
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

	// get the next available key
	jobRunID := nextID{}
	_, err := store.Get("nextID", &jobRunID)
	if err != nil {
		return fmt.Errorf("storage error: %v", err)
	}

	// assign that key to the jobrun as its ID
	jobrun.ID = jobRunID.ID

	// increment the next available key
	increment := jobRunID
	increment.ID++
	err = store.Set("nextID", increment)
	if err != nil {
		return fmt.Errorf("storage error: %v", err)
	}

	// persist the jobrun
	key := strconv.Itoa(jobRunID.ID)
	return store.Set(key, jobrun)
}

// Index the job runs
func indexJobRuns(store gokv.Store, jobrun *jobRun) error {

	// get the index
	index := jobRunIndex{}
	_, err := store.Get(jobrun.JobName, &index)
	if err != nil {
		return fmt.Errorf("storage error: %v", err)
	}

	// add the jobrun ID to the index
	index.JobRunIDs = append(index.JobRunIDs, jobrun.ID)
	return store.Set(jobrun.JobName, index)
}

// Read all the persisted jobruns for a given job.
func readJobRuns(store gokv.Store, jobName string) ([]*jobRun, error) {
	index := jobRunIndex{}
	_, err := store.Get(jobName, &index)
	if err != nil {
		return nil, fmt.Errorf("storage error: %v", err)
	}

	jobRuns := make([]*jobRun, 0)
	for _, i := range index.JobRunIDs {
		value := jobRun{}
		key := strconv.Itoa(i)
		_, err := store.Get(key, &value)
		if err != nil {
			return nil, fmt.Errorf("storage error: %v", err)
		}
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

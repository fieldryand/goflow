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

// Read all the persisted jobruns for a given job.
func readJobRuns(store gokv.Store, jobName string) ([]*jobRun, error) {
	jobRuns := make([]*jobRun, 0)
	index := 1
	for {
		value := jobRun{}
		key := strconv.Itoa(index)
		found, err := store.Get(key, &value)
		if err != nil {
			panic(err)
		}
		if !found {
			break
		} else if value.JobName == jobName {
			jobRuns = append(jobRuns, &value)
		}
		index++
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

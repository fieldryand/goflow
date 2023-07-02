package goflow

import (
	"log"
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

func (j *Job) newJobRun() *jobRun {
	return &jobRun{
		JobName:   j.Name,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		JobState:  j.jobState}
}

// Persist a new jobrun.
func persistNewJobRun(store gokv.Store, jobrun *jobRun) error {
	// find the next available key
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
		}
		index++
	}
	// assign that key to the jobrun as its ID
	jobrun.ID = index
	key := strconv.Itoa(index)
	// persist it
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
		} else {
			jobRuns = append(jobRuns, &value)
		}
		index++
	}
	return jobRuns, nil
}

// Sync the current jobstate to the persisted jobrun.
func updateJobState(store gokv.Store, jobrun *jobRun, jobstate *jobState) error {
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
		} else if index == jobrun.ID {
			// when we find the jobrun's key, set it to its current value
			// first need to obtain the lock
			jobstate.TaskState.RLock()
			jobrun.JobState = jobstate
			err := store.Set(key, jobrun)
			if err != nil {
				log.Panicf("error: %v", err)
			}
			jobstate.TaskState.RUnlock()
		}
		index++
	}
	return nil
}

package goflow

import (
	"log"
	"strconv"

	"github.com/philippgille/gokv"
)

func writeJobRun(store gokv.Store, jobrun *jobRun) error {
	key := strconv.Itoa(jobrun.ID)
	return store.Set(key, jobrun)
}

func readJobRuns(store gokv.Store, jobName string) (*jobRunList, error) {
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
	jobRunList := newJobRunList(jobName, jobRuns)
	return jobRunList, nil
}

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

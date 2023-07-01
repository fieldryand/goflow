package goflow

import (
	"log"
	"strconv"

	"github.com/philippgille/gokv"
)

type database interface {
	writeJobRun(*jobRun) error
	readJobRuns(string) (*jobRunList, error)
	updateJobState(*jobRun, *jobState) error
}

type genericStore struct{ gokv.Store }

func newStore(store gokv.Store) genericStore {
	return genericStore{store}
}

func (store genericStore) writeJobRun(jobrun *jobRun) error {
	key := strconv.Itoa(jobrun.ID)
	return store.Set(key, jobrun)
}

func (store genericStore) readJobRuns(jobName string) (*jobRunList, error) {
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

func (store genericStore) updateJobState(jr *jobRun, js *jobState) error {
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
		} else if index == jr.ID {
			// when we find the jobrun's key, set it to its current value
			// first need to obtain the lock
			js.TaskState.RLock()
			err := store.Set(key, js)
			if err != nil {
				log.Panicf("error: %v", err)
			}
			js.TaskState.RUnlock()
		}
		index++
	}
	return nil
}

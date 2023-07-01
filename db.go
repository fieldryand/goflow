package goflow

import (
	"encoding/binary"
	"encoding/json"
	"log"

	bolt "go.etcd.io/bbolt"
)

type database interface {
	writeJobRun(*jobRun) error
	readJobRuns(string) (*jobRunList, error)
	updateJobState(*jobRun, *jobState) error
}

type boltDB struct{ *bolt.DB }

var jobRunBucket string = "jobRuns"

func (db *boltDB) writeJobRun(jobrun *jobRun) error {
	value, _ := json.Marshal(jobrun)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(jobRunBucket))
		key := make([]byte, 4)
		binary.BigEndian.PutUint32(key, uint32(jobrun.ID))
		err := b.Put(key, value)
		return err
	})

	return err
}

func (db *boltDB) readJobRuns(jobName string) (*jobRunList, error) {
	jobRuns := make([]*jobRun, 0)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(jobRunBucket))
		index := 1
		for {
			key := make([]byte, 4)
			binary.BigEndian.PutUint32(key, uint32(index))
			v := b.Get(key)
			if v == nil {
				break
			}
			value := jobRun{}
			_ = json.Unmarshal(v, &value)
			jobRuns = append(jobRuns, &value)
			index++
		}
		return nil
	})

	jobRunList := newJobRunList(jobName, jobRuns)

	return jobRunList, err
}

func (db *boltDB) updateJobState(jr *jobRun, js *jobState) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(jobRunBucket))
		index := 1
		for {
			key := make([]byte, 4)
			binary.BigEndian.PutUint32(key, uint32(index))
			value := b.Get(key)
			if value == nil {
				break
			}
			if index == jr.ID {
				jobrun := &jobRun{}
				_ = json.Unmarshal(value, jobrun)
				updatedJobRun, _ := marshalJobRun(jobrun, js)
				err := b.Put(key, updatedJobRun)
				if err != nil {
					log.Panicf("error: %v", err)
				}
			}
			index++
		}
		return nil
	})

	return err
}

type memoryDB struct{ jobRuns []*jobRun }

func (db *memoryDB) writeJobRun(jr *jobRun) error {
	db.jobRuns = append(db.jobRuns, jr)
	return nil
}

func (db *memoryDB) readJobRuns(jobName string) (*jobRunList, error) {
	jobRunList := newJobRunList(jobName, db.jobRuns)
	return jobRunList, nil
}

func (db *memoryDB) updateJobState(jr *jobRun, js *jobState) error {
	for _, jobRun := range db.jobRuns {
		if jobRun.name() == jr.name() {
			jobRun.JobState.Update(js)
		}
	}
	return nil
}

package goflow

import (
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
)

type database interface {
	writeJobRun(*jobRun) error
	readJobRuns(string) (*jobRunList, error)
	updateJobState(*jobRun, *jobState) error
}

type boltDB struct{ *bolt.DB }

var jobRunBucket string = "jobRuns"

func (db *boltDB) writeJobRun(jr *jobRun) error {
	jrMarshalled, _ := json.Marshal(jr)

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(jobRunBucket))
		err := b.Put([]byte(jr.name()), jrMarshalled)
		return err
	})

	return err
}

func (db *boltDB) readJobRuns(jobName string) (*jobRunList, error) {
	jobRuns := make([]*jobRun, 0)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(jobRunBucket))
		cursor := b.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			j := jobRun{}
			_ = json.Unmarshal(v, &j)
			jobRuns = append(jobRuns, &j)
		}
		return nil
	})

	jobRunList := newJobRunList(jobName, jobRuns)

	return jobRunList, err
}

func (db *boltDB) updateJobState(jr *jobRun, js *jobState) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(jobRunBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(k) == jr.name() {
				j := jobRun{}
				_ = json.Unmarshal(v, &j)
				j.JobState = js
				updatedJobRun, _ := json.Marshal(j)
				err := b.Put([]byte(jr.name()), updatedJobRun)
				if err != nil {
					log.Panicf("error: %v", err)
				}
			}
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
			jr.JobState = js
		}
	}
	return nil
}

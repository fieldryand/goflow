package goflow

import (
	"fmt"
	"time"
)

type jobRun struct {
	JobName   string    `json:"jobName"`
	StartedAt string    `json:"startedAt"`
	JobState  *jobState `json:"jobState"`
}

func newJobRun(name string) *jobRun {
	return &jobRun{
		JobName:   name,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		JobState:  newJobState()}
}

func (j *jobRun) name() string {
	return fmt.Sprintf("%s_%s", j.JobName, j.StartedAt)
}

type jobRunList struct {
	JobName string    `json:"jobName"`
	JobRuns []*jobRun `json:"jobRuns"`
}

func newJobRunList(name string, jobRuns []*jobRun) *jobRunList {
	list := make([]*jobRun, 0)

	for _, jr := range jobRuns {
		if jr.JobName == name {
			list = append(list, jr)
		}
	}

	return &jobRunList{JobName: name, JobRuns: list}
}

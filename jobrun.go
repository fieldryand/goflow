package goflow

import (
	"encoding/json"
	"fmt"
	"time"
)

type jobRun struct {
	JobName   string    `json:"jobName"`
	StartedAt string    `json:"startedAt"`
	JobState  *jobState `json:"jobState"`
}

func (j *Job) newJobRun() *jobRun {
	return &jobRun{
		JobName:   j.Name,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		JobState:  j.jobState}
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

func marshalJobRunList(jrl *jobRunList) ([]byte, error) {
	for _, jobRun := range jrl.JobRuns {
		jobRun.JobState.TaskState.RLock()
	}

	result, ok := json.Marshal(jrl)

	for _, jobRun := range jrl.JobRuns {
		jobRun.JobState.TaskState.RUnlock()
	}

	return result, ok
}

func marshalJobRun(jr *jobRun, js *jobState) ([]byte, error) {
	js.TaskState.RLock()
	jr.JobState = js
	result, ok := json.Marshal(jr)
	js.TaskState.RUnlock()
	return result, ok
}

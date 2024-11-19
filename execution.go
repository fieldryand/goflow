package goflow

import (
	"time"

	"github.com/google/uuid"
	"github.com/philippgille/gokv"
)

type execution struct {
	ID         uuid.UUID       `json:"id"`
	Job        string          `json:"job"`
	StartTs    time.Time       `json:"startTs"`
	ModifiedTs time.Time       `json:"modifiedTs"`
	State      state           `json:"state"`
	Tasks      []taskExecution `json:"tasks"`
}

type taskExecution struct {
	Name    string    `json:"name"`
	State   state     `json:"state"`
	StartTs time.Time `json:"startTs"`
}

func (j *Job) newExecution() *execution {
	taskExecutions := make([]taskExecution, 0)
	for _, task := range j.Tasks {
		taskrun := taskExecution{
			Name:    task.Name,
			State:   none,
			StartTs: time.Time{}}
		taskExecutions = append(taskExecutions, taskrun)
	}
	return &execution{
		ID:         uuid.New(),
		Job:        j.Name,
		StartTs:    time.Now().UTC(),
		ModifiedTs: time.Now().UTC(),
		State:      none,
		Tasks:      taskExecutions}
}

// Persist a new execution.
func persistNewExecution(s gokv.Store, e *execution) error {
	key := e.ID
	return s.Set(key.String(), e)
}

type executionIndex struct {
	ExecutionIDs []string `json:"executions"`
}

func timeToDatestring(t time.Time) string {
	return t.Format("2006-01-02")
}

// The store contains key-value pairs such as
// "2024-10-13": [{job-id-0}, {job-id-1}, ...].
// This function updates such a pair.
func indexExecutions(s gokv.Store, e *execution) error {
	date := timeToDatestring(e.StartTs)
	i := executionIndex{}
	_, err := s.Get(date, &i)
	if err != nil {
		return err
	}
	i.ExecutionIDs = append(i.ExecutionIDs, e.ID.String())
	return s.Set(date, i)
}

// Read all the persisted executions on or after a given date.
func readExecutions(s gokv.Store, d time.Time) ([]*execution, error) {

	i := executionIndex{}
	executions := make([]*execution, 0)

	for {

		date := timeToDatestring(d)
		found, _ := s.Get(date, &i)

		if found {
			for _, key := range i.ExecutionIDs {
				val := execution{}
				_, err := s.Get(key, &val)
				if err != nil {
					return nil, err
				}
				executions = append(executions, &val)
			}
		}

		d = d.AddDate(0, 0, 1)

		if d.After(time.Now()) {
			break
		}

	}

	return executions, nil
}

func (e *execution) setTaskState(task string, s state) {
	for ix, t := range e.Tasks {
		if t.Name == task {
			e.Tasks[ix].State = s
		}
	}
}

func (e *execution) setTaskStartTs(task string) {
	for ix, t := range e.Tasks {
		if t.Name == task {
			e.Tasks[ix].StartTs = time.Now().UTC()
		}
	}
}

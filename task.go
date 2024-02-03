package goflow

import (
	"math"
	"time"
)

// A Task is the unit of work that makes up a job. Whenever a task is executed, it
// calls its associated operator.
type Task struct {
	Name        string
	Operator    Operator
	TriggerRule triggerRule
	Retries     int
	RetryDelay  RetryDelay
	remaining   int
	state       state
}

type triggerRule string

const (
	allDone       triggerRule = "allDone"
	allSuccessful triggerRule = "allSuccessful"
)

func (t *Task) run(writes chan writeOp) error {

	_, err := t.Operator.Run()

	// retry
	if err != nil && t.remaining > 0 {
		writes <- writeOp{t.Name, upForRetry}
		return nil
	}

	// failed
	if err != nil && t.remaining <= 0 {
		writes <- writeOp{t.Name, failed}
		return err
	}

	// success
	writes <- writeOp{t.Name, successful}
	return nil
}

func (t *Task) skip(writes chan writeOp) error {
	writes <- writeOp{t.Name, skipped}
	return nil
}

// RetryDelay is a type that implements a Wait() method, which is called in between
// task retry attempts.
type RetryDelay interface {
	wait(taskName string, attempt int)
}

// ConstantDelay waits a constant number of seconds between task retries.
type ConstantDelay struct{ Period int }

func (d ConstantDelay) wait(task string, attempt int) {
	time.Sleep(time.Duration(d.Period) * time.Second)
}

// ExponentialBackoff waits exponentially longer between each retry attempt.
type ExponentialBackoff struct{}

func (d ExponentialBackoff) wait(task string, attempt int) {
	delay := math.Pow(2, float64(attempt))
	time.Sleep(time.Duration(delay) * time.Second)
}

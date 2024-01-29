package goflow

import (
	"log"
	"math"
	"time"
)

// A Task is the unit of work that makes up a job. Whenever a task is executed, it
// calls its associated operator.
type Task struct {
	Name         string
	Operator     Operator
	PipeOperator PipeOperator
	TriggerRule  triggerRule
	Retries      int
	RetryDelay   RetryDelay
	attempts     int // attempts remaining
	state        state
}

type triggerRule string

const (
	allDone       triggerRule = "allDone"
	allSuccessful triggerRule = "allSuccessful"
)

func (t *Task) log(s state, res interface{}) {
	msg := "task update: name=%v, state=%v, remainingattempts=%v, result=%v"
	log.Printf(msg, t.Name, s, t.attempts, res)
}

func (t *Task) run(p pipe, writes chan writeOp) error {

	log.Printf("starting task: name=%v", t.Name)

	res, err := t.Operator.Run()

	// retry
	if err != nil && t.attempts > 0 {
		t.log(upForRetry, err)
		writes <- writeOp{t.Name, upForRetry, p}
		return nil
	}

	// failed
	if err != nil && t.attempts <= 0 {
		t.log(failed, err)
		writes <- writeOp{t.Name, failed, p}
		return err
	}

	// success
	t.log(successful, res)
	writes <- writeOp{t.Name, successful, p}
	return nil

}

func (t *Task) runWithPipe(p pipe, writes chan writeOp) error {

	log.Printf("starting task: name=%v", t.Name)

	res, err := t.PipeOperator.RunWithPipe(p)

	// retry
	if err != nil && t.attempts > 0 {
		t.log(upForRetry, err)
		writes <- writeOp{t.Name, upForRetry, res}
		return nil
	}

	// failed
	if err != nil && t.attempts <= 0 {
		t.log(failed, err)
		writes <- writeOp{t.Name, failed, res}
		return err
	}

	// success
	t.log(successful, res)
	writes <- writeOp{t.Name, successful, res}
	return nil

}

func (t *Task) skip(p pipe, writes chan writeOp) error {
	t.log(skipped, nil)
	writes <- writeOp{t.Name, skipped, p}
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
	log.Printf("task update: name=%v, secondsuntilretry=%v", task, d.Period)
	time.Sleep(time.Duration(d.Period) * time.Second)
}

// ExponentialBackoff waits exponentially longer between each retry attempt.
type ExponentialBackoff struct{}

func (d ExponentialBackoff) wait(task string, attempt int) {
	delay := math.Pow(2, float64(attempt))
	log.Printf("task update: name=%v, secondsuntilretry=%v", task, delay)
	time.Sleep(time.Duration(delay) * time.Second)
}

package goflow

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

var reads = make(chan readOp)

func TestJob(t *testing.T) {
	j := NewJob("example", JobParams{})

	j.AddTask(
		"addOneOne",
		NewAddition(1, 1),
		TaskParams{},
	)
	j.AddTask(
		"sleepTwo",
		BashOp("sleep", "2"),
		TaskParams{},
	)
	j.AddTask(
		"addTwoFour",
		BashOp("sh", "-c", "echo $((2 + 4))"),
		TaskParams{},
	)
	j.AddTask(
		"addThreeFour",
		NewAddition(3, 4),
		TaskParams{},
	)
	j.AddTask(
		"whoopsWithConstantDelay",
		BashOp("whoops"),
		TaskParams{
			Retries:    5,
			RetryDelay: ConstantDelay(1),
		},
	)
	j.AddTask(
		"whoopsWithExponentialBackoff",
		BashOp("whoops"),
		TaskParams{
			Retries:    1,
			RetryDelay: ExponentialBackoff(),
		},
	)
	j.AddTask(
		"totallySkippable",
		BashOp("sh", "-c", "echo 'everything succeeded'"),
		TaskParams{
			TriggerRule: "allSuccessful",
		},
	)
	j.AddTask(
		"cleanUp",
		BashOp("sh", "-c", "echo 'cleaning up now'"),
		TaskParams{
			TriggerRule: "allDone",
		},
	)

	j.SetDownstream(j.Task("addOneOne"), j.Task("sleepTwo"))
	j.SetDownstream(j.Task("sleepTwo"), j.Task("addTwoFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("addThreeFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("whoopsWithConstantDelay"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("whoopsWithExponentialBackoff"))
	j.SetDownstream(j.Task("whoopsWithConstantDelay"), j.Task("totallySkippable"))
	j.SetDownstream(j.Task("whoopsWithExponentialBackoff"), j.Task("totallySkippable"))
	j.SetDownstream(j.Task("totallySkippable"), j.Task("cleanUp"))

	go j.run(reads)
	go func() {
		read := readOp{resp: make(chan *jobState), allDone: j.allDone()}
		reads <- read
		<-read.resp
	}()

	time.Sleep(time.Duration(7) * time.Second)

	expectedState := map[string]state{
		"addOneOne":                    successful,
		"sleepTwo":                     successful,
		"addTwoFour":                   successful,
		"addThreeFour":                 successful,
		"whoopsWithConstantDelay":      failed,
		"whoopsWithExponentialBackoff": failed,
		"totallySkippable":             skipped,
		"cleanUp":                      successful,
	}

	if !reflect.DeepEqual(j.jobState.TaskState, expectedState) {
		t.Errorf("Got status %v, expected %v", j.jobState.TaskState, expectedState)
	}
}

func TestCyclicJob(t *testing.T) {
	j := NewJob("cyclic", JobParams{})

	j.AddTask("addTwoTwo", NewAddition(2, 2), TaskParams{})
	j.AddTask("addFourFour", NewAddition(4, 4), TaskParams{})
	j.SetDownstream(j.Task("addTwoTwo"), j.Task("addFourFour"))
	j.SetDownstream(j.Task("addFourFour"), j.Task("addTwoTwo"))

	j.run(reads)
}

func TestTaskFailure(t *testing.T) {
	j := NewJob("with bad task", JobParams{})
	j.AddTask("badTask", NewAddition(-1, -1), TaskParams{})

	go j.run(reads)
	go func() {
		read := readOp{resp: make(chan *jobState), allDone: j.allDone()}
		reads <- read
		<-read.resp
	}()

	time.Sleep(time.Duration(1) * time.Second)

	if j.jobState.State != failed {
		t.Errorf("Got status %v, expected %v", j.jobState.State, failed)
	}
}

// Adds two nonnegative numbers.
type Addition struct{ a, b int }

func NewAddition(a, b int) *Addition {
	o := Addition{a, b}
	return &o
}

func (o Addition) Run() (interface{}, error) {

	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}

	result := o.a + o.b
	return result, nil
}

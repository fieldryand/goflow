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

	j.Add(&Task{
		Name:     "addOneOne",
		Operator: Addition{1, 1},
	})
	j.Add(&Task{
		Name:     "sleepTwo",
		Operator: Bash{Cmd: "sleep", Args: []string{"2"}},
	})
	j.Add(&Task{
		Name:     "addTwoFour",
		Operator: Bash{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})
	j.Add(&Task{
		Name:     "addThreeFour",
		Operator: Addition{3, 4},
	})
	j.Add(&Task{
		Name:       "whoopsWithConstantDelay",
		Operator:   Bash{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: ConstantDelay{1},
	})
	j.Add(&Task{
		Name:       "whoopsWithExponentialBackoff",
		Operator:   Bash{Cmd: "whoops", Args: []string{}},
		Retries:    1,
		RetryDelay: ExponentialBackoff{},
	})
	j.Add(&Task{
		Name:        "totallySkippable",
		Operator:    Bash{Cmd: "sh", Args: []string{"-c", "echo 'everything succeeded'"}},
		TriggerRule: "allSuccessful",
	})
	j.Add(&Task{
		Name:        "cleanUp",
		Operator:    Bash{Cmd: "sh", Args: []string{"-c", "echo 'cleaning up now'"}},
		TriggerRule: "allDone",
	})

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

	j.Add(&Task{Name: "addTwoTwo", Operator: Addition{2, 2}})
	j.Add(&Task{Name: "addFourFour", Operator: Addition{4, 4}})
	j.SetDownstream(j.Task("addTwoTwo"), j.Task("addFourFour"))
	j.SetDownstream(j.Task("addFourFour"), j.Task("addTwoTwo"))

	j.run(reads)
}

func TestTaskFailure(t *testing.T) {
	j := NewJob("with bad task", JobParams{})
	j.Add(&Task{Name: "badTask", Operator: Addition{-1, -1}})

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

func (o Addition) Run() (interface{}, error) {

	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}

	result := o.a + o.b
	return result, nil
}

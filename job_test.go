package goflow

import (
	"errors"
	"reflect"
	"testing"
)

func TestJob(t *testing.T) {
	j := &Job{Name: "example", Schedule: "* * * * *"}

	j.Add(&Task{
		Name:     "addOneOne",
		Operator: Addition{1, 1},
	})
	j.Add(&Task{
		Name:     "sleepTwo",
		Operator: Command{Cmd: "sleep", Args: []string{"2"}},
	})
	j.Add(&Task{
		Name:     "addTwoFour",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})
	j.Add(&Task{
		Name:     "addThreeFour",
		Operator: Addition{3, 4},
	})
	j.Add(&Task{
		Name:       "whoopsWithConstantDelay",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: ConstantDelay{1},
	})
	j.Add(&Task{
		Name:       "whoopsWithExponentialBackoff",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    1,
		RetryDelay: ExponentialBackoff{},
	})
	j.Add(&Task{
		Name:        "totallySkippable",
		Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'everything succeeded'"}},
		TriggerRule: "allSuccessful",
	})
	j.Add(&Task{
		Name:        "cleanUp",
		Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'cleaning up now'"}},
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

	go j.run()
	func() {
		for {
			jobState := j.getJobState()
			if jobState.State != running && jobState.State != none {
				break
			}
		}
	}()

	expectedState := newStringStateMap()
	expectedState.Store("addOneOne", successful)
	expectedState.Store("sleepTwo", successful)
	expectedState.Store("addTwoFour", successful)
	expectedState.Store("addThreeFour", successful)
	expectedState.Store("whoopsWithConstantDelay", failed)
	expectedState.Store("whoopsWithExponentialBackoff", failed)
	expectedState.Store("totallySkippable", skipped)
	expectedState.Store("cleanUp", successful)

	if !reflect.DeepEqual(j.jobState.TaskState.Internal, expectedState.Internal) {
		t.Errorf("Got status %v, expected %v", j.jobState.TaskState.Internal, expectedState.Internal)
	}
}

func TestCyclicJob(t *testing.T) {
	j := &Job{Name: "cyclic", Schedule: "* * * * *"}

	j.Add(&Task{Name: "addTwoTwo", Operator: Addition{2, 2}})
	j.Add(&Task{Name: "addFourFour", Operator: Addition{4, 4}})
	j.SetDownstream(j.Task("addTwoTwo"), j.Task("addFourFour"))
	j.SetDownstream(j.Task("addFourFour"), j.Task("addTwoTwo"))

	j.run()
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

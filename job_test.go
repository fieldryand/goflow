package goflow

import (
	"errors"
	"reflect"
	"testing"

	"github.com/fieldryand/goflow/op"
)

var reads = make(chan readOp)

func TestJob(t *testing.T) {
	j := NewJob("example")

	j.AddTask("addOneOne", NewAddition(1, 1))
	j.AddTask("sleepTwo", op.Bash("sleep", "2"))
	j.AddTask("addTwoFour", op.Bash("sh", "-c", "echo $((2 + 4))"))
	j.AddTask("addThreeFour", NewAddition(3, 4))

	j.SetDownstream(j.Task("addOneOne"), j.Task("sleepTwo"))
	j.SetDownstream(j.Task("sleepTwo"), j.Task("addTwoFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("addThreeFour"))

	j.run(reads)

	expectedState := map[string]state{
		"addOneOne":    successful,
		"sleepTwo":     successful,
		"addTwoFour":   successful,
		"addThreeFour": successful,
	}

	if !reflect.DeepEqual(j.jobState.TaskState, expectedState) {
		t.Errorf("Got status %v, expected %v", j.jobState.TaskState, expectedState)
	}
}

func TestCyclicJob(t *testing.T) {
	j := NewJob("cyclic")

	j.AddTask("addTwoTwo", NewAddition(2, 2))
	j.AddTask("addFourFour", NewAddition(4, 4))
	j.SetDownstream(j.Task("addTwoTwo"), j.Task("addFourFour"))
	j.SetDownstream(j.Task("addFourFour"), j.Task("addTwoTwo"))

	j.run(reads)
}

func TestTaskFailure(t *testing.T) {
	j := NewJob("with bad task").AddTask("badTask", NewAddition(-1, -1))
	j.run(reads)

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

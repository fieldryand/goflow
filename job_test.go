package goflow

import (
	"errors"
	"reflect"
	"testing"

	"github.com/fieldryand/goflow/op"
)

var reads = make(chan readOp)

func TestJob(t *testing.T) {
	addOneOne := NewTask("addOneOne", NewAddition(1, 1))
	sleepTwo := NewTask("sleepTwo", op.Bash("sleep", "2"))
	addTwoFour := NewTask("addTwoFour", NewAddition(2, 4))
	addThreeFour := NewTask("addThreeFour", NewAddition(3, 4))

	j := NewJob("example").
		AddTask(addOneOne).
		AddTask(sleepTwo).
		AddTask(addTwoFour).
		AddTask(addThreeFour).
		SetDownstream(addOneOne, sleepTwo).
		SetDownstream(sleepTwo, addTwoFour).
		SetDownstream(addOneOne, addThreeFour)

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
	addTwoTwo := NewTask("addTwoTwo", NewAddition(2, 2))
	addFourFour := NewTask("addFourFour", NewAddition(4, 4))

	j := NewJob("cyclic").
		AddTask(addTwoTwo).
		AddTask(addFourFour).
		SetDownstream(addTwoTwo, addFourFour).
		SetDownstream(addFourFour, addTwoTwo)

	j.run(reads)
}

func TestTaskFailure(t *testing.T) {
	badTask := NewTask("badTask", NewAddition(-1, -1))

	j := NewJob("with bad task").
		AddTask(badTask)

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

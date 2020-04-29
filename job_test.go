package goflow

import (
	"errors"
	"reflect"
	"testing"

	"github.com/fieldryand/goflow/operator"
)

var reads = make(chan readOp)

func TestJob(t *testing.T) {
	add_1_1 := NewTask("add 1 1", NewAddition(1, 1))
	sleep_2 := NewTask("sleep 2", operator.NewSleep(2))
	add_2_4 := NewTask("add 2 4", NewAddition(2, 4))
	add_3_4 := NewTask("add 3 4", NewAddition(3, 4))

	j := NewJob("example").
		AddTask(add_1_1).
		AddTask(sleep_2).
		AddTask(add_2_4).
		AddTask(add_3_4).
		SetDownstream(add_1_1, sleep_2).
		SetDownstream(sleep_2, add_2_4).
		SetDownstream(add_1_1, add_3_4)

	j.run(reads)

	expectedState := map[string]string{
		"add 1 1": "Success",
		"sleep 2": "Success",
		"add 2 4": "Success",
		"add 3 4": "Success",
	}

	if !reflect.DeepEqual(j.TaskState, expectedState) {
		t.Errorf("Got status %v, expected %v", j.TaskState, expectedState)
	}
}

func TestCyclicJob(t *testing.T) {
	add_2_2 := NewTask("add 2 2", NewAddition(2, 2))
	add_4_4 := NewTask("add 4 4", NewAddition(4, 4))

	j := NewJob("cyclic").
		AddTask(add_2_2).
		AddTask(add_4_4).
		SetDownstream(add_2_2, add_4_4).
		SetDownstream(add_4_4, add_2_2)

	j.run(reads)
}

func TestTaskFailure(t *testing.T) {
	badTask := NewTask("add -1 -1", NewAddition(-1, -1))

	j := NewJob("with bad task").
		AddTask(badTask)

	err := j.run(reads)

	if err == nil {
		t.Errorf("Job returned nil, expected a jobError")
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

package goflow

import (
	"github.com/fieldryand/goflow/operators"
	"reflect"
	"testing"
)

var reads = make(chan readOp)

func TestJob(t *testing.T) {
	add_1_1 := Task("add 1 1", operators.AddOperator(1, 1))
	sleep_2 := Task("sleep 2", operators.SleepOperator(2))
	add_2_4 := Task("add 2 4", operators.AddOperator(2, 4))
	add_3_4 := Task("add 3 4", operators.AddOperator(3, 4))

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
	add_2_2 := Task("add 2 2", operators.AddOperator(2, 2))
	add_4_4 := Task("add 4 4", operators.AddOperator(4, 4))

	j := NewJob("cyclic").
		AddTask(add_2_2).
		AddTask(add_4_4).
		SetDownstream(add_2_2, add_4_4).
		SetDownstream(add_4_4, add_2_2)

	j.run(reads)
}

func TestTaskFailure(t *testing.T) {
	badTask := Task("add -1 -1", operators.AddOperator(-1, -1))

	j := NewJob("with bad task").
		AddTask(badTask)

	err := j.run(reads)

	if err == nil {
		t.Errorf("Job returned nil, expected a jobError")
	}

}
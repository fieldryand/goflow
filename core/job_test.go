package core

import (
	"reflect"
	"testing"
)

func TestJob(t *testing.T) {
	stat := make(chan string, 1)

	add_1_1 := Task("add 1 1", AddOperator(1, 1))
	sleep_2 := Task("sleep 2", SleepOperator(2))
	add_2_4 := Task("add 2 4", AddOperator(2, 4))
	add_3_4 := Task("add 3 4", AddOperator(3, 4))

	j := Job("example").
		AddTask(add_1_1).
		AddTask(sleep_2).
		AddTask(add_2_4).
		AddTask(add_3_4).
		SetDownstream(add_1_1, sleep_2).
		SetDownstream(sleep_2, add_2_4).
		SetDownstream(add_1_1, add_3_4)

	j.Run(stat)
	<-stat

	expectedStatus := map[string]string{
		"add 1 1": "Success",
		"sleep 2": "Success",
		"add 2 4": "Success",
		"add 3 4": "Success",
	}

	gotStatus := j.Status()

	if !reflect.DeepEqual(gotStatus, expectedStatus) {
		t.Errorf("Got status %v, expected %v", gotStatus, expectedStatus)
	}
}

func TestCyclicJob(t *testing.T) {
	stat := make(chan string, 1)

	add_2_2 := Task("add 2 2", AddOperator(2, 2))
	add_4_4 := Task("add 4 4", AddOperator(4, 4))

	j := Job("cyclic").
		AddTask(add_2_2).
		AddTask(add_4_4).
		SetDownstream(add_2_2, add_4_4).
		SetDownstream(add_4_4, add_2_2)

	j.Run(stat)
	<-stat
}

func TestTaskFailure(t *testing.T) {
	stat := make(chan string, 1)

	badTask := Task("add -1 -1", AddOperator(-1, -1))

	j := Job("with bad task").
		AddTask(badTask)

	err := j.Run(stat)

	<-stat

	if err == nil {
		t.Errorf("Job returned nil, expected %v", err)
	}

}

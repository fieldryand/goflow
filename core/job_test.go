package core

import (
	"testing"
)

func TestJob(t *testing.T) {
	add_1_1 := Task("add 1 1", AddOperator(1, 1))
	sleep_2 := Task("sleep 2", SleepOperator(2))
	add_2_4 := Task("add 2 4", AddOperator(2, 4))
	add_3_4 := Task("add 3 4", AddOperator(3, 4))

	stat := make(chan string, 1)
	j := Job("job 1").
		AddTask(add_1_1).
		AddTask(sleep_2).
		AddTask(add_2_4).
		AddTask(add_3_4).
		SetDownstream(add_1_1, sleep_2).
		SetDownstream(sleep_2, add_2_4).
		SetDownstream(add_1_1, add_3_4)

	j.Run(stat)
	<-stat

	add_2_2 := Task("add 2 2", AddOperator(2, 2))
	add_4_4 := Task("add 4 4", AddOperator(4, 4))

	b := Job("bad job").
		AddTask(add_2_2).
		AddTask(add_4_4).
		SetDownstream(add_2_2, add_4_4).
		SetDownstream(add_4_4, add_2_2)

	b.Run(stat)
	<-stat
}

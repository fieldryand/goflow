package core

import (
	"testing"
)

func TestJob(t *testing.T) {
	j := NewJob("job 1")

	add_1_1 := NewTask("add 1 1", NewAddOperator(1, 1))
	sleep_2 := NewTask("sleep 2", NewSleepOperator(2))
	add_2_4 := NewTask("add 2 4", NewAddOperator(2, 4))
	add_3_4 := NewTask("add 3 4", NewAddOperator(3, 4))

	j.addTask(add_1_1)
	j.addTask(sleep_2)
	j.addTask(add_2_4)
	j.addTask(add_3_4)

	j.setDownstream(add_1_1, sleep_2)
	j.setDownstream(sleep_2, add_2_4)
	j.setDownstream(add_1_1, add_3_4)

	j.run()

	b := NewJob("bad job")
	add_2_2 := NewTask("add 2 2", NewAddOperator(2, 2))
	add_4_4 := NewTask("add 4 4", NewAddOperator(4, 4))
	b.addTask(add_2_2)
	b.addTask(add_4_4)
	b.setDownstream(add_2_2, add_4_4)
	b.setDownstream(add_4_4, add_2_2)
	b.run()
}

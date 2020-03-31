package jobs

import "github.com/fieldryand/goflow/core"

// ExampleJob returns an example job.
// goflow: ExampleJob example
func ExampleJob() *core.Job {
	sleep_1 := core.Task("sleep 1", core.SleepOperator(1))
	add_1_1 := core.Task("add 1 1", core.AddOperator(1, 1))
	sleep_2 := core.Task("sleep 2", core.SleepOperator(2))
	add_2_4 := core.Task("add 2 4", core.AddOperator(2, 4))
	add_3_4 := core.Task("add 3 4", core.AddOperator(3, 4))

	j := core.NewJob("example").
		AddTask(sleep_1).
		AddTask(add_1_1).
		AddTask(sleep_2).
		AddTask(add_2_4).
		AddTask(add_3_4).
		SetDownstream(sleep_1, add_1_1).
		SetDownstream(add_1_1, sleep_2).
		SetDownstream(sleep_2, add_2_4).
		SetDownstream(add_1_1, add_3_4)

	return j
}

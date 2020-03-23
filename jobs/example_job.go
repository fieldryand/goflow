package jobs

import "github.com/fieldryand/goflow/core"

var add_1_1 = core.Task("add 1 1", core.AddOperator(1, 1))
var sleep_2 = core.Task("sleep 2", core.SleepOperator(2))
var add_2_4 = core.Task("add 2 4", core.AddOperator(2, 4))
var add_3_4 = core.Task("add 3 4", core.AddOperator(3, 4))

var ExampleJob = core.Job("example").
	AddTask(add_1_1).
	AddTask(sleep_2).
	AddTask(add_2_4).
	AddTask(add_3_4).
	SetDownstream(add_1_1, sleep_2).
	SetDownstream(sleep_2, add_2_4).
	SetDownstream(add_1_1, add_3_4)

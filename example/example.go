// An example program demonstrating the usage of goflow.
package main

import (
	"github.com/fieldryand/goflow"
	"github.com/fieldryand/goflow/operators"
)

func main() {
	jobs := map[string]*goflow.Job{"example": ExampleJob()}
	goflow.Start(jobs)
}

func ExampleJob() *goflow.Job {
	sleep_1 := goflow.Task("sleep 1", operators.SleepOperator(1))
	add_1_1 := goflow.Task("add 1 1", operators.AddOperator(1, 1))
	sleep_2 := goflow.Task("sleep 2", operators.SleepOperator(2))
	add_2_4 := goflow.Task("add 2 4", operators.AddOperator(2, 4))
	add_3_4 := goflow.Task("add 3 4", operators.AddOperator(3, 4))

	j := goflow.NewJob("example").
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

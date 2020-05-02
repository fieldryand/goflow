// An example program demonstrating the usage of goflow.
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
	"github.com/fieldryand/goflow/operator"
)

func main() {
	jobs := map[string](func() *goflow.Job){"example": ExampleJob}
	goflow := goflow.Goflow(jobs)
	goflow.Run(":8090")
}

func ExampleJob() *goflow.Job {
	sleep_1 := goflow.NewTask("sleep 1", operator.NewSleep(1))
	add_1_1 := goflow.NewTask("add 1 1", NewAddition(1, 1))
	sleep_2 := goflow.NewTask("sleep 2", operator.NewSleep(2))
	add_2_4 := goflow.NewTask("add 2 4", NewAddition(2, 4))
	add_3_4 := goflow.NewTask("add 3 4", NewAddition(3, 4))

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

// We can create custom operators by implementing the Run() method.

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

// An example program demonstrating the usage of goflow.
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
)

func main() {
	gf := goflow.NewEngine()

	gf.AddJob(complexAnalyticsJob)
	gf.AddJob(messedUpJob)
	gf.AddJob(customOperatorJob)

	gf.Use(goflow.DefaultLogger())

	gf.Run(":8100")
}

// Crunch some numbers
func complexAnalyticsJob() *goflow.Job {
	j := goflow.NewJob(
		"ComplexAnalytics",
		goflow.JobParams{},
	)

	j.AddTask(
		"sleepOne",
		goflow.BashOp("sleep", "1"),
		goflow.TaskParams{},
	)
	j.AddTask(
		"addOneOne",
		goflow.BashOp("sh", "-c", "echo $((1 + 1))"),
		goflow.TaskParams{},
	)
	j.AddTask(
		"sleepTwo",
		goflow.BashOp("sleep", "2"),
		goflow.TaskParams{},
	)
	j.AddTask(
		"addTwoFour",
		goflow.BashOp("sh", "-c", "echo $((2 + 4))"),
		goflow.TaskParams{},
	)
	j.AddTask(
		"addThreeFour",
		goflow.BashOp("sh", "-c", "echo $((3 + 4))"),
		goflow.TaskParams{},
	)

	j.SetDownstream(j.Task("sleepOne"), j.Task("addOneOne"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("sleepTwo"))
	j.SetDownstream(j.Task("sleepTwo"), j.Task("addTwoFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("addThreeFour"))

	return j
}

// A task that throws an error
func messedUpJob() *goflow.Job {
	j := goflow.NewJob("MessedUp", goflow.JobParams{})
	j.AddTask("whoops", goflow.BashOp("whoops"), goflow.TaskParams{})
	return j
}

// We can create custom operators by implementing the Run() method.

// Add two nonnegative numbers
type positiveAdditionOperator struct{ a, b int }

func positiveAddition(a, b int) *positiveAdditionOperator {
	o := positiveAdditionOperator{a, b}
	return &o
}

// Run implements the custom operation
func (o positiveAdditionOperator) Run() (interface{}, error) {
	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}
	result := o.a + o.b
	return result, nil
}

// Use our custom operation in a job
func customOperatorJob() *goflow.Job {
	j := goflow.NewJob("CustomOperator", goflow.JobParams{})
	j.AddTask("posAdd", positiveAddition(5, 6), goflow.TaskParams{})
	return j
}

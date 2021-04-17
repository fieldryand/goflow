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
	j := goflow.NewJob("ComplexAnalytics")

	j.AddTask("sleepOne", goflow.BashOp("sleep", "1"))
	j.AddTask("addOneOne", goflow.BashOp("sh", "-c", "echo $((1 + 1))"))
	j.AddTask("sleepTwo", goflow.BashOp("sleep", "2"))
	j.AddTask("addTwoFour", goflow.BashOp("sh", "-c", "echo $((2 + 4))"))
	j.AddTask("addThreeFour", goflow.BashOp("sh", "-c", "echo $((2 + 4))"))

	j.SetDownstream(j.Task("sleepOne"), j.Task("addOneOne"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("sleepTwo"))
	j.SetDownstream(j.Task("sleepTwo"), j.Task("addTwoFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("addThreeFour"))

	return j
}

// A task that throws an error
func messedUpJob() *goflow.Job {
	return goflow.NewJob("MessedUp").AddTask("whoops", goflow.BashOp("whoops"))
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
	j := goflow.NewJob("CustomOperator")
	j.AddTask("posAdd", positiveAddition(5, 6))
	return j
}

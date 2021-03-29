// An example program demonstrating the usage of goflow.
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
	"github.com/fieldryand/goflow/op"
)

func main() {
	gf := goflow.New(ExampleJobOne, ExampleJobTwo, ExampleJobThree)
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8100")
}

// ExampleJobOne returns a simple job consisting of calls to "sleep" and a
// custom Addition operator.
func ExampleJobOne() *goflow.Job {
	j := goflow.NewJob("exampleOne")

	j.AddTask("sleepOne", op.Bash("sleep", "1"))
	j.AddTask("addOneOne", NewAddition(1, 1))
	j.AddTask("sleepTwo", op.Bash("sleep", "2"))
	j.AddTask("addTwoFour", op.Bash("sh", "-c", "echo $((2 + 4))"))
	j.AddTask("addThreeFour", NewAddition(3, 4))

	j.SetDownstream(j.Task("sleepOne"), j.Task("addOneOne"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("sleepTwo"))
	j.SetDownstream(j.Task("sleepTwo"), j.Task("addTwoFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("addThreeFour"))

	return j
}

// ExampleJobTwo returns an even simpler job consisting of a single "sleep" task.
func ExampleJobTwo() *goflow.Job {
	return goflow.NewJob("exampleTwo").AddTask("sleepTen", op.Bash("sleep", "10"))
}

// ExampleJobThree returns a job with a task that throws an error.
func ExampleJobThree() *goflow.Job {
	return goflow.NewJob("exampleThree").AddTask("whoops", op.Bash("whoops"))
}

// We can create custom operators by implementing the Run() method.

// Addition is an operation that adds two nonnegative numbers.
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

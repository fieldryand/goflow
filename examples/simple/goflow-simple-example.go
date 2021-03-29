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
	sleepOne := goflow.NewTask("sleepOne", op.Bash("sleep", "1"))
	addOneOne := goflow.NewTask("addOneOne", NewAddition(1, 1))
	sleepTwo := goflow.NewTask("sleepTwo", op.Bash("sleep", "2"))
	addTwoFour := goflow.NewTask("addTwoFour", op.Bash("sh", "-c", "echo $((2 + 4))"))
	addThreeFour := goflow.NewTask("addThreeFour", NewAddition(3, 4))

	j := goflow.NewJob("exampleOne").
		AddTask(sleepOne).
		AddTask(addOneOne).
		AddTask(sleepTwo).
		AddTask(addTwoFour).
		AddTask(addThreeFour).
		SetDownstream(sleepOne, addOneOne).
		SetDownstream(addOneOne, sleepTwo).
		SetDownstream(sleepTwo, addTwoFour).
		SetDownstream(addOneOne, addThreeFour)

	return j
}

// ExampleJobTwo returns an even simpler job consisting of a single "sleep" task.
func ExampleJobTwo() *goflow.Job {
	sleepTen := goflow.NewTask("sleepTen", op.Bash("sleep", "10"))
	j := goflow.NewJob("exampleTwo").AddTask(sleepTen)
	return j
}

// ExampleJobThree returns a job with a task that throws an error.
func ExampleJobThree() *goflow.Job {
	badTask := goflow.NewTask("badTask", NewAddition(-10, 0))
	j := goflow.NewJob("exampleThree").AddTask(badTask)
	return j
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

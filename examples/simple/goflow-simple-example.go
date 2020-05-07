// An example program demonstrating the usage of goflow.
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
	"github.com/fieldryand/goflow/operator"
)

func main() {
	jobs := map[string](func() *goflow.Job){
		"exampleOne": ExampleJobOne,
		"exampleTwo": ExampleJobTwo,
	}
	goflow := goflow.Goflow(jobs)
	goflow.Run(":8090")
}

// Returns a simple job consisting of Addition and Sleep operators.
func ExampleJobOne() *goflow.Job {
	sleepOne := goflow.NewTask("sleepOne", operator.NewSleep(1))
	addOneOne := goflow.NewTask("addOneOne", NewAddition(1, 1))
	sleepTwo := goflow.NewTask("sleepTwo", operator.NewSleep(2))
	addTwoFour := goflow.NewTask("addTwoFour", NewAddition(2, 4))
	addThreeFour := goflow.NewTask("addThreeFour", NewAddition(3, 4))

	j := goflow.NewJob("example").
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

// Returns an even simpler job consisting of a single Sleep task.
func ExampleJobTwo() *goflow.Job {
	sleepTen := goflow.NewTask("sleepTen", operator.NewSleep(10))
	j := goflow.NewJob("example").AddTask(sleepTen)
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

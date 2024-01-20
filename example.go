package goflow

import (
	"errors"
)

// Crunch some numbers
func complexAnalyticsJob() *Job {
	j := &Job{
		Name:     "exampleComplexAnalytics",
		Schedule: "* * * * * *",
		Active:   false,
	}

	j.Add(&Task{
		Name:     "sleepOne",
		Operator: Command{Cmd: "sleep", Args: []string{"1"}},
	})
	j.Add(&Task{
		Name:     "addOneOne",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((1 + 1))"}},
	})
	j.Add(&Task{
		Name:     "sleepTwo",
		Operator: Command{Cmd: "sleep", Args: []string{"2"}},
	})
	j.Add(&Task{
		Name:     "addTwoFour",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})
	j.Add(&Task{
		Name:     "addThreeFour",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
	})
	j.Add(&Task{
		Name:       "whoopsWithConstantDelay",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: ConstantDelay{Period: 1},
	})
	j.Add(&Task{
		Name:       "whoopsWithExponentialBackoff",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    1,
		RetryDelay: ExponentialBackoff{},
	})
	j.Add(&Task{
		Name:        "totallySkippable",
		Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'everything succeeded'"}},
		TriggerRule: "allSuccessful",
	})
	j.Add(&Task{
		Name:        "cleanUp",
		Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'cleaning up now'"}},
		TriggerRule: "allDone",
	})

	j.SetDownstream("sleepOne", "addOneOne")
	j.SetDownstream("addOneOne", "sleepTwo")
	j.SetDownstream("sleepTwo", "addTwoFour")
	j.SetDownstream("addOneOne", "addThreeFour")
	j.SetDownstream("sleepOne", "whoopsWithConstantDelay")
	j.SetDownstream("sleepOne", "whoopsWithExponentialBackoff")
	j.SetDownstream("whoopsWithConstantDelay", "totallySkippable")
	j.SetDownstream("whoopsWithExponentialBackoff", "totallySkippable")
	j.SetDownstream("totallySkippable", "cleanUp")

	return j
}

// PositiveAddition adds two nonnegative numbers. This is just a contrived example to
// demonstrate the usage of custom operators.
type PositiveAddition struct{ a, b int }

// Run implements the custom operation.
func (o PositiveAddition) Run() (interface{}, error) {
	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}
	result := o.a + o.b
	return result, nil
}

// Use our custom operation in a job.
func customOperatorJob() *Job {
	j := &Job{Name: "exampleCustomOperator", Schedule: "* * * * * *", Active: true}
	j.Add(&Task{Name: "posAdd", Operator: PositiveAddition{5, 6}})
	return j
}

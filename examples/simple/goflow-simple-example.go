// An example program demonstrating the usage of goflow.
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
)

func main() {
	gf := goflow.New(goflow.Options{StreamJobRuns: true})

	gf.AddJob(complexAnalyticsJob)
	gf.AddJob(customOperatorJob)

	gf.Use(goflow.DefaultLogger())

	gf.Run(":8181")
}

// Crunch some numbers
func complexAnalyticsJob() *goflow.Job {
	j := &goflow.Job{
		Name:     "ComplexAnalytics",
		Schedule: "* * * * *",
	}

	j.Add(&goflow.Task{
		Name:     "sleepOne",
		Operator: goflow.Command{Cmd: "sleep", Args: []string{"1"}},
	})
	j.Add(&goflow.Task{
		Name:     "addOneOne",
		Operator: goflow.Command{Cmd: "sh", Args: []string{"-c", "echo $((1 + 1))"}},
	})
	j.Add(&goflow.Task{
		Name:     "sleepTwo",
		Operator: goflow.Command{Cmd: "sleep", Args: []string{"2"}},
	})
	j.Add(&goflow.Task{
		Name:     "addTwoFour",
		Operator: goflow.Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})
	j.Add(&goflow.Task{
		Name:     "addThreeFour",
		Operator: goflow.Command{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
	})
	j.Add(&goflow.Task{
		Name:       "whoopsWithConstantDelay",
		Operator:   goflow.Command{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: goflow.ConstantDelay{Period: 1},
	})
	j.Add(&goflow.Task{
		Name:       "whoopsWithExponentialBackoff",
		Operator:   goflow.Command{Cmd: "whoops", Args: []string{}},
		Retries:    1,
		RetryDelay: goflow.ExponentialBackoff{},
	})
	j.Add(&goflow.Task{
		Name:        "totallySkippable",
		Operator:    goflow.Command{Cmd: "sh", Args: []string{"-c", "echo 'everything succeeded'"}},
		TriggerRule: "allSuccessful",
	})
	j.Add(&goflow.Task{
		Name:        "cleanUp",
		Operator:    goflow.Command{Cmd: "sh", Args: []string{"-c", "echo 'cleaning up now'"}},
		TriggerRule: "allDone",
	})

	j.SetDownstream(j.Task("sleepOne"), j.Task("addOneOne"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("sleepTwo"))
	j.SetDownstream(j.Task("sleepTwo"), j.Task("addTwoFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("addThreeFour"))
	j.SetDownstream(j.Task("sleepOne"), j.Task("whoopsWithConstantDelay"))
	j.SetDownstream(j.Task("sleepOne"), j.Task("whoopsWithExponentialBackoff"))
	j.SetDownstream(j.Task("whoopsWithConstantDelay"), j.Task("totallySkippable"))
	j.SetDownstream(j.Task("whoopsWithExponentialBackoff"), j.Task("totallySkippable"))
	j.SetDownstream(j.Task("totallySkippable"), j.Task("cleanUp"))

	return j
}

// PositiveAddition adds two nonnegative numbers.
type PositiveAddition struct{ a, b int }

// Run implements the custom operation
func (o PositiveAddition) Run() (interface{}, error) {
	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}
	result := o.a + o.b
	return result, nil
}

// Use our custom operation in a job.
func customOperatorJob() *goflow.Job {
	j := &goflow.Job{Name: "CustomOperator", Schedule: "* * * * *"}
	j.Add(&goflow.Task{Name: "posAdd", Operator: PositiveAddition{5, 6}})
	return j
}

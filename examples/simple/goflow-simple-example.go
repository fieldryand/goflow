// An example program demonstrating the usage of goflow.
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
)

func main() {
	gf := goflow.NewEngine()

	gf.AddJob(complexAnalyticsJob)
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
		goflow.Bash{Cmd: "sleep", Args: []string{"1"}},
		goflow.TaskParams{},
	)
	j.AddTask(
		"addOneOne",
		goflow.Bash{Cmd: "sh", Args: []string{"-c", "echo $((1 + 1))"}},
		goflow.TaskParams{},
	)
	j.AddTask(
		"sleepTwo",
		goflow.Bash{Cmd: "sleep", Args: []string{"2"}},
		goflow.TaskParams{},
	)
	j.AddTask(
		"addTwoFour",
		goflow.Bash{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
		goflow.TaskParams{},
	)
	j.AddTask(
		"addThreeFour",
		goflow.Bash{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
		goflow.TaskParams{},
	)
	j.AddTask(
		"whoopsWithConstantDelay",
		goflow.Bash{Cmd: "whoops", Args: []string{}},
		goflow.TaskParams{
			Retries:    5,
			RetryDelay: &goflow.ConstantDelay{Period: 1},
		},
	)
	j.AddTask(
		"whoopsWithExponentialBackoff",
		goflow.Bash{Cmd: "whoops", Args: []string{}},
		goflow.TaskParams{
			Retries:    1,
			RetryDelay: &goflow.ExponentialBackoff{},
		},
	)
	j.AddTask(
		"totallySkippable",
		goflow.Bash{Cmd: "sh", Args: []string{"-c", "echo 'everything succeeded'"}},
		goflow.TaskParams{
			TriggerRule: "allSuccessful",
		},
	)
	j.AddTask(
		"cleanUp",
		goflow.Bash{Cmd: "sh", Args: []string{"-c", "echo 'cleaning up now'"}},
		goflow.TaskParams{
			TriggerRule: "allDone",
		},
	)

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
	j := goflow.NewJob("CustomOperator", goflow.JobParams{})
	j.AddTask("posAdd", PositiveAddition{5, 6}, goflow.TaskParams{})
	return j
}

package goflow

import (
	"errors"
	"math/rand"
)

// Crunch some numbers
func complexAnalyticsJob() *Job {
	j := &Job{
		Name:     "example-complex-analytics",
		Schedule: "* * * * * *",
		Active:   false,
	}

	j.Add(&Task{
		Name:     "sleep-one",
		Operator: Command{Cmd: "sleep", Args: []string{"1"}},
	})
	j.Add(&Task{
		Name:     "add-one-one",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((1 + 1))"}},
	})
	j.Add(&Task{
		Name:     "sleep-two",
		Operator: Command{Cmd: "sleep", Args: []string{"2"}},
	})
	j.Add(&Task{
		Name:     "add-two-four",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})
	j.Add(&Task{
		Name:     "add-three-four",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
	})
	j.Add(&Task{
		Name:       "whoops-with-constant-delay",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: ConstantDelay{Period: 1},
	})
	j.Add(&Task{
		Name:       "whoops-with-exponential-backoff",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    1,
		RetryDelay: ExponentialBackoff{},
	})
	j.Add(&Task{
		Name:        "totally-skippable",
		Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'everything succeeded'"}},
		TriggerRule: "allSuccessful",
	})
	j.Add(&Task{
		Name:        "clean-up",
		Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'cleaning up now'"}},
		TriggerRule: "allDone",
	})

	j.SetDownstream(j.Task("sleep-one"), j.Task("add-one-one"))
	j.SetDownstream(j.Task("add-one-one"), j.Task("sleep-two"))
	j.SetDownstream(j.Task("sleep-two"), j.Task("add-two-four"))
	j.SetDownstream(j.Task("add-one-one"), j.Task("add-three-four"))
	j.SetDownstream(j.Task("sleep-one"), j.Task("whoops-with-constant-delay"))
	j.SetDownstream(j.Task("sleep-one"), j.Task("whoops-with-exponential-backoff"))
	j.SetDownstream(j.Task("whoops-with-constant-delay"), j.Task("totally-skippable"))
	j.SetDownstream(j.Task("whoops-with-exponential-backoff"), j.Task("totally-skippable"))
	j.SetDownstream(j.Task("totally-skippable"), j.Task("clean-up"))

	return j
}

// RandomFailure fails randomly. This is a contrived example for demo purposes.
type RandomFailure struct{ n int }

// rng with seed=1
var r = rand.New(rand.NewSource(1))

// Run implements failures at random intervals.
func (o RandomFailure) Run() (interface{}, error) {
	x := r.Intn(o.n)

	if x == o.n-1 {
		return nil, errors.New("unlucky")
	}

	return x, nil
}

// Use our custom operation in a job.
func customOperatorJob() *Job {
	j := &Job{Name: "example-custom-operator", Schedule: "* * * * * *", Active: true}
	j.Add(&Task{Name: "random-failure", Operator: RandomFailure{4}})
	return j
}

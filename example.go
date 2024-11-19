package goflow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
)

// Crunch some numbers
func complexAnalyticsJob() *Job {
	j := &Job{
		Name:     "example-complex-analytics",
		Schedule: "* * * * * *",
		Active:   false,
	}

	err := j.AddTask(
		&Task{
			Name:     "sleep-one",
			Operator: Command{Cmd: "sleep", Args: []string{"1"}},
		},
		&Task{
			Name:     "add-one-one",
			Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((1 + 1))"}},
		},
		&Task{
			Name:     "sleep-two",
			Operator: Command{Cmd: "sleep", Args: []string{"2"}},
		},
		&Task{
			Name:     "add-two-four",
			Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
		},
		&Task{
			Name:     "add-three-four",
			Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
		},
		&Task{
			Name:       "whoops-with-constant-delay",
			Operator:   Command{Cmd: "whoops", Args: []string{}},
			Retries:    5,
			RetryDelay: ConstantDelay{Period: 1},
		},
		&Task{
			Name:       "whoops-with-exponential-backoff",
			Operator:   Command{Cmd: "whoops", Args: []string{}},
			Retries:    1,
			RetryDelay: ExponentialBackoff{},
		},
		&Task{
			Name:        "totally-skippable",
			Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'everything succeeded'"}},
			TriggerRule: "allSuccessful",
		},
		&Task{
			Name:        "clean-up",
			Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'cleaning up now'"}},
			TriggerRule: "allDone",
		},
	)

	if err != nil {
		log.Printf("error adding task: %v", err)
	}

	j.SetDownstream("sleep-one", "add-one-one")
	j.SetDownstream("add-one-one", "sleep-two")
	j.SetDownstream("sleep-two", "add-two-four")
	j.SetDownstream("add-one-one", "add-three-four")
	j.SetDownstream("sleep-one", "whoops-with-constant-delay")
	j.SetDownstream("sleep-one", "whoops-with-exponential-backoff")
	j.SetDownstream("whoops-with-constant-delay", "totally-skippable")
	j.SetDownstream("whoops-with-exponential-backoff", "totally-skippable")
	j.SetDownstream("totally-skippable", "clean-up")

	return j
}

// randomFailure fails randomly. This is a contrived example for demo purposes.
type randomFailure struct{ n int }

// rng with seed=1
var r = rand.New(rand.NewSource(1))

// Run implements failures at random intervals.
func (o randomFailure) Run(ctx context.Context) (any, error) {
	select {
	case <-ctx.Done():
		return nil, errors.New("context cancelled")
	default:
	}
	x := r.Intn(o.n)
	if x == o.n-1 {
		return "randomly failed", errors.New("unlucky")
	}
	return fmt.Sprintf("the result is %v", x), nil
}

func randomFailureJob() *Job {
	j := &Job{Name: "example-random-failure", Schedule: "* * * * * *", Active: true}
	err := j.AddTask(&Task{Name: "random-failure", Operator: randomFailure{4}})
	if err != nil {
		log.Printf("error adding task: %v", err)
	}
	return j
}

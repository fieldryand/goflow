package goflow

import (
	"errors"
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
func (o randomFailure) Run(e *Execution) (any, error) {
	x := r.Intn(o.n)

	if x == o.n-1 {
		return nil, errors.New("unlucky")
	}

	return x, nil
}

func randomFailureJob() *Job {
	j := &Job{Name: "example-random-failure", Schedule: "* * * * * *", Active: true}
	j.Add(&Task{Name: "random-failure", Operator: randomFailure{4}})
	return j
}

// summation is a sum of values. If a summation task is downstream of another,
// then the final result will be the accumulated sum.
type summation struct {
	Value int
}

// Run performs summation.
func (o summation) Run(e *Execution) (any, error) {

	result := o.Value

	for _, task := range e.Tasks {
		if task.State == "successful" {
			if i, ok := task.Operator.(summation); ok {
				result = result + i.Value
			}
		}
	}

	log.Printf("sum value=%v", result)

	return result, nil
}

func summationJob() *Job {
	j := &Job{Name: "example-summation-job", Schedule: "* * * * * *", Active: true}
	j.Add(&Task{Name: "summation-1", Operator: summation{1}})
	j.Add(&Task{Name: "summation-2", Operator: summation{1}})
	j.Add(&Task{Name: "summation-3", Operator: summation{1}})
	j.Add(&Task{Name: "summation-4", Operator: summation{1}})
	j.SetDownstream("summation-1", "summation-2")
	j.SetDownstream("summation-2", "summation-3")
	j.SetDownstream("summation-3", "summation-4")
	return j
}

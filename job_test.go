package goflow

import (
	"errors"
	"testing"

	"github.com/philippgille/gokv/gomap"
)

func TestJob(t *testing.T) {
	j := &Job{Name: "example", Schedule: "* * * * *"}

	j.Add(&Task{
		Name:     "add-one-one",
		Operator: Addition{1, 1},
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
		Operator: Addition{3, 4},
	})
	j.Add(&Task{
		Name:       "whoops-with-constant-delay",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: ConstantDelay{1},
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

	j.SetDownstream(j.Task("add-one-one"), j.Task("sleep-two"))
	j.SetDownstream(j.Task("sleep-two"), j.Task("add-two-four"))
	j.SetDownstream(j.Task("add-one-one"), j.Task("add-three-four"))
	j.SetDownstream(j.Task("add-one-one"), j.Task("whoops-with-constant-delay"))
	j.SetDownstream(j.Task("add-one-one"), j.Task("whoops-with-exponential-backoff"))
	j.SetDownstream(j.Task("whoops-with-constant-delay"), j.Task("totally-skippable"))
	j.SetDownstream(j.Task("whoops-with-exponential-backoff"), j.Task("totally-skippable"))
	j.SetDownstream(j.Task("totally-skippable"), j.Task("clean-up"))

	store := gomap.NewStore(gomap.DefaultOptions)

	go j.run(store, j.newExecution())

	for {
		if j.allDone() {
			break
		}
	}

	if j.loadTaskState("add-one-one") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("add-one-one"), successful)
	}
	if j.loadTaskState("sleep-two") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("sleep-two"), successful)
	}
	if j.loadTaskState("add-two-four") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("add-two-four"), successful)
	}
	if j.loadTaskState("add-three-four") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("add-three-four"), successful)
	}
	if j.loadTaskState("whoops-with-constant-delay") != failed {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("whoops-with-constant-delay"), failed)
	}
	if j.loadTaskState("whoops-with-exponential-backoff") != failed {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("whoops-with-exponential-backoff"), failed)
	}
	if j.loadTaskState("totally-skippable") != skipped {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("totally-skippable"), skipped)
	}
	if j.loadTaskState("clean-up") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("clean-up"), successful)
	}

}

func TestCyclicJob(t *testing.T) {
	j := &Job{Name: "cyclic", Schedule: "* * * * *"}

	j.Add(&Task{Name: "addTwoTwo", Operator: Addition{2, 2}})
	j.Add(&Task{Name: "addFourFour", Operator: Addition{4, 4}})
	j.SetDownstream(j.Task("addTwoTwo"), j.Task("addFourFour"))
	j.SetDownstream(j.Task("addFourFour"), j.Task("addTwoTwo"))

	store := gomap.NewStore(gomap.DefaultOptions)

	j.run(store, j.newExecution())
}

// Adds two nonnegative numbers.
type Addition struct{ a, b int }

func (o Addition) Run() (interface{}, error) {

	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}

	result := o.a + o.b
	return result, nil
}

package goflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/philippgille/gokv/gomap"
)

func TestInvalidTask(t *testing.T) {
	j := &Job{Name: "example", Schedule: "* * * * *", Timeout: 1 * time.Second}

	err := j.Add(&Task{
		Name:     "",
		Operator: Addition{1, 1},
	})

	if err == nil {
		t.Errorf("task with invalid name should be rejected")
	}

	j.Add(&Task{
		Name:     "independent-task",
		Operator: Addition{1, 1},
	})

	j.Add(&Task{
		Name:     "dependent-task",
		Operator: Addition{1, 1},
	})

	err = j.SetDownstream("does-not-exist", "dependent-task")

	if err == nil {
		t.Errorf("edge should not be set between nonexistent tasks")
	}

	err = j.SetDownstream("independent-task", "does-not-exist")

	if err == nil {
		t.Errorf("edge should not be set between nonexistent tasks")
	}
}

func TestJob(t *testing.T) {
	j := &Job{Name: "example", Schedule: "* * * * *", Timeout: 1 * time.Second}

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

	j.SetDownstream("add-one-one", "sleep-two")
	j.SetDownstream("sleep-two", "add-two-four")
	j.SetDownstream("add-one-one", "add-three-four")
	j.SetDownstream("add-one-one", "whoops-with-constant-delay")
	j.SetDownstream("add-one-one", "whoops-with-exponential-backoff")
	j.SetDownstream("whoops-with-constant-delay", "totally-skippable")
	j.SetDownstream("whoops-with-exponential-backoff", "totally-skippable")
	j.SetDownstream("totally-skippable", "clean-up")

	store := gomap.NewStore(gomap.DefaultOptions)

	go j.run(store, j.newExecution(false))

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

// Adds two nonnegative numbers.
type Addition struct{ a, b int }

func (o Addition) Run(ctx context.Context) (interface{}, error) {

	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}

	result := o.a + o.b
	return result, nil
}

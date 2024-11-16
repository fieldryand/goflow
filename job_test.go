package goflow

import (
	"context"
	"testing"

	"github.com/philippgille/gokv/gomap"
)

func TestJob(t *testing.T) {
	j := &Job{Name: "example", Schedule: "* * * * *"}

	err := j.AddTask(
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
			RetryDelay: ConstantDelay{1},
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
		&Task{
			Name:     "failure",
			Operator: randomFailure{1},
		})

	if err != nil {
		t.Errorf("Error adding tasks")
	}

	j.SetDownstream("add-one-one", "sleep-two")
	j.SetDownstream("sleep-two", "add-two-four")
	j.SetDownstream("add-one-one", "add-three-four")
	j.SetDownstream("add-one-one", "whoops-with-constant-delay")
	j.SetDownstream("add-one-one", "whoops-with-exponential-backoff")
	j.SetDownstream("whoops-with-constant-delay", "totally-skippable")
	j.SetDownstream("whoops-with-exponential-backoff", "totally-skippable")
	j.SetDownstream("totally-skippable", "clean-up")

	store := gomap.NewStore(gomap.DefaultOptions)

	ctx := context.Background()
	go j.run(ctx, store, j.newExecution())

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
	if j.loadTaskState("failure") != failed {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("failure"), failed)
	}

}

func TestInvalidTaskName(t *testing.T) {
	j := &Job{Name: "test-invalid-task-name", Schedule: "* * * * *"}

	err := j.AddTask(&Task{
		Name:     "",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})

	if err == nil {
		t.Errorf("Expected error creating a task with an invalid name")
	}

}

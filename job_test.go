package goflow

import (
	"testing"

	"github.com/philippgille/gokv/gomap"
)

func TestJob(t *testing.T) {
	j := &Job{Name: "example", Schedule: "* * * * *"}

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
	j.Add(&Task{
		Name:     "failure",
		Operator: randomFailure{1},
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

	go j.run(store, j.newExecution())

	for {
		if j.allDone() {
			break
		}
	}

	if j.loadTaskState("add-one-one") != Successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("add-one-one"), Successful)
	}
	if j.loadTaskState("sleep-two") != Successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("sleep-two"), Successful)
	}
	if j.loadTaskState("add-two-four") != Successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("add-two-four"), Successful)
	}
	if j.loadTaskState("add-three-four") != Successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("add-three-four"), Successful)
	}
	if j.loadTaskState("whoops-with-constant-delay") != Failed {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("whoops-with-constant-delay"), Failed)
	}
	if j.loadTaskState("whoops-with-exponential-backoff") != Failed {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("whoops-with-exponential-backoff"), Failed)
	}
	if j.loadTaskState("totally-skippable") != Skipped {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("totally-skippable"), Skipped)
	}
	if j.loadTaskState("clean-up") != Successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("clean-up"), Successful)
	}
	if j.loadTaskState("failure") != Failed {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("failure"), Failed)
	}

}

func TestCyclicJob(t *testing.T) {
	j := &Job{Name: "cyclic", Schedule: "* * * * *"}

	j.Add(&Task{
		Name:     "add-two-four",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})
	j.Add(&Task{
		Name:     "add-three-four",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
	})

	j.SetDownstream("add-two-four", "add-three-four")
	err := j.SetDownstream("add-three-four", "add-two-four")

	if err == nil {
		t.Errorf("Expected error creating a cyclic dag")
	}

}

func TestSetDownstream(t *testing.T) {
	j := &Job{Name: "test-downstream", Schedule: "* * * * *"}

	j.Add(&Task{
		Name:     "add-two-four",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})
	j.Add(&Task{
		Name:     "add-three-four",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
	})

	err := j.SetDownstream("does-not-exist", "add-three-four")

	if err == nil {
		t.Errorf("Expected error setting a dependency on a non-existent task")
	}

	err = j.SetDownstream("add-two-four", "does-not-exist")

	if err == nil {
		t.Errorf("Expected error setting a non-existent task as a dependency")
	}

}

func TestInvalidTaskName(t *testing.T) {
	j := &Job{Name: "test-invalid-task-name", Schedule: "* * * * *"}

	err := j.Add(&Task{
		Name:     "",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})

	if err == nil {
		t.Errorf("Expected error creating a task with an invalid name")
	}

}

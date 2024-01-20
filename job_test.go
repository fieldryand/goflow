package goflow

import (
	"errors"
	"testing"
)

func TestJob(t *testing.T) {
	j := &Job{Name: "example", Schedule: "* * * * *"}

	j.Add(&Task{
		Name:     "addOneOne",
		Operator: Addition{1, 1},
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
		Operator: Addition{3, 4},
	})
	j.Add(&Task{
		Name:       "whoopsWithConstantDelay",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: ConstantDelay{1},
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

	j.SetDownstream("addOneOne", "sleepTwo")
	j.SetDownstream("sleepTwo", "addTwoFour")
	j.SetDownstream("addOneOne", "addThreeFour")
	j.SetDownstream("addOneOne", "whoopsWithConstantDelay")
	j.SetDownstream("addOneOne", "whoopsWithExponentialBackoff")
	j.SetDownstream("whoopsWithConstantDelay", "totallySkippable")
	j.SetDownstream("whoopsWithExponentialBackoff", "totallySkippable")
	j.SetDownstream("totallySkippable", "cleanUp")

	go j.run()
	func() {
		for {
			jobState := j.loadState()
			if jobState != running && jobState != none {
				break
			}
		}
	}()

	if j.loadTaskState("addOneOne") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("addOneOne"), successful)
	}
	if j.loadTaskState("sleepTwo") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("sleepTwo"), successful)
	}
	if j.loadTaskState("addTwoFour") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("addTwoFour"), successful)
	}
	if j.loadTaskState("addThreeFour") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("addThreeFour"), successful)
	}
	if j.loadTaskState("whoopsWithConstantDelay") != failed {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("whoopsWithConstantDelay"), failed)
	}
	if j.loadTaskState("whoopsWithExponentialBackoff") != failed {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("whoopsWithExponentialBackoff"), failed)
	}
	if j.loadTaskState("totallySkippable") != skipped {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("totallySkippable"), skipped)
	}
	if j.loadTaskState("cleanUp") != successful {
		t.Errorf("Got status %v, expected %v", j.loadTaskState("cleanUp"), successful)
	}
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

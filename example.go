package goflow

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Crunch some numbers
func complexAnalyticsJob() *Job {
	j := &Job{
		Name:     "exampleComplexAnalytics",
		Schedule: "* * * * * *",
		Active:   false,
	}

	j.Add(&Task{
		Name:     "sleepOne",
		Operator: Command{Cmd: "sleep", Args: []string{"1"}},
	})
	j.Add(&Task{
		Name:     "addOneOne",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((1 + 1))"}},
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
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
	})
	j.Add(&Task{
		Name:       "whoopsWithConstantDelay",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: ConstantDelay{Period: 1},
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

// PositiveAddition adds two nonnegative numbers. This is just a contrived example to
// demonstrate the usage of custom operators.
type PositiveAddition struct{ a, b int }

// Run implements the custom operation.
func (o PositiveAddition) Run() (interface{}, error) {
	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}
	result := o.a + o.b
	return result, nil
}

// Use our custom operation in a job.
func customOperatorJob() *Job {
	j := &Job{Name: "exampleCustomOperator", Schedule: "* * * * * *", Active: false}
	j.Add(&Task{Name: "posAdd", Operator: PositiveAddition{5, 6}})
	return j
}

type PassKey string

type CtxPass1 struct{ J *Job }

func (o CtxPass1) Run() (interface{}, error) {
	fmt.Println(o.J.Name)
	o.J.Ctx = context.WithValue(o.J.Ctx, PassKey("ids"), "1,2,3,4")
	time.Sleep(3 * time.Second)

	return nil, nil
}

type CtxPass2 struct{ J *Job }

func (o CtxPass2) Run() (interface{}, error) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	num := rand.Intn(100)
	fmt.Println(o.J.Ctx.Value(PassKey("ids")))
	o.J.Ctx = context.WithValue(o.J.Ctx, PassKey("CtxPass2"), "done2")
	o.J.Ctx = context.WithValue(o.J.Ctx, PassKey("CtxPass2Value"), num)
	return nil, nil
}

type CtxPass3 struct{ J *Job }

func (o CtxPass3) Run() (interface{}, error) {
	fmt.Println(o.J.Ctx.Value(PassKey("ids")))
	o.J.Ctx = context.WithValue(o.J.Ctx, PassKey("CtxPass3"), "done3")

	return nil, nil
}

type CtxPass4 struct{ J *Job }

func (o CtxPass4) Run() (interface{}, error) {

	fmt.Println(o.J.Ctx.Value(PassKey("CtxPass2")))
	fmt.Println(o.J.Ctx.Value(PassKey("CtxPass2Value")))
	fmt.Println(o.J.Ctx.Value(PassKey("CtxPass3")))
	return nil, nil
}

func ctxPassJob() *Job {
	j := &Job{Name: "exampleCtxOperator", Schedule: "0 */1 * * * *", Active: false, Ctx: context.Background()}
	j.Add(&Task{Name: "CtxPass1", Operator: CtxPass1{j}})
	j.Add(&Task{Name: "CtxPass2", Operator: CtxPass2{j}})
	j.Add(&Task{Name: "CtxPass3", Operator: CtxPass3{j}})
	j.Add(&Task{Name: "CtxPass4", Operator: CtxPass4{j}})

	j.SetDownstream(j.Task("CtxPass1"), j.Task("CtxPass2"))
	j.SetDownstream(j.Task("CtxPass1"), j.Task("CtxPass3"))
	j.SetDownstream(j.Task("CtxPass2"), j.Task("CtxPass4"))
	j.SetDownstream(j.Task("CtxPass3"), j.Task("CtxPass4"))
	return j
}

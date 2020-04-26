// Package operators defines goflow operators.
package operators

import (
	"time"
)

type Operator interface {
	Run() (interface{}, error)
}

type operatorError struct {
	msg string
}

func (e *operatorError) Error() string {
	return e.msg
}

type addOperator struct{ a, b int }

func AddOperator(a, b int) *addOperator {
	o := addOperator{a, b}
	return &o
}

func (addop addOperator) Run() (interface{}, error) {

	if addop.a < 0 || addop.b < 0 {
		return 0, &operatorError{"Can't add negative numbers"}
	}

	result := addop.a + addop.b
	return result, nil
}

type sleepOperator struct{ s int }

func SleepOperator(s int) *sleepOperator {
	o := sleepOperator{s}
	return &o
}

func (slpop sleepOperator) Run() (interface{}, error) {
	time.Sleep(time.Duration(slpop.s) * time.Second)
	return true, nil
}

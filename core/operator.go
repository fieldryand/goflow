package core

import (
	"time"
)

type Operator interface {
	run() (interface{}, error)
}

type AddOperator struct {a, b int}

func (addop AddOperator) run() (interface{}, error) {
	result := addop.a + addop.b
	return result, nil
}

type SleepOperator struct {s int}

func (slpop SleepOperator) run() (interface{}, error) {
	time.Sleep(time.Duration(slpop.s) * time.Second)
	return true, nil
}

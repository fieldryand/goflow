package core

import (
	"time"
)

type operator interface {
	run() (interface{}, error)
}

type addOperator struct{ a, b int }

func AddOperator(a, b int) *addOperator {
	o := addOperator{a, b}
	return &o
}

func (addop addOperator) run() (interface{}, error) {
	result := addop.a + addop.b
	return result, nil
}

type sleepOperator struct{ s int }

func SleepOperator(s int) *sleepOperator {
	o := sleepOperator{s}
	return &o
}

func (slpop sleepOperator) run() (interface{}, error) {
	time.Sleep(time.Duration(slpop.s) * time.Second)
	return true, nil
}

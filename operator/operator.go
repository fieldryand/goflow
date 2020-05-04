// Package operator defines an interface for goflow operators.
package operator

import (
	"time"
)

// An Operator implements a Run() method. When a job executes a task that
// uses the operator, the Run() method is called.
type Operator interface {
	Run() (interface{}, error)
}

// A Sleep operator sleeps for s seconds.
type Sleep struct{ s int }

// NewSleep returns a sleep operator.
func NewSleep(s int) *Sleep {
	o := Sleep{s}
	return &o
}

// Run implements the sleep interface.
func (o Sleep) Run() (interface{}, error) {
	time.Sleep(time.Duration(o.s) * time.Second)
	return true, nil
}

// Package operator defines an interface for goflow operators.
package operator

import (
	"time"
)

// Operators implement a Run() method. When a job executes a task that
// uses the operator, the Run() method is called.
type Operator interface {
	Run() (interface{}, error)
}

// Sleeps for s seconds.
type Sleep struct{ s int }

func NewSleep(s int) *Sleep {
	o := Sleep{s}
	return &o
}

func (o Sleep) Run() (interface{}, error) {
	time.Sleep(time.Duration(o.s) * time.Second)
	return true, nil
}

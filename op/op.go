// Package op defines an interface for goflow operators.
package op

import (
	"os/exec"
)

// An Operator implements a Run() method. When a job executes a task that
// uses the operator, the Run() method is called.
type Operator interface {
	Run() (interface{}, error)
}

// A bash operator executes a shell command.
type BashOperator struct {
	cmd  string
	args []string
}

// NewBashOperator returns a bash operator.
func Bash(cmd string, args ...string) *BashOperator {
	o := BashOperator{cmd, args}
	return &o
}

// Run passes the command and arguments to exec.Command and captures the
// output.
func (o BashOperator) Run() (interface{}, error) {
	out, err := exec.Command(o.cmd, o.args...).Output()
	return string(out), err
}

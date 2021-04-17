package goflow

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
)

// An Operator implements a Run() method. When a job executes a task that
// uses the operator, the Run() method is called.
type Operator interface {
	Run() (interface{}, error)
}

// BashOperator executes a shell command.
type BashOperator struct {
	Cmd  string
	Args []string
}

// BashOp returns a bash operator.
func BashOp(cmd string, args ...string) *BashOperator {
	return &BashOperator{cmd, args}
}

// Run passes the command and arguments to exec.Command and captures the
// output.
func (o BashOperator) Run() (interface{}, error) {
	out, err := exec.Command(o.Cmd, o.Args...).Output()
	return string(out), err
}

// GetOperator makes a GET request.
type GetOperator struct {
	URL string
}

// GetOp returns a get operator.
func GetOp(url string) *GetOperator {
	return &GetOperator{url}
}

// Run sends the request and returns an error if the status code is
// outside the 2xx range.
func (o GetOperator) Run() (interface{}, error) {
	res, err := http.Get(o.URL)
	if err != nil {
		return nil, err
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("Received status code %v", res.StatusCode)
	} else {
		content, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return string(content), err
	}
}

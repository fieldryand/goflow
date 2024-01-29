package goflow

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

// An Operator implements a Run() method. When a job executes a task that
// uses the operator, the Run() method is called.
type Operator interface {
	Run() (interface{}, error)
}

// An OperatorWithContext implements a RunWithContext() method.
// When a job executes a task that uses the operator, the operator
// has access to a Context passed from previous tasks.
type OperatorWithContext interface {
	RunWithContext(ctx context.Context) (interface{}, error)
}

// Command executes a shell command.
type Command struct {
	Cmd  string
	Args []string
}

// Run passes the command and arguments to exec.Command and captures the
// output.
func (o Command) Run() (interface{}, error) {
	out, err := exec.Command(o.Cmd, o.Args...).Output()
	return string(out), err
}

// Get makes a GET request.
type Get struct {
	Client *http.Client
	URL    string
}

// Run sends the request and returns an error if the status code is
// outside the 2xx range.
func (o Get) Run() (interface{}, error) {
	res, err := o.Client.Get(o.URL)
	if err != nil {
		return nil, err
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("Received status code %v", res.StatusCode)
	} else {
		content, err := io.ReadAll(res.Body)
		res.Body.Close()
		return string(content), err
	}
}

// Post makes a POST request.
type Post struct {
	Client *http.Client
	URL    string
	Body   io.Reader
}

// Run sends the request and returns an error if the status code is
// outside the 2xx range.
func (o Post) Run() (interface{}, error) {
	res, err := o.Client.Post(o.URL, "application/json", o.Body)
	if err != nil {
		return nil, err
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("Received status code %v", res.StatusCode)
	} else {
		content, err := io.ReadAll(res.Body)
		res.Body.Close()
		return string(content), err
	}
}

[![Build Status](https://travis-ci.org/fieldryand/goflow.svg?branch=master)](https://travis-ci.org/fieldryand/goflow)
[![codecov](https://codecov.io/gh/fieldryand/goflow/branch/master/graph/badge.svg)](https://codecov.io/gh/fieldryand/goflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/fieldryand/goflow)](https://goreportcard.com/report/github.com/fieldryand/goflow)
[![Release](https://img.shields.io/github/v/release/fieldryand/goflow)](https://github.com/fieldryand/goflow/releases)

# Goflow

A workflow scheduler written in Go and geared toward orchestration of ETL or analytics workloads. Goflow comes complete with a web UI for inspecting and triggering jobs. It is heavily inspired by Apache Airflow and [Dagobah](https://github.com/thieman/dagobah), but the emphasis is on ease of setup and cost (runnable on an EC2 or GCE micro instance).

## Usage

```go
// An example program demonstrating the usage of goflow.
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
	"github.com/fieldryand/goflow/op"
)

func main() {
	gf := goflow.NewEngine(complexAnalyticsJob, messedUpJob, customOperatorJob)
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8100")
}

// Crunch some numbers
func complexAnalyticsJob() *goflow.Job {
	j := goflow.NewJob("ComplexAnalytics")

	j.AddTask("sleepOne", op.Bash("sleep", "1"))
	j.AddTask("addOneOne", op.Bash("sh", "-c", "echo $((1 + 1))"))
	j.AddTask("sleepTwo", op.Bash("sleep", "2"))
	j.AddTask("addTwoFour", op.Bash("sh", "-c", "echo $((2 + 4))"))
	j.AddTask("addThreeFour", op.Bash("sh", "-c", "echo $((2 + 4))"))

	j.SetDownstream(j.Task("sleepOne"), j.Task("addOneOne"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("sleepTwo"))
	j.SetDownstream(j.Task("sleepTwo"), j.Task("addTwoFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("addThreeFour"))

	return j
}

// A task that throws an error
func messedUpJob() *goflow.Job {
	return goflow.NewJob("MessedUp").AddTask("whoops", op.Bash("whoops"))
}

// We can create custom operators by implementing the Run() method.

// Add two nonnegative numbers
type positiveAdditionOperator struct{ a, b int }

func positiveAddition(a, b int) *positiveAdditionOperator {
	o := positiveAdditionOperator{a, b}
	return &o
}

// Run implements the custom operation
func (o positiveAdditionOperator) Run() (interface{}, error) {
	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}
	result := o.a + o.b
	return result, nil
}

// Use our custom operation in a job
func customOperatorJob() *goflow.Job {
	j := goflow.NewJob("CustomOperator")
	j.AddTask("posAdd", positiveAddition(5, 6))
	return j
}

```

You can run this example yourself:

```
go install examples/simple/goflow-simple-example.go
eval "$GOPATH/bin/goflow-simple-example"
```

Then browse to `localhost:8100` to explore the UI, where you can submit jobs and view their current state.

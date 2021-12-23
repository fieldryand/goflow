[![Build Status](https://travis-ci.com/fieldryand/goflow.svg?branch=master)](https://travis-ci.com/fieldryand/goflow)
[![codecov](https://codecov.io/gh/fieldryand/goflow/branch/master/graph/badge.svg)](https://codecov.io/gh/fieldryand/goflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/fieldryand/goflow)](https://goreportcard.com/report/github.com/fieldryand/goflow)
[![GoDoc](https://pkg.go.dev/badge/github.com/fieldryand/goflow?status.svg)](https://pkg.go.dev/github.com/fieldryand/goflow?tab=doc)
[![Release](https://img.shields.io/github/v/release/fieldryand/goflow)](https://github.com/fieldryand/goflow/releases)

# Goflow

A workflow/DAG orchestrator written in Go and meant for ETL or analytics pipelines. Goflow comes complete with a web UI for inspecting and triggering jobs.

## Contents

1. [Screenshots](#screenshots)
2. [Motivation](#motivation)
3. [Concepts and features](#concepts-and-features)
   1. [Jobs and tasks](#jobs-and-tasks)
   2. [Custom Operators](#custom-operators)
   3. [Retries](#retries)
   4. [Trigger rules](#trigger-rules)
   5. [The Goflow engine](#the-goflow-engine)
4. [Installation and development](#installation-and-development)
   1. [Running the example](#running-the-example)
   2. [TODO: Docker image](#todo-docker-image)

## Screenshots

Job details page:

![screenshot-jobs-complex-analytics-02](https://user-images.githubusercontent.com/3333324/120916675-3375d680-c6ab-11eb-8a28-a5adecd34512.png)

## Motivation

Goflow was built as a simple replacement for Apache Airflow to manage some small data pipeline projects. Airflow started to feel too heavyweight for these projects where all the computation was offloaded to independent services. I wanted a solution with minimal memory requirements to save costs and avoid the occasional high memory usage/leak issues I was facing with Airflow.

## Concepts and features

- `Job`: A Goflow workflow is called a `Job`. Jobs can be scheduled using cron syntax.
- `Task`: Each job consists of one or more tasks organized into a dependency graph. A task can be run under certain conditions; by default, a task runs when all of its dependencies finish successfully.
- Concurrency: Jobs and tasks execute concurrently.
- `Operator`: An `Operator` defines the work done by a `Task`. Goflow comes with two basic operators: `Bash` for running shell commands and `Get` for HTTP GET requests. Implementing your own `Operator` is straightforward.
- Retries: You can allow a `Task` a given number of retry attempts. Goflow comes with two retry strategies, `ConstantDelay` and `ExponentialBackoff`.
- Database: Goflow supports two database types, in-memory and BoltDB. BoltDB will persist your history of job runs, whereas in-memory means the history will be lost each time the Goflow server is stopped. The default is BoltDB.

### Jobs and tasks

Let's start by creating a function that returns a job called `myJob`. There is a single task in this job that sleeps for one second.

```go
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
)

func myJob() *goflow.Job {
	j := &goflow.Job{Name: "myJob", Schedule: "* * * * *"}
	j.Add(&goflow.Task{
		Name:     "sleepForOneSecond",
		Operator: goflow.Bash{Cmd: "sleep", Args: []string{"1"}},
	})
	return j
}
```

### Custom operators

A custom `Operator` needs to implement the `Run` method. Here's an example of an operator that adds two positive numbers.

```go
type PositiveAddition struct{ a, b int }

func (o PositiveAddition) Run() (interface{}, error) {
	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}
	result := o.a + o.b
	return result, nil
}
```

### Retries

Let's add a retry strategy to `myJob`:

```go
func myJob() *goflow.Job {
	j := &goflow.Job{Name: "myJob", Schedule: "* * * * *"}
	j.Add(&goflow.Task{
		Name:       "sleepForOneSecond",
		Operator:   goflow.Bash{Cmd: "sleep", Args: []string{"1"}},
		Retries:    5,
		RetryDelay: goflow.ConstantDelay{Period: 1},
	})
	return j
}
```

Instead of `ConstantDelay`, we could also use `ExponentialBackoff` (see https://en.wikipedia.org/wiki/Exponential_backoff).

### Trigger rules

By default, a task has the trigger rule `allSuccessful`, meaning the task starts executing when all the tasks directly
upstream exit successfully. If any dependency exits with an error, all downstream tasks are skipped, and the job exits with an error.

Sometimes you want a downstream task to execute even if there are upstream failures. Often these are situations where you want
to perform some cleanup task, such as shutting down a server. In such cases, you can give a task the trigger rule `allDone`.

Let's modify `myJob` to have the trigger rule `allDone`.


```go
func myJob() *goflow.Job {
	j := &goflow.Job{Name: "myJob", Schedule: "* * * * *"}
	j.Add(&goflow.Task{
		Name:        "sleepForOneSecond",
		Operator:    goflow.Bash{Cmd: "sleep", Args: []string{"1"}},
		Retries:     5,
		RetryDelay:  goflow.ConstantDelay{Period: 1},
		TriggerRule: "allDone",
	})
	return j
}
```

### The Goflow Engine

Finally, let's create a Goflow engine, register our job, attach a logger, and run the application.

```go
func main() {
	gf := goflow.New(goflow.Options{})
	gf.AddJob(myJob)
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8181")
}
```

You can pass different options to the engine. Options currently supported:
- `DBType`: `boltdb` (default) or `memory`
- `BoltDBPath`: This will be the filepath of the Bolt database on disk.

Goflow is built on the [Gin framework](https://github.com/gin-gonic/gin), so you can pass any Gin handler to `Use`.

## Installation and development

In order to use Goflow you need Go and NPM installed on your system.

### Running the example

Here's how to run the example application included in this repo. First, clone this repo into your `GOPATH`.

```shell
mkdir -p $GOPATH/src/github.com/fieldryand
cd $GOPATH/src/github.com/fieldryand
git clone https://github.com/fieldryand/goflow.git
```

Next, run `compile_assets.sh` to build the frontend.

```shell
./compile_assets.sh
```

Install the Go dependencies and run the application.

```shell
go get
go run examples/simple/goflow-simple-example.go
```

Finally, browse to `localhost:8181` to explore the UI, where you can submit jobs and view their current state.

### TODO: Docker image

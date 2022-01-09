![Build Status](https://github.com/fieldryand/goflow/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/fieldryand/goflow/branch/master/graph/badge.svg)](https://codecov.io/gh/fieldryand/goflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/fieldryand/goflow)](https://goreportcard.com/report/github.com/fieldryand/goflow)
[![GoDoc](https://pkg.go.dev/badge/github.com/fieldryand/goflow?status.svg)](https://pkg.go.dev/github.com/fieldryand/goflow?tab=doc)
[![Release](https://img.shields.io/github/v/release/fieldryand/goflow)](https://github.com/fieldryand/goflow/releases)

# Goflow

A workflow/DAG orchestrator written in Go for rapid prototyping of ETL/ML/AI pipelines. Goflow comes complete with a web UI for inspecting and triggering jobs.

## Contents

1. [Quick start](#quick-start)
2. [Use case](#use-case)
3. [Concepts and features](#concepts-and-features)
   1. [Jobs and tasks](#jobs-and-tasks)
   2. [Custom Operators](#custom-operators)
   3. [Retries](#retries)
   4. [Task dependencies](#task-dependencies)
   5. [Trigger rules](#trigger-rules)
   6. [The Goflow engine](#the-goflow-engine)

## Quick start

### With Docker

```shell
docker run -p 8181:8181 ghcr.io/fieldryand/goflow-example
```

Browse to `localhost:8181` to explore the UI.

![goflow-demo](https://user-images.githubusercontent.com/3333324/147818084-ade84547-4404-4d58-a697-c18ecb06fd30.gif)

### Without Docker

```shell
go get github.com/fieldryand/goflow
```

TODO: UI dependencies

## Use case

Goflow was built as a simple replacement for Apache Airflow to manage some small data pipeline projects. Airflow started to feel too heavyweight for these projects where all the computation was offloaded to independent services, but there was still a need for basic orchestration, concurrency, retries, visibility etc.

Goflow prioritizes ease of deployment over features and scalability. If you need distributed workers, backfilling over time slices, a durable database of job runs, etc, then Goflow is not for you. On the other hand, if you want to rapidly prototype some pipelines, then Goflow might be a good fit.

## Concepts and features

- `Job`: A Goflow workflow is called a `Job`. Jobs can be scheduled using cron syntax.
- `Task`: Each job consists of one or more tasks organized into a dependency graph. A task can be run under certain conditions; by default, a task runs when all of its dependencies finish successfully.
- Concurrency: Jobs and tasks execute concurrently.
- `Operator`: An `Operator` defines the work done by a `Task`. Goflow comes with a handful of basic operators, and implementing your own `Operator` is straightforward.
- Retries: You can allow a `Task` a given number of retry attempts. Goflow comes with two retry strategies, `ConstantDelay` and `ExponentialBackoff`.
- Database: Goflow supports two database types, in-memory and BoltDB. BoltDB will persist your history of job runs, whereas in-memory means the history will be lost each time the Goflow server is stopped. The default is BoltDB.
- Streaming: Goflow uses server-sent events to stream the status of jobs and tasks to the UI in real time.

### Jobs and tasks

Let's start by creating a function that returns a job called `myJob`. There is a single task in this job that sleeps for one second.

```go
package main

import (
	"errors"

	"github.com/fieldryand/goflow"
)

func myJob() *goflow.Job {
	j := &goflow.Job{Name: "myJob", Schedule: "* * * * *", Active: true}
	j.Add(&goflow.Task{
		Name:     "sleepForOneSecond",
		Operator: goflow.Command{Cmd: "sleep", Args: []string{"1"}},
	})
	return j
}
```

By setting `Active: true`, we are telling Goflow to apply the provided cron schedule for this job when the application starts.
Job scheduling can be activated and deactivated from the UI.

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

Let's add a retry strategy to the `sleepForOneSecond` task:

```go
func myJob() *goflow.Job {
	j := &goflow.Job{Name: "myJob", Schedule: "* * * * *"}
	j.Add(&goflow.Task{
		Name:       "sleepForOneSecond",
		Operator:   goflow.Command{Cmd: "sleep", Args: []string{"1"}},
		Retries:    5,
		RetryDelay: goflow.ConstantDelay{Period: 1},
	})
	return j
}
```

Instead of `ConstantDelay`, we could also use `ExponentialBackoff` (see https://en.wikipedia.org/wiki/Exponential_backoff).

### Task dependencies

A job can define a directed acyclic graph (DAG) of independent and dependent tasks. Let's use the `SetDownstream` method to
define two tasks that are dependent on `sleepForOneSecond`. The tasks will use the `PositiveAddition` operator we defined earlier,
as well as a new operator provided by Goflow, `Get`.

```go
func myJob() *goflow.Job {
	j := &goflow.Job{Name: "myJob", Schedule: "* * * * *"}
	j.Add(&goflow.Task{
		Name:       "sleepForOneSecond",
		Operator:   goflow.Command{Cmd: "sleep", Args: []string{"1"}},
		Retries:    5,
		RetryDelay: goflow.ConstantDelay{Period: 1},
	})
	j.Add(&goflow.Task{
		Name:       "getGoogle",
		Operator:   goflow.Get{Client: &http.Client{}, URL: "https://www.google.com"},
	})
	j.Add(&goflow.Task{
		Name:       "AddTwoPlusThree",
		Operator:   PositiveAddition{a: 2, b: 3},
	})
	j.SetDownstream(j.Task("sleepForOneSecond"), j.Task("getGoogle"))
	j.SetDownstream(j.Task("sleepForOneSecond"), j.Task("AddTwoPlusThree"))
	return j
}
```

### Trigger rules

By default, a task has the trigger rule `allSuccessful`, meaning the task starts executing when all the tasks directly
upstream exit successfully. If any dependency exits with an error, all downstream tasks are skipped, and the job exits with an error.

Sometimes you want a downstream task to execute even if there are upstream failures. Often these are situations where you want
to perform some cleanup task, such as shutting down a server. In such cases, you can give a task the trigger rule `allDone`.

Let's modify `sleepForOneSecond` to have the trigger rule `allDone`.


```go
func myJob() *goflow.Job {
	// other stuff
	j.Add(&goflow.Task{
		Name:        "sleepForOneSecond",
		Operator:    goflow.Command{Cmd: "sleep", Args: []string{"1"}},
		Retries:     5,
		RetryDelay:  goflow.ConstantDelay{Period: 1},
		TriggerRule: "allDone",
	})
	// other stuff
}
```

### The Goflow Engine

Finally, let's create a Goflow engine, register our job, attach a logger, and run the application.

```go
func main() {
	gf := goflow.New(goflow.Options{StreamJobRuns: true})
	gf.AddJob(myJob)
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8181")
}
```

You can pass different options to the engine. Options currently supported:
- `DBType`: `boltdb` (default) or `memory`
- `BoltDBPath`: This will be the filepath of the Bolt database on disk.
- `StreamJobRuns`: Whether to stream updates to the UI.
- `ShowExamples`: Whether to show the example jobs.

Goflow is built on the [Gin framework](https://github.com/gin-gonic/gin), so you can pass any Gin handler to `Use`.

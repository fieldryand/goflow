![Build Status](https://github.com/fieldryand/goflow/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/fieldryand/goflow/branch/master/graph/badge.svg)](https://codecov.io/gh/fieldryand/goflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/fieldryand/goflow)](https://goreportcard.com/report/github.com/fieldryand/goflow)
[![GoDoc](https://pkg.go.dev/badge/github.com/fieldryand/goflow/v2?status.svg)](https://pkg.go.dev/github.com/fieldryand/goflow/v2?tab=doc)
[![Release](https://img.shields.io/github/v/release/fieldryand/goflow)](https://github.com/fieldryand/goflow/releases)

# Goflow

A simple but powerful DAG scheduler and dashboard, written in Go.

![goflow-demo](https://user-images.githubusercontent.com/3333324/147818084-ade84547-4404-4d58-a697-c18ecb06fd30.gif)

------

**Use it if:**
- you need a directed acyclic graph (DAG) scheduler like Apache Airflow, but without the complexity.
- you have a variety of clusters or services performing heavy computations and you want something small and light to orchestrate them.
- you want a monitoring dashboard.
- you want the easiest possible deployment with a single binary or container, saving you time. Volume mounts etc are too much headache.
- you want it to run on a single tiny VM, saving on cloud costs.
- you want to choose your storage technology--embedded, Postgres, Redis, S3, DynamoDB or something else.
- you prefer to define your DAGs with code rather than configuration files. This approach can make it easier to manage complex DAGs.

**Don't use it if:**
- you need to queue a huge number of tasks. Goflow is not tested at massive scale and does not support horizontal scaling.

## Contents

- [Quick start](#quick-start)
   - [With Docker](#with-docker)
   - [Without Docker](#without-docker)
- [Development overview](#development-overview)
   - [Jobs and tasks](#jobs-and-tasks)
   - [Custom Operators](#custom-operators)
   - [Retries](#retries)
   - [Task dependencies](#task-dependencies)
   - [Trigger rules](#trigger-rules)
   - [The Goflow engine](#the-goflow-engine)
   - [Available operators](#available-operators)
- [Storage](#storage)
- [API and integration](#api-and-integration)

## Quick start

### With Docker

```shell
docker run -p 8181:8181 ghcr.io/fieldryand/goflow-example:latest
```

Check out the dashboard at `localhost:8181`.

### Without Docker

In a fresh project directory:

```shell
go mod init # create a new module
go get github.com/fieldryand/goflow/v2 # install dependencies
```

Create a file `main.go` with contents:
```go
package main

import "github.com/fieldryand/goflow/v2"

func main() {
        options := goflow.Options{
                UIPath: "ui/",
                ShowExamples:  true,
                WithSeconds:  true,
        }
        gf := goflow.New(options)
        gf.Use(goflow.DefaultLogger())
        gf.Run(":8181")
}
```

Download and untar the dashboard:

```shell
wget https://github.com/fieldryand/goflow/releases/latest/download/goflow-ui.tar.gz
tar -xvzf goflow-ui.tar.gz
rm goflow-ui.tar.gz
```

Now run the application with `go run main.go` and see it in the browser at localhost:8181.

## Development overview

First a few definitions.

- `Job`: A Goflow workflow is called a `Job`. Jobs can be scheduled using cron syntax.
- `Task`: Each job consists of one or more tasks organized into a dependency graph. A task can be run under certain conditions; by default, a task runs when all of its dependencies finish successfully.
- Concurrency: Jobs and tasks execute concurrently.
- `Operator`: An `Operator` defines the work done by a `Task`. Goflow comes with a handful of basic operators, and implementing your own `Operator` is straightforward.
- Retries: You can allow a `Task` a given number of retry attempts. Goflow comes with two retry strategies, `ConstantDelay` and `ExponentialBackoff`.
- Streaming: Goflow uses server-sent events to stream the status of jobs and tasks to the dashboard in real time.

### Jobs and tasks

Let's start by creating a function that returns a job called `my-job`. There is a single task in this job that sleeps for one second.

```go
package main

import (
	"errors"

	"github.com/fieldryand/goflow/v2"
)

func myJob() *goflow.Job {
	j := &goflow.Job{Name: "my-job", Schedule: "* * * * *", Active: true}
	j.Add(&goflow.Task{
		Name:     "sleep-for-one-second",
		Operator: goflow.Command{Cmd: "sleep", Args: []string{"1"}},
	})
	return j
}
```

By setting `Active: true`, we are telling Goflow to apply the provided cron schedule for this job when the application starts.
Job scheduling can be activated and deactivated from the dashboard.

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

Let's add a retry strategy to the `sleep-for-one-second` task:

```go
func myJob() *goflow.Job {
	j := &goflow.Job{Name: "my-job", Schedule: "* * * * *"}
	j.Add(&goflow.Task{
		Name:       "sleep-for-one-second",
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
define two tasks that are dependent on `sleep-for-one-second`. The tasks will use the `PositiveAddition` operator we defined earlier,
as well as a new operator provided by Goflow, `Get`.

```go
func myJob() *goflow.Job {
	j := &goflow.Job{Name: "my-job", Schedule: "* * * * *"}
	j.Add(&goflow.Task{
		Name:       "sleep-for-one-second",
		Operator:   goflow.Command{Cmd: "sleep", Args: []string{"1"}},
		Retries:    5,
		RetryDelay: goflow.ConstantDelay{Period: 1},
	})
	j.Add(&goflow.Task{
		Name:       "get-google",
		Operator:   goflow.Get{Client: &http.Client{}, URL: "https://www.google.com"},
	})
	j.Add(&goflow.Task{
		Name:       "add-two-plus-three",
		Operator:   PositiveAddition{a: 2, b: 3},
	})
	j.SetDownstream(j.Task("sleep-for-one-second"), j.Task("get-google"))
	j.SetDownstream(j.Task("sleep-for-one-second"), j.Task("add-two-plus-three"))
	return j
}
```

### Trigger rules

By default, a task has the trigger rule `allSuccessful`, meaning the task starts executing when all the tasks directly
upstream exit successfully. If any dependency exits with an error, all downstream tasks are skipped, and the job exits with an error.

Sometimes you want a downstream task to execute even if there are upstream failures. Often these are situations where you want
to perform some cleanup task, such as shutting down a server. In such cases, you can give a task the trigger rule `allDone`.

Let's modify `sleep-for-one-second` to have the trigger rule `allDone`.


```go
func myJob() *goflow.Job {
	// other stuff
	j.Add(&goflow.Task{
		Name:        "sleep-for-one-second",
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
	gf := goflow.New(goflow.Options{Streaming: true})
	gf.AddJob(myJob)
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8181")
}
```

You can pass different options to the engine. Options currently supported:
- `Store`: This is [described in more detail below.](#storage)
- `UIPath`: The path to the dashboard code. The default value is an empty string, meaning Goflow serves only the API and not the dashboard. Suggested value if you want the dashboard: `ui/`
- `ShowExamples`: Whether to show the example jobs. Default value: `false`
- `WithSeconds`: Whether to include the seconds field in the cron spec. See the [cron package documentation](https://github.com/robfig/cron) for details. Default value: `false`

Goflow is built on the [Gin framework](https://github.com/gin-gonic/gin), so you can pass any Gin handler to `Use`.

### Available operators

Goflow provides several operators for common tasks. [See the package documentation](https://pkg.go.dev/github.com/fieldryand/goflow) for details on each.

- `Command` executes a shell command.
- `Get` makes a GET request.
- `Post` makes a POST request.

## Storage

For persisting your job execution history, Goflow allows you to plug in many different key-value stores thanks to the [excellent gokv package](https://github.com/philippgille/gokv/). This way you can recover from a crash or deploy a new version of your app without losing your data.

> Note: the gokv API is not yet stable. Goflow has been tested against v0.6.0.

By default, Goflow uses an in-memory database, but you can easily replace it with Postgres, Redis, S3 or any other `gokv.Store`. Here is an example:

```go
package main

import "github.com/fieldryand/goflow/v2"
import "github.com/philippgille/gokv/redis"

func main() {
        // create a storage client
        client, err := redis.NewClient(redis.DefaultOptions)
        if err != nil {
                panic(err)
        }
        defer client.Close()

        // pass the client as a Goflow option
        options := goflow.Options{
                Store: client,
                UIPath: "ui/",
                Streaming: true,
                ShowExamples:  true,
        }
        gf := goflow.New(options)
        gf.Use(goflow.DefaultLogger())
        gf.Run(":8181")
}
```


## API and integration

You can use the API to integrate Goflow with other applications, such as an existing dashboard. Here is an overview of available endpoints:
- `GET /api/health`: Check health of the service
- `GET /api/jobs`: List registered jobs
- `GET /api/jobs/{jobname}`: Get the details for a given job
- `GET /api/executions`: Query and list job executions
- `POST /api/jobs/{jobname}/submit`: Submit a job for execution
- `POST /api/jobs/{jobname}/toggle`: Toggle a job schedule on or off
- `/stream`: This endpoint returns Server-Sent Events with a `data` payload matching the one returned by `/api/executions`. The dashboard that ships with Goflow uses this endpoint.

Check out the OpenAPI spec for more details. Easiest way is to clone the repo, then within the repo use Swagger as in the following:

```shell
docker run -p 8080:8080 -e SWAGGER_JSON=/app/swagger.json -v $(pwd):/app swaggerapi/swagger-ui
```

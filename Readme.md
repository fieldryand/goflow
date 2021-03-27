[![Build Status](https://travis-ci.org/fieldryand/goflow.svg?branch=master)](https://travis-ci.org/fieldryand/goflow)
[![codecov](https://codecov.io/gh/fieldryand/goflow/branch/master/graph/badge.svg)](https://codecov.io/gh/fieldryand/goflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/fieldryand/goflow)](https://goreportcard.com/report/github.com/fieldryand/goflow)
[![Release](https://img.shields.io/github/v/release/fieldryand/goflow)](https://github.com/fieldryand/goflow/releases)

# Goflow

A minimal workflow scheduler written in Go, inspired by Apache Airflow and github.com/thieman/dagobah. Goflow retains the code-as-configuration philosophy of Airflow, but is much more lightweight because it strips away the emphasis on Celery-style distributed task execution. Instead, Goflow is suited for situations where your workflow tasks can be executed by microservices.

## Usage

See `examples/` for examples of using the library. You can run a simple example webserver with
```
go install examples/simple/goflow-simple-example.go
eval "$GOPATH/bin/goflow-simple-example"
```

Then browse to `localhost:8100` to explore the UI, where you can submit jobs and view their current state.

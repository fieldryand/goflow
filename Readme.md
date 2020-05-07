[![Build Status](https://travis-ci.org/fieldryand/goflow.svg?branch=master)](https://travis-ci.org/fieldryand/goflow)
[![codecov](https://codecov.io/gh/fieldryand/goflow/branch/master/graph/badge.svg)](https://codecov.io/gh/fieldryand/goflow)
[![Go Report Card](https://goreportcard.com/badge/github.com/fieldryand/goflow)](https://goreportcard.com/report/github.com/fieldryand/goflow)

# Goflow

A minimal workflow scheduler written in Go, inspired by Apache Airflow and github.com/thieman/dagobah.

## Motivation

For personal projects, I want the cheapest architecture possible. I want to run Airflow on a .6GB-memory GCP f1-micro instance for a couple bucks a month, but Airflow running in a Docker container requires more memory than that. All my Airflow DAGS basically start a cluster or VM, send some requests to services that do the heavy compute, then stop the cluster/VM. I'd also like the option of an in-memory database, because I only need task logs for the last few days, and Airflow requires a Postgres instance (more $) if you want concurrent task execution (yes please, so I can stop that cluster sooner). Goflow is designed to meet these minimal requirements.

## Usage

See `examples/` for examples of using the library. You can run a simple example webserver with
```
go install examples/simple/goflow-simple-example.go
eval "$GOPATH/bin/goflow-simple-example"
```

Then browse to `localhost:8090` to explore the UI, where you can submit jobs and view their current state. 

## TODO

- http operator
- scheduling
- ...

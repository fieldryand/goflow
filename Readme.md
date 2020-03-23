# Goflow

A minimal workflow scheduler written in Go, inspired by Apache Airflow and github.com/thieman/dagobah.

## Motivation

For personal projects, I want the cheapest architecture possible. I want to run Airflow on a .6GB-memory GCP f1-micro instance for a couple bucks a month, but Airflow running in a Docker container requires more memory than that. All my Airflow DAGS basically start a cluster or VM, send some requests to services that do the heavy compute, then stop the cluster/VM. I'd also like the option of an in-memory database, because I only need task logs for the last few days, and Airflow requires a Postgres instance (more $) if you want concurrent task execution (yes please, so I can stop that cluster sooner). Goflow is designed to meet these minimal requirements.

## Usage

1. Clone the repo.
2. Create a Job in the `jobs` directory. See `example_job.go` for an example.
3. Add the job to `goflow.go`.
4. `go run goflow.go`

Output:
```
taskLogger:job.go:125: Task add 1 1 succeeded with result 2
taskLogger:job.go:125: Task add 3 4 succeeded with result 7
taskLogger:job.go:125: Task sleep 2 succeeded with result true
taskLogger:job.go:125: Task add 2 4 succeeded with result 6
```

## TODO

- webserver
- UI
- scheduling
- ...

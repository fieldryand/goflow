# Goflow

A minimal workflow scheduler written in Go, inspired by Apache Airflow and github.com/thieman/dagobah.

## Motivation

For personal projects, I want the cheapest architecture possible. I want to run Airflow on a .6GB-memory GCP f1-micro instance for a couple bucks a month, but Airflow running in a Docker container requires more memory than that. All my Airflow DAGS basically start a cluster or VM, send some requests to services that do the heavy compute, then stop the cluster/VM. I'd also like the option of an in-memory database, because I only need task logs for the last few days, and Airflow requires a Postgres instance (more $) if you want concurrent task execution (yes please, so I can stop that cluster sooner). Goflow is designed to meet these minimal requirements.

## Usage

1. Clone the repo.
2. To create a job: create a file in the `jobs` directory that matches the pattern `*_job.go`. In this file, declare a function that returns a `*core.Job`, and in the line above, write a comment that matches the pattern `// goflow: {{ function }} {{ job name }}`. See `example_job.go` for an example.
3. Run `go generate` to generate the `flow` function in `flow.go`. This function exposes the jobs.
4. Run the webserver with `go run .`.

Run the demo with `./demo.sh`.

Demo client output:
```
job submitted
{"add 1 1":"None","add 2 4":"None","add 3 4":"None","sleep 1":"None","sleep 2":"None"}
{"add 1 1":"Success","add 2 4":"None","add 3 4":"Success","sleep 1":"Success","sleep 2":"Running"}
{"add 1 1":"Success","add 2 4":"Success","add 3 4":"Success","sleep 1":"Success","sleep 2":"Success"}
```

`go generate` output:
```
2020/03/30 21:41:35 Found job example (ExampleJob) in ./jobs/example_job.go
```

## TODO

- UI
- scheduling
- ...

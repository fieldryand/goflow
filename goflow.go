// Package goflow implements a simple but powerful DAG scheduler and dashboard.
package goflow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/gomap"
	"github.com/robfig/cron/v3"
)

// Goflow contains job data and a router.
type Goflow struct {
	Store       gokv.Store
	Options     Options
	Jobs        map[string](func() *Job)
	Router      *http.ServeMux
	cron        *cron.Cron
	jobs        []string
	cronEntries map[string]cron.EntryID
	queue       chan string
}

// Options to control various Goflow behavior.
type Options struct {
	Store        gokv.Store
	UIPath       string
	Streaming    bool
	ShowExamples bool
	WithSeconds  bool
}

// New returns a Goflow engine.
func New(opts Options) *Goflow {

	// Add a default store if necessary
	if opts.Store == nil {
		opts.Store = gomap.NewStore(gomap.DefaultOptions)
	}

	// Add the cron schedule
	var c *cron.Cron
	if opts.WithSeconds {
		c = cron.New(cron.WithSeconds())
	} else {
		c = cron.New()
	}

	g := &Goflow{
		Store:       opts.Store,
		Options:     opts,
		Jobs:        make(map[string](func() *Job)),
		Router:      http.NewServeMux(),
		cron:        c,
		cronEntries: make(map[string]cron.EntryID),
		queue:       make(chan string),
	}

	if opts.ShowExamples {
		err := g.AddJob(complexAnalyticsJob, randomFailureJob)
		if err != nil {
			log.Println("error adding example jobs")
		}
	}

	return g
}

// AddJob takes a job-emitting function and registers it
// with the engine.
func (g *Goflow) AddJob(jobFunc ...func() *Job) error {
	for _, k := range jobFunc {
		err := g.addJob(k)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Goflow) addJob(jobFunc func() *Job) error {

	j := jobFunc()

	// "" is not a valid key in the storage layer
	if j.Name == "" {
		return errors.New("\"\" is not a valid job name")
	}

	// Validate and register the job
	if !j.Dag.validate() {
		return fmt.Errorf("Invalid Dag for job %s", j.Name)
	}
	g.Jobs[j.Name] = jobFunc
	g.jobs = append(g.jobs, j.Name)

	// If the job is active by default, add it to the cron schedule
	if j.Active {
		entryID, err := g.cron.AddFunc(j.Schedule, func() { g.queue <- j.Name })
		if err != nil {
			return err
		}
		g.cronEntries[j.Name] = entryID
	}

	return nil
}

// toggle flips a job's cron schedule status from active to inactive
// and vice versa. It returns true if the new status is active and false
// if it is inactive.
func (g *Goflow) toggle(jobName string) (bool, error) {

	// if the job is found in the list of entries, remove it
	for job, entryID := range g.cronEntries {
		if job == jobName {
			g.cron.Remove(entryID)
			delete(g.cronEntries, job)
			return false, nil
		}
	}

	// else add a new entry
	jobFunc := g.Jobs[jobName]
	entryID, err := g.cron.AddFunc(jobFunc().Schedule, func() { g.queue <- jobName })
	if err != nil {
		return false, err
	}
	g.cronEntries[jobName] = entryID
	return true, nil
}

// Execute tells the engine to run a given job in a new goroutine.
func (g *Goflow) Execute(ctx context.Context, job string) (*uuid.UUID, error) {

	// find the job if it exists and create a new execution instance
	jobFunc, ok := g.Jobs[job]
	if !ok {
		return nil, fmt.Errorf("job %s does not exist", job)
	}
	j := jobFunc()
	e := j.newExecution()

	// write it to the storage layer
	err := persistNewExecution(g.Store, e)
	if err != nil {
		return &e.ID, err
	}
	err = indexExecutions(g.Store, e)
	if err != nil {
		return &e.ID, err
	}

	// start the job
	go j.run(ctx, g.Store, e)

	return &e.ID, nil
}

// Run is a blocking call that listens for jobs.
func (g *Goflow) Run(ctx context.Context) error {
	g.cron.Start()
	for {
		select {
		case <-ctx.Done():
			g.cron.Stop()
			log.Println("goflow error: context cancelled")
			return errors.New("context cancelled")
		default:
		}
		i := <-g.queue
		_, err := g.Execute(ctx, i)
		if err != nil {
			return err
		}
	}
}

// RunWithWebserver will listen for jobs and web requests.
func (g *Goflow) RunWithWebserver(ctx context.Context, port string) error {
	go func() {
		err := g.Run(ctx)
		if err != nil {
			log.Printf("goflow error: %v", err)
		}
	}()
	g.addRoutes()
	return http.ListenAndServe(port, g.Router)
}

// Package goflow implements a simple but powerful DAG scheduler and dashboard.
package goflow

import (
	"errors"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/gomap"
	"github.com/robfig/cron/v3"
)

// Goflow contains job data and a router.
type Goflow struct {
	Store            gokv.Store
	Options          Options
	Jobs             map[string](func() *Job)
	router           *gin.Engine
	cron             *cron.Cron
	activeJobCronIDs map[string]cron.EntryID
}

// Options to control various Goflow behavior.
type Options struct {
	Streaming    bool
	ShowExamples bool
	WithSeconds  bool
}

// New returns a Goflow engine.
func New(opts Options) *Goflow {
	var c *cron.Cron
	if opts.WithSeconds {
		c = cron.New(cron.WithSeconds())
	} else {
		c = cron.New()
	}

	g := &Goflow{
		Store:            gomap.NewStore(gomap.DefaultOptions),
		Options:          opts,
		Jobs:             make(map[string](func() *Job)),
		router:           gin.New(),
		cron:             c,
		activeJobCronIDs: make(map[string]cron.EntryID),
	}

	if opts.ShowExamples {
		g.Add(complexAnalyticsJob)
		g.Add(customOperatorJob)
	}

	return g
}

// AttachStore attaches a store.
func (g *Goflow) AttachStore(store gokv.Store) {
	g.Store = store
}

// Add takes a job-emitting function and registers it
// with the engine.
func (g *Goflow) Add(jobFn func() *Job) error {

	jobName := jobFn().Name

	if jobName == "" {
		return errors.New("\"\" is not a valid job name")
	}

	if !jobFn().Dag.validate() {
		return fmt.Errorf("Invalid Dag for job %s", jobName)
	}

	// Register the job
	g.Jobs[jobName] = jobFn

	// If the job is active by default, add it to the cron schedule
	if jobFn().Active {
		entryID, _ := g.cron.AddFunc(jobFn().Schedule, func() { g.runJob(jobName) })
		g.activeJobCronIDs[jobName] = entryID
	}

	return nil
}

// setUnsetActive takes a job-emitting function and modifies it so it emits
// jobs with the desired active value.
func setUnsetActive(fn func() *Job, active bool) func() *Job {
	return func() *Job {
		job := fn()
		job.Active = active
		return job
	}
}

// toggle flips a job's cron schedule status from active to inactive
// and vice versa. It returns true if the new status is active and false
// if it is inactive.
func (g *Goflow) toggle(jobName string) (bool, error) {
	if g.Jobs[jobName]().Active {
		g.Jobs[jobName] = setUnsetActive(g.Jobs[jobName], false)
		g.cron.Remove(g.activeJobCronIDs[jobName])
		delete(g.activeJobCronIDs, jobName)
		return false, nil
	}

	g.Jobs[jobName] = setUnsetActive(g.Jobs[jobName], true)
	entryID, _ := g.cron.AddFunc(g.Jobs[jobName]().Schedule, func() { g.runJob(jobName) })
	g.activeJobCronIDs[jobName] = entryID
	return true, nil
}

// runJob tells the engine to run a given job and returns
// the corresponding Execution.
func (g *Goflow) runJob(jobName string) {
	job := g.Jobs[jobName]()
	go job.run(g.Store)
}

// Use middleware in the Gin router.
func (g *Goflow) Use(middleware gin.HandlerFunc) gin.IRoutes {
	return g.router.Use(middleware)
}

// Run runs the webserver.
func (g *Goflow) Run(port string) error {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	g.router.Use(gin.Recovery())
	g.addStreamRoute()
	g.addAPIRoutes()
	g.addStaticRoutes()
	g.addUIRoutes()
	g.cron.Start()
	return g.router.Run(port)
}

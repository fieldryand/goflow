// Package goflow implements a simple but powerful DAG scheduler and dashboard.
package goflow

import (
	"errors"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	Store        gokv.Store // deprecated, use AttachStore()
	UIPath       string
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
		g.RegisterJob(exampleAnalytics)
		g.RegisterJob(exampleHTTP)
	}

	return g
}

// AttachStore attaches a store.
func (g *Goflow) AttachStore(store gokv.Store) {
	g.Store = store
}

// RegisterJob takes a job-emitting function and registers it
// with the engine.
func (g *Goflow) RegisterJob(jobFn func() *Job) error {

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
		entryID, _ := g.cron.AddFunc(jobFn().Schedule, func() { g.scheduledExecute(jobName) })
		g.activeJobCronIDs[jobName] = entryID
	}

	return nil
}

// AddJob (DEPRECATED) takes a job-emitting function and registers it
// with the engine.
func (g *Goflow) AddJob(jobFn func() *Job) *Goflow {

	jobName := jobFn().Name

	// TODO: change the return type here to error
	// "" is not a valid key in the storage layer
	//if jobName == "" {
	//		return errors.New("\"\" is not a valid job name")
	//	}

	// Register the job
	g.Jobs[jobName] = jobFn

	// If the job is active by default, add it to the cron schedule
	if jobFn().Active {
		entryID, _ := g.cron.AddFunc(jobFn().Schedule, func() { g.scheduledExecute(jobName) })
		g.activeJobCronIDs[jobName] = entryID
	}

	return g
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
	entryID, _ := g.cron.AddFunc(g.Jobs[jobName]().Schedule, func() { g.scheduledExecute(jobName) })
	g.activeJobCronIDs[jobName] = entryID
	return true, nil
}

// Execute tells the engine to run a given job in a goroutine.
// The job state is readable from the engine's store.
func (g *Goflow) Execute(job string) uuid.UUID {

	// create job
	j := g.Jobs[job]()

	// create and persist a new execution
	e := j.newExecution(false)
	persistNewExecution(g.Store, e)
	indexExecutions(g.Store, e)

	// execute
	go j.run(g.Store, e)

	return e.ID
}

func (g *Goflow) scheduledExecute(job string) uuid.UUID {

	// create job
	j := g.Jobs[job]()

	// create and persist a new execution
	e := j.newExecution(true)
	persistNewExecution(g.Store, e)
	indexExecutions(g.Store, e)

	// execute
	go j.run(g.Store, e)

	return e.ID
}

// Use middleware in the Gin router.
func (g *Goflow) Use(middleware gin.HandlerFunc) *Goflow {
	g.router.Use(middleware)
	return g
}

// Run runs the webserver.
func (g *Goflow) Run(port string) {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	g.router.Use(gin.Recovery())
	g.addStreamRoute()
	g.addAPIRoutes()
	if g.Options.UIPath != "" {
		g.addUIRoutes()
		g.addStaticRoutes()
	}
	g.cron.Start()
	g.router.Run(port)
}

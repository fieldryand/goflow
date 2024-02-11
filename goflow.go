// Package goflow implements a simple but powerful DAG scheduler and dashboard.
package goflow

import (
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
	jobs             []string
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
		Store:            opts.Store,
		Options:          opts,
		Jobs:             make(map[string](func() *Job)),
		router:           gin.New(),
		cron:             c,
		activeJobCronIDs: make(map[string]cron.EntryID),
	}

	if opts.ShowExamples {
		g.AddJob(complexAnalyticsJob)
		g.AddJob(customOperatorJob)
	}

	return g
}

// AddJob takes a job-emitting function and registers it
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
	g.jobs = append(g.jobs, jobName)

	// If the job is active by default, add it to the cron schedule
	if jobFn().Active {
		entryID, _ := g.cron.AddFunc(jobFn().Schedule, func() { g.executeScheduled(jobName) })
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
	entryID, _ := g.cron.AddFunc(g.Jobs[jobName]().Schedule, func() { g.executeScheduled(jobName) })
	g.activeJobCronIDs[jobName] = entryID
	return true, nil
}

// executeScheduled tells the engine to run a given job. The cron scheduler will
// run it in a new goroutine.
func (g *Goflow) executeScheduled(job string) uuid.UUID {

	// create job
	j := g.Jobs[job]()

	// create and persist a new execution
	e := j.newExecution()
	persistNewExecution(g.Store, e)
	indexExecutions(g.Store, e)

	// start running the job
	j.run(g.Store, e)

	return e.ID
}

// execute tells the engine to run a given job in a new goroutine.
func (g *Goflow) execute(job string) uuid.UUID {

	// create job
	j := g.Jobs[job]()

	// create and persist a new execution
	e := j.newExecution()
	persistNewExecution(g.Store, e)
	indexExecutions(g.Store, e)

	// start running the job
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

// Package goflow implements a simple but powerful DAG scheduler and dashboard.
package goflow

import (
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
	Store        gokv.Store
	UIPath       string
	Streaming    bool
	ShowExamples bool
}

// New returns a Goflow engine.
func New(opts Options) *Goflow {
	if opts.Store == nil {
		storeOptions := gomap.DefaultOptions
		opts.Store = gomap.NewStore(storeOptions)
	}

	g := &Goflow{
		Store:            opts.Store,
		Options:          opts,
		Jobs:             make(map[string](func() *Job)),
		router:           gin.New(),
		cron:             cron.New(),
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
	g.Jobs[jobFn().Name] = jobFn

	if jobFn().Active {
		entryID, _ := g.cron.AddFunc(jobFn().Schedule, func() { g.runJob(jobFn().Name) })
		g.activeJobCronIDs[jobFn().Name] = entryID
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
	entryID, _ := g.cron.AddFunc(g.Jobs[jobName]().Schedule, func() { g.runJob(jobName) })
	g.activeJobCronIDs[jobName] = entryID
	return true, nil
}

// runJob tells the engine to run a given job and returns
// the corresponding jobRun.
func (g *Goflow) runJob(jobName string) *jobRun {
	// generate the job
	job := g.Jobs[jobName]()

	// create and persist a new jobrun record
	jobrun := job.newJobRun()
	persistNewJobRun(g.Store, jobrun)

	// start running the job
	go job.run()

	// in parallel, keep syncing the job state to the store
	go func() {
		for {
			// get the current state
			jobState := job.getJobState()

			// sync to the store
			updateJobState(g.Store, jobrun, jobState)

			// stop syncing when the job is done
			if jobState.State != running && jobState.State != none {
				log.Printf("job <%v> reached state <%v>", job.Name, job.jobState.State)
				break
			}
		}
	}()

	return jobrun
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

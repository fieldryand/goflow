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
	Store   gokv.Store
	Options Options
	Jobs    map[string](func() *Job)
	router  *gin.Engine
	cron    *cron.Cron
	jobs    []string
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
		Store:   opts.Store,
		Options: opts,
		Jobs:    make(map[string](func() *Job)),
		router:  gin.New(),
		cron:    c,
	}

	if opts.ShowExamples {
		g.AddJob(complexAnalyticsJob)
		g.AddJob(customOperatorJob)
	}

	return g
}

// scheduledExecution implements cron.Job
type scheduledExecution struct {
	store   gokv.Store
	jobFunc func() *Job
}

func (schedExec *scheduledExecution) Run() {

	// create job
	job := schedExec.jobFunc()

	// create and persist a new execution
	e := job.newExecution()
	persistNewExecution(schedExec.store, e)
	indexExecutions(schedExec.store, e)

	// start running the job
	job.run(schedExec.store, e)
}

// AddJob takes a job-emitting function and registers it
// with the engine.
func (g *Goflow) AddJob(jobFunc func() *Job) *Goflow {

	j := jobFunc()

	// TODO: change the return type here to error
	// "" is not a valid key in the storage layer
	//if j.Name == "" {
	//		return errors.New("\"\" is not a valid job name")
	//	}

	// Register the job
	g.Jobs[j.Name] = jobFunc
	g.jobs = append(g.jobs, j.Name)

	// If the job is active by default, add it to the cron schedule
	if j.Active {
		e := &scheduledExecution{g.Store, jobFunc}
		_, err := g.cron.AddJob(j.Schedule, e)

		if err != nil {
			panic(err)
		}
	}

	return g
}

// toggle flips a job's cron schedule status from active to inactive
// and vice versa. It returns true if the new status is active and false
// if it is inactive.
func (g *Goflow) toggle(jobName string) (bool, error) {

	// if the job is found in the list of entries, remove it
	for _, entry := range g.cron.Entries() {
		if name := entry.Job.(*scheduledExecution).jobFunc().Name; name == jobName {
			g.cron.Remove(entry.ID)
			return false, nil
		}
	}

	// else add a new entry
	jobFunc := g.Jobs[jobName]
	e := &scheduledExecution{g.Store, jobFunc}
	g.cron.AddJob(jobFunc().Schedule, e)
	return true, nil
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
	g.addStreamRoute(true)
	g.addAPIRoutes()
	if g.Options.UIPath != "" {
		g.addUIRoutes()
		g.addStaticRoutes()
	}
	g.cron.Start()
	g.router.Run(port)
}

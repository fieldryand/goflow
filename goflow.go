// Package goflow implements a simple but powerful DAG scheduler and dashboard that is easy to set up and integrate with other applications.
package goflow

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	bolt "go.etcd.io/bbolt"
)

// Goflow contains job data and a router.
type Goflow struct {
	Options          Options
	Jobs             map[string](func() *Job)
	router           *gin.Engine
	cron             *cron.Cron
	activeJobCronIDs map[string]cron.EntryID
	db               database
}

// Options to control various Goflow behavior.
type Options struct {
	DBType        string
	BoltDBPath    string
	UIPath        string
	StreamJobRuns bool
	ShowExamples  bool
}

// New returns a Goflow engine.
func New(opts Options) *Goflow {
	if opts.DBType == "" {
		opts.DBType = "boltdb"
	}
	if opts.BoltDBPath == "" {
		opts.BoltDBPath = "goflow.db"
	}

	g := &Goflow{
		Options:          opts,
		Jobs:             make(map[string](func() *Job)),
		router:           gin.New(),
		cron:             cron.New(),
		activeJobCronIDs: make(map[string]cron.EntryID),
	}

	if opts.DBType == "boltdb" {
		g.initializeBoltDB()
	} else {
		g.initializeMemoryDB()
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
	job := g.Jobs[jobName]()
	jr := job.newJobRun()
	g.db.writeJobRun(jr)

	go job.run()
	go func() {
		for {
			jobState := job.getJobState()
			g.db.updateJobState(jr, jobState)
			if jobState.State != running && jobState.State != none {
				log.Printf("job <%v> reached state <%v>", job.Name, job.jobState.State)
				break
			}
		}
	}()

	return jr
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
	g.addStaticRoutes()
	g.addStreamRoute()
	g.addUIRoutes()
	g.addAPIRoutes()
	g.cron.Start()
	g.router.Run(port)
}

func (g *Goflow) initializeMemoryDB() *Goflow {
	g.db = &memoryDB{make([]*jobRun, 0)}
	return g
}

func (g *Goflow) initializeBoltDB() *Goflow {
	db, err := bolt.Open(g.Options.BoltDBPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(jobRunBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	g.db = &boltDB{db}
	return g
}

func (g *Goflow) getJobRuns(clientDisconnect bool) func(*gin.Context) {
	return func(c *gin.Context) {
		chanStream := make(chan string)

		go func() {
			defer close(chanStream)

			// Periodically push the list of job runs into the stream
			for {
				for job := range g.Jobs {
					jrl, _ := g.db.readJobRuns(job)
					marshalled, _ := marshalJobRunList(jrl)
					chanStream <- string(marshalled)
				}
				time.Sleep(time.Second * 1)
			}
		}()

		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-chanStream; ok {
				c.SSEvent("message", msg)
				return clientDisconnect
			}
			return false
		})
	}
}

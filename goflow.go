// Package goflow implements a web UI-based workflow orchestrator
// inspired by Apache Airflow.
package goflow

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	StreamJobRuns bool
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

// toggleActive flips a job's cron schedule status from active to inactive
// and vice versa. It returns true if the new status is active and false
// if it is inactive.
func (g *Goflow) toggleActive(jobName string) (bool, error) {
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
				log.Printf("job %v reached state %v", job.Name, job.jobState.State)
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
	g.router.Use(gin.Recovery())
	g.addRoutes()
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

func (g *Goflow) addStaticRoutes() *Goflow {
	goPath := os.Getenv("GOPATH")
	assetPath := goPath + "/src/github.com/fieldryand/goflow/assets/"
	g.router.Static("/css", assetPath+"css")
	g.router.Static("/dist", assetPath+"dist")
	g.router.Static("/src", assetPath+"src")
	g.router.LoadHTMLGlob(assetPath + "html/*.html.tmpl")
	return g
}

func (g *Goflow) addRoutes() *Goflow {
	g.addStaticRoutes()

	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	g.router.GET("/", func(c *gin.Context) {
		jobs := make([]*Job, 0)
		for _, job := range g.Jobs {
			jobs = append(jobs, job())
		}
		c.HTML(http.StatusOK, "index.html.tmpl", gin.H{"jobs": jobs})
	})

	g.router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	g.router.GET("/jobs", func(c *gin.Context) {
		jobNames := make([]string, 0)
		for _, job := range g.Jobs {
			jobNames = append(jobNames, job().Name)
		}
		c.JSON(http.StatusOK, jobNames)
	})

	g.router.GET("/jobs/:name", func(c *gin.Context) {
		name := c.Param("name")
		jobFn, ok := g.Jobs[name]

		if ok {
			tasks := jobFn().Tasks
			taskNames := make([]string, 0)
			for _, task := range tasks {
				taskNames = append(taskNames, task.Name)
			}

			c.HTML(http.StatusOK, "job.html.tmpl", gin.H{
				"jobName":   name,
				"taskNames": taskNames,
				"schedule":  g.Jobs[name]().Schedule,
			})
		} else {
			c.String(http.StatusNotFound, "Not found")
		}

	})

	g.router.POST("/jobs/:name/submit", func(c *gin.Context) {
		name := c.Param("name")
		_, ok := g.Jobs[name]

		if ok {
			jobRun := g.runJob(name)
			c.String(http.StatusOK, fmt.Sprintf("submitted job run %s", jobRun.name()))
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

	g.router.POST("/jobs/:name/toggleActive", func(c *gin.Context) {
		name := c.Param("name")
		_, ok := g.Jobs[name]

		if ok {
			isActive, _ := g.toggleActive(name)
			c.String(http.StatusOK, fmt.Sprintf("job %s set to active=%v", name, isActive))
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

	g.router.GET("/jobs/:name/isActive", func(c *gin.Context) {
		name := c.Param("name")
		jobFn, ok := g.Jobs[name]

		if ok {
			c.JSON(http.StatusOK, jobFn().Active)
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

	g.router.GET("/stream", g.getJobRuns(g.Options.StreamJobRuns))

	g.router.GET("/jobs/:name/dag", func(c *gin.Context) {
		name := c.Param("name")
		jobFn, ok := g.Jobs[name]

		if ok {
			c.JSON(http.StatusOK, jobFn().Dag)
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

	return g
}

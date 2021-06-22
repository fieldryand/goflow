// Package goflow implements a workflow scheduler geared
// toward orchestration of ETL or analytics workloads.
package goflow

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

// Goflow contains job data and a router.
type Goflow struct {
	Jobs             map[string](func() *Job)
	jobRuns          []*jobRun
	router           *gin.Engine
	cron             *cron.Cron
	activeJobFlags   map[string]bool
	activeJobCronIDs map[string]cron.EntryID
}

// New returns a Goflow engine.
func New() *Goflow {
	return &Goflow{
		Jobs:             make(map[string](func() *Job)),
		jobRuns:          make([]*jobRun, 0),
		router:           gin.New(),
		cron:             cron.New(),
		activeJobFlags:   make(map[string]bool),
		activeJobCronIDs: make(map[string]cron.EntryID),
	}
}

// AddJob takes a job-emitting function and registers it
// with the engine.
func (g *Goflow) AddJob(jobFn func() *Job) *Goflow {
	g.Jobs[jobFn().Name] = jobFn

	if jobFn().ActiveByDefault {
		entryID, _ := g.cron.AddFunc(jobFn().Schedule, func() { g.runJob(jobFn().Name) })
		g.activeJobCronIDs[jobFn().Name] = entryID
		g.activeJobFlags[jobFn().Name] = true
	} else {
		g.activeJobFlags[jobFn().Name] = false
	}

	return g
}

// toggleActive flips a job's cron schedule status from active to inactive
// and vice versa. It returns true if the new status is active and false
// if it is inactive.
func (g *Goflow) toggleActive(jobName string) (bool, error) {
	if g.activeJobFlags[jobName] {
		g.cron.Remove(g.activeJobCronIDs[jobName])
		delete(g.activeJobCronIDs, jobName)
		g.activeJobFlags[jobName] = false
		return false, nil
	}

	jobFn := g.Jobs[jobName]
	entryID, _ := g.cron.AddFunc(jobFn().Schedule, func() { g.runJob(jobName) })
	g.activeJobCronIDs[jobName] = entryID
	g.activeJobFlags[jobName] = true
	return true, nil
}

// runJob tells the engine to run a given job and returns
// the corresponding jobRun.
func (g *Goflow) runJob(jobName string) *jobRun {
	job := g.Jobs[jobName]()
	jobRun := newJobRun(jobName)

	g.jobRuns = append(g.jobRuns, jobRun)
	reads := make(chan readOp)

	go job.run(reads)
	go func() {
		read := readOp{resp: make(chan *jobState), allDone: job.allDone()}
		reads <- read
		for _, jr := range g.jobRuns {
			if jr.name() == jobRun.name() {
				jr.JobState = <-read.resp
			}
		}
	}()

	return jobRun
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

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format(time.RFC3339) + " [GOFLOW] - " + string(bytes))
}

func (g *Goflow) addRoutes() *Goflow {
	goPath := os.Getenv("GOPATH")
	assetPath := goPath + "/src/github.com/fieldryand/goflow/assets/"
	g.router.Static("/css", assetPath+"css")
	g.router.Static("/dist", assetPath+"dist")
	g.router.Static("/src", assetPath+"src")
	g.router.LoadHTMLGlob(assetPath + "html/*.html.tmpl")

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
			c.String(http.StatusOK, fmt.Sprintf("isActive flag for job %s set to %v", name, isActive))
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

	g.router.GET("/jobs/:name/isActive", func(c *gin.Context) {
		name := c.Param("name")
		isActive, ok := g.activeJobFlags[name]

		if ok {
			c.JSON(http.StatusOK, isActive)
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

	g.router.GET("/jobs/:name/jobRuns", func(c *gin.Context) {
		name := c.Param("name")
		_, ok := g.Jobs[name]

		if ok {
			jobRunList := newJobRunList(name, g.jobRuns)
			c.JSON(http.StatusOK, jobRunList)
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

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

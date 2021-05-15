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
)

// Engine contains job data and a router.
type Engine struct {
	Jobs    map[string](func() *Job)
	jobRuns []*jobRun
	router  *gin.Engine
}

// NewEngine returns a Goflow engine.
func NewEngine() *Engine {
	return &Engine{
		Jobs:    make(map[string](func() *Job)),
		jobRuns: make([]*jobRun, 0),
		router:  gin.New(),
	}
}

// AddJob takes a job-emitting function and registers it
// with the engine.
func (g *Engine) AddJob(jobFn func() *Job) *Engine {
	g.Jobs[jobFn().Name] = jobFn
	return g
}

// Use middleware in the Gin router.
func (g *Engine) Use(middleware gin.HandlerFunc) *Engine {
	g.router.Use(middleware)
	return g
}

// Run runs the webserver.
func (g *Engine) Run(port string) {
	g.addRoutes()
	g.router.Run(port)
}

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format(time.RFC3339) + " [GOFLOW] - " + string(bytes))
}

func (g *Engine) addRoutes() *Engine {
	goPath := os.Getenv("GOPATH")
	assetPath := goPath + "/src/github.com/fieldryand/goflow/assets/"
	g.router.Static("/css", assetPath+"css")
	g.router.Static("/dist", assetPath+"dist")
	g.router.Static("/src", assetPath+"src")
	g.router.LoadHTMLGlob(assetPath + "html/*.html.tmpl")

	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	g.router.GET("/", func(c *gin.Context) {
		jobNames := make([]string, 0)
		for _, job := range g.Jobs {
			jobNames = append(jobNames, job().Name)
		}

		c.HTML(http.StatusOK, "index.html.tmpl", gin.H{
			"jobNames": jobNames,
		})
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

		tasks := g.Jobs[name]().Tasks
		taskNames := make([]string, 0)
		for _, task := range tasks {
			taskNames = append(taskNames, task.Name)
		}

		jobRuns := make([]*jobRun, 0)
		for _, jr := range g.jobRuns {
			if jr.JobName == name {
				jobRuns = append(jobRuns, jr)
			}
		}

		c.HTML(http.StatusOK, "job.html.tmpl", gin.H{
			"jobName":   name,
			"taskNames": taskNames,
			"jobRuns":   jobRuns,
		})
	})

	g.router.POST("/jobs/:name/submit", func(c *gin.Context) {
		name := c.Param("name")
		job := g.Jobs[name]()
		jobRun := newJobRun(name)

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
		c.String(http.StatusOK, fmt.Sprintf("submitted job run %s", jobRun.name()))
	})

	g.router.GET("/jobs/:name/jobRuns", func(c *gin.Context) {
		name := c.Param("name")
		jobRunList := newJobRunList(name, g.jobRuns)
		c.JSON(http.StatusOK, jobRunList)
	})

	g.router.GET("/jobs/:name/dag", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, g.Jobs[name]().Dag)
	})

	return g
}

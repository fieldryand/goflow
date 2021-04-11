// Package goflow implements a workflow scheduler geared
// toward orchestration of ETL or analytics workloads.
package goflow

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Engine contains job data and a router.
type Engine struct {
	jobMap  map[string](func() *Job)
	jobRuns []*jobRun
	router  *gin.Engine
}

// NewEngine returns a Goflow engine.
func NewEngine(jobs ...func() *Job) *Engine {
	jobMap := make(map[string](func() *Job))

	for _, job := range jobs {
		jobMap[job().Name] = job
	}

	router := gin.New()

	g := Engine{
		jobMap:  jobMap,
		jobRuns: make([]*jobRun, 0),
		router:  router,
	}

	return &g
}

// Use passes middleware to the Gin router.
func (g *Engine) Use(middleware gin.HandlerFunc) *Engine {
	g.router.Use(middleware)
	return g
}

// Run runs the webserver.
func (g *Engine) Run(port string) {
	g.addRoutes()
	g.router.Run(port)
}

func (g *Engine) addRoutes() *Engine {
	goPath := os.Getenv("GOPATH")
	assetPath := "/src/github.com/fieldryand/goflow/assets/*.html.tmpl"

	g.router.Static("/static", "assets/static")
	g.router.LoadHTMLGlob(goPath + assetPath)

	g.router.GET("/", func(c *gin.Context) {
		jobNames := make([]string, 0)
		for _, job := range g.jobMap {
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
		for _, job := range g.jobMap {
			jobNames = append(jobNames, job().Name)
		}
		c.JSON(http.StatusOK, jobNames)
	})

	g.router.GET("/jobs/:name", func(c *gin.Context) {
		name := c.Param("name")

		jobRuns := make([]*jobRun, 0)
		for _, jr := range g.jobRuns {
			if jr.JobName == name {
				jobRuns = append(jobRuns, jr)
			}
		}

		c.HTML(http.StatusOK, "job.html.tmpl", gin.H{
			"jobName": name,
			"jobRuns": jobRuns,
		})
	})

	g.router.GET("/jobs/:name/submit", func(c *gin.Context) {
		name := c.Param("name")
		job := g.jobMap[name]()
		jobRun := newJobRun(name)

		g.jobRuns = append(g.jobRuns, jobRun)

		reads := make(chan readOp)
		go job.run(reads)
		go func() {
			read := readOp{resp: make(chan *jobState)}
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
		c.JSON(http.StatusOK, g.jobMap[name]().Dag)
	})

	return g
}

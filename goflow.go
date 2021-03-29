//Package goflow implements a minimal workflow scheduler.
package goflow

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Goflow struct {
	jobNames  []string
	jobMap    map[string](func() *Job)
	jobStates map[string]*jobState
	router    *gin.Engine
}

// New returns a Goflow engine.
func New(jobs ...func() *Job) *Goflow {
	jobNames := make([]string, 0)
	jobMap := make(map[string](func() *Job))
	jobStates := make(map[string]*jobState)

	for _, job := range jobs {
		jobNames = append(jobNames, job().Name)
		jobMap[job().Name] = job
		jobStates[job().Name] = newJobState()
	}

	router := gin.New()

	g := Goflow{
		jobNames:  jobNames,
		jobMap:    jobMap,
		jobStates: jobStates,
		router:    router,
	}

	return &g
}

func (g *Goflow) Use(middleware gin.HandlerFunc) *Goflow {
	g.router.Use(middleware)
	return g
}

func (g *Goflow) Run(port string) {
	g.addRoutes()
	g.router.Run(port)
}

func (g *Goflow) addRoutes() *Goflow {
	goPath := os.Getenv("GOPATH")
	assetPath := "/src/github.com/fieldryand/goflow/assets/*.html.tmpl"

	g.router.Static("/static", "assets/static")
	g.router.LoadHTMLGlob(goPath + assetPath)

	g.router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html.tmpl", gin.H{
			"jobStates": g.jobStates,
		})
	})

	g.router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	g.router.GET("/jobs", func(c *gin.Context) {
		encoded, _ := json.Marshal(g.jobNames)
		c.String(http.StatusOK, string(encoded))
	})

	g.router.GET("/jobs/:name", func(c *gin.Context) {
		name := c.Param("name")

		c.HTML(http.StatusOK, "job.html.tmpl", gin.H{
			"jobName":   name,
			"jobStates": g.jobStates[name],
		})
	})

	g.router.GET("/jobs/:name/submit", func(c *gin.Context) {
		name := c.Param("name")
		job := g.jobMap[name]()
		g.jobStates[name] = job.jobState
		reads := make(chan readOp)
		go job.run(reads)
		go func() {
			read := readOp{resp: make(chan *jobState)}
			reads <- read
			g.jobStates[name] = <-read.resp
		}()
		c.String(http.StatusOK, "job submitted")
	})

	g.router.GET("/jobs/:name/state", func(c *gin.Context) {
		name := c.Param("name")
		encoded, _ := json.Marshal(g.jobStates[name])
		c.String(http.StatusOK, string(encoded))
	})

	g.router.GET("/jobs/:name/dag", func(c *gin.Context) {
		name := c.Param("name")
		encoded, _ := json.Marshal(g.jobMap[name]().Dag)
		c.String(http.StatusOK, string(encoded))
	})

	return g
}

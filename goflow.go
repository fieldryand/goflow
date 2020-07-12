//Package goflow implements a minimal workflow scheduler.
package goflow

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Goflow returns the application router.
// BUG(rf): no validation on Job or Task names to prevent collision.
// BUG(rf): no validation on parameters to ensure the job exists.
func Goflow(jobs map[string](func() *Job), middleware [](func() gin.HandlerFunc)) *gin.Engine {

	jobNames := make([]string, 0)
	for name := range jobs {
		jobNames = append(jobNames, name)
	}

	js := make(map[string]*jobState)

	for j := range jobs {
		js[j] = newJobState()
	}

	router := gin.New()
	router.Use(gin.Recovery())

	for _, m := range middleware {
		router.Use(m())
	}

	router.Static("/static", "assets/static")

	goPath := os.Getenv("GOPATH")
	assetPath := "/src/github.com/fieldryand/goflow/assets/*.html.tmpl"
	router.LoadHTMLGlob(goPath + assetPath)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html.tmpl", gin.H{
			"js": js,
		})
	})

	router.GET("/jobs", func(c *gin.Context) {
		encoded, _ := json.Marshal(jobNames)
		c.String(http.StatusOK, string(encoded))
	})

	router.GET("/jobs/:name", func(c *gin.Context) {
		name := c.Param("name")

		c.HTML(http.StatusOK, "job.html.tmpl", gin.H{
			"jobName": name,
			"js":      js[name],
		})
	})

	router.GET("/jobs/:name/submit", func(c *gin.Context) {
		name := c.Param("name")
		job := jobs[name]()
		js[name] = job.jobState
		reads := make(chan readOp)
		go job.run(reads)
		go func() {
			read := readOp{resp: make(chan *jobState)}
			reads <- read
			js[name] = <-read.resp
		}()
		c.String(http.StatusOK, "job submitted")
	})

	router.GET("/jobs/:name/state", func(c *gin.Context) {
		name := c.Param("name")
		encoded, _ := json.Marshal(js[name])
		c.String(http.StatusOK, string(encoded))
	})

	router.GET("/jobs/:name/dag", func(c *gin.Context) {
		name := c.Param("name")
		encoded, _ := json.Marshal(jobs[name]().Dag)
		c.String(http.StatusOK, string(encoded))
	})

	return router
}

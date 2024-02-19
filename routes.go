package goflow

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (g *Goflow) addStaticRoutes() {
	g.router.Static("/css", g.Options.UIPath+"css")
	g.router.Static("/dist", g.Options.UIPath+"dist")
	g.router.Static("/src", g.Options.UIPath+"src")
	g.router.LoadHTMLGlob(g.Options.UIPath + "html/*.html.tmpl")
}

func (g *Goflow) addStreamRoute(keepOpen bool) {
	g.router.GET("/stream", g.stream(keepOpen))
}

func (g *Goflow) addUIRoutes() {
	ui := g.router.Group("/ui")
	{
		ui.GET("/", g.handleRoot())
		ui.GET("/jobs/:name", g.handleJobsPage())
	}

	g.router.GET("/", handleRedirect)
}

func (g *Goflow) addAPIRoutes() {
	api := g.router.Group("/api")
	{
		api.GET("/health", handleHealth)
		api.GET("/jobs", g.handleJobs())
		api.GET("/jobs/:name", g.handleParameterizedJobs())
		api.POST("/jobs/:name/submit", g.handleSubmittedJobs())
		api.POST("/jobs/:name/toggle", g.handleToggledJobs())
		api.GET("/executions", g.handleExecutions())
	}
}

func handleHealth(c *gin.Context) {
	var msg struct {
		Health string `json:"health"`
	}
	msg.Health = "OK"
	c.JSON(http.StatusOK, msg)
}

func handleRedirect(c *gin.Context) {
	c.Redirect(http.StatusFound, "/ui/")
}

func (g *Goflow) handleJobs() gin.HandlerFunc {
	return func(c *gin.Context) {
		var msg struct {
			Jobs []string `json:"jobs"`
		}
		msg.Jobs = g.jobs
		c.JSON(http.StatusOK, msg)
	}
}

func (g *Goflow) handleParameterizedJobs() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		jobFn, ok := g.Jobs[name]

		var msg struct {
			JobName   string   `json:"job"`
			TaskNames []string `json:"tasks"`
			Dag       dag      `json:"dag"`
			Schedule  string   `json:"schedule"`
			Active    bool     `json:"active"`
		}

		if ok {
			msg.JobName = name
			msg.TaskNames = jobFn().tasks
			msg.Dag = jobFn().Dag
			msg.Schedule = g.Jobs[name]().Schedule

			// check if the job is active by looking in the list of cron entries
			for _, entry := range g.cron.Entries() {
				if jobName := entry.Job.(*scheduledExecution).jobFunc().Name; name == jobName {
					msg.Active = true
				}
			}

			c.JSON(http.StatusOK, msg)
		} else {
			c.JSON(http.StatusNotFound, msg)
		}
	}
}

func (g *Goflow) handleSubmittedJobs() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		_, ok := g.Jobs[name]

		var msg struct {
			Job       string `json:"job"`
			Success   bool   `json:"success"`
			Submitted string `json:"submitted"`
		}
		msg.Job = name

		if ok {
			g.execute(name)
			msg.Success = true
			msg.Submitted = time.Now().UTC().Format(time.RFC3339Nano)
			c.JSON(http.StatusOK, msg)
		} else {
			msg.Success = false
			c.JSON(http.StatusNotFound, msg)
		}
	}
}

func (g *Goflow) handleToggledJobs() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		_, ok := g.Jobs[name]

		var msg struct {
			Job     string `json:"job"`
			Success bool   `json:"success"`
			Active  bool   `json:"active"`
		}
		msg.Job = name

		if ok {
			isActive, _ := g.toggle(name)
			msg.Success = true
			msg.Active = isActive
			c.JSON(http.StatusOK, msg)
		} else {
			msg.Success = false
			c.JSON(http.StatusNotFound, msg)
		}
	}
}

func (g *Goflow) handleExecutions() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobName := c.Query("jobname")
		stateQuery := c.Query("state")

		executions := make([]*Execution, 0)

		for job := range g.Jobs {
			stored, _ := readExecutions(g.Store, job)
			for _, execution := range stored {
				if stateQuery != "" && stateQuery != string(execution.State) {
				} else if jobName != "" && jobName != execution.JobName {
				} else {
					executions = append(executions, execution)
				}
			}
		}

		var msg struct {
			Executions []*Execution `json:"executions"`
		}
		msg.Executions = executions

		c.JSON(http.StatusOK, msg)
	}
}

func (g *Goflow) handleRoot() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs := make([]*Job, 0)
		for _, job := range g.jobs {

			// create the job, assume it's inactive
			j := g.Jobs[job]()
			j.Active = false

			// check if the job is active by looking in the list of cron entries
			for _, entry := range g.cron.Entries() {
				if name := entry.Job.(*scheduledExecution).jobFunc().Name; name == j.Name {
					j.Active = true
				}
			}

			jobs = append(jobs, j)
		}
		c.HTML(http.StatusOK, "index.html.tmpl", gin.H{"jobs": jobs})
	}
}

func (g *Goflow) handleJobsPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		jobFn, ok := g.Jobs[name]

		if ok {
			c.HTML(http.StatusOK, "job.html.tmpl", gin.H{
				"jobName":   name,
				"taskNames": jobFn().tasks,
				"schedule":  g.Jobs[name]().Schedule,
			})
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	}
}

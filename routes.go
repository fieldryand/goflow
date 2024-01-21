package goflow

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed ui/*
var f embed.FS // nolint: unused

func (g *Goflow) addStreamRoute() *Goflow {
	g.router.GET("/stream", g.stream(g.Options.Streaming))
	return g
}

func (g *Goflow) addAPIRoutes() *Goflow {
	api := g.router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			var msg struct {
				Health string `json:"health"`
			}
			msg.Health = "OK"
			c.JSON(http.StatusOK, msg)
		})

		api.GET("/jobs", func(c *gin.Context) {
			jobNames := make([]string, 0)
			for _, job := range g.Jobs {
				jobNames = append(jobNames, job().Name)
			}
			var msg struct {
				Jobs []string `json:"jobs"`
			}
			msg.Jobs = jobNames
			c.JSON(http.StatusOK, msg)
		})

		api.GET("/executions", func(c *gin.Context) {
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
		})

		api.GET("/jobs/:name", func(c *gin.Context) {
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
				tasks := jobFn().Tasks
				taskNames := make([]string, 0)
				for _, task := range tasks {
					taskNames = append(taskNames, task.Name)
				}

				msg.JobName = name
				msg.TaskNames = taskNames
				msg.Dag = jobFn().Dag
				msg.Schedule = g.Jobs[name]().Schedule
				msg.Active = jobFn().Active

				c.JSON(http.StatusOK, msg)
			} else {
				c.JSON(http.StatusNotFound, msg)
			}
		})

		api.POST("/jobs/:name/submit", func(c *gin.Context) {
			name := c.Param("name")
			_, ok := g.Jobs[name]

			var msg struct {
				Job     string `json:"job"`
				Success bool   `json:"success"`
			}
			msg.Job = name

			if ok {
				g.runJob(name)
				msg.Success = true
				c.JSON(http.StatusOK, msg)
			} else {
				msg.Success = false
				c.JSON(http.StatusNotFound, msg)
			}
		})

		api.POST("/jobs/:name/toggle", func(c *gin.Context) {
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
		})
	}

	return g
}

func (g *Goflow) addUIRoutes() *Goflow {
	ui := g.router.Group("/ui")
	{
		ui.GET("/", func(c *gin.Context) {
			jobs := make([]*Job, 0)
			for _, job := range g.Jobs {
				jobs = append(jobs, job())
			}
			c.HTML(http.StatusOK, "index.html.tmpl", gin.H{"jobs": jobs})
		})

		ui.GET("/jobs/:name", func(c *gin.Context) {
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
	}

	g.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui/")
	})

	return g
}

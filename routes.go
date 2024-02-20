package goflow

import (
	"encoding/json"
	"net/http"
	"text/template"
	"time"

	"github.com/julienschmidt/httprouter"
)

func (g *Goflow) addTestRoute() {
	g.Router.GET("/api/jobs", g.handleJobs)
	g.Router.GET("/api/jobs/:name", g.handleParameterizedJobs)
	g.Router.POST("/api/jobs/:name/submit", g.handleSubmittedJobs)
	g.Router.POST("/api/jobs/:name/toggle", g.handleToggledJobs)
	g.Router.GET("/api/executions", g.handleExecutions)
	g.Router.GET("/", g.handleRedirect)
	g.Router.GET("/ui", g.handleRoot)
	g.Router.GET("/ui/jobs/:name", g.handleJobsPage)
	g.Router.ServeFiles("/static/*filepath", http.Dir(g.Options.UIPath))
}

//func (g *Goflow) addStreamRoute(keepOpen bool) {
//	g.router.GET("/stream", g.stream(keepOpen))
//}

func (g *Goflow) handleRedirect(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "/ui", http.StatusMovedPermanently)
}

func (g *Goflow) handleJobs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var msg struct {
		Jobs []string `json:"jobs"`
	}
	msg.Jobs = g.jobs
	out, _ := json.Marshal(msg)
	w.Write(out)
}

func (g *Goflow) handleParameterizedJobs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
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

		out, _ := json.Marshal(msg)
		w.Write(out)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (g *Goflow) handleSubmittedJobs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
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
		out, _ := json.Marshal(msg)
		w.Write(out)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (g *Goflow) handleToggledJobs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
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
		out, _ := json.Marshal(msg)
		w.Write(out)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (g *Goflow) handleExecutions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jobName := ps.ByName("jobname")
	stateQuery := ps.ByName("state")

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

	out, _ := json.Marshal(msg)
	w.Write(out)
}

func (g *Goflow) handleRoot(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	tmpl, _ := template.ParseFiles("ui/html/index.html.tmpl")
	tmpl.Execute(w, map[string]any{"jobs": jobs})
}

func (g *Goflow) handleJobsPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
	jobFn, ok := g.Jobs[name]

	if ok {
		tmpl, _ := template.ParseFiles("ui/html/job.html.tmpl")
		tmpl.Execute(w,
			map[string]any{
				"jobName":   name,
				"taskNames": jobFn().tasks,
				"schedule":  g.Jobs[name]().Schedule,
			})
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

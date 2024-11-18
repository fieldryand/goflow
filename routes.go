package goflow

import (
	"encoding/json"
	"net/http"
	"text/template"
	"time"
)

func (g *Goflow) addRoutes() {
	g.Router.HandleFunc("GET /api/jobs", g.handleJobs)
	g.Router.HandleFunc("GET /api/jobs/{name}", g.handleParameterizedJobs)
	g.Router.HandleFunc("POST /api/jobs/{name}/submit", g.handleSubmittedJobs)
	g.Router.HandleFunc("POST /api/jobs/{name}/toggle", g.handleToggledJobs)
	g.Router.HandleFunc("GET /api/executions", g.handleExecutions)
	g.Router.HandleFunc("GET /{$}", g.handleRedirect)
	g.Router.HandleFunc("GET /ui", g.handleRoot)
	g.Router.HandleFunc("GET /ui/jobs/{name}", g.handleJobsPage)
	g.Router.HandleFunc("GET /ui/diagrams/{name}", g.handleDiagramsPage)
	g.Router.HandleFunc("/events", g.handleStream)
	g.Router.HandleFunc("/events/{name}", g.handleStream)
	g.Router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(g.Options.UIPath))))
}

func (g *Goflow) handleRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/ui", http.StatusMovedPermanently)
}

func (g *Goflow) handleJobs(w http.ResponseWriter, r *http.Request) {
	var msg struct {
		Jobs []string `json:"jobs"`
	}
	msg.Jobs = g.jobs
	out, _ := json.Marshal(msg)
	_, err := w.Write(out)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (g *Goflow) handleParameterizedJobs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
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
		for job, _ := range g.cronEntries {
			if job == name {
				msg.Active = true
			}
		}

		out, _ := json.Marshal(msg)
		_, err := w.Write(out)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (g *Goflow) handleSubmittedJobs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	_, ok := g.Jobs[name]

	var msg struct {
		Job     string `json:"job"`
		Success bool   `json:"success"`
		StartTs string `json:"startTs"`
	}
	msg.Job = name

	if ok {
		go func() { g.queue <- name }()
		msg.Success = true
		msg.StartTs = time.Now().UTC().Format(time.RFC3339Nano)
		out, _ := json.Marshal(msg)
		_, err := w.Write(out)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (g *Goflow) handleToggledJobs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
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

func (g *Goflow) handleExecutions(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	jobName := r.PathValue("jobname")
	stateQuery := r.PathValue("state")

	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	executions := make([]*execution, 0)
	stored, _ := readExecutions(g.Store, d)

	for _, execution := range stored {
		if stateQuery != "" && stateQuery != string(execution.State) {
		} else if jobName != "" && jobName != execution.Job {
		} else {
			executions = append(executions, execution)
		}
	}

	var msg struct {
		Executions []*execution `json:"executions"`
	}
	msg.Executions = executions

	out, _ := json.Marshal(msg)
	w.Write(out)
}

func (g *Goflow) handleRoot(w http.ResponseWriter, r *http.Request) {
	jobs := make([]*Job, 0)
	for _, job := range g.jobs {

		// create the job, assume it's inactive
		j := g.Jobs[job]()
		j.Active = false

		// check if the job is active by looking in the list of cron entries
		for job, _ := range g.cronEntries {
			if job == j.Name {
				j.Active = true
			}
		}

		jobs = append(jobs, j)
	}

	tmpl, _ := template.ParseFiles("ui/html/index.html.tmpl")
	err := tmpl.Execute(w, map[string]any{"jobs": jobs})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (g *Goflow) handleJobsPage(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	jobFn, ok := g.Jobs[name]

	if ok {
		tmpl, _ := template.ParseFiles("ui/html/job.html.tmpl")
		err := tmpl.Execute(w,
			map[string]any{
				"jobName":   name,
				"taskNames": jobFn().tasks,
				"schedule":  g.Jobs[name]().Schedule,
			})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (g *Goflow) handleDiagramsPage(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	jobFn, ok := g.Jobs[name]

	if ok {
		tmpl, _ := template.ParseFiles("ui/html/diagram.html.tmpl")
		err := tmpl.Execute(w,
			map[string]any{
				"jobName":   name,
				"taskNames": jobFn().tasks,
				"schedule":  g.Jobs[name]().Schedule,
			})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

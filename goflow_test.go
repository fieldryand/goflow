package goflow

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fieldryand/goflow/op"
	"github.com/gin-gonic/gin"
)

var router = exampleRouter()
var w = httptest.NewRecorder()

func TestIndexRoute(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/jobs status is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestHealthRoute(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobsRoute(t *testing.T) {
	req, _ := http.NewRequest("GET", "/jobs", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobSubmitRoute(t *testing.T) {
	req, _ := http.NewRequest("GET", "/jobs/example/submit", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobStateRoute(t *testing.T) {
	req, _ := http.NewRequest("GET", "/jobs/example/state", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobDagRoute(t *testing.T) {
	req, _ := http.NewRequest("GET", "/jobs/example/dag", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func exampleJob() *Job {
	sleepOne := NewTask("sleepOne", op.Bash("sleep", "1"))
	j := NewJob("example").AddTask(sleepOne)
	return j
}

func exampleRouter() *gin.Engine {
	g := New(exampleJob)
	g.Use(DefaultLogger())
	g.addRoutes()
	return g.router
}

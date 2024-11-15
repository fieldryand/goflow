package goflow

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var router = exampleRouter()
var today = timeToDatestring(time.Now())

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func TestIndexRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/ui status is %d, expected %d", w.Code, http.StatusOK)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/ status is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobsRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/jobs", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	req, _ = http.NewRequest("GET", "/api/jobs/example-complex-analytics", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestExecutionsRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/executions?date=%s", today), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobSubmitToRouter(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs/example-complex-analytics/submit", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/jobs/example-random-failure/submit", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/jobs/bla/submit", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusNotFound)
	}
}

func TestJobToggleActiveRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs/example-complex-analytics/toggle", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	req, _ = http.NewRequest("POST", "/api/jobs/example-random-failure/toggle", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/jobs/bla/toggle", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusNotFound)
	}
}

func TestRouteNotFound(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/blaaaa", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusNotFound)
	}

	req, _ = http.NewRequest("GET", "/api/jobs/blaaaa", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusNotFound)
	}

	req, _ = http.NewRequest("GET", "/ui/jobs/blaaaa", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusNotFound)
	}
}

func TestJobOverviewRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ui/jobs/example-complex-analytics", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/jobs/bla", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusNotFound)
	}
}

func TestStreamRoute(t *testing.T) {
	var w = CreateTestResponseRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/events?stream=messages&date=%s&keepopen=false", today), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = CreateTestResponseRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/events/jobname=example-complex-analytics?stream=messages&date=%s&keepopen=false", today), nil)
	router.ServeHTTP(w, req)

	// wait for all goroutines to exit
	time.Sleep(2 * time.Second)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

// check for a race against /stream
func TestToggleRaceCondition(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs/example-complex-analytics/toggle", nil)
	router.ServeHTTP(w, req)
}

func exampleRouter() *http.ServeMux {
	ctx := context.Background()
	g := New(Options{UIPath: "ui/", ShowExamples: true, WithSeconds: true})
	g.addTestRoute()
	g.start(ctx)
	return g.Router
}

func TestInvalidJobName(t *testing.T) {
	g := New(Options{})

	err := g.AddJob(func() *Job { return &Job{Name: "", Schedule: "* * * * *"} })

	if err == nil {
		t.Errorf("Expected error adding a job with an invalid name")
	}
}

func TestExecutionOfNonexistentJob(t *testing.T) {
	g := New(Options{})
	ctx := context.Background()
	_, err := g.execute(ctx, "job")

	if err == nil {
		t.Errorf("Expected error executing a nonexistent job")
	}
}

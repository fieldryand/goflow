package goflow

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/philippgille/gokv/gomap"
)

var router = exampleRouter()

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
	req, _ := http.NewRequest("GET", "/ui/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/ui/ status is %d, expected %d", w.Code, http.StatusOK)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/ status is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestHealthRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
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
	req, _ := http.NewRequest("GET", "/api/executions", nil)
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
	req, _ = http.NewRequest("POST", "/api/jobs/example-custom-operator/submit", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/jobs/example-summation-job/submit", nil)

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

	req, _ = http.NewRequest("POST", "/api/jobs/example-custom-operator/toggle", nil)
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
	req, _ := http.NewRequest("GET", "/stream", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = CreateTestResponseRecorder()
	req, _ = http.NewRequest("GET", "/stream?jobname=example-complex-analytics", nil)
	router.ServeHTTP(w, req)

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

func exampleRouter() *gin.Engine {
	g := New(Options{UIPath: "ui/", ShowExamples: true, WithSeconds: true})
	g.execute("example-custom-operator")
	g.Use(DefaultLogger())
	g.addStaticRoutes()
	g.addStreamRoute(false)
	g.addUIRoutes()
	g.addAPIRoutes()
	return g.router
}

func TestScheduledExecution(t *testing.T) {
	store := gomap.NewStore(gomap.DefaultOptions)
	schedExec := scheduledExecution{store, customOperatorJob}
	schedExec.Run()
}

func TestGoflowWithoutOptions(t *testing.T) {
	g := New(Options{})
	g.Use(DefaultLogger())
}

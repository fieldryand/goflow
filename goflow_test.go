package goflow

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

	req, _ = http.NewRequest("GET", "/api/jobs/exampleComplexAnalytics", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobRunsRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/jobruns", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobSubmitToRouter(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs/exampleComplexAnalytics/submit", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/jobs/exampleCustomOperator/submit", nil)
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
	req, _ := http.NewRequest("POST", "/api/jobs/exampleComplexAnalytics/toggle", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	req, _ = http.NewRequest("POST", "/api/jobs/exampleCustomOperator/toggle", nil)
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
	req, _ := http.NewRequest("GET", "/ui/jobs/exampleComplexAnalytics", nil)
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
}

func exampleRouter() *gin.Engine {
	g := New(Options{UIPath: "ui/", ShowExamples: true})
	g.Use(DefaultLogger())
	g.addStaticRoutes()
	g.addStreamRoute()
	g.addUIRoutes()
	g.addAPIRoutes()
	return g.router
}

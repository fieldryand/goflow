package goflow

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

var router = exampleRouter()
var routerWithMemoryDB = exampleRouterWithMemoryDB()

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
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/jobs status is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestHealthRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobsRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobSubmitToRouterWithMemoryDB(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/jobs/exampleComplexAnalytics/submit", nil)

	routerWithMemoryDB.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func TestJobSubmitToRouter(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/jobs/exampleComplexAnalytics/submit", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/jobs/exampleCustomOperator/submit", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/jobs/bla/submit", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusNotFound)
	}
}

func TestJobToggleActiveRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/jobs/exampleComplexAnalytics/toggleActive", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	req, _ = http.NewRequest("POST", "/jobs/exampleActiveSchedule/toggleActive", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/jobs/bla/toggleActive", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusNotFound)
	}
}

func TestJobIsActiveRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/exampleComplexAnalytics/isActive", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/jobs/bla/isActive", nil)
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
}

func TestJobOverviewRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/exampleComplexAnalytics", nil)
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

func TestJobDagRoute(t *testing.T) {
	var w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/jobs/exampleComplexAnalytics/dag", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/jobs/bla/dag", nil)
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

func TestStreamRouteMemoryDB(t *testing.T) {
	var w = CreateTestResponseRecorder()
	req, _ := http.NewRequest("GET", "/stream", nil)
	routerWithMemoryDB.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}
}

func exampleRouter() *gin.Engine {
	g := New(Options{AssetBasePath: "assets/", ShowExamples: true})
	g.Use(DefaultLogger())
	g.addRoutes()
	return g.router
}

func exampleRouterWithMemoryDB() *gin.Engine {
	g := New(Options{AssetBasePath: "assets/", ShowExamples: true, DBType: "memory"})
	g.Use(DefaultLogger())
	g.addRoutes()
	return g.router
}

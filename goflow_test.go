package goflow

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/philippgille/gokv/gomap"
)

var router = exampleRouter()
var routerWithSeconds = exampleRouterWithSeconds()

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

	req, _ = http.NewRequest("GET", "/api/health", nil)
	routerWithSeconds.ServeHTTP(w, req)

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
	req, _ := http.NewRequest("POST", "/api/jobs/example-complex-analytics/execute", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("httpStatus is %d, expected %d", w.Code, http.StatusOK)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/jobs/bla/execute", nil)
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
}

func simpleJob() *Job {
	j := &Job{Name: "simple", Schedule: "* * * * *", Timeout: 1 * time.Second}
	j.Add(&Task{Name: "add-two-two", Operator: Addition{2, 2}})
	return j
}

func exampleRouter() *gin.Engine {
	store := gomap.NewStore(gomap.DefaultOptions)
	g := New(Options{ShowExamples: true})
	g.AttachStore(store)
	g.Add(simpleJob)
	g.Execute("simple")
	g.scheduledExecute("simple")
	g.Use(DefaultLogger())
	g.addStaticRoutes()
	g.addStreamRoute()
	g.addUIRoutes()
	g.addAPIRoutes()
	return g.router
}

func exampleRouterWithSeconds() *gin.Engine {
	g := New(Options{ShowExamples: true, WithSeconds: true})
	return g.router
}

func cyclicJob() *Job {
	j := &Job{Name: "cyclic", Schedule: "* * * * *"}
	j.Add(&Task{Name: "addTwoTwo", Operator: Addition{2, 2}})
	j.Add(&Task{Name: "addFourFour", Operator: Addition{4, 4}})
	j.SetDownstream("addTwoTwo", "addFourFour")
	j.SetDownstream("addFourFour", "addTwoTwo")
	return j
}

func TestCyclicJob(t *testing.T) {
	g := New(Options{})
	if g.Add(cyclicJob) == nil {
		t.Errorf("cyclic job should be rejected")
	}
}

func invalidJob() *Job {
	return &Job{Name: "", Schedule: "* * * * *"}
}

func TestInvalidJobName(t *testing.T) {
	g := New(Options{})
	err := g.Add(invalidJob)
	if err == nil {
		t.Errorf("job with invalid name should be rejected")
	}
}

// An example program demonstrating the usage of goflow.
package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/fieldryand/goflow"
	"github.com/fieldryand/goflow/operator"

	"github.com/gin-gonic/gin"
)

func main() {
	jobs := map[string](func() *goflow.Job){
		"exampleOne":   ExampleJobOne,
		"exampleTwo":   ExampleJobTwo,
		"exampleThree": ExampleJobThree,
	}

	middleware := [](func() gin.HandlerFunc){structuredLogger}

	go heartbeat()

	goflow := goflow.Goflow(jobs, middleware)

	goflow.Run(":8090")
}

// ExampleJobOne returns a simple job consisting of Addition and Sleep operators.
func ExampleJobOne() *goflow.Job {
	sleepOne := goflow.NewTask("sleepOne", operator.NewSleep(1))
	addOneOne := goflow.NewTask("addOneOne", NewAddition(1, 1))
	sleepTwo := goflow.NewTask("sleepTwo", operator.NewSleep(2))
	addTwoFour := goflow.NewTask("addTwoFour", NewAddition(2, 4))
	addThreeFour := goflow.NewTask("addThreeFour", NewAddition(3, 4))

	j := goflow.NewJob("exampleOne").
		AddTask(sleepOne).
		AddTask(addOneOne).
		AddTask(sleepTwo).
		AddTask(addTwoFour).
		AddTask(addThreeFour).
		SetDownstream(sleepOne, addOneOne).
		SetDownstream(addOneOne, sleepTwo).
		SetDownstream(sleepTwo, addTwoFour).
		SetDownstream(addOneOne, addThreeFour)

	return j
}

// ExampleJobTwo returns an even simpler job consisting of a single Sleep task.
func ExampleJobTwo() *goflow.Job {
	sleepTen := goflow.NewTask("sleepTen", operator.NewSleep(10))
	j := goflow.NewJob("exampleTwo").AddTask(sleepTen)
	return j
}

// ExampleJobThree returns a job with a task that throws an error.
func ExampleJobThree() *goflow.Job {
	badTask := goflow.NewTask("badTask", NewAddition(-10, 0))
	j := goflow.NewJob("exampleThree").AddTask(badTask)
	return j
}

// We can create custom operators by implementing the Run() method.

// Addition is an operation that adds two nonnegative numbers.
type Addition struct{ a, b int }

func NewAddition(a, b int) *Addition {
	o := Addition{a, b}
	return &o
}

func (o Addition) Run() (interface{}, error) {

	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}

	result := o.a + o.b
	return result, nil
}

func structuredLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		entry := &structuredLogEntry{
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		}

		encoded, _ := json.Marshal(entry)
		return string(encoded) + "\n"
	})
}

type structuredLogEntry struct {
	ClientIp     string        `json:"clientIp"`
	Timestamp    string        `json:"timestamp"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	Proto        string        `json:"protocol"`
	Status       int           `json:"status"`
	Latency      time.Duration `json:"latency"`
	UserAgent    string        `json:"userAgent"`
	ErrorMessage string        `json:"errorMessage"`
}

func heartbeat() {
	for {
		time.Sleep(5 * time.Second)
		http.Get("http://localhost:8090/health")
	}
}

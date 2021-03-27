// An example program demonstrating custom middleware.
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fieldryand/goflow"
	"github.com/fieldryand/goflow/operator"

	"github.com/gin-gonic/gin"
)

func main() {
	go heartbeat()

	goflow := goflow.New(ExampleJobTwo)
	goflow.Use(gin.Recovery())
	goflow.Use(Logger())
	goflow.Run(":8100")
}

// ExampleJobTwo returns an even simpler job consisting of a single Sleep task.
func ExampleJobTwo() *goflow.Job {
	sleepTen := goflow.NewTask("sleepTen", operator.NewSleep(10))
	j := goflow.NewJob("exampleTwo").AddTask(sleepTen)
	return j
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[GOFLOW] [%s] %s - \"%s %s %s %d %s \"%s\" %s\"\n",
			param.TimeStamp.Format(time.RFC1123),
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func heartbeat() {
	for {
		time.Sleep(5 * time.Second)
		http.Get("http://localhost:8100/health")
	}
}

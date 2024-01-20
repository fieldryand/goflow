package goflow

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// Setting clientDisconnect to false makes testing possible.
func (g *Goflow) stream(clientDisconnect bool) func(*gin.Context) {

	return func(c *gin.Context) {
		chanStream := make(chan string)
		go func() {
			defer close(chanStream)
			// Periodically push the list of job runs into the stream
			for {
				for jobname := range g.Jobs {
					executions, _ := readExecutions(g.Store, jobname)
					marshalled, _ := marshalExecutions(jobname, executions)
					chanStream <- string(marshalled)
				}
				time.Sleep(time.Second * 1)
			}
		}()
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-chanStream; ok {
				c.SSEvent("message", msg)
				return clientDisconnect
			}
			return false
		})
	}

}

// Obtain locks and put the response in the structure expected
// by the streaming endpoint.
func marshalExecutions(name string, executions []*Execution) ([]byte, error) {
	var msg struct {
		JobName    string       `json:"jobName"`
		Executions []*Execution `json:"executions"`
	}
	msg.JobName = name
	msg.Executions = executions
	result, ok := json.Marshal(msg)
	return result, ok
}

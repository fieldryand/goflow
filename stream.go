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
					jobruns, _ := readJobRuns(g.Store, jobname)
					marshalled, _ := marshalJobRuns(jobname, jobruns)
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
func marshalJobRuns(name string, jobruns []*jobRun) ([]byte, error) {
	var msg struct {
		JobName string    `json:"jobName"`
		JobRuns []*jobRun `json:"jobRuns"`
	}
	msg.JobName = name
	msg.JobRuns = jobruns
	result, ok := json.Marshal(msg)
	return result, ok
}

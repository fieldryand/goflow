package goflow

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// Set keepOpen to false when testing--one event will be sent and
// then the channel is closed by the server.
func (g *Goflow) stream(keepOpen bool) func(*gin.Context) {

	return func(c *gin.Context) {
		job := c.Query("jobname")

		history := make([]*Execution, 0)

		// open a channel for live executions
		chanStream := make(chan *Execution)

		// periodically push the list of job runs into the stream
		go func() {
			defer close(chanStream)
			for {
				for jobname := range g.Jobs {
					executions, _ := readExecutions(g.Store, jobname)
					for _, e := range executions {

						// make sure it wasn't already sent
						inHistory := false

						for _, h := range history {
							if e.ID == h.ID && e.ModifiedTimestamp == h.ModifiedTimestamp {
								inHistory = true
							}
						}

						if !inHistory {
							if job != "" && job == e.JobName {
								chanStream <- e
								history = append(history, e)
							} else if job == "" {
								chanStream <- e
								history = append(history, e)
							}
						}

					}
				}
				time.Sleep(time.Second * 1)
			}
		}()

		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-chanStream; ok {
				c.SSEvent("message", msg)
				return keepOpen
			}
			return false
		})
	}

}

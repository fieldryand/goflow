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

		history := make([]*execution, 0)

		// periodically push the list of job runs into the stream
		c.Stream(func(w io.Writer) bool {
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
						if (job != "" && job == e.JobName) || job == "" {
							c.SSEvent("message", e)
							history = append(history, e)
						}
					}

				}
			}

			time.Sleep(time.Second * 1)

			return keepOpen
		})
	}

}

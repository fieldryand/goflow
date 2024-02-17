package goflow

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// Setting clientDisconnect to false makes testing possible.
func (g *Goflow) stream(clientDisconnect bool) func(*gin.Context) {

	return func(c *gin.Context) {

		history := make([]*execution, 0)

		// open a channel for live executions
		chanStream := make(chan *execution)

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
							chanStream <- e
							history = append(history, e)
						}

					}
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

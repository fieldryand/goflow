package goflow

import (
	"encoding/json"
	"net/http"
	"time"

	sse "github.com/r3labs/sse/v2"
)

// Set keepOpen to false when testing--one event will be sent and
// then the channel is closed by the server.
func (g *Goflow) handleStream(w http.ResponseWriter, r *http.Request) {
	job := r.PathValue("name")

	server := sse.New()
	server.CreateStream("messages")

	history := make([]*Execution, 0)

	// periodically push the list of job runs into the stream
	go func() {
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
							out, _ := json.Marshal(e)
							server.Publish("messages", &sse.Event{
								Data: []byte(out),
							})
							history = append(history, e)
						} else if job == "" {
							out, _ := json.Marshal(e)
							server.Publish("messages", &sse.Event{
								Data: []byte(out),
							})
							history = append(history, e)
						}
					}

				}
			}
			time.Sleep(time.Second * 1)
		}
	}()

	server.ServeHTTP(w, r)

}

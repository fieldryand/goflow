package goflow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Set keepOpen to false when testing--one event will be sent and
// then the channel is closed by the server.
func (g *Goflow) handleStream(w http.ResponseWriter, r *http.Request) {
	job := r.PathValue("name")
	keepOpen := r.URL.Query().Get("keepopen")

	flusher, err := w.(http.Flusher)
	if !err {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	history := make([]*Execution, 0)

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			for jobname := range g.Jobs {
				executions, _ := readExecutions(g.Store, jobname)
				for _, e := range executions {

					// make sure it wasn't already sent
					inHistory := false

					for _, h := range history {
						if e.ID == h.ID && e.ModifiedTs == h.ModifiedTs {
							inHistory = true
						}
					}

					if !inHistory {
						if (job != "" && job == e.JobName) || job == "" {
							out, _ := json.Marshal(e)
							w.Write([]byte(fmt.Sprintf("data: %s\n", out)))
							w.Write([]byte("\n"))
							flusher.Flush()
							history = append(history, e)
						}
					}

				}
			}

			if keepOpen == "false" {
				return
			}

			time.Sleep(time.Second * 1)
		}
	}

}

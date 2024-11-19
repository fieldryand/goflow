package goflow

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (g *Goflow) handleStream(w http.ResponseWriter, r *http.Request) {
	job := r.PathValue("name")
	date := r.URL.Query().Get("date")
	keepOpen := r.URL.Query().Get("keepopen")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	history := make([]*execution, 0)

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			executions, _ := readExecutions(g.Store, d)
			for _, e := range executions {

				// make sure it wasn't already sent
				inHistory := false

				for _, h := range history {
					if e.ID == h.ID && e.ModifiedTs == h.ModifiedTs {
						inHistory = true
					}
				}

				if !inHistory {
					if (job != "" && job == e.Job) || job == "" {
						out, _ := json.Marshal(e)
						_, err := w.Write([]byte(fmt.Sprintf("data: %s\n", out)))
						if err != nil {
							log.Printf("write failed: %v", err)
						}
						_, err = w.Write([]byte("\n"))
						if err != nil {
							log.Printf("write failed: %v", err)
						}
						flusher.Flush()
						history = append(history, e)
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

// This example demonstrates a chain of results passed from one task
// to another. Each job execution performs a sum of three integers:
// 1 + 2 + 3 = 6, with each task adding a term to the running sum.
// The task operator gets its jobID from the context, then updates a
// synchronized map, i.e. map[jobID] = sum.
package main

import (
	"context"
	"log"
	"sync"

	"github.com/fieldryand/goflow/v3"
)

// A synchronized map prevents concurrent writes.
type sum struct {
	values map[any]int
	sync.RWMutex
}

var s = &sum{values: make(map[any]int)}

type summation struct{ term int }

func (o summation) Run(ctx context.Context) (any, error) {
	// The jobID is available in the context.
	key := goflow.ContextKey("jobID")
	jobID := ctx.Value(key)
	s.Lock()
	s.values[jobID] = s.values[jobID] + o.term
	log.Printf("the sum is %v", s.values[jobID])
	s.Unlock()
	return nil, nil
}

func chainedJob() *goflow.Job {
	j := &goflow.Job{Name: "example-chained", Schedule: "* * * * * *", Active: true}
	err := j.AddTask(
		&goflow.Task{Name: "a", Operator: summation{1}},
		&goflow.Task{Name: "b", Operator: summation{2}},
		&goflow.Task{Name: "c", Operator: summation{3}},
	)
	if err != nil {
		log.Printf("error adding task: %v", err)
	}
	return j
}

func main() {
	gf := goflow.New(goflow.Options{ShowExamples: false, WithSeconds: true})
	err := gf.AddJob(chainedJob)
	if err != nil {
		log.Printf("error adding task: %v", err)
	}
	gf.Run(context.Background())
}

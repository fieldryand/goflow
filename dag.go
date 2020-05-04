package goflow

import (
	"github.com/ef-ds/deque"
)

type dag struct {
	Graph map[string][]string
}

func newDag() *dag {
	d := dag{make(map[string][]string)}
	return &d
}

// A node has a name and 0 or more dependent nodes
func (d *dag) addNode(name string) {
	deps := make([]string, 0)
	d.Graph[name] = deps
}

// Create an edge between an independent and dependent node
func (d *dag) setDownstream(ind, dep string) {
	d.Graph[ind] = append(d.Graph[ind], dep)
}

type invalidDagError struct {
}

func (e *invalidDagError) Error() string {
	return "Invalid DAG"
}

// Ensure the DAG is acyclic
func (d *dag) validate() bool {
	degree := make(map[string]int)

	for node := range d.Graph {
		degree[node] = 0
	}

	for _, ds := range d.Graph {
		for _, i := range ds {
			degree[i]++
		}
	}

	var deq deque.Deque

	for node, val := range degree {
		if val == 0 {
			deq.PushFront(node)
		}
	}

	l := make([]string, 0)

	for {
		popped, ok := deq.PopBack()

		if ok == false {
			break
		} else {
			node := popped.(string)
			l = append(l, node)
			for _, ds := range d.Graph[node] {
				degree[ds]--
				if degree[ds] == 0 {
					deq.PushFront(ds)
				}
			}
		}
	}

	if len(l) == len(d.Graph) {
		return true
	}

	return false
}

// Return the immediately upstream nodes for a given node
func (d *dag) dependencies(node string) []string {

	dependencies := make([]string, 0)

	for dep, ds := range d.Graph {
		for _, i := range ds {
			if node == i {
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies
}

// Return all the independent nodes in the graph
func (d *dag) independentNodes() []string {

	downstream := make([]string, 0)

	for _, ds := range d.Graph {
		for _, i := range ds {
			downstream = append(downstream, i)
		}
	}

	ind := make([]string, 0)

	for node := range d.Graph {
		ctr := 0
		for _, i := range downstream {
			if node == i {
				ctr++
			}
		}
		if ctr == 0 {
			ind = append(ind, node)
		}
	}

	return ind

}

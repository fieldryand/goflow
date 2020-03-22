package core

import (
	"github.com/ef-ds/deque"
)

type Dag struct {
	graph map[string][]string
}

// A node has a name and 0 or more dependent nodes
func (d *Dag) addNode(name string) {
	deps := make([]string, 0)
	d.graph[name] = deps
}

// Create an edge between an independent and dependent node
func (d *Dag) setDownstream(ind, dep string) {
	d.graph[ind] = append(d.graph[ind], dep)
}

type InvalidDagError struct {
}

func (e *InvalidDagError) Error() string {
	return "Invalid DAG"
}

// Ensure the DAG is acyclic
func (d *Dag) validate() bool {
	degree := make(map[string]int)

	for node, _ := range d.graph {
		degree[node] = 0
	}

	for _, ds := range d.graph {
		for _, i := range ds {
			degree[i] += 1
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
			for _, ds := range d.graph[node] {
				degree[ds] -= 1
				if degree[ds] == 0 {
					deq.PushFront(ds)
				}
			}
		}
	}

	if len(l) == len(d.graph) {
		return true
	} else {
		return false
	}
}

// Return the immediately upstream nodes for a given node
func (d *Dag) dependencies(node string) []string {

	dependencies := make([]string, 0)

	for dep, ds := range d.graph {
		for _, i := range ds {
			if node == i {
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies
}

// Return all the independent nodes in the graph
func (d *Dag) independentNodes() []string {

	downstream := make([]string, 0)

	for _, ds := range d.graph {
		for _, i := range ds {
			downstream = append(downstream, i)
		}
	}

	ind := make([]string, 0)

	for node, _ := range d.graph {
		ctr := 0
		for _, i := range downstream {
			if node == i {
				ctr += 1
			}
		}
		if ctr == 0 {
			ind = append(ind, node)
		}
	}

	return ind

}

package goflow

import (
	"github.com/ef-ds/deque"
)

// Credit: The DAG implementation here is roughly a port of the
// one from this Python project: https://github.com/thieman/dagobah

// A DAG is a directed acyclic graph represented by a simple map
// where a key is a node in the graph, and a value is a slice of
// immediately downstream dependent nodes.
type dag map[string][]string

// A node has a name and 0 or more dependent nodes
func (d dag) addNode(name string) {
	deps := make([]string, 0)
	d[name] = deps
}

// Create an edge between an independent and dependent node
func (d dag) setDownstream(ind, dep string) {
	d[ind] = append(d[ind], dep)
}

// Returns true if a node is a downstream node, false if it
// is independent.
func (d dag) isDownstream(nodeName string) bool {
	ind := d.independentNodes()

	for _, name := range ind {
		if nodeName == name {
			return false
		}
	}

	return true
}

// Ensure the DAG is acyclic
func (d dag) validate() bool {
	degree := make(map[string]int)

	for node := range d {
		degree[node] = 0
	}

	for _, ds := range d {
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

		if !ok {
			break
		} else {
			node := popped.(string)
			l = append(l, node)
			for ds := range d {
				degree[ds]--
				if degree[ds] == 0 {
					deq.PushFront(ds)
				}
			}
		}
	}

	return len(l) == len(d)
}

// Return the immediately upstream nodes for a given node
func (d dag) dependencies(node string) []string {

	dependencies := make([]string, 0)

	for dep, ds := range d {
		for _, i := range ds {
			if node == i {
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies
}

// Return all the independent nodes in the graph
func (d dag) independentNodes() []string {

	downstream := make([]string, 0)

	for _, ds := range d {
		downstream = append(downstream, ds...)
	}

	ind := make([]string, 0)

	for node := range d {
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

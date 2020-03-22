package core

import (
	"testing"
)

func TestDag(t *testing.T) {
	d := NewDag()

	d.addNode("a")
	d.addNode("b")
	d.addNode("c")
	d.addNode("d")
	d.setDownstream("a", "b")
	d.setDownstream("a", "c")
	d.setDownstream("b", "d")
	d.setDownstream("c", "d")

	if !d.validate() {
		t.Errorf("Valid dag failed validation check")
	}

	if !Equal(d.dependencies("b"), []string{"a"}) {
		t.Errorf("d.dependencies() returned %s, expected %s",
			d.dependencies("b"),
			[]string{"a"})
	}

	if !Equal(d.independentNodes(), []string{"a"}) {
		t.Errorf("d.independentNodes() returned %s, expected %s",
			d.dependencies("b"),
			[]string{"a"})
	}

	e := NewDag()

	e.addNode("a")
	e.addNode("b")
	e.setDownstream("a", "b")
	e.setDownstream("b", "a")

	if e.validate() {
		t.Errorf("Invalid dag passed validation check")
	}

	invalidErr := InvalidDagError{}

	if invalidErr.Error() != "Invalid DAG" {
		t.Errorf("InvalidDagError returned unexpected message")
	}
}

package goflow

import (
	"testing"
)

type test struct {
	a   []string
	b   []string
	out bool
}

var tests = []test{
	{[]string{"a"}, []string{"a"}, true},
	{[]string{"a"}, []string{"a", "b"}, false},
	{[]string{"a"}, []string{"b"}, false},
}

func TestEqual(t *testing.T) {
	for _, v := range tests {
		got := equal(v.a, v.b)
		if got != v.out {
			t.Errorf("Test failed: got %t, expected %t", got, v.out)
		}
	}
}

package core

import (
	"testing"
)

func TestAddOperator(t *testing.T) {
	ao := AddOperator(3, 2)
	result, _ := ao.run()

	if result != 5 {
		t.Errorf("AddOperator.run() expected %d, got %d", result, 5)
	}
}

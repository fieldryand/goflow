package operators

import (
	"testing"
)

func TestAddOperator(t *testing.T) {
	ao := AddOperator(3, 2)
	result, _ := ao.Run()

	if result != 5 {
		t.Errorf("Expected %d, got %d", result, 5)
	}
}

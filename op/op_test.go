package op

import (
	"testing"
)

func TestBash(t *testing.T) {
	o := Bash("sh", "-c", "echo $((2 + 4))")
	result, _ := o.Run()

	if result != 6 {
		t.Errorf("Expected %t, got %d", true, result)
	}
}

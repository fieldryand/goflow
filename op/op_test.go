package op

import (
	"fmt"
	"testing"
)

func TestBash(t *testing.T) {
	result, _ := Bash("sh", "-c", "echo $((2 + 4))").Run()
	result_str := fmt.Sprintf("%v", result)
	expected := "6\n"

	if result_str != expected {
		t.Errorf("Expected %s, got %s", expected, result_str)
	}
}

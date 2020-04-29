package operator

import (
	"testing"
)

func TestSleep(t *testing.T) {
	o := NewSleep(1)
	result, _ := o.Run()

	if result != true {
		t.Errorf("Expected %t, got %d", true, result)
	}
}

package p

import "testing"

func TestCrash(t *testing.T) {
	err := Do()
	if err != nil {
		t.Fatal(err)
	}
}

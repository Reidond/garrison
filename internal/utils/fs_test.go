package utils

import "testing"

func TestExpandPath(t *testing.T) {
	p, err := ExpandPath("~/")
	if err != nil || p == "" {
		t.Fatalf("expected expanded path, got %q err=%v", p, err)
	}
}

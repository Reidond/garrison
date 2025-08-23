package steamcmd

import (
	"context"
	"testing"
)

// We can't run steamcmd here, but we can at least ensure loginArgs shape.
func TestLoginArgs(t *testing.T) {
	if got := New(Config{}).loginArgs(); len(got) != 2 || got[0] != "+login" || got[1] != "anonymous" {
		t.Fatalf("expected anonymous login, got %#v", got)
	}
	if got := New(Config{Username: "u"}).loginArgs(); len(got) != 2 || got[1] != "u" {
		t.Fatalf("expected username only, got %#v", got)
	}
	if got := New(Config{Username: "u", Password: "p"}).loginArgs(); len(got) != 3 || got[2] != "p" {
		t.Fatalf("expected user+pass, got %#v", got)
	}
	_ = context.TODO()
}

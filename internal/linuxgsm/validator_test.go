package linuxgsm

import "testing"

func TestConfigValidate(t *testing.T) {
	ok := Config{
		ContainerName: "test-armar",
		GameID:        "armar",
		Volumes:       []string{"/tmp/lgsm:/data"},
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bad := Config{}
	if err := bad.Validate(); err == nil {
		t.Fatalf("expected error for empty config")
	}

	bad2 := Config{ContainerName: "c", GameID: "invalid", Volumes: []string{"/tmp/x:/data"}}
	if err := bad2.Validate(); err == nil {
		t.Fatalf("expected unsupported game id error")
	}
}

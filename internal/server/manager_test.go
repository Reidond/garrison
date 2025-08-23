package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPIDFileLifecycle(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	m := NewManager(Metadata{Name: "t", InstallDir: tmp})
	if err := m.writePID(12345); err != nil {
		t.Fatalf("writePID: %v", err)
	}
	pid, err := m.readPID()
	if err != nil || pid != 12345 {
		t.Fatalf("readPID: %v pid=%d", err, pid)
	}
	m.removePID()
	if _, err := os.Stat(filepath.Join(tmp, ".garrison", "t.pid")); !os.IsNotExist(err) {
		t.Fatalf("pid file should be gone, err=%v", err)
	}
}

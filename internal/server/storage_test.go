package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorageCRUD(t *testing.T) {
	// Redirect HOME to a temp dir to avoid touching real user data
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)

	md := Metadata{Name: "test", InstallDir: tmp, AppID: 1}
	if err := SaveMetadata(md); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := LoadMetadata("test")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Name != md.Name || got.InstallDir != md.InstallDir || got.AppID != md.AppID {
		t.Fatalf("roundtrip mismatch: %+v vs %+v", got, md)
	}
	list, err := ListMetadata()
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v len=%d", err, len(list))
	}
	if err := DeleteMetadata("test"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".garrison", "test.json")); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, err=%v", err)
	}
}

package tests

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

// Minimal integration: build the binary and run --help.
func TestCLIHelp(t *testing.T) {
    wd, _ := os.Getwd()
    repo := filepath.Dir(wd)
    bin := filepath.Join(repo, "garrison")

    cmd := exec.Command("go", "build", "-o", bin, "./cmd/garrison")
    cmd.Dir = repo
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("build failed: %v\n%s", err, string(out))
    }

    help := exec.Command(bin, "--help")
    help.Dir = repo
    if out, err := help.CombinedOutput(); err != nil {
        t.Fatalf("running --help failed: %v\n%s", err, string(out))
    }
}

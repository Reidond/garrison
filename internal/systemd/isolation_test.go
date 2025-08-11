package systemd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	execu "github.com/example/garrison/internal/executil"
)

func withXDGEnv(t *testing.T) (dataHome, cacheHome, stateHome, configHome string) {
	t.Helper()
	base := t.TempDir()
	dataHome = filepath.Join(base, "data")
	cacheHome = filepath.Join(base, "cache")
	stateHome = filepath.Join(base, "state")
	configHome = filepath.Join(base, "config")
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("XDG_CACHE_HOME", cacheHome)
	t.Setenv("XDG_STATE_HOME", stateHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", base)
	return
}

func TestComputePathsWithXDG(t *testing.T) {
	dataHome, cacheHome, stateHome, configHome := withXDGEnv(t)
	p, err := computePaths("my-game")
	if err != nil {
		t.Fatalf("computePaths error: %v", err)
	}
	wantUnit := filepath.Join(configHome, "systemd", "user", "garrison-my-game.service")
	if p.Unit != wantUnit {
		t.Fatalf("Unit = %q, want %q", p.Unit, wantUnit)
	}
	if p.UnitDir != filepath.Dir(wantUnit) {
		t.Fatalf("UnitDir = %q, want %q", p.UnitDir, filepath.Dir(wantUnit))
	}
	if p.Base != filepath.Join(dataHome, "garrison", "my-game") {
		t.Fatalf("Base = %q", p.Base)
	}
	if p.App != filepath.Join(p.Base, "app") {
		t.Fatalf("App = %q", p.App)
	}
	if p.Profiles != filepath.Join(p.Base, "profiles") {
		t.Fatalf("Profiles = %q", p.Profiles)
	}
	if p.Config != filepath.Join(p.Base, "profiles", "config.json") {
		t.Fatalf("Config = %q", p.Config)
	}
	if p.Logs != filepath.Join(stateHome, "garrison", "my-game") {
		t.Fatalf("Logs = %q", p.Logs)
	}
	if p.Cache != filepath.Join(cacheHome, "garrison", "steamcmd") {
		t.Fatalf("Cache = %q", p.Cache)
	}
}

func TestBuildUnitIncludesSandboxingAndExec(t *testing.T) {
	withXDGEnv(t)
	p, err := computePaths("game")
	if err != nil {
		t.Fatalf("computePaths: %v", err)
	}
	unit := buildUnit([]string{"echo pre-one", "echo pre-two"}, "/bin/true -f", p)
	checks := []string{
		"NoNewPrivileges=yes",
		"ProtectSystem=strict",
		"ReadWritePaths=" + p.Base,
		"ReadWritePaths=" + p.Cache,
		"ReadWritePaths=" + p.Logs,
		"ExecStartPre=/usr/bin/bash -lc 'echo pre-one'",
		"ExecStartPre=/usr/bin/bash -lc 'echo pre-two'",
		"ExecStart=/bin/true -f",
	}
	for _, c := range checks {
		if !strings.Contains(unit, c) {
			t.Fatalf("unit content missing %q\n%s", c, unit)
		}
	}
}

type fakeRunner struct {
	calls   []string
	nextOut string
	nextErr error
}

func (f *fakeRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	return nil
}

func (f *fakeRunner) CombinedOutput(name string, args ...string) (string, error) {
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	return f.nextOut, f.nextErr
}

func TestRunSteamcmdTransient_InvokesSystemdRun(t *testing.T) {
	_, _, _, _ = withXDGEnv(t)
	fr := &fakeRunner{}
	execu.SetDefault(fr)
	t.Cleanup(func() { execu.SetDefault(execu.HostRunner{}) })

	if err := RunSteamcmdTransient("srv", "123", true); err != nil {
		t.Fatalf("RunSteamcmdTransient: %v", err)
	}
	if len(fr.calls) == 0 || !strings.HasPrefix(fr.calls[0], "systemd-run ") {
		t.Fatalf("expected systemd-run call, got %v", fr.calls)
	}
	found := false
	for _, c := range fr.calls {
		if strings.Contains(c, "--unit=garrison-scmd-srv") && strings.Contains(c, "+app_update 123 validate") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing expected args in calls: %v", fr.calls)
	}
}

func TestInstallAndStartUnit_InvokesSystemctlAndWritesUnit(t *testing.T) {
	_, _, _, _ = withXDGEnv(t)
	fr := &fakeRunner{}
	execu.SetDefault(fr)
	t.Cleanup(func() { execu.SetDefault(execu.HostRunner{}) })

	if err := InstallAndStartUnit("srvx", []string{"echo pre"}, "/bin/true"); err != nil {
		t.Fatalf("InstallAndStartUnit: %v", err)
	}
	p, _ := computePaths("srvx")
	b, err := os.ReadFile(p.Unit)
	if err != nil {
		t.Fatalf("unit missing: %v", err)
	}
	unit := string(b)
	if !strings.Contains(unit, "ExecStart=/bin/true") || !strings.Contains(unit, "ExecStartPre=/usr/bin/bash -lc 'echo pre'") {
		t.Fatalf("unexpected unit content: %s", unit)
	}
	want := []string{
		"systemctl --user daemon-reload",
		"systemctl --user enable " + serviceName("srvx"),
		"systemctl --user start " + serviceName("srvx"),
	}
	for _, w := range want {
		ok := false
		for _, c := range fr.calls {
			if c == w {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("missing call %q in %v", w, fr.calls)
		}
	}
}

func TestStopAndStatus_UseSystemctl(t *testing.T) {
	fr := &fakeRunner{nextOut: "active\n"}
	execu.SetDefault(fr)
	t.Cleanup(func() { execu.SetDefault(execu.HostRunner{}) })
	if err := Stop("srv"); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if err := Status("srv"); err != nil {
		t.Fatalf("Status: %v", err)
	}
	// Expect at least stop and is-active
	has := func(s string) bool {
		for _, c := range fr.calls {
			if strings.HasPrefix(c, s) {
				return true
			}
		}
		return false
	}
	if !has("systemctl --user stop garrison-srv.service") {
		t.Fatalf("stop not called: %v", fr.calls)
	}
	if !has("systemctl --user is-active garrison-srv.service") {
		t.Fatalf("is-active not called: %v", fr.calls)
	}
}

func TestStatus_UnknownStateFallsBackToStatus(t *testing.T) {
	fr := &fakeRunner{nextOut: "weird\n", nextErr: fmt.Errorf("boom")}
	execu.SetDefault(fr)
	t.Cleanup(func() { execu.SetDefault(execu.HostRunner{}) })
	_ = Status("srv2")
	want := "systemctl --user status --no-pager garrison-srv2.service"
	ok := false
	for _, c := range fr.calls {
		if c == want {
			ok = true
			break
		}
	}
	if !ok {
		t.Fatalf("expected fallback status call; got %v", fr.calls)
	}
}

func TestDeleteAll_RemovesArtifactsAndReloads(t *testing.T) {
	_, _, _, _ = withXDGEnv(t)
	p, err := computePaths("srvd")
	if err != nil {
		t.Fatal(err)
	}
	// Create expected files/dirs
	if err := os.MkdirAll(p.UnitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.Unit, []byte("[Unit]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.Base, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.Cache, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.Logs, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := DeleteAll("srvd"); err != nil {
		t.Fatalf("DeleteAll error: %v", err)
	}

	// Verify removals
	if _, err := os.Stat(p.Unit); !os.IsNotExist(err) {
		t.Fatalf("unit not removed: err=%v", err)
	}
	if _, err := os.Stat(p.Base); !os.IsNotExist(err) {
		t.Fatalf("base not removed: err=%v", err)
	}
	if _, err := os.Stat(p.Cache); !os.IsNotExist(err) {
		t.Fatalf("cache not removed: err=%v", err)
	}
	if _, err := os.Stat(p.Logs); !os.IsNotExist(err) {
		t.Fatalf("logs not removed: err=%v", err)
	}
}

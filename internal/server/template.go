package server

import (
	"os"
	"path/filepath"
	"strings"
)

// Interpolate replaces placeholders in a template with values.
// Supported: {{install_dir}}, {{app_id}}, {{name}}, {{args}}
func Interpolate(tmpl string, vars map[string]string) string {
	out := tmpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}

// ResolvePath makes path absolute relative to installDir if not absolute.
func ResolvePath(installDir, p string) string {
	if p == "" {
		return p
	}
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(installDir, p)
}

// InheritEnv merges existing process env with map, returning list of "K=V" strings.
func InheritEnv(extra map[string]string) []string {
	env := os.Environ()
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	return env
}

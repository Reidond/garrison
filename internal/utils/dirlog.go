package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// DirLogger emits a pretty directory tree listing to a printf-like sink.
// The sink is typically testing.T.Logf or slog.Logger.Infof competitor.
// It prints entries with a prefix and aligns file sizes.
type DirLogger struct {
	Printf func(format string, args ...any)
	// Prefix (e.g., "│ ", "├─", "└─") used for stylistic output
	Prefix string
}

// LogDirTree logs the directory contents recursively in a stable sorted order.
// maxEntries caps the number of emitted entries (directories and files combined);
// when the cap is reached, logging is truncated with a note.
func (dl DirLogger) LogDirTree(root string, maxEntries int) {
	dl.Printf("dir: %s", root)
	count := 0
	var walk func(dir, prefix string)
	walk = func(dir, prefix string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			dl.Printf("%s!! (error reading %s): %v", dl.Prefix, dir, err)
			return
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		for i, e := range entries {
			if maxEntries > 0 && count >= maxEntries {
				return
			}
			isLast := i == len(entries)-1
			branch := "├─"
			nextPrefix := prefix + "│ "
			if isLast {
				branch = "└─"
				nextPrefix = prefix + "  "
			}
			name := e.Name()
			path := filepath.Join(dir, name)
			if e.IsDir() {
				dl.Printf("%s%s %s/", prefix, branch, name)
				count++
				if maxEntries > 0 && count >= maxEntries {
					return
				}
				walk(path, nextPrefix)
			} else {
				var sizeStr string
				if info, err := e.Info(); err == nil {
					sizeStr = humanBytes(info.Size())
				}
				dl.Printf("%s%s %s (%s)", prefix, branch, name, sizeStr)
				count++
			}
		}
	}
	walk(root, dl.Prefix)
	if maxEntries > 0 && count >= maxEntries {
		dl.Printf("%s… output truncated at %d entries", dl.Prefix, maxEntries)
	}
}

func humanBytes(n int64) string {
	const unit = 1024.0
	labels := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	d := float64(n)
	idx := 0
	for d >= unit && idx < len(labels)-1 {
		d /= unit
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%d %s", n, labels[idx])
	}
	return fmt.Sprintf("%.1f %s", d, labels[idx])
}

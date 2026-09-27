package tui

import (
	"os"
	"path/filepath"
	"sort"
)

// discoverApps lists the buildable/runnable app names (subdirectories of
// cmd/) relative to the project root at dir, so targets requiring an
// APP=<name> argument (`it`, `build`) can offer real choices instead of a
// hardcoded list.
func discoverApps(dir string) []string {
	entries, err := os.ReadDir(filepath.Join(dir, "cmd"))
	if err != nil {
		return nil
	}
	var apps []string
	for _, e := range entries {
		if e.IsDir() {
			apps = append(apps, e.Name())
		}
	}
	sort.Strings(apps)
	return apps
}

// Package tui implements a terminal UI for discovering, running, and
// inspecting the results of this project's Makefile targets.
package tui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Target describes a single runnable Makefile target.
type Target struct {
	Name        string
	Description string
	Category    string
	// NeedsApp is true when the target's recipe references $(APP), meaning
	// the caller must supply an APP=<name> argument (e.g. `all`, `one`, `sub`).
	NeedsApp bool
}

var (
	targetRe   = regexp.MustCompile(`^([a-zA-Z_0-9-]+):.*?##\s*(.*)$`)
	categoryRe = regexp.MustCompile(`^##@\s*(.*)$`)
)

// FindMakefile walks upward from the current working directory looking for
// a file named "Makefile", so the TUI works from any subdirectory of the
// project, not just the root.
func FindMakefile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "Makefile")
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no Makefile found starting from %s", dir)
		}
		dir = parent
	}
}

// ParseMakefile reads the Makefile at path and returns the documented
// targets (i.e. those with a trailing `## description` comment), in the
// order they're declared, grouped under their most recent `##@ Category`
// header.
func ParseMakefile(path string) ([]Target, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		targets  []Target
		category string
		current  *Target
	)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := categoryRe.FindStringSubmatch(line); m != nil {
			category = strings.TrimSpace(m[1])
			current = nil
			continue
		}

		if m := targetRe.FindStringSubmatch(line); m != nil {
			targets = append(targets, Target{
				Name:        m[1],
				Description: strings.TrimSpace(m[2]),
				Category:    category,
			})
			current = &targets[len(targets)-1]
			continue
		}

		// Recipe lines are indented with a tab and belong to the most
		// recently declared target; scan them for an APP reference.
		if current != nil && strings.HasPrefix(line, "\t") && strings.Contains(line, "$(APP)") {
			current.NeedsApp = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

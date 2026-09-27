// Command tui launches an interactive terminal UI for discovering, running,
// and inspecting the results of this project's Makefile targets.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nelsw/bytelyon-client/internal/tui"
)

func main() {
	makefilePath, err := tui.FindMakefile()
	if err != nil {
		fmt.Fprintln(os.Stderr, "tui:", err)
		os.Exit(1)
	}

	targets, err := tui.ParseMakefile(makefilePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tui:", err)
		os.Exit(1)
	}

	model := tui.NewModel(filepath.Dir(makefilePath), targets)

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui:", err)
		os.Exit(1)
	}
}

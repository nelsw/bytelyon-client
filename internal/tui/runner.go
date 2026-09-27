package tui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// outputLineMsg carries a single line of streamed `make` output.
type outputLineMsg struct {
	runID int
	text  string
}

// runDoneMsg reports the terminal state of a finished run.
type runDoneMsg struct {
	runID    int
	err      error
	exitCode int
	duration time.Duration
}

// activeRun tracks the state needed to observe and, if necessary, kill a
// running `make` invocation.
type activeRun struct {
	id  int
	cmd *exec.Cmd
}

// kill terminates the whole process group spawned for this run, so that
// child processes started by `make` (e.g. go test, golangci-lint) are
// cleaned up too.
func (r *activeRun) kill() {
	if r == nil || r.cmd == nil || r.cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL)
}

// startRun shells out to `make <target> [APP=<app>]` in dir, streaming its
// combined stdout/stderr line-by-line over the returned channel as
// outputLineMsg values, followed by exactly one runDoneMsg. The caller is
// responsible for reading from the channel (e.g. via waitForActivity) until
// the runDoneMsg is observed.
func startRun(dir string, runID int, target Target, app string) (chan tea.Msg, *activeRun) {
	msgs := make(chan tea.Msg)

	args := []string{target.Name}
	if target.NeedsApp && app != "" {
		args = append(args, fmt.Sprintf("APP=%s", app))
	}

	cmd := exec.Command("make", args...)
	cmd.Dir = dir
	// Run in its own process group so we can kill the whole tree on cancel.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	run := &activeRun{id: runID, cmd: cmd}
	start := time.Now()

	go func() {
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			msgs <- outputLineMsg{runID: runID, text: scanner.Text()}
		}
		if err := scanner.Err(); err != nil {
			msgs <- outputLineMsg{runID: runID, text: fmt.Sprintf("[tui] output scan error: %v", err)}
		}
	}()

	go func() {
		var exitCode int
		err := cmd.Start()
		if err != nil {
			exitCode = -1
		} else {
			waitErr := cmd.Wait()
			err = waitErr
			if waitErr != nil {
				if exitErr, ok := waitErr.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					exitCode = -1
				}
			}
		}
		_ = pw.Close()
		msgs <- runDoneMsg{runID: runID, err: err, exitCode: exitCode, duration: time.Since(start)}
	}()

	return msgs, run
}

// waitForActivity returns a tea.Cmd that blocks until the next message
// arrives on msgs, then delivers it into the Bubble Tea event loop.
func waitForActivity(msgs chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-msgs
	}
}

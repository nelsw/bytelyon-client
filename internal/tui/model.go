package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focus int

const (
	focusList focus = iota
	focusOutput
)

type mode int

const (
	modeNormal mode = iota
	modeAppPicker
)

// Model is the top-level Bubble Tea model for the Makefile target runner.
type Model struct {
	rootDir string

	list     list.Model
	viewport viewport.Model
	spinner  spinner.Model

	apps      []string
	appCursor int
	pending   Target

	mode  mode
	focus focus

	runID     int
	running   bool
	hasResult bool
	runTarget Target
	runApp    string
	startedAt time.Time
	duration  time.Duration
	exitCode  int
	runErr    error

	lines []string
	msgs  chan tea.Msg
	run   *activeRun

	width, height int
	ready         bool
}

// NewModel builds the initial Model for the given project root directory
// (where the Makefile lives) and its parsed targets.
func NewModel(rootDir string, targets []Target) Model {
	items := make([]list.Item, len(targets))
	for i, t := range targets {
		items[i] = targetItem{target: t}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(colorAccent).
		BorderLeftForeground(colorAccent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(colorAccent)

	l := list.New(items, delegate, 0, 0)
	l.Title = "Make Targets"
	l.Styles.Title = titleStyle
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = statusRunningStyle

	vp := viewport.New(0, 0)

	return Model{
		rootDir:  rootDir,
		list:     l,
		viewport: vp,
		spinner:  sp,
		apps:     discoverApps(rootDir),
		focus:    focusList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.layout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		if !m.running {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case outputLineMsg:
		if msg.runID != m.runID {
			return m, nil
		}
		m.lines = append(m.lines, msg.text)
		m.viewport.SetContent(strings.Join(m.lines, "\n"))
		m.viewport.GotoBottom()
		return m, waitForActivity(m.msgs)

	case runDoneMsg:
		if msg.runID != m.runID {
			return m, nil
		}
		m.running = false
		m.hasResult = true
		m.duration = msg.duration
		m.exitCode = msg.exitCode
		m.runErr = msg.err
		m.run = nil
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == modeAppPicker {
		switch msg.String() {
		case "esc":
			m.mode = modeNormal
			return m, nil
		case "up", "k":
			if m.appCursor > 0 {
				m.appCursor--
			}
			return m, nil
		case "down", "j":
			if m.appCursor < len(m.apps)-1 {
				m.appCursor++
			}
			return m, nil
		case "enter":
			app := ""
			if len(m.apps) > 0 {
				app = m.apps[m.appCursor]
			}
			m.mode = modeNormal
			return m.beginRun(m.pending, app)
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		m.killActive()
		return m, tea.Quit

	case "q":
		if !m.list.SettingFilter() {
			m.killActive()
			return m, tea.Quit
		}

	case "tab":
		if m.focus == focusList {
			m.focus = focusOutput
		} else {
			m.focus = focusList
		}
		return m, nil

	case "x":
		m.killActive()
		return m, nil

	case "enter":
		if m.focus == focusList && !m.list.SettingFilter() {
			if item, ok := m.list.SelectedItem().(targetItem); ok {
				if item.target.NeedsApp {
					m.mode = modeAppPicker
					m.pending = item.target
					m.appCursor = 0
					return m, nil
				}
				return m.beginRun(item.target, "")
			}
		}
	}

	var cmd tea.Cmd
	if m.focus == focusList {
		m.list, cmd = m.list.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

// beginRun kills any in-flight run, then starts target (with an optional
// APP=<app> argument) and kicks off the goroutines/commands needed to
// stream its output.
func (m Model) beginRun(target Target, app string) (tea.Model, tea.Cmd) {
	m.killActive()

	m.runID++
	m.running = true
	m.hasResult = false
	m.runTarget = target
	m.runApp = app
	m.startedAt = time.Now()
	m.duration = 0
	m.exitCode = 0
	m.runErr = nil
	m.lines = nil
	m.viewport.SetContent("")
	m.viewport.GotoTop()

	msgs, run := startRun(m.rootDir, m.runID, target, app)
	m.msgs = msgs
	m.run = run

	return m, tea.Batch(waitForActivity(msgs), m.spinner.Tick)
}

func (m *Model) killActive() {
	if m.running && m.run != nil {
		m.run.kill()
	}
	m.running = false
}

func (m *Model) layout() {
	if !m.ready {
		return
	}

	const headerHeight = 1
	const footerHeight = 1
	const statusHeight = 1

	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	listWidth := m.width / 3
	if listWidth < 24 {
		listWidth = 24
	}
	if listWidth > 48 {
		listWidth = 48
	}
	outputWidth := m.width - listWidth

	frameW, frameH := paneStyle.GetHorizontalFrameSize(), paneStyle.GetVerticalFrameSize()

	m.list.SetSize(listWidth-frameW, bodyHeight-frameH)

	m.viewport.Width = outputWidth - frameW
	m.viewport.Height = bodyHeight - frameH - statusHeight
	if m.viewport.Height < 1 {
		m.viewport.Height = 1
	}
}

func (m Model) View() string {
	if !m.ready {
		return "loading…"
	}

	header := titleStyle.Render(" bytelyon-client — make target runner ")

	listPane := paneStyle
	outputPane := paneStyle
	if m.focus == focusList {
		listPane = activePaneStyle
	} else {
		outputPane = activePaneStyle
	}

	listView := listPane.Render(m.list.View())

	status := m.renderStatus()
	outputContent := lipgloss.JoinVertical(lipgloss.Left, status, m.viewport.View())
	outputView := outputPane.Render(outputContent)

	body := lipgloss.JoinHorizontal(lipgloss.Top, listView, outputView)

	footer := helpStyle.Render(m.helpText())

	view := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	if m.mode == modeAppPicker {
		return m.renderModal(view)
	}
	return view
}

func (m Model) renderStatus() string {
	if m.running {
		return statusRunningStyle.Render(fmt.Sprintf(
			"%s running `make %s%s`… %s",
			m.spinner.View(), m.runTarget.Name, appSuffix(m.runApp),
			time.Since(m.startedAt).Round(time.Second),
		))
	}
	if m.hasResult {
		dur := m.duration.Round(10 * time.Millisecond)
		if m.exitCode == 0 && m.runErr == nil {
			return statusGoodStyle.Render(fmt.Sprintf(
				"✓ make %s%s succeeded in %s",
				m.runTarget.Name, appSuffix(m.runApp), dur,
			))
		}
		return statusBadStyle.Render(fmt.Sprintf(
			"✗ make %s%s failed (exit %d) in %s",
			m.runTarget.Name, appSuffix(m.runApp), m.exitCode, dur,
		))
	}
	return statusIdleStyle.Render("Select a target and press enter to run it")
}

func (m Model) helpText() string {
	if m.mode == modeAppPicker {
		return "↑/↓ choose  •  enter confirm  •  esc cancel"
	}
	base := "tab switch pane  •  / filter  •  q quit"
	if m.focus == focusList {
		return "↑/↓ navigate  •  enter run  •  x cancel run  •  " + base
	}
	return "↑/↓ or mouse scroll  •  " + base
}

func (m Model) renderModal(background string) string {
	title := modalTitleStyle.Render(fmt.Sprintf("Choose APP for `%s`", m.pending.Name))
	lines := []string{title, ""}

	for i, app := range m.apps {
		style := choiceStyle
		if i == m.appCursor {
			style = choiceActiveStyle
		}
		lines = append(lines, style.Render(app))
	}
	if len(m.apps) == 0 {
		lines = append(lines, helpStyle.Render("no cmd/* apps found"))
	}

	content := modalStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func appSuffix(app string) string {
	if app == "" {
		return ""
	}
	return fmt.Sprintf(" APP=%s", app)
}

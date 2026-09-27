package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary = lipgloss.Color("212") // pink, echoes the Makefile banner
	colorAccent  = lipgloss.Color("111") // blue
	colorMuted   = lipgloss.Color("241")
	colorGood    = lipgloss.Color("42")
	colorBad     = lipgloss.Color("203")
	colorWarn    = lipgloss.Color("221")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(colorPrimary).
			Padding(0, 1)

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	activePaneStyle = paneStyle.
			BorderForeground(colorAccent)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	statusRunningStyle = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
	statusGoodStyle    = lipgloss.NewStyle().Foreground(colorGood).Bold(true)
	statusBadStyle     = lipgloss.NewStyle().Foreground(colorBad).Bold(true)
	statusIdleStyle    = lipgloss.NewStyle().Foreground(colorMuted)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)

	choiceStyle       = lipgloss.NewStyle().Padding(0, 1)
	choiceActiveStyle = lipgloss.NewStyle().Padding(0, 1).
				Foreground(lipgloss.Color("0")).
				Background(colorAccent).
				Bold(true)

	categoryStyle = lipgloss.NewStyle().Foreground(colorAccent)
)

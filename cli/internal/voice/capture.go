package voice

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
)

// CaptureResult holds the outcome of a voice capture session.
type CaptureResult struct {
	Text      string
	Cancelled bool
	Error     error
}

// captureModel is a Bubble Tea model for the voice capture TUI.
type captureModel struct {
	theme          tui.Theme
	listener       *Listener
	transcript     string
	ready          bool
	done           bool
	result         CaptureResult
	spinner        spinner.Model
	silenceTimeout float64
}

// listenerStartedMsg is sent when the dictation helper has been spawned.
type listenerStartedMsg struct {
	listener *Listener
}

// transcriptMsg carries a transcript event from the listener.
type transcriptMsg struct {
	event *TranscriptEvent
	err   error
}

func newCaptureModel(silenceTimeout float64) captureModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tui.DefaultTheme.Primary)

	return captureModel{
		theme:          tui.DefaultTheme,
		spinner:        s,
		silenceTimeout: silenceTimeout,
	}
}

func (m captureModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startListener(),
	)
}

func (m captureModel) startListener() tea.Cmd {
	timeout := m.silenceTimeout
	return func() tea.Msg {
		listener, err := NewListener(timeout)
		if err != nil {
			return transcriptMsg{err: err}
		}
		return listenerStartedMsg{listener: listener}
	}
}

// waitForEvent returns a Cmd that blocks until the next transcript event.
func waitForEvent(l *Listener) tea.Cmd {
	return func() tea.Msg {
		evt, err := l.Next()
		return transcriptMsg{event: evt, err: err}
	}
}

func (m captureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.transcript != "" {
				m.done = true
				m.result = CaptureResult{Text: m.transcript}
				if m.listener != nil {
					go m.listener.Stop() // non-blocking to avoid race
				}
				return m, tea.Quit
			}
			return m, nil
		case "esc", "ctrl+c":
			m.done = true
			m.result = CaptureResult{Cancelled: true}
			if m.listener != nil {
				go m.listener.Stop()
			}
			return m, tea.Quit
		}

	case listenerStartedMsg:
		m.listener = msg.listener
		return m, waitForEvent(m.listener)

	case transcriptMsg:
		// Ignore late messages if we already have a result (e.g., user pressed Enter
		// and the listener EOF races with the quit message).
		if m.done {
			return m, nil
		}

		if msg.err != nil {
			m.done = true
			m.result = CaptureResult{Error: msg.err}
			return m, tea.Quit
		}

		switch msg.event.Type {
		case "ready":
			m.ready = true
			return m, waitForEvent(m.listener)
		case "partial":
			m.transcript = msg.event.Text
			return m, waitForEvent(m.listener)
		case "final":
			m.transcript = msg.event.Text
			m.done = true
			m.result = CaptureResult{Text: m.transcript}
			return m, tea.Quit
		case "error":
			m.done = true
			m.result = CaptureResult{Error: fmt.Errorf("%s", msg.event.Text)}
			return m, tea.Quit
		}
		return m, waitForEvent(m.listener)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m captureModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Render("Voice Capture")
	b.WriteString("\n  " + title + "\n\n")

	if !m.ready {
		b.WriteString("  " + m.spinner.View() + " Initializing microphone...\n")
	} else {
		b.WriteString("  " + m.spinner.View() + " Listening...\n\n")

		if m.transcript != "" {
			textStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5E7EB")).
				PaddingLeft(2)
			b.WriteString(textStyle.Render(m.transcript) + "\n")
		} else {
			b.WriteString("  " + m.theme.MutedStyle().Render("Speak now...") + "\n")
		}
	}

	b.WriteString("\n")

	helpParts := []string{}
	if m.transcript != "" {
		helpParts = append(helpParts, "enter: send")
	}
	helpParts = append(helpParts, "esc: cancel")
	if m.ready && m.transcript != "" {
		helpParts = append(helpParts, fmt.Sprintf("auto-sends after %.0fs silence", m.silenceTimeout))
	}
	b.WriteString("  " + m.theme.HelpStyle().Render(strings.Join(helpParts, " | ")) + "\n")

	return b.String()
}

// Capture runs the voice capture TUI and returns the transcribed text.
// Returns empty string with nil error if the user cancelled.
func Capture(silenceTimeout float64) (string, error) {
	model := newCaptureModel(silenceTimeout)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("running voice capture: %w", err)
	}

	result := finalModel.(captureModel).result
	if result.Error != nil {
		return "", result.Error
	}
	if result.Cancelled {
		return "", nil
	}
	return result.Text, nil
}

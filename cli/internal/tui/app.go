package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// View represents a screen or component that can be rendered.
type View interface {
	// Init returns an initial command to run.
	Init() tea.Cmd

	// Update handles messages and returns the updated view and any commands.
	Update(msg tea.Msg) (View, tea.Cmd)

	// View renders the view as a string.
	View() string
}

// App is the main application model that manages views and theme.
type App struct {
	theme      Theme
	width      int
	height     int
	activeView View
	ready      bool
}

// NewApp creates a new App with the default theme.
func NewApp(initialView View) *App {
	return &App{
		theme:      DefaultTheme,
		activeView: initialView,
	}
}

// NewAppWithTheme creates a new App with a custom theme.
func NewAppWithTheme(initialView View, theme Theme) *App {
	return &App{
		theme:      theme,
		activeView: initialView,
	}
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	if a.activeView != nil {
		return a.activeView.Init()
	}
	return nil
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
	}

	// Pass message to active view
	if a.activeView != nil {
		newView, cmd := a.activeView.Update(msg)
		a.activeView = newView
		return a, cmd
	}

	return a, nil
}

// View implements tea.Model.
func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	if a.activeView != nil {
		return a.activeView.View()
	}

	return ""
}

// SetView changes the active view.
func (a *App) SetView(v View) tea.Cmd {
	a.activeView = v
	return v.Init()
}

// Theme returns the app's theme.
func (a *App) Theme() Theme {
	return a.theme
}

// Width returns the terminal width.
func (a *App) Width() int {
	return a.width
}

// Height returns the terminal height.
func (a *App) Height() int {
	return a.height
}

// SetViewMsg is a message to change the active view.
type SetViewMsg struct {
	View View
}

// QuitMsg signals the app should quit.
type QuitMsg struct{}

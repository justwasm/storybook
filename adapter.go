package storybook

import (
	tea "charm.land/bubbletea/v2"
)

// WrapOption configures model adaptation options.
type WrapOption func(*ModelAdapter)

// WithResetFactory provides a factory to re-instantiate the model upon restart.
func WithResetFactory(fn func() tea.Model) WrapOption {
	return func(a *ModelAdapter) {
		a.resetFactory = fn
	}
}

// WithCustomQuitKey configures custom keys that trigger mock quit.
func WithCustomQuitKey(keys ...string) WrapOption {
	return func(a *ModelAdapter) {
		a.customQuitKeys = keys
	}
}

// ModelAdapter wraps any standard tea.Model into a Storybook Instance with mock quit and restart.
type ModelAdapter struct {
	inner          tea.Model
	resetFactory   func() tea.Model
	quitting       bool
	customQuitKeys []string
	width          int
	height         int
}

// WrapModel wraps a standard Bubble Tea model into an Instance.
func WrapModel(m tea.Model, opts ...WrapOption) *ModelAdapter {
	a := &ModelAdapter{
		inner: m,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Init initializes the inner model.
func (a *ModelAdapter) Init() tea.Cmd {
	if a.inner == nil {
		return nil
	}
	return a.inner.Init()
}

// Update handles messages, intercepts tea.Quit, and manages restart triggers.
func (a *ModelAdapter) Update(msg tea.Msg) (Instance, tea.Cmd) {
	if a.inner == nil {
		return a, nil
	}

	// 1. If currently in simulated exit state, listen for restart keys ('r', 'R', space)
	if a.quitting {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			k := keyMsg.String()
			if k == "r" || k == "R" || k == " " || k == "space" {
				return a, a.Restart()
			}
		}
		// In quitting state, ignore other commands
		return a, nil
	}

	// 2. Check for manual quit triggers if configured
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && len(a.customQuitKeys) > 0 {
		k := keyMsg.String()
		for _, qk := range a.customQuitKeys {
			if k == qk {
				a.quitting = true
				return a, nil
			}
		}
	}

	// 3. Delegate to inner model
	newInner, cmd := a.inner.Update(msg)
	a.inner = newInner

	// 4. Check if inner model attempted to quit (tea.Quit simulated)
	if cmd != nil {
		// Detect tea.Quit command safely by checking tea.Quit() behavior
		if isQuitCmd(cmd) {
			a.quitting = true
			return a, nil
		}
	}

	return a, cmd
}

// View renders the inner model's view, appending the MockQuit banner if exited.
func (a *ModelAdapter) View() string {
	if a.inner == nil {
		return ""
	}
	content := a.inner.View().Content
	if a.quitting {
		content += "\n" + RenderMockQuit()
	}
	return content
}

// SetSize delivers canvas dimension updates to the inner model.
func (a *ModelAdapter) SetSize(width, height int) {
	a.width = width
	a.height = height

	// If inner model implements Resizable directly
	if r, ok := a.inner.(Resizable); ok {
		r.SetSize(width, height)
		return
	}

	// Otherwise, forward as standard tea.WindowSizeMsg
	if a.inner != nil {
		newInner, _ := a.inner.Update(tea.WindowSizeMsg{Width: width, Height: height})
		a.inner = newInner
	}
}

// Restart resets the model state and re-invokes Init().
func (a *ModelAdapter) Restart() tea.Cmd {
	a.quitting = false
	if a.resetFactory != nil {
		a.inner = a.resetFactory()
		if a.width > 0 && a.height > 0 {
			a.SetSize(a.width, a.height)
		}
		return a.Init()
	}
	if r, ok := a.inner.(Restarter); ok {
		return r.Restart()
	}
	return a.Init()
}

// HandleKey implements InteractiveHandler.
func (a *ModelAdapter) HandleKey(key tea.KeyMsg) (tea.Cmd, bool) {
	if a.inner == nil {
		return nil, false
	}
	var cmd tea.Cmd
	_, cmd = a.Update(key)
	return cmd, true
}

// HandleMouse implements MouseHandler.
func (a *ModelAdapter) HandleMouse(ev tea.MouseMsg) (tea.Cmd, bool) {
	if a.inner == nil {
		return nil, false
	}
	var cmd tea.Cmd
	_, cmd = a.Update(ev)
	return cmd, true
}

// isQuitCmd checks whether a tea.Cmd represents tea.Quit.
func isQuitCmd(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	// Execute the command in an isolated manner: tea.Quit returns tea.QuitMsg{}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		return true
	}
	return false
}

// IsQuitting returns whether the adapter is currently in mock-quit state.
func (a *ModelAdapter) IsQuitting() bool {
	return a.quitting
}

// Inner returns the underlying wrapped model.
func (a *ModelAdapter) Inner() tea.Model {
	return a.inner
}

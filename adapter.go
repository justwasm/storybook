package storybook

import (
	"reflect"

	tea "charm.land/bubbletea/v2"
)

var teaQuitPtr = reflect.ValueOf(tea.Quit).Pointer()

// isTeaQuit checks whether a tea.Cmd directly points to tea.Quit without executing it.
func isTeaQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	return reflect.ValueOf(cmd).Pointer() == teaQuitPtr
}

// MockQuitMsg indicates an inner model dispatched a quit command.
type MockQuitMsg struct{}

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

// WithAutoFocus enables or disables automatic focus invocation on focusable models.
func WithAutoFocus(enable bool) WrapOption {
	return func(a *ModelAdapter) {
		a.autoFocus = enable
	}
}

// WithInitialSize sets initial canvas dimensions before initialization.
func WithInitialSize(width, height int) WrapOption {
	return func(a *ModelAdapter) {
		a.width = width
		a.height = height
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
	autoFocus      bool
}

// WrapModel wraps a standard Bubble Tea model into an Instance.
func WrapModel(m tea.Model, opts ...WrapOption) *ModelAdapter {
	a := &ModelAdapter{
		inner:     m,
		autoFocus: true,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// wrapCmd safely wraps a tea.Cmd so that tea.QuitMsg is intercepted asynchronously
// without executing blocking commands or sleeping on the UI event thread.
func (a *ModelAdapter) wrapCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	if isTeaQuit(cmd) {
		return func() tea.Msg {
			return MockQuitMsg{}
		}
	}
	return func() tea.Msg {
		msg := cmd()
		switch m := msg.(type) {
		case tea.QuitMsg:
			return MockQuitMsg{}
		case tea.BatchMsg:
			wrapped := make(tea.BatchMsg, len(m))
			for i, c := range m {
				wrapped[i] = a.wrapCmd(c)
			}
			return wrapped
		default:
			return msg
		}
	}
}

// Init initializes the inner model and dispatches any initial commands or focus triggers.
func (a *ModelAdapter) Init() tea.Cmd {
	if a.inner == nil {
		return nil
	}
	var cmds []tea.Cmd
	if a.width > 0 && a.height > 0 {
		a.SetSize(a.width, a.height)
	}
	if cmd := a.inner.Init(); cmd != nil {
		if isTeaQuit(cmd) {
			a.quitting = true
		} else {
			cmds = append(cmds, a.wrapCmd(cmd))
		}
	}
	if a.autoFocus {
		if cmd := a.triggerFocus(); cmd != nil {
			if isTeaQuit(cmd) {
				a.quitting = true
			} else {
				cmds = append(cmds, a.wrapCmd(cmd))
			}
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	if len(cmds) == 1 {
		return cmds[0]
	}
	return tea.Batch(cmds...)
}

func (a *ModelAdapter) triggerFocus() tea.Cmd {
	if f, ok := a.inner.(Focusable); ok {
		return f.Focus()
	}
	if f, ok := a.inner.(SimpleFocusable); ok {
		f.Focus()
		return nil
	}
	return nil
}

// Update handles messages, intercepts tea.Quit, and manages restart triggers.
func (a *ModelAdapter) Update(msg tea.Msg) (Instance, tea.Cmd) {
	if a.inner == nil {
		return a, nil
	}

	// 0. Intercept MockQuitMsg delivered from background commands
	if _, ok := msg.(MockQuitMsg); ok {
		a.quitting = true
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

	// 4. Check if inner model returned tea.Quit directly
	if cmd != nil {
		if isTeaQuit(cmd) {
			a.quitting = true
			return a, nil
		}
	}

	return a, a.wrapCmd(cmd)
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

// SetState delivers state changes to inner models implementing Stateful.
func (a *ModelAdapter) SetState(stateIndex int) {
	if s, ok := a.inner.(Stateful); ok {
		s.SetState(stateIndex)
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
		cmd := r.Restart()
		if isTeaQuit(cmd) {
			a.quitting = true
			return nil
		}
		return a.wrapCmd(cmd)
	}
	return a.Init()
}

// HandleKey implements InteractiveHandler.
func (a *ModelAdapter) HandleKey(key tea.KeyMsg) (tea.Cmd, bool) {
	if a.inner == nil {
		return nil, false
	}
	if a.autoFocus {
		_ = a.triggerFocus()
	}
	if h, ok := a.inner.(InteractiveHandler); ok {
		cmd, handled := h.HandleKey(key)
		if isTeaQuit(cmd) {
			a.quitting = true
			return nil, handled
		}
		return a.wrapCmd(cmd), handled
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
	if h, ok := a.inner.(MouseHandler); ok {
		cmd, handled := h.HandleMouse(ev)
		if isTeaQuit(cmd) {
			a.quitting = true
			return nil, handled
		}
		return a.wrapCmd(cmd), handled
	}
	var cmd tea.Cmd
	_, cmd = a.Update(ev)
	return cmd, true
}

// IsQuitting returns whether the adapter is currently in mock-quit state.
func (a *ModelAdapter) IsQuitting() bool {
	return a.quitting
}

// Inner returns the underlying wrapped model.
func (a *ModelAdapter) Inner() tea.Model {
	return a.inner
}

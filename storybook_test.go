package storybook

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Dummy model for testing
type testModel struct {
	count    int
	lastMsg  string
	width    int
	height   int
	quitting bool
}

func (m *testModel) Init() tea.Cmd {
	return nil
}

func (m *testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		m.lastMsg = msg.String()
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "inc":
			m.count++
			return m, nil
		}
	}
	return m, nil
}

func (m *testModel) View() tea.View {
	return tea.NewView("TestModel count: 42")
}

func TestRegistry_BasicOperations(t *testing.T) {
	reg := NewRegistry()

	reg.Register(Story{
		ID:          "comp-1",
		Name:        "Component One",
		Category:    CategoryBasics,
		Level:       LevelAtom,
		Description: "First test component",
		SourceCode:  "package main\n\nfunc main() {}",
		Factory: func() Instance {
			return WrapModel(&testModel{})
		},
	})

	reg.Register(Story{
		ID:          "comp-2",
		Name:        "Component Two",
		Category:    CategoryBasics,
		Level:       LevelMolecule,
		Description: "Second test component",
		Factory: func() Instance {
			return WrapModel(&testModel{})
		},
	})

	stories := reg.Stories()
	if len(stories) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(stories))
	}

	s1, ok := reg.Find("comp-1")
	if !ok || s1.Name != "Component One" {
		t.Errorf("failed to find comp-1: got %+v", s1)
	}

	items := reg.VisibleNavItems("")
	// 1 category header + 2 stories = 3 items
	if len(items) != 3 {
		t.Errorf("expected 3 visible nav items, got %d", len(items))
	}

	// Test Search Filter
	filtered := reg.VisibleNavItems("Two")
	if len(filtered) != 2 { // 1 category + 1 matching story
		t.Errorf("expected 2 items matching 'Two', got %d", len(filtered))
	}
}

func TestWrapModel_MockQuitAndRestart(t *testing.T) {
	inner := &testModel{}
	adapter := WrapModel(inner)

	adapter.Init()

	// Initial view should not contain mock quit notice
	initialView := adapter.View()
	if strings.Contains(initialView, MockQuitMessage) {
		t.Errorf("initial view should not contain MockQuitMessage")
	}

	// Deliver 'q' key to model, which returns tea.Quit
	adapter.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})

	if !adapter.IsQuitting() {
		t.Errorf("expected adapter to be in quitting state after tea.Quit")
	}

	quittingView := adapter.View()
	if !strings.Contains(quittingView, MockQuitMessage) {
		t.Errorf("expected quitting view to contain MockQuitMessage, got:\n%s", quittingView)
	}

	// Press 'r' to restart
	adapter.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})

	if adapter.IsQuitting() {
		t.Errorf("expected adapter quitting to be false after 'r' restart")
	}

	restartedView := adapter.View()
	if strings.Contains(restartedView, MockQuitMessage) {
		t.Errorf("restarted view should not contain MockQuitMessage")
	}
}

func TestWrapModel_SetSize(t *testing.T) {
	inner := &testModel{}
	adapter := WrapModel(inner)

	adapter.SetSize(80, 24)
	if inner.width != 80 || inner.height != 24 {
		t.Errorf("expected inner dimensions (80, 24), got (%d, %d)", inner.width, inner.height)
	}
}

func TestModel_NavigationAndTabs(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Story{
		ID:          "test-story",
		Name:        "Test Story",
		Category:    CategoryBasics,
		Level:       LevelExample,
		Description: "A test story",
		SourceCode:  "// Hello World\npackage main\n\nfunc main() {}",
		Doc:         "Architecture details here.",
		Factory: func() Instance {
			return WrapModel(&testModel{})
		},
	})

	m := NewModel(reg)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	if m.ActiveStory() == nil || m.ActiveStory().ID != "test-story" {
		t.Fatalf("expected active story 'test-story', got %+v", m.ActiveStory())
	}

	// 1. Initial tab is Live Demo (0)
	if m.ActiveTab() != TabLiveDemo {
		t.Errorf("expected initial tab LiveDemo, got %d", m.ActiveTab())
	}

	// 2. Press Tab -> Source Code (1)
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveTab() != TabSourceCode {
		t.Errorf("expected tab SourceCode after Tab, got %d", m.ActiveTab())
	}

	// 3. Press Tab -> Architecture & Info (2)
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.ActiveTab() != TabArchitectureInfo {
		t.Errorf("expected tab ArchitectureInfo after second Tab, got %d", m.ActiveTab())
	}

	// 4. Press '1' -> Back to Live Demo
	m.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	if m.ActiveTab() != TabLiveDemo {
		t.Errorf("expected tab LiveDemo after pressing '1', got %d", m.ActiveTab())
	}

	// 5. Test Theme Switching
	initialTheme := m.CurrentTheme().Name
	m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	if m.CurrentTheme().Name == initialTheme {
		t.Errorf("expected theme to change after pressing 't'")
	}

	// 6. Test Rendering
	view := m.RenderViewString()
	if !strings.Contains(view, "Test Story") {
		t.Errorf("expected view to contain story name, got:\n%s", view)
	}
	if !strings.Contains(view, "TestModel count: 42") {
		t.Errorf("expected view to contain inner model output, got:\n%s", view)
	}
}

func TestModel_CrossLayoutAndAltScreen(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Story{
		ID:          "simple",
		Name:        "simple",
		Category:    CategoryBasics,
		Description: "A simple story",
		Factory: func() Instance {
			return WrapModel(&testModel{})
		},
	})

	m := NewModel(reg)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	v := m.View()
	if !v.AltScreen {
		t.Errorf("expected v.AltScreen to be true for clean full-terminal presentation")
	}

	lines := strings.Split(v.Content, "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines, got %d", len(lines))
	}

	// Line 0 is top title bar
	if !strings.Contains(lines[0], "Storybook") {
		t.Errorf("line 0 should contain top title bar, got: %s", lines[0])
	}

	// Line 1 is header row with STORIES and breadcrumb
	if !strings.Contains(lines[1], "STORIES") || !strings.Contains(lines[1], "│") {
		t.Errorf("line 1 should contain STORIES and divider │, got: %s", lines[1])
	}

	// Line 2 is the horizontal cross bar of the "十字"!
	// It should contain ─── on both sides of the vertical divider │
	if !strings.Contains(lines[2], "───") || !strings.Contains(lines[2], "│") {
		t.Errorf("line 2 should be the horizontal cross bar containing ─── and │, got: %s", lines[2])
	}

	// Bottom row contains tabs and theme indicator
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "Tokyo Night") {
		t.Errorf("bottom line should contain theme indicator, got: %s", lastLine)
	}

	// Test quitting resets AltScreen
	m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	quitView := m.View()
	if quitView.AltScreen {
		t.Errorf("expected quitView.AltScreen to be false upon exit")
	}
}

func TestModel_MouseWheelAndNoLineOverflow(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Story{
		ID:          "story-1",
		Name:        "Story One",
		Category:    CategoryBasics,
		Description: "First story",
		SourceCode:  "line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8\nline 9\nline 10\nline 11\nline 12",
		Factory: func() Instance {
			return WrapModel(&testModel{})
		},
	})
	reg.Register(Story{
		ID:          "story-2",
		Name:        "Story Two",
		Category:    CategoryBasics,
		Description: "Second story",
		SourceCode:  "package main\n\nfunc main() {\n\t// Very long line that should be clamped and never bleed into left sidebar or wrap onto column 0\n\tprintln(\"1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890\")\n}",
		Factory: func() Instance {
			return WrapModel(&testModel{})
		},
	})

	m := NewModel(reg)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 1. Mouse wheel down in sidebar moves selection
	m.Update(tea.MouseWheelMsg(tea.Mouse{X: 10, Y: 10, Button: tea.MouseWheelDown}))
	if m.ActiveStory() == nil || m.ActiveStory().ID != "story-2" {
		t.Errorf("expected mouse wheel down in sidebar to navigate to story-2, got %v", m.ActiveStory())
	}

	// 2. Mouse wheel up in sidebar moves back
	m.Update(tea.MouseWheelMsg(tea.Mouse{X: 10, Y: 10, Button: tea.MouseWheelUp}))
	if m.ActiveStory() == nil || m.ActiveStory().ID != "story-1" {
		t.Errorf("expected mouse wheel up in sidebar to navigate to story-1, got %v", m.ActiveStory())
	}

	// 3. Switch to source code tab and verify mouse wheel scrolls source code
	m.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	if m.ActiveTab() != TabSourceCode {
		t.Fatalf("expected active tab SourceCode")
	}

	m.Update(tea.MouseWheelMsg(tea.Mouse{X: 50, Y: 10, Button: tea.MouseWheelDown}))
	if m.sourceScroll != 3 {
		t.Errorf("expected sourceScroll to be 3 after wheel down, got %d", m.sourceScroll)
	}

	m.Update(tea.MouseWheelMsg(tea.Mouse{X: 50, Y: 10, Button: tea.MouseWheelUp}))
	if m.sourceScroll != 0 {
		t.Errorf("expected sourceScroll to be 0 after wheel up, got %d", m.sourceScroll)
	}

	// 4. Verify no line in full output exceeds width (100) or contains \r
	view := m.RenderViewString()
	lines := strings.Split(view, "\n")
	for row, l := range lines {
		if strings.Contains(l, "\r") {
			t.Errorf("row %d contains forbidden carriage return \\r: %q", row, l)
		}
		w := ansi.StringWidth(l)
		if w > 100 {
			t.Errorf("row %d visible width %d exceeds layout.TotalWidth 100: %q", row, w, l)
		}
	}
}

type tickerTestModel struct {
	tickCount int
	stateIdx  int
	lastMouse tea.Mouse
}

func (t *tickerTestModel) Init() tea.Cmd {
	return func() tea.Msg { return "test-tick-dispatched" }
}

func (t *tickerTestModel) Update(msg tea.Msg) (Instance, tea.Cmd) {
	if str, ok := msg.(string); ok && str == "test-tick-dispatched" {
		t.tickCount++
	}
	return t, nil
}

func (t *tickerTestModel) View() string {
	return "TickerTestModel"
}

func (t *tickerTestModel) SetState(idx int) {
	t.stateIdx = idx
}

func (t *tickerTestModel) HandleMouse(ev tea.MouseMsg) (tea.Cmd, bool) {
	t.lastMouse = ev.Mouse()
	return nil, true
}

func TestModel_DynamicTickerLifecycle(t *testing.T) {
	reg := NewRegistry()
	inst := &tickerTestModel{}

	reg.Register(Story{
		ID:       "s1",
		Name:     "Static Story",
		Category: CategoryBasics,
		Factory: func() Instance {
			return WrapModel(&testModel{})
		},
	})

	reg.Register(Story{
		ID:       "s2",
		Name:     "Ticker Story",
		Category: CategoryBasics,
		Factory: func() Instance {
			return inst
		},
	})

	m := NewModel(reg)
	// Initially on s1
	if m.ActiveStory().ID != "s1" {
		t.Fatalf("expected initial story s1, got %s", m.ActiveStory().ID)
	}

	// Move down to s2
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'j'}))
	if cmd == nil {
		t.Fatalf("expected loadActiveStory() to return Init() command for s2, got nil")
	}

	// Execute command and verify tick message
	msg := cmd()
	if str, ok := msg.(string); !ok || str != "test-tick-dispatched" {
		t.Fatalf("expected tick message, got %+v", msg)
	}

	// Deliver to model
	m.Update(msg)
	if inst.tickCount != 1 {
		t.Fatalf("expected inst.tickCount == 1, got %d", inst.tickCount)
	}
}

func TestModel_DualAxisStateVsTabSwitching(t *testing.T) {
	reg := NewRegistry()
	inst := &tickerTestModel{}

	reg.Register(Story{
		ID:       "multi-state",
		Name:     "Multi-State Component",
		Category: CategoryBasics,
		Level:    LevelAtom,
		States: []State{
			{ID: "s0", Name: "State Zero"},
			{ID: "s1", Name: "State One"},
			{ID: "s2", Name: "State Two"},
			{ID: "s3", Name: "State Three"},
		},
		Factory: func() Instance {
			return inst
		},
	})

	reg.Register(Story{
		ID:       "single-state",
		Name:     "Single State App",
		Category: CategoryBasics,
		Factory: func() Instance {
			return WrapModel(&testModel{})
		},
	})

	m := NewModel(reg)
	// 1. In multi-state story, pressing '2' selects State One (index 1), NOT SourceCode tab!
	m.Update(tea.KeyPressMsg(tea.Key{Code: '2'}))
	if m.stateIdx != 1 {
		t.Fatalf("expected stateIdx == 1, got %d", m.stateIdx)
	}
	if inst.stateIdx != 1 {
		t.Fatalf("expected inst.stateIdx == 1, got %d", inst.stateIdx)
	}
	if m.ActiveTab() != TabLiveDemo {
		t.Fatalf("expected activeTab to remain TabLiveDemo, got %v", m.ActiveTab())
	}

	// Pressing '4' selects State Three (index 3)
	m.Update(tea.KeyPressMsg(tea.Key{Code: '4'}))
	if m.stateIdx != 3 {
		t.Fatalf("expected stateIdx == 3, got %d", m.stateIdx)
	}

	// Pressing 'v' cycles view modality to SourceCode
	m.Update(tea.KeyPressMsg(tea.Key{Code: 'v'}))
	if m.ActiveTab() != TabSourceCode {
		t.Fatalf("expected TabSourceCode after pressing v, got %v", m.ActiveTab())
	}

	// Pressing 'v' again cycles to ArchitectureInfo
	m.Update(tea.KeyPressMsg(tea.Key{Code: 'v'}))
	if m.ActiveTab() != TabArchitectureInfo {
		t.Fatalf("expected TabArchitectureInfo after pressing v, got %v", m.ActiveTab())
	}

	// Pressing 'd' jumps back to LiveDemo
	m.Update(tea.KeyPressMsg(tea.Key{Code: 'd'}))
	if m.ActiveTab() != TabLiveDemo {
		t.Fatalf("expected TabLiveDemo after pressing d, got %v", m.ActiveTab())
	}

	// 2. Now navigate to single-state story
	m.Update(tea.KeyPressMsg(tea.Key{Code: 'j'}))
	if m.ActiveStory().ID != "single-state" {
		t.Fatalf("expected active story single-state, got %s", m.ActiveStory().ID)
	}

	// In single-state story, pressing '2' selects TabSourceCode
	m.Update(tea.KeyPressMsg(tea.Key{Code: '2'}))
	if m.ActiveTab() != TabSourceCode {
		t.Fatalf("expected TabSourceCode for single-state story pressing '2', got %v", m.ActiveTab())
	}
}

func TestModel_MouseCoordinateTranslation(t *testing.T) {
	reg := NewRegistry()
	inst := &tickerTestModel{}

	reg.Register(Story{
		ID:       "mouse-story",
		Name:     "Mouse Story",
		Category: CategoryInput,
		Factory: func() Instance {
			return inst
		},
	})

	m := NewModel(reg)
	m.width = 100
	m.height = 30
	layout := ComputeLayout(m.width, m.height)

	// Screen coords: sidebarWidth + 1 + 10, row 3 + 5
	screenX := layout.SidebarWidth + 1 + 10
	screenY := 3 + 5

	m.Update(tea.MouseClickMsg(tea.Mouse{X: screenX, Y: screenY, Button: tea.MouseLeft}))

	// Verify inst received local coords (10, 5)
	if inst.lastMouse.X != 10 || inst.lastMouse.Y != 5 {
		t.Fatalf("expected translated mouse coords (10, 5), got (%d, %d)", inst.lastMouse.X, inst.lastMouse.Y)
	}
}

type focusableTestModel struct {
	focused bool
	resized bool
	w, h    int
}

func (m *focusableTestModel) Init() tea.Cmd {
	return nil
}

func (m *focusableTestModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *focusableTestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.resized = true
		m.w = ws.Width
		m.h = ws.Height
	}
	return m, nil
}

func (m *focusableTestModel) View() tea.View {
	return tea.NewView("Focusable Content")
}

func TestRegisterModel_ZeroInvasiveAndAutoFocus(t *testing.T) {
	origReg := DefaultRegistry
	DefaultRegistry = NewRegistry()
	defer func() { DefaultRegistry = origReg }()

	var capturedModel *focusableTestModel
	story := RegisterModel("focus-demo", "Focus Demo", CategoryForms, func() tea.Model {
		m := &focusableTestModel{}
		capturedModel = m
		return m
	},
		WithDescription("Tests zero-invasive auto focus and sizing"),
		WithLevel(LevelAtom),
	)

	if story.ID != "focus-demo" || story.Level != LevelAtom {
		t.Fatalf("expected registered story ID 'focus-demo' and LevelAtom, got %s / %v", story.ID, story.Level)
	}

	inst := story.Factory()
	adapter, ok := inst.(*ModelAdapter)
	if !ok {
		t.Fatalf("expected instance to be *ModelAdapter, got %T", inst)
	}

	// Sizing delivered
	adapter.SetSize(80, 24)
	if !capturedModel.resized || capturedModel.w != 80 || capturedModel.h != 24 {
		t.Fatalf("expected model resized to 80x24, got %v (%dx%d)", capturedModel.resized, capturedModel.w, capturedModel.h)
	}

	// Auto-focus triggered on Init()
	adapter.Init()
	if !capturedModel.focused {
		t.Fatalf("expected focusable model to be auto-focused")
	}
}

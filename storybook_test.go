package storybook

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	if !strings.Contains(lines[0], "Theme:") {
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

	// Test quitting resets AltScreen
	m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	quitView := m.View()
	if quitView.AltScreen {
		t.Errorf("expected quitView.AltScreen to be false upon exit")
	}
}

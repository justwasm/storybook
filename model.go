package storybook

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model is the top-level Elm Architecture model for Storybook.
type Model struct {
	registry       *Registry
	activeIdx      int
	activeStory    *Story
	activeInstance Instance
	activeTab      ViewTab
	stateIdx       int
	sourceScroll   int
	interactive    bool
	searching      bool
	filterQuery    string
	themeIdx       int
	width          int
	height         int
	sidebarWidth   int
	copyNotice     string

	// Mouse text selection for Source Code tab
	selecting    bool
	hasSelection bool
	selStartLine int
	selStartCol  int
	selEndLine   int
	selEndCol    int
}

// NewModel constructs a Storybook model with the given registry (defaults to DefaultRegistry).
func NewModel(reg ...*Registry) *Model {
	r := DefaultRegistry
	if len(reg) > 0 && reg[0] != nil {
		r = reg[0]
	}

	m := &Model{
		registry:     r,
		activeTab:    TabLiveDemo,
		themeIdx:     0,
		sidebarWidth: 32,
		width:        100,
		height:       30,
	}

	// Auto-select first story
	items := m.registry.VisibleNavItems("")
	for i, it := range items {
		if !it.IsCategory && it.Story != nil {
			m.activeIdx = i
			m.activeStory = it.Story
			m.loadActiveStory()
			break
		}
	}

	return m
}

// CurrentTheme returns the active Theme.
func (m *Model) CurrentTheme() Theme {
	if m.themeIdx < 0 || m.themeIdx >= len(AllThemes) {
		return ThemeTokyoNight
	}
	return AllThemes[m.themeIdx]
}

// Init initializes the active story instance if present.
func (m *Model) Init() tea.Cmd {
	if m.activeInstance != nil {
		if initier, ok := m.activeInstance.(Initializer); ok {
			return initier.Init()
		}
	}
	return nil
}

// Update processes incoming messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateDimensions()
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward non-key messages to active instance (e.g. Tickers, custom msgs)
	if m.activeInstance != nil {
		var cmd tea.Cmd
		m.activeInstance, cmd = m.activeInstance.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) updateDimensions() {
	if m.width > 120 {
		m.sidebarWidth = 34
	} else if m.width > 90 {
		m.sidebarWidth = 30
	} else {
		m.sidebarWidth = 24
	}

	if m.activeInstance != nil {
		canvasW := m.width - m.sidebarWidth - 1
		canvasH := m.height - 4
		if canvasW > 0 && canvasH > 0 {
			if r, ok := m.activeInstance.(Resizable); ok {
				r.SetSize(canvasW, canvasH)
			}
		}
	}
}

func (m *Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// 1. Interactive Playground Mode
	if m.interactive {
		if key == "esc" {
			m.interactive = false
			return m, nil
		}

		if m.activeInstance != nil {
			if h, ok := m.activeInstance.(InteractiveHandler); ok {
				cmd, consumed := h.HandleKey(msg)
				if consumed {
					return m, cmd
				}
			}
			var cmd tea.Cmd
			m.activeInstance, cmd = m.activeInstance.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// 2. Search Box Mode
	if m.searching {
		switch key {
		case "esc":
			m.searching = false
			m.filterQuery = ""
			return m, nil
		case "enter":
			m.searching = false
			return m, nil
		case "backspace":
			if len(m.filterQuery) > 0 {
				m.filterQuery = m.filterQuery[:len(m.filterQuery)-1]
				m.adjustSelectionAfterFilter()
			}
			return m, nil
		default:
			if len(key) == 1 {
				m.filterQuery += key
				m.adjustSelectionAfterFilter()
			}
			return m, nil
		}
	}

	// 3. Normal Sidebar Navigation Mode
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		m.moveNav(-1)
		return m, nil

	case "down", "j":
		m.moveNav(1)
		return m, nil

	case "enter", "l":
		items := m.registry.VisibleNavItems(m.filterQuery)
		if m.activeIdx >= 0 && m.activeIdx < len(items) {
			it := items[m.activeIdx]
			if it.IsCategory {
				m.registry.ToggleCategory(it.Category)
			} else if it.Story != nil {
				m.interactive = true
			}
		}
		return m, nil

	case "1":
		m.activeTab = TabLiveDemo
		return m, nil
	case "2":
		m.activeTab = TabSourceCode
		return m, nil
	case "3":
		m.activeTab = TabArchitectureInfo
		return m, nil

	case "tab":
		// If story has state presets in Live Demo, cycle states; otherwise cycle tabs
		if m.activeTab == TabLiveDemo && m.activeStory != nil && len(m.activeStory.States) > 1 {
			m.stateIdx = (m.stateIdx + 1) % len(m.activeStory.States)
			if s, ok := m.activeInstance.(Stateful); ok {
				s.SetState(m.stateIdx)
			}
		} else {
			m.activeTab = (m.activeTab + 1) % 3
		}
		return m, nil

	case "/":
		m.searching = true
		return m, nil

	case "t", "T":
		m.themeIdx = (m.themeIdx + 1) % len(AllThemes)
		return m, nil

	case "pgup":
		if m.activeTab == TabSourceCode {
			m.sourceScroll -= 15
			if m.sourceScroll < 0 {
				m.sourceScroll = 0
			}
		}
		return m, nil

	case "pgdown":
		if m.activeTab == TabSourceCode {
			m.sourceScroll += 15
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) moveNav(delta int) {
	items := m.registry.VisibleNavItems(m.filterQuery)
	if len(items) == 0 {
		return
	}

	newIdx := m.activeIdx + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(items) {
		newIdx = len(items) - 1
	}

	m.activeIdx = newIdx
	it := items[newIdx]
	if !it.IsCategory && it.Story != nil && it.Story != m.activeStory {
		m.activeStory = it.Story
		m.stateIdx = 0
		m.sourceScroll = 0
		m.copyNotice = ""
		m.loadActiveStory()
	}
}

func (m *Model) adjustSelectionAfterFilter() {
	items := m.registry.VisibleNavItems(m.filterQuery)
	if len(items) == 0 {
		m.activeIdx = -1
		m.activeStory = nil
		m.activeInstance = nil
		return
	}

	for i, it := range items {
		if !it.IsCategory && it.Story != nil {
			m.activeIdx = i
			m.activeStory = it.Story
			m.stateIdx = 0
			m.sourceScroll = 0
			m.loadActiveStory()
			return
		}
	}
	m.activeIdx = 0
}

func (m *Model) loadActiveStory() {
	if m.activeStory == nil || m.activeStory.Factory == nil {
		m.activeInstance = nil
		return
	}

	m.activeInstance = m.activeStory.Factory()

	canvasW := m.width - m.sidebarWidth - 1
	canvasH := m.height - 4
	if canvasW > 0 && canvasH > 0 {
		if r, ok := m.activeInstance.(Resizable); ok {
			r.SetSize(canvasW, canvasH)
		}
	}
}

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	mouse := msg.Mouse()

	// If in Live Demo tab and interactive, forward mouse to active instance
	if m.activeTab == TabLiveDemo && m.interactive && m.activeInstance != nil {
		if mh, ok := m.activeInstance.(MouseHandler); ok {
			cmd, _ := mh.HandleMouse(msg)
			return m, cmd
		}
	}

	// If in Source Code tab, handle scrolling and drag selection
	if m.activeTab == TabSourceCode && m.activeStory != nil && m.activeStory.SourceCode != "" {
		switch mouse.Button {
		case tea.MouseWheelUp:
			m.sourceScroll -= 3
			if m.sourceScroll < 0 {
				m.sourceScroll = 0
			}
			return m, nil
		case tea.MouseWheelDown:
			m.sourceScroll += 3
			return m, nil
		}

		// Drag-to-select and copy text
		if mouse.X > m.sidebarWidth {
			switch msg.(type) {
			case tea.MouseClickMsg:
				if mouse.Button == tea.MouseLeft {
					m.selecting = true
					m.hasSelection = false
					m.selStartLine = m.sourceScroll + (mouse.Y - 2)
					m.selEndLine = m.selStartLine
					return m, nil
				}
			case tea.MouseReleaseMsg:
				if m.selecting {
					m.selecting = false
					m.selEndLine = m.sourceScroll + (mouse.Y - 2)
					m.copySelectedSourceText()
					return m, nil
				}
			}
		}
	}

	return m, nil
}

func (m *Model) copySelectedSourceText() {
	if m.activeStory == nil || m.activeStory.SourceCode == "" {
		return
	}
	lines := strings.Split(m.activeStory.SourceCode, "\n")
	start := m.selStartLine
	end := m.selEndLine
	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}
	if end >= len(lines) {
		end = len(lines) - 1
	}

	if start <= end {
		selectedText := strings.Join(lines[start:end+1], "\n")
		if strings.TrimSpace(selectedText) != "" {
			_ = clipboard.WriteAll(selectedText)
			m.copyNotice = fmt.Sprintf("✓ Copied %d lines to clipboard", end-start+1)
		}
	}
}

// View renders the dual-pane storybook layout.
func (m *Model) View() tea.View {
	return tea.NewView(m.RenderViewString())
}

// RenderViewString renders the dual-pane storybook layout as a plain string.
func (m *Model) RenderViewString() string {
	theme := m.CurrentTheme()
	items := m.registry.VisibleNavItems(m.filterQuery)

	sidebarH := m.height - 1
	if sidebarH < 5 {
		sidebarH = 5
	}
	sidebarView := NewSidebar(m.sidebarWidth, sidebarH, theme).
		Render(items, m.activeIdx, m.filterQuery, m.searching)

	canvasW := m.width - m.sidebarWidth - 1
	if canvasW < 10 {
		canvasW = 10
	}
	canvasH := m.height - 4
	if canvasH < 5 {
		canvasH = 5
	}

	canvas := NewCanvas(canvasW, canvasH, theme)

	// Header + Tabs
	headerLine := canvas.RenderHeader(m.activeStory, m.interactive, theme)
	var states []State
	if m.activeStory != nil {
		states = m.activeStory.States
	}
	tabsLine := canvas.RenderTabs(m.activeTab, m.stateIdx, states)

	// Content
	var content string
	switch m.activeTab {
	case TabLiveDemo:
		if m.activeInstance != nil {
			content = m.activeInstance.View()
		} else {
			content = lipgloss.NewStyle().Foreground(theme.Muted).Render("  (No component instance loaded)")
		}
	case TabSourceCode:
		src := ""
		if m.activeStory != nil {
			src = m.activeStory.SourceCode
		}
		content = canvas.RenderSourceCode(src, m.sourceScroll, m.copyNotice, canvasH)
	case TabArchitectureInfo:
		content = canvas.RenderInfoTab(m.activeStory)
	}

	// Footer
	footerLine := canvas.RenderFooter(m.interactive)

	// Combine right-hand canvas view
	rightPaneLines := []string{
		headerLine,
		tabsLine,
		lipgloss.NewStyle().Foreground(theme.Border).Render(strings.Repeat("─", canvasW)),
	}

	contentLines := strings.Split(content, "\n")
	for i := 0; i < canvasH; i++ {
		if i < len(contentLines) {
			rightPaneLines = append(rightPaneLines, PadOrTruncate(contentLines[i], canvasW))
		} else {
			rightPaneLines = append(rightPaneLines, strings.Repeat(" ", canvasW))
		}
	}
	rightPaneLines = append(rightPaneLines, footerLine)

	rightPane := strings.Join(rightPaneLines, "\n")

	// Join Left Sidebar and Right Pane with vertical divider
	sepStyle := lipgloss.NewStyle().Foreground(theme.Border)
	vsep := sepStyle.Render("│")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		sidebarView,
		vsep,
		rightPane,
	)
}

// ActiveStory returns the currently selected Story.
func (m *Model) ActiveStory() *Story {
	return m.activeStory
}

// ActiveTab returns the currently active ViewTab.
func (m *Model) ActiveTab() ViewTab {
	return m.activeTab
}

package storybook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	tea "charm.land/bubbletea/v2"
)

// Model is the top-level Elm Architecture model for Storybook.
type Model struct {
	title          string
	registry       *Registry
	activeIdx      int
	sidebarScroll  int
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
	quitting       bool
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
		title:     "❖ Storybook",
		registry:  r,
		activeTab: TabLiveDemo,
		themeIdx:  0,
		width:     100,
		height:    30,
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

// SetTitle configures the window title prefix.
func (m *Model) SetTitle(title string) {
	m.title = title
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
	layout := ComputeLayout(m.width, m.height)
	if m.activeInstance != nil {
		if r, ok := m.activeInstance.(Resizable); ok {
			r.SetSize(layout.CanvasWidth, layout.CanvasContentHeight)
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
			cmd := m.adjustSelectionAfterFilter()
			return m, cmd
		case "enter":
			m.searching = false
			return m, nil
		case "backspace":
			if len(m.filterQuery) > 0 {
				m.filterQuery = m.filterQuery[:len(m.filterQuery)-1]
				cmd := m.adjustSelectionAfterFilter()
				return m, cmd
			}
			return m, nil
		default:
			if len(key) == 1 {
				m.filterQuery += key
				cmd := m.adjustSelectionAfterFilter()
				return m, cmd
			}
			return m, nil
		}
	}

	// 3. Normal Sidebar Navigation Mode
	switch key {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		if m.filterQuery != "" {
			m.filterQuery = ""
			m.adjustSelectionAfterFilter()
		}
		return m, nil

	case "up", "k":
		cmd := m.moveNav(-1)
		return m, cmd

	case "down", "j":
		cmd := m.moveNav(1)
		return m, cmd

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

	case "v", "V":
		m.activeTab = (m.activeTab + 1) % 3
		return m, nil

	case "d", "D":
		m.activeTab = TabLiveDemo
		return m, nil

	case "s", "S":
		m.activeTab = TabSourceCode
		return m, nil

	case "i", "I":
		m.activeTab = TabArchitectureInfo
		return m, nil

	case "y", "Y":
		if m.activeTab == TabSourceCode {
			m.copyAllSourceText()
			return m, nil
		}
		return m, nil

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		hasStoryStates := m.activeStory != nil && len(m.activeStory.States) > 1
		if hasStoryStates {
			idx := int(key[0] - '1')
			if idx < len(m.activeStory.States) {
				m.stateIdx = idx
				if s, ok := m.activeInstance.(Stateful); ok {
					s.SetState(m.stateIdx)
				}
				m.activeTab = TabLiveDemo
			}
			return m, nil
		}

		// When story has no states (e.g. standalone showcase), 1/2/3 switch tabs
		switch key {
		case "1":
			m.activeTab = TabLiveDemo
		case "2":
			m.activeTab = TabSourceCode
		case "3":
			m.activeTab = TabArchitectureInfo
		}
		return m, nil

	case "tab":
		hasStoryStates := m.activeStory != nil && len(m.activeStory.States) > 1
		if hasStoryStates && m.activeTab == TabLiveDemo {
			m.stateIdx = (m.stateIdx + 1) % len(m.activeStory.States)
			if s, ok := m.activeInstance.(Stateful); ok {
				s.SetState(m.stateIdx)
			}
		} else {
			m.activeTab = (m.activeTab + 1) % 3
		}
		return m, nil

	case "shift+tab":
		hasStoryStates := m.activeStory != nil && len(m.activeStory.States) > 1
		if hasStoryStates && m.activeTab == TabLiveDemo {
			m.stateIdx = (m.stateIdx - 1 + len(m.activeStory.States)) % len(m.activeStory.States)
			if s, ok := m.activeInstance.(Stateful); ok {
				s.SetState(m.stateIdx)
			}
		} else {
			m.activeTab = (m.activeTab + 2) % 3
		}
		return m, nil

	case "r", "R", " ":
		if re, ok := m.activeInstance.(Restarter); ok {
			return m, re.Restart()
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

func (m *Model) moveNav(delta int) tea.Cmd {
	items := m.registry.VisibleNavItems(m.filterQuery)
	if len(items) == 0 {
		return nil
	}

	newIdx := m.activeIdx + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(items) {
		newIdx = len(items) - 1
	}

	m.activeIdx = newIdx

	// Sync scroll window with selection
	layout := ComputeLayout(m.width, m.height)
	availRows := layout.BodyHeight - 2
	if availRows < 1 {
		availRows = 1
	}
	if m.activeIdx < m.sidebarScroll {
		m.sidebarScroll = m.activeIdx
	}
	if m.activeIdx >= m.sidebarScroll+availRows {
		m.sidebarScroll = m.activeIdx - availRows + 1
	}

	it := items[newIdx]
	if !it.IsCategory && it.Story != nil && it.Story != m.activeStory {
		m.activeStory = it.Story
		m.stateIdx = 0
		m.sourceScroll = 0
		m.copyNotice = ""
		return m.loadActiveStory()
	}
	return nil
}

func (m *Model) adjustSelectionAfterFilter() tea.Cmd {
	items := m.registry.VisibleNavItems(m.filterQuery)
	if len(items) == 0 {
		m.activeIdx = -1
		m.activeStory = nil
		m.activeInstance = nil
		return nil
	}

	for i, it := range items {
		if !it.IsCategory && it.Story != nil {
			m.activeIdx = i
			m.activeStory = it.Story
			m.stateIdx = 0
			m.sourceScroll = 0
			return m.loadActiveStory()
		}
	}
	m.activeIdx = 0
	return nil
}

func (m *Model) loadActiveStory() tea.Cmd {
	if m.activeStory == nil || m.activeStory.Factory == nil {
		m.activeInstance = nil
		return nil
	}

	m.activeInstance = m.activeStory.Factory()

	layout := ComputeLayout(m.width, m.height)
	if r, ok := m.activeInstance.(Resizable); ok {
		r.SetSize(layout.CanvasWidth, layout.CanvasContentHeight)
	}
	if s, ok := m.activeInstance.(Stateful); ok {
		s.SetState(m.stateIdx)
	}

	if initier, ok := m.activeInstance.(Initializer); ok {
		return initier.Init()
	}
	return nil
}

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	mouse := msg.Mouse()
	layout := ComputeLayout(m.width, m.height)
	items := m.registry.VisibleNavItems(m.filterQuery)

	// 1. Mouse in Left Sidebar (mouse.X < layout.SidebarWidth)
	if mouse.X < layout.SidebarWidth {
		switch mouse.Button {
		case tea.MouseWheelUp:
			cmd := m.moveNav(-1)
			return m, cmd
		case tea.MouseWheelDown:
			cmd := m.moveNav(1)
			return m, cmd
		case tea.MouseLeft:
			// Row 0 is top title, row 1 is header ("STORIES"), row 2 is divider ("───").
			// Items start at screen Y = 3.
			rowOffset := mouse.Y - 3
			if rowOffset >= 0 {
				targetIdx := m.sidebarScroll + rowOffset
				if targetIdx >= 0 && targetIdx < len(items) {
					m.activeIdx = targetIdx
					it := items[targetIdx]
					if it.IsCategory {
						m.registry.ToggleCategory(it.Category)
					} else if it.Story != nil && it.Story != m.activeStory {
						m.activeStory = it.Story
						m.stateIdx = 0
						m.sourceScroll = 0
						m.copyNotice = ""
						cmd := m.loadActiveStory()
						return m, cmd
					}
				}
			}
			return m, nil
		}
		return m, nil
	}

	// 2. Mouse over bottom controls line (mouse.Y >= m.height - 1)
	if mouse.Y >= m.height-1 && mouse.Button == tea.MouseLeft {
		relX := mouse.X - layout.SidebarWidth - 1
		hasStates := m.activeStory != nil && len(m.activeStory.States) > 1

		if hasStates && m.activeTab == TabLiveDemo {
			curX := 1
			for i, st := range m.activeStory.States {
				badgeLen := len(fmt.Sprintf("[%d] %s", i+1, st.Name)) + 4
				if relX >= curX && relX < curX+badgeLen {
					m.stateIdx = i
					if s, ok := m.activeInstance.(Stateful); ok {
						s.SetState(m.stateIdx)
					}
					return m, nil
				}
				curX += badgeLen
			}
			if relX >= curX {
				m.activeTab = (m.activeTab + 1) % 3
				return m, nil
			}
		} else if m.activeTab == TabSourceCode || m.activeTab == TabArchitectureInfo {
			if relX >= 0 && relX < 24 {
				m.activeTab = TabLiveDemo
				return m, nil
			}
		} else {
			if relX >= 0 && relX < 16 {
				m.activeTab = TabLiveDemo
				return m, nil
			} else if relX >= 16 && relX < 34 {
				m.activeTab = TabSourceCode
				return m, nil
			} else if relX >= 34 && relX < 60 {
				m.activeTab = TabArchitectureInfo
				return m, nil
			}
		}
		return m, nil
	}

	// 3. Mouse Wheel in Right Canvas
	switch mouse.Button {
	case tea.MouseWheelUp:
		if m.activeTab == TabSourceCode {
			m.sourceScroll -= 3
			if m.sourceScroll < 0 {
				m.sourceScroll = 0
			}
			return m, nil
		}
	case tea.MouseWheelDown:
		if m.activeTab == TabSourceCode {
			m.sourceScroll += 3
			return m, nil
		}
	}

	// 4. Mouse inside Component Viewport (X > SidebarWidth, Y in [3 .. m.height - 2])
	if mouse.X > layout.SidebarWidth && mouse.Y >= 3 && mouse.Y < m.height-1 {
		originX := layout.SidebarWidth + 1
		originY := 3
		localMsg := TranslateMouseMsg(msg, -originX, -originY)

		if m.activeTab == TabLiveDemo && m.activeInstance != nil {
			if mouse.Button == tea.MouseLeft && !m.interactive {
				m.interactive = true
			}

			if mh, ok := m.activeInstance.(MouseHandler); ok {
				cmd, handled := mh.HandleMouse(localMsg)
				if handled {
					return m, cmd
				}
			}
			if m.interactive {
				var cmd tea.Cmd
				m.activeInstance, cmd = m.activeInstance.Update(localMsg)
				return m, cmd
			}
		}

		// Drag-to-select and copy text in Source Code tab
		if m.activeTab == TabSourceCode && m.activeStory != nil {
			switch msg.(type) {
			case tea.MouseClickMsg:
				if mouse.Button == tea.MouseLeft {
					m.selecting = true
					m.hasSelection = false
					m.selStartLine = m.sourceScroll + (mouse.Y - 3)
					m.selEndLine = m.selStartLine
					return m, nil
				}
			case tea.MouseMotionMsg:
				if m.selecting {
					m.selEndLine = m.sourceScroll + (mouse.Y - 3)
					return m, nil
				}
			case tea.MouseReleaseMsg:
				if m.selecting {
					m.selecting = false
					m.selEndLine = m.sourceScroll + (mouse.Y - 3)
					m.copySelectedSourceText()
					return m, nil
				}
			}
		}
	}

	return m, nil
}

func (m *Model) copySelectedSourceText() {
	src := ""
	if m.activeStory != nil {
		src = m.activeStory.SourceCode
		if src == "" && m.activeStory.UpstreamRef != "" {
			src = loadUpstreamSource(m.activeStory.UpstreamRef)
		}
	}
	if src == "" {
		return
	}
	lines := strings.Split(src, "\n")
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

func (m *Model) copyAllSourceText() {
	src := ""
	if m.activeStory != nil {
		src = m.activeStory.SourceCode
		if src == "" && m.activeStory.UpstreamRef != "" {
			src = loadUpstreamSource(m.activeStory.UpstreamRef)
		}
	}
	if strings.TrimSpace(src) != "" {
		_ = clipboard.WriteAll(src)
		lines := strings.Split(src, "\n")
		m.copyNotice = fmt.Sprintf("✓ Copied all %d lines to clipboard", len(lines))
	}
}

// View renders the dual-pane storybook layout with Altscreen enabled.
func (m *Model) View() tea.View {
	if m.quitting {
		v := tea.NewView("")
		v.AltScreen = false
		return v
	}

	fullView := m.RenderViewString()
	v := tea.NewView(fullView)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	v.DisableBracketedPasteMode = false
	return v
}

// RenderViewString renders the dual-pane storybook layout as a plain string.
func (m *Model) RenderViewString() string {
	layout := ComputeLayout(m.width, m.height)
	theme := m.CurrentTheme()
	items := m.registry.VisibleNavItems(m.filterQuery)

	// Clean Top Title Bar (Row 0 across full TotalWidth)
	var modeIndicator string
	if m.interactive {
		modeIndicator = green(bold("● INTERACTIVE "))
	} else {
		modeIndicator = dim("○ NAV ")
	}

	titleText := fmt.Sprintf(" %s %s%s", m.title, modeIndicator, dim("— Component Catalog & Playground"))
	topBar := PadOrTruncateANSI(titleText, layout.TotalWidth)

	// Left: Sidebar
	sidebarContent := RenderSidebar(items, m.activeIdx, m.sidebarScroll, layout.SidebarWidth, layout.BodyHeight, m.searching, m.filterQuery, m.interactive, theme)

	// Right: Canvas
	viewBody := ""
	if m.activeInstance != nil {
		viewBody = m.activeInstance.View()
	}
	srcCode := ""
	if m.activeStory != nil {
		srcCode = m.activeStory.SourceCode
		if srcCode == "" && m.activeStory.UpstreamRef != "" {
			srcCode = loadUpstreamSource(m.activeStory.UpstreamRef)
		}
	}
	canvasContent := RenderCanvas(m.activeStory, m.activeTab, m.stateIdx, m.interactive, theme, viewBody, srcCode, m.sourceScroll, m.copyNotice, layout.CanvasWidth, layout.BodyHeight)

	// Combine Columns via JoinColumns: creates the continuous "十字" (Cross) at Body Row 1!
	body := JoinColumns(sidebarContent, layout.SidebarWidth, canvasContent, layout.CanvasWidth, layout.BodyHeight, dim("│"))

	fullView := topBar + "\n" + body
	lines := strings.Split(fullView, "\n")
	if len(lines) > m.height {
		lines = lines[:m.height]
		fullView = strings.Join(lines, "\n")
	}

	return fullView
}

// ActiveStory returns the currently selected Story.
func (m *Model) ActiveStory() *Story {
	return m.activeStory
}

// ActiveTab returns the currently active ViewTab.
func (m *Model) ActiveTab() ViewTab {
	return m.activeTab
}

func loadUpstreamSource(ref string) string {
	candidates := []string{
		ref,
		filepath.Join(ref, "main.go"),
		filepath.Join(ref, "plugin.go"),
	}

	prefixes := []string{"", ".", "..", "../..", "../../.."}
	for _, pref := range prefixes {
		for _, cand := range candidates {
			target := filepath.Join(pref, cand)
			if data, err := os.ReadFile(target); err == nil && len(data) > 0 {
				return string(data)
			}
		}
	}
	return ""
}

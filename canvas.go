package storybook

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// ViewTab represents the active inspection tab on the right canvas.
type ViewTab int

const (
	TabLiveDemo ViewTab = iota
	TabSourceCode
	TabArchitectureInfo
)

// Canvas handles rendering the component viewport, source code viewer, and architecture info tab.
type Canvas struct {
	width  int
	height int
	theme  Theme
}

// NewCanvas creates a new canvas viewport manager.
func NewCanvas(width, height int, theme Theme) *Canvas {
	return &Canvas{
		width:  width,
		height: height,
		theme:  theme,
	}
}

// RenderHeader renders the breadcrumbs and theme indicator at the top of the canvas.
func (c *Canvas) RenderHeader(story *Story, interactive bool, theme Theme) string {
	w := c.width
	if w <= 0 {
		return ""
	}

	left := ""
	if story != nil {
		modePill := ""
		if interactive {
			modePill = " " + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1a1b26")).
				Background(theme.Accent).
				Bold(true).
				Render("INTERACTIVE · Esc to return")
		}
		catName := string(story.Category)
		left = fmt.Sprintf(" %s > %s%s", catName, story.Name, modePill)
	}

	right := fmt.Sprintf("Theme: %s [T] ", theme.Name)
	if story != nil && story.UpstreamRef != "" {
		right = fmt.Sprintf("[%s]  %s", story.UpstreamRef, right)
	}

	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Fg).
		Background(theme.HeaderBg)

	line := left + strings.Repeat(" ", gap) + right
	return headerStyle.Render(PadOrTruncate(line, w))
}

// RenderTabs renders the 3-state navigation tabs (`[1] Live Demo`, `[2] Source Code`, `[3] Architecture & Info`).
func (c *Canvas) RenderTabs(activeTab ViewTab, stateIdx int, states []State) string {
	w := c.width
	if w <= 0 {
		return ""
	}

	tabDemo := " [1] Live Demo "
	tabSource := " [2] Source Code "
	tabInfo := " [3] Architecture & Info "

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(c.theme.SelectedFg).
		Background(c.theme.SelectedBg)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(c.theme.Muted)

	switch activeTab {
	case TabLiveDemo:
		tabDemo = activeStyle.Render(tabDemo)
		tabSource = inactiveStyle.Render(tabSource)
		tabInfo = inactiveStyle.Render(tabInfo)
	case TabSourceCode:
		tabDemo = inactiveStyle.Render(tabDemo)
		tabSource = activeStyle.Render(tabSource)
		tabInfo = inactiveStyle.Render(tabInfo)
	case TabArchitectureInfo:
		tabDemo = inactiveStyle.Render(tabDemo)
		tabSource = inactiveStyle.Render(tabSource)
		tabInfo = activeStyle.Render(tabInfo)
	}

	tabsLine := tabDemo + " " + tabSource + " " + tabInfo

	// If there are state presets, show them on the right
	if len(states) > 0 && activeTab == TabLiveDemo {
		statePill := fmt.Sprintf("State: [%d/%d] %s ", stateIdx+1, len(states), states[stateIdx].Name)
		gap := w - lipgloss.Width(tabsLine) - len(statePill)
		if gap > 0 {
			tabsLine += strings.Repeat(" ", gap) + lipgloss.NewStyle().Foreground(c.theme.Accent).Render(statePill)
		}
	}

	return PadOrTruncate(tabsLine, w)
}

// RenderSourceCode renders the formatted source code with line number gutters and scroll offset.
func (c *Canvas) RenderSourceCode(sourceCode string, scrollOffset int, copyNotice string, availableH int) string {
	w := c.width
	if w <= 0 || availableH <= 0 {
		return ""
	}

	if strings.TrimSpace(sourceCode) == "" {
		return lipgloss.NewStyle().Foreground(c.theme.Muted).Render("  (No source code embedded for this component)")
	}

	// Expand tabs to 4 spaces for consistent terminal alignment
	expanded := strings.ReplaceAll(sourceCode, "\t", "    ")
	rawLines := strings.Split(expanded, "\n")
	totalLines := len(rawLines)

	if scrollOffset >= totalLines {
		scrollOffset = totalLines - 1
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	maxDigits := len(fmt.Sprintf("%d", totalLines))
	gutterFmt := fmt.Sprintf("%%%dd │ ", maxDigits)

	var lines []string

	if copyNotice != "" {
		noticeStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#73daca")).
			Background(c.theme.SelectedBg)
		lines = append(lines, PadOrTruncate(noticeStyle.Render(" "+copyNotice+" "), w))
	}

	for i := 0; i < availableH-len(lines); i++ {
		lineIdx := scrollOffset + i
		if lineIdx < totalLines {
			gutter := lipgloss.NewStyle().
				Foreground(c.theme.CodeGutterFg).
				Background(c.theme.CodeGutterBg).
				Render(fmt.Sprintf(gutterFmt, lineIdx+1))

			codeText := rawLines[lineIdx]
			rendered := gutter + codeText
			lines = append(lines, PadOrTruncate(rendered, w))
		} else {
			lines = append(lines, strings.Repeat(" ", w))
		}
	}

	return strings.Join(lines, "\n")
}

// RenderInfoTab renders metadata, component documentation, atomic design role, and shortcuts.
func (c *Canvas) RenderInfoTab(story *Story) string {
	w := c.width
	if w <= 0 || story == nil {
		return ""
	}

	var lines []string

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(c.theme.Accent)
	boldLabel := func(k, v string) string {
		return fmt.Sprintf("  %s %s", lipgloss.NewStyle().Bold(true).Foreground(c.theme.Fg).Render(k), v)
	}

	lines = append(lines, "")
	lines = append(lines, titleStyle.Render(fmt.Sprintf("  ❖ %s", story.Name)))
	lines = append(lines, "")
	lines = append(lines, boldLabel("ID:          ", story.ID))
	lines = append(lines, boldLabel("Category:    ", string(story.Category)))
	if story.Level != "" {
		lines = append(lines, boldLabel("Atomic Level:", fmt.Sprintf("%s %s", story.Level.LevelTag(), story.Level)))
	}
	if story.UpstreamRef != "" {
		lines = append(lines, boldLabel("Upstream Ref:", story.UpstreamRef))
	}
	lines = append(lines, boldLabel("Description: ", story.Description))
	lines = append(lines, "")

	if story.Doc != "" {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(c.theme.CategoryFg).Render("  ── Architecture & Details ──"))
		lines = append(lines, "")
		for _, docLine := range strings.Split(story.Doc, "\n") {
			lines = append(lines, "  "+docLine)
		}
		lines = append(lines, "")
	}

	if len(story.States) > 0 {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(c.theme.CategoryFg).Render("  ── Preset States ──"))
		for i, st := range story.States {
			lines = append(lines, fmt.Sprintf("  [%d] %s: %s", i+1, st.Name, st.Description))
		}
		lines = append(lines, "")
	}

	lines = append(lines, lipgloss.NewStyle().Foreground(c.theme.Muted).Render("  ── Shortcuts ──"))
	lines = append(lines, "  • Enter / l: Enter interactive playground")
	lines = append(lines, "  • 1 / 2 / 3: Switch view tabs (Demo, Source, Info)")
	lines = append(lines, "  • Tab:       Cycle state presets or tabs")
	lines = append(lines, "  • t:         Switch theme")
	lines = append(lines, "  • Esc:       Exit playground back to sidebar")

	var output []string
	for _, l := range lines {
		output = append(output, PadOrTruncate(l, w))
	}
	return strings.Join(output, "\n")
}

// RenderFooter renders the bottom status bar.
func (c *Canvas) RenderFooter(interactive bool) string {
	w := c.width
	if w <= 0 {
		return ""
	}

	statusText := "  [1] Live Demo   [2] Source Code   [3] Architecture & Info   "
	if interactive {
		statusText += lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#73daca")).
			Render("● INTERACTIVE: Press Esc to return to sidebar")
	} else {
		statusText += lipgloss.NewStyle().
			Foreground(c.theme.Muted).
			Render("↑/↓: Nav • Enter: Interact • /: Search • t: Theme • q: Quit")
	}

	statusStyle := lipgloss.NewStyle().
		Background(c.theme.StatusBg).
		Foreground(c.theme.StatusFg)

	return statusStyle.Render(PadOrTruncate(statusText, w))
}

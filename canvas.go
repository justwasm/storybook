package storybook

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// ViewTab represents the active inspection tab on the right canvas.
type ViewTab int

const (
	TabLiveDemo ViewTab = iota
	TabSourceCode
	TabArchitectureInfo
)

// RenderCanvas draws the preview canvas area with breadcrumb header, component viewport, and controls bar.
func RenderCanvas(story *Story, activeTab ViewTab, stateIdx int, interactive bool, theme Theme, viewContent string, sourceCode string, sourceScroll int, copyNotice string, width, height int) string {
	if story == nil {
		return renderEmptyCanvas(width, height)
	}

	var lines []string

	// 1. Breadcrumb & upstream header bar (Row 0)
	var focusIcon string
	var focusHint string
	if interactive {
		focusIcon = green(bold("● "))
		focusHint = green(bold("(INTERACTIVE · Esc to return)"))
	} else {
		focusIcon = dim("○ ")
		focusHint = dim("(Enter to interact)")
	}

	catStr := dim(string(story.Category) + " > ")
	nameStr := bold(cyan(story.Name))
	levelBadge := ""
	if story.Level != "" {
		levelBadge = " " + yellow(story.Level.ShortTag())
	}
	pathStr := ""
	if story.UpstreamRef != "" {
		pathStr = dim("[" + story.UpstreamRef + "]")
	}

	headerLeft := " " + focusIcon + catStr + nameStr + levelBadge + " " + focusHint
	headerRight := pathStr + " "
	if activeTab == TabSourceCode {
		headerRight = dim("[Scroll: ↑↓/PgDn/Wheel]") + " " + pathStr + " "
	}

	headerLeftW := ansi.StringWidth(headerLeft)
	headerRightW := ansi.StringWidth(headerRight)

	gapW := width - headerLeftW - headerRightW
	if gapW < 1 {
		gapW = 1
	}
	headerRow := headerLeft + strings.Repeat(" ", gapW) + headerRight
	lines = append(lines, headerRow)

	// Row 1: The right horizontal arm of the "十字" (Cross)
	sepLine := dim(strings.Repeat("─", width))
	if interactive {
		sepLine = cyan(strings.Repeat("─", width))
	}
	lines = append(lines, sepLine)

	// Content Rows = height - header(1) - topSep(1) - bottomSep(1) - controls(1) = height - 4
	contentHeight := height - 4
	if contentHeight < 2 {
		contentHeight = 2
	}

	// 2. Component Content Area (Rows 2 .. height - 3)
	var bodyText string
	switch activeTab {
	case TabLiveDemo:
		bodyText = viewContent
	case TabSourceCode:
		bodyText = renderSourceCodeLines(sourceCode, sourceScroll, copyNotice, width, contentHeight)
	case TabArchitectureInfo:
		bodyText = renderInfoContent(story, width, contentHeight)
	}

	contentLines := strings.Split(bodyText, "\n")
	for r := 0; r < contentHeight; r++ {
		if r < len(contentLines) {
			lines = append(lines, contentLines[r])
		} else {
			lines = append(lines, "")
		}
	}

	// 3. Separator above controls (Row height - 2)
	lines = append(lines, dim(strings.Repeat("─", width)))

	// 4. Controls & States line (Row height - 1)
	tabDemo := "[1] Live Demo"
	tabSource := "[2] Source Code"
	tabInfo := "[3] Architecture & Info"

	switch activeTab {
	case TabLiveDemo:
		tabDemo = reverse(bold(" " + tabDemo + " "))
		tabSource = dim(tabSource)
		tabInfo = dim(tabInfo)
	case TabSourceCode:
		tabDemo = dim(tabDemo)
		tabSource = reverse(bold(" " + tabSource + " "))
		tabInfo = dim(tabInfo)
	case TabArchitectureInfo:
		tabDemo = dim(tabDemo)
		tabSource = dim(tabSource)
		tabInfo = reverse(bold(" " + tabInfo + " "))
	}

	tabsBar := " " + tabDemo + "  " + tabSource + "  " + tabInfo

	// If there are state presets, append them
	if len(story.States) > 0 && activeTab == TabLiveDemo {
		stateBadge := fmt.Sprintf("State: [%d/%d] %s", stateIdx+1, len(story.States), story.States[stateIdx].Name)
		tabsBar += "  " + yellow(stateBadge)
	}

	modeBadge := ""
	if interactive {
		if activeTab == TabSourceCode {
			modeBadge = green(bold(" ● SOURCE: Drag to Auto-Copy | Esc: Back "))
		} else {
			modeBadge = green(bold(" ● INTERACTIVE: Press Esc to return "))
		}
	} else {
		if activeTab == TabSourceCode {
			modeBadge = dim(" ○ [Enter: Focus & Copy | Tab: View | q: Quit] ")
		} else {
			modeBadge = dim(" ○ [Enter: Interact | Tab: View/State | /: Filter | q: Quit] ")
		}
	}

	tabsW := ansi.StringWidth(tabsBar)
	modeW := ansi.StringWidth(modeBadge)
	ctrlGap := width - tabsW - modeW
	if ctrlGap < 1 {
		ctrlGap = 1
	}
	controlsLine := tabsBar + strings.Repeat(" ", ctrlGap) + modeBadge
	lines = append(lines, controlsLine)

	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

func renderSourceCodeLines(sourceCode string, scrollOffset int, copyNotice string, width, height int) string {
	if strings.TrimSpace(sourceCode) == "" {
		return dim("  (No source code embedded for this component)")
	}

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
		lines = append(lines, green(bold(" "+copyNotice+" ")))
	}

	for i := 0; i < height-len(lines); i++ {
		lineIdx := scrollOffset + i
		if lineIdx < totalLines {
			gutter := dim(fmt.Sprintf(gutterFmt, lineIdx+1))
			lines = append(lines, gutter+rawLines[lineIdx])
		} else {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

func renderInfoContent(story *Story, width, height int) string {
	if story == nil {
		return ""
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, bold(cyan(fmt.Sprintf("  ❖ %s", story.Name))))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s %s", bold("ID:          "), story.ID))
	lines = append(lines, fmt.Sprintf("  %s %s", bold("Category:    "), string(story.Category)))
	if story.Level != "" {
		lines = append(lines, fmt.Sprintf("  %s %s (%s)", bold("Atomic Level:"), yellow(story.Level.LevelTag()), story.Level))
	}
	if story.UpstreamRef != "" {
		lines = append(lines, fmt.Sprintf("  %s %s", bold("Upstream Ref:"), story.UpstreamRef))
	}
	lines = append(lines, fmt.Sprintf("  %s %s", bold("Description: "), story.Description))
	lines = append(lines, "")

	if story.Doc != "" {
		lines = append(lines, bold(blue("  ── Architecture & Details ──")))
		lines = append(lines, "")
		for _, docLine := range strings.Split(story.Doc, "\n") {
			lines = append(lines, "  "+docLine)
		}
		lines = append(lines, "")
	}

	if len(story.States) > 0 {
		lines = append(lines, bold(blue("  ── Preset States ──")))
		for i, st := range story.States {
			lines = append(lines, fmt.Sprintf("  [%d] %s: %s", i+1, st.Name, st.Description))
		}
		lines = append(lines, "")
	}

	lines = append(lines, dim("  ── Shortcuts ──"))
	lines = append(lines, dim("  • Enter / l: Enter interactive playground"))
	lines = append(lines, dim("  • 1 / 2 / 3: Switch view tabs (Demo, Source, Info)"))
	lines = append(lines, dim("  • Tab:       Cycle state presets or tabs"))
	lines = append(lines, dim("  • t:         Switch theme"))
	lines = append(lines, dim("  • Esc:       Exit playground back to sidebar"))

	return strings.Join(lines, "\n")
}

func renderEmptyCanvas(width, height int) string {
	var lines []string
	header := bold(cyan(" PREVIEW CANVAS"))
	lines = append(lines, " "+header)
	lines = append(lines, dim(strings.Repeat("─", width)))
	emptyMsg := dim("No story selected. Use Up/Down or / to filter stories.")
	lines = append(lines, " "+emptyMsg)
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines[:height], "\n")
}

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

// RenderCanvas draws the preview canvas area with breadcrumb header, component viewport, and bottom controls bar.
func RenderCanvas(story *Story, activeTab ViewTab, stateIdx int, interactive bool, theme Theme, viewContent string, sourceCode string, sourceScroll int, copyNotice string, width, height int) string {
	if story == nil {
		return renderEmptyCanvas(width, height)
	}

	var lines []string

	// 1. Clean, minimal breadcrumb & upstream header bar (Row 0)
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
		levelBadge = " " + yellow(story.Level.LevelTag())
	}
	pathStr := ""
	if story.UpstreamRef != "" {
		pathStr = dim("[" + story.UpstreamRef + "]")
	}

	headerLeft := " " + focusIcon + catStr + nameStr + levelBadge + " " + focusHint
	headerRight := pathStr + " "

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
	// Tabs and State indicators live cleanly on the bottom row
	var controlsLeft string
	hasStates := len(story.States) > 1

	if hasStates && activeTab == TabLiveDemo {
		// Render state buttons: [1] PresetA  [2] PresetB  [3] PresetC
		var stateBadges []string
		for i, st := range story.States {
			badge := fmt.Sprintf("[%d] %s", i+1, st.Name)
			if i == stateIdx {
				badge = reverse(bold(" " + badge + " "))
			} else {
				badge = dim(badge)
			}
			stateBadges = append(stateBadges, badge)
		}
		controlsLeft = " " + strings.Join(stateBadges, "  ")
	} else if activeTab == TabSourceCode {
		controlsLeft = " " + reverse(bold(" [v/d] Back to Demo ")) + "  " + dim("[y: Copy All | Drag: Select Lines]")
	} else if activeTab == TabArchitectureInfo {
		controlsLeft = " " + reverse(bold(" [v/d] Back to Demo ")) + "  " + dim("[s: Source Code]")
	} else {
		// Standalone story or single-state story: render the 3 tabs [1] Demo  [2] Source  [3] Info
		tabDemo := "[1] Live Demo"
		tabSource := "[2] Source Code"
		tabInfo := "[3] Architecture & Info"
		if width < 85 {
			tabDemo = "[1] Demo"
			tabSource = "[2] Source"
			tabInfo = "[3] Info"
		}

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
		controlsLeft = " " + tabDemo + "  " + tabSource + "  " + tabInfo
	}

	// Move Theme tag and Mode indicators to the right
	themePill := cyan(fmt.Sprintf("[%s (t)]", theme.Name))
	var modeBadge string

	if interactive {
		modeBadge = fmt.Sprintf(" %s %s ", themePill, green(bold("● INTERACTIVE: Esc to return")))
	} else if hasStates && activeTab == TabLiveDemo {
		viewPill := dim("│ View: [v/s] Source  [v/i] Info")
		if width < 115 {
			viewPill = dim("│ [v: View]")
		}
		modeBadge = fmt.Sprintf(" %s %s %s ", viewPill, themePill, dim("○ [1..9: State | Enter: Focus | q: Quit]"))
	} else {
		if width < 85 {
			modeBadge = fmt.Sprintf(" %s ", themePill)
		} else {
			modeBadge = fmt.Sprintf(" %s %s ", themePill, dim("○ [Enter: Focus | Tab: Cycle | q: Quit]"))
		}
	}

	tabsW := ansi.StringWidth(controlsLeft)
	modeW := ansi.StringWidth(modeBadge)
	ctrlGap := width - tabsW - modeW
	if ctrlGap < 1 {
		ctrlGap = 1
	}
	controlsLine := controlsLeft + strings.Repeat(" ", ctrlGap) + modeBadge
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

	// Normalize CRLF to LF and expand tabs to 4 spaces
	cleanCode := strings.ReplaceAll(sourceCode, "\r", "")
	expanded := strings.ReplaceAll(cleanCode, "\t", "    ")
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

	// Fix: evaluate remaining lines count once, rendering all available rows
	remainingRows := height - len(lines)
	for r := 0; r < remainingRows; r++ {
		lineIdx := scrollOffset + r
		if lineIdx < totalLines {
			codeLine := rawLines[lineIdx]
			gutter := dim(fmt.Sprintf(gutterFmt, lineIdx+1))
			// Truncate codeLine so that gutter + codeLine never overflows width
			availableWidth := width - ansi.StringWidth(gutter)
			if availableWidth > 0 && ansi.StringWidth(codeLine) > availableWidth {
				codeLine = ansi.Truncate(codeLine, availableWidth, "")
			}
			lines = append(lines, gutter+codeLine)
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

	if len(story.KeyFeatures) > 0 {
		lines = append(lines, bold(blue("  ── Key Features ──")))
		for _, kf := range story.KeyFeatures {
			lines = append(lines, fmt.Sprintf("  • %s", kf))
		}
		lines = append(lines, "")
	}

	lines = append(lines, dim("  ── Shortcuts ──"))
	lines = append(lines, dim("  • Enter / l: Enter interactive playground mode"))
	lines = append(lines, dim("  • 1..9:      Directly select component state preset (when available)"))
	lines = append(lines, dim("  • Tab:       Cycle state presets or view tabs"))
	lines = append(lines, dim("  • v / V:     Cycle view modality (Live Demo -> Source Code -> Info)"))
	lines = append(lines, dim("  • d, s, i:   Jump to Demo, Source, or Architecture Info"))
	lines = append(lines, dim("  • t / T:     Cycle visual color theme"))
	lines = append(lines, dim("  • Esc:       Exit interactive playground back to sidebar"))

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

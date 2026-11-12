package storybook

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Layout holds the computed geometry of the dual-pane Storybook TUI.
type Layout struct {
	TotalWidth          int
	TotalHeight         int
	SidebarWidth        int
	CanvasWidth         int
	BodyHeight          int // Content rows between top header and terminal edge
	CanvasContentHeight int // Exact rows available for component viewport (BodyHeight - 4)
}

// ComputeLayout calculates dual-pane dimensions based on terminal width and height.
func ComputeLayout(width, height int) Layout {
	if width < 40 {
		width = 40
	}
	if height < 12 {
		height = 12
	}

	// Sidebar takes ~26-30% of screen, clamped between 26 and 38 columns.
	sidebarW := width * 28 / 100
	if sidebarW < 26 {
		sidebarW = 26
	}
	if sidebarW > 38 {
		sidebarW = 38
	}
	if sidebarW > width-24 {
		sidebarW = width - 24
	}

	// 1 column for the vertical divider.
	canvasW := width - sidebarW - 1
	if canvasW < 10 {
		canvasW = 10
	}

	// Top title bar (1 row) + bottom line (1 row) = 2 rows
	bodyH := height - 2
	if bodyH < 4 {
		bodyH = 4
	}

	// Component content rows = bodyH - header(1) - topSep(1) - bottomSep(1) - controls(1) = bodyH - 4
	contentH := bodyH - 4
	if contentH < 2 {
		contentH = 2
	}

	return Layout{
		TotalWidth:          width,
		TotalHeight:         height,
		SidebarWidth:        sidebarW,
		CanvasWidth:         canvasW,
		BodyHeight:          bodyH,
		CanvasContentHeight: contentH,
	}
}

// JoinColumns renders two multi-line column strings side-by-side with an exact vertical divider on every row.
func JoinColumns(left string, leftWidth int, right string, rightWidth int, height int, dividerChar string) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	var b strings.Builder
	for row := 0; row < height; row++ {
		var lStr, rStr string
		if row < len(leftLines) {
			lStr = leftLines[row]
		}
		if row < len(rightLines) {
			rStr = rightLines[row]
		}

		// Pad or truncate left column to exact leftWidth using visible ANSI column width
		lPadded := PadOrTruncateANSI(lStr, leftWidth)
		// Truncate right column to exact rightWidth
		rPadded := PadOrTruncateANSI(rStr, rightWidth)

		b.WriteString(lPadded)
		b.WriteString(dividerChar)
		b.WriteString(rPadded)
		if row < height-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// PadOrTruncateANSI pads or truncates a string to fit exact width according to terminal ANSI visible width.
func PadOrTruncateANSI(s string, w int) string {
	if w <= 0 {
		return ""
	}
	sw := ansi.StringWidth(s)
	if sw > w {
		return ansi.Truncate(s, w, "")
	}
	if sw < w {
		return s + strings.Repeat(" ", w-sw)
	}
	return s
}

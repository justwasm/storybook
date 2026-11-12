package storybook

import (
	"fmt"
	"strings"
)

// RenderSidebar draws the navigation sidebar with category headers and showcase story rows.
func RenderSidebar(items []NavItem, selectedIdx int, scrollOffset int, width, height int, filtering bool, filterText string, interactive bool, theme Theme) string {
	var lines []string

	// Row 0: Search filter prompt or title
	header := bold(cyan(" STORIES"))
	if filtering || filterText != "" {
		header = yellow(" /"+filterText) + dim(" (esc clears)")
	}
	lines = append(lines, header)

	// Row 1: The left horizontal arm of the "十字" (Cross)
	lines = append(lines, dim(strings.Repeat("─", width)))

	// Available content rows for items (height - 2 header rows)
	availRows := height - 2
	if availRows < 1 {
		availRows = 1
	}

	// Clamp scroll window around selectedIdx
	if selectedIdx < scrollOffset {
		scrollOffset = selectedIdx
	}
	if selectedIdx >= scrollOffset+availRows {
		scrollOffset = selectedIdx - availRows + 1
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	endIdx := scrollOffset + availRows
	if endIdx > len(items) {
		endIdx = len(items)
	}

	for i := scrollOffset; i < endIdx; i++ {
		item := items[i]
		isSel := i == selectedIdx

		var row string
		if item.IsCategory {
			arrow := "▼"
			catLabel := fmt.Sprintf(" %s %s", arrow, bold(string(item.Category)))
			if isSel {
				row = reverse(catLabel)
			} else {
				row = catLabel
			}
		} else {
			prefix := "    • "
			badge := ""
			if item.Story.Level != "" {
				badge = " " + item.Story.Level.ShortTag()
			}
			if isSel {
				indicator := "○"
				if interactive {
					indicator = "●"
				}
				row = reverse("  " + indicator + " " + item.Story.Name + badge)
			} else {
				row = prefix + item.Story.Name + dim(badge)
			}
		}
		lines = append(lines, row)
	}

	// Pad remaining height with empty rows
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

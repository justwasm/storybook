package storybook

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// Sidebar renders the left navigation panel containing categories, stories, and filter box.
type Sidebar struct {
	width  int
	height int
	theme  Theme
}

// NewSidebar creates a new sidebar renderer.
func NewSidebar(width, height int, theme Theme) *Sidebar {
	return &Sidebar{
		width:  width,
		height: height,
		theme:  theme,
	}
}

// Render renders the sidebar content given the navigation items, selected index, and search query.
func (s *Sidebar) Render(items []NavItem, selectedIdx int, filterQuery string, searching bool) string {
	w := s.width
	if w <= 0 {
		return ""
	}

	var lines []string

	// 1. Header Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(s.theme.Accent)
	header := titleStyle.Render("STORIES")
	if filterQuery != "" {
		header += fmt.Sprintf(" (%d)", countStories(items))
	}
	lines = append(lines, PadOrTruncate(header, w))

	// Separator
	sep := lipgloss.NewStyle().Foreground(s.theme.Border).Render(strings.Repeat("─", w))
	lines = append(lines, sep)

	// 2. Filter input row
	if searching || filterQuery != "" {
		filterPrompt := "🔍 " + filterQuery
		if searching {
			filterPrompt += "█"
		}
		filterStyle := lipgloss.NewStyle().
			Foreground(s.theme.Accent).
			Background(s.theme.SelectedBg)
		lines = append(lines, PadOrTruncate(filterStyle.Render(filterPrompt), w))
	}

	// 3. Story items with scroll window calculation
	headerRows := len(lines)
	availableRows := s.height - headerRows
	if availableRows < 3 {
		availableRows = 3
	}

	scrollOffset := 0
	if selectedIdx >= availableRows {
		scrollOffset = selectedIdx - availableRows + 1
	}

	for i := scrollOffset; i < len(items) && len(lines) < s.height; i++ {
		item := items[i]
		isSelected := (i == selectedIdx)

		if item.IsCategory {
			catName := string(item.Category)
			catStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(s.theme.CategoryFg)
			row := "▼ " + catName
			if isSelected {
				row = "► " + catName
				row = lipgloss.NewStyle().
					Background(s.theme.SelectedBg).
					Foreground(s.theme.SelectedFg).
					Bold(true).
					Render(PadOrTruncate(row, w))
			} else {
				row = catStyle.Render(PadOrTruncate(row, w))
			}
			lines = append(lines, row)
			continue
		}

		// Story row
		bullet := "  • "
		if isSelected {
			bullet = "  ● "
		}

		storyName := item.Story.Name
		badge := ""
		if item.Story.Level != "" {
			badge = " " + FormatLevelTag(item.Story.Level, s.theme)
		}

		rowText := bullet + storyName
		if isSelected {
			selStyle := lipgloss.NewStyle().
				Background(s.theme.SelectedBg).
				Foreground(s.theme.SelectedFg).
				Bold(true)
			lines = append(lines, PadOrTruncate(selStyle.Render(rowText+badge), w))
		} else {
			mutedStyle := lipgloss.NewStyle().Foreground(s.theme.Fg)
			lines = append(lines, PadOrTruncate(mutedStyle.Render(rowText)+badge, w))
		}
	}

	// Pad remaining vertical lines
	for len(lines) < s.height {
		lines = append(lines, strings.Repeat(" ", w))
	}

	return strings.Join(lines, "\n")
}

func countStories(items []NavItem) int {
	count := 0
	for _, it := range items {
		if !it.IsCategory && it.Story != nil {
			count++
		}
	}
	return count
}

package storybook

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// MockQuitMessage is the standard simulated exit banner for embedded examples.
const MockQuitMessage = "[Program Exited (tea.Quit simulated) — press 'r' or 'space' to restart]"

// RenderMockQuit returns the formatted simulated program exit notice.
func RenderMockQuit() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f7768e")).
		Background(lipgloss.Color("#24283b")).
		Padding(0, 1).
		MarginTop(1)
	return style.Render(MockQuitMessage)
}

// Theme encapsulates styling colors for the Storybook chrome and preview canvas.
type Theme struct {
	Name             string
	Bg               color.Color
	Fg               color.Color
	Muted            color.Color
	Border           color.Color
	ActiveBorder     color.Color
	SelectedBg       color.Color
	SelectedFg       color.Color
	HeaderBg         color.Color
	Accent           color.Color
	CategoryFg       color.Color
	StatusBg         color.Color
	StatusFg         color.Color
	CodeGutterBg     color.Color
	CodeGutterFg     color.Color
	CodeSelectBg     color.Color
	CodeSelectFg     color.Color
	LevelAtomFg      color.Color
	LevelMoleculeFg  color.Color
	LevelOrganismFg  color.Color
	LevelModalFg     color.Color
	LevelExampleFg   color.Color
}

// ThemeTokyoNight defines Tokyo Night theme palette.
var ThemeTokyoNight = Theme{
	Name:             "Tokyo Night",
	Bg:               lipgloss.Color("#1a1b26"),
	Fg:               lipgloss.Color("#c0caf5"),
	Muted:            lipgloss.Color("#565f89"),
	Border:           lipgloss.Color("#3b4261"),
	ActiveBorder:     lipgloss.Color("#7aa2f7"),
	SelectedBg:       lipgloss.Color("#283457"),
	SelectedFg:       lipgloss.Color("#ffffff"),
	HeaderBg:         lipgloss.Color("#24283b"),
	Accent:           lipgloss.Color("#7dcfff"),
	CategoryFg:       lipgloss.Color("#bb9af7"),
	StatusBg:         lipgloss.Color("#1f2335"),
	StatusFg:         lipgloss.Color("#9aa5ce"),
	CodeGutterBg:     lipgloss.Color("#1e2030"),
	CodeGutterFg:     lipgloss.Color("#565f89"),
	CodeSelectBg:     lipgloss.Color("#3d59a1"),
	CodeSelectFg:     lipgloss.Color("#ffffff"),
	LevelAtomFg:      lipgloss.Color("#73daca"),
	LevelMoleculeFg:  lipgloss.Color("#7aa2f7"),
	LevelOrganismFg:  lipgloss.Color("#bb9af7"),
	LevelModalFg:     lipgloss.Color("#f7768e"),
	LevelExampleFg:   lipgloss.Color("#e0af68"),
}

// ThemeSolarizedDark defines Solarized Dark palette.
var ThemeSolarizedDark = Theme{
	Name:             "Solarized Dark",
	Bg:               lipgloss.Color("#002b36"),
	Fg:               lipgloss.Color("#839496"),
	Muted:            lipgloss.Color("#586e75"),
	Border:           lipgloss.Color("#073642"),
	ActiveBorder:     lipgloss.Color("#268bd2"),
	SelectedBg:       lipgloss.Color("#073642"),
	SelectedFg:       lipgloss.Color("#eee8d5"),
	HeaderBg:         lipgloss.Color("#073642"),
	Accent:           lipgloss.Color("#2aa198"),
	CategoryFg:       lipgloss.Color("#b58900"),
	StatusBg:         lipgloss.Color("#073642"),
	StatusFg:         lipgloss.Color("#93a1a1"),
	CodeGutterBg:     lipgloss.Color("#073642"),
	CodeGutterFg:     lipgloss.Color("#586e75"),
	CodeSelectBg:     lipgloss.Color("#268bd2"),
	CodeSelectFg:     lipgloss.Color("#ffffff"),
	LevelAtomFg:      lipgloss.Color("#2aa198"),
	LevelMoleculeFg:  lipgloss.Color("#268bd2"),
	LevelOrganismFg:  lipgloss.Color("#6c71c4"),
	LevelModalFg:     lipgloss.Color("#dc322f"),
	LevelExampleFg:   lipgloss.Color("#b58900"),
}

// ThemeSolarizedLight defines Solarized Light palette.
var ThemeSolarizedLight = Theme{
	Name:             "Solarized Light",
	Bg:               lipgloss.Color("#fdf6e3"),
	Fg:               lipgloss.Color("#657b83"),
	Muted:            lipgloss.Color("#93a1a1"),
	Border:           lipgloss.Color("#eee8d5"),
	ActiveBorder:     lipgloss.Color("#268bd2"),
	SelectedBg:       lipgloss.Color("#eee8d5"),
	SelectedFg:       lipgloss.Color("#073642"),
	HeaderBg:         lipgloss.Color("#eee8d5"),
	Accent:           lipgloss.Color("#2aa198"),
	CategoryFg:       lipgloss.Color("#b58900"),
	StatusBg:         lipgloss.Color("#eee8d5"),
	StatusFg:         lipgloss.Color("#586e75"),
	CodeGutterBg:     lipgloss.Color("#eee8d5"),
	CodeGutterFg:     lipgloss.Color("#93a1a1"),
	CodeSelectBg:     lipgloss.Color("#93a1a1"),
	CodeSelectFg:     lipgloss.Color("#002b36"),
	LevelAtomFg:      lipgloss.Color("#2aa198"),
	LevelMoleculeFg:  lipgloss.Color("#268bd2"),
	LevelOrganismFg:  lipgloss.Color("#6c71c4"),
	LevelModalFg:     lipgloss.Color("#dc322f"),
	LevelExampleFg:   lipgloss.Color("#b58900"),
}

// ThemeTerminal defines high-contrast 16-color ANSI terminal palette.
var ThemeTerminal = Theme{
	Name:             "Terminal 16",
	Bg:               lipgloss.Color("0"),
	Fg:               lipgloss.Color("7"),
	Muted:            lipgloss.Color("8"),
	Border:           lipgloss.Color("8"),
	ActiveBorder:     lipgloss.Color("4"),
	SelectedBg:       lipgloss.Color("4"),
	SelectedFg:       lipgloss.Color("15"),
	HeaderBg:         lipgloss.Color("8"),
	Accent:           lipgloss.Color("6"),
	CategoryFg:       lipgloss.Color("5"),
	StatusBg:         lipgloss.Color("8"),
	StatusFg:         lipgloss.Color("7"),
	CodeGutterBg:     lipgloss.Color("0"),
	CodeGutterFg:     lipgloss.Color("8"),
	CodeSelectBg:     lipgloss.Color("4"),
	CodeSelectFg:     lipgloss.Color("15"),
	LevelAtomFg:      lipgloss.Color("2"),
	LevelMoleculeFg:  lipgloss.Color("4"),
	LevelOrganismFg:  lipgloss.Color("5"),
	LevelModalFg:     lipgloss.Color("1"),
	LevelExampleFg:   lipgloss.Color("3"),
}

// AllThemes lists the standard available color schemes.
var AllThemes = []Theme{
	ThemeTokyoNight,
	ThemeSolarizedDark,
	ThemeSolarizedLight,
	ThemeTerminal,
}

// LevelColor returns the badge color corresponding to an atomic level under current theme.
func (t Theme) LevelColor(level AtomicLevel) color.Color {
	switch level {
	case LevelAtom:
		return t.LevelAtomFg
	case LevelMolecule:
		return t.LevelMoleculeFg
	case LevelOrganism:
		return t.LevelOrganismFg
	case LevelModal:
		return t.LevelModalFg
	case LevelExample:
		return t.LevelExampleFg
	default:
		return t.Muted
	}
}

// TruncateString truncates a string with ellipsis if it exceeds max width.
func TruncateString(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-3]) + "..."
}

// PadOrTruncate pads or truncates a string to fit exact width.
func PadOrTruncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > width {
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// FormatLevelTag returns a styled level badge for the component.
func FormatLevelTag(level AtomicLevel, t Theme) string {
	if level == "" {
		return ""
	}
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(t.LevelColor(level))
	return style.Render(level.ShortTag())
}

// FormatPill returns a rounded pill badge.
func FormatPill(label string, fg, bg color.Color) string {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Bold(true).
		Padding(0, 1).
		Render(label)
}

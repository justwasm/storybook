package storybook

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

// RunOption configures storybook execution.
type RunOption func(*runConfig)

type runConfig struct {
	registry *Registry
	title    string
}

// WithRegistry specifies a custom registry instead of DefaultRegistry.
func WithRegistry(r *Registry) RunOption {
	return func(c *runConfig) {
		c.registry = r
	}
}

// WithTitle sets the storybook window title.
func WithTitle(title string) RunOption {
	return func(c *runConfig) {
		c.title = title
	}
}

// Run executes the interactive storybook program.
func Run(opts ...RunOption) error {
	cfg := &runConfig{
		registry: DefaultRegistry,
		title:    "❖ Terminal Component Storybook & Substrate",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" || arg == "help" {
			printUsage(cfg.title, cfg.registry)
			return nil
		}
	}

	model := NewModel(cfg.registry)
	if cfg.title != "" {
		model.SetTitle(cfg.title)
	}
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("storybook: error running interactive program: %w", err)
	}
	return nil
}

func printUsage(title string, r *Registry) {
	fmt.Println(title)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run . [flags]")
	fmt.Println()
	fmt.Println("Navigation & Controls:")
	fmt.Println("  ↑ / ↓, j / k      Navigate through component stories and categories")
	fmt.Println("  Enter / l         Focus playground / interact with selected component")
	fmt.Println("  Tab               Cycle component state variations or view tabs")
	fmt.Println("  1, 2, 3           Directly jump to Live Demo, Source Code, or Info view")
	fmt.Println("  /                 Filter component catalog by name or category")
	fmt.Println("  t                 Cycle visual color themes (Tokyo Night, Solarized, Terminal)")
	fmt.Println("  q / Ctrl+C        Quit storybook")
	fmt.Println()
	fmt.Println("Registered Categories:")
	for _, c := range r.Categories() {
		fmt.Printf("  • %s\n", c)
	}
}

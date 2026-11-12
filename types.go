package storybook

import (
	tea "charm.land/bubbletea/v2"
)

// Category groups stories by architectural or functional domain.
type Category string

const (
	CategoryBasics       Category = "1. Basics & Lifecycle"
	CategoryTerminal     Category = "2. Terminal & Screen"
	CategoryInput        Category = "3. Input & Events"
	CategoryForms        Category = "4. Forms & Text"
	CategoryLists        Category = "5. Lists & Layouts"
	CategoryProgress     Category = "6. Progress & Timing"
	CategoryApps         Category = "7. Graphics & Apps"
	CategoryFoundations  Category = "1. Foundations"
	CategoryComposer     Category = "2. Composer & Inputs"
	CategoryPopups       Category = "3. Autocomplete Popups"
	CategoryToolSteps    Category = "4. Tool Steps & Execution"
	CategoryPanels       Category = "5. Panels & Dialogs"
	CategoryStatus       Category = "6. Status & Overlays"
	CategoryWorkspace    Category = "7. Workspace & Multiplexer"
)

// AtomicLevel defines the Atomic Design hierarchy level for the component.
type AtomicLevel string

const (
	LevelAtom      AtomicLevel = "Atom"
	LevelMolecule  AtomicLevel = "Molecule"
	LevelOrganism  AtomicLevel = "Organism"
	LevelModal     AtomicLevel = "Modal"
	LevelExample   AtomicLevel = "Example"
	LevelComponent AtomicLevel = "Component"
)

// LevelTag returns a full uppercase badge for the atomic design level.
func (l AtomicLevel) LevelTag() string {
	switch l {
	case LevelAtom:
		return "[ATOM]"
	case LevelMolecule:
		return "[MOLECULE]"
	case LevelOrganism:
		return "[ORGANISM]"
	case LevelModal:
		return "[MODAL]"
	case LevelExample:
		return "[EXAMPLE]"
	default:
		return "[COMPONENT]"
	}
}

// ShortTag returns a compact 3-character badge for sidebar listings.
func (l AtomicLevel) ShortTag() string {
	switch l {
	case LevelAtom:
		return "[A]"
	case LevelMolecule:
		return "[M]"
	case LevelOrganism:
		return "[O]"
	case LevelModal:
		return "[D]"
	case LevelExample:
		return "[E]"
	default:
		return "[C]"
	}
}

// State represents a preset configuration or visual variation of a component.
type State struct {
	ID          string
	Name        string
	Description string
}

// Instance is an active, stateful or rendered instance of a story.
type Instance interface {
	// View renders the component's visual representation.
	View() string
	// Update delivers a message to the component.
	Update(msg tea.Msg) (Instance, tea.Cmd)
}

// Resizable is an optional interface for instances that adapt to canvas dimensions.
type Resizable interface {
	SetSize(width, height int)
}

// Stateful is an optional interface for instances that support state variations.
type Stateful interface {
	SetState(stateIndex int)
}

// InteractiveHandler is an optional interface for instances handling raw keys in playground mode.
type InteractiveHandler interface {
	HandleKey(key tea.KeyMsg) (tea.Cmd, bool)
}

// MouseHandler is an optional interface for instances handling mouse events.
type MouseHandler interface {
	HandleMouse(ev tea.MouseMsg) (tea.Cmd, bool)
}

// Restarter is an optional interface for instances that can reset their state.
type Restarter interface {
	Restart() tea.Cmd
}

// Initializer is an optional interface for instances with an initial command.
type Initializer interface {
	Init() tea.Cmd
}

// Story defines a single component or example exhibit in the Storybook catalogue.
type Story struct {
	ID          string                  // Unique identifier, e.g. "simple"
	Name        string                  // Display name
	Level       AtomicLevel             // Atomic Design level
	Category    Category                // Parent category
	Description string                  // Component purpose and role
	UpstreamRef string                  // Upstream symbol or file path
	SourceCode  string                  // Source code content for the source tab
	Doc         string                  // Documentation / architecture explanation
	KeyFeatures []string                // Bullet points or key capabilities
	States      []State                 // Preset states
	Factory     func() Instance         // Instantiation factory
}

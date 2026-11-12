package storybook

import (
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
)

// NavItem represents a row in the sidebar list (either a category header or a story).
type NavItem struct {
	IsCategory bool
	Category   Category
	Story      *Story
}

// Registry holds all registered stories and category arrangements.
type Registry struct {
	mu           sync.RWMutex
	stories      []Story
	byID         map[string]Story
	byCategory   map[Category][]Story
	categories   []Category
	collapsedCat map[Category]bool
}

// NewRegistry constructs an empty story catalog registry.
func NewRegistry() *Registry {
	return &Registry{
		byID:         make(map[string]Story),
		byCategory:   make(map[Category][]Story),
		collapsedCat: make(map[Category]bool),
	}
}

// DefaultRegistry is the global singleton registry populated by plugins via init().
var DefaultRegistry = NewRegistry()

// Register adds a Story to the global default registry.
func Register(s Story) {
	DefaultRegistry.Register(s)
}

// RegisterCategory ensures a category exists and specifies its order in the global default registry.
func RegisterCategory(cat Category) {
	DefaultRegistry.RegisterCategory(cat)
}

// MustRegister adds a Story to the global default registry or panics on duplicate ID.
func MustRegister(s Story) {
	DefaultRegistry.MustRegister(s)
}

// Register adds a Story to this registry.
func (r *Registry) Register(s Story) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byID[s.ID]; exists {
		// Replace existing story while preserving order
		r.byID[s.ID] = s
		for i, existing := range r.stories {
			if existing.ID == s.ID {
				r.stories[i] = s
				break
			}
		}
		// Update category slice
		catStories := r.byCategory[s.Category]
		for i, existing := range catStories {
			if existing.ID == s.ID {
				catStories[i] = s
				r.byCategory[s.Category] = catStories
				return
			}
		}
		r.byCategory[s.Category] = append(r.byCategory[s.Category], s)
		return
	}

	r.stories = append(r.stories, s)
	r.byID[s.ID] = s
	r.byCategory[s.Category] = append(r.byCategory[s.Category], s)

	// Add category to category list if not already present
	found := false
	for _, c := range r.categories {
		if c == s.Category {
			found = true
			break
		}
	}
	if !found {
		r.categories = append(r.categories, s.Category)
	}
}

// MustRegister adds a story or panics if already registered.
func (r *Registry) MustRegister(s Story) {
	r.mu.Lock()
	if _, exists := r.byID[s.ID]; exists {
		r.mu.Unlock()
		panic("storybook: story already registered: " + s.ID)
	}
	r.mu.Unlock()
	r.Register(s)
}

// RegisterCategory ensures a category appears in the catalog in a specific order.
func (r *Registry) RegisterCategory(cat Category) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.categories {
		if c == cat {
			return
		}
	}
	r.categories = append(r.categories, cat)
}

// Stories returns all registered stories in order.
func (r *Registry) Stories() []Story {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]Story, len(r.stories))
	copy(res, r.stories)
	return res
}

// StoriesByCategory returns the stories belonging to a category.
func (r *Registry) StoriesByCategory(cat Category) []Story {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byCategory[cat]
}

// Categories returns all registered categories in order.
func (r *Registry) Categories() []Category {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]Category, len(r.categories))
	copy(res, r.categories)
	return res
}

// Find looks up a story by ID.
func (r *Registry) Find(id string) (Story, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byID[id]
	return s, ok
}

// ToggleCategory toggles whether a category is collapsed in the sidebar.
func (r *Registry) ToggleCategory(cat Category) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.collapsedCat[cat] = !r.collapsedCat[cat]
}

// IsCategoryCollapsed reports whether a category is collapsed.
func (r *Registry) IsCategoryCollapsed(cat Category) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collapsedCat[cat]
}

// Reset clears all registered stories and categories. Primarily used in unit tests.
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stories = nil
	r.byID = make(map[string]Story)
	r.byCategory = make(map[Category][]Story)
	r.categories = nil
	r.collapsedCat = make(map[Category]bool)
}

// ResetDefault clears the global default registry.
func ResetDefault() {
	DefaultRegistry.Reset()
}

// VisibleNavItems returns the flattened list of navigable rows based on collapse states and search query.
func (r *Registry) VisibleNavItems(filterQuery string) []NavItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []NavItem
	filterLower := strings.ToLower(strings.TrimSpace(filterQuery))

	for _, cat := range r.categories {
		stories := r.byCategory[cat]
		if len(stories) == 0 {
			continue
		}

		var matchingStories []*Story
		for i := range stories {
			s := &stories[i]
			if filterLower == "" ||
				strings.Contains(strings.ToLower(s.Name), filterLower) ||
				strings.Contains(strings.ToLower(s.ID), filterLower) ||
				strings.Contains(strings.ToLower(s.UpstreamRef), filterLower) ||
				strings.Contains(strings.ToLower(s.Description), filterLower) {
				matchingStories = append(matchingStories, s)
			}
		}

		if filterLower != "" && len(matchingStories) == 0 {
			continue
		}

		items = append(items, NavItem{
			IsCategory: true,
			Category:   cat,
		})

		// If collapsed (and not searching), don't show stories
		if filterLower == "" && r.collapsedCat[cat] {
			continue
		}

		for _, s := range matchingStories {
			items = append(items, NavItem{
				IsCategory: false,
				Category:   cat,
				Story:      s,
			})
		}
	}
	return items
}

// StoryOption configures optional metadata on a Story.
type StoryOption func(*Story)

// WithDescription sets the component purpose and role.
func WithDescription(desc string) StoryOption {
	return func(s *Story) { s.Description = desc }
}

// WithLevel sets the Atomic Design hierarchy level of a story.
func WithLevel(lvl AtomicLevel) StoryOption {
	return func(s *Story) { s.Level = lvl }
}

// WithUpstream sets the upstream source reference or file path.
func WithUpstream(ref string) StoryOption {
	return func(s *Story) { s.UpstreamRef = ref }
}

// WithSourceCode sets the source code snippet for the source tab.
func WithSourceCode(src string) StoryOption {
	return func(s *Story) { s.SourceCode = src }
}

// WithDoc sets the documentation or architectural notes.
func WithDoc(doc string) StoryOption {
	return func(s *Story) { s.Doc = doc }
}

// WithKeyFeatures sets the key bullet points for the architecture tab.
func WithKeyFeatures(features ...string) StoryOption {
	return func(s *Story) { s.KeyFeatures = features }
}

// WithStates sets the state variations for dual-axis navigation (1..9 keys).
func WithStates(states ...State) StoryOption {
	return func(s *Story) { s.States = states }
}

// RegisterModel registers a standard Bubble Tea v2 model factory with automatic
// lifecycle adaptation, canvas resizing, mouse coordinate translation, and mock-quit isolation.
func RegisterModel(id, name string, cat Category, factory func() tea.Model, opts ...StoryOption) Story {
	story := Story{
		ID:       id,
		Name:     name,
		Category: cat,
		Level:    LevelComponent,
		Factory: func() Instance {
			m := factory()
			return WrapModel(m, WithResetFactory(factory))
		},
	}
	for _, opt := range opts {
		opt(&story)
	}
	Register(story)
	return story
}

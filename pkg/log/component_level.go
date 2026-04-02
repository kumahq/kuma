package log

import (
	"maps"
	"strings"
	"sync"
)

var globalRegistry = &ComponentLevelRegistry{
	overrides: make(map[string]LogLevel),
}

// GlobalComponentLevelRegistry returns the singleton registry used to manage
// per-component log level overrides.
func GlobalComponentLevelRegistry() *ComponentLevelRegistry {
	return globalRegistry
}

// NewComponentLevelRegistry creates a new empty registry, mainly for testing.
func NewComponentLevelRegistry() *ComponentLevelRegistry {
	return &ComponentLevelRegistry{
		overrides: make(map[string]LogLevel),
	}
}

// ComponentLevelRegistry holds per-component log level overrides.
// Components are identified by dot-separated names matching the hierarchy
// built by logr.Logger.WithName() calls (e.g. "xds.server").
type ComponentLevelRegistry struct {
	mu        sync.RWMutex
	overrides map[string]LogLevel
}

// SetLevel sets a log level override for the given component.
func (r *ComponentLevelRegistry) SetLevel(component string, level LogLevel) {
	r.mu.Lock()
	r.overrides[component] = level
	r.mu.Unlock()
}

// ResetLevel removes the log level override for the given component.
func (r *ComponentLevelRegistry) ResetLevel(component string) {
	r.mu.Lock()
	delete(r.overrides, component)
	r.mu.Unlock()
}

// ResetAll removes all component log level overrides.
func (r *ComponentLevelRegistry) ResetAll() {
	r.mu.Lock()
	r.overrides = make(map[string]LogLevel)
	r.mu.Unlock()
}

// GetEffectiveLevel returns the effective log level for a component.
// It checks for an exact match first, then walks up the hierarchy
// (e.g. "xds.server" → "xds"). Returns the level and true if an
// override was found, or zero value and false otherwise.
func (r *ComponentLevelRegistry) GetEffectiveLevel(component string) (LogLevel, bool) {
	if component == "" {
		return 0, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Walk from most specific to least specific
	name := component
	for {
		if level, ok := r.overrides[name]; ok {
			return level, true
		}
		idx := strings.LastIndex(name, ".")
		if idx < 0 {
			break
		}
		name = name[:idx]
	}

	return 0, false
}

// ListOverrides returns a snapshot of all current component level overrides.
func (r *ComponentLevelRegistry) ListOverrides() map[string]LogLevel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]LogLevel, len(r.overrides))
	maps.Copy(result, r.overrides)
	return result
}

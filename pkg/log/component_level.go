package log

import (
	"fmt"
	"maps"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

const maxComponentNameLen = 256

// MaxOverrides is the maximum number of per-component level overrides
// allowed in a single registry. This prevents unbounded memory growth
// from misconfigured or malicious API clients.
const MaxOverrides = 200

var validComponentName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

var globalRegistry = NewComponentLevelRegistry()

// GlobalComponentLevelRegistry returns the singleton registry used to manage
// per-component log level overrides.
func GlobalComponentLevelRegistry() *ComponentLevelRegistry {
	return globalRegistry
}

// NewComponentLevelRegistry creates a new empty registry, mainly for testing.
func NewComponentLevelRegistry() *ComponentLevelRegistry {
	r := &ComponentLevelRegistry{}
	m := make(map[string]LogLevel)
	r.snapshot.Store(&m)
	return r
}

// ComponentLevelRegistry holds per-component log level overrides.
// Components are identified by dot-separated names matching the hierarchy
// built by logr.Logger.WithName() calls (e.g. "xds.server").
//
// Reads are lock-free via an atomic pointer to an immutable snapshot.
// Writes acquire a mutex and swap in a new copy (copy-on-write).
type ComponentLevelRegistry struct {
	mu       sync.Mutex
	snapshot atomic.Pointer[map[string]LogLevel]
}

// ValidateComponentName checks that the component name is well-formed:
// non-empty, at most 256 characters, alphanumeric with dots/dashes/underscores.
func ValidateComponentName(name string) error {
	if name == "" {
		return fmt.Errorf("component name must not be empty")
	}
	if len(name) > maxComponentNameLen {
		return fmt.Errorf("component name exceeds %d characters", maxComponentNameLen)
	}
	if !validComponentName.MatchString(name) {
		return fmt.Errorf("component name %q contains invalid characters", name)
	}
	return nil
}

// SetLevel sets a log level override for the given component.
// Returns an error if the maximum number of overrides would be exceeded.
func (r *ComponentLevelRegistry) SetLevel(component string, level LogLevel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current := *r.snapshot.Load()
	if _, exists := current[component]; !exists && len(current) >= MaxOverrides {
		return fmt.Errorf("maximum number of component overrides (%d) reached", MaxOverrides)
	}
	next := make(map[string]LogLevel, len(current)+1)
	maps.Copy(next, current)
	next[component] = level
	r.snapshot.Store(&next)
	return nil
}

// ResetLevel removes the log level override for the given component.
func (r *ComponentLevelRegistry) ResetLevel(component string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	current := *r.snapshot.Load()
	next := make(map[string]LogLevel, len(current))
	maps.Copy(next, current)
	delete(next, component)
	r.snapshot.Store(&next)
}

// ResetAll removes all component log level overrides and returns
// the overrides that were active before the reset.
func (r *ComponentLevelRegistry) ResetAll() map[string]LogLevel {
	r.mu.Lock()
	defer r.mu.Unlock()
	prev := *r.snapshot.Load()
	m := make(map[string]LogLevel)
	r.snapshot.Store(&m)
	result := make(map[string]LogLevel, len(prev))
	maps.Copy(result, prev)
	return result
}

// GetEffectiveLevelForNames returns the effective log level given a
// pre-computed name hierarchy (most-specific first). This avoids string
// splitting on the hot path when called from componentAwareSink.Enabled.
func (r *ComponentLevelRegistry) GetEffectiveLevelForNames(names []string) (LogLevel, bool) {
	m := r.snapshot.Load()
	for _, name := range names {
		if level, ok := (*m)[name]; ok {
			return level, true
		}
	}
	return 0, false
}

// ListOverrides returns a snapshot of all current component level overrides.
func (r *ComponentLevelRegistry) ListOverrides() map[string]LogLevel {
	m := r.snapshot.Load()
	result := make(map[string]LogLevel, len(*m))
	maps.Copy(result, *m)
	return result
}

// SplitHierarchy returns the name and all its ancestors, most-specific first.
// e.g. "xds.server" → ["xds.server", "xds"]
func SplitHierarchy(name string) []string {
	var names []string
	for {
		names = append(names, name)
		idx := strings.LastIndex(name, ".")
		if idx < 0 {
			break
		}
		name = name[:idx]
	}
	return names
}

package log

import (
	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// componentAwareSink wraps a logr.LogSink to add per-component log level
// control. It intercepts WithName calls to track the hierarchical component
// name (dot-separated) and checks the ComponentLevelRegistry on each
// Enabled call to apply per-component level overrides.
//
// The inner sink must be created at maximum verbosity (-10) so that Info()
// never filters — all filtering is done by this sink's Enabled() method.
type componentAwareSink struct {
	inner     logr.LogSink
	name      string
	names     []string // pre-computed hierarchy, most-specific first
	registry  *ComponentLevelRegistry
	baseLevel *zap.AtomicLevel // base level for fallback when no override
}

// NewComponentAwareSink wraps an existing LogSink with per-component level
// awareness. The inner sink should be created at max verbosity. baseLevel
// is used as fallback when no per-component override is set.
func NewComponentAwareSink(inner logr.LogSink, registry *ComponentLevelRegistry, baseLevel *zap.AtomicLevel) logr.LogSink {
	return &componentAwareSink{
		inner:     inner,
		registry:  registry,
		baseLevel: baseLevel,
	}
}

func (s *componentAwareSink) Init(info logr.RuntimeInfo) {
	s.inner.Init(info)
}

func (s *componentAwareSink) Enabled(level int) bool {
	if len(s.names) > 0 {
		if override, ok := s.registry.GetEffectiveLevelForNames(s.names); ok {
			switch override {
			case OffLevel:
				return false
			case DebugLevel:
				return level <= -maxVerbosity
			case InfoLevel:
				return level <= 0
			default:
				// unknown override — fall back to base level
			}
		}
	}
	// No override — check base level
	// logr V-level N maps to zap level (InfoLevel - N)
	return s.baseLevel.Enabled(zapcore.InfoLevel - zapcore.Level(level))
}

func (s *componentAwareSink) Info(level int, msg string, keysAndValues ...any) {
	s.inner.Info(level, msg, keysAndValues...)
}

func (s *componentAwareSink) Error(err error, msg string, keysAndValues ...any) {
	s.inner.Error(err, msg, keysAndValues...)
}

func (s *componentAwareSink) WithValues(keysAndValues ...any) logr.LogSink {
	return &componentAwareSink{
		inner:     s.inner.WithValues(keysAndValues...),
		name:      s.name,
		names:     s.names,
		registry:  s.registry,
		baseLevel: s.baseLevel,
	}
}

func (s *componentAwareSink) WithName(name string) logr.LogSink {
	fullName := name
	if s.name != "" {
		fullName = s.name + "." + name
	}
	return &componentAwareSink{
		inner:     s.inner.WithName(name),
		name:      fullName,
		names:     SplitHierarchy(fullName),
		registry:  s.registry,
		baseLevel: s.baseLevel,
	}
}

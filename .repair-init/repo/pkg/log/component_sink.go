package log

import (
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logLinesCounter is nil until SetupMetrics is called.
var logLinesCounter atomic.Pointer[prometheus.CounterVec]

// ResetMetrics clears the global log-lines counter. Intended for test teardown
// to prevent counter state from leaking across packages.
func ResetMetrics() {
	logLinesCounter.Store(nil)
}

// SetupMetrics registers the kuma_cp_log_lines_total counter with the given registerer.
func SetupMetrics(reg prometheus.Registerer) (*prometheus.CounterVec, error) {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kuma_cp_log_lines_total",
		Help: "Total number of log lines emitted by the control plane, by logger name and level.",
	}, []string{"logger", "level"})
	if err := reg.Register(cv); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			cv = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			return nil, err
		}
	}
	logLinesCounter.Store(cv)
	return cv, nil
}

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
	cntInfo   prometheus.Counter
	cntDebug  prometheus.Counter
	cntError  prometheus.Counter
}

// NewComponentAwareSink wraps an existing LogSink with per-component level
// awareness. The inner sink should be created at max verbosity. baseLevel
// is used as fallback when no per-component override is set.
func NewComponentAwareSink(inner logr.LogSink, registry *ComponentLevelRegistry, baseLevel *zap.AtomicLevel) logr.LogSink {
	cntInfo, cntDebug, cntError := resolveCounters("root")
	return &componentAwareSink{
		inner:     inner,
		registry:  registry,
		baseLevel: baseLevel,
		cntInfo:   cntInfo,
		cntDebug:  cntDebug,
		cntError:  cntError,
	}
}

// resolveCounters pre-resolves the info/debug/error counters for loggerName
// from the current global CounterVec snapshot. Returns nils if metrics are not
// set up yet; callers must guard with nil checks.
func resolveCounters(loggerName string) (prometheus.Counter, prometheus.Counter, prometheus.Counter) {
	cv := logLinesCounter.Load()
	if cv == nil {
		return nil, nil, nil
	}
	return cv.WithLabelValues(loggerName, "info"),
		cv.WithLabelValues(loggerName, "debug"),
		cv.WithLabelValues(loggerName, "error")
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
	// level == 0 is Info (logr .Info()); level > 0 is a V-level (debug).
	// Negative values are not produced by zapr but treated as debug defensively.
	if level == 0 {
		if s.cntInfo != nil {
			s.cntInfo.Inc()
		}
	} else {
		if s.cntDebug != nil {
			s.cntDebug.Inc()
		}
	}
	s.inner.Info(level, msg, keysAndValues...)
}

func (s *componentAwareSink) Error(err error, msg string, keysAndValues ...any) {
	if s.cntError != nil {
		s.cntError.Inc()
	}
	s.inner.Error(err, msg, keysAndValues...)
}

func (s *componentAwareSink) WithValues(keysAndValues ...any) logr.LogSink {
	return &componentAwareSink{
		inner:     s.inner.WithValues(keysAndValues...),
		name:      s.name,
		names:     s.names,
		registry:  s.registry,
		baseLevel: s.baseLevel,
		cntInfo:   s.cntInfo,
		cntDebug:  s.cntDebug,
		cntError:  s.cntError,
	}
}

func (s *componentAwareSink) WithName(name string) logr.LogSink {
	fullName := name
	if s.name != "" {
		fullName = s.name + "." + name
	}
	cntInfo, cntDebug, cntError := resolveCounters(fullName)
	return &componentAwareSink{
		inner:     s.inner.WithName(name),
		name:      fullName,
		names:     SplitHierarchy(fullName),
		registry:  s.registry,
		baseLevel: s.baseLevel,
		cntInfo:   cntInfo,
		cntDebug:  cntDebug,
		cntError:  cntError,
	}
}

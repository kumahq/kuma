package component

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type logEntry struct {
	msg  string
	tags map[string]any
}

type recordingLogSinkRoot struct {
	mu      sync.Mutex
	entries []logEntry
}

func (r *recordingLogSinkRoot) snapshot() []logEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]logEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

func (r *recordingLogSinkRoot) entriesByMsg(msg string) []logEntry {
	var out []logEntry
	for _, e := range r.snapshot() {
		if e.msg == msg {
			out = append(out, e)
		}
	}
	return out
}

type recordingLogSink struct {
	tags []any
	root *recordingLogSinkRoot
}

func (s *recordingLogSink) Init(logr.RuntimeInfo) {}
func (s *recordingLogSink) Enabled(int) bool      { return true }

func (s *recordingLogSink) record(msg string, extra []any) {
	tags := map[string]any{}
	all := append(append([]any(nil), s.tags...), extra...)
	for i := 0; i+1 < len(all); i += 2 {
		k, ok := all[i].(string)
		if !ok {
			continue
		}
		tags[k] = all[i+1]
	}
	s.root.mu.Lock()
	defer s.root.mu.Unlock()
	s.root.entries = append(s.root.entries, logEntry{msg: msg, tags: tags})
}

func (s *recordingLogSink) Info(_ int, msg string, vals ...any) {
	s.record(msg, vals)
}

func (s *recordingLogSink) Error(err error, msg string, vals ...any) {
	extra := append([]any{"error", err}, vals...)
	s.record(msg, extra)
}

func (s *recordingLogSink) WithValues(vals ...any) logr.LogSink {
	return &recordingLogSink{tags: append(append([]any(nil), s.tags...), vals...), root: s.root}
}

func (s *recordingLogSink) WithName(string) logr.LogSink {
	return &recordingLogSink{tags: append([]any(nil), s.tags...), root: s.root}
}

func newRecordingLogger() (logr.Logger, *recordingLogSinkRoot) {
	root := &recordingLogSinkRoot{}
	return logr.New(&recordingLogSink{root: root}), root
}

// scriptedComponent runs the next scripted behavior per Start call. Once the
// script is exhausted it blocks until stop closes.
type scriptedComponent struct {
	mu     sync.Mutex
	script []func() error
	calls  atomic.Int64
}

func (c *scriptedComponent) Start(stop <-chan struct{}) error {
	c.calls.Add(1)
	c.mu.Lock()
	if len(c.script) == 0 {
		c.mu.Unlock()
		<-stop
		return nil
	}
	f := c.script[0]
	c.script = c.script[1:]
	c.mu.Unlock()
	return f()
}

func (c *scriptedComponent) NeedLeaderElection() bool { return false }

// scriptOfErrors lifts a list of errors into a script that returns one error per Start call.
func scriptOfErrors(errs ...error) []func() error {
	out := make([]func() error, len(errs))
	for i, e := range errs {
		out[i] = func() error { return e }
	}
	return out
}

var _ = Describe("resilientComponent", func() {
	const tinyBase = 1 * time.Millisecond
	const tinyMax = 5 * time.Millisecond

	It("logs an Info on start and an Info on stop", func() {
		log, root := newRecordingLogger()
		stop := make(chan struct{})
		c := &scriptedComponent{}
		r := NewResilientComponent(log, c, tinyBase, tinyMax)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load).Should(BeNumerically(">=", int64(1)))
		close(stop)
		Eventually(done).Should(BeClosed())

		Expect(root.entriesByMsg("starting resilient component")).To(HaveLen(1))
		Expect(root.entriesByMsg("done")).To(HaveLen(1))
	})

	It("logs scheduling restart with attempt, sinceLastSuccess, nextBackoff and lastError", func() {
		log, root := newRecordingLogger()
		stop := make(chan struct{})

		c := &scriptedComponent{script: scriptOfErrors(
			errors.New("boom-1"),
			errors.New("boom-2"),
			errors.New("boom-3"),
		)}
		r := NewResilientComponent(log, c, tinyBase, tinyMax)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load, "2s", "5ms").Should(BeNumerically(">=", int64(4)))
		close(stop)
		Eventually(done).Should(BeClosed())

		scheduled := root.entriesByMsg("scheduling component restart")
		Expect(len(scheduled)).To(BeNumerically(">=", 3))

		Expect(scheduled[0].tags).To(HaveKeyWithValue("attempt", uint64(1)))
		Expect(scheduled[0].tags).To(HaveKey("generationID"))
		Expect(scheduled[0].tags).To(HaveKey("sinceLastSuccess"))
		Expect(scheduled[0].tags).To(HaveKey("nextBackoff"))
		Expect(scheduled[0].tags).To(HaveKeyWithValue("lastError", "boom-1"))

		Expect(scheduled[1].tags).To(HaveKeyWithValue("attempt", uint64(2)))
		Expect(scheduled[1].tags).To(HaveKeyWithValue("lastError", "boom-2"))
	})

	It("logs the component-terminated Error per failure", func() {
		log, root := newRecordingLogger()
		stop := make(chan struct{})

		c := &scriptedComponent{script: scriptOfErrors(
			errors.New("kaboom-1"),
			errors.New("kaboom-2"),
		)}
		r := NewResilientComponent(log, c, tinyBase, tinyMax)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load, "2s", "5ms").Should(BeNumerically(">=", int64(3)))
		close(stop)
		Eventually(done).Should(BeClosed())

		terminated := root.entriesByMsg("component terminated with an error")
		Expect(len(terminated)).To(BeNumerically(">=", 2))
		Expect(terminated[0].tags).To(HaveKey("generationID"))
		Expect(terminated[0].tags).To(HaveKey("error"))
	})

	It("samples the restart line after the first burst and reports suppressed", func() {
		log, root := newRecordingLogger()
		stop := make(chan struct{})

		const failures = resilientLogSampleAfterAttempt + 7
		errs := make([]error, failures)
		for i := range errs {
			errs[i] = fmt.Errorf("err-%d", i)
		}
		c := &scriptedComponent{script: scriptOfErrors(errs...)}
		r := NewResilientComponent(log, c, 100*time.Microsecond, 1*time.Millisecond)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load, "2s", "5ms").Should(BeNumerically(">=", int64(failures)))
		close(stop)
		Eventually(done).Should(BeClosed())

		scheduled := root.entriesByMsg("scheduling component restart")
		Expect(len(scheduled)).To(BeNumerically(">=", resilientLogSampleAfterAttempt))
		Expect(len(scheduled)).To(BeNumerically("<", failures))

		for i := 0; i < resilientLogSampleAfterAttempt && i < len(scheduled); i++ {
			Expect(scheduled[i].tags).ToNot(HaveKey("suppressed"))
		}
	})

	It("reports sinceLastSuccess relative to the most recent stable-run termination", func() {
		log, root := newRecordingLogger()
		stop := make(chan struct{})

		const stableRun = 30 * time.Millisecond
		const tinyMax = 5 * time.Millisecond
		c := &scriptedComponent{script: []func() error{
			func() error { time.Sleep(stableRun); return errors.New("after-stable") },
			func() error { return errors.New("quick") },
		}}
		r := NewResilientComponent(log, c, 1*time.Millisecond, tinyMax)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load, "2s", "5ms").Should(BeNumerically(">=", int64(2)))
		close(stop)
		Eventually(done).Should(BeClosed())

		scheduled := root.entriesByMsg("scheduling component restart")
		Expect(scheduled).ToNot(BeEmpty())
		sinceLastSuccess, ok := scheduled[0].tags["sinceLastSuccess"].(time.Duration)
		Expect(ok).To(BeTrue())
		Expect(sinceLastSuccess).To(BeNumerically("<", stableRun/2))
	})

	It("always-logs the first burst of a new outage after a stable run", func() {
		log, root := newRecordingLogger()
		stop := make(chan struct{})

		const stableRun = 30 * time.Millisecond
		const tinyMax = 1 * time.Millisecond

		// Pre-burst pushes the sampler past the always-log threshold and
		// accumulates suppressed entries inside the same 10s window.
		preBurst := make([]error, resilientLogSampleAfterAttempt*4)
		for i := range preBurst {
			preBurst[i] = fmt.Errorf("pre-%d", i)
		}
		script := scriptOfErrors(preBurst...)
		// Stable run >backoffMaxTime ends one outage and starts a new one.
		script = append(script, func() error {
			time.Sleep(stableRun)
			return errors.New("after-stable")
		})
		// Post-burst: the new outage's first 5 attempts must all log
		// (post-stable plus 4 quick failures = 5 attempts), neither
		// suppressed by leftover nextSampledAt nor carrying suppressed counts.
		postBurst := make([]error, resilientLogSampleAfterAttempt-1)
		for i := range postBurst {
			postBurst[i] = fmt.Errorf("post-%d", i)
		}
		script = append(script, scriptOfErrors(postBurst...)...)

		c := &scriptedComponent{script: script}
		r := NewResilientComponent(log, c, 100*time.Microsecond, tinyMax)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load, "5s", "5ms").Should(BeNumerically(">=", int64(len(script))))
		close(stop)
		Eventually(done).Should(BeClosed())

		scheduled := root.entriesByMsg("scheduling component restart")
		start := slices.IndexFunc(scheduled, func(e logEntry) bool {
			return e.tags["lastError"] == "after-stable"
		})
		Expect(start).To(BeNumerically(">=", 0))

		// The new outage's first resilientLogSampleAfterAttempt entries must
		// all be present with outage-local attempt counters 1..N and no
		// suppressed counts from the prior outage.
		Expect(len(scheduled)).To(BeNumerically(">=", start+resilientLogSampleAfterAttempt))
		for i := range resilientLogSampleAfterAttempt {
			entry := scheduled[start+i]
			Expect(entry.tags).To(HaveKeyWithValue("attempt", uint64(i+1)))
			Expect(entry.tags).ToNot(HaveKey("suppressed"))
		}
	})

	It("skips the restart log on clean exit", func() {
		log, root := newRecordingLogger()
		stop := make(chan struct{})

		c := &scriptedComponent{script: []func() error{
			func() error { return errors.New("transient") },
			func() error { return nil },
			func() error { return nil },
		}}
		r := NewResilientComponent(log, c, 1*time.Millisecond, 5*time.Millisecond)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load, "2s", "5ms").Should(BeNumerically(">=", int64(3)))
		close(stop)
		Eventually(done).Should(BeClosed())

		scheduled := root.entriesByMsg("scheduling component restart")
		// Only the transient-failure restart should emit; the two clean
		// exits restart silently to avoid "restart with no error" noise.
		Expect(scheduled).To(HaveLen(1))
		Expect(scheduled[0].tags).To(HaveKeyWithValue("lastError", "transient"))
	})

	It("recovers from a panic in the wrapped component", func() {
		log, root := newRecordingLogger()
		stop := make(chan struct{})

		c := &scriptedComponent{script: []func() error{
			func() error { panic(errors.New("synthetic-panic")) },
			func() error { panic("non-error-panic") },
			func() error { return errors.New("after-panic") },
		}}
		r := NewResilientComponent(log, c, 1*time.Millisecond, 5*time.Millisecond)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load, "2s", "5ms").Should(BeNumerically(">=", int64(3)))
		close(stop)
		Eventually(done).Should(BeClosed())

		terminated := root.entriesByMsg("component terminated with an error")
		Expect(len(terminated)).To(BeNumerically(">=", 3))
		errStr := func(e logEntry) string {
			err, _ := e.tags["error"].(error)
			if err == nil {
				return ""
			}
			return err.Error()
		}
		Expect(errStr(terminated[0])).To(ContainSubstring("synthetic-panic"))
		Expect(errStr(terminated[1])).To(ContainSubstring("non-error-panic"))
		Expect(errStr(terminated[2])).To(ContainSubstring("after-panic"))
	})

	It("returns promptly when stop fires during the backoff sleep", func() {
		log, _ := newRecordingLogger()
		stop := make(chan struct{})

		// Backoff window large enough to dominate any prompt-shutdown budget.
		const backoff = 2 * time.Second
		c := &scriptedComponent{script: scriptOfErrors(errors.New("trigger-backoff"))}
		r := NewResilientComponent(log, c, backoff, backoff)

		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			Expect(r.Start(stop)).To(Succeed())
		}()
		Eventually(c.calls.Load, "2s", "5ms").Should(BeNumerically(">=", int64(1)))
		// The component has returned its error; the wrapper is now in the
		// backoff sleep. Closing stop must unblock it well before backoff fires.
		closedAt := time.Now()
		close(stop)
		Eventually(done, "500ms").Should(BeClosed())
		Expect(time.Since(closedAt)).To(BeNumerically("<", backoff))
	})
})

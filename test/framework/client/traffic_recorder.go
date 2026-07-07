package client

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

type TrafficRecorder struct {
	name  string
	limit int

	mu       sync.Mutex
	firstErr error
	reported bool
	records  []TrafficRecord
}

type TrafficRecord struct {
	Timestamp time.Time      `json:"timestamp"`
	Label     string         `json:"label"`
	Success   bool           `json:"success"`
	Error     string         `json:"error,omitempty"`
	Instances map[string]int `json:"instances,omitempty"`
}

func NewTrafficRecorder(name string, limit int) *TrafficRecorder {
	if limit <= 0 {
		limit = 50
	}
	return &TrafficRecorder{
		name:  name,
		limit: limit,
	}
}

func (r *TrafficRecorder) RecordSuccess(label string) {
	r.record(TrafficRecord{
		Timestamp: time.Now().UTC(),
		Label:     label,
		Success:   true,
	})
}

func (r *TrafficRecorder) RecordInstances(label string, instances map[string]int, err error) {
	if err != nil {
		r.RecordError(label, err)
		return
	}
	r.record(TrafficRecord{
		Timestamp: time.Now().UTC(),
		Label:     label,
		Success:   true,
		Instances: instances,
	})
}

func (r *TrafficRecorder) RecordError(label string, err error) {
	if err == nil {
		r.RecordSuccess(label)
		return
	}
	r.record(TrafficRecord{
		Timestamp: time.Now().UTC(),
		Label:     label,
		Success:   false,
		Error:     err.Error(),
	})
}

func (r *TrafficRecorder) FirstError() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.firstErr == nil {
		return nil
	}
	r.addReportEntryLocked()
	return fmt.Errorf("%s traffic failed: %w", r.name, r.firstErr)
}

func (r *TrafficRecorder) AddReportEntry() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.addReportEntryLocked()
}

func (r *TrafficRecorder) record(record TrafficRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !record.Success && r.firstErr == nil {
		r.firstErr = fmt.Errorf("%s: %s", record.Label, record.Error)
	}
	r.records = append(r.records, record)
	if len(r.records) > r.limit {
		r.records = r.records[len(r.records)-r.limit:]
	}
}

func (r *TrafficRecorder) addReportEntryLocked() {
	if r.reported {
		return
	}
	r.reported = true

	payload := map[string]any{
		"name":     r.name,
		"limit":    r.limit,
		"records":  r.records,
		"firstErr": nil,
	}
	if r.firstErr != nil {
		payload["firstErr"] = r.firstErr.Error()
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		framework.Logf("failed to marshal traffic recorder report: %v", err)
		return
	}
	report.AddFileToReportEntry(
		fmt.Sprintf("traffic-recorder-%d-%s.json", time.Now().UnixNano(), r.name),
		data,
	)
}

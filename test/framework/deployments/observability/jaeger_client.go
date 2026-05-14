package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type jaegerServicesOutput struct {
	Data []string `json:"data"`
}

// Tag is a Jaeger span tag (Jaeger normalizes all values to string for transport).
type Tag struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Span is a slim Jaeger span representation; only the fields used by tests
// are included.
type Span struct {
	TraceID       string `json:"traceID"`
	SpanID        string `json:"spanID"`
	OperationName string `json:"operationName"`
	Tags          []Tag  `json:"tags"`
}

// Trace is a slim Jaeger trace representation; only the fields used by tests
// are included.
type Trace struct {
	TraceID string `json:"traceID"`
	Spans   []Span `json:"spans"`
}

type jaegerTracesOutput struct {
	Data []Trace `json:"data"`
}

func tracedServices(url string) ([]string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/api/services", url), http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	output := &jaegerServicesOutput{}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return nil, err
	}
	sort.Strings(output.Data)
	return output.Data, nil
}

// tracesForService fetches recent traces for a given Jaeger service name.
// Used by tests that need to inspect span tags (e.g., kuma.workload) rather
// than just verifying a service name is present.
func tracesForService(url, service string, limit int) ([]Trace, error) {
	q := fmt.Sprintf("%s/api/traces?service=%s&limit=%d&lookback=10m", url, service, limit)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, q, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	output := &jaegerTracesOutput{}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return nil, err
	}
	return output.Data, nil
}

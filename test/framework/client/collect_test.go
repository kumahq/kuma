package client

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kumahq/kuma/v2/test/server/types"
)

func TestRedactedHeaders(t *testing.T) {
	headers := map[string]string{
		"Authorization":  "Bearer token",
		"X-Api-Token":    "token",
		"X-Secret-Name":  "secret",
		"Cookie":         "session=secret",
		"X-Credential":   "credential",
		"Host":           "example.kuma.io",
		"X-Regular-Flag": "value",
	}

	got := redactedHeaders(headers)
	want := map[string]string{
		"Authorization":  redactedDiagnosticValue,
		"X-Api-Token":    redactedDiagnosticValue,
		"X-Secret-Name":  redactedDiagnosticValue,
		"Cookie":         redactedDiagnosticValue,
		"X-Credential":   redactedDiagnosticValue,
		"Host":           "example.kuma.io",
		"X-Regular-Flag": "value",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected redacted headers %v, got %v", want, got)
	}
	if headers["Authorization"] != "Bearer token" {
		t.Fatalf("expected original headers to stay unchanged, got %q", headers["Authorization"])
	}
}

func TestRedactedCommand(t *testing.T) {
	command := []string{
		"curl",
		"--header", "Authorization: Bearer token",
		"--header", "'X-Api-Token: token'",
		"-H", "\"X-Secret: secret\"",
		"--header=Cookie: session=secret",
		"-HX-Credential: credential",
		"--header", "Host: example.kuma.io",
		"--max-time", "5",
	}

	got := redactedCommand(command)
	want := []string{
		"curl",
		"--header", "Authorization: [redacted]",
		"--header", "'X-Api-Token: [redacted]'",
		"-H", "\"X-Secret: [redacted]\"",
		"--header=Cookie: [redacted]",
		"-HX-Credential: [redacted]",
		"--header", "Host: example.kuma.io",
		"--max-time", "5",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected redacted command %v, got %v", want, got)
	}
	if command[1] != "--header" || command[2] != "Authorization: Bearer token" {
		t.Fatalf("expected original command to stay unchanged, got %v", command)
	}
}

func TestRedactedCollectStdout(t *testing.T) {
	stdout, err := json.Marshal(types.EchoResponse{
		Instance: "echo-1",
		Received: types.EchoResponseReceived{
			Method: "GET",
			Path:   "/",
			Headers: map[string][]string{
				"Authorization": {"Bearer token"},
				"Cookie":        {"session=secret"},
				"Host":          {"example.kuma.io"},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal test response: %v", err)
	}

	got := redactedCollectStdout(string(stdout))

	var response types.EchoResponse
	if err := json.Unmarshal([]byte(got), &response); err != nil {
		t.Fatalf("failed to unmarshal redacted stdout: %v", err)
	}

	want := map[string][]string{
		"Authorization": {redactedDiagnosticValue},
		"Cookie":        {redactedDiagnosticValue},
		"Host":          {"example.kuma.io"},
	}
	if !reflect.DeepEqual(response.Received.Headers, want) {
		t.Fatalf("expected redacted echoed headers %v, got %v", want, response.Received.Headers)
	}
}

func TestRedactedCollectStdoutPreservesUnknownJSON(t *testing.T) {
	stdout := `{"error":"denied","details":{"reason":"rbac"}}`

	got := redactedCollectStdout(stdout)

	var response map[string]any
	if err := json.Unmarshal([]byte(got), &response); err != nil {
		t.Fatalf("failed to unmarshal preserved stdout: %v", err)
	}
	if response["error"] != "denied" {
		t.Fatalf("expected original JSON fields to be preserved, got %v", response)
	}
	details, ok := response["details"].(map[string]any)
	if !ok || details["reason"] != "rbac" {
		t.Fatalf("expected nested JSON fields to be preserved, got %v", response)
	}
}

func TestRedactedDiagnosticResponses(t *testing.T) {
	responses := []any{
		types.EchoResponse{
			Instance: "echo-1",
			Received: types.EchoResponseReceived{
				Headers: map[string][]string{
					"Authorization": {"Bearer token"},
					"Host":          {"example.kuma.io"},
				},
			},
		},
		FailureResponse{ResponseCode: 503},
	}

	got := redactedDiagnosticResponses(responses)

	echo, ok := got[0].(types.EchoResponse)
	if !ok {
		t.Fatalf("expected first response to stay an EchoResponse, got %T", got[0])
	}
	if !reflect.DeepEqual(echo.Received.Headers, map[string][]string{
		"Authorization": {redactedDiagnosticValue},
		"Host":          {"example.kuma.io"},
	}) {
		t.Fatalf("expected redacted echoed headers, got %v", echo.Received.Headers)
	}
	if !reflect.DeepEqual(got[1], responses[1]) {
		t.Fatalf("expected non-echo response to remain unchanged, got %v", got[1])
	}

	original := responses[0].(types.EchoResponse)
	if original.Received.Headers["Authorization"][0] != "Bearer token" {
		t.Fatalf("expected original responses to stay unchanged, got %v", original.Received.Headers)
	}
}

func TestBuildConcurrentDiagnosticUsesInnerReference(t *testing.T) {
	diagnostic := buildConcurrentDiagnostic(
		"echo.mesh",
		CollectResponsesOpts{
			numberOfRequests:      5,
			maxConcurrentRequests: 2,
			URL:                   "http://echo.mesh",
			Method:                "GET",
			Headers: map[string]string{
				"Authorization": "Bearer token",
			},
			Flags: []string{"--fail"},
		},
		result{
			idx: 1,
			err: diagnosticError{
				err:  errTest("boom"),
				file: "client-collect-cluster-source.json",
			},
		},
		[]any{
			types.EchoResponse{
				Instance: "echo-1",
				Received: types.EchoResponseReceived{
					Headers: map[string][]string{
						"Authorization": {"Bearer token"},
					},
				},
			},
		},
	)

	if diagnostic.FailedRequestDiagnostic != "client-collect-cluster-source.json" {
		t.Fatalf("expected failed request reference to be preserved, got %q", diagnostic.FailedRequestDiagnostic)
	}
	if diagnostic.URL != "" || diagnostic.Method != "" || diagnostic.Headers != nil || diagnostic.Flags != nil {
		t.Fatalf("expected overlapping request context to be omitted when inner diagnostic exists, got %+v", diagnostic)
	}

	echo := diagnostic.PartialResponses[0].(types.EchoResponse)
	if got := echo.Received.Headers["Authorization"][0]; got != redactedDiagnosticValue {
		t.Fatalf("expected concurrent partial responses to be redacted, got %q", got)
	}
}

type errTest string

func (e errTest) Error() string {
	return string(e)
}

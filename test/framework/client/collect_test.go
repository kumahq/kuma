package client

import (
	"reflect"
	"testing"
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

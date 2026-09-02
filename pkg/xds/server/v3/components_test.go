package v3

import (
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestADSServerRejectsSOTWStream(t *testing.T) {
	err := (&adsServer{}).StreamAggregatedResources(nil)
	if err == nil {
		t.Fatal("expected SOTW stream to be rejected")
	}
	if status.Code(err) != codes.Unimplemented {
		t.Fatalf("expected %v, got %v", codes.Unimplemented, status.Code(err))
	}

	msg := err.Error()
	if !strings.Contains(msg, "unsupported SOTW/state-of-the-world xDS") {
		t.Fatalf("expected error to mention unsupported SOTW/state-of-the-world xDS, got %q", msg)
	}
	if !strings.Contains(msg, "Delta ADS") {
		t.Fatalf("expected error to mention Delta ADS, got %q", msg)
	}
}

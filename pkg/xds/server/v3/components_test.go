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
	st := status.Convert(err)
	if st.Code() != codes.Unimplemented {
		t.Fatalf("expected %v, got %v", codes.Unimplemented, st.Code())
	}

	msg := st.Message()
	if !strings.Contains(msg, "unsupported SOTW/state-of-the-world xDS") {
		t.Fatalf("expected error to mention unsupported SOTW/state-of-the-world xDS, got %q", msg)
	}
	if !strings.Contains(msg, "restart the proxy") {
		t.Fatalf("expected error to mention restarting the proxy, got %q", msg)
	}
	if !strings.Contains(msg, "delta-capable bootstrap") {
		t.Fatalf("expected error to mention a delta-capable bootstrap, got %q", msg)
	}
	if !strings.Contains(msg, "Delta ADS") {
		t.Fatalf("expected error to mention Delta ADS, got %q", msg)
	}
}

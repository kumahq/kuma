package multizone

import (
	"strings"
	"testing"
	"time"

	config_types "github.com/kumahq/kuma/v3/pkg/config/types"
)

func TestEventBasedWatchdogConfigValidateRejectsNonPositiveIntervals(t *testing.T) {
	cfg := EventBasedWatchdogConfig{
		FlushInterval:      config_types.Duration{Duration: 0},
		FullResyncInterval: config_types.Duration{Duration: -1 * time.Second},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), ".FlushInterval must be positive") {
		t.Fatalf("expected flush interval validation error, got %v", err)
	}
	if !strings.Contains(err.Error(), ".FullResyncInterval must be positive") {
		t.Fatalf("expected full resync validation error, got %v", err)
	}
}

func TestKdsServerConfigValidateRequiresClientCaForClientCert(t *testing.T) {
	cfg := DefaultGlobalConfig().KDS
	cfg.RequireClientCert = true

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), ".TlsClientCaFile cannot be empty if RequireClientCert is true") {
		t.Fatalf("expected client CA validation error, got %v", err)
	}

	cfg.TlsClientCaFile = "/ca.crt"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no validation error, got %v", err)
	}
}

func TestKdsClientConfigValidateRequiresCertAndKeyTogether(t *testing.T) {
	cfg := DefaultZoneConfig().KDS
	cfg.TlsCertFile = "/tls.crt"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), ".TlsCertFile and .TlsKeyFile have to be set together") {
		t.Fatalf("expected cert/key validation error, got %v", err)
	}

	cfg.TlsKeyFile = "/tls.key"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no validation error, got %v", err)
	}
}

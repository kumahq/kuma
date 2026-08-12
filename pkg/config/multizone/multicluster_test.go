package multizone

import (
	"strings"
	"testing"
)

func TestZoneConfigValidateName(t *testing.T) {
	cases := []struct {
		name      string
		zoneName  string
		wantErr   bool
		errSubstr string
	}{
		{name: "empty name", zoneName: "", wantErr: true, errSubstr: "Name is mandatory"},
		{name: "valid lowercase name", zoneName: "zone-1", wantErr: false},
		{name: "rejects uppercase letters", zoneName: "Zone-1", wantErr: true, errSubstr: "valid RFC1123 DNS label"},
		{name: "rejects underscore", zoneName: "zone_1", wantErr: true, errSubstr: "valid RFC1123 DNS label"},
		{name: "rejects name longer than 63 characters", zoneName: strings.Repeat("a", 64), wantErr: true, errSubstr: "valid RFC1123 DNS label"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := ZoneConfig{
				Name: c.zoneName,
			}
			err := cfg.Validate()
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected validation error for zone name %q", c.zoneName)
				}
				if !strings.Contains(err.Error(), c.errSubstr) {
					t.Fatalf("expected error to contain %q, got %v", c.errSubstr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no validation error for zone name %q, got %v", c.zoneName, err)
			}
		})
	}
}

package system_names_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/core/system_names"
)

func TestIsSystem(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "system_foo", want: true},
		{name: "system_", want: true},
		{name: "foo", want: false},
		{name: "sys", want: false},
		{name: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := system_names.IsSystem(tt.name); got != tt.want {
				t.Fatalf("IsSystem(%q) = %t, want %t", tt.name, got, tt.want)
			}
		})
	}
}

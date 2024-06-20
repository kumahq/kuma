package validators

import "testing"

func TestValidateBandwidth(t *testing.T) {
	path := []string{"path"}

	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "sanity",
			input: "1kbps",
		},
		{
			name:  "without number",
			input: "Mbps",
		},
		{
			name:  "not exact match",
			input: "1bpsp",
			err: func() string {
				e := &ValidationError{}
				e.AddViolationAt(path, MustHaveBPSUnit)
				return e.Error()
			}(),
		},
		{
			name:  "bps is not allowed",
			input: "1bps",
			err: func() string {
				e := &ValidationError{}
				e.AddViolationAt(path, MustHaveBPSUnit)
				return e.Error()
			}(),
		},
		{
			name:  "float point number is not supported",
			input: "0.1kbps",
			err: func() string {
				e := &ValidationError{}
				e.AddViolationAt(path, MustHaveBPSUnit)
				return e.Error()
			}(),
		},
		{
			name:  "not defined",
			input: "",
			err: func() string {
				e := &ValidationError{}
				e.AddViolationAt(path, MustBeDefined)
				return e.Error()
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := ValidateBandwidth(path, tt.input)
			if actual.Error() != tt.err {
				t.Errorf("ValidateBandwidth(%s): expected %s, actual %s", tt.input, tt.err, actual)
			}
		})
	}
}

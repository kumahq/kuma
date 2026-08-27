package client

import (
	"errors"
	"testing"

	"github.com/kumahq/kuma/v2/test/framework"
)

type execOutcome struct {
	stdout string
	stderr string
	err    error
}

// fakeExecCluster serves canned Exec outcomes. Only Exec is ever called, so the
// embedded interface stays nil.
type fakeExecCluster struct {
	framework.Cluster

	outcomes []execOutcome
	calls    int
}

func (c *fakeExecCluster) Exec(string, string, string, ...string) (string, string, error) {
	o := c.outcomes[min(c.calls, len(c.outcomes)-1)]
	c.calls++
	return o.stdout, o.stderr, o.err
}

func TestExecCollectRetriesOnLostOutput(t *testing.T) {
	boom := errors.New("command terminated with exit code 22")

	cases := map[string]struct {
		outcomes  []execOutcome
		wantCalls int
		wantOut   string
		wantErr   error
	}{
		"lost output then success": {
			// client-go reports a dropped exec stream as exit 0 with no output.
			outcomes:  []execOutcome{{}, {stdout: `{"instance":"server"}`}},
			wantCalls: 2,
			wantOut:   `{"instance":"server"}`,
		},
		"output lost on every attempt": {
			outcomes:  []execOutcome{{}},
			wantCalls: collectExecAttempts,
		},
		"successful first attempt is not repeated": {
			outcomes:  []execOutcome{{stdout: `{"instance":"server"}`}},
			wantCalls: 1,
			wantOut:   `{"instance":"server"}`,
		},
		"reported exec failure is not retried": {
			outcomes:  []execOutcome{{err: boom}},
			wantCalls: 1,
			wantErr:   boom,
		},
		"stderr without stdout is not retried": {
			// curl reported something, so the transport delivered the output.
			outcomes:  []execOutcome{{stderr: "curl: (6) Could not resolve host"}},
			wantCalls: 1,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cluster := &fakeExecCluster{outcomes: tc.outcomes}

			stdout, _, err := execCollect(cluster, "ns", "pod", "container", "curl", "http://example")

			if cluster.calls != tc.wantCalls {
				t.Errorf("Exec called %d times, want %d", cluster.calls, tc.wantCalls)
			}
			if stdout != tc.wantOut {
				t.Errorf("stdout = %q, want %q", stdout, tc.wantOut)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

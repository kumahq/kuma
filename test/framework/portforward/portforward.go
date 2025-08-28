package portforward

import (
	"errors"
	"math"

	"github.com/gruntwork-io/terratest/modules/k8s"
)

var EnvoyAdminDefaultSpec = Spec{RemotePort: 9901}

type Tunnel struct {
	tunnel   *k8s.Tunnel
	Endpoint string `json:"endpoint"`
}

func NewTunnel(tunnel *k8s.Tunnel, endpoint string) Tunnel {
	return Tunnel{
		tunnel:   tunnel,
		Endpoint: endpoint,
	}
}

func (p Tunnel) Close() {
	if p.tunnel != nil {
		p.tunnel.Close()
	}
}

type Spec struct {
	AppName    string
	Namespace  string
	RemotePort int
}

func (a Spec) WithDefaults(def Spec) Spec {
	if a.AppName == "" {
		a.AppName = def.AppName
	}

	if a.Namespace == "" {
		a.Namespace = def.Namespace
	}

	if a.RemotePort == 0 {
		a.RemotePort = def.RemotePort
	}

	return a
}

func (a Spec) ValidateFullSpec() error {
	var errs []error

	if a.AppName == "" {
		errs = append(errs, errors.New(".AppName is required"))
	}

	if a.Namespace == "" {
		errs = append(errs, errors.New(".Namespace is required"))
	}

	if a.RemotePort < 1 || a.RemotePort > math.MaxUint16 {
		errs = append(errs, errors.New(".Port must be between 1 and 65535"))
	}

	return errors.Join(errs...)
}

// Matches reports whether 'args' matches the receiver
// Matching semantics:
//   - Zero value in 'args' means "don't care" (wildcard) for that field.
//   - All non-zero fields in 'args' must equal the corresponding fields in the
//     defaulted receiver
//   - An entirely zero 'args' acts as a catch‑all (matches everything)
func (a Spec) Matches(args Spec) bool {
	// Catch‑all: no specific properties requested.
	if args == (Spec{}) {
		return true
	}

	// Exact match fast-path.
	if args == a {
		return true
	}

	// Field-wise matching (wildcard on zero values).
	if args.Namespace != "" && args.Namespace != a.Namespace {
		return false
	}

	if args.AppName != "" && args.AppName != a.AppName {
		return false
	}

	if args.RemotePort != 0 && args.RemotePort != a.RemotePort {
		return false
	}

	return true
}

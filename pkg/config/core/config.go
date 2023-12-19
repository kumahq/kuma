package core

import "github.com/pkg/errors"

type EnvironmentType = string

const (
	KubernetesEnvironment EnvironmentType = "kubernetes"
	UniversalEnvironment  EnvironmentType = "universal"
)

// Control Plane mode

type CpMode = string

const (
	// Deprecated: use zone
	Standalone CpMode = "standalone"
	Zone       CpMode = "zone"
	Global     CpMode = "global"
)

// ValidateCpMode to check modes of kuma-cp
func ValidateCpMode(mode CpMode) error {
	if mode != Standalone && mode != Zone && mode != Global {
		return errors.Errorf("invalid mode. Available modes: %s, %s, %s", Standalone, Zone, Global)
	}
	return nil
}

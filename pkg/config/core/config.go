package core

import "fmt"

type EnvironmentType = string

const (
	KubernetesEnvironment EnvironmentType = "kubernetes"
	UniversalEnvironment  EnvironmentType = "universal"
)

// Control Plane mode

type CpMode = string

const (
	Standalone CpMode = "standalone"
	Zone       CpMode = "zone"
	Global     CpMode = "global"
)

// ValidateCpMode to check modes of kuma-cp
func ValidateCpMode(mode CpMode) error {
	if mode != Standalone && mode != Zone && mode != Global {
		return fmt.Errorf("invalid mode. Available modes: %s, %s, %s", Standalone, Zone, Global)
	}
	return nil
}

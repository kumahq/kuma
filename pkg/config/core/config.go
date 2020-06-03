package core

import "github.com/pkg/errors"

type EnvironmentType = string

// Control Plane mode
type CpMode = string

const (
	KubernetesEnvironment EnvironmentType = "kubernetes"
	UniversalEnvironment  EnvironmentType = "universal"
)

const (
	Standalone CpMode = "standalone"
	Local      CpMode = "local"
	Global     CpMode = "global"
)

// ValidateCpMode to check modes of kuma-cp
func ValidateCpMode(mode CpMode) error {
	if mode != Standalone && mode != Local && mode != Global {
		return errors.Errorf("invalid mode. Available modes: %s, %s, %s", Standalone, Local, Global)
	}
	return nil
}

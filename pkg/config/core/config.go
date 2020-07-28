package core

import "github.com/go-errors/errors"

type EnvironmentType = string

const (
	KubernetesEnvironment EnvironmentType = "kubernetes"
	UniversalEnvironment  EnvironmentType = "universal"
)

// Control Plane mode

type CpMode = string

const (
	Standalone CpMode = "standalone"
	Remote     CpMode = "remote"
	Global     CpMode = "global"
)

// ValidateCpMode to check modes of kuma-cp
func ValidateCpMode(mode CpMode) error {
	if mode != Standalone && mode != Remote && mode != Global {
		return errors.Errorf("invalid mode. Available modes: %s, %s, %s", Standalone, Remote, Global)
	}
	return nil
}

package core

import "github.com/pkg/errors"

type EnvironmentType = string

// mode type for multi-cluster
type CpMode = string

const (
	KubernetesEnvironment EnvironmentType = "kubernetes"
	UniversalEnvironment  EnvironmentType = "universal"
)

const (
	StandAlone CpMode = "standalone"
	Local      CpMode = "local"
	Global     CpMode = "global"
)

// ValidateCpMode to check modes of kuma-cp
func ValidateCpMode(mode CpMode) error {
	if mode != StandAlone && mode != Local && mode != Global {
		return errors.Errorf("mode should be either %s, %s or %s", StandAlone, Local, Global)
	}
	return nil
}

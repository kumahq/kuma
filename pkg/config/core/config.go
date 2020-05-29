package core

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

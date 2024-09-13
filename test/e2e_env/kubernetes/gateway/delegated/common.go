package delegated

type Config struct {
	Namespace                   string
	NamespaceOutsideMesh        string
	Mesh                        string
	KicIP                       string
	CpNamespace                 string
	ObservabilityDeploymentName string
	IPV6                        bool
	MeshServiceMode             string
}

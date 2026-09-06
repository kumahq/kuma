package delegated

// TestServerReplicas is how many test-server instances the delegated gateway
// suite runs. The gateway load balances over all of them, so a spec that waits
// for a policy to land has to wait for every one.
const TestServerReplicas = 3

type Config struct {
	Namespace                   string
	NamespaceOutsideMesh        string
	Mesh                        string
	KicIP                       string
	CpNamespace                 string
	ObservabilityDeploymentName string
	IPV6                        bool
}

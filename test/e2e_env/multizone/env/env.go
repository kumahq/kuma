package env

import "github.com/kumahq/kuma/test/framework"

var Global *framework.UniversalCluster
var KubeZone1 *framework.K8sCluster
var KubeZone2 *framework.K8sCluster
var UniZone1 *framework.UniversalCluster
var UniZone2 *framework.UniversalCluster

func Zones() []framework.Cluster {
	return []framework.Cluster{KubeZone1, KubeZone2, UniZone1, UniZone2}
}

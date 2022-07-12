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

func GlobalCluster() framework.Cluster {
	return Global
}

func KubeZone1Cluster() framework.Cluster {
	return KubeZone1
}

func KubeZone2Cluster() framework.Cluster {
	return KubeZone1
}

func UniZone1Cluster() framework.Cluster {
	return UniZone1
}

func UniZone2Cluster() framework.Cluster {
	return UniZone2
}

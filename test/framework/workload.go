package framework

import (
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
)

func GetWorkload(cluster Cluster, workloadName, meshName string) (*workload_api.WorkloadResource, error) {
	out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "workload", "-m", meshName, workloadName, "-ojson")
	if err != nil {
		return nil, err
	}
	res, err := rest.JSON.UnmarshalCore([]byte(out))
	if err != nil {
		return nil, err
	}
	return res.(*workload_api.WorkloadResource), nil
}

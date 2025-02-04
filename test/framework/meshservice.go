package framework

import (
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

func GetMeshServiceStatus(cluster Cluster, meshServiceName, meshName string) (*meshservice_api.MeshService, *meshservice_api.MeshServiceStatus, error) {
	out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshservice", "-m", meshName, meshServiceName, "-ojson")
	if err != nil {
		return nil, nil, err
	}
	res, err := rest.JSON.Unmarshal([]byte(out), meshservice_api.MeshServiceResourceTypeDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return res.GetSpec().(*meshservice_api.MeshService), res.GetStatus().(*meshservice_api.MeshServiceStatus), nil
}

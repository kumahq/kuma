package framework

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

func DeleteAllResourcesUniversal(kumactl KumactlOptions, descriptor core_model.ResourceTypeDescriptor, mesh string) error {
	dpps, err := kumactl.RunKumactlAndGetOutput("get", descriptor.KumactlListArg, "-m", mesh, "-o", "json")
	if err != nil {
		return err
	}
	list := descriptor.NewList()
	if err := rest.JSON.UnmarshalListToCore([]byte(dpps), list); err != nil {
		return err
	}
	for _, item := range list.GetItems() {
		_, err := kumactl.RunKumactlAndGetOutput("delete", descriptor.KumactlArg, item.GetMeta().GetName(), "-m", mesh)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteMeshResourcesKubernetes(cluster Cluster, mesh string, resources ...string) error {
	var errs []string

	for _, resource := range resources {
		if err := k8s.RunKubectlE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(),
			"delete",
			resource,
			"--all-namespaces",
			"--selector",
			fmt.Sprintf("%s=%s", mesh_proto.MeshTag, mesh),
		); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf(
			"deleting mesh resources failed with errors:\n\t%s",
			strings.Join(errs, "\n\t"),
		)
	}

	return nil
}

func WaitForMesh(mesh string, clusters []Cluster) error {
	for _, cluster := range clusters {
		if err := WaitForResource(cluster, core_mesh.MeshResourceTypeDescriptor, core_model.ResourceKey{Name: mesh}); err != nil {
			return err
		}
	}
	return nil
}

func WaitForResource(cluster Cluster, descriptor core_model.ResourceTypeDescriptor, key core_model.ResourceKey) error {
	_, err := retry.DoWithRetryE(cluster.GetTesting(), "wait for resource "+key.Mesh+"/"+key.Name, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			args := []string{"get", descriptor.KumactlArg, key.Name}
			if key.Mesh != "" {
				args = append(args, "-m", key.Mesh)
			}
			_, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput(args...)
			return "", err
		})
	return err
}

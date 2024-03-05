package framework

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/test/framework/kumactl"
)

func DeleteMeshResources(cluster Cluster, mesh string, descriptor ...core_model.ResourceTypeDescriptor) error {
	var errs []error

	for _, desc := range descriptor {
		if _, ok := cluster.(*K8sCluster); ok {
			errs = append(errs, deleteMeshResourcesKubernetes(cluster, mesh, desc))
			continue
		}

		errs = append(errs, deleteMeshResourcesUniversal(*cluster.GetKumactlOptions(), mesh, desc))
	}

	return errors.Join(errs...)
}

func DeleteMeshPolicyOrError(cluster Cluster, descriptor core_model.ResourceTypeDescriptor, policyName string) error {
	_, err := retry.DoWithRetryE(
		cluster.GetTesting(),
		"delete policy",
		10,
		time.Second,
		func() (string, error) {
			return k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(Config.KumaNamespace),
				"delete",
				descriptor.KumactlArg,
				policyName,
			)
		},
	)

	return err
}

func deleteMeshResourcesUniversal(kumactl kumactl.KumactlOptions, mesh string, descriptor core_model.ResourceTypeDescriptor) error {
	list, err := allResourcesOfType(kumactl, descriptor, mesh)
	if err != nil {
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

func allResourcesOfType(kumactl kumactl.KumactlOptions, descriptor core_model.ResourceTypeDescriptor, mesh string) (core_model.ResourceList, error) {
	dpps, err := kumactl.RunKumactlAndGetOutput("get", descriptor.KumactlListArg, "-m", mesh, "-o", "json")
	if err != nil {
		return nil, err
	}
	list := descriptor.NewList()
	if err := rest.JSON.UnmarshalListToCore([]byte(dpps), list); err != nil {
		return nil, err
	}
	return list, err
}

func deleteMeshResourcesKubernetes(cluster Cluster, mesh string, resource core_model.ResourceTypeDescriptor) error {
	args := []string{"delete", strings.ReplaceAll(strings.ToLower(resource.PluralDisplayName), " ", "")}
	if resource.IsPluginOriginated {
		// because all new policies have a mesh label, we can just delete by selecting a label
		args = append(args, "--all-namespaces", "--selector", fmt.Sprintf("%s=%s", mesh_proto.MeshTag, mesh))
		if err := k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(), args...); err != nil {
			return err
		}
	} else {
		list, err := allResourcesOfType(*cluster.GetKumactlOptions(), resource, mesh)
		if err != nil {
			return err
		}
		for _, item := range list.GetItems() {
			itemDelArgs := append(args, item.GetMeta().GetName())
			if err := k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(), itemDelArgs...); err != nil {
				return err
			}
		}
	}
	return nil
}

func WaitForMesh(mesh string, clusters []Cluster) error {
	return WaitForResource(core_mesh.MeshResourceTypeDescriptor, core_model.ResourceKey{Name: mesh}, clusters...)
}

func WaitForResource(descriptor core_model.ResourceTypeDescriptor, key core_model.ResourceKey, clusters ...Cluster) error {
	for _, c := range clusters {
		_, err := retry.DoWithRetryE(c.GetTesting(), "wait for resource "+key.Mesh+"/"+key.Name, DefaultRetries, DefaultTimeout,
			func() (string, error) {
				args := []string{"get", descriptor.KumactlArg, key.Name}
				if key.Mesh != "" {
					args = append(args, "-m", key.Mesh)
				}
				_, err := c.GetKumactlOptions().RunKumactlAndGetOutput(args...)
				return "", err
			})
		if err != nil {
			return err
		}
	}
	return nil
}

package framework

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v2/test/framework/kumactl"
)

var kumactlConnectionErrorSubstrings = []string{
	"connect: connection refused",
	"connection reset by peer",
	"EOF",
	"failed to retrieve server version",
	"use of closed network connection",
	"unexpected EOF",
}

func DeleteMeshResources(cluster Cluster, mesh string, descriptor ...core_model.ResourceTypeDescriptor) error {
	mode := cluster.GetKuma().Mode()
	_, err := retry.DoWithRetryContextE(
		cluster.GetTesting(), context.Background(),
		"delete mesh resources",
		10,
		time.Second,
		func() (string, error) {
			var errs []error

			for _, desc := range descriptor {
				if _, ok := cluster.(*K8sCluster); ok {
					errs = append(errs, deleteMeshResourcesKubernetes(cluster, mesh, mode, desc))
					continue
				}

				errs = append(errs, deleteMeshResourcesUniversal(*cluster.GetKumactlOptions(), mesh, mode, desc))
			}

			return "", errors.Join(errs...)
		},
	)
	return err
}

// isManagedByMode reports whether a resource with the given metadata is
// managed by a CP running in the given mode. A CP can only delete resources
// it owns: Global owns global-origin resources, Zone owns zone-origin
// resources. Unlabeled resources are considered owned by any CP (matching
// the validator in pkg/api-server/resource_endpoints.go which only rejects
// when origin is explicitly set and wrong).
func isManagedByMode(meta core_model.ResourceMeta, mode core.CpMode) bool {
	origin, ok := core_model.ResourceOrigin(meta)
	if !ok {
		return true
	}
	switch mode {
	case core.Global:
		return origin == mesh_proto.GlobalResourceOrigin
	case core.Zone:
		return origin == mesh_proto.ZoneResourceOrigin
	}
	return true
}

func DeleteMeshPolicyOrError(cluster Cluster, descriptor core_model.ResourceTypeDescriptor, policyName string, mesh ...string) error {
	_, err := retry.DoWithRetryContextE(
		cluster.GetTesting(), context.Background(),
		"delete policy",
		10,
		time.Second,
		func() (string, error) {
			if _, ok := cluster.(*K8sCluster); ok {
				return k8s.RunKubectlAndGetOutputContextE(
					cluster.GetTesting(), context.Background(),
					cluster.GetKubectlOptions(Config.KumaNamespace),
					"delete",
					descriptor.KumactlArg,
					policyName,
				)
			}
			return cluster.GetKumactlOptions().RunKumactlAndGetOutput("delete", descriptor.KumactlArg, policyName, "-m", mesh[0])
		},
	)

	return err
}

func deleteMeshResourcesUniversal(kumactl kumactl.KumactlOptions, mesh string, mode core.CpMode, descriptor core_model.ResourceTypeDescriptor) error {
	list, err := allResourcesOfType(kumactl, descriptor, mesh)
	if err != nil {
		return err
	}
	for _, item := range list.GetItems() {
		if !isManagedByMode(item.GetMeta(), mode) {
			continue
		}
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

func deleteMeshResourcesKubernetes(cluster Cluster, mesh string, mode core.CpMode, resource core_model.ResourceTypeDescriptor) error {
	args := []string{"delete", strings.ReplaceAll(strings.ToLower(resource.PluralDisplayName), " ", "")}
	if resource.IsPluginOriginated {
		// Delete only resources owned by this CP: those matching the CP mode
		// in their origin label and those without an origin label at all.
		// Resources synced from other CPs (origin != mode) are managed by the
		// originating CP and must be skipped — the REST API rejects writes to
		// them (see pkg/api-server/resource_endpoints.go validateOriginForWrite).
		selectors := []string{
			fmt.Sprintf("%s=%s,%s=%s", mesh_proto.MeshTag, mesh, mesh_proto.ResourceOriginLabel, originLabelFor(mode)),
			fmt.Sprintf("%s=%s,!%s", mesh_proto.MeshTag, mesh, mesh_proto.ResourceOriginLabel),
		}
		for _, sel := range selectors {
			delArgs := append([]string{}, args...)
			delArgs = append(delArgs, "--all-namespaces", "--selector", sel)
			if err := k8s.RunKubectlContextE(cluster.GetTesting(), context.Background(), cluster.GetKubectlOptions(), delArgs...); err != nil {
				return err
			}
		}
	} else {
		list, err := allResourcesOfType(*cluster.GetKumactlOptions(), resource, mesh)
		if err != nil {
			return err
		}
		for _, item := range list.GetItems() {
			if !isManagedByMode(item.GetMeta(), mode) {
				continue
			}
			itemDelArgs := append(args, item.GetMeta().GetName())
			if err := k8s.RunKubectlContextE(cluster.GetTesting(), context.Background(), cluster.GetKubectlOptions(), itemDelArgs...); err != nil {
				return err
			}
		}
	}
	return nil
}

func originLabelFor(mode core.CpMode) string {
	if mode == core.Global {
		return string(mesh_proto.GlobalResourceOrigin)
	}
	return string(mesh_proto.ZoneResourceOrigin)
}

func WaitForMesh(mesh string, clusters []Cluster) error {
	return WaitForResource(core_mesh.MeshResourceTypeDescriptor, core_model.ResourceKey{Name: mesh}, clusters...)
}

// WaitForZoneOnline polls `kumactl inspect zones` on the given Global CP
// cluster until the named zone appears with status Online. Use this after
// starting a new zone CP dynamically (e.g. NewUniversalCluster + Kuma(core.Zone))
// to ensure KDS registration and MeshService propagation have completed
// before running cross-zone assertions.
func WaitForZoneOnline(global Cluster, zoneName string) error {
	_, err := retry.DoWithRetryContextE(
		global.GetTesting(), context.Background(),
		fmt.Sprintf("wait for zone %s online", zoneName),
		// 120 retries * 3s = 6 min. Cold-start KDS connection on
		// IPv6 kindIpv6 clusters can take 60-90s; the previous 60s
		// budget was not enough.
		4*DefaultRetries, DefaultTimeout,
		func() (string, error) {
			out, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			if err != nil {
				return "", err
			}
			if !strings.Contains(out, zoneName) {
				return "", fmt.Errorf("zone %s not found in zones list:\n%s", zoneName, out)
			}
			for line := range strings.SplitSeq(out, "\n") {
				if strings.Contains(line, zoneName) {
					if !strings.Contains(line, "Online") {
						return "", fmt.Errorf("zone %s not Online: %s", zoneName, line)
					}
					return "", nil
				}
			}
			return "", fmt.Errorf("zone %s not found in zones output", zoneName)
		},
	)
	return err
}

func WaitForResource(descriptor core_model.ResourceTypeDescriptor, key core_model.ResourceKey, clusters ...Cluster) error {
	for _, c := range clusters {
		_, err := retry.DoWithRetryContextE(c.GetTesting(), context.Background(), "wait for resource "+key.Mesh+"/"+key.Name, DefaultRetries, DefaultTimeout,
			func() (string, error) {
				args := []string{"get", descriptor.KumactlArg, key.Name}
				if key.Mesh != "" {
					args = append(args, "-m", key.Mesh)
				}
				out, err := c.GetKumactlOptions().RunKumactlAndGetOutput(args...)
				if err != nil && isKumactlConnectionError(out, err) {
					if cluster, ok := c.(*K8sCluster); ok {
						if refreshErr := cluster.RefreshKumaCPPortForwards(); refreshErr != nil {
							return "", errors.Join(err, refreshErr)
						}
					}
				}
				return "", err
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func isKumactlConnectionError(output string, err error) bool {
	if err == nil {
		return false
	}

	msg := output + "\n" + err.Error()
	for _, substring := range kumactlConnectionErrorSubstrings {
		if strings.Contains(msg, substring) {
			return true
		}
	}
	return false
}

func NumberOfResources(c Cluster, resource core_model.ResourceTypeDescriptor) (int, error) {
	output, err := c.GetKumactlOptions().RunKumactlAndGetOutput("get", resource.KumactlListArg, "-o", "json")
	if err != nil {
		return 0, err
	}
	t := struct {
		Total int `json:"total"`
	}{}
	if err := json.Unmarshal([]byte(output), &t); err != nil {
		return 0, err
	}
	return t.Total, nil
}

// NumberOfResourcesByPath counts resources by querying the cluster's REST API
// directly at the given path (e.g. "/zone-ingresses"). Use this instead of
// NumberOfResources when the cluster runs an old CP that may not support the
// current kumactl endpoint names.
func NumberOfResourcesByPath(c Cluster, path string) (int, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.GetKuma().GetAPIServerAddress()+path, http.NoBody)
	if err != nil {
		return 0, err
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	t := struct {
		Total int `json:"total"`
	}{}
	if err := json.Unmarshal(body, &t); err != nil {
		return 0, err
	}
	return t.Total, nil
}

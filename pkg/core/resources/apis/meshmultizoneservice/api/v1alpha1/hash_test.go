package v1alpha1_test

import (
	"bytes"
	"testing"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

func TestMeshMultiZoneServiceXDSHashIgnoresConditionsButTracksResolvedBackends(t *testing.T) {
	first := builders.MeshMultiZoneService().
		WithMesh("default").
		WithName("backend").
		WithServiceLabelSelector(map[string]string{"app": "backend"}).
		AddIntPort(8080, core_meta.ProtocolHTTP).
		Build()
	first.Meta.(*test_model.ResourceMeta).Version = "1"
	first.Status = &api.MeshMultiZoneServiceStatus{
		Addresses: []hostnamegenerator_api.Address{{Hostname: "backend.mesh"}},
		VIPs:      []meshservice_api.VIP{{IP: "243.0.0.1"}},
		MeshServices: []api.MatchedMeshService{{
			Name:      "backend",
			Namespace: "default",
			Zone:      "zone-1",
			Mesh:      "default",
		}},
		Conditions: []common_api.Condition{{
			Type:   api.MeshServicesMatchedCondition,
			Status: v1.ConditionTrue,
			Reason: api.MatchesFoundReason,
		}},
	}

	sameXDSDifferentConditions := builders.MeshMultiZoneService().
		WithMesh("default").
		WithName("backend").
		WithServiceLabelSelector(map[string]string{"app": "backend"}).
		AddIntPort(8080, core_meta.ProtocolHTTP).
		Build()
	sameXDSDifferentConditions.Meta.(*test_model.ResourceMeta).Version = "2"
	sameXDSDifferentConditions.Status = &api.MeshMultiZoneServiceStatus{
		Addresses: []hostnamegenerator_api.Address{{Hostname: "backend.mesh"}},
		VIPs:      []meshservice_api.VIP{{IP: "243.0.0.1"}},
		MeshServices: []api.MatchedMeshService{{
			Name:      "backend",
			Namespace: "default",
			Zone:      "zone-1",
			Mesh:      "default",
		}},
		Conditions: []common_api.Condition{{
			Type:   api.MeshServicesMatchedCondition,
			Status: v1.ConditionFalse,
			Reason: api.NoMatchesFoundReason,
		}},
	}

	changedBackends := builders.MeshMultiZoneService().
		WithMesh("default").
		WithName("backend").
		WithServiceLabelSelector(map[string]string{"app": "backend"}).
		AddIntPort(8080, core_meta.ProtocolHTTP).
		Build()
	changedBackends.Meta.(*test_model.ResourceMeta).Version = "3"
	changedBackends.Status = &api.MeshMultiZoneServiceStatus{
		Addresses: []hostnamegenerator_api.Address{{Hostname: "backend.mesh"}},
		VIPs:      []meshservice_api.VIP{{IP: "243.0.0.2"}},
		MeshServices: []api.MatchedMeshService{{
			Name:      "backend-v2",
			Namespace: "default",
			Zone:      "zone-1",
			Mesh:      "default",
		}},
	}

	if !bytes.Equal(first.XDSHash(), sameXDSDifferentConditions.XDSHash()) {
		t.Fatal("expected MeshMultiZoneService XDSHash to ignore version and conditions")
	}
	if bytes.Equal(first.XDSHash(), changedBackends.XDSHash()) {
		t.Fatal("expected MeshMultiZoneService XDSHash to change when resolved backends change")
	}
	if bytes.Equal(first.Hash(), sameXDSDifferentConditions.Hash()) {
		t.Fatal("expected MeshMultiZoneService Hash to change when version/status changes")
	}
}

package v1alpha1_test

import (
	"bytes"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func TestMeshOpenTelemetryBackendXDSHashIgnoresConditions(t *testing.T) {
	first := &api.MeshOpenTelemetryBackendResource{
		Meta: &test_model.ResourceMeta{
			Mesh:    "default",
			Name:    "collector",
			Version: "1",
		},
		Spec: &api.MeshOpenTelemetryBackend{
			Endpoint: &api.Endpoint{
				Address: pointer.To("otel-collector.mesh"),
				Port:    pointer.To(int32(4317)),
			},
			Protocol: pointer.To(api.ProtocolGRPC),
		},
		Status: &api.MeshOpenTelemetryBackendStatus{
			Conditions: []common_api.Condition{{
				Type:   api.ReferencedByPoliciesCondition,
				Status: v1.ConditionTrue,
				Reason: api.ReferencedReason,
			}},
		},
	}

	second := &api.MeshOpenTelemetryBackendResource{
		Meta: &test_model.ResourceMeta{
			Mesh:    "default",
			Name:    "collector",
			Version: "2",
		},
		Spec: &api.MeshOpenTelemetryBackend{
			Endpoint: &api.Endpoint{
				Address: pointer.To("otel-collector.mesh"),
				Port:    pointer.To(int32(4317)),
			},
			Protocol: pointer.To(api.ProtocolGRPC),
		},
		Status: &api.MeshOpenTelemetryBackendStatus{
			Conditions: []common_api.Condition{{
				Type:   api.ReferencedByPoliciesCondition,
				Status: v1.ConditionFalse,
				Reason: api.NotReferencedReason,
			}},
		},
	}

	if !bytes.Equal(first.XDSHash(), second.XDSHash()) {
		t.Fatal("expected MeshOpenTelemetryBackend XDSHash to ignore version and conditions")
	}
	if bytes.Equal(first.Hash(), second.Hash()) {
		t.Fatal("expected MeshOpenTelemetryBackend Hash to change when version/status changes")
	}
}

package v1alpha1_test

import (
	"bytes"
	"testing"

	api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func TestMeshTrustXDSHashIgnoresStatusOrigin(t *testing.T) {
	first := builders.MeshTrust().WithMesh("default").WithName("trust-1").Build()
	first.Meta.(*test_model.ResourceMeta).Version = "1"
	first.Status = &api.MeshTrustStatus{
		Origin: &api.Origin{KRI: pointer.To("meshidentity:default/identity-1")},
	}

	second := builders.MeshTrust().WithMesh("default").WithName("trust-1").Build()
	second.Meta.(*test_model.ResourceMeta).Version = "2"
	second.Status = &api.MeshTrustStatus{
		Origin: &api.Origin{KRI: pointer.To("meshidentity:default/identity-2")},
	}

	if !bytes.Equal(first.XDSHash(), second.XDSHash()) {
		t.Fatal("expected MeshTrust XDSHash to ignore version and status origin")
	}
	if bytes.Equal(first.Hash(), second.Hash()) {
		t.Fatal("expected MeshTrust Hash to change when version/status changes")
	}
}

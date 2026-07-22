package v1alpha1_test

import (
	"bytes"
	"testing"

	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

func TestMeshExternalServiceXDSHashTracksDNSStatusButNotVersion(t *testing.T) {
	first := builders.MeshExternalService().
		WithMesh("default").
		WithName("payments").
		WithKumaVIP("242.0.0.1").
		Build()
	first.Meta.(*test_model.ResourceMeta).Version = "1"
	first.Status.Addresses = []hostnamegenerator_api.Address{{Hostname: "payments.mesh"}}

	sameContent := builders.MeshExternalService().
		WithMesh("default").
		WithName("payments").
		WithKumaVIP("242.0.0.1").
		Build()
	sameContent.Meta.(*test_model.ResourceMeta).Version = "2"
	sameContent.Status.Addresses = []hostnamegenerator_api.Address{{Hostname: "payments.mesh"}}

	changedStatus := builders.MeshExternalService().
		WithMesh("default").
		WithName("payments").
		WithKumaVIP("242.0.0.2").
		Build()
	changedStatus.Meta.(*test_model.ResourceMeta).Version = "3"
	changedStatus.Status.Addresses = []hostnamegenerator_api.Address{{Hostname: "payments-alt.mesh"}}

	if !bytes.Equal(first.XDSHash(), sameContent.XDSHash()) {
		t.Fatal("expected MeshExternalService XDSHash to ignore version-only writes")
	}
	if bytes.Equal(first.XDSHash(), changedStatus.XDSHash()) {
		t.Fatal("expected MeshExternalService XDSHash to change when VIP/domains change")
	}
	if bytes.Equal(first.Hash(), sameContent.Hash()) {
		t.Fatal("expected MeshExternalService Hash to change when version changes")
	}
}

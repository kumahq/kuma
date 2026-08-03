package v1alpha1_test

import (
	"bytes"
	"testing"

	api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

func meshZoneAddress(version string, address string, port int32) *api.MeshZoneAddressResource {
	return &api.MeshZoneAddressResource{
		Meta: &test_model.ResourceMeta{
			Mesh:    "default",
			Name:    "zone-proxy-east",
			Version: version,
		},
		Spec: &api.MeshZoneAddress{
			Address: address,
			Port:    port,
		},
	}
}

func TestMeshZoneAddressHashTracksResolvedAddress(t *testing.T) {
	first := meshZoneAddress("1", "10.0.0.1", 10001)
	sameContent := meshZoneAddress("1", "10.0.0.1", 10001)
	rotatedIP := meshZoneAddress("1", "10.0.0.2", 10001)
	changedPort := meshZoneAddress("1", "10.0.0.1", 10002)

	if !bytes.Equal(first.Hash(), sameContent.Hash()) {
		t.Fatal("expected MeshZoneAddress Hash to be stable for identical resources")
	}
	// the mesh context stores the address after DNS resolution, so a load balancer
	// hostname keeps the same resourceVersion while resolving to a new IP
	if bytes.Equal(first.Hash(), rotatedIP.Hash()) {
		t.Fatal("expected MeshZoneAddress Hash to change when the address changes")
	}
	if bytes.Equal(first.Hash(), changedPort.Hash()) {
		t.Fatal("expected MeshZoneAddress Hash to change when the port changes")
	}
}

func TestMeshZoneAddressHashTracksVersion(t *testing.T) {
	first := meshZoneAddress("1", "10.0.0.1", 10001)
	newVersion := meshZoneAddress("2", "10.0.0.1", 10001)

	if bytes.Equal(first.Hash(), newVersion.Hash()) {
		t.Fatal("expected MeshZoneAddress Hash to change when the version changes")
	}
}

func TestMeshZoneAddressHashHandlesNilSpec(t *testing.T) {
	withNilSpec := &api.MeshZoneAddressResource{
		Meta: &test_model.ResourceMeta{Mesh: "default", Name: "zone-proxy-east"},
	}

	if len(withNilSpec.Hash()) == 0 {
		t.Fatal("expected MeshZoneAddress Hash to be computed for a nil spec")
	}
}

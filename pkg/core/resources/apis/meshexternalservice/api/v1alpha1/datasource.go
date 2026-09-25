package v1alpha1

import (
	"errors"
	"fmt"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	datasource_api "github.com/kumahq/kuma/v2/api/common/v1alpha1/datasource"
	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
)

func (ds *VerificationDataSource) IsLegacy() bool {
	return ds.Secret != nil || ds.Inline != nil || ds.InlineString != nil
}

// ToProto converts both the legacy and the SecureDataSource shape to the DataSource
// read by the mesh-snapshot loader, so neither shape adds store I/O on xDS generation.
func (ds *VerificationDataSource) ToProto() (*system_proto.DataSource, error) {
	if ds.Type == nil {
		legacy := &common_api.DataSource{Secret: ds.Secret, Inline: ds.Inline, InlineString: ds.InlineString}
		return legacy.ConvertToProto(), nil
	}
	switch *ds.Type {
	case datasource_api.SecureDataSourceSecretRef:
		if ds.SecretRef == nil {
			return nil, errors.New("secretRef must be defined")
		}
		return &system_proto.DataSource{Type: &system_proto.DataSource_Secret{Secret: ds.SecretRef.Name}}, nil
	case datasource_api.SecureDataSourceInline:
		if ds.InsecureInline == nil {
			return nil, errors.New("insecureInline must be defined")
		}
		return &system_proto.DataSource{Type: &system_proto.DataSource_InlineString{InlineString: ds.InsecureInline.Value}}, nil
	default:
		return nil, fmt.Errorf("datasource type: %s is not supported on MeshExternalService", *ds.Type)
	}
}

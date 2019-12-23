package xds

import (
	"strconv"

	"github.com/Kong/kuma/pkg/core"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

var metadataLog = core.Log.WithName("xds-server").WithName("metadata-tracker")

const (
	// Supported Envoy node metadata fields.

	fieldDataplaneTokenPath = "dataplaneTokenPath"
	fieldDataplaneAdminPort = "dataplane.admin.port"
)

// DataplaneMetadata represents environment-specific part of a dataplane configuration.
//
// This information might change from one dataplane run to another
// and therefore it cannot be a part of Dataplane resource.
//
// On start-up, a dataplane captures its effective configuration (that might come
// from a file, environment variables and command line options) and includes it
// into request for a bootstrap config.
// Control Plane can use this information to fill in node metadata in the bootstrap
// config.
// Envoy will include node metadata from the bootstrap config
// at least into the very first discovery request on every xDS stream.
// This way, xDS server will be able to use Envoy node metadata
// to generate xDS resources that depend on environment-specific configuration.
type DataplaneMetadata struct {
	DataplaneTokenPath string
	AdminPort          uint32
}

func (m *DataplaneMetadata) GetDataplaneTokenPath() string {
	if m == nil {
		return ""
	}
	return m.DataplaneTokenPath
}

func (m *DataplaneMetadata) GetAdminPort() uint32 {
	if m == nil {
		return 0
	}
	return m.AdminPort
}

func DataplaneMetadataFromNode(node *envoy_core.Node) *DataplaneMetadata {
	metadata := DataplaneMetadata{}
	if node.Metadata == nil {
		return &metadata
	}
	if field := node.Metadata.Fields[fieldDataplaneTokenPath]; field != nil {
		metadata.DataplaneTokenPath = field.GetStringValue()
	}
	if value := node.Metadata.Fields[fieldDataplaneAdminPort]; value != nil {
		if port, err := strconv.Atoi(value.GetStringValue()); err == nil {
			metadata.AdminPort = uint32(port)
		} else {
			metadataLog.Error(err, "invalid value in dataplane metadata", "field", fieldDataplaneAdminPort, "value", value)
		}
	}
	return &metadata
}

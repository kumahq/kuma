package xds

import (
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var metadataLog = core.Log.WithName("xds-server").WithName("metadata-tracker")

const (
	// Supported Envoy node metadata fields.
	FieldDataplaneAdminPort            = "dataplane.admin.port"
	FieldDataplaneAdminAddress         = "dataplane.admin.address"
	FieldDataplaneReadinessPort        = "dataplane.readinessReporter.port"
	FieldDataplaneAppProbeProxyEnabled = "dataplane.appProbeProxy.enabled"
	FieldDataplaneDNSPort              = "dataplane.dns.port"
	FieldDataplaneDataplaneResource    = "dataplane.resource"
	FieldDynamicMetadata               = "dynamicMetadata"
	FieldDataplaneProxyType            = "dataplane.proxyType"
	FieldVersion                       = "version"
	FieldPrefixDependenciesVersion     = "version.dependencies"
	FieldFeatures                      = "features"
	FieldWorkdir                       = "workdir"
	FieldMetricsCertPath               = "metricsCertPath"
	FieldMetricsKeyPath                = "metricsKeyPath"
	FieldSystemCaPath                  = "systemCaPath"
<<<<<<< HEAD
=======
	FieldTransparentProxy              = "transparentProxy"
	FieldIPv6Enabled                   = "ipv6Enabled"
>>>>>>> fa3eb620b (fix(kuma-cp): configure Envoy internal addresses based on dp IPv6 support (#14652))
)

// DataplaneMetadata represents environment-specific part of a dataplane configuration.
//
// This information might change from one dataplane run to another,
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
	Resource             model.Resource
	AdminPort            uint32
	AdminAddress         string
	ReadinessPort        uint32
	AppProbeProxyEnabled bool
	DNSPort              uint32
	DynamicMetadata      map[string]string
	ProxyType            mesh_proto.ProxyType
	Version              *mesh_proto.Version
	Features             Features
	WorkDir              string
	MetricsCertPath      string
	MetricsKeyPath       string
	SystemCaPath         string
<<<<<<< HEAD
=======
	TransparentProxy     *tproxy_dp.DataplaneConfig
	IPv6Enabled          bool
>>>>>>> fa3eb620b (fix(kuma-cp): configure Envoy internal addresses based on dp IPv6 support (#14652))
}

// GetDataplaneResource returns the underlying DataplaneResource, if present.
// If the resource is of a different type, it returns nil.
func (m *DataplaneMetadata) GetDataplaneResource() *core_mesh.DataplaneResource {
	if m != nil {
		if d, ok := m.Resource.(*core_mesh.DataplaneResource); ok {
			return d
		}
	}

	return nil
}

// GetZoneIngressResource returns the underlying ZoneIngressResource, if present.
// If the resource is of a different type, it returns nil.
func (m *DataplaneMetadata) GetZoneIngressResource() *core_mesh.ZoneIngressResource {
	if m != nil {
		if z, ok := m.Resource.(*core_mesh.ZoneIngressResource); ok {
			return z
		}
	}

	return nil
}

// GetZoneEgressResource returns the underlying ZoneEgressResource, if present.
// If the resource is of a different type, it returns nil.
func (m *DataplaneMetadata) GetZoneEgressResource() *core_mesh.ZoneEgressResource {
	if m != nil {
		if z, ok := m.Resource.(*core_mesh.ZoneEgressResource); ok {
			return z
		}
	}

	return nil
}

func (m *DataplaneMetadata) GetProxyType() mesh_proto.ProxyType {
	if m == nil || m.ProxyType == "" {
		return mesh_proto.DataplaneProxyType
	}
	return m.ProxyType
}

func (m *DataplaneMetadata) GetAdminPort() uint32 {
	if m == nil {
		return 0
	}
	return m.AdminPort
}

func (m *DataplaneMetadata) GetReadinessPort() uint32 {
	if m == nil {
		return 0
	}
	return m.ReadinessPort
}

func (m *DataplaneMetadata) GetAppProbeProxyEnabled() bool {
	if m == nil {
		return false
	}
	return m.AppProbeProxyEnabled
}

func (m *DataplaneMetadata) GetAdminAddress() string {
	if m == nil {
		return ""
	}
	return m.AdminAddress
}

func (m *DataplaneMetadata) GetDNSPort() uint32 {
	if m == nil {
		return 0
	}
	return m.DNSPort
}

func (m *DataplaneMetadata) GetDynamicMetadata(key string) string {
	if m == nil || m.DynamicMetadata == nil {
		return ""
	}
	return m.DynamicMetadata[key]
}

func (m *DataplaneMetadata) GetVersion() *mesh_proto.Version {
	if m == nil {
		return nil
	}
	return m.Version
}

<<<<<<< HEAD
=======
func (m *DataplaneMetadata) GetTransparentProxy() *tproxy_dp.DataplaneConfig {
	if m == nil {
		return nil
	}
	return m.TransparentProxy
}

func (m *DataplaneMetadata) GetIPv6Enabled() bool {
	if m == nil {
		return false
	}
	return m.IPv6Enabled
}

func (m *DataplaneMetadata) HasFeature(feature string) bool {
	if m == nil || m.Features == nil {
		return false
	}
	return m.Features.HasFeature(feature)
}

>>>>>>> fa3eb620b (fix(kuma-cp): configure Envoy internal addresses based on dp IPv6 support (#14652))
func DataplaneMetadataFromXdsMetadata(xdsMetadata *structpb.Struct) *DataplaneMetadata {
	// Be extra careful here about nil checks since xdsMetadata is a "user" input.
	// Even if we know that something should not be nil since we are generating metadata,
	// the DiscoveryRequest can still be crafted manually to crash the CP.
	metadata := DataplaneMetadata{}
	if xdsMetadata == nil {
		return &metadata
	}
	if field := xdsMetadata.Fields[FieldDataplaneProxyType]; field != nil {
		metadata.ProxyType = mesh_proto.ProxyType(field.GetStringValue())
	}
	metadata.AdminPort = uint32Metadata(xdsMetadata, FieldDataplaneAdminPort)
	metadata.AdminAddress = xdsMetadata.Fields[FieldDataplaneAdminAddress].GetStringValue()
	metadata.ReadinessPort = uint32Metadata(xdsMetadata, FieldDataplaneReadinessPort)
	metadata.AppProbeProxyEnabled = boolMetadata(xdsMetadata, FieldDataplaneAppProbeProxyEnabled)
	metadata.DNSPort = uint32Metadata(xdsMetadata, FieldDataplaneDNSPort)
	if value := xdsMetadata.Fields[FieldDataplaneDataplaneResource]; value != nil {
		res, err := rest.YAML.UnmarshalCore([]byte(value.GetStringValue()))
		if err != nil {
			metadataLog.Error(err, "invalid value in dataplane metadata", "field", FieldDataplaneDataplaneResource, "value", value)
		} else {
			switch r := res.(type) {
			case *core_mesh.DataplaneResource,
				*core_mesh.ZoneIngressResource,
				*core_mesh.ZoneEgressResource:
				metadata.Resource = r
			default:
				metadataLog.Error(err, "invalid dataplane resource type",
					"resource", r.Descriptor().Name,
					"field", FieldDataplaneDataplaneResource,
					"value", value)
			}
		}
	}

	metadata.WorkDir = xdsMetadata.Fields[FieldWorkdir].GetStringValue()

	if xdsMetadata.Fields[FieldMetricsCertPath] != nil {
		metadata.MetricsCertPath = xdsMetadata.Fields[FieldMetricsCertPath].GetStringValue()
	}
	if xdsMetadata.Fields[FieldMetricsKeyPath] != nil {
		metadata.MetricsKeyPath = xdsMetadata.Fields[FieldMetricsKeyPath].GetStringValue()
	}
	if xdsMetadata.Fields[FieldSystemCaPath] != nil {
		metadata.SystemCaPath = xdsMetadata.Fields[FieldSystemCaPath].GetStringValue()
	}

	if listValue := xdsMetadata.Fields[FieldFeatures]; listValue != nil {
		metadata.Features = Features{}
		for _, feature := range listValue.GetListValue().GetValues() {
			metadata.Features[feature.GetStringValue()] = true
		}
	}

	if value := xdsMetadata.Fields[FieldVersion]; value.GetStructValue() != nil {
		version := &mesh_proto.Version{}
		if err := util_proto.ToTyped(value.GetStructValue(), version); err != nil {
			metadataLog.Error(err, "invalid value in dataplane metadata", "field", FieldVersion, "value", value)
		}
		version.KumaDp.KumaCpCompatible = kuma_version.DeploymentVersionCompatible(kuma_version.Build.Version, version.KumaDp.GetVersion())
		metadata.Version = version
	}

<<<<<<< HEAD
=======
	if v := xdsMetadata.Fields[FieldTransparentProxy]; v.GetStructValue() != nil {
		metadata.TransparentProxy = util_proto.MustFromMapOfAny[*tproxy_dp.DataplaneConfig](v.GetStructValue())
	}

	if xdsMetadata.Fields[FieldIPv6Enabled] != nil {
		metadata.IPv6Enabled = xdsMetadata.Fields[FieldIPv6Enabled].GetBoolValue()
	} else {
		// For backward compatibility as previously this was always enabled
		metadata.IPv6Enabled = true
	}

>>>>>>> fa3eb620b (fix(kuma-cp): configure Envoy internal addresses based on dp IPv6 support (#14652))
	if value := xdsMetadata.Fields[FieldDynamicMetadata]; value != nil {
		dynamicMetadata := map[string]string{}
		for field, val := range value.GetStructValue().GetFields() {
			if strings.HasPrefix(field, FieldPrefixDependenciesVersion) {
				dependencyName := strings.TrimPrefix(field, FieldPrefixDependenciesVersion+".")
				if metadata.GetVersion().GetDependencies() != nil {
					metadata.Version.Dependencies[dependencyName] = val.GetStringValue()
				}
			} else {
				dynamicMetadata[field] = val.GetStringValue()
			}
		}
		metadata.DynamicMetadata = dynamicMetadata
	}

	return &metadata
}

func uint32Metadata(xdsMetadata *structpb.Struct, field string) uint32 {
	value := xdsMetadata.Fields[field]
	if value == nil {
		return 0
	}
	port, err := strconv.ParseInt(value.GetStringValue(), 10, 32)
	if err != nil {
		metadataLog.Error(err, "invalid value in dataplane metadata", "field", field, "value", value)
		return 0
	}
	return uint32(port)
}

func boolMetadata(xdsMetadata *structpb.Struct, field string) bool {
	value := xdsMetadata.Fields[field]
	if value == nil {
		return false
	}
	return value.GetStringValue() == "true"
}

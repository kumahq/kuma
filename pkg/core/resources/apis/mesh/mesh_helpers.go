package mesh

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/ca/provided/config"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *MeshResource) HasPrometheusMetricsEnabled() bool {
	return m != nil && m.GetEnabledMetricsBackend().GetType() == mesh_proto.MetricsPrometheusType
}

func (m *MeshResource) GetEnabledMetricsBackend() *mesh_proto.MetricsBackend {
	return m.GetMetricsBackend(m.Spec.GetMetrics().GetEnabledBackend())
}

func (m *MeshResource) GetMetricsBackend(name string) *mesh_proto.MetricsBackend {
	for _, backend := range m.Spec.GetMetrics().GetBackends() {
		if backend.Name == name {
			return backend
		}
	}
	return nil
}

func (m *MeshResource) MTLSEnabled() bool {
	return m != nil && m.Spec.GetMtls().GetEnabledBackend() != ""
}

// ZoneEgress works only when mTLS is enabled.
// Configuration of mTLS is validated on Mesh configuration
// change and when zoneEgress is enabled.
func (m *MeshResource) ZoneEgressEnabled() bool {
	return m != nil && m.Spec.GetRouting().GetZoneEgress()
}

func (m *MeshResource) LocalityAwareLbEnabled() bool {
	return m != nil && m.Spec.GetRouting().GetLocalityAwareLoadBalancing()
}

func (m *MeshResource) GetLoggingBackend(name string) *mesh_proto.LoggingBackend {
	backends := map[string]*mesh_proto.LoggingBackend{}
	for _, backend := range m.Spec.GetLogging().GetBackends() {
		backends[backend.Name] = backend
	}
	if name == "" {
		return backends[m.Spec.GetLogging().GetDefaultBackend()]
	}
	return backends[name]
}

func (m *MeshResource) GetTracingBackend(name string) *mesh_proto.TracingBackend {
	backends := map[string]*mesh_proto.TracingBackend{}
	for _, backend := range m.Spec.GetTracing().GetBackends() {
		backends[backend.Name] = backend
	}
	if name == "" {
		return backends[m.Spec.GetTracing().GetDefaultBackend()]
	}
	return backends[name]
}

// GetLoggingBackends will return logging backends as comma separated strings
// if empty return empty string
func (m *MeshResource) GetLoggingBackends() string {
	var backends []string
	for _, backend := range m.Spec.GetLogging().GetBackends() {
		backend := fmt.Sprintf("%s/%s", backend.GetType(), backend.GetName())
		backends = append(backends, backend)
	}
	return strings.Join(backends, ", ")
}

// GetTracingBackends will return tracing backends as comma separated strings
// if empty return empty string
func (m *MeshResource) GetTracingBackends() string {
	var backends []string
	for _, backend := range m.Spec.GetTracing().GetBackends() {
		backend := fmt.Sprintf("%s/%s", backend.GetType(), backend.GetName())
		backends = append(backends, backend)
	}
	return strings.Join(backends, ", ")
}

func (m *MeshResource) GetEnabledCertificateAuthorityBackend() *mesh_proto.CertificateAuthorityBackend {
	return m.GetCertificateAuthorityBackend(m.Spec.GetMtls().GetEnabledBackend())
}

func (m *MeshResource) GetCertificateAuthorityBackend(name string) *mesh_proto.CertificateAuthorityBackend {
	for _, backend := range m.Spec.GetMtls().GetBackends() {
		if backend.Name == name {
			return backend
		}
	}
	return nil
}

var durationRE = regexp.MustCompile(`^(\d+)(y|w|d|h|m|s|ms)$`)

// ParseDuration parses a string into a time.Duration
func ParseDuration(durationStr string) (time.Duration, error) {
	// Allow 0 without a unit.
	if durationStr == "0" {
		return 0, nil
	}
	matches := durationRE.FindStringSubmatch(durationStr)
	if len(matches) != 3 {
		return 0, fmt.Errorf("not a valid duration string: %q", durationStr)
	}
	var (
		n, _ = strconv.Atoi(matches[1])
		dur  = time.Duration(n) * time.Millisecond
	)
	switch unit := matches[2]; unit {
	case "y":
		dur *= 1000 * 60 * 60 * 24 * 365
	case "w":
		dur *= 1000 * 60 * 60 * 24 * 7
	case "d":
		dur *= 1000 * 60 * 60 * 24
	case "h":
		dur *= 1000 * 60 * 60
	case "m":
		dur *= 1000 * 60
	case "s":
		dur *= 1000
	case "ms":
		// Value already correct
	default:
		return 0, fmt.Errorf("invalid time unit in duration string: %q", unit)
	}
	return dur, nil
}

func (ml *MeshResourceList) MarshalLog() interface{} {
	maskedList := make([]*MeshResource, 0, len(ml.Items))
	for _, mesh := range ml.Items {
		maskedList = append(maskedList, mesh.MarshalLog().(*MeshResource))
	}
	return MeshResourceList{
		Items:      maskedList,
		Pagination: ml.Pagination,
	}
}

func (m *MeshResource) MarshalLog() interface{} {
	spec := proto.Clone(m.Spec).(*mesh_proto.Mesh)
	if spec == nil {
		return m
	}
	mtls := spec.Mtls
	if mtls == nil {
		return m
	}
	for _, backend := range mtls.Backends {
		conf := backend.Conf
		if conf == nil {
			continue
		}
		cfg := &config.ProvidedCertificateAuthorityConfig{}
		err := util_proto.ToTyped(conf, cfg)
		if err != nil {
			continue
		}
		cfg.Key = cfg.Key.MaskInlineDatasource()
		cfg.Cert = cfg.Cert.MaskInlineDatasource()
		backend.Conf, err = util_proto.ToStruct(cfg)
		if err != nil {
			continue
		}
	}
	return &MeshResource{
		Meta: m.Meta,
		Spec: spec,
	}
}

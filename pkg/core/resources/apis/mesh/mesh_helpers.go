package mesh

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/ca/provided/config"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
)

func (r *MeshResource) HasPrometheusMetricsEnabled() bool {
	return r != nil && r.GetEnabledMetricsBackend().GetType() == mesh_proto.MetricsPrometheusType
}

func (r *MeshResource) GetEnabledMetricsBackend() *mesh_proto.MetricsBackend {
	return r.GetMetricsBackend(r.Spec.GetMetrics().GetEnabledBackend())
}

func (r *MeshResource) GetMetricsBackend(name string) *mesh_proto.MetricsBackend {
	for _, backend := range r.Spec.GetMetrics().GetBackends() {
		if backend.Name == name {
			return backend
		}
	}
	return nil
}

func (r *MeshResource) MTLSEnabled() bool {
	return r != nil && r.Spec.GetMtls().GetEnabledBackend() != ""
}

// ZoneEgress works only when mTLS is enabled.
// Configuration of mTLS is validated on Mesh configuration
// change and when zoneEgress is enabled.
func (r *MeshResource) ZoneEgressEnabled() bool {
	return r != nil && r.Spec.GetRouting().GetZoneEgress()
}

func (r *MeshResource) LocalityAwareLbEnabled() bool {
	return r != nil && r.Spec.GetRouting().GetLocalityAwareLoadBalancing()
}

func (r *MeshResource) GetLoggingBackend(name string) *mesh_proto.LoggingBackend {
	backends := map[string]*mesh_proto.LoggingBackend{}
	for _, backend := range r.Spec.GetLogging().GetBackends() {
		backends[backend.Name] = backend
	}
	if name == "" {
		return backends[r.Spec.GetLogging().GetDefaultBackend()]
	}
	return backends[name]
}

func (r *MeshResource) GetTracingBackend(name string) *mesh_proto.TracingBackend {
	backends := map[string]*mesh_proto.TracingBackend{}
	for _, backend := range r.Spec.GetTracing().GetBackends() {
		backends[backend.Name] = backend
	}
	if name == "" {
		return backends[r.Spec.GetTracing().GetDefaultBackend()]
	}
	return backends[name]
}

// GetLoggingBackends will return logging backends as comma separated strings
// if empty return empty string
func (r *MeshResource) GetLoggingBackends() string {
	var backends []string
	for _, backend := range r.Spec.GetLogging().GetBackends() {
		backend := fmt.Sprintf("%s/%s", backend.GetType(), backend.GetName())
		backends = append(backends, backend)
	}
	return strings.Join(backends, ", ")
}

// GetTracingBackends will return tracing backends as comma separated strings
// if empty return empty string
func (r *MeshResource) GetTracingBackends() string {
	var backends []string
	for _, backend := range r.Spec.GetTracing().GetBackends() {
		backend := fmt.Sprintf("%s/%s", backend.GetType(), backend.GetName())
		backends = append(backends, backend)
	}
	return strings.Join(backends, ", ")
}

func (r *MeshResource) GetEnabledCertificateAuthorityBackend() *mesh_proto.CertificateAuthorityBackend {
	return r.GetCertificateAuthorityBackend(r.Spec.GetMtls().GetEnabledBackend())
}

func (r *MeshResource) GetCertificateAuthorityBackend(name string) *mesh_proto.CertificateAuthorityBackend {
	for _, backend := range r.Spec.GetMtls().GetBackends() {
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

func (ml *MeshResourceList) MarshalLog() any {
	maskedList := make([]*MeshResource, 0, len(ml.Items))
	for _, mesh := range ml.Items {
		maskedList = append(maskedList, mesh.MarshalLog().(*MeshResource))
	}
	return MeshResourceList{
		Items:      maskedList,
		Pagination: ml.Pagination,
	}
}

func (r *MeshResource) MarshalLog() any {
	spec := proto.Clone(r.Spec).(*mesh_proto.Mesh)
	if spec == nil {
		return r
	}
	mtls := spec.Mtls
	if mtls == nil {
		return r
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
		Meta: r.Meta,
		Spec: spec,
	}
}

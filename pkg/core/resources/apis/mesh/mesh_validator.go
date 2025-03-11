package mesh

import (
	"fmt"
	"net"
	"net/url"
	"regexp"

	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var AllowedMTLSBackends = 1

func (m *MeshResource) Validate() error {
	var verr validators.ValidationError
	verr.AddError("mtls", validateMtls(m.Spec.Mtls))
	verr.AddError("logging", validateLogging(m.Spec.Logging))
	verr.AddError("tracing", validateTracing(m.Spec.Tracing))
	verr.AddError("metrics", validateMetrics(m.Spec.Metrics))
	verr.AddError("constraints", validateConstraints(m.Spec.Constraints))
	verr.AddError("", validateZoneEgress(m.Spec.Routing, m.Spec.Mtls))
	return verr.OrNil()
}

func validateConstraints(constraints *mesh_proto.Mesh_Constraints) validators.ValidationError {
	var verr validators.ValidationError
	if constraints == nil {
		return verr
	}
	verr.AddError("dataplaneProxy", validateDppConstraints(constraints.DataplaneProxy))
	return verr
}

func validateDppConstraints(constraints *mesh_proto.Mesh_DataplaneProxyConstraints) validators.ValidationError {
	var verr validators.ValidationError
	if constraints == nil {
		return verr
	}

	for i, requirement := range constraints.GetRequirements() {
		verr.Add(ValidateSelector(
			validators.RootedAt("requirements").Index(i).Field("tags"),
			requirement.Tags,
			ValidateTagsOpts{RequireAtLeastOneTag: true},
		))
	}

	for i, requirement := range constraints.GetRestrictions() {
		verr.Add(ValidateSelector(
			validators.RootedAt("restrictions").Index(i).Field("tags"),
			requirement.Tags,
			ValidateTagsOpts{RequireAtLeastOneTag: true},
		))
	}

	return verr
}

func validateMtls(mtls *mesh_proto.Mesh_Mtls) validators.ValidationError {
	var verr validators.ValidationError
	if mtls == nil {
		return verr
	}
	if len(mtls.GetBackends()) > AllowedMTLSBackends {
		verr.AddViolationAt(validators.RootedAt("backends"), fmt.Sprintf("cannot have more than %d backends", AllowedMTLSBackends))
	}

	usedNames := map[string]bool{}
	for i, backend := range mtls.GetBackends() {
		if usedNames[backend.Name] {
			verr.AddViolationAt(validators.RootedAt("backends").Index(i).Field("name"), fmt.Sprintf("%q name is already used for another backend", backend.Name))
		}
		usedNames[backend.Name] = true
	}
	if mtls.GetEnabledBackend() != "" && !usedNames[mtls.GetEnabledBackend()] {
		verr.AddViolation("enabledBackend", "has to be set to one of the backends in the mesh")
	}
	for _, backend := range mtls.Backends {
		if backend.GetDpCert() != nil {
			_, err := ParseDuration(backend.GetDpCert().GetRotation().GetExpiration())
			if err != nil {
				verr.AddViolation("dpcert.rotation.expiration", "has to be a valid format")
			}
		}
	}
	return verr
}

func validateLogging(logging *mesh_proto.Logging) validators.ValidationError {
	var verr validators.ValidationError
	if logging == nil {
		return verr
	}
	usedNames := map[string]bool{}
	for i, backend := range logging.Backends {
		verr.AddError(validators.RootedAt("backends").Index(i).String(), validateLoggingBackend(backend))
		if usedNames[backend.Name] {
			verr.AddViolationAt(validators.RootedAt("backends").Index(i).Field("name"), fmt.Sprintf("%q name is already used for another backend", backend.Name))
		}
		usedNames[backend.Name] = true
	}
	if logging.DefaultBackend != "" && !usedNames[logging.DefaultBackend] {
		verr.AddViolation("defaultBackend", "has to be set to one of the logging backend in mesh")
	}
	return verr
}

func validateLoggingBackend(backend *mesh_proto.LoggingBackend) validators.ValidationError {
	var verr validators.ValidationError
	if backend.Name == "" {
		verr.AddViolation("name", "cannot be empty")
	}
	switch backend.GetType() {
	case mesh_proto.LoggingFileType:
		verr.AddError("config", validateLoggingFile(backend.Conf))
	case mesh_proto.LoggingTcpType:
		verr.AddError("config", validateLoggingTcp(backend.Conf))
	default:
		verr.AddViolation("type", fmt.Sprintf("unknown backend type. Available backends: %q, %q", mesh_proto.LoggingTcpType, mesh_proto.LoggingFileType))
	}
	return verr
}

func validateLoggingTcp(cfgStr *structpb.Struct) validators.ValidationError {
	var verr validators.ValidationError
	cfg := mesh_proto.TcpLoggingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		verr.AddViolation("", fmt.Sprintf("could not parse config: %s", err.Error()))
		return verr
	}
	if cfg.Address == "" {
		verr.AddViolation("address", "cannot be empty")
		return verr
	}
	host, port, err := net.SplitHostPort(cfg.Address)
	if host == "" || port == "" || err != nil {
		verr.AddViolation("address", "has to be in format of HOST:PORT")
	}
	return verr
}

func validateLoggingFile(cfgStr *structpb.Struct) validators.ValidationError {
	var verr validators.ValidationError
	cfg := mesh_proto.FileLoggingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		verr.AddViolation("", fmt.Sprintf("could not parse config: %s", err.Error()))
	} else if cfg.Path == "" {
		verr.AddViolation("path", "cannot be empty")
	}
	return verr
}

func validateTracing(tracing *mesh_proto.Tracing) validators.ValidationError {
	var verr validators.ValidationError
	if tracing == nil {
		return verr
	}
	usedNames := map[string]bool{}
	for i, backend := range tracing.Backends {
		verr.AddError(validators.RootedAt("backends").Index(i).String(), validateTracingBackend(backend))
		if usedNames[backend.Name] {
			verr.AddViolationAt(validators.RootedAt("backends").Index(i).Field("name"), fmt.Sprintf("%q name is already used for another backend", backend.Name))
		}
		usedNames[backend.Name] = true
	}
	if tracing.DefaultBackend != "" && !usedNames[tracing.DefaultBackend] {
		verr.AddViolation("defaultBackend", "has to be set to one of the tracing backend in mesh")
	}
	return verr
}

func validateTracingBackend(backend *mesh_proto.TracingBackend) validators.ValidationError {
	var verr validators.ValidationError
	if backend.Name == "" {
		verr.AddViolation("name", "cannot be empty")
	}
	if backend.Sampling.GetValue() < 0.0 || backend.Sampling.GetValue() > 100.0 {
		verr.AddViolation("sampling", "has to be in [0.0 - 100.0] range")
	}
	switch backend.GetType() {
	case mesh_proto.TracingZipkinType:
		verr.AddError("config", validateZipkin(backend.Conf))
	case mesh_proto.TracingDatadogType:
		verr.AddError("config", validateDatadog(backend.Conf))
	default:
		verr.AddViolation("type", fmt.Sprintf("unknown backend type. Available backends: %q, %q", mesh_proto.TracingZipkinType, mesh_proto.TracingDatadogType))
	}
	return verr
}

func validateDatadog(cfgStr *structpb.Struct) validators.ValidationError {
	var verr validators.ValidationError
	cfg := mesh_proto.DatadogTracingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		verr.AddViolation("", fmt.Sprintf("could not parse config: %s", err.Error()))
		return verr
	}

	if cfg.Address == "" {
		verr.AddViolation("address", "cannot be empty")
	}

	verr.Add(ValidatePort(validators.RootedAt("port"), cfg.GetPort()))
	return verr
}

func validateZipkin(cfgStr *structpb.Struct) validators.ValidationError {
	var verr validators.ValidationError
	cfg := mesh_proto.ZipkinTracingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		verr.AddViolation("", fmt.Sprintf("could not parse config: %s", err.Error()))
		return verr
	}
	if cfg.ApiVersion != "" && cfg.ApiVersion != "httpJsonV1" && cfg.ApiVersion != "httpJson" && cfg.ApiVersion != "httpProto" {
		verr.AddViolation("apiVersion", fmt.Sprintf(`has invalid value. %s`, AllowedValuesHint("httpJsonV1", "httpJson", "httpProto")))
	}
	if cfg.Url == "" {
		verr.AddViolation("url", "cannot be empty")
		return verr
	}
	uri, err := url.ParseRequestURI(cfg.Url)
	if err != nil {
		verr.AddViolation("url", "invalid URL")
	} else if uri.Port() == "" {
		verr.AddViolation("url", "port has to be explicitly specified")
	}
	return verr
}

func validateMetrics(metrics *mesh_proto.Metrics) validators.ValidationError {
	var verr validators.ValidationError
	if metrics == nil {
		return verr
	}
	usedNames := map[string]bool{}
	for i, backend := range metrics.GetBackends() {
		if usedNames[backend.Name] {
			verr.AddViolationAt(validators.RootedAt("backends").Index(i).Field("name"), fmt.Sprintf("%q name is already used for another backend", backend.Name))
		}
		if backend.GetType() != mesh_proto.MetricsPrometheusType {
			verr.AddViolationAt(validators.RootedAt("backends").Index(i).Field("type"), fmt.Sprintf("unknown backend type. Available backends: %q", mesh_proto.MetricsPrometheusType))
		} else {
			verr.AddErrorAt(validators.RootedAt("backends").Index(i).Field("conf"), validatePrometheusConfig(backend.GetConf()))
		}
		usedNames[backend.Name] = true
	}
	if metrics.GetEnabledBackend() != "" && !usedNames[metrics.GetEnabledBackend()] {
		verr.AddViolation("enabledBackend", "has to be set to one of the backends in the mesh")
	}
	return verr
}

func validatePrometheusConfig(cfgStr *structpb.Struct) validators.ValidationError {
	var verr validators.ValidationError
	cfg := mesh_proto.PrometheusMetricsBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		verr.AddViolation("", fmt.Sprintf("could not parse config: %s", err.Error()))
		return verr
	}
	if cfg.SkipMTLS != nil && cfg.Tls != nil && !hasEqualConfiguration(&cfg) {
		verr.AddViolation("", "skipMTLS and tls configuration cannot be defined together")
		return verr
	}
	if _, err := regexp.Compile(cfg.GetEnvoy().GetFilterRegex()); err != nil {
		verr.AddViolationAt(validators.RootedAt("envoy").Field("filterRegex"), fmt.Sprintf("provided regexp isn't correct: %s", err.Error()))
		return verr
	}
	usedName := make(map[string]bool)
	for i, config := range cfg.GetAggregate() {
		if _, ok := usedName[config.GetName()]; ok {
			verr.AddViolationAt(validators.RootedAt("aggregate").Index(i).Field("name"), fmt.Sprintf("duplicate entry: %s, values have to be unique", config.GetName()))
			continue
		}
		usedName[config.GetName()] = true
	}
	return verr
}

func validateZoneEgress(routing *mesh_proto.Routing, mtls *mesh_proto.Mesh_Mtls) validators.ValidationError {
	var verr validators.ValidationError
	if routing == nil {
		return verr
	}
	if routing.ZoneEgress {
		if mtls.GetEnabledBackend() == "" {
			verr.AddViolation("mtls", "has to be set when zoneEgress enabled")
		}
	}
	return verr
}

func hasEqualConfiguration(cfg *mesh_proto.PrometheusMetricsBackendConfig) bool {
	return (!cfg.SkipMTLS.GetValue() && cfg.Tls.GetMode() == mesh_proto.PrometheusTlsConfig_activeMTLSBackend) ||
		(cfg.SkipMTLS.GetValue() && cfg.Tls.GetMode() == mesh_proto.PrometheusTlsConfig_disabled)
}

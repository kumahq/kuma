package mesh

import (
	"fmt"
	"net"
	"net/url"

	structpb "github.com/golang/protobuf/ptypes/struct"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/envoy/accesslog"
	"github.com/Kong/kuma/pkg/util/proto"
)

func (m *MeshResource) Validate() error {
	var verr validators.ValidationError
	verr.AddError("mtls", validateMtls(m.Spec.Mtls))
	verr.AddError("logging", validateLogging(m.Spec.Logging))
	verr.AddError("tracing", validateTracing(m.Spec.Tracing))
	return verr.OrNil()
}

func validateMtls(mtls *mesh_proto.Mesh_Mtls) validators.ValidationError {
	var verr validators.ValidationError
	if mtls == nil {
		return verr
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
	if err := accesslog.ValidateFormat(backend.Format); err != nil {
		verr.AddViolation("format", err.Error())
	}
	switch backend.GetType() {
	case mesh_proto.LoggingFileType:
		verr.AddError("config", validateLoggingFile(backend.Config))
	case mesh_proto.LoggingTcpType:
		verr.AddError("config", validateLoggingTcp(backend.Config))
	}
	return verr
}

func validateLoggingTcp(cfgStr *structpb.Struct) validators.ValidationError {
	var verr validators.ValidationError
	cfg := mesh_proto.TcpLoggingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		verr.AddViolation("", fmt.Sprintf("could not parse config: %s", err.Error()))
	} else {
		if cfg.Address == "" {
			verr.AddViolation("address", "cannot be empty")
		} else {
			host, port, err := net.SplitHostPort(cfg.Address)
			if host == "" || port == "" || err != nil {
				verr.AddViolation("address", "has to be in format of HOST:PORT")
			}
		}
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
	if zipkin, ok := backend.GetType().(*mesh_proto.TracingBackend_Zipkin_); ok {
		verr.AddError("zipkin", validateZipkin(zipkin.Zipkin))
	}
	return verr
}

func validateZipkin(zipkin *mesh_proto.TracingBackend_Zipkin) validators.ValidationError {
	var verr validators.ValidationError
	if zipkin.Url == "" {
		verr.AddViolation("url", "cannot be empty")
	} else {
		uri, err := url.ParseRequestURI(zipkin.Url)
		if err != nil {
			verr.AddViolation("url", "invalid URL")
		} else if uri.Port() == "" {
			verr.AddViolation("url", "port has to be explicitly specified")
		}
	}
	if zipkin.ApiVersion != "" && zipkin.ApiVersion != "httpJsonV1" && zipkin.ApiVersion != "httpJson" && zipkin.ApiVersion != "httpProto" {
		verr.AddViolation("apiVersion", fmt.Sprintf(`has invalid value. %s`, AllowedValuesHint("httpJsonV1", "httpJson", "httpProto")))
	}
	return verr
}

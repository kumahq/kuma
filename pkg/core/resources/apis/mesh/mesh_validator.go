package mesh

import (
	"fmt"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
	"net"
)

func (m *MeshResource) Validate() error {
	var verr validators.ValidationError
	verr.AddError("mtls", validateMtls(m.Spec.Mtls))
	verr.AddError("logging", validateLogging(m.Spec.Logging))
	return verr.OrNil()
}

func validateMtls(mtls *mesh_proto.Mesh_Mtls) validators.ValidationError {
	var verr validators.ValidationError
	if mtls == nil {
		return verr
	}
	if mtls.Enabled && mtls.Ca == nil {
		verr.AddViolation("ca", "has to be set when mTLS is enabled")
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
		verr.AddError(validators.RootedAt("backends").Index(i).String(), validateBackend(backend))
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

func validateBackend(backend *mesh_proto.LoggingBackend) validators.ValidationError {
	var verr validators.ValidationError
	if backend.Name == "" {
		verr.AddViolation("name", "cannot be empty")
	}
	if file, ok := backend.GetType().(*mesh_proto.LoggingBackend_File_); ok {
		verr.AddError("file", validateLoggingFile(file))
	} else if tcp, ok := backend.GetType().(*mesh_proto.LoggingBackend_Tcp_); ok {
		verr.AddError("tcp", validateLoggingTcp(tcp))
	}
	return verr
}

func validateLoggingTcp(tcp *mesh_proto.LoggingBackend_Tcp_) validators.ValidationError {
	var verr validators.ValidationError
	if tcp.Tcp.Address == "" {
		verr.AddViolation("address", "cannot be empty")
	} else {
		host, port, err := net.SplitHostPort(tcp.Tcp.Address)
		if host == "" || port == "" || err != nil {
			verr.AddViolation("address", "has to be in format of HOST:PORT")
		}
	}
	return verr
}

func validateLoggingFile(file *mesh_proto.LoggingBackend_File_) validators.ValidationError {
	var veer validators.ValidationError
	if file.File.Path == "" {
		veer.AddViolation("path", "cannot be empty")
	}
	return veer
}

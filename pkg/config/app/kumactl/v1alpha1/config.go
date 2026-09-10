package v1alpha1

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// mustBeObject keeps a scalar or a list where an object belongs from reporting the
// anonymous struct the proto spelled aliases are decoded through.
func mustBeObject(data []byte, name string) error {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return errors.Errorf("%s must be an object", name)
	}
	return nil
}

// Configuration defines configuration of `kumactl`.
//
// The field names below are the ones jsonpb wrote, so a file written by an older
// kumactl round trips unchanged. Every type whose name spans more than one word also
// accepts the proto field name, because jsonpb read both spellings and configs written
// by hand use the underscored one.
type Configuration struct {
	// List of known Control Planes.
	ControlPlanes []*ControlPlane `json:"controlPlanes,omitempty"`
	// List of configured `kumactl` contexts.
	Contexts []*Context `json:"contexts,omitempty"`
	// Name of the context to use by default.
	CurrentContext string `json:"currentContext,omitempty"`
}

func (c *Configuration) UnmarshalJSON(data []byte) error {
	if err := mustBeObject(data, "configuration"); err != nil {
		return err
	}
	type alias Configuration
	aux := struct {
		*alias
		ProtoControlPlanes  []*ControlPlane `json:"control_planes"`
		ProtoCurrentContext *string         `json:"current_context"`
	}{alias: (*alias)(c)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if c.ControlPlanes == nil {
		c.ControlPlanes = aux.ProtoControlPlanes
	}
	if c.CurrentContext == "" && aux.ProtoCurrentContext != nil {
		c.CurrentContext = *aux.ProtoCurrentContext
	}
	return nil
}

func (c *Configuration) GetControlPlanes() []*ControlPlane {
	if c == nil {
		return nil
	}
	return c.ControlPlanes
}

func (c *Configuration) GetContexts() []*Context {
	if c == nil {
		return nil
	}
	return c.Contexts
}

func (c *Configuration) GetCurrentContext() string {
	if c == nil {
		return ""
	}
	return c.CurrentContext
}

// ControlPlane defines a Control Plane.
type ControlPlane struct {
	// Name defines a reference name for a given Control Plane.
	Name string `json:"name,omitempty"`
	// Coordinates defines coordinates of a given Control Plane.
	Coordinates *ControlPlaneCoordinates `json:"coordinates,omitempty"`
}

func (c *ControlPlane) GetName() string {
	if c == nil {
		return ""
	}
	return c.Name
}

func (c *ControlPlane) GetCoordinates() *ControlPlaneCoordinates {
	if c == nil {
		return nil
	}
	return c.Coordinates
}

// ControlPlaneCoordinates defines coordinates of a Control Plane.
type ControlPlaneCoordinates struct {
	ApiServer *ControlPlaneCoordinates_ApiServer `json:"apiServer,omitempty"`
}

func (c *ControlPlaneCoordinates) UnmarshalJSON(data []byte) error {
	if err := mustBeObject(data, "coordinates"); err != nil {
		return err
	}
	type alias ControlPlaneCoordinates
	aux := struct {
		*alias
		ProtoApiServer *ControlPlaneCoordinates_ApiServer `json:"api_server"`
	}{alias: (*alias)(c)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if c.ApiServer == nil {
		c.ApiServer = aux.ProtoApiServer
	}
	return nil
}

func (c *ControlPlaneCoordinates) GetApiServer() *ControlPlaneCoordinates_ApiServer {
	if c == nil {
		return nil
	}
	return c.ApiServer
}

type ControlPlaneCoordinates_Headers struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (h *ControlPlaneCoordinates_Headers) GetKey() string {
	if h == nil {
		return ""
	}
	return h.Key
}

func (h *ControlPlaneCoordinates_Headers) GetValue() string {
	if h == nil {
		return ""
	}
	return h.Value
}

type ControlPlaneCoordinates_ApiServer struct {
	// URL defines URL of the Control Plane API Server.
	Url string `json:"url,omitempty"`
	// CaCert defines the certificate authority which will be used to verify
	// connection to the control plane API server
	CaCertFile string `json:"caCertFile,omitempty"`
	// ClientCert defines the certificate of the authorized client of the
	// control plane API server
	ClientCertFile string `json:"clientCertFile,omitempty"`
	// ClientKey defines the key of the authorized client of the control plane
	// API server
	ClientKeyFile string `json:"clientKeyFile,omitempty"`
	// Headers to be added for communication with Kuma control plane
	Headers []*ControlPlaneCoordinates_Headers `json:"headers,omitempty"`
	// Authentication type
	AuthType string `json:"authType,omitempty"`
	// Authentication configuration for defined authentication type
	AuthConf map[string]string `json:"authConf,omitempty"`
	// SkipVerify disables verification of the Control Plane API Server's TLS certificate.
	SkipVerify bool `json:"skipVerify,omitempty"`
}

func (s *ControlPlaneCoordinates_ApiServer) UnmarshalJSON(data []byte) error {
	if err := mustBeObject(data, "apiServer"); err != nil {
		return err
	}
	type alias ControlPlaneCoordinates_ApiServer
	aux := struct {
		*alias
		ProtoCaCertFile     *string           `json:"ca_cert_file"`
		ProtoClientCertFile *string           `json:"client_cert_file"`
		ProtoClientKeyFile  *string           `json:"client_key_file"`
		ProtoAuthType       *string           `json:"auth_type"`
		ProtoAuthConf       map[string]string `json:"auth_conf"`
		ProtoSkipVerify     *bool             `json:"skip_verify"`
	}{alias: (*alias)(s)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if s.CaCertFile == "" && aux.ProtoCaCertFile != nil {
		s.CaCertFile = *aux.ProtoCaCertFile
	}
	if s.ClientCertFile == "" && aux.ProtoClientCertFile != nil {
		s.ClientCertFile = *aux.ProtoClientCertFile
	}
	if s.ClientKeyFile == "" && aux.ProtoClientKeyFile != nil {
		s.ClientKeyFile = *aux.ProtoClientKeyFile
	}
	if s.AuthType == "" && aux.ProtoAuthType != nil {
		s.AuthType = *aux.ProtoAuthType
	}
	if s.AuthConf == nil {
		s.AuthConf = aux.ProtoAuthConf
	}
	if !s.SkipVerify && aux.ProtoSkipVerify != nil {
		s.SkipVerify = *aux.ProtoSkipVerify
	}
	return nil
}

func (s *ControlPlaneCoordinates_ApiServer) GetUrl() string {
	if s == nil {
		return ""
	}
	return s.Url
}

func (s *ControlPlaneCoordinates_ApiServer) GetCaCertFile() string {
	if s == nil {
		return ""
	}
	return s.CaCertFile
}

func (s *ControlPlaneCoordinates_ApiServer) GetClientCertFile() string {
	if s == nil {
		return ""
	}
	return s.ClientCertFile
}

func (s *ControlPlaneCoordinates_ApiServer) GetClientKeyFile() string {
	if s == nil {
		return ""
	}
	return s.ClientKeyFile
}

func (s *ControlPlaneCoordinates_ApiServer) GetHeaders() []*ControlPlaneCoordinates_Headers {
	if s == nil {
		return nil
	}
	return s.Headers
}

func (s *ControlPlaneCoordinates_ApiServer) GetAuthType() string {
	if s == nil {
		return ""
	}
	return s.AuthType
}

func (s *ControlPlaneCoordinates_ApiServer) GetAuthConf() map[string]string {
	if s == nil {
		return nil
	}
	return s.AuthConf
}

func (s *ControlPlaneCoordinates_ApiServer) GetSkipVerify() bool {
	if s == nil {
		return false
	}
	return s.SkipVerify
}

// Context defines a context in which individual `kumactl` commands run.
type Context struct {
	// Name defines a reference name for a given context.
	Name string `json:"name,omitempty"`
	// ControlPlane defines a reference to a known Control Plane.
	ControlPlane string `json:"controlPlane,omitempty"`
	// Defaults defines default settings for a given context.
	Defaults *Context_Defaults `json:"defaults,omitempty"`
}

func (c *Context) UnmarshalJSON(data []byte) error {
	if err := mustBeObject(data, "context"); err != nil {
		return err
	}
	type alias Context
	aux := struct {
		*alias
		ProtoControlPlane *string `json:"control_plane"`
	}{alias: (*alias)(c)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if c.ControlPlane == "" && aux.ProtoControlPlane != nil {
		c.ControlPlane = *aux.ProtoControlPlane
	}
	return nil
}

func (c *Context) GetName() string {
	if c == nil {
		return ""
	}
	return c.Name
}

func (c *Context) GetControlPlane() string {
	if c == nil {
		return ""
	}
	return c.ControlPlane
}

func (c *Context) GetDefaults() *Context_Defaults {
	if c == nil {
		return nil
	}
	return c.Defaults
}

// Context_Defaults defines default settings for a context.
type Context_Defaults struct {
	// Mesh defines a Mesh to use in requests if one is not provided explicitly.
	Mesh string `json:"mesh,omitempty"`
}

func (d *Context_Defaults) GetMesh() string {
	if d == nil {
		return ""
	}
	return d.Mesh
}

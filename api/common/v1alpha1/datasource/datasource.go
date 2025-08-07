// +kubebuilder:object:generate=true
package datasource

import "github.com/kumahq/kuma/pkg/core/validators"

// +kubebuilder:validation:Enum=File;Secret;EnvVar;InsecureInline
type SecureDataSourceType string

const (
	SecureDataSourceFile      SecureDataSourceType = "File"
	SecureDataSourceSecretRef SecureDataSourceType = "Secret"
	SecureDataSourceEnvVar    SecureDataSourceType = "EnvVar"
	SecureDataSourceInline    SecureDataSourceType = "InsecureInline"
)

// SecureDataSource is a way to securely provide data to the component
type SecureDataSource struct {
	// +kuma:discriminator
	Type           SecureDataSourceType `json:"type"`
	File           *File                `json:"file,omitempty"`
	EnvVar         *EnvVar              `json:"envVar,omitempty"`
	InsecureInline *Inline              `json:"insecureInline,omitempty"`
	SecretRef      *SecretRef           `json:"secretRef,omitempty"`
}

// +kubebuilder:validation:Enum=File;EnvVar;Inline
type DataSourceType string

const (
	DataSourceFile   DataSourceType = "File"
	DataSourceEnvVar DataSourceType = "EnvVar"
	DataSourceInline DataSourceType = "Inline"
)

func (s *SecureDataSource) ValidateSecureDataSource(path validators.PathBuilder) validators.ValidationError {
	var verr validators.ValidationError
	if s == nil {
		return verr
	}
	switch s.Type {
	case SecureDataSourceEnvVar:
		if s.EnvVar == nil {
			verr.AddViolationAt(path.Field("envVar"), validators.MustBeDefined)
		} else if s.EnvVar.Name == "" {
			verr.AddViolationAt(path.Field("envVar").Field("name"), validators.MustBeDefined)
		}
	case SecureDataSourceInline:
		if s.InsecureInline == nil {
			verr.AddViolationAt(path.Field("insecureInline"), validators.MustBeDefined)
		} else if s.InsecureInline.Value == "" {
			verr.AddViolationAt(path.Field("insecureInline").Field("value"), validators.MustBeDefined)
		}
	case SecureDataSourceFile:
		if s.File == nil {
			verr.AddViolationAt(path.Field("file"), validators.MustBeDefined)
		} else if s.File.Path == "" {
			verr.AddViolationAt(path.Field("file").Field("path"), validators.MustBeDefined)
		}
	case SecureDataSourceSecretRef:
		if s.SecretRef == nil {
			verr.AddViolationAt(path.Field("secretRef"), validators.MustBeDefined)
		}
		if s.SecretRef != nil {
			if s.SecretRef.Kind != SecretRefType {
				verr.AddViolationAt(path.Field("secretRef").Field("kind"), validators.MustBeOneOf(string(s.SecretRef.Kind), string(SecretRefType)))
			}
			if s.SecretRef.Name == "" {
				verr.AddViolationAt(path.Field("secretRef").Field("name"), validators.MustBeDefined)
			}
		}
	default:
		verr.AddViolationAt(path.Field("type"), validators.MustBeOneOf(string(s.Type), string(SecureDataSourceEnvVar), string(SecureDataSourceInline), string(SecureDataSourceFile), string(SecureDataSourceSecretRef)))
	}
	return verr
}

// DataSource is just a way to provide data. Not necessarily secrets,
// can be any data, i.e. certs, configs, OPA policies written in rego, lua plugins etc.
type DataSource struct {
	// +kuma:discriminator
	Type   DataSourceType `json:"type"`
	File   *File          `json:"file,omitempty"`
	EnvVar *EnvVar        `json:"envVar,omitempty"`
	Inline *Inline        `json:"inline,omitempty"`
}

type File struct {
	Path string `json:"path"`
}

type EnvVar struct {
	Name string `json:"name"`
}

// +kubebuilder:validation:Enum=Secret
type RefType string

const (
	SecretRefType RefType = "Secret"
)

type SecretRef struct {
	Kind RefType `json:"kind"`
	Name string  `json:"name"`
}

type Inline struct {
	Value string `json:"value"`
}

// +kubebuilder:object:generate=true
package datasource

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
	Type           SecureDataSourceType `json:"type,omitempty"`
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

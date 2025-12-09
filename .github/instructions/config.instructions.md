---
applyTo:
  - "pkg/config/**"
---

# Config Development Guidelines

## Structure

```
pkg/config/
├── config.go, loader.go, display.go, deprecate.go  # Core
├── app/{kuma-cp,kuma-dp,kumactl}/                  # Applications
├── core/resources/store/                           # Store backends
├── {api-server,xds,dp-server,multizone}/           # Components
└── types/                                          # Duration, TLS
```

**Lifecycle:** Create → Load → PostProcess → Validate → Sanitize → Use

## Core Patterns

### Config Interface (`pkg/config/config.go:10`)

```go
type Config interface {
    Sanitize()           // Redact sensitive values
    Validate() error     // Validate config values/dependencies
    PostProcess() error  // Apply env var overrides, transformations
}

type BaseConfig struct{}  // Embed to avoid boilerplate
func (c BaseConfig) Sanitize()          {}
func (c BaseConfig) PostProcess() error { return nil }
func (c BaseConfig) Validate() error    { return nil }
```

### Loader Pattern (`pkg/config/loader.go:20`)

```go
config.NewLoader(cfg).
    WithStrictParsing().                    // Reject unknown fields
    WithEnvVarsLoading("KUMA_").            // Env var prefix
    WithValidation().                       // Call Validate() after load
    Load(stdin, "config.yaml", "override.yaml")  // Last wins
```

Features: YAML/stdin loading • Strict/permissive parsing • Env var override (`envconfig`) • Validation

### Hierarchical Nested Configs (`pkg/config/app/kuma-cp/config.go:50`)

```go
type Config struct {
    config.BaseConfig `json:"-"`
    General    *GeneralConfig              `json:"general" envconfig:"kuma_general"`
    Store      *store.StoreConfig          `json:"store" envconfig:"kuma_store"`
    XdsServer  *xds.XdsServerConfig        `json:"xdsServer" envconfig:"kuma_xds_server"`
}

func (c *Config) Validate() error {
    var errs error
    if err := c.General.Validate(); err != nil {
        errs = multierr.Append(errs, errors.Wrap(err, ".General validation failed"))
    }
    if err := c.Store.Validate(); err != nil {
        errs = multierr.Append(errs, errors.Wrap(err, ".Store validation failed"))
    }
    return errs
}
```

Rules: Parent calls child methods • Wrap errors with field path (`.General`) • Use `multierr.Append()` • Never panic

### Default Factory (`pkg/config/app/kuma-cp/kuma-cp.defaults.go:10`)

```go
func DefaultConfig() *Config {
    return &Config{
        Environment: core.UniversalEnvironment,
        Mode:        core.Zone,
        General:     DefaultGeneralConfig(),
        Store:       store.DefaultStoreConfig(),
        XdsServer:   xds.DefaultXdsServerConfig(),
    }
}

func DefaultXdsServerConfig() *XdsServerConfig {
    return &XdsServerConfig{
        GrpcPort:                              5678,
        DataplaneConfigurationRefreshInterval: config_types.Duration{Duration: 1 * time.Second},
        DataplaneStatusFlushInterval:          config_types.Duration{Duration: 10 * time.Second},
        NACKBackoff:                           config_types.Duration{Duration: 5 * time.Second},
    }
}
```

Rules: Hardcode production defaults • Paths empty (filled in `PostProcess`) • Every struct has `Default*Config()`

### Tags (`pkg/config/xds/config.go:15`)

```go
type DataplaneMetrics struct {
    SubscriptionLimit int                      `json:"subscriptionLimit" envconfig:"kuma_metrics_dataplane_subscription_limit"`
    IdleTimeout       config_types.Duration    `json:"idleTimeout" envconfig:"kuma_metrics_dataplane_idle_timeout"`
}
```

- `json`: YAML/JSON marshaling (camelCase)
- `envconfig`: `kuma_<component>_<field>` (snake_case, lowercase), processed by `envconfig`

### Validation (`pkg/config/app/kuma-cp/config.go:200`)

```go
func (c *Config) Validate() error {
    var errs error
    switch c.Mode {
    case core.Zone:
        if err := c.XdsServer.Validate(); err != nil {
            errs = multierr.Append(errs, errors.Wrap(err, "Zone mode requires valid XdsServer"))
        }
        if c.Multizone.Zone.GlobalAddress == "" {
            errs = multierr.Append(errs, errors.New("Zone mode requires .Multizone.Zone.GlobalAddress"))
        }
    case core.Global:
        if err := c.Multizone.Global.Kds.Validate(); err != nil {
            errs = multierr.Append(errs, errors.Wrap(err, "Global mode requires valid KDS"))
        }
    }
    // Mutually exclusive fields
    if (c.TlsKeyFile == "" && c.TlsCertFile != "") ||
       (c.TlsKeyFile != "" && c.TlsCertFile == "") {
        errs = multierr.Append(errs, errors.New("Both TlsCertFile and TlsKeyFile must be specified together"))
    }
    return errs
}
```

Common checks: Port range (`> 65535`) • Positive duration • CIDR (`net.ParseCIDR`) • TLS (`config_types.TLSVersion/Ciphers`) • Mutually exclusive • Conditional deps

### PostProcess (`pkg/config/app/kuma-dp/config.go:150`)

```go
func (d *Dataplane) PostProcess() error {
    if d.Name == "" {
        if podName, ok := os.LookupEnv("POD_NAME"); ok {
            d.Name = podName + "." + os.Getenv("POD_NAMESPACE")
        }
    }
    if err := envconfig.Process("KUMA", d); err != nil {
        return errors.Wrap(err, "failed to process env vars")
    }
    return nil
}
```

Use: Env var override • Auto-fill from OS • Path expansion • Normalize values

### Sanitize (`pkg/config/display.go:10`)

```go
func (c *Config) Sanitize() {
    c.General.Sanitize()    // TLS keys, CA certs
    c.Store.Sanitize()      // DB passwords
    c.ApiServer.Sanitize()  // OAuth secrets
}

func (a *ApiServerConfig) Sanitize() {
    if a.Auth != nil {
        a.Auth.ClientCertsDir = "[redacted]"
    }
    if a.Authn != nil && a.Authn.Tokens != nil {
        for i := range a.Authn.Tokens.BootstrapToken {
            a.Authn.Tokens.BootstrapToken[i] = "[redacted]"
        }
    }
}
```

Rules: Replace sensitive with `[redacted]` • Called before logging • Deep copy first (JSON marshal/unmarshal)

## Custom Types

### Duration (`pkg/config/types/duration.go:10`)

Parses `"1s"`, `"5m"`, `"1h30m"` and numeric seconds (`60` → `1m`) • Converts to/from `time.Duration` • UnmarshalJSON/YAML

```go
type TimeoutConfig struct {
    ConnectionTimeout config_types.Duration `json:"connectionTimeout"`
    IdleTimeout       config_types.Duration `json:"idleTimeout"`
}
```

### TLS (`pkg/config/types/tls.go:10`)

- `TLSVersion(string) (uint16, error)` - Validates TLS 1.0-1.3
- `TLSCiphers([]string) ([]uint16, error)` - Validates cipher suites

```go
tlsVersion, err := config_types.TLSVersion(c.TlsMinVersion)
ciphers, err := config_types.TLSCiphers(c.TlsCipherSuites)
```

## Environment & Mode (`pkg/config/core/config.go`)

```go
// Environment
const (
    KubernetesEnvironment Environment = "kubernetes"  // plugins/runtime/k8s
    UniversalEnvironment  Environment = "universal"   // plugins/runtime/universal
)

// Mode
const (
    Standalone CpMode = "standalone"  // Single-zone
    Global     CpMode = "global"      // Multi-zone coordinator
    Zone       CpMode = "zone"        // Multi-zone member
)

switch cfg.Mode {
case core.Zone:      // Requires XdsServer, GlobalAddress
case core.Global:    // Requires KDS server
case core.Standalone:// No multizone deps
}
```

## Component Configs

| Component | Location | Key Fields |
|-----------|----------|------------|
| **Store** | `pkg/config/core/resources/store/config.go:10` | `Type` (kubernetes/postgres/pgx/memory) • `Kubernetes.SystemNamespace` • `Postgres.{Host,Port,User,Password,Database}` • `Cache.{Enabled,ExpirationTime}` • `Upsert.ConflictRetry` |
| **XDS Server** | `pkg/config/xds/config.go:10` | `GrpcPort` (5678) • `DataplaneConfigurationRefreshInterval` (1s) • `DataplaneStatusFlushInterval` (10s) • `NACKBackoff` (5s) • `DataplaneDeregistrationDelay` (10s) |
| **API Server** | `pkg/config/api-server/config.go:10` | `HTTP.{Interface,Port}` • `HTTPS.{TlsCertFile,TlsKeyFile,TlsMinVersion}` • `Auth.ClientCertsDir` • `Authn.{Type,Tokens}` (types: tokens/clientCerts/none) • `CORS` • `ReadOnly` • `GUI` |
| **DP Server** | `pkg/config/dp-server/config.go:10` | `Authn.DpProxy.{Type,DpToken,ZoneToken}` • `Authn.ZoneProxy` • `Hds.{Enabled,Interval,RefreshInterval}` • `TlsCertFile/TlsKeyFile` • Auth types: dpToken/serviceAccountToken/zoneToken/none |
| **Multizone KDS** | `pkg/config/multizone/kds.go:10` | **Global:** `GrpcPort` (5685) • `RefreshInterval` (1s) • `TlsEnabled/TlsCertFile/TlsKeyFile` • `MaxMsgSize` (10MB) • `ZoneHealthCheck`<br>**Zone:** `RefreshInterval` • `TlsSkipVerify/RootCAFile` • `ReconnectDelay` |

## Deprecation (`pkg/config/deprecate.go:10`)

```go
type Deprecation struct {
    Env             string
    EnvMsg          string
    ConfigValuePath func(cfg Config) (string, bool)
    ConfigValueMsg  string
}

deprecations := []Deprecation{
    {Env: "KUMA_METRICS_MESH_MIN_RESYNC_TIMEOUT", EnvMsg: "Use KUMA_METRICS_MESH_MIN_RESYNC_INTERVAL"},
}
config.PrintDeprecations(deprecations, cfg, stdout)
```

Track deprecated env vars/fields • Print warnings on startup • Migration guidance • Remove after 2-3 releases

## Testing

```go
// Loading
It("should load config from YAML", func() {
    cfg := kuma_cp.DefaultConfig()
    yaml := `mode: zone\nstore:\n  type: postgres`
    err := config.NewLoader(cfg).WithValidation().LoadBytes([]byte(yaml))
    Expect(err).ToNot(HaveOccurred())
    Expect(cfg.Mode).To(Equal(core.Zone))
})

// Env vars
It("should override from env vars", func() {
    cfg := kuma_cp.DefaultConfig()
    os.Setenv("KUMA_MODE", "global")
    defer os.Unsetenv("KUMA_MODE")
    err := cfg.PostProcess()
    Expect(cfg.Mode).To(Equal(core.Global))
})

// Validation (table-driven)
DescribeTable("should validate",
    func(given *XdsServerConfig, expectedErr string) {
        err := given.Validate()
        if expectedErr == "" {
            Expect(err).ToNot(HaveOccurred())
        } else {
            Expect(err).To(MatchError(ContainSubstring(expectedErr)))
        }
    },
    Entry("valid", &XdsServerConfig{GrpcPort: 5678}, ""),
    Entry("invalid port", &XdsServerConfig{GrpcPort: 99999}, "Port must be in range"),
)

// Sanitization
It("should sanitize sensitive values", func() {
    cfg := kuma_cp.DefaultConfig()
    cfg.Store.Postgres.Password = "secret123"
    cfg.Sanitize()
    Expect(cfg.Store.Postgres.Password).To(Equal("[redacted]"))
})
```

## Common Mistakes

### ❌ NEVER

```go
// Empty validation
func (c *Config) Validate() error { return nil }

// Panic in config
func (c *Config) Validate() error { panic("Port must be set") }

// Direct envconfig.Process() (should be in PostProcess)
func LoadConfig() (*Config, error) {
    cfg := DefaultConfig()
    envconfig.Process("KUMA", cfg)  // ❌
}

// No defaults
func NewConfig() *Config { return &Config{} }

// Missing envconfig tag
type ServerConfig struct {
    Port int `json:"port"`  // ❌ Can't override via env
}

// No error wrapping
func (c *Config) Validate() error {
    if err := c.Store.Validate(); err != nil {
        return err  // ❌ Which field?
    }
}
```

### ✅ ALWAYS

```go
// Complete validation
func (c *Config) Validate() error {
    var errs error
    if err := c.Store.Validate(); err != nil {
        errs = multierr.Append(errs, errors.Wrap(err, ".Store validation failed"))
    }
    if c.Port > 65535 {
        errs = multierr.Append(errs, errors.New(".Port must be in range [0, 65535]"))
    }
    return errs
}

// Proper PostProcess
func (c *Config) PostProcess() error {
    if err := c.Store.PostProcess(); err != nil {
        return err
    }
    if err := envconfig.Process("KUMA", c); err != nil {
        return errors.Wrap(err, "failed to process env vars")
    }
    return nil
}

// Complete sanitization
func (c *Config) Sanitize() {
    c.Store.Sanitize()
    c.ApiServer.Sanitize()
    if c.PrivateKey != "" {
        c.PrivateKey = "[redacted]"
    }
}

// Default factory with all fields
func DefaultXdsServerConfig() *XdsServerConfig {
    return &XdsServerConfig{
        GrpcPort:                              5678,
        DataplaneConfigurationRefreshInterval: config_types.Duration{Duration: 1 * time.Second},
        DataplaneStatusFlushInterval:          config_types.Duration{Duration: 10 * time.Second},
        NACKBackoff:                           config_types.Duration{Duration: 5 * time.Second},
    }
}

// Both tags
type ServerConfig struct {
    Port    int                   `json:"port" envconfig:"kuma_server_port"`
    Timeout config_types.Duration `json:"timeout" envconfig:"kuma_server_timeout"`
}
```

## Adding New Config

```go
// 1. Create struct
type MyComponentConfig struct {
    config.BaseConfig `json:"-"`
    Enabled bool                   `json:"enabled" envconfig:"kuma_my_component_enabled"`
    Port    int                    `json:"port" envconfig:"kuma_my_component_port"`
    Timeout config_types.Duration  `json:"timeout" envconfig:"kuma_my_component_timeout"`
}

// 2. Default factory
func DefaultMyComponentConfig() *MyComponentConfig {
    return &MyComponentConfig{Enabled: true, Port: 8080, Timeout: config_types.Duration{Duration: 30 * time.Second}}
}

// 3. Validate
func (c *MyComponentConfig) Validate() error {
    var errs error
    if c.Port > 65535 {
        errs = multierr.Append(errs, errors.New(".Port must be in range [0, 65535]"))
    }
    if c.Timeout.Duration <= 0 {
        errs = multierr.Append(errs, errors.New(".Timeout must be positive"))
    }
    return errs
}

// 4. Sanitize (if needed)
func (c *MyComponentConfig) Sanitize() {
    if c.ApiKey != "" { c.ApiKey = "[redacted]" }
}

// 5. Add to parent
type Config struct {
    MyComponent *MyComponentConfig `json:"myComponent" envconfig:"kuma_my_component"`
}

func DefaultConfig() *Config {
    return &Config{MyComponent: DefaultMyComponentConfig()}
}

func (c *Config) Validate() error {
    var errs error
    if err := c.MyComponent.Validate(); err != nil {
        errs = multierr.Append(errs, errors.Wrap(err, ".MyComponent validation failed"))
    }
    return errs
}

// 6. Write tests (validation table-driven, env var override, sanitization)
```

## Review Checklist

- **Validation:** All fields • Port ranges (0-65535) • Positive durations • TLS cert/key pairs • Conditional deps • Wrap errors • `multierr`
- **Defaults:** All fields initialized • Production defaults • Consistent
- **Env vars:** All fields have `envconfig` tags • `kuma_<component>_<field>` • Processed in `PostProcess()`
- **Sanitization:** Sensitive fields redacted • Called before logging
- **Testing:** Validation (table-driven) • Env var override • Sanitization • YAML loading
- **Structure:** Embed `BaseConfig` • Hierarchical nesting • JSON tags • Consistent

## Quick Reference

| Type | Location |
|------|----------|
| Config, Loader | `pkg/config/config.go:10`, `pkg/config/loader.go:20` |
| kuma-cp, kuma-dp | `pkg/config/app/{kuma-cp,kuma-dp}/config.go` |
| Store, XDS, API/DP Server, KDS | `pkg/config/{core/resources/store,xds,api-server,dp-server,multizone}/config.go` |
| Duration, TLS | `pkg/config/types/{duration,tls}.go:10` |
| Display, Deprecate | `pkg/config/{display,deprecate}.go:10` |

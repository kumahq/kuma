---
applyTo:
  - "pkg/config/**"
---

# Configuration Development Guidelines

## Architecture Overview

**Key Directories:**
```
pkg/config/
├── config.go            # Config interface, BaseConfig
├── loader.go            # Fluent config loader (YAML, env)
├── display.go           # Safe config serialization
├── deprecate.go         # Deprecation tracking
├── app/                 # Application configs
│   ├── kuma-cp/         # Control plane master config
│   ├── kuma-dp/         # Data plane proxy config
│   └── kumactl/         # CLI config
├── core/                # Core types, environment modes
│   └── resources/store/ # Store backends (k8s, postgres)
├── api-server/          # REST API server
├── xds/                 # Envoy XDS server
├── dp-server/           # Dataplane gRPC server
├── multizone/           # KDS sync (global/zone)
├── plugins/             # Plugin configs
│   ├── resources/       # Store plugins
│   ├── runtime/         # K8s/Universal runtime
│   └── policies/        # Policy plugin config
└── types/               # Custom types (Duration, TLS)
```

**Config Lifecycle:**
```
Create → LoadFile/LoadBytes → PostProcess → Validate → Sanitize → Use
   ↓           ↓                  ↓             ↓          ↓
Defaults    Parse YAML      Env vars      Check      Redact logs
```

## Core Patterns

### Config Interface

**Location:** `pkg/config/config.go:10`

**Interface:**
```go
type Config interface {
    Sanitize()           // Redact sensitive values (passwords, tokens, certs)
    Validate() error     // Validate config values and dependencies
    PostProcess() error  // Apply env var overrides, transformations
}
```

**BaseConfig Struct:**
```go
type BaseConfig struct{}

func (c BaseConfig) Sanitize()          {}
func (c BaseConfig) PostProcess() error { return nil }
func (c BaseConfig) Validate() error    { return nil }
```

**Usage:** Embed `BaseConfig` in component configs to avoid boilerplate

### Loader Pattern (Fluent Builder)

**Location:** `pkg/config/loader.go:20`

**Interface:**
```go
loader := config.NewLoader(cfg).
    WithStrictParsing().                    // Reject unknown fields
    WithEnvVarsLoading("KUMA_").            // Load env vars with prefix
    WithValidation().                       // Call Validate() after load
    Load(stdin, "config.yaml", "override.yaml")  // Load in order
```

**Features:**
- YAML file loading (multiple files, last wins)
- stdin reading (`-` or empty string)
- Strict vs. permissive parsing (unknown field rejection)
- Environment variable override via `kelseyhightower/envconfig`
- Optional validation after loading

### Hierarchical Nested Configs

**Location:** `pkg/config/app/kuma-cp/config.go:50`

**Pattern:**
```go
type Config struct {
    config.BaseConfig `json:"-"`           // Embedded base

    General    *GeneralConfig              `json:"general" envconfig:"kuma_general"`
    Store      *store.StoreConfig          `json:"store" envconfig:"kuma_store"`
    XdsServer  *xds.XdsServerConfig        `json:"xdsServer" envconfig:"kuma_xds_server"`
    ApiServer  *api_server.ApiServerConfig `json:"apiServer" envconfig:"kuma_api_server"`
}

func (c *Config) Validate() error {
    var errs error

    if err := c.General.Validate(); err != nil {
        errs = multierr.Append(errs, errors.Wrap(err, ".General validation failed"))
    }
    if err := c.Store.Validate(); err != nil {
        errs = multierr.Append(errs, errors.Wrap(err, ".Store validation failed"))
    }

    return errs  // Returns nil if no errors
}
```

**Rules:**
- Parent calls child `Validate()`, `PostProcess()`, `Sanitize()`
- Wrap errors with field path: `.General`, `.Store.Postgres`
- Accumulate errors with `multierr.Append()`
- Never panic in validation

### Default Factory Functions

**Location:** `pkg/config/app/kuma-cp/kuma-cp.defaults.go:10`

**Pattern:**
```go
func DefaultConfig() *Config {
    return &Config{
        Environment: core.UniversalEnvironment,
        Mode:        core.Zone,
        General:     DefaultGeneralConfig(),
        Store:       store.DefaultStoreConfig(),
        XdsServer:   xds.DefaultXdsServerConfig(),
        // ... all fields initialized
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

**Rules:**
- All server ports, timeouts, intervals hardcoded
- Sensible defaults for production
- Paths often empty (auto-filled in `PostProcess`)
- Every config struct has `Default*Config()` factory

### JSON Tags with Environment Variable Support

**Location:** `pkg/config/xds/config.go:15`

**Pattern:**
```go
type DataplaneMetrics struct {
    SubscriptionLimit int                      `json:"subscriptionLimit" envconfig:"kuma_metrics_dataplane_subscription_limit"`
    IdleTimeout       config_types.Duration    `json:"idleTimeout" envconfig:"kuma_metrics_dataplane_idle_timeout"`
}
```

**Tag Requirements:**
- `json` tag: YAML/JSON marshaling (lowercase or camelCase)
- `envconfig` tag: Environment variable name for override
  - Format: `kuma_<component>_<field>` (snake_case, all lowercase)
  - Nested: `kuma_store_postgres_connection_string`
- Processed by `github.com/kelseyhightower/envconfig`

### Validation with Dependencies

**Location:** `pkg/config/app/kuma-cp/config.go:200`

**Pattern:**
```go
func (c *Config) Validate() error {
    var errs error

    // Conditional validation based on mode
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

**Common Validations:**
- Port range: `if port > 65535 { ... }`
- Positive duration: `if duration.Duration <= 0 { ... }`
- CIDR: `net.ParseCIDR(cidr)`
- TLS version/ciphers: `config_types.TLSVersion()`, `config_types.TLSCiphers()`
- Mutually exclusive fields
- Conditional dependencies (mode, environment)

### PostProcess for Transformations

**Location:** `pkg/config/app/kuma-dp/config.go:150`

**Pattern:**
```go
func (d *Dataplane) PostProcess() error {
    // Auto-fill from environment if not set
    if d.Name == "" {
        podName, ok := os.LookupEnv("POD_NAME")
        if ok {
            d.Name = podName + "." + os.Getenv("POD_NAMESPACE")
        }
    }

    // Apply envconfig overrides
    if err := envconfig.Process("KUMA", d); err != nil {
        return errors.Wrap(err, "failed to process env vars")
    }

    return nil
}
```

**Use Cases:**
- Environment variable override (via `envconfig.Process()`)
- Auto-fill from OS environment
- Path expansion (`~` → home directory)
- Normalize values (trim whitespace, lowercase)

### Sanitization for Display

**Location:** `pkg/config/display.go:10`

**Pattern:**
```go
func (c *Config) Sanitize() {
    c.General.Sanitize()    // Redact TLS keys, CA certs
    c.Store.Sanitize()      // Redact DB passwords
    c.ApiServer.Sanitize()  // Redact OAuth secrets
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

**Rules:**
- Replace sensitive values with `[redacted]`
- Called before logging/display (`config.Display()`)
- Sensitive: passwords, tokens, private keys, certs, connection strings
- Deep copy via JSON marshal/unmarshal before sanitizing

## Custom Types

### Duration

**Location:** `pkg/config/types/duration.go:10`

**Features:**
- Parses `"1s"`, `"5m"`, `"1h30m"` from YAML/JSON
- Supports numeric seconds: `60` → `1m`
- Converts to/from Go's `time.Duration`
- UnmarshalJSON/UnmarshalYAML for both formats

**Usage:**
```go
type TimeoutConfig struct {
    ConnectionTimeout config_types.Duration `json:"connectionTimeout"`
    IdleTimeout       config_types.Duration `json:"idleTimeout"`
}
```

### TLS Types

**Location:** `pkg/config/types/tls.go:10`

**Features:**
- `TLSVersion(string) (uint16, error)` - Validates TLS 1.0-1.3
- `TLSCiphers([]string) ([]uint16, error)` - Validates cipher suites
- Uses Go's `crypto/tls` package constants
- Supports both secure and insecure cipher pools

**Usage:**
```go
tlsVersion, err := config_types.TLSVersion(c.TlsMinVersion)
if err != nil {
    return errors.Wrap(err, ".TlsMinVersion invalid")
}

ciphers, err := config_types.TLSCiphers(c.TlsCipherSuites)
if err != nil {
    return errors.Wrap(err, ".TlsCipherSuites invalid")
}
```

## Environment & Mode Switches

### Environment Types

**Location:** `pkg/config/core/config.go:10`

**Enum:**
```go
const (
    KubernetesEnvironment Environment = "kubernetes"
    UniversalEnvironment  Environment = "universal"
)
```

**Usage:**
```go
// K8s-specific logic
if cfg.Environment == core.KubernetesEnvironment {
    // Use plugins/runtime/k8s
}

// Universal-specific logic
if cfg.Environment == core.UniversalEnvironment {
    // Use plugins/runtime/universal
}
```

### Mode Types

**Location:** `pkg/config/core/config.go:20`

**Enum:**
```go
const (
    Standalone CpMode = "standalone"  // Single-zone, no multizone
    Global     CpMode = "global"      // Multi-zone coordinator
    Zone       CpMode = "zone"        // Multi-zone member
)
```

**Conditional Config:**
```go
switch cfg.Mode {
case core.Zone:
    // Zone requires XdsServer, GlobalAddress
case core.Global:
    // Global requires KDS server
case core.Standalone:
    // Standalone has no multizone deps
}
```

## Configuration Components

### Store Configuration

**Location:** `pkg/config/core/resources/store/config.go:10`

**Key Fields:**
- `Type` - `kubernetes`, `postgres`, `pgx`, `memory`
- `Kubernetes.SystemNamespace` - Control plane namespace
- `Postgres.{Host, Port, User, Password, Database}` - Connection params
- `Cache.{Enabled, ExpirationTime}` - In-memory cache
- `Upsert.ConflictRetry{...}` - Retry backoff for conflicts

**Environment-Specific:**
- K8s: Uses `SystemNamespace`
- Universal: Uses Postgres/Memory

### XDS Server Configuration

**Location:** `pkg/config/xds/config.go:10`

**Key Fields:**
- `GrpcPort` - Envoy xDS server port (default: 5678)
- `DataplaneConfigurationRefreshInterval` - Config regen cadence (1s)
- `DataplaneStatusFlushInterval` - Status update cadence (10s)
- `NACKBackoff` - Backoff for rejected configs (5s)
- `DataplaneDeregistrationDelay` - Delay before removing proxy (10s)

**Performance Critical:**
- Short refresh intervals increase load
- NACK backoff prevents xDS storms
- Deregistration delay prevents flapping

### API Server Configuration

**Location:** `pkg/config/api-server/config.go:10`

**Key Fields:**
- `HTTP.{Interface, Port}` - HTTP listener
- `HTTPS.{TlsCertFile, TlsKeyFile, TlsMinVersion}` - HTTPS listener
- `Auth.ClientCertsDir` - Client cert directory for mTLS
- `Authn.{Type, Tokens, Localhostlsadmin}` - Authentication
- `CORS.{AllowedDomains, AllowedHeaders}` - CORS policy
- `ReadOnly` - Disable mutating operations
- `GUI.{Enabled, BasePath}` - Web UI serving

**Auth Types:**
- `tokens` - Bearer token authentication
- `clientCerts` - Client certificate authentication
- `none` - No authentication (localhost only)

### Dataplane Server Configuration

**Location:** `pkg/config/dp-server/config.go:10`

**Key Fields:**
- `Authn.DpProxy.{Type, DpToken, ZoneToken}` - Dataplane auth
- `Authn.ZoneProxy.{Type, ZoneToken}` - Zone proxy auth
- `Hds.{Enabled, Interval, RefreshInterval}` - Health Discovery Service
- `TlsCertFile`, `TlsKeyFile` - Server TLS
- `ReadHeaderTimeout` - HTTP header read timeout

**Auth Types:**
- `dpToken` - Dataplane token from token service
- `serviceAccountToken` - Kubernetes service account token
- `zoneToken` - Zone-to-global authentication
- `none` - No authentication (dev only)

### Multizone KDS Configuration

**Location:** `pkg/config/multizone/kds.go:10`

**Global (Server):**
- `GrpcPort` - KDS server port (5685)
- `RefreshInterval` - State update cadence (1s)
- `TlsEnabled`, `TlsCertFile`, `TlsKeyFile` - TLS
- `MaxMsgSize` - Max resource batch size (10MB)
- `ZoneHealthCheck.{PollInterval, Timeout}` - Health monitoring

**Zone (Client):**
- `RefreshInterval` - Poll interval (1s)
- `TlsSkipVerify`, `RootCAFile` - TLS validation
- `ReconnectDelay` - Backoff for reconnects (1s)

## Deprecation Handling

**Location:** `pkg/config/deprecate.go:10`

**Pattern:**
```go
type Deprecation struct {
    Env             string                                    // Old env var
    EnvMsg          string                                    // Migration message
    ConfigValuePath func(cfg Config) (string, bool)          // Config path accessor
    ConfigValueMsg  string                                    // Migration message
}

deprecations := []Deprecation{
    {
        Env:    "KUMA_METRICS_MESH_MIN_RESYNC_TIMEOUT",
        EnvMsg: "Use KUMA_METRICS_MESH_MIN_RESYNC_INTERVAL instead",
    },
}

config.PrintDeprecations(deprecations, cfg, stdout)
```

**Usage:**
- Track deprecated env vars and config fields
- Print warnings on startup
- Provide migration guidance
- Remove after 2-3 releases

## Testing Patterns

### Config Loading Tests

**Location:** `pkg/config/loader/loader_test.go:10`

```go
It("should load config from YAML", func() {
    // given
    cfg := kuma_cp.DefaultConfig()
    yaml := `
      mode: zone
      store:
        type: postgres
    `

    // when
    err := config.NewLoader(cfg).
        WithValidation().
        LoadBytes([]byte(yaml))

    // then
    Expect(err).ToNot(HaveOccurred())
    Expect(cfg.Mode).To(Equal(core.Zone))
    Expect(cfg.Store.Type).To(Equal("postgres"))
})
```

### Environment Variable Tests

**Location:** `pkg/config/app/kuma-cp/config_test.go:50`

```go
It("should override from env vars", func() {
    // given
    cfg := kuma_cp.DefaultConfig()
    os.Setenv("KUMA_MODE", "global")
    os.Setenv("KUMA_STORE_TYPE", "memory")
    defer os.Unsetenv("KUMA_MODE")
    defer os.Unsetenv("KUMA_STORE_TYPE")

    // when
    err := cfg.PostProcess()

    // then
    Expect(err).ToNot(HaveOccurred())
    Expect(cfg.Mode).To(Equal(core.Global))
    Expect(cfg.Store.Type).To(Equal("memory"))
})
```

### Validation Tests

**Location:** `pkg/config/xds/config_test.go:20`

```go
DescribeTable("should validate",
    func(given *XdsServerConfig, expectedErr string) {
        // when
        err := given.Validate()

        // then
        if expectedErr == "" {
            Expect(err).ToNot(HaveOccurred())
        } else {
            Expect(err).To(MatchError(ContainSubstring(expectedErr)))
        }
    },
    Entry("valid", &XdsServerConfig{GrpcPort: 5678}, ""),
    Entry("invalid port", &XdsServerConfig{GrpcPort: 99999}, "Port must be in range"),
    Entry("negative timeout", &XdsServerConfig{NACKBackoff: config_types.Duration{Duration: -1}}, "must be positive"),
)
```

### Sanitization Tests

**Location:** `pkg/config/app/kuma-cp/config_test.go:100`

```go
It("should sanitize sensitive values", func() {
    // given
    cfg := kuma_cp.DefaultConfig()
    cfg.Store.Postgres.Password = "secret123"
    cfg.ApiServer.Authn.Tokens.BootstrapToken = []string{"token1"}

    // when
    cfg.Sanitize()

    // then
    Expect(cfg.Store.Postgres.Password).To(Equal("[redacted]"))
    Expect(cfg.ApiServer.Authn.Tokens.BootstrapToken[0]).To(Equal("[redacted]"))
})
```

## Common Mistakes

### ❌ NEVER Do

**Missing Validation:**
```go
// ❌ No validation for required fields
func (c *Config) Validate() error {
    return nil  // Always returns nil
}
```

**Panic in Config:**
```go
// ❌ Panic instead of returning error
func (c *Config) Validate() error {
    if c.Port == 0 {
        panic("Port must be set")  // DON'T PANIC
    }
}
```

**Direct envconfig.Process():**
```go
// ❌ Direct call without PostProcess pattern
func LoadConfig() (*Config, error) {
    cfg := DefaultConfig()
    envconfig.Process("KUMA", cfg)  // Should be in PostProcess()
}
```

**No Default Factory:**
```go
// ❌ No default values
func NewConfig() *Config {
    return &Config{}  // All fields zero-valued
}
```

**Missing envconfig Tags:**
```go
// ❌ No envconfig tag
type ServerConfig struct {
    Port int `json:"port"`  // Can't override via env var
}
```

**No Error Wrapping:**
```go
// ❌ No context in validation error
func (c *Config) Validate() error {
    if err := c.Store.Validate(); err != nil {
        return err  // Which field failed?
    }
}
```

### ✅ ALWAYS Do

**Complete Validation:**
```go
func (c *Config) Validate() error {
    var errs error

    // Validate all nested configs
    if err := c.Store.Validate(); err != nil {
        errs = multierr.Append(errs, errors.Wrap(err, ".Store validation failed"))
    }

    // Validate own fields
    if c.Port > 65535 {
        errs = multierr.Append(errs, errors.New(".Port must be in range [0, 65535]"))
    }

    return errs
}
```

**Proper PostProcess:**
```go
func (c *Config) PostProcess() error {
    // Process nested configs first
    if err := c.Store.PostProcess(); err != nil {
        return err
    }

    // Apply envconfig overrides
    if err := envconfig.Process("KUMA", c); err != nil {
        return errors.Wrap(err, "failed to process env vars")
    }

    return nil
}
```

**Complete Sanitization:**
```go
func (c *Config) Sanitize() {
    // Sanitize nested configs
    c.Store.Sanitize()
    c.ApiServer.Sanitize()

    // Sanitize own sensitive fields
    if c.PrivateKey != "" {
        c.PrivateKey = "[redacted]"
    }
}
```

**Default Factory with All Fields:**
```go
func DefaultXdsServerConfig() *XdsServerConfig {
    return &XdsServerConfig{
        GrpcPort:                              5678,
        DataplaneConfigurationRefreshInterval: config_types.Duration{Duration: 1 * time.Second},
        DataplaneStatusFlushInterval:          config_types.Duration{Duration: 10 * time.Second},
        NACKBackoff:                           config_types.Duration{Duration: 5 * time.Second},
        DataplaneDeregistrationDelay:          config_types.Duration{Duration: 10 * time.Second},
    }
}
```

**Both JSON and Envconfig Tags:**
```go
type ServerConfig struct {
    Port    int                   `json:"port" envconfig:"kuma_server_port"`
    Timeout config_types.Duration `json:"timeout" envconfig:"kuma_server_timeout"`
}
```

## Adding New Configuration

### Step-by-Step Process

**1. Create Config Struct:**
```go
type MyComponentConfig struct {
    config.BaseConfig `json:"-"`

    Enabled bool                   `json:"enabled" envconfig:"kuma_my_component_enabled"`
    Port    int                    `json:"port" envconfig:"kuma_my_component_port"`
    Timeout config_types.Duration  `json:"timeout" envconfig:"kuma_my_component_timeout"`
}
```

**2. Add Default Factory:**
```go
func DefaultMyComponentConfig() *MyComponentConfig {
    return &MyComponentConfig{
        Enabled: true,
        Port:    8080,
        Timeout: config_types.Duration{Duration: 30 * time.Second},
    }
}
```

**3. Implement Validate:**
```go
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
```

**4. Implement Sanitize (if needed):**
```go
func (c *MyComponentConfig) Sanitize() {
    // Only if component has sensitive fields
    if c.ApiKey != "" {
        c.ApiKey = "[redacted]"
    }
}
```

**5. Add to Parent Config:**
```go
type Config struct {
    // ... existing fields ...
    MyComponent *MyComponentConfig `json:"myComponent" envconfig:"kuma_my_component"`
}

func DefaultConfig() *Config {
    return &Config{
        // ... existing fields ...
        MyComponent: DefaultMyComponentConfig(),
    }
}

func (c *Config) Validate() error {
    var errs error
    // ... existing validations ...
    if err := c.MyComponent.Validate(); err != nil {
        errs = multierr.Append(errs, errors.Wrap(err, ".MyComponent validation failed"))
    }
    return errs
}

func (c *Config) Sanitize() {
    // ... existing sanitizations ...
    c.MyComponent.Sanitize()
}
```

**6. Write Tests:**
```go
var _ = Describe("MyComponentConfig", func() {
    Describe("Validate", func() {
        DescribeTable("should validate",
            func(given *MyComponentConfig, expectedErr string) {
                err := given.Validate()
                if expectedErr == "" {
                    Expect(err).ToNot(HaveOccurred())
                } else {
                    Expect(err).To(MatchError(ContainSubstring(expectedErr)))
                }
            },
            Entry("valid", DefaultMyComponentConfig(), ""),
            Entry("invalid port", &MyComponentConfig{Port: 99999}, "Port must be in range"),
        )
    })

    It("should load from env vars", func() {
        // Test envconfig override
    })
})
```

## Review Focus

**Validation:**
- All required fields validated
- Port ranges checked (0-65535)
- Positive durations enforced
- TLS cert/key pairs validated together
- Conditional dependencies checked (mode, environment)
- Errors wrapped with field path context
- Multi-error accumulation with `multierr`

**Defaults:**
- All fields initialized in `Default*Config()`
- Sensible production defaults
- Consistent with existing components

**Environment Variables:**
- All fields have `envconfig` tags
- Naming convention: `kuma_<component>_<field>` (snake_case, lowercase)
- Processed in `PostProcess()` via `envconfig.Process()`

**Sanitization:**
- Sensitive fields redacted in `Sanitize()`
- Called before logging/display
- Covers passwords, tokens, keys, certs

**Testing:**
- Validation tests (table-driven)
- Environment variable override tests
- Sanitization tests for sensitive fields
- Load from YAML tests

**Structure:**
- Embedded `BaseConfig` to avoid boilerplate
- Hierarchical nesting with clear ownership
- JSON tags for YAML marshaling
- Consistent with existing patterns

## Quick Reference

**Core Interfaces:**
- `Config` - `pkg/config/config.go:10`
- `Loader` - `pkg/config/loader.go:20`

**Key Configs:**
- kuma-cp: `pkg/config/app/kuma-cp/config.go:50`
- kuma-dp: `pkg/config/app/kuma-dp/config.go:30`
- Store: `pkg/config/core/resources/store/config.go:10`
- XDS: `pkg/config/xds/config.go:10`
- API Server: `pkg/config/api-server/config.go:10`
- DP Server: `pkg/config/dp-server/config.go:10`
- Multizone KDS: `pkg/config/multizone/kds.go:10`

**Custom Types:**
- Duration: `pkg/config/types/duration.go:10`
- TLS: `pkg/config/types/tls.go:10`

**Utilities:**
- Display: `pkg/config/display.go:10`
- Deprecate: `pkg/config/deprecate.go:10`
- Util: `pkg/config/util.go:10`

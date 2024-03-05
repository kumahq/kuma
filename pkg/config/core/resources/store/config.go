package store

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/k8s"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var _ config.Config = &StoreConfig{}

type StoreType = string

const (
	KubernetesStore StoreType = "kubernetes"
	PostgresStore   StoreType = "postgres"
	PgxStore        StoreType = "pgx"
	MemoryStore     StoreType = "memory"
)

// StoreConfig defines Resource Store configuration
type StoreConfig struct {
	// Type of Store used in the Control Plane. Can be either "kubernetes", "postgres" or "memory"
	Type StoreType `json:"type" envconfig:"kuma_store_type"`
	// Postgres Store configuration
	Postgres *postgres.PostgresStoreConfig `json:"postgres"`
	// Kubernetes Store configuration
	Kubernetes *k8s.KubernetesStoreConfig `json:"kubernetes"`
	// Cache configuration
	Cache CacheStoreConfig `json:"cache"`
	// Upsert configuration
	Upsert UpsertConfig `json:"upsert"`
	// UnsafeDelete skips validation of resource delete.
	// For example you don't have to delete all Dataplane objects before you delete a Mesh
	UnsafeDelete bool `json:"unsafeDelete" envconfig:"kuma_store_unsafe_delete"`
}

func DefaultStoreConfig() *StoreConfig {
	return &StoreConfig{
		Type:       MemoryStore,
		Postgres:   postgres.DefaultPostgresStoreConfig(),
		Kubernetes: k8s.DefaultKubernetesStoreConfig(),
		Cache:      DefaultCacheStoreConfig(),
		Upsert:     DefaultUpsertConfig(),
	}
}

func (s *StoreConfig) Sanitize() {
	s.Kubernetes.Sanitize()
	s.Postgres.Sanitize()
	s.Cache.Sanitize()
}

func (s *StoreConfig) PostProcess() error {
	return multierr.Combine(
		s.Kubernetes.PostProcess(),
		s.Postgres.PostProcess(),
		s.Cache.PostProcess(),
	)
}

func (s *StoreConfig) Validate() error {
	switch s.Type {
	case PostgresStore:
		if err := s.Postgres.Validate(); err != nil {
			return errors.Wrap(err, "Postgres validation failed")
		}
	case KubernetesStore:
		if err := s.Kubernetes.Validate(); err != nil {
			return errors.Wrap(err, "Kubernetes validation failed")
		}
		return nil
	case MemoryStore:
		return nil
	default:
		return errors.Errorf("Type should be either %s, %s or %s", PostgresStore, KubernetesStore, MemoryStore)
	}
	if err := s.Cache.Validate(); err != nil {
		return errors.Wrap(err, "Cache validation failed")
	}
	return nil
}

var _ config.Config = &CacheStoreConfig{}

type CacheStoreConfig struct {
	config.BaseConfig

	Enabled        bool                  `json:"enabled" envconfig:"kuma_store_cache_enabled"`
	ExpirationTime config_types.Duration `json:"expirationTime" envconfig:"kuma_store_cache_expiration_time"`
}

func DefaultCacheStoreConfig() CacheStoreConfig {
	return CacheStoreConfig{
		Enabled:        true,
		ExpirationTime: config_types.Duration{Duration: time.Second},
	}
}

func DefaultUpsertConfig() UpsertConfig {
	return UpsertConfig{
		ConflictRetryBaseBackoff:   config_types.Duration{Duration: 200 * time.Millisecond},
		ConflictRetryMaxTimes:      10,
		ConflictRetryJitterPercent: 30,
	}
}

type UpsertConfig struct {
	config.BaseConfig

	// Base time for exponential backoff on upsert (get and update) operations when retry is enabled
	ConflictRetryBaseBackoff config_types.Duration `json:"conflictRetryBaseBackoff" envconfig:"kuma_store_upsert_conflict_retry_base_backoff"`
	// Max retries on upsert (get and update) operation when retry is enabled
	ConflictRetryMaxTimes uint `json:"conflictRetryMaxTimes" envconfig:"kuma_store_upsert_conflict_retry_max_times"`
	// Percentage of jitter. For example: if backoff is 20s, and this value 10, the backoff will be between 18s and 22s.
	ConflictRetryJitterPercent uint `json:"conflictRetryJitterPercent" envconfig:"kuma_store_upsert_conflict_retry_jitter_percent"`
}

func (u *UpsertConfig) Validate() error {
	if u.ConflictRetryBaseBackoff.Duration < 0 {
		return errors.New("RetryBaseBackoff cannot be lower than 0")
	}
	return nil
}

var _ config.Config = &UpsertConfig{}

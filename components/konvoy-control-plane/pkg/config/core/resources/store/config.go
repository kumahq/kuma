package store

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/plugins/resources/postgres"
	"github.com/pkg/errors"
)

var _ config.Config = &StoreConfig{}

type StoreType = string

const (
	KubernetesStore StoreType = "kubernetes"
	PostgresStore   StoreType = "postgres"
	MemoryStore     StoreType = "memory"
)

// Resource Store configuration
type StoreConfig struct {
	// Type of Store used in the Control Plane
	Type StoreType `yaml:"type" envconfig:"konvoy_store_type"`
	// Postgres Store configuration
	Postgres *postgres.PostgresStoreConfig `yaml:"postgres"`
}

func DefaultStoreConfig() *StoreConfig {
	return &StoreConfig{
		Type: MemoryStore,
	}
}

func (s *StoreConfig) Validate() error {
	switch s.Type {
	case PostgresStore:
		if err := s.Postgres.Validate(); err != nil {
			return errors.Wrap(err, "Postgres validation failed")
		}
	case KubernetesStore:
		return nil
	case MemoryStore:
		return nil
	default:
		return errors.Errorf("Type should be either %s, %s or %s", PostgresStore, KubernetesStore, MemoryStore)
	}
	return nil
}

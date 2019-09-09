package store

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
	"github.com/Kong/kuma/pkg/config/plugins/resources/k8s"
	"github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
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
	// Type of Store used in the Control Plane. Can be either "kubernetes", "postgres" or "memory"
	Type StoreType `yaml:"type" envconfig:"kuma_store_type"`
	// Postgres Store configuration
	Postgres *postgres.PostgresStoreConfig `yaml:"postgres"`
	// Kubernetes Store configuration
	Kubernetes *k8s.KubernetesStoreConfig `yaml:"kubernetes"`
}

func DefaultStoreConfig() *StoreConfig {
	return &StoreConfig{
		Type:       MemoryStore,
		Postgres:   postgres.DefaultPostgresStoreConfig(),
		Kubernetes: k8s.DefaultKubernetesStoreConfig(),
	}
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
	return nil
}

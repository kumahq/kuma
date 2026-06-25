package store

import (
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
)

type SecretStore interface {
	core_store.ResourceStore
}

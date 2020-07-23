package store

import (
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

type SecretStore interface {
	core_store.ResourceStore
}

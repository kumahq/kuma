package store

import (
	core_store "github.com/kumahq/kuma/v2/pkg/core/resources/store"
)

type SecretStore interface {
	core_store.ResourceStore
}

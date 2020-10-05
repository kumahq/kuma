package defaults

import (
	"context"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func (d *defaultsComponent) signingKeyExists() (bool, error) {
	_, err := issuer.GetSigningKey(d.resManager)
	switch err {
	case issuer.SigningKeyNotFound:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (d *defaultsComponent) createSigningKeyIfNotExist() error {
	exists, err := d.signingKeyExists()
	if err != nil {
		return err
	}
	if exists {
		log.V(1).Info("default Signing Key already exists. Skip creating default Signing Key.")
	} else {
		log.Info("trying to create Signing Key")
		key, err := issuer.CreateSigningKey()
		if err != nil {
			return err
		}
		// We are using directly Resource Store instead of Resource Manager to bypass Mesh validation. Even if there is no default Mesh, we still want to create signing key.
		// Drop this if we will switch to have Secrets not scoped to the Mesh.
		if err := d.resStore.Create(context.Background(), &key, core_store.CreateBy(issuer.SigningKeyResourceKey)); err != nil {
			log.V(1).Info("could not create Signing Key", "err", err)
			return err
		}
		log.Info("Signing Key for generating Dataplane Token created")
	}
	return nil
}

func (d *defaultsComponent) shouldCreateSigningKey() bool {
	switch d.cpMode {
	case config_core.Standalone:
		return d.environment == config_core.UniversalEnvironment // Signing Key should be created only on Universal since it is not used on K8S
	case config_core.Global:
		return true // Signing Key with multi-zone should be created on Global even if the Environment is K8S, because we may connect Universal Remote
	case config_core.Remote:
		return false // Signing Key should be synced from Global
	}
	return false
}

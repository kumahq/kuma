package manager

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/model"

	secret_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secret_cipher "github.com/Kong/kuma/pkg/core/secrets/cipher"
	secret_store "github.com/Kong/kuma/pkg/core/secrets/store"
)

type SecretManager interface {
	Create(context.Context, *secret_model.SecretResource, ...core_store.CreateOptionsFunc) error
	Update(context.Context, *secret_model.SecretResource, ...core_store.UpdateOptionsFunc) error
	Delete(context.Context, *secret_model.SecretResource, ...core_store.DeleteOptionsFunc) error
	DeleteAll(context.Context, ...core_store.DeleteAllOptionsFunc) error
	Get(context.Context, *secret_model.SecretResource, ...core_store.GetOptionsFunc) error
	List(context.Context, *secret_model.SecretResourceList, ...core_store.ListOptionsFunc) error
}

func NewSecretManager(secretStore secret_store.SecretStore, cipher secret_cipher.Cipher) SecretManager {
	return &secretManager{
		secretStore: secretStore,
		cipher:      cipher,
	}
}

var _ SecretManager = &secretManager{}

type secretManager struct {
	secretStore secret_store.SecretStore
	cipher      secret_cipher.Cipher
}

func (s *secretManager) Get(ctx context.Context, secret *secret_model.SecretResource, fs ...core_store.GetOptionsFunc) error {
	if err := s.secretStore.Get(ctx, secret, fs...); err != nil {
		return err
	}
	return s.decrypt(secret)
}

func (s *secretManager) List(ctx context.Context, secrets *secret_model.SecretResourceList, fs ...core_store.ListOptionsFunc) error {
	if err := s.secretStore.List(ctx, secrets, fs...); err != nil {
		return err
	}
	for _, secret := range secrets.Items {
		if err := s.decrypt(secret); err != nil {
			return err
		}
	}
	return nil
}

func (s *secretManager) Create(ctx context.Context, secret *secret_model.SecretResource, fs ...core_store.CreateOptionsFunc) error {
	if err := s.encrypt(secret); err != nil {
		return err
	}
	return s.secretStore.Create(ctx, secret, fs...)
}

func (s *secretManager) Update(ctx context.Context, secret *secret_model.SecretResource, fs ...core_store.UpdateOptionsFunc) error {
	if err := s.encrypt(secret); err != nil {
		return err
	}
	return s.secretStore.Update(ctx, secret, fs...)
}

func (s *secretManager) Delete(ctx context.Context, secret *secret_model.SecretResource, fs ...core_store.DeleteOptionsFunc) error {
	return s.secretStore.Delete(ctx, secret, fs...)
}

func (s *secretManager) DeleteAll(ctx context.Context, fs ...core_store.DeleteAllOptionsFunc) error {
	list := &secret_model.SecretResourceList{}
	opts := core_store.NewDeleteAllOptions(fs...)
	if err := s.secretStore.List(context.Background(), list, core_store.ListByMesh(opts.Mesh)); err != nil {
		return err
	}
	for _, item := range list.Items {
		if err := s.Delete(ctx, item, core_store.DeleteBy(model.MetaToResourceKey(item.Meta))); err != nil {
			return err
		}
	}
	return nil
}

func (s *secretManager) encrypt(secret *secret_model.SecretResource) error {
	if len(secret.Spec.Value) > 0 {
		value, err := s.cipher.Encrypt(secret.Spec.Value)
		if err != nil {
			return err
		}
		secret.Spec.Value = value
	}
	return nil
}

func (s *secretManager) decrypt(secret *secret_model.SecretResource) error {
	if len(secret.Spec.Value) > 0 {
		value, err := s.cipher.Decrypt(secret.Spec.Value)
		if err != nil {
			return err
		}
		secret.Spec.Value = value
	}
	return nil
}

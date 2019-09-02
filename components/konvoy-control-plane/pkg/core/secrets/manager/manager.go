package manager

import (
	"context"

	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	secret_cryptor "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/cryptor"
	secret_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/model"
	secret_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/store"
)

type SecretManager interface {
	Create(context.Context, *secret_model.Secret, ...core_store.CreateOptionsFunc) error
	Update(context.Context, *secret_model.Secret, ...core_store.UpdateOptionsFunc) error
	Delete(context.Context, *secret_model.Secret, ...core_store.DeleteOptionsFunc) error
	Get(context.Context, *secret_model.Secret, ...core_store.GetOptionsFunc) error
	List(context.Context, *secret_model.SecretList, ...core_store.ListOptionsFunc) error
}

func NewSecretManager(secretStore secret_store.SecretStore, cryptor secret_cryptor.Cryptor) SecretManager {
	return &secretManager{
		secretStore: secretStore,
		cryptor:     cryptor,
	}
}

var _ SecretManager = &secretManager{}

type secretManager struct {
	secretStore secret_store.SecretStore
	cryptor     secret_cryptor.Cryptor
}

func (s *secretManager) Get(ctx context.Context, secret *secret_model.Secret, fs ...core_store.GetOptionsFunc) error {
	if err := s.secretStore.Get(ctx, secret, fs...); err != nil {
		return err
	}
	return s.decrypt(secret)
}

func (s *secretManager) List(ctx context.Context, secrets *secret_model.SecretList, fs ...core_store.ListOptionsFunc) error {
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

func (s *secretManager) Create(ctx context.Context, secret *secret_model.Secret, fs ...core_store.CreateOptionsFunc) error {
	if err := s.encrypt(secret); err != nil {
		return err
	}
	return s.secretStore.Create(ctx, secret, fs...)
}

func (s *secretManager) Update(ctx context.Context, secret *secret_model.Secret, fs ...core_store.UpdateOptionsFunc) error {
	if err := s.encrypt(secret); err != nil {
		return err
	}
	return s.secretStore.Update(ctx, secret, fs...)
}

func (s *secretManager) Delete(ctx context.Context, secret *secret_model.Secret, fs ...core_store.DeleteOptionsFunc) error {
	return s.secretStore.Delete(ctx, secret, fs...)
}

func (s *secretManager) encrypt(secret *secret_model.Secret) error {
	if 0 < len(secret.Spec.Value) {
		value, err := s.cryptor.Encrypt(secret.Spec.Value)
		if err != nil {
			return err
		}
		secret.Spec.Value = value
	}
	return nil
}

func (s *secretManager) decrypt(secret *secret_model.Secret) error {
	if 0 < len(secret.Spec.Value) {
		value, err := s.cryptor.Decrypt(secret.Spec.Value)
		if err != nil {
			return err
		}
		secret.Spec.Value = value
	}
	return nil
}

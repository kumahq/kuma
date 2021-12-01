package manager

import (
	"context"
	"time"

	"github.com/pkg/errors"

	secret_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	secret_cipher "github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
)

func NewSecretManager(secretStore secret_store.SecretStore, cipher secret_cipher.Cipher, validator SecretValidator) manager.ResourceManager {
	return &secretManager{
		secretStore: secretStore,
		cipher:      cipher,
		validator:   validator,
	}
}

var _ manager.ResourceManager = &secretManager{}

type secretManager struct {
	secretStore secret_store.SecretStore
	cipher      secret_cipher.Cipher
	validator   SecretValidator
}

func (s *secretManager) Get(ctx context.Context, resource model.Resource, fs ...core_store.GetOptionsFunc) error {
	secret, ok := resource.(*secret_model.SecretResource)
	if !ok {
		return newInvalidTypeError()
	}
	if err := s.secretStore.Get(ctx, secret, fs...); err != nil {
		return err
	}
	return s.decrypt(secret)
}

func (s *secretManager) List(ctx context.Context, resources model.ResourceList, fs ...core_store.ListOptionsFunc) error {
	secrets, ok := resources.(*secret_model.SecretResourceList)
	if !ok {
		return newInvalidTypeError()
	}
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

func (s *secretManager) Create(ctx context.Context, resource model.Resource, fs ...core_store.CreateOptionsFunc) error {
	secret, ok := resource.(*secret_model.SecretResource)
	if !ok {
		return newInvalidTypeError()
	}
	if err := s.encrypt(secret); err != nil {
		return err
	}
	if err := s.secretStore.Create(ctx, secret, append(fs, core_store.CreatedAt(time.Now()))...); err != nil {
		return err
	}
	return s.decrypt(secret)
}

func (s *secretManager) Update(ctx context.Context, resource model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	secret, ok := resource.(*secret_model.SecretResource)
	if !ok {
		return newInvalidTypeError()
	}
	if err := s.encrypt(secret); err != nil {
		return err
	}
	if err := s.secretStore.Update(ctx, secret, append(fs, core_store.ModifiedAt(time.Now()))...); err != nil {
		return err
	}
	return s.decrypt(secret)
}

func (s *secretManager) Delete(ctx context.Context, resource model.Resource, fs ...core_store.DeleteOptionsFunc) error {
	secret, ok := resource.(*secret_model.SecretResource)
	if !ok {
		return newInvalidTypeError()
	}
	opts := core_store.NewDeleteOptions(fs...)
	if err := s.validator.ValidateDelete(ctx, opts.Name, opts.Mesh); err != nil {
		return err
	}
	return s.secretStore.Delete(ctx, secret, fs...)
}

func (s *secretManager) DeleteAll(ctx context.Context, secrets model.ResourceList, fs ...core_store.DeleteAllOptionsFunc) error {
	list := &secret_model.SecretResourceList{}
	opts := core_store.NewDeleteAllOptions(fs...)
	if err := s.secretStore.List(ctx, list, core_store.ListByMesh(opts.Mesh)); err != nil {
		return err
	}
	for _, item := range list.Items {
		if err := s.Delete(ctx, item, core_store.DeleteBy(model.MetaToResourceKey(item.Meta))); err != nil && !core_store.IsResourceNotFound(err) {
			return err
		}
	}
	return nil
}

func newInvalidTypeError() error {
	return errors.New("resource has a wrong type")
}

func (s *secretManager) encrypt(secret *secret_model.SecretResource) error {
	if len(secret.Spec.GetData().GetValue()) > 0 {
		value, err := s.cipher.Encrypt(secret.Spec.Data.Value)
		if err != nil {
			return err
		}
		secret.Spec.Data.Value = value
	}
	return nil
}

func (s *secretManager) decrypt(secret *secret_model.SecretResource) error {
	if len(secret.Spec.GetData().GetValue()) > 0 {
		value, err := s.cipher.Decrypt(secret.Spec.Data.Value)
		if err != nil {
			return err
		}
		secret.Spec.Data.Value = value
	}
	return nil
}

package gatewayapi

import (
	"context"
	"encoding/pem"
	"fmt"

	"google.golang.org/protobuf/types/known/wrapperspb"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func (r *GatewayReconciler) createSecretIfMissing(
	ctx context.Context,
	key types.NamespacedName,
	mesh string,
) (model.ResourceKey, error) {
	resKey := model.ResourceKey{
		Name: gatewaySecretKeyName(key),
		Mesh: mesh,
	}
	log := r.Log.WithValues("key", key, "name", resKey.Name, "mesh", resKey.Mesh)
	log.V(1).Info("trying to copy Gateway API Secret as Kuma Secret")

	exists := true
	kumaSecret := system.NewSecretResource()
	if err := r.ResourceManager.Get(ctx, kumaSecret, core_store.GetBy(resKey)); err != nil {
		if core_store.IsResourceNotFound(err) {
			exists = false
		} else {
			return model.ResourceKey{}, err
		}
	}

	if exists {
		log.V(1).Info("secret already exist, no need to copy it")
		return resKey, nil
	}

	secret := &kube_core.Secret{}
	if err := r.Client.Get(ctx, key, secret); err != nil {
		return resKey, err
	}

	data, err := convertSecret(secret)
	if err != nil {
		return model.ResourceKey{}, err
	}

	kumaSecret.Spec.Data = &wrapperspb.BytesValue{
		Value: data,
	}

	log.Info("creating a copy of a secret")
	if err := r.ResourceManager.Create(ctx, kumaSecret, core_store.CreateBy(resKey)); err != nil {
		return model.ResourceKey{}, err
	}
	log.Info("secret copied")
	return resKey, nil
}

func convertSecret(secret *kube_core.Secret) ([]byte, error) {
	if secret.Type != kube_core.SecretTypeTLS {
		return nil, fmt.Errorf("only secrets of type %q are supported", secret.Type)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: secret.Data["tls.key"],
	})
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: secret.Data["tls.crt"],
	})

	return append(privatePEM, publicPEM...), nil
}

func gatewaySecretKeyName(key types.NamespacedName) string {
	return fmt.Sprintf("gapi-%s-%s", key.Namespace, key.Name)
}

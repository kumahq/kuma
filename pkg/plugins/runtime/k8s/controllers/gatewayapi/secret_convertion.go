package gatewayapi

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/wrapperspb"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
)

func (r *GatewayReconciler) createSecretIfMissing(
	ctx context.Context,
	key types.NamespacedName,
	data []byte,
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
		if core_store.IsNotFound(err) {
			exists = false
		} else {
			return model.ResourceKey{}, err
		}
	}

	if exists {
		log.V(1).Info("secret already exist, no need to copy it")
		return resKey, nil
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

// convertSecret returns the data to be stored as a Kuma secret or an error to
// be displayed in a condition message.
func convertSecret(secret *kube_core.Secret) ([]byte, error) {
	if secret.Type != kube_core.SecretTypeTLS {
		return nil, errors.Errorf("only secrets of type %q are supported", kube_core.SecretTypeTLS)
	}

	data := append(secret.Data["tls.key"], secret.Data["tls.crt"]...)

	if _, _, err := gateway.NewServerSecret(data); err != nil {
		// Don't return the exact error
		return nil, fmt.Errorf("malformed secret")
	}

	return data, nil
}

func gatewaySecretKeyName(key types.NamespacedName) string {
	return fmt.Sprintf("gapi-%s-%s", key.Namespace, key.Name)
}

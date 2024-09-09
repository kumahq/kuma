package gatewayapi

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretController tracks a Kuma Secret copy of Gateway API Secret
// Whenever Secret is used for TLS termination in Gateway API, we copy this to system namespace.
// Secret is created in GatewayController, SecretController updates and deletes the secret.
type SecretController struct {
	Log                                  logr.Logger
	Client                               kube_client.Client
	SystemNamespace                      string
	SupportGatewaySecretsInAllNamespaces bool
}

func (r *SecretController) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	if !r.SupportGatewaySecretsInAllNamespaces && req.Namespace != r.SystemNamespace {
		r.Log.V(1).Info("ignoring reconcile because SupportGatewaySecretsInAllNamespaces is disabled", "req", req)
		return kube_ctrl.Result{}, nil
	}
	r.Log.Info("reconcile", "req", req)

	copiedSecretKey := types.NamespacedName{
		Namespace: r.SystemNamespace,
		Name:      gatewaySecretKeyName(req.NamespacedName),
	}

	secret := &kube_core.Secret{}
	if err := r.Client.Get(ctx, req.NamespacedName, secret); err != nil {
		if kube_apierrs.IsNotFound(err) {
			r.Log.V(1).Info("secret was removed, trying to remove Kuma copy", "req", req)
			if err := r.deleteCopiedSecret(ctx, copiedSecretKey); err != nil {
				return kube_ctrl.Result{}, err
			}
			return kube_ctrl.Result{}, nil
		}
		return kube_ctrl.Result{}, errors.Wrapf(err, "unable to fetch Secret %s", req.NamespacedName.String())
	}

	if secret.Type != kube_core.SecretTypeTLS {
		r.Log.V(1).Info("secret is not of a type TLS, nothing to update", "req", req)
		return kube_ctrl.Result{}, nil
	}

	if err := r.updateCopiedSecret(ctx, copiedSecretKey, secret); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *SecretController) updateCopiedSecret(ctx context.Context, key types.NamespacedName, originalSecret *kube_core.Secret) error {
	r.Log.V(1).Info("trying to update Kuma copy of a secret", "key", key.String())
	copiedSecret := &kube_core.Secret{}
	if err := r.Client.Get(ctx, key, copiedSecret); err != nil {
		if kube_apierrs.IsNotFound(err) {
			r.Log.V(1).Info("secret was not copied, nothing to update", "key", key.String())
			return nil
		}
		return errors.Wrapf(err, "unable to fetch Secret %s", key.String())
	}

	bytes, err := convertSecret(originalSecret)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(copiedSecret.Data["value"], bytes) {
		r.Log.V(1).Info("secret data is the same, nothing to update", "key", key.String())
		return nil
	}

	r.Log.Info("updating copied secret", "key", key.String())
	copiedSecret.Data["value"] = bytes
	return r.Client.Update(ctx, copiedSecret)
}

func (r *SecretController) deleteCopiedSecret(ctx context.Context, key types.NamespacedName) error {
	copiedSecret := &kube_core.Secret{}
	copiedSecret.Name = key.Name
	copiedSecret.Namespace = key.Namespace
	r.Log.V(1).Info("deleting copied secret", "key", key.String())
	if err := r.Client.Delete(ctx, copiedSecret); err != nil {
		if kube_apierrs.IsNotFound(err) {
			r.Log.V(1).Info("secret was not copied, nothing to delete", "key", key.String())
			return nil
		}
		return err
	}
	r.Log.Info("copied secret deleted", "key", key.String())
	return nil
}

func (r *SecretController) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-secret-controller").
		For(&kube_core.Secret{}).
		Complete(r)
}

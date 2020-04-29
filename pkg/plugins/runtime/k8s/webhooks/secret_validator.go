package webhooks

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/validators"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"

	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	meshLabel = "kuma.io/mesh"
)

type SecretValidator struct {
	Decoder *admission.Decoder
	Client  kube_client.Client
}

func (v *SecretValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	secret := &kube_core.Secret{}

	err := v.Decoder.Decode(req, secret)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	oldSecret := &kube_core.Secret{}
	if len(req.OldObject.Raw) != 0 {
		err := v.Decoder.DecodeRaw(req.OldObject, oldSecret)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	} else {
		oldSecret = nil
	}

	if err := v.validate(ctx, secret, oldSecret); err != nil {
		if verr, ok := err.(*validators.ValidationError); ok {
			return convertValidationErrorOf(*verr, secret, secret)
		}
		return admission.Denied(err.Error())
	}

	return admission.Allowed("")
}

func (v *SecretValidator) validate(ctx context.Context, secret *kube_core.Secret, oldSecret *kube_core.Secret) error {
	verr := &validators.ValidationError{}
	if !isKumaSecret(secret) {
		return nil
	}

	// validate mesh exists
	mesh := mesh_k8s.Mesh{}
	key := kube_types.NamespacedName{
		Name: meshOfSecret(secret),
	}
	if err := v.Client.Get(ctx, key, &mesh); err != nil {
		if kube_apierrs.IsNotFound(err) {
			verr.AddViolationAt(validators.RootedAt("metadata").Field("labels").Key(meshLabel), "mesh does not exist")
		} else {
			return errors.Wrap(err, "could not fetch mesh")
		}
	}

	// block change of the mesh on the secret
	if oldSecret != nil {
		if meshOfSecret(secret) != meshOfSecret(oldSecret) {
			verr.AddViolationAt(validators.RootedAt("metadata").Field("labels").Key(meshLabel), "cannot change mesh of the Secret. Delete the Secret first and apply it again.")
		}
	}

	// validate convention of the secret
	if len(secret.Data["value"]) == 0 {
		verr.AddViolationAt(validators.RootedAt("data").Field("value"), "cannot be empty.")
	}
	return verr.OrNil()
}

func isKumaSecret(secret *kube_core.Secret) bool {
	return secret.Type == "system.kuma.io/secret"
}

func meshOfSecret(secret *kube_core.Secret) string {
	meshName := secret.GetLabels()[meshLabel]
	if meshName == "" {
		return "default"
	}
	return meshName
}

func (v *SecretValidator) InjectDecoder(d *admission.Decoder) error {
	v.Decoder = d
	return nil
}

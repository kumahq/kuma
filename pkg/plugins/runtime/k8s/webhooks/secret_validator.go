package webhooks

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	admissionv1 "k8s.io/api/admission/v1"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	"github.com/kumahq/kuma/pkg/core/validators"
	common_k8s "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

type SecretValidator struct {
	Decoder      admission.Decoder
	Client       kube_client.Reader
	Validator    secret_manager.SecretValidator
	UnsafeDelete bool
	CpMode       config_core.CpMode
}

func (v *SecretValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionv1.Delete:
		return v.handleDelete(ctx, req)
	case admissionv1.Create, admissionv1.Update:
		return v.handleUpdate(ctx, req)
	}
	return admission.Allowed("")
}

func (v *SecretValidator) handleUpdate(ctx context.Context, req admission.Request) admission.Response {
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

func (v *SecretValidator) handleDelete(ctx context.Context, req admission.Request) admission.Response {
	if v.UnsafeDelete {
		return admission.Allowed("")
	}
	secret := &kube_core.Secret{}
	if err := v.Client.Get(ctx, kube_types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, secret); err != nil {
		if kube_apierrs.IsNotFound(err) { // let K8S handle case when resource is not found
			return admission.Allowed("")
		}
		return admission.Errored(http.StatusBadRequest, err)
	}
	if secret.Type == common_k8s.MeshSecretType {
		if syncedResource(v.CpMode, secret.GetLabels()) {
			return admission.Allowed("ignore. It's synced resource.")
		}
		if err := v.Validator.ValidateDelete(ctx, req.Name, meshOfSecret(secret)); err != nil {
			if verr, ok := err.(*validators.ValidationError); ok {
				return convertValidationErrorOf(*verr, secret, secret)
			}
			return admission.Denied(err.Error())
		}
	}
	return admission.Allowed("")
}

func (v *SecretValidator) validate(ctx context.Context, secret *kube_core.Secret, oldSecret *kube_core.Secret) error {
	verr := &validators.ValidationError{}

	switch secret.Type {
	case common_k8s.MeshSecretType:
		if err := v.validateMeshSecret(ctx, verr, secret, oldSecret); err != nil {
			return err
		}
	case common_k8s.GlobalSecretType:
		v.validateGlobalSecret(verr, secret)
	}

	return verr.OrNil()
}

func (v *SecretValidator) validateMeshSecret(ctx context.Context, verr *validators.ValidationError, secret *kube_core.Secret, oldSecret *kube_core.Secret) error {
	// validate mesh exists
	mesh := mesh_k8s.Mesh{}
	key := kube_types.NamespacedName{
		Name: meshOfSecret(secret),
	}
	if err := v.Client.Get(ctx, key, &mesh); err != nil {
		if !kube_apierrs.IsNotFound(err) {
			return errors.Wrap(err, "could not fetch mesh")
		}
	}

	// block change of the mesh on the secret
	if oldSecret != nil {
		if meshOfSecret(secret) != meshOfSecret(oldSecret) {
			verr.AddViolationAt(validators.RootedAt("metadata").Field("labels").Key(metadata.KumaMeshLabel), "cannot change mesh of the Secret. Delete the Secret first and apply it again.")
		}
	}
	v.validateSecretData(verr, secret)
	return nil
}

func (v *SecretValidator) validateGlobalSecret(verr *validators.ValidationError, secret *kube_core.Secret) {
	if _, ok := secret.GetLabels()[metadata.KumaMeshLabel]; ok {
		verr.AddViolationAt(validators.RootedAt("metadata").Field("labels").Key(metadata.KumaMeshLabel), "mesh cannot be set on global secret")
	}
	v.validateSecretData(verr, secret)
}

func (v *SecretValidator) validateSecretData(verr *validators.ValidationError, secret *kube_core.Secret) {
	// validate convention of the secret
	if len(secret.Data["value"]) == 0 {
		verr.AddViolationAt(validators.RootedAt("data").Field("value"), "cannot be empty.")
	}
}

func meshOfSecret(secret *kube_core.Secret) string {
	meshName := secret.GetLabels()[metadata.KumaMeshLabel]
	if meshName == "" {
		return "default"
	}
	return meshName
}

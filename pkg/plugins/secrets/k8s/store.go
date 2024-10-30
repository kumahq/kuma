package k8s

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	secret_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	common_k8s "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ secret_store.SecretStore = &KubernetesStore{}

type KubernetesStore struct {
	reader           kube_client.Reader
	writer           kube_client.Writer
	scheme           *runtime.Scheme
	secretsConverter Converter
	// secrets have a special converter, and we need to convert the mesh object since it's the owner
	resourcesConverter common_k8s.Converter
	// Namespace to store Secrets in, e.g. namespace where Control Plane is installed to
	namespace string
}

func NewStore(reader kube_client.Reader, writer kube_client.Writer, scheme *runtime.Scheme, namespace string) (secret_store.SecretStore, error) {
	return &KubernetesStore{
		reader:             reader,
		writer:             writer,
		scheme:             scheme,
		secretsConverter:   DefaultConverter(),
		resourcesConverter: k8s.NewSimpleConverter(),
		namespace:          namespace,
	}, nil
}

func (s *KubernetesStore) Create(ctx context.Context, r core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	opts := core_store.NewCreateOptions(fs...)
	secret, err := s.secretsConverter.ToKubernetesObject(r)
	if err != nil {
		return errors.Wrap(err, "failed to convert core Secret into k8s counterpart")
	}
	secret.Namespace = s.namespace
	secret.Name = opts.Name
<<<<<<< HEAD
	if r.Descriptor().Name == secret_model.SecretType {
		labels := map[string]string{
			metadata.KumaMeshLabel: opts.Mesh,
		}
		secret.SetLabels(labels)
	}
=======

	setLabelsAnnotationsAndMesh(secret, opts.Mesh, opts.Labels)
>>>>>>> e1179afd8 (fix(k8s): set annotation kuma.io/display-name for Secrets and Configs (#11923))

	if opts.Owner != nil {
		k8sOwner, err := s.resourcesConverter.ToKubernetesObject(opts.Owner)
		if err != nil {
			return errors.Wrap(err, "failed to convert core model into k8s counterpart")
		}
		if err := controllerutil.SetOwnerReference(k8sOwner, secret, s.scheme); err != nil {
			return errors.Wrap(err, "failed to set owner reference for object")
		}
	}

	if err := s.writer.Create(ctx, secret); err != nil {
		if kube_apierrs.IsAlreadyExists(err) {
			return core_store.ErrorResourceAlreadyExists(r.Descriptor().Name, secret.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to create k8s Secret")
	}
	err = s.secretsConverter.ToCoreResource(secret, r)
	if err != nil {
		return errors.Wrap(err, "failed to convert k8s Secret into core counterpart")
	}
	return nil
}

func (s *KubernetesStore) Update(ctx context.Context, r core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	secret, err := s.secretsConverter.ToKubernetesObject(r)
	if err != nil {
		return errors.Wrap(err, "failed to convert core Secret into k8s counterpart")
	}
	secret.Namespace = s.namespace
<<<<<<< HEAD
=======

	updateLabels := r.GetMeta().GetLabels()
	if opts.ModifyLabels {
		updateLabels = opts.Labels
	}

	setLabelsAnnotationsAndMesh(secret, r.GetMeta().GetMesh(), updateLabels)

>>>>>>> e1179afd8 (fix(k8s): set annotation kuma.io/display-name for Secrets and Configs (#11923))
	if err := s.writer.Update(ctx, secret); err != nil {
		if kube_apierrs.IsConflict(err) {
			return core_store.ErrorResourceConflict(r.Descriptor().Name, secret.Name, r.GetMeta().GetMesh())
		}
		return errors.Wrap(err, "failed to update k8s Secret")
	}
	err = s.secretsConverter.ToCoreResource(secret, r)
	if err != nil {
		return errors.Wrap(err, "failed to convert k8s Secret into core counterpart")
	}
	return nil
}

<<<<<<< HEAD
=======
func setLabelsAnnotationsAndMesh(s *kube_core.Secret, mesh string, labels map[string]string) {
	if labels == nil {
		labels = map[string]string{}
	}
	if mesh != "" {
		labels[metadata.KumaMeshLabel] = mesh
	}

	labels, annotations := k8s.SplitLabelsAndAnnotations(labels, s.GetAnnotations())
	s.GetObjectMeta().SetLabels(labels)
	s.GetObjectMeta().SetAnnotations(annotations)
}

>>>>>>> e1179afd8 (fix(k8s): set annotation kuma.io/display-name for Secrets and Configs (#11923))
func (s *KubernetesStore) Delete(ctx context.Context, r core_model.Resource, fs ...core_store.DeleteOptionsFunc) error {
	opts := core_store.NewDeleteOptions(fs...)
	if err := s.Get(ctx, r, core_store.GetByKey(opts.Name, opts.Mesh)); err != nil {
		return errors.Wrap(err, "failed to delete k8s secret")
	}

	secret, err := s.secretsConverter.ToKubernetesObject(r)
	if err != nil {
		return errors.Wrap(err, "failed to convert core Secret into k8s counterpart")
	}
	secret.Namespace = s.namespace
	secret.Name = opts.Name

	if err := s.writer.Delete(ctx, secret); err != nil {
		return errors.Wrap(err, "failed to delete k8s Secret")
	}
	return nil
}

func (s *KubernetesStore) Get(ctx context.Context, r core_model.Resource, fs ...core_store.GetOptionsFunc) error {
	opts := core_store.NewGetOptions(fs...)
	secret := &kube_core.Secret{}
	if err := s.reader.Get(ctx, kube_client.ObjectKey{Namespace: s.namespace, Name: opts.Name}, secret); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return core_store.ErrorResourceNotFound(r.Descriptor().Name, opts.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to get k8s secret")
	}
	if err := s.secretsConverter.ToCoreResource(secret, r); err != nil {
		return errors.Wrap(err, "failed to convert k8s Secret into core counterpart")
	}
	if err := assertFound(r, secret, opts.Name, opts.Mesh); err != nil {
		return err
	}
	return nil
}

func assertFound(r core_model.Resource, secret *kube_core.Secret, name string, mesh string) error {
	switch r.Descriptor().Name {
	case secret_model.SecretType:
		// secret must match mesh and be a proper type, otherwise return not found
		if r.GetMeta().GetMesh() != mesh || secret.Type != common_k8s.MeshSecretType {
			if err := r.SetSpec(&system_proto.Secret{}); err != nil {
				return err
			}
			return core_store.ErrorResourceNotFound(r.Descriptor().Name, name, mesh)
		}
	case secret_model.GlobalSecretType:
		// secret must be a proper type, otherwise return not found
		if secret.Type != common_k8s.GlobalSecretType {
			if err := r.SetSpec(&system_proto.Secret{}); err != nil {
				return err
			}
			return core_store.ErrorResourceNotFound(r.Descriptor().Name, name, mesh)
		}
	}
	return nil
}

func (s *KubernetesStore) List(ctx context.Context, rs core_model.ResourceList, fs ...core_store.ListOptionsFunc) error {
	opts := core_store.NewListOptions(fs...)
	secrets := &kube_core.SecretList{}

	fields := kube_client.MatchingFields{} // list only Kuma System secrets
	labels := kube_client.MatchingLabels{}
	switch rs.GetItemType() {
	case secret_model.SecretType:
		fields = kube_client.MatchingFields{ // list only Kuma System secrets
			"type": common_k8s.MeshSecretType,
		}
		if opts.Mesh != "" {
			labels[metadata.KumaMeshLabel] = opts.Mesh
		}
	case secret_model.GlobalSecretType:
		fields = kube_client.MatchingFields{ // list only Kuma System secrets
			"type": common_k8s.GlobalSecretType,
		}
	}
	if err := s.reader.List(ctx, secrets, kube_client.InNamespace(s.namespace), labels, fields); err != nil {
		return errors.Wrap(err, "failed to list k8s Secrets")
	}
	if err := s.secretsConverter.ToCoreList(secrets, rs); err != nil {
		return errors.Wrap(err, "failed to convert k8s Secret into core counterpart")
	}
	return nil
}

var _ core_model.ResourceMeta = &KubernetesMetaAdapter{}

type KubernetesMetaAdapter struct {
	kube_meta.ObjectMeta
	SecretType kube_core.SecretType
}

func (m *KubernetesMetaAdapter) GetNameExtensions() core_model.ResourceNameExtensions {
	return common_k8s.ResourceNameExtensions(m.ObjectMeta.Namespace, m.ObjectMeta.Name)
}

func (m *KubernetesMetaAdapter) GetVersion() string {
	return m.ObjectMeta.GetResourceVersion()
}

func (m *KubernetesMetaAdapter) GetMesh() string {
	if m.SecretType == common_k8s.GlobalSecretType {
		return ""
	}
	mesh, exist := m.Labels[metadata.KumaMeshLabel]
	if !exist {
		mesh = core_model.DefaultMesh
	}
	return mesh
}

func (m *KubernetesMetaAdapter) GetCreationTime() time.Time {
	return m.GetObjectMeta().GetCreationTimestamp().Time
}

func (m *KubernetesMetaAdapter) GetModificationTime() time.Time {
	return m.GetObjectMeta().GetCreationTimestamp().Time
}

func (m *KubernetesMetaAdapter) GetLabels() map[string]string {
	labels := maps.Clone(m.GetObjectMeta().GetLabels())
	if labels == nil {
		labels = map[string]string{}
	}
	if displayName, ok := m.GetObjectMeta().GetAnnotations()[v1alpha1.DisplayName]; ok {
		labels[v1alpha1.DisplayName] = displayName
	} else {
		labels[v1alpha1.DisplayName] = m.GetObjectMeta().GetName()
	}
	return labels
}

type Converter interface {
	ToKubernetesObject(resource core_model.Resource) (*kube_core.Secret, error)
	ToCoreResource(secret *kube_core.Secret, out core_model.Resource) error
	ToCoreList(list *kube_core.SecretList, out core_model.ResourceList) error
}

func DefaultConverter() Converter {
	return &SimpleConverter{}
}

var _ Converter = &SimpleConverter{}

type SimpleConverter struct{}

func (c *SimpleConverter) ToKubernetesObject(r core_model.Resource) (*kube_core.Secret, error) {
	secret := &kube_core.Secret{}
	switch r.Descriptor().Name {
	case secret_model.SecretType:
		secret.Type = common_k8s.MeshSecretType
		secret.Data = map[string][]byte{
			"value": r.(*secret_model.SecretResource).Spec.GetData().GetValue(),
		}
		if r.GetMeta() != nil {
			labels := map[string]string{
				metadata.KumaMeshLabel: r.GetMeta().GetMesh(),
			}
			secret.SetLabels(labels)
		}
	case secret_model.GlobalSecretType:
		secret.Type = common_k8s.GlobalSecretType
		secret.Data = map[string][]byte{
			"value": r.(*secret_model.GlobalSecretResource).Spec.GetData().GetValue(),
		}
	default:
		return nil, errors.Errorf("invalid type %s, expected %s or %s", r.Descriptor().Name, secret_model.SecretType, secret_model.GlobalSecretType)
	}
	if r.GetMeta() != nil {
		if adapter, ok := r.GetMeta().(*KubernetesMetaAdapter); ok {
			secret.ObjectMeta = adapter.ObjectMeta
		} else {
			return nil, fmt.Errorf("meta has unexpected type: %#v", r.GetMeta())
		}
	}
	return secret, nil
}

func (c *SimpleConverter) ToCoreResource(secret *kube_core.Secret, out core_model.Resource) error {
	out.SetMeta(&KubernetesMetaAdapter{
		ObjectMeta: secret.ObjectMeta,
		SecretType: secret.Type,
	})
	if secret.Data != nil {
		_ = out.SetSpec(&system_proto.Secret{
			Data: util_proto.Bytes(secret.Data["value"]),
		})
	}
	return nil
}

func (c *SimpleConverter) ToCoreList(in *kube_core.SecretList, out core_model.ResourceList) error {
	switch out.GetItemType() {
	case secret_model.SecretType:
		secOut := out.(*secret_model.SecretResourceList)
		secOut.Items = make([]*secret_model.SecretResource, len(in.Items))
		for i := range in.Items {
			r := secret_model.NewSecretResource()
			if err := c.ToCoreResource(&in.Items[i], r); err != nil {
				return err
			}
			secOut.Items[i] = r
		}
	case secret_model.GlobalSecretType:
		secOut := out.(*secret_model.GlobalSecretResourceList)
		secOut.Items = make([]*secret_model.GlobalSecretResource, len(in.Items))
		for i := range in.Items {
			r := secret_model.NewGlobalSecretResource()
			if err := c.ToCoreResource(&in.Items[i], r); err != nil {
				return err
			}
			secOut.Items[i] = r
		}
	default:
		return errors.Errorf("invalid type %s, expected %s or %s", out.GetItemType(), secret_model.SecretType, secret_model.GlobalSecretType)
	}
	return nil
}

package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	secret_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secret_store "github.com/Kong/kuma/pkg/core/secrets/store"
	common_k8s "github.com/Kong/kuma/pkg/plugins/common/k8s"

	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Every Kuma Secret will be annotated with kuma.io/mesh to be able to expose secrets only for given mesh
	meshLabel = "kuma.io/mesh"
	// Kubernetes secret type to differentiate Kuma System secrets
	secretType = "system.kuma.io/secret"
)

var _ secret_store.SecretStore = &KubernetesStore{}

type KubernetesStore struct {
	reader    kube_client.Reader
	writer    kube_client.Writer
	converter Converter
	// Namespace to store Secrets in, e.g. namespace where Control Plane is installed to
	namespace string
}

func NewStore(reader kube_client.Reader, writer kube_client.Writer, namespace string) (secret_store.SecretStore, error) {
	return &KubernetesStore{
		reader:    reader,
		writer:    writer,
		converter: DefaultConverter(),
		namespace: namespace,
	}, nil
}

func (s *KubernetesStore) Create(ctx context.Context, r *secret_model.SecretResource, fs ...core_store.CreateOptionsFunc) error {
	opts := core_store.NewCreateOptions(fs...)
	secret, err := s.converter.ToKubernetesObject(r)
	if err != nil {
		return errors.Wrap(err, "failed to convert core Secret into k8s counterpart")
	}
	secret.Namespace = s.namespace
	secret.Name = opts.Name
	labels := map[string]string{
		meshLabel: opts.Mesh,
	}
	secret.SetLabels(labels)

	if err := s.writer.Create(ctx, secret); err != nil {
		if kube_apierrs.IsAlreadyExists(err) {
			return core_store.ErrorResourceAlreadyExists(r.GetType(), secret.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to create k8s Secret")
	}
	err = s.converter.ToCoreResource(secret, r)
	if err != nil {
		return errors.Wrap(err, "failed to convert k8s Secret into core counterpart")
	}
	return nil
}
func (s *KubernetesStore) Update(ctx context.Context, r *secret_model.SecretResource, fs ...core_store.UpdateOptionsFunc) error {
	secret, err := s.converter.ToKubernetesObject(r)
	if err != nil {
		return errors.Wrap(err, "failed to convert core Secret into k8s counterpart")
	}
	secret.Namespace = s.namespace
	if err := s.writer.Update(ctx, secret); err != nil {
		if kube_apierrs.IsConflict(err) {
			return core_store.ErrorResourceConflict(r.GetType(), secret.Name, r.GetMeta().GetMesh())
		}
		return errors.Wrap(err, "failed to update k8s Secret")
	}
	err = s.converter.ToCoreResource(secret, r)
	if err != nil {
		return errors.Wrap(err, "failed to convert k8s Secret into core counterpart")
	}
	return nil
}
func (s *KubernetesStore) Delete(ctx context.Context, r *secret_model.SecretResource, fs ...core_store.DeleteOptionsFunc) error {
	opts := core_store.NewDeleteOptions(fs...)
	secret, err := s.converter.ToKubernetesObject(r)
	if err != nil {
		return errors.Wrap(err, "failed to convert core Secret into k8s counterpart")
	}
	secret.Namespace = s.namespace
	secret.Name = opts.Name

	if err := s.writer.Delete(ctx, secret); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "failed to delete k8s Secret")
	}
	return nil
}
func (s *KubernetesStore) Get(ctx context.Context, r *secret_model.SecretResource, fs ...core_store.GetOptionsFunc) error {
	opts := core_store.NewGetOptions(fs...)
	secret := &kube_core.Secret{}
	if err := s.reader.Get(ctx, kube_client.ObjectKey{Namespace: s.namespace, Name: opts.Name}, secret); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return core_store.ErrorResourceNotFound(r.GetType(), opts.Name, opts.Mesh)
		}
		return errors.Wrap(err, "failed to get k8s secret")
	}
	if err := s.converter.ToCoreResource(secret, r); err != nil {
		return errors.Wrap(err, "failed to convert k8s Secret into core counterpart")
	}
	if r.GetMeta().GetMesh() != opts.Mesh {
		if err := r.SetSpec(&system_proto.Secret{}); err != nil {
			return err
		}
		return core_store.ErrorResourceNotFound(r.GetType(), opts.Name, opts.Mesh)
	}
	return nil
}
func (s *KubernetesStore) List(ctx context.Context, rs *secret_model.SecretResourceList, fs ...core_store.ListOptionsFunc) error {
	opts := core_store.NewListOptions(fs...)
	secrets := &kube_core.SecretList{}

	fields := kube_client.MatchingFields{ // list only Kuma System secrets
		"type": secretType,
	}
	labels := kube_client.MatchingLabels{}
	if opts.Mesh != "" {
		labels[meshLabel] = opts.Mesh
	}
	if err := s.reader.List(ctx, secrets, kube_client.InNamespace(s.namespace), labels, fields); err != nil {
		return errors.Wrap(err, "failed to list k8s Secrets")
	}
	if err := s.converter.ToCoreList(secrets, rs); err != nil {
		return errors.Wrap(err, "failed to convert k8s Secret into core counterpart")
	}
	return nil
}

var _ core_model.ResourceMeta = &KubernetesMetaAdapter{}

type KubernetesMetaAdapter struct {
	kube_meta.ObjectMeta
}

func (m *KubernetesMetaAdapter) GetNameExtensions() core_model.ResourceNameExtensions {
	return common_k8s.ResourceNameExtensions(m.ObjectMeta.Namespace, m.ObjectMeta.Name)
}

func (m *KubernetesMetaAdapter) GetVersion() string {
	return m.ObjectMeta.GetResourceVersion()
}

func (m *KubernetesMetaAdapter) GetMesh() string {
	mesh, exist := m.Labels[meshLabel]
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

type Converter interface {
	ToKubernetesObject(*secret_model.SecretResource) (*kube_core.Secret, error)
	ToCoreResource(secret *kube_core.Secret, out *secret_model.SecretResource) error
	ToCoreList(list *kube_core.SecretList, out *secret_model.SecretResourceList) error
}

func DefaultConverter() Converter {
	return &SimpleConverter{}
}

var _ Converter = &SimpleConverter{}

type SimpleConverter struct {
}

func (c *SimpleConverter) ToKubernetesObject(r *secret_model.SecretResource) (*kube_core.Secret, error) {
	secret := &kube_core.Secret{}
	secret.Type = secretType
	secret.Data = map[string][]byte{
		"value": r.Spec.GetData().GetValue(),
	}
	if r.GetMeta() != nil {
		if adapter, ok := r.GetMeta().(*KubernetesMetaAdapter); ok {
			secret.ObjectMeta = adapter.ObjectMeta
			labels := map[string]string{
				meshLabel: r.GetMeta().GetMesh(),
			}
			secret.SetLabels(labels)
		} else {
			return nil, fmt.Errorf("meta has unexpected type: %#v", r.GetMeta())
		}
	}
	return secret, nil
}

func (c *SimpleConverter) ToCoreResource(secret *kube_core.Secret, out *secret_model.SecretResource) error {
	out.SetMeta(&KubernetesMetaAdapter{secret.ObjectMeta})
	if secret.Data != nil {
		out.Spec = system_proto.Secret{
			Data: &wrappers.BytesValue{
				Value: secret.Data["value"],
			},
		}
	}
	return nil
}

func (c *SimpleConverter) ToCoreList(in *kube_core.SecretList, out *secret_model.SecretResourceList) error {
	out.Items = make([]*secret_model.SecretResource, len(in.Items))
	for i, secret := range in.Items {
		r := &secret_model.SecretResource{}
		if err := c.ToCoreResource(&secret, r); err != nil {
			return err
		}
		out.Items[i] = r
	}
	return nil
}

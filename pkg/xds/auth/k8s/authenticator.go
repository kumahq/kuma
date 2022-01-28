package k8s

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	kube_auth "k8s.io/api/authentication/v1"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
	"github.com/kumahq/kuma/pkg/xds/auth"
)

func New(client kube_client.Client) auth.Authenticator {
	return &kubeAuthenticator{
		client: client,
	}
}

type kubeAuthenticator struct {
	client kube_client.Client
}

var _ auth.Authenticator = &kubeAuthenticator{}

func (k *kubeAuthenticator) Authenticate(ctx context.Context, resource model.Resource, credential auth.Credential) error {
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource:
		return k.authDataplane(ctx, resource, credential)
	case *core_mesh.ZoneIngressResource:
		return k.authZoneIngress(ctx, resource, credential)
	case *core_mesh.ZoneEgressResource:
		return k.authZoneEgress(ctx, resource, credential)
	default:
		return errors.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
	}
}

func (k *kubeAuthenticator) authDataplane(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential auth.Credential) error {
	proxyName, proxyNamespace, err := util_k8s.CoreNameToK8sName(dataplane.Meta.GetName())
	if err != nil {
		return err
	}
	serviceAccountName, err := k.podServiceAccountName(ctx, proxyName, proxyNamespace)
	if err != nil {
		return err
	}
	if err := k.verifyToken(ctx, credential, proxyNamespace, serviceAccountName); err != nil {
		return errors.Wrap(err, "authentication failed")
	}
	return nil
}

func (k *kubeAuthenticator) verifyToken(ctx context.Context, credential auth.Credential, proxyNamespace, serviceAccountName string) error {
	tokenReview := &kube_auth.TokenReview{
		Spec: kube_auth.TokenReviewSpec{
			Token: credential,
		},
	}
	if err := k.client.Create(ctx, tokenReview); err != nil {
		return errors.Wrap(err, "call to TokenReview API failed")
	}
	if !tokenReview.Status.Authenticated {
		return errors.Errorf("token doesn't belong to a valid user")
	}
	userInfo := strings.Split(tokenReview.Status.User.Username, ":")
	if len(userInfo) != 4 {
		return errors.Errorf("username inside TokenReview response has unexpected format: %q", tokenReview.Status.User.Username)
	}
	if !(userInfo[0] == "system" && userInfo[1] == "serviceaccount") {
		return errors.Errorf("user %q is not a service account", tokenReview.Status.User.Username)
	}
	namespace := userInfo[2]
	if namespace != proxyNamespace {
		return errors.Errorf("token belongs to a namespace %q different from proxyId %q", namespace, proxyNamespace)
	}
	name := userInfo[3]
	if name != serviceAccountName {
		return errors.Errorf("service account name of the pod %q is different than token that was provided %q", serviceAccountName, name)
	}
	return nil
}

func (k *kubeAuthenticator) authZoneIngress(ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource, credential auth.Credential) error {
	proxyName, proxyNamespace, err := util_k8s.CoreNameToK8sName(zoneIngress.Meta.GetName())
	if err != nil {
		return err
	}
	serviceAccountName, err := k.podServiceAccountName(ctx, proxyName, proxyNamespace)
	if err != nil {
		return err
	}
	if err := k.verifyToken(ctx, credential, proxyNamespace, serviceAccountName); err != nil {
		return errors.Wrap(err, "authentication failed")
	}
	return nil
}

func (k *kubeAuthenticator) authZoneEgress(ctx context.Context, zoneEgress *core_mesh.ZoneEgressResource, credential auth.Credential) error {
	proxyName, proxyNamespace, err := util_k8s.CoreNameToK8sName(zoneEgress.Meta.GetName())
	if err != nil {
		return err
	}

	serviceAccountName, err := k.podServiceAccountName(ctx, proxyName, proxyNamespace)
	if err != nil {
		return err
	}

	if err := k.verifyToken(ctx, credential, proxyNamespace, serviceAccountName); err != nil {
		return errors.Wrap(err, "authentication failed")
	}

	return nil
}

func (k *kubeAuthenticator) podServiceAccountName(ctx context.Context, podName, podNamespace string) (string, error) {
	pod := &kube_core.Pod{}
	if err := k.client.Get(ctx, types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	}, pod); err != nil {
		return "", errors.Wrapf(err, "could not retrieve Pod %s/%s to verify identity of a dataplane proxy", podNamespace, podName)
	}
	if pod.Spec.ServiceAccountName != "" {
		return pod.Spec.ServiceAccountName, nil
	}
	return "default", nil // if ServiceAccount is not expicitly defined in a Pod, it's "default" SA in a namespace.
}

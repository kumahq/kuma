package k8s

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	kube_auth "k8s.io/api/authentication/v1"
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
	default:
		return errors.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
	}
}

func (k *kubeAuthenticator) authDataplane(ctx context.Context, dataplane *core_mesh.DataplaneResource, credential auth.Credential) error {
	if credential == "" {
		return errors.New("authentication failed: k8s token is missing")
	}
	tokenReview := &kube_auth.TokenReview{
		Spec: kube_auth.TokenReviewSpec{
			Token: credential,
		},
	}
	if err := k.client.Create(ctx, tokenReview); err != nil {
		return errors.Wrap(err, "authentication failed: call to TokenReview API failed")
	}
	if !tokenReview.Status.Authenticated {
		return errors.Errorf("authentication failed: token doesn't belong to a valid user")
	}
	userInfo := strings.Split(tokenReview.Status.User.Username, ":")
	if len(userInfo) != 4 {
		return errors.Errorf("authentication failed: username inside TokenReview response has unexpected format: %q", tokenReview.Status.User.Username)
	}
	if !(userInfo[0] == "system" && userInfo[1] == "serviceaccount") {
		return errors.Errorf("authentication failed: token must belong to a k8s system account, got %q", tokenReview.Status.User.Username)
	}
	_, proxyNamespace, err := util_k8s.CoreNameToK8sName(dataplane.Meta.GetName())
	if err != nil {
		return err
	}
	namespace := userInfo[2]
	if namespace != proxyNamespace {
		return errors.Errorf("authentication failed: token belongs to a namespace (%q) different from proxyId (%q)", namespace, proxyNamespace)
	}
	return nil
}

func (k *kubeAuthenticator) authZoneIngress(ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource, credential auth.Credential) error {
	if credential == "" {
		return errors.New("authentication failed: k8s token is missing")
	}
	tokenReview := &kube_auth.TokenReview{
		Spec: kube_auth.TokenReviewSpec{
			Token: credential,
		},
	}
	if err := k.client.Create(ctx, tokenReview); err != nil {
		return errors.Wrap(err, "authentication failed: call to TokenReview API failed")
	}
	if !tokenReview.Status.Authenticated {
		return errors.Errorf("authentication failed: token doesn't belong to a valid user")
	}
	userInfo := strings.Split(tokenReview.Status.User.Username, ":")
	if len(userInfo) != 4 {
		return errors.Errorf("authentication failed: username inside TokenReview response has unexpected format: %q", tokenReview.Status.User.Username)
	}
	if !(userInfo[0] == "system" && userInfo[1] == "serviceaccount") {
		return errors.Errorf("authentication failed: token must belong to a k8s system account, got %q", tokenReview.Status.User.Username)
	}
	_, proxyNamespace, err := util_k8s.CoreNameToK8sName(zoneIngress.Meta.GetName())
	if err != nil {
		return err
	}
	namespace := userInfo[2]
	if namespace != proxyNamespace {
		return errors.Errorf("authentication failed: token belongs to a namespace (%q) different from proxyId (%q)", namespace, proxyNamespace)
	}
	return nil
}

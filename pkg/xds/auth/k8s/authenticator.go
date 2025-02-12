package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kube_auth "k8s.io/api/authentication/v1"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
	"github.com/kumahq/kuma/pkg/xds/auth"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

var log = core.Log.WithName("kube-token-validator")

func New(client kube_client.Client, metrics *xds_metrics.Metrics) auth.Authenticator {
	return &kubeAuthenticator{
		client: client,
	}
}

type kubeAuthenticator struct {
	client kube_client.Client
}

var _ auth.Authenticator = &kubeAuthenticator{}

func (k *kubeAuthenticator) Authenticate(ctx context.Context, resource model.Resource, credential auth.Credential) error {
	var err error
	switch resource := resource.(type) {
	case *core_mesh.DataplaneResource, *core_mesh.ZoneIngressResource, *core_mesh.ZoneEgressResource:
		err = k.authResource(ctx, resource, credential)
	default:
		err = errors.Errorf("no matching authenticator for %s resource", resource.Descriptor().Name)
	}

	if err != nil {
		return err
	}
	return nil
}

func (k *kubeAuthenticator) authResource(ctx context.Context, resource model.Resource, credential auth.Credential) error {
	resourceName := resource.GetMeta().GetName()
	podName, podNamespace, err := util_k8s.CoreNameToK8sName(resourceName)
	if err != nil {
		return err
	}

	serviceAccountName, err := k.podServiceAccountName(ctx, podName, podNamespace)
	if err != nil {
		return err
	}

	tokenReview := &kube_auth.TokenReview{
		Spec: kube_auth.TokenReviewSpec{
			Token: credential,
		},
	}

	if err := k.client.Create(ctx, tokenReview); err != nil {
		log.Error(err, "could not call Kubernetes API Server to verify dataplane token",
			"proxy", resourceName, "serviceAccountName", serviceAccountName)
		return errors.New("could not call Kubernetes API Server to verify dataplane token")
	}
	if !tokenReview.Status.Authenticated {
		log.Info("[WARNING] fail to verify dataplane token", "error", tokenReview.Status.Error)
		return errors.Errorf("dataplane token verification failed with an error")
	}

	serviceAccountAuthErr := errors.Errorf("invalid service account token")
	userInfo := strings.Split(tokenReview.Status.User.Username, ":")
	if len(userInfo) != 4 {
		log.Info(fmt.Sprintf("[WARNING] invalid service account token: username inside TokenReview response has unexpected format: %q",
			tokenReview.Status.User.Username), "proxy", resourceName, "serviceAccountName", serviceAccountName)
		return serviceAccountAuthErr
	}

	if !(userInfo[0] == "system" && userInfo[1] == "serviceaccount") {
		log.Info(fmt.Sprintf("[WARNING] invalid service account token: user %q is not a service account", tokenReview.Status.User.Username),
			"proxy", resourceName, "serviceAccountName", serviceAccountName)
		return serviceAccountAuthErr
	}

	tokenSANamespace := userInfo[2]
	if tokenSANamespace != podNamespace {
		log.Info(fmt.Sprintf("[WARNING] invalid service account token: token belongs to a namespace %q different from proxyId %q", tokenSANamespace, podNamespace),
			"proxy", resourceName, "serviceAccountName", serviceAccountName)
		return serviceAccountAuthErr
	}

	tokenSAName := userInfo[3]
	if tokenSAName != serviceAccountName {
		log.Info(fmt.Sprintf("[WARNING] invalid service account token: service account name in token %q does not match pod service account", tokenSAName),
			"proxy", resourceName, "serviceAccountName", serviceAccountName)
		return serviceAccountAuthErr
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

package stub

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	common_auth "github.com/Kong/kuma/pkg/sds/auth/common"
	util_k8s "github.com/Kong/kuma/pkg/util/k8s"

	kube_auth "k8s.io/api/authentication/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

func New(client kube_client.Client, dataplaneResolver common_auth.DataplaneResolver) sds_auth.Authenticator {
	return &kubeAuthenticator{
		client:            client,
		dataplaneResolver: dataplaneResolver,
	}
}

type kubeAuthenticator struct {
	client            kube_client.Client
	dataplaneResolver common_auth.DataplaneResolver
}

func (k *kubeAuthenticator) Authenticate(ctx context.Context, proxyId core_xds.ProxyId, credential sds_auth.Credential) (sds_auth.Identity, error) {
	if err := k.reviewToken(ctx, proxyId, credential); err != nil {
		return sds_auth.Identity{}, err
	}
	// at this point we know that proxyId belongs to the same namespace as token.
	// since legacy k8s tokens are not bound to a specific Pod,
	// we have to rely on information included into proxyId.
	dataplane, err := k.dataplaneResolver(ctx, proxyId)
	if err != nil {
		return sds_auth.Identity{}, errors.Wrapf(err, "unable to find Dataplane for proxy %q", proxyId)
	}
	return common_auth.GetDataplaneIdentity(dataplane)
}

func (k *kubeAuthenticator) reviewToken(ctx context.Context, proxyId core_xds.ProxyId, credential sds_auth.Credential) error {
	if credential == "" {
		return errors.New("authentication failed: k8s token is missing")
	}
	tokenReview := &kube_auth.TokenReview{
		Spec: kube_auth.TokenReviewSpec{
			Token: string(credential),
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
	_, proxyNamespace, err := util_k8s.CoreNameToK8sName(proxyId.Name)
	if err != nil {
		return err
	}
	namespace := userInfo[2]
	if namespace != proxyNamespace {
		return errors.Errorf("authentication failed: token belongs to a namespace (%q) different from proxyId (%q)", namespace, proxyNamespace)
	}
	return nil
}

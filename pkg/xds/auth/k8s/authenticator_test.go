package k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_auth "k8s.io/api/authentication/v1"
	kube_core "k8s.io/api/core/v1"
	kube_client_scheme "k8s.io/client-go/kubernetes/scheme"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	auth_k8s "github.com/kumahq/kuma/v2/pkg/xds/auth/k8s"
)

type tokenReviewClient struct {
	kube_client.Client
	username      string
	authenticated bool
	reviewed      bool
}

func (c *tokenReviewClient) Create(ctx context.Context, obj kube_client.Object, opts ...kube_client.CreateOption) error {
	tokenReview, ok := obj.(*kube_auth.TokenReview)
	if !ok {
		return c.Client.Create(ctx, obj, opts...)
	}
	c.reviewed = true
	tokenReview.Status = kube_auth.TokenReviewStatus{
		Authenticated: c.authenticated,
		User:          kube_auth.UserInfo{Username: c.username},
	}
	return nil
}

var _ = Describe("Authenticate", func() {
	type testCase struct {
		dpLabels     map[string]string
		podSA        string
		tokenUser    string
		expectedErr  string
		expectReview bool
	}

	DescribeTable("should verify the identity of a dataplane proxy",
		func(given testCase) {
			// given
			pod := &kube_core.Pod{
				Name: "backend", Namespace: "example",
				Spec: kube_core.PodSpec{ServiceAccountName: given.podSA},
			}
			client := &tokenReviewClient{
				Client: kube_client_fake.NewClientBuilder().
					WithScheme(kube_client_scheme.Scheme).
					WithObjects(pod).
					Build(),
				username:      given.tokenUser,
				authenticated: true,
			}
			dataplane := core_mesh.NewDataplaneResource()
			dataplane.SetMeta(&test_model.ResourceMeta{
				Mesh:   "default",
				Name:   "backend.example",
				Labels: given.dpLabels,
			})

			// when
			err := auth_k8s.New(client, nil).Authenticate(context.Background(), dataplane, "token")

			// then
			if given.expectedErr == "" {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(MatchError(given.expectedErr))
			}
			Expect(client.reviewed).To(Equal(given.expectReview))
		},
		Entry("service account label matches the Pod service account", testCase{
			dpLabels:     map[string]string{metadata.KumaServiceAccount: "backend-sa"},
			podSA:        "backend-sa",
			tokenUser:    "system:serviceaccount:example:backend-sa",
			expectReview: true,
		}),
		Entry("service account label is forged", testCase{
			dpLabels:    map[string]string{metadata.KumaServiceAccount: "victim-sa"},
			podSA:       "backend-sa",
			tokenUser:   "system:serviceaccount:example:backend-sa",
			expectedErr: "invalid service account token",
		}),
		Entry("service account label is not set", testCase{
			podSA:        "backend-sa",
			tokenUser:    "system:serviceaccount:example:backend-sa",
			expectReview: true,
		}),
		Entry("token was issued for another service account", testCase{
			dpLabels:     map[string]string{metadata.KumaServiceAccount: "backend-sa"},
			podSA:        "backend-sa",
			tokenUser:    "system:serviceaccount:example:other-sa",
			expectedErr:  "invalid service account token",
			expectReview: true,
		}),
	)
})

package verifier

import (
	"context"
	"fmt"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/kds"
)

func Create(ctx context.Context, r model.Resource, opts ...store.CreateOptionsFunc) Executable {
	return func(tc TestContext) error {
		return tc.Store().Create(ctx, r, opts...)
	}
}

func DiscoveryRequest(node *envoy_core.Node, resourceType model.ResourceType) Executable {
	return func(tc TestContext) error {
		tc.Stream().RecvCh <- &v2.DiscoveryRequest{
			Node:    node,
			TypeUrl: string(resourceType),
		}
		return nil
	}
}

func Ack(node *envoy_core.Node, resourceType model.ResourceType) Executable {
	return func(tc TestContext) error {
		tc.Stream().RecvCh <- &v2.DiscoveryRequest{
			Node:          node,
			TypeUrl:       string(resourceType),
			ResponseNonce: tc.LastResponse(string(resourceType)).Nonce,
			VersionInfo:   tc.LastResponse(string(resourceType)).VersionInfo,
		}
		return nil
	}
}

func WaitResponse(timeout time.Duration, testFunc func(krs []*mesh_proto.KumaResource)) Executable {
	return func(tc TestContext) error {
		select {
		case resp := <-tc.Stream().SentCh:
			krs, err := kumaResources(resp)
			if err != nil {
				return err
			}
			if len(krs) > 0 {
				tc.SaveLastResponse(kds.ResourceType(krs[0].Spec.TypeUrl), resp)
			}
			testFunc(krs)
		case <-time.After(timeout):
			return fmt.Errorf("timeout exceeded")
		}
		return nil
	}
}

func kumaResources(response *v2.DiscoveryResponse) (resources []*mesh_proto.KumaResource, _ error) {
	for _, r := range response.Resources {
		kr := &mesh_proto.KumaResource{}
		if err := ptypes.UnmarshalAny(r, kr); err != nil {
			return nil, err
		}
		resources = append(resources, kr)
	}
	return
}

func ExpectNoResponseDuring(timeout time.Duration) Executable {
	return func(tc TestContext) error {
		t := time.Now()
		select {
		case resp := <-tc.Stream().SentCh:
			return fmt.Errorf("received response after %v: %v", time.Since(t), resp)
		case <-time.After(timeout):
			return nil
		}
	}
}

func CloseStream() Executable {
	return func(tc TestContext) error {
		close(tc.Stream().RecvCh)
		return nil
	}
}

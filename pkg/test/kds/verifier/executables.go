package verifier

import (
	"context"
	"fmt"
	"time"

	"github.com/Kong/kuma/pkg/kds/util"

	"github.com/golang/protobuf/ptypes/any"

	"github.com/Kong/kuma/pkg/util/proto"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func Create(ctx context.Context, r model.Resource, opts ...store.CreateOptionsFunc) Executable {
	return func(tc TestContext) error {
		return tc.Store().Create(ctx, r, opts...)
	}
}

func DiscoveryRequest(node *envoy_core.Node, resourceType model.ResourceType) Executable {
	return func(tc TestContext) error {
		tc.ServerStream().RecvCh <- &v2.DiscoveryRequest{
			Node:    node,
			TypeUrl: string(resourceType),
		}
		return nil
	}
}

func ACK(node *envoy_core.Node, resourceType model.ResourceType) Executable {
	return func(tc TestContext) error {
		tc.ServerStream().RecvCh <- &v2.DiscoveryRequest{
			Node:          node,
			TypeUrl:       string(resourceType),
			ResponseNonce: tc.LastResponse(string(resourceType)).Nonce,
			VersionInfo:   tc.LastResponse(string(resourceType)).VersionInfo,
		}
		tc.SaveLastACKedResponse(string(resourceType), tc.LastResponse(string(resourceType)))
		return nil
	}
}

func NACK(node *envoy_core.Node, resourceType model.ResourceType) Executable {
	return func(tc TestContext) error {
		tc.ServerStream().RecvCh <- &v2.DiscoveryRequest{
			Node:          node,
			TypeUrl:       string(resourceType),
			ResponseNonce: tc.LastResponse(string(resourceType)).Nonce,
			VersionInfo:   tc.LastACKedResponse(string(resourceType)).VersionInfo,
		}
		return nil
	}
}

func WaitResponse(timeout time.Duration, testFunc func(rs []model.Resource)) Executable {
	return func(tc TestContext) error {
		select {
		case resp := <-tc.ServerStream().SentCh:
			rs, err := util.ToCoreResourceList(resp)
			if err != nil {
				return err
			}
			if len(rs.GetItems()) > 0 {
				tc.SaveLastResponse(string(rs.GetItemType()), resp)
			}
			testFunc(rs.GetItems())
		case <-time.After(timeout):
			return fmt.Errorf("timeout exceeded")
		}
		return nil
	}
}

func ExpectNoResponseDuring(timeout time.Duration) Executable {
	return func(tc TestContext) error {
		t := time.Now()
		select {
		case resp := <-tc.ServerStream().SentCh:
			return fmt.Errorf("received response after %v: %v", time.Since(t), resp)
		case <-time.After(timeout):
			return nil
		}
	}
}

func CloseStream() Executable {
	return func(tc TestContext) error {
		close(tc.ServerStream().RecvCh)
		return nil
	}
}

func WaitRequest(timeout time.Duration, testFunc func(rs *v2.DiscoveryRequest)) Executable {
	return func(tc TestContext) error {
		select {
		case req := <-tc.ClientStream().SentCh:
			testFunc(req)
		case <-time.After(timeout):
			return fmt.Errorf("timeout exceeded")
		}
		return nil
	}
}

func DiscoveryResponse(rs model.ResourceList, nonce, version string) Executable {
	return func(tc TestContext) error {
		envoyRes, err := util.ToEnvoyResources(rs)
		if err != nil {
			return err
		}
		resources := make([]*any.Any, 0, len(envoyRes))
		for i := 0; i < len(envoyRes); i++ {
			pbaby, err := proto.MarshalAnyDeterministic(envoyRes[i])
			if err != nil {
				return err
			}
			resources = append(resources, pbaby)
		}
		tc.ClientStream().RecvCh <- &v2.DiscoveryResponse{
			TypeUrl:     string(rs.GetItemType()),
			Nonce:       nonce,
			VersionInfo: version,
			Resources:   resources,
		}
		return nil
	}
}

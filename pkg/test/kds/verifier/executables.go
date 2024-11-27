package verifier

import (
	"context"
	"fmt"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds/util"
	"github.com/kumahq/kuma/pkg/util/proto"
)

func Create(ctx context.Context, r model.Resource, opts ...store.CreateOptionsFunc) Executable {
	return func(tc TestContext) error {
		return tc.Store().Create(ctx, r, opts...)
	}
}

func DiscoveryRequest(node *envoy_core.Node, resourceType model.ResourceType) Executable {
	return func(tc TestContext) error {
		tc.ServerStream().RecvCh <- &envoy_sd.DiscoveryRequest{
			Node:    node,
			TypeUrl: string(resourceType),
		}
		return nil
	}
}

func ACK(node *envoy_core.Node, resourceType model.ResourceType) Executable {
	return func(tc TestContext) error {
		tc.ServerStream().RecvCh <- &envoy_sd.DiscoveryRequest{
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
		tc.ServerStream().RecvCh <- &envoy_sd.DiscoveryRequest{
			Node:          node,
			TypeUrl:       string(resourceType),
			ResponseNonce: tc.LastResponse(string(resourceType)).Nonce,
			VersionInfo:   tc.LastACKedResponse(string(resourceType)).GetVersionInfo(),
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

func WaitRequest(timeout time.Duration, testFunc func(rs *envoy_sd.DiscoveryRequest)) Executable {
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
		resources := make([]*anypb.Any, 0, len(envoyRes))
		for i := 0; i < len(envoyRes); i++ {
			pbaby, err := proto.MarshalAnyDeterministic(envoyRes[i])
			if err != nil {
				return err
			}
			resources = append(resources, pbaby)
		}
		tc.ClientStream().RecvCh <- &envoy_sd.DiscoveryResponse{
			TypeUrl:     string(rs.GetItemType()),
			Nonce:       nonce,
			VersionInfo: version,
			Resources:   resources,
		}
		return nil
	}
}

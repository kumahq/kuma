package envoyadmin

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_util "github.com/kumahq/kuma/pkg/config"
	config_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/kds/service"
	util_grpc "github.com/kumahq/kuma/pkg/util/grpc"
	"github.com/kumahq/kuma/pkg/util/k8s"
)

const reverseUnaryRPCService = "kuma.mesh.v1alpha1.KDSZoneEnvoyAdminService"

type kdsEnvoyAdminClient struct {
	rpcs       service.EnvoyAdminRPCs
	resManager manager.ReadOnlyResourceManager
	tracer     trace.Tracer
}

func NewClient(rpcs service.EnvoyAdminRPCs, resManager manager.ReadOnlyResourceManager) admin.EnvoyAdminClient {
	tracer := otel.GetTracerProvider().Tracer(otelgrpc.ScopeName)
	return &kdsEnvoyAdminClient{
		rpcs:       rpcs,
		resManager: resManager,
		tracer:     tracer,
	}
}

var _ admin.EnvoyAdminClient = &kdsEnvoyAdminClient{}

func (k *kdsEnvoyAdminClient) PostQuit(context.Context, *core_mesh.DataplaneResource) error {
	panic("not implemented")
}

type message interface {
	util_grpc.ReverseUnaryMessage
	GetError() string
}

func startTrace(ctx context.Context, tracer trace.Tracer, name string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(
		ctx,
		name,
		trace.WithSpanKind(trace.SpanKindClient),
		// We make up attributes for the reverse unary gRPC service
		trace.WithAttributes(
			semconv.RPCService(reverseUnaryRPCService),
			semconv.RPCMethod(name),
		),
	)
	return ctx, span
}

func doRequest[T message]( //nolint:nonamedreturns
	ctx context.Context,
	tracer trace.Tracer,
	resManager manager.ReadOnlyResourceManager,
	proxy core_model.ResourceWithAddress,
	requestType string,
	rpcs util_grpc.ReverseUnaryRPCs,
	mkMsg func(id, typ, name, mesh string) util_grpc.ReverseUnaryMessage,
) (resp T, retErr error) {
	var t T
	ctx, span := startTrace(ctx, tracer, requestType)
	defer func() {
		if retErr != nil {
			span.SetStatus(codes.Error, retErr.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}()

	zone := core_model.ZoneOfResource(proxy)
	tenantZoneID := service.TenantZoneClientIDFromCtx(ctx, zone)

	reqId := core.NewUUID()
	nameInZone, err := resNameInZone(ctx, resManager, proxy)
	if err != nil {
		return t, &KDSTransportError{requestType: requestType, reason: err.Error()}
	}
	msg := mkMsg(
		reqId,
		string(proxy.Descriptor().Name),
		nameInZone,                // send the name which without the added prefix
		proxy.GetMeta().GetMesh(), // should be empty for ZoneIngress/ZoneEgress
	)

	if err = rpcs.Send(tenantZoneID.String(), msg); err != nil {
		return t, &KDSTransportError{requestType: requestType, reason: err.Error()}
	}

	defer rpcs.DeleteWatch(tenantZoneID.String(), reqId)
	ch := make(chan util_grpc.ReverseUnaryMessage)
	if err := rpcs.WatchResponse(tenantZoneID.String(), reqId, ch); err != nil {
		return t, errors.Wrapf(err, "could not watch the response")
	}

	select {
	case <-ctx.Done():
		return t, ctx.Err()
	case resp := <-ch:
		var t T
		tResp, ok := resp.(T)
		if !ok {
			return t, errors.New("invalid request type")
		}
		if tResp.GetError() != "" {
			return t, &KDSTransportError{requestType: requestType, reason: tResp.GetError()}
		}
		return tResp, nil
	}
}

func (k *kdsEnvoyAdminClient) Stats(ctx context.Context, proxy core_model.ResourceWithAddress, format mesh_proto.AdminOutputFormat) ([]byte, error) {
	requestType := "StatsRequest"
	mkMsg := func(reqId, typ, name, mesh string) util_grpc.ReverseUnaryMessage {
		return &mesh_proto.StatsRequest{
			RequestId:    reqId,
			ResourceType: typ,
			ResourceName: name,
			ResourceMesh: mesh,
			Format:       format,
		}
	}
	resp, err := doRequest[*mesh_proto.StatsResponse](ctx, k.tracer, k.resManager, proxy, requestType, k.rpcs.Stats, mkMsg)
	if err != nil {
		return nil, err
	}
	return resp.GetStats(), nil
}

func (k *kdsEnvoyAdminClient) ConfigDump(ctx context.Context, proxy core_model.ResourceWithAddress, includeEds bool) ([]byte, error) {
	requestType := "XDSConfigRequest"
	mkMsg := func(reqId, typ, name, mesh string) util_grpc.ReverseUnaryMessage {
		return &mesh_proto.XDSConfigRequest{
			RequestId:    reqId,
			ResourceType: typ,
			ResourceName: name,
			ResourceMesh: mesh,
			IncludeEds:   includeEds,
		}
	}
	resp, err := doRequest[*mesh_proto.XDSConfigResponse](ctx, k.tracer, k.resManager, proxy, requestType, k.rpcs.XDSConfigDump, mkMsg)
	if err != nil {
		return nil, err
	}
	return resp.GetConfig(), nil
}

func (k *kdsEnvoyAdminClient) Clusters(ctx context.Context, proxy core_model.ResourceWithAddress, format mesh_proto.AdminOutputFormat) ([]byte, error) {
	requestType := "ClustersRequest"
	mkMsg := func(reqId, typ, name, mesh string) util_grpc.ReverseUnaryMessage {
		return &mesh_proto.ClustersRequest{
			RequestId:    reqId,
			ResourceType: typ,
			ResourceName: name,
			ResourceMesh: mesh,
			Format:       format,
		}
	}
	resp, err := doRequest[*mesh_proto.ClustersResponse](ctx, k.tracer, k.resManager, proxy, requestType, k.rpcs.Clusters, mkMsg)
	if err != nil {
		return nil, err
	}
	return resp.GetClusters(), nil
}

func resNameInZone(
	ctx context.Context,
	resManager manager.ReadOnlyResourceManager,
	r core_model.Resource,
) (string, error) {
	name := core_model.GetDisplayName(r.GetMeta())
	zone := core_model.ZoneOfResource(r)
	// we need to check for the legacy name which starts with zoneName
	if strings.HasPrefix(r.GetMeta().GetName(), zone) {
		return name, nil
	}
	storeType, err := getZoneStoreType(ctx, resManager, zone)
	if err != nil {
		return "", err
	}
	// only K8s needs namespace added to the resource name
	if storeType != store.KubernetesStore {
		return name, nil
	}

	if ns := r.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag]; ns != "" {
		name = k8s.K8sNamespacedNameToCoreName(name, ns)
	}
	return name, nil
}

func getZoneStoreType(
	ctx context.Context,
	resManager manager.ReadOnlyResourceManager,
	zone string,
) (store.StoreType, error) {
	zoneInsightRes := core_system.NewZoneInsightResource()
	if err := resManager.Get(ctx, zoneInsightRes, core_store.GetByKey(zone, core_model.NoMesh)); err != nil {
		return "", err
	}
	subscription := zoneInsightRes.Spec.GetLastSubscription()
	if !subscription.IsOnline() {
		return "", fmt.Errorf("zone is offline")
	}
	kdsSubscription, ok := subscription.(*system_proto.KDSSubscription)
	if !ok {
		return "", fmt.Errorf("cannot map subscription")
	}
	config := kdsSubscription.GetConfig()
	cfg := &config_cp.Config{}
	if err := config_util.FromYAML([]byte(config), cfg); err != nil {
		return "", fmt.Errorf("cannot read control-plane configuration")
	}
	return cfg.Store.Type, nil
}

type KDSTransportError struct {
	requestType string
	reason      string
}

func (e *KDSTransportError) Error() string {
	if e.reason == "" {
		return fmt.Sprintf("could not send %s", e.requestType)
	} else {
		return fmt.Sprintf("could not send %s: %s", e.requestType, e.reason)
	}
}

func (e *KDSTransportError) Is(err error) bool {
	return reflect.TypeOf(e) == reflect.TypeOf(err)
}

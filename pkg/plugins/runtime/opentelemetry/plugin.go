package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kumahq/kuma/pkg/config/tracing"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func init() {
	core_plugins.Register("tracing.opentelemetry", &plugin{})
}

var log = core.Log.WithName("tracing").WithName("opentelemetry")

type plugin struct{}

var _ core_plugins.RuntimePlugin = &plugin{}

type tracer struct {
	config tracing.OpenTelemetry
}

var _ component.Component = &tracer{}

func (t *tracer) Start(stop <-chan struct{}) error {
	shutdown, err := initOtel(context.Background(), t.config)
	if err != nil {
		return err
	}

	go func() {
		<-stop
		log.Info("stopping")
		if err := shutdown(context.Background()); err != nil {
			log.Error(err, "shutting down")
		}
	}()

	return nil
}

func (t *tracer) NeedLeaderElection() bool {
	return false
}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	otel := rt.Config().Tracing.OpenTelemetry
	if otel.Endpoint == "" {
		return nil
	}

	t := tracer{
		config: otel,
	}
	if err := rt.Add(component.NewResilientComponent(core.Log.WithName("otel-tracer"), &t)); err != nil {
		return err
	}

	return nil
}

func initOtel(ctx context.Context, otelConfig tracing.OpenTelemetry) (func(context.Context) error, error) {
	res, err := resource.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		otelConfig.Endpoint,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := trace.NewBatchSpanProcessor(traceExporter)

	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

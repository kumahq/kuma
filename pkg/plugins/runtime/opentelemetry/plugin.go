package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	datadog "github.com/tonglil/opentelemetry-go-datadog-propagator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

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
	shutdown, err := initOtel(context.Background(), log, t.config)
	if err != nil {
		return err
	}

	<-stop
	log.Info("stopping")
	if err := shutdown(context.Background()); err != nil {
		log.Error(err, "shutting down")
	}

	return nil
}

func (t *tracer) NeedLeaderElection() bool {
	return false
}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	otel := rt.Config().Tracing.OpenTelemetry
	if !otel.Enabled && otel.Endpoint == "" {
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

func initOtel(ctx context.Context, log logr.Logger, otelConfig tracing.OpenTelemetry) (func(context.Context) error, error) {
	res, err := resource.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var opts []otlptracegrpc.Option
	if otelConfig.Endpoint != "" {
		log.Info("DEPRECATED: KUMA_TRACING_OPENTELEMETRY_ENDPOINT is deprecated, use OTEL_EXPORTER_OTLP_ENDPOINT and OTEL_EXPORTER_OTLP_INSECURE instead")
		opts = append(opts, otlptracegrpc.WithEndpoint(otelConfig.Endpoint), otlptracegrpc.WithInsecure())
	}

	traceExporter, err := otlptracegrpc.New(ctx, opts...)
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

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, datadog.Propagator{}))

	return tracerProvider.Shutdown, nil
}

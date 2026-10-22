package opentelemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	datadog "github.com/tonglil/opentelemetry-go-datadog-propagator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/kumahq/kuma/v3/pkg/config/tracing"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/v3/pkg/core/runtime"
	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
	otel_extensions "github.com/kumahq/kuma/v3/pkg/plugins/extensions/otel"
)

func init() {
	// Registered as "tracing.opentelemetry" for historical reasons.
	// This plugin also handles CP metrics OTLP push (auto-enabled via OTEL_EXPORTER_OTLP_ENDPOINT).
	core_plugins.Register("tracing.opentelemetry", &plugin{})
}

var log = core.Log.WithName("tracing").WithName("opentelemetry")

type plugin struct{}

var _ core_plugins.RuntimePlugin = &plugin{}

type tracer struct {
	config    tracing.OpenTelemetry
	factories []otel_extensions.SpanProcessorFactory
}

var _ component.Component = &tracer{}

func (t *tracer) Start(stop <-chan struct{}) error {
	shutdown, err := initOtel(context.Background(), log, t.config, t.factories)
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
	otelTracing := rt.Config().Tracing.OpenTelemetry
	if otelTracing.Enabled || otelTracing.Endpoint != "" {
		t := tracer{
			config:    otelTracing,
			factories: otel_extensions.SpanProcessorFactoriesFromContext(rt.Extensions()),
		}
		if err := rt.Add(component.NewResilientComponent(core.Log.WithName("otel-tracer"), &t, rt.Config().General.ResilientComponentBaseBackoff.Duration, rt.Config().General.ResilientComponentMaxBackoff.Duration)); err != nil {
			return err
		}
	}

	if rt.Config().Metrics.OpenTelemetry.Enabled && os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		metricsLog := core.Log.WithName("otel-metrics-pusher")
		mp := &metricsPusher{
			gatherer: rt.Metrics(),
			log:      metricsLog,
		}
		if err := rt.Add(component.NewResilientComponent(metricsLog, mp, rt.Config().General.ResilientComponentBaseBackoff.Duration, rt.Config().General.ResilientComponentMaxBackoff.Duration)); err != nil {
			return err
		}
	}

	return nil
}

func initOtel(ctx context.Context, log logr.Logger, otelConfig tracing.OpenTelemetry, factories []otel_extensions.SpanProcessorFactory) (func(context.Context) error, error) {
	res, err := resource.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	processors, err := newSpanProcessors(ctx, log, otelConfig, factories)
	if err != nil {
		return nil, err
	}

	tracerProvider := newTracerProvider(res, processors)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, datadog.Propagator{}))

	return tracerProvider.Shutdown, nil
}

func newSpanProcessors(ctx context.Context, log logr.Logger, otelConfig tracing.OpenTelemetry, factories []otel_extensions.SpanProcessorFactory) ([]trace.SpanProcessor, error) {
	var opts []otlptracegrpc.Option
	if otelConfig.Endpoint != "" {
		log.Info("DEPRECATED: KUMA_TRACING_OPENTELEMETRY_ENDPOINT is deprecated, use OTEL_EXPORTER_OTLP_ENDPOINT and OTEL_EXPORTER_OTLP_INSECURE instead")
		opts = append(opts, otlptracegrpc.WithEndpoint(otelConfig.Endpoint), otlptracegrpc.WithInsecure())
	}

	primary, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	processors := []trace.SpanProcessor{trace.NewBatchSpanProcessor(primary)}
	for i, factory := range factories {
		processor, err := factory(ctx)
		if err != nil {
			for _, created := range processors {
				_ = created.Shutdown(ctx)
			}
			return nil, fmt.Errorf("failed to create additional span processor %d: %w", i, err)
		}
		processors = append(processors, processor)
	}
	if len(factories) > 0 {
		log.Info("registered additional span processors", "count", len(factories))
	}

	return processors, nil
}

func newTracerProvider(res *resource.Resource, processors []trace.SpanProcessor) *trace.TracerProvider {
	opts := []trace.TracerProviderOption{
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
	}
	for _, processor := range processors {
		opts = append(opts, trace.WithSpanProcessor(processor))
	}

	return trace.NewTracerProvider(opts...)
}

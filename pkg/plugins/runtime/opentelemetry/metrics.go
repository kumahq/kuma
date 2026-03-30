package opentelemetry

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	prombridge "go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
)

type metricsPusher struct {
	gatherer prometheus.Gatherer
}

var _ component.Component = &metricsPusher{}

func (m *metricsPusher) Start(stop <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	bridge := prombridge.NewMetricProducer(prombridge.WithGatherer(m.gatherer))

	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithProducer(bridge))

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	log.Info("OTLP metrics push started")
	<-stop
	log.Info("stopping OTLP metrics push")

	if err := provider.Shutdown(context.Background()); err != nil {
		log.Error(err, "shutting down OTLP metrics provider")
	}
	return nil
}

func (m *metricsPusher) NeedLeaderElection() bool {
	return false
}

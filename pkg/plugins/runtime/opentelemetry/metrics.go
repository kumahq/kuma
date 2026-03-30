package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	prombridge "go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
)

type metricsPusher struct {
	gatherer prometheus.Gatherer
	log      logr.Logger
}

var _ component.Component = &metricsPusher{}

func (m *metricsPusher) Start(stop <-chan struct{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	exporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	bridge := prombridge.NewMetricProducer(prombridge.WithGatherer(m.gatherer))

	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithProducer(bridge))

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	m.log.Info("OTLP metrics push started")
	<-stop
	m.log.Info("stopping OTLP metrics push")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if shutdownErr := provider.Shutdown(shutdownCtx); shutdownErr != nil {
		return fmt.Errorf("shutting down OTLP metrics provider: %w", shutdownErr)
	}
	return nil
}

func (m *metricsPusher) NeedLeaderElection() bool {
	return false
}

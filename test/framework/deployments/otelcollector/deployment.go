package otelcollector

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

const (
	DeploymentName         = "otel-collector"
	GRPCPort               = 4317
	HTTPPort               = 4318
	MetricsPort            = 8888
	PrometheusExporterPort = 8889
)

type OpenTelemetryCollector interface {
	CollectorEndpoint() string
	ExporterEndpoint() string
}

type Deployment interface {
	framework.Deployment
	OpenTelemetryCollector
}

type DeploymentOpts struct {
	namespace          string
	image              string
	networks           []string
	logLevel           string
	serviceAccountName string
	waitingToBeReady   bool
}

func DefaultOpts() DeploymentOpts {
	return DeploymentOpts{
		image:              "otel/opentelemetry-collector-contrib:0.92.0",
		networks:           []string{"kind"},
		logLevel:           "info",
		waitingToBeReady:   true,
		serviceAccountName: "otlp-collector",
	}
}

type DeploymentOpt = func(opts *DeploymentOpts)

func WithNamespace(namespace string) DeploymentOpt {
	return func(opts *DeploymentOpts) {
		opts.namespace = namespace
	}
}

func WithImage(image string) DeploymentOpt {
	return func(opts *DeploymentOpts) {
		opts.image = image
	}
}

func WithNetworks(networks ...string) DeploymentOpt {
	return func(opts *DeploymentOpts) {
		opts.networks = networks
	}
}

func WithoutWaitingToBeReady() DeploymentOpt {
	return func(opts *DeploymentOpts) {
		opts.waitingToBeReady = false
	}
}

func WithServiceAccount(serviceAccountName string) DeploymentOpt {
	return func(opts *DeploymentOpts) {
		opts.serviceAccountName = serviceAccountName
	}
}

func From(cluster framework.Cluster) OpenTelemetryCollector {
	return cluster.Deployment(DeploymentName).(OpenTelemetryCollector)
}

func Install(fs ...DeploymentOpt) framework.InstallFunc {
	opts := DefaultOpts()
	for _, fn := range fs {
		fn(&opts)
	}

	return func(cluster framework.Cluster) error {
		var deployment Deployment

		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = newK8sDeployment().
				WithImage(opts.image).
				WithNamespace(opts.namespace).
				WithWaitingToBeReady(opts.waitingToBeReady).
				WithServiceAccount(opts.serviceAccountName)
		default:
			return errors.New("invalid cluster")
		}

		if err := cluster.Deploy(deployment); err != nil {
			return err
		}

		return nil
	}
}

package kic

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

const DeploymentName = "kongingresscontroller"

type KIC interface {
	IP(namespace string) (string, error)
}

type Deployment interface {
	framework.Deployment
	KIC
}

type deployOptions struct {
	namespace string
	mesh      string
	name      string
}

type deployOptionsFunc func(*deployOptions)

func newDeployOpt(fs ...deployOptionsFunc) *deployOptions {
	rv := &deployOptions{}
	for _, f := range fs {
		f(rv)
	}
	return rv
}

func From(cluster framework.Cluster) KIC {
	return cluster.Deployment(DeploymentName).(KIC)
}

func Install(fs ...deployOptionsFunc) framework.InstallFunc {
	opts := newDeployOpt(fs...)
	return func(cluster framework.Cluster) error {
		var deployment *k8sDeployment
		switch cluster.(type) {
		case *framework.K8sCluster:
			deployment = &k8sDeployment{
				ingressNamespace: opts.namespace,
				mesh:             opts.mesh,
				name:             opts.name,
			}
		default:
			return errors.New("invalid cluster")
		}
		return cluster.Deploy(deployment)
	}
}

func WithNamespace(namespace string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.namespace = namespace
	}
}

func WithMesh(mesh string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.mesh = mesh
	}
}

func WithName(name string) deployOptionsFunc {
	return func(o *deployOptions) {
		o.name = name
	}
}

func KongIngressController(fs ...deployOptionsFunc) framework.InstallFunc {
	return Install(fs...)
}

func KongIngressService(fs ...deployOptionsFunc) framework.InstallFunc {
	opts := newDeployOpt(fs...)
	if opts.namespace == "" {
		opts.namespace = framework.Config.DefaultGatewayNamespace
	}
	svc := `
apiVersion: v1
kind: Service
metadata:
  name: gateway
  namespace: %s
spec:
  type: ClusterIP
  selector:
    app: %s-gateway
  ports:
    - name: proxy
      targetPort: 8000
      port: 80
    - name: proxy-ssl
      targetPort: 8443
      port: 443
`
	return framework.YamlK8s(fmt.Sprintf(svc, opts.namespace, opts.name))
}

package kic

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
)

const DeploymentName = "kongingresscontroller"

type KIC interface{}

type Deployment interface {
	framework.Deployment
	KIC
}

type deployOptions struct {
	namespace string
	mesh      string
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

func KongIngressController(fs ...deployOptionsFunc) framework.InstallFunc {
	return Install(fs...)
}

func KongIngressNodePort(fs ...deployOptionsFunc) framework.InstallFunc {
	proxy_port := NodePortHTTP()
	proxy_ssl_port := NodePortHTTPS()
	opts := newDeployOpt(fs...)
	if opts.namespace == "" {
		opts.namespace = framework.Config.DefaultGatewayNamespace
	}
	nodeport := `
apiVersion: v1
kind: Service
metadata:
  name: gateway-nodeport
  namespace: %s
spec:
  type: NodePort
  selector:
    app: ingress-kong
  ports:
    - name: proxy
      targetPort: 8000
      nodePort: %d
      port: %d
    - name: proxy-ssl
      targetPort: 8443
      nodePort: %d
      port: %d
`
	return framework.YamlK8s(fmt.Sprintf(nodeport, opts.namespace, proxy_port, proxy_port, proxy_ssl_port, proxy_ssl_port))
}

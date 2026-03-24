package zoneproxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/test/framework"
)

type universalDeployment struct {
	opts DeploymentOpts
}

func (d *universalDeployment) Name() string {
	return d.opts.Name
}

func (d *universalDeployment) Deploy(cluster framework.Cluster) error {
	uniCluster := cluster.(*framework.UniversalCluster)

	if d.opts.IngressPort > 0 {
		if err := d.deployProxy(uniCluster, "ZoneIngress", d.ingressName(), int(d.opts.IngressPort)); err != nil {
			return err
		}
	}
	if d.opts.EgressPort > 0 {
		if err := d.deployProxy(uniCluster, "ZoneEgress", d.egressName(), int(d.opts.EgressPort)); err != nil {
			return err
		}
	}
	return nil
}

func (d *universalDeployment) deployProxy(uniCluster *framework.UniversalCluster, listenerType, name string, port int) error {
	mode := framework.AppMode(framework.AppIngress)
	if listenerType == "ZoneEgress" {
		mode = framework.AppMode(framework.AppEgress)
	}

	app, err := framework.NewUniversalApp(
		uniCluster.GetTesting(),
		uniCluster.Name(),
		name,
		d.opts.Mesh,
		mode,
		framework.UniversalAppRunOptions{
			DockerBackend: uniCluster.GetDockerBackend(),
		},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create Universal app for %q", name)
	}

	if err := app.StartMainApp(); err != nil {
		return errors.Wrapf(err, "failed to start main app for %q", name)
	}

	ip := app.GetIP()

	workload := d.opts.Workload
	if workload == "" {
		workload = name
	}
	dpYAML, err := framework.RenderDataplaneTemplate(framework.DataplaneTemplateData{
		Mesh: d.opts.Mesh,
		Labels: map[string]string{
			"kuma.io/workload": workload,
		},
		Listeners: []framework.ListenerConfig{
			{Type: listenerType, Name: name, Port: port},
		},
	})
	if err != nil {
		return errors.Wrapf(err, "failed to render Dataplane YAML for %q", name)
	}

	token, err := uniCluster.GetKuma().GenerateDpToken(d.opts.Mesh, "", workload)
	if err != nil {
		return errors.Wrapf(err, "failed to generate DP token for %q", name)
	}

	return uniCluster.CreateDataplaneProxy(app, name, ip, dpYAML, token)
}

func (d *universalDeployment) ingressName() string {
	return fmt.Sprintf("%s-ingress", d.opts.Name)
}

func (d *universalDeployment) egressName() string {
	return fmt.Sprintf("%s-egress", d.opts.Name)
}

func (d *universalDeployment) Delete(_ framework.Cluster) error {
	return nil
}

var _ framework.Deployment = &universalDeployment{}

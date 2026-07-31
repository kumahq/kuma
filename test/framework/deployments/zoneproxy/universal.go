package zoneproxy

import (
	"context"
	"fmt"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/test/framework"
)

type universalDeployment struct {
	opts DeploymentOpts
}

func (d *universalDeployment) Name() string {
	return d.opts.Name
}

func (d *universalDeployment) Deploy(cluster framework.Cluster) error {
	uniCluster := cluster.(*framework.UniversalCluster)

	if d.opts.Ingress {
		if err := d.deployProxy(uniCluster, "ZoneIngress", d.ingressName(), int(d.opts.IngressPort)); err != nil {
			return err
		}
	}
	if d.opts.Egress {
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
			EnableIPv6:    framework.Config.IPV6,
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

	if err := uniCluster.CreateDataplaneProxy(app, name, ip, dpYAML, token, d.opts.DpEnvs); err != nil {
		return err
	}

	// CreateDataplaneProxy only starts the process. Wait for the proxy to
	// register before returning, so a proxy that never comes up fails here
	// rather than as a connection error in an unrelated assertion later. The
	// Kubernetes deployment gets the same gate from WaitPodsAvailable.
	if err := d.waitDataplaneOnline(uniCluster, name); err != nil {
		return err
	}

	if listenerType == "ZoneIngress" {
		return d.createMeshZoneAddress(uniCluster, name, ip, port)
	}
	return nil
}

func (d *universalDeployment) waitDataplaneOnline(uniCluster *framework.UniversalCluster, name string) error {
	_, err := retry.DoWithRetryContextE(
		uniCluster.GetTesting(), context.Background(),
		"wait for zone proxy "+name+" to come online",
		framework.DefaultRetries, framework.DefaultTimeout,
		func() (string, error) {
			online, found, err := framework.IsDataplaneOnline(uniCluster, d.opts.Mesh, name)
			if err != nil {
				return "", err
			}
			if !found {
				return "", errors.Errorf("zone proxy %q not registered yet", name)
			}
			if !online {
				return "", errors.Errorf("zone proxy %q not online yet", name)
			}
			return "", nil
		})
	return err
}

// createMeshZoneAddress publishes the address other zones dial to reach this
// ingress. On Kubernetes the meshzoneaddress controller derives it from the
// zone proxy Service; on Universal there is no Service, so the deployment
// creates the resource itself.
func (d *universalDeployment) createMeshZoneAddress(uniCluster *framework.UniversalCluster, name, ip string, port int) error {
	mza := fmt.Sprintf(`
type: MeshZoneAddress
name: %s
mesh: %s
labels:
  kuma.io/origin: zone
  kuma.io/zone: %s
spec:
  address: %s
  port: %d
`, name, d.opts.Mesh, uniCluster.ZoneName(), ip, port)

	return framework.NewClusterSetup().
		Install(framework.YamlUniversal(mza)).
		Setup(uniCluster)
}

func (d *universalDeployment) ingressName() string {
	return ingressName(d.opts.Name)
}

func (d *universalDeployment) egressName() string {
	return egressName(d.opts.Name)
}

func (d *universalDeployment) Delete(_ framework.Cluster) error {
	return nil
}

var _ framework.Deployment = &universalDeployment{}

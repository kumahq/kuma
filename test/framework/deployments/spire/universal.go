package spire

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"

	"github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments"
)

const (
	SpireServerPort = uint32(8999)
	SpireHealthPort = uint32(8080)
	AppSpireServer  = "spire-server"
)

type universalDeployment struct {
	name          string
	container     *deployments.DockerContainer
	trustDomain   string
	dockerBackend framework.DockerBackend
}

var _ framework.Deployment = &universalDeployment{}

func NewUniversalDeployment(cluster framework.Cluster, name string, opts *deployOptions) *universalDeployment {
	container, err := deployments.NewDockerContainer(
		deployments.WithDockerBackend(cluster.(*framework.UniversalCluster).GetDockerBackend()),
		deployments.AllocatePublicPortsFor(SpireServerPort, SpireHealthPort),
		deployments.WithContainerName(cluster.Name()+"_"+AppSpireServer),
		deployments.WithTestingT(cluster.GetTesting()),
		deployments.WithNetwork(framework.DockerNetworkName),
		deployments.WithImage(framework.Config.GetUniversalImage()),
		deployments.WithCommand("spire-server", "run", "-config", "/spire/spire-server.conf", "--trustDomain", opts.trustDomain),
	)
	if err != nil {
		panic(err)
	}

	return &universalDeployment{
		name:          name,
		container:     container,
		trustDomain:   opts.trustDomain,
		dockerBackend: cluster.(*framework.UniversalCluster).GetDockerBackend(),
	}
}

func (u *universalDeployment) Name() string {
	return u.name
}

func (u *universalDeployment) GetTrustDomain() string {
	return u.trustDomain
}

func (u *universalDeployment) GetIP() (string, error) {
	return u.container.GetIP()
}

func (u *universalDeployment) GetContainerID() string {
	return u.container.GetID()
}

func (u *universalDeployment) Deploy(cluster framework.Cluster) error {
	if err := u.container.Run(); err != nil {
		return err
	}

	return u.waitTillReady(cluster.GetTesting())
}

func (u *universalDeployment) Delete(cluster framework.Cluster) error {
	return u.container.Stop()
}

func (u *universalDeployment) waitTillReady(t testing.TestingT) error {
	containerID := u.container.GetID()

	retry.DoWithRetry(t, "health-check "+containerID, framework.DefaultRetries, framework.DefaultTimeout,
		func() (string, error) {
			// Check Spire server health endpoint - just check status code
			args := []string{
				"exec",
				containerID,
				"curl",
				"-s",
				"-o", "/dev/null",
				"-w", "%{http_code}",
				fmt.Sprintf("http://localhost:%d/ready", SpireHealthPort),
			}

			out, err := u.dockerBackend.RunCommandAndGetStdOutE(t, "exec", args, logger.Discard)
			if err != nil {
				return "Spire server health check failed", fmt.Errorf("curl command failed: %w", err)
			}

			// Check if we got HTTP 200
			statusCode := strings.TrimSpace(out)
			if statusCode != "200" {
				return "Spire server not ready", fmt.Errorf("got HTTP status %s, expected 200", statusCode)
			}

			return "Spire server is ready", nil
		})

	return nil
}

// RegisterWorkload registers a workload entry in Spire server
func (u *universalDeployment) RegisterWorkload(cluster framework.Cluster, spiffeID, parentID, selector string) error {
	args := []string{
		"exec",
		u.container.GetID(),
		"spire-server",
		"entry", "create",
		"-spiffeID", spiffeID,
		"-parentID", parentID,
		"-selector", selector,
	}

	_, err := u.dockerBackend.RunCommandAndGetStdOutE(cluster.GetTesting(), "exec", args, logger.Discard)
	return err
}

// GetAgentJoinToken generates a join token for a Spire agent
func (u *universalDeployment) GetAgentJoinToken(cluster framework.Cluster, agentSpiffeID string) (string, error) {
	args := []string{
		"exec",
		u.container.GetID(),
		"spire-server",
		"token", "generate",
		"-spiffeID", agentSpiffeID,
	}

	out, err := u.dockerBackend.RunCommandAndGetStdOutE(cluster.GetTesting(), "exec", args, logger.Discard)
	if err != nil {
		return "", fmt.Errorf("failed to generate join token: %w", err)
	}

	// Parse the token from output format: "Token: <token-value>"
	out = strings.TrimSpace(out)
	if !strings.HasPrefix(out, "Token: ") {
		return "", fmt.Errorf("unexpected token output format: %s", out)
	}

	token := strings.TrimPrefix(out, "Token: ")
	token = strings.TrimSpace(token)

	if token == "" {
		return "", fmt.Errorf("generated token is empty")
	}

	return token, nil
}

// ExecSpireServerCommand executes a spire-server command in the container
func (u *universalDeployment) ExecSpireServerCommand(cluster framework.Cluster, cmdArgs ...string) (string, error) {
	args := []string{
		"exec",
		u.container.GetID(),
		"spire-server",
	}
	args = append(args, cmdArgs...)

	out, err := u.dockerBackend.RunCommandAndGetStdOutE(cluster.GetTesting(), "exec", args, logger.Discard)
	return strings.TrimSpace(out), err
}

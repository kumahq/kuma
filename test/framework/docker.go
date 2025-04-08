package framework

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/random"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework/universal"
)

const DockerNetworkName = "kind"

type DockerBackend interface {
	// RunAndGetIDE runs the 'docker run' command on the given image with the given options and returns the container ID
	// that is returned in stdout, or any error.
	RunAndGetIDE(t testing.TestingT, image string, options *docker.RunOptions) (string, error)

	// StopE runs the 'docker stop' command for the given containers and returns any errors.
	StopE(t testing.TestingT, containers []string, options *docker.StopOptions) (string, error)

	GetPublishedDockerPorts(t testing.TestingT, l *logger.Logger, container string, ports []uint32) (map[uint32]uint32, error)

	RunCommandAndGetStdOutE(t testing.TestingT, cmdName string, args []string, log *logger.Logger) (string, error)
}

type LocalDockerBackend struct{}

func (*LocalDockerBackend) RunAndGetIDE(t testing.TestingT, image string, options *docker.RunOptions) (string, error) {
	return docker.RunAndGetIDE(t, image, options)
}

// StopE runs the 'docker stop' command for the given containers and returns any errors.
func (*LocalDockerBackend) StopE(t testing.TestingT, containers []string, options *docker.StopOptions) (string, error) {
	return docker.StopE(t, containers, options)
}

func (*LocalDockerBackend) GetPublishedDockerPorts(
	t testing.TestingT,
	l *logger.Logger,
	container string,
	ports []uint32,
) (map[uint32]uint32, error) {
	result := map[uint32]uint32{}
	for _, port := range ports {
		cmd := shell.Command{
			Command: "docker",
			Args:    []string{"port", container, strconv.Itoa(int(port))},
			Logger:  l,
		}
		var out string
		var err error
		// Sometimes the port may not be available immediately, and it can take some time.
		// Since we didn't retry, tests were failing with and an error
		// `missing port in address` on OSX.
		for i := 0; i < 10; i++ {
			out, err = shell.RunCommandAndGetStdOutE(t, cmd)
			if err != nil {
				time.Sleep(time.Millisecond * 500)
			}
			if out != "" {
				break
			}
		}
		if err != nil {
			return nil, err
		}
		addresses := strings.Split(out, "\n")
		if len(addresses) < 1 {
			return nil, errors.Errorf("there are no addresses for port %d", port)
		}
		addr := addresses[0]
		// on CircleCI, we get the ipv6 address in the format of ":::port",
		// which is not parsable by the "net.SplitHostPort"
		if strings.HasPrefix(addr, ":::") {
			addr = "[::]:" + addr[3:]
		}
		_, pubPortStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		pubPort, _ := strconv.ParseInt(pubPortStr, 10, 32)
		result[port] = uint32(pubPort)
	}
	return result, nil
}

func (*LocalDockerBackend) RunCommandAndGetStdOutE(t testing.TestingT, cmdName string, args []string, log *logger.Logger) (string, error) {
	cmd := shell.Command{
		Command: "docker",
		Args:    args,
		Logger:  log,
	}

	return shell.RunCommandAndGetStdOutE(t, cmd)
}

type VmPortForward struct {
	StopChannel   chan struct{}
	LocalPort     int
	RemoteAddress string
}

type RemoteDockerBackend struct {
	networking   *universal.Networking
	portForwards map[string][]*VmPortForward
}

func (u *RemoteDockerBackend) RunAndGetIDE(t testing.TestingT, image string, options *docker.RunOptions) (string, error) {
	if u.networking.RemoteHost == nil || u.networking.RemoteHost.Address == "" {
		return "", errors.New("RemoteHost is not set for the RemoteDockerBackend")
	}

	if len(options.Volumes) > 0 {
		files := make(map[string]string)
		uploadDir := fmt.Sprintf("/tmp/smoke-%s", random.UniqueId())

		for i := 0; i < len(options.Volumes); i++ {
			parts := strings.Split(options.Volumes[i], ":")
			if len(parts) != 2 {
				continue
			}

			hostPath := parts[0]
			mountPath := parts[1]
			absPath, err := filepath.Abs(hostPath)
			if err != nil {
				continue
			}
			stat, err := os.Stat(absPath)
			if err != nil {
				continue
			}
			if stat.IsDir() {
				continue
			}

			uploadPath := fmt.Sprintf("%s/%s", uploadDir, filepath.Base(absPath))
			files[absPath] = uploadPath
			options.Volumes[i] = fmt.Sprintf("%s:%s", uploadPath, mountPath)
		}

		if len(files) > 0 {
			err := u.networking.CopyFiles(t, files)
			if err != nil {
				return "", err
			}
		}
	}

	cmd := []string{"docker"}
	args := formatDockerRunArgs(image, options)
	cmd = append(cmd, args...)

	return u.runDockerCommandInSession(t, "run", strings.Join(cmd, " "), options.Logger)
}

// StopE runs the 'docker stop' command for the given containers and returns any errors.
func (u *RemoteDockerBackend) StopE(t testing.TestingT, containers []string, options *docker.StopOptions) (string, error) {
	if u.networking.RemoteHost == nil || u.networking.RemoteHost.Address == "" {
		return "", errors.New("RemoteHost is not set for the RemoteDockerBackend")
	}

	cmd := []string{"docker"}
	args := formatDockerStopArgs(containers, options)
	cmd = append(cmd, args...)
	strings.Join(cmd, " ")

	for _, container := range containers {
		pfList, exists := u.portForwards[container]
		if exists {
			for _, vmPortForward := range pfList {
				close(vmPortForward.StopChannel)
			}
			delete(u.portForwards, container)
		}
	}

	return u.runDockerCommandInSession(t, "stop", strings.Join(cmd, " "), options.Logger)
}

func (u *RemoteDockerBackend) runDockerCommandInSession(t testing.TestingT, cmdName, cmd string, log *logger.Logger) (string, error) {
	log.Logf(t, "Running 'docker %s' on remote host '%s', returning stdout: %q", cmdName, u.networking.RemoteHost.Address, cmd)
	sshSession, err := u.networking.NewSession(path.Join("docker-commands", u.networking.RemoteHost.Address), cmdName, false, cmd)
	if err != nil {
		return "", err
	}
	err = sshSession.Run()
	_ = sshSession.Close()
	if err != nil {
		return "", err
	}

	stdOut, err := os.ReadFile(sshSession.StdOutFile())
	if err != nil {
		return "", fmt.Errorf("failed to read stdout of docker %s command: %w", cmdName, err)
	}
	out := string(stdOut)
	return strings.TrimRight(out, "\n"), nil
}

func (u *RemoteDockerBackend) GetPublishedDockerPorts(
	t testing.TestingT,
	log *logger.Logger,
	container string,
	ports []uint32,
) (map[uint32]uint32, error) {
	listPortsCmd := []string{"docker", "port", container}
	out, err := u.runDockerCommandInSession(t, "port", strings.Join(listPortsCmd, " "), log)
	if err != nil {
		// Sometimes the port may not be available immediately, and it can take some time.
		// Since we didn't retry, tests were failing with and an error
		// `missing port in address` on OSX.
		time.Sleep(time.Millisecond * 500)
		return nil, err
	}

	allPortMappings := strings.Split(out, "\n")

	result := map[uint32]uint32{}
	for _, port := range ports {
		// example output format:
		// 30094/tcp -> 0.0.0.0:40094
		// 30094/tcp -> [::]:40094

		mappingPrefix := fmt.Sprintf("%d/tcp", port)
		var addresses []string
		for _, portMapping := range allPortMappings {
			if strings.HasPrefix(portMapping, mappingPrefix) {
				parts := strings.Split(portMapping, " ")
				addresses = append(addresses, parts[len(parts)-1])
			}
		}

		if len(addresses) < 1 {
			return nil, errors.Errorf("there are no addresses for port %d", port)
		}
		addr := addresses[0]
		// on CircleCI (or other possible hosts), we get the ipv6 address in the format of ":::port",
		// which is not parsable by the "net.SplitHostPort"
		if strings.HasPrefix(addr, ":::") {
			addr = "[::]:" + addr[3:]
		}
		_, pubPortStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		// setup SSH port forwarding for the remote port
		pf, err := portForward(u.networking, net.JoinHostPort("127.0.0.1", pubPortStr))
		if err != nil {
			return nil, err
		}
		pfList, exists := u.portForwards[container]
		if !exists {
			if u.portForwards == nil {
				u.portForwards = make(map[string][]*VmPortForward)
			}
			u.portForwards[container] = []*VmPortForward{}
		}
		pfList = append(pfList, pf)
		u.portForwards[container] = pfList
		result[port] = uint32(pf.LocalPort)
	}

	return result, nil
}

func (u *RemoteDockerBackend) RunCommandAndGetStdOutE(t testing.TestingT, cmdName string, args []string, log *logger.Logger) (string, error) {
	args = append(args, "docker")
	return u.runDockerCommandInSession(t, cmdName, strings.Join(args, " "), log)
}

func portForward(networking *universal.Networking, remoteAddress string) (*VmPortForward, error) {
	stopChan := make(chan struct{})
	addr, err := networking.PortForward(remoteAddress, stopChan)
	if err != nil {
		return nil, fmt.Errorf("could not establish ssh port forwarding to remote host %q: %w",
			networking.RemoteHost, err)
	}

	pf := &VmPortForward{
		StopChannel:   stopChan,
		LocalPort:     addr.(*net.TCPAddr).Port,
		RemoteAddress: remoteAddress,
	}
	return pf, nil
}

// FormatDockerRunArgs formats the arguments for the 'docker run' command.
// it is taken from terratest/modules/docker
func formatDockerRunArgs(image string, options *docker.RunOptions) []string {
	args := []string{"run"}

	if options.Detach {
		args = append(args, "--detach")
	}

	if options.Entrypoint != "" {
		args = append(args, "--entrypoint", options.Entrypoint)
	}

	for _, envVar := range options.EnvironmentVariables {
		args = append(args, "--env", envVar)
	}

	if options.Init {
		args = append(args, "--init")
	}

	if options.Name != "" {
		args = append(args, "--name", options.Name)
	}

	if options.Platform != "" {
		args = append(args, "--platform", options.Platform)
	}

	if options.Privileged {
		args = append(args, "--privileged")
	}

	if options.Remove {
		args = append(args, "--rm")
	}

	if options.Tty {
		args = append(args, "--tty")
	}

	if options.User != "" {
		args = append(args, "--user", options.User)
	}

	for _, volume := range options.Volumes {
		args = append(args, "--volume", volume)
	}

	args = append(args, options.OtherOptions...)

	args = append(args, image)

	args = append(args, options.Command...)

	return args
}

// FormatDockerStopArgs formats the arguments for the 'docker stop' command
// it is taken from terratest/modules/docker
func formatDockerStopArgs(containers []string, options *docker.StopOptions) []string {
	args := []string{"stop"}

	if options.Time != 0 {
		args = append(args, "--time", strconv.Itoa(options.Time))
	}

	args = append(args, containers...)

	return args
}

var _ = []DockerBackend{
	&LocalDockerBackend{},
	&RemoteDockerBackend{},
}

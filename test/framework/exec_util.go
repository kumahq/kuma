package framework

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"

	"github.com/gruntwork-io/terratest/modules/k8s"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	restclient "k8s.io/client-go/rest"

	. "github.com/onsi/gomega"
)

// inspired by https://github.com/kubernetes/kubernetes/blob/v1.6.1/test/e2e/framework/exec_util.go

// ExecOptions passed to ExecWithOptions
type ExecOptions struct {
	Command []string

	Namespace     string
	PodName       string
	ContainerName string

	Stdin         io.Reader
	CaptureStdout bool
	CaptureStderr bool
	// If false, whitespace in std{err,out} will be removed.
	PreserveWhitespace bool
}

// ExecWithOptions executes a command in the specified container,
// returning stdout, stderr and error. `options` allowed for
// additional parameters to be passed.
func (c *K8sCluster) ExecWithOptions(options ExecOptions) (string, string, error) {
	const tty = false
	config, err := k8s.LoadApiClientConfigE(c.kubeconfig, "")
	Expect(err).NotTo(HaveOccurred())

	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(options.PodName).
		Namespace(options.Namespace).
		SubResource("exec").
		Param("container", options.ContainerName)

	req.VersionedParams(&kube_core.PodExecOptions{
		Container: options.ContainerName,
		Command:   options.Command,
		Stdin:     true,
		Stdout:    options.CaptureStdout,
		Stderr:    options.CaptureStderr,
		TTY:       tty,
	}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	err = execute("POST", req.URL(), config, strings.NewReader(""), &stdout, &stderr, tty)

	if options.PreserveWhitespace {
		return stdout.String(), stderr.String(), err
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

// Exec executes a command in the
// specified container and return stdout, stderr and error
func (c *K8sCluster) Exec(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	return c.ExecWithOptions(ExecOptions{
		Command:       cmd,
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,

		Stdin:              nil,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
	})
}

func (c *K8sCluster) ExecWithRetries(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	var stdout string
	var stderr string
	_, err := retry.DoWithRetryE(
		c.t,
		fmt.Sprintf("kubectl exec -- %s", strings.Join(cmd, " ")),
		DefaultRetries,
		DefaultTimeout,
		func() (string, error) {
			var err error
			stdout, stderr, err = c.Exec(namespace, podName, containerName, cmd...)
			return "", err
		},
	)
	return stdout, stderr, err
}

func execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}

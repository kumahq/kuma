package framework

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
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

	Retries int
	Timeout time.Duration
}

// ExecWithOptions executes a command in the specified container,
// returning stdout, stderr and error. `options` allowed for
// additional parameters to be passed.
func (c *K8sCluster) ExecWithOptions(options ExecOptions) (string, string, error) {
	kubectlExec := func() (string, string, error) {
		const tty = false
		config, err := k8s.LoadApiClientConfigE(c.kubeconfig, "")
		Expect(err).NotTo(HaveOccurred())

		clientset, err := k8s.GetKubernetesClientFromOptionsE(c.t, c.GetKubectlOptions())
		Expect(err).NotTo(HaveOccurred())

		req := clientset.CoreV1().RESTClient().Post().
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
		err = executeK8s("POST", req.URL(), config, strings.NewReader(""), &stdout, &stderr, tty)

		if options.PreserveWhitespace {
			return stdout.String(), stderr.String(), err
		}

		return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
	}

	var stdout string
	var stderr string
	_, err := retry.DoWithRetryE(
		c.t,
		fmt.Sprintf("kubectl exec -c %q -n %q %s -- %s",
			options.ContainerName,
			options.Namespace,
			options.PodName,
			strings.Join(options.Command, " ")),
		options.Retries,
		options.Timeout,
		func() (string, error) {
			var err error
			stdout, stderr, err = kubectlExec()
			return "", err
		},
	)
	return stdout, stderr, err
}

// Exec executes a command in the specified container and return stdout,
// stderr and error.
func (c *K8sCluster) Exec(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	return c.ExecWithOptions(ExecOptions{
		Command:            cmd,
		Namespace:          namespace,
		PodName:            podName,
		ContainerName:      containerName,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
		Retries:            0,
	})
}

// ExecWithRetries executes a command in the specified container and
// return stdout, stderr and error. It retries a default number of times
// if the command fails.
func (c *K8sCluster) ExecWithRetries(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	return c.ExecWithOptions(ExecOptions{
		Command:            cmd,
		Namespace:          namespace,
		PodName:            podName,
		ContainerName:      containerName,
		CaptureStdout:      true,
		CaptureStderr:      true,
		PreserveWhitespace: false,
		Retries:            DefaultRetries,
		Timeout:            DefaultTimeout,
	})
}

func executeK8s(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
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

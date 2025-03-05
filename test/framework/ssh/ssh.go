package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
	k8s_strings "k8s.io/utils/strings"

	"github.com/kumahq/kuma/pkg/util/files"
	"github.com/kumahq/kuma/test/framework/report"
	"github.com/kumahq/kuma/test/framework/utils"
)

// logf logs a test progress message.
func logf(format string, args ...interface{}) {
	logger.Default.Logf(ginkgo.GinkgoT(), format, args...)
}

type App struct {
	containerName string
	cmd           *exec.Cmd
	stdin         bytes.Buffer
	stdoutFile    string
	stderrFile    string
	port          string
	clusterName   string
	args          []string
}

func NewApp(containerName string, cluster string, verbose bool, port string, envMap map[string]string, args []string) *App {
	app := &App{
		containerName: containerName,
		port:          port,
		clusterName:   cluster,
		args:          args,
	}
	sshArgs := []string{
		"-q", "-tt",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"root@localhost", "-p", port,
	}
	for k, v := range envMap {
		sshArgs = append(sshArgs, fmt.Sprintf("%s=%s", k, utils.ShellEscape(v)))
	}

	sshArgs = append(sshArgs, args...)
	app.cmd = exec.Command("ssh", sshArgs...)

	sout, _ := os.CreateTemp(os.TempDir(), "ssh_out")
	serr, _ := os.CreateTemp(os.TempDir(), "ssh_err")
	app.stderrFile = serr.Name()
	app.stdoutFile = sout.Name()

	outWriters := []io.Writer{sout}
	errWriters := []io.Writer{serr}

	if verbose {
		outWriters = append(outWriters, os.Stdout)
		errWriters = append(errWriters, os.Stderr)
	}

	app.cmd.Stdout = io.MultiWriter(outWriters...)
	app.cmd.Stderr = io.MultiWriter(errWriters...)
	app.cmd.Stdin = &app.stdin

	return app
}

func (s *App) done() {
	base := path.Join(s.clusterName, "universal", "ssh", s.containerName, k8s_strings.ShortenString(files.ToValidUnixFilename(strings.Join(s.args, "_")), 64), strconv.Itoa(s.cmd.Process.Pid))
	report.AddFileToReportEntry(path.Join(base, "cmd-debug.txt"), fmt.Sprintf("cmd:%s\nargs:%s\ncontainer:%s\n", s.cmd.String(), s.args, s.containerName))
	report.AddFileToReportEntry(path.Join(base, "std-out.log"), s.Out())
	report.AddFileToReportEntry(path.Join(base, "std-err.log"), s.Err())
}

func (s *App) Run() error {
	defer s.done()
	logf("Running %v", s.cmd)
	return s.cmd.Run()
}

func (s *App) Signal(signal syscall.Signal, wait bool) error {
	defer s.done()
	logf("Signaling %s %v", signal, s.cmd)
	if err := s.cmd.Process.Signal(signal); err != nil {
		return err
	}
	if wait {
		logf("Waiting %v", s.cmd)
		_, err := s.cmd.Process.Wait()
		return err
	}
	return nil
}

func (s *App) Start() error {
	logf("Starting %v", s.cmd)
	return s.cmd.Start()
}

func (s *App) Out() string {
	r, err := os.ReadFile(s.stdoutFile)
	if err != nil {
		return fmt.Sprintf("failed to read stderr file: %v", err)
	}
	return string(r)
}

func (s *App) Err() string {
	r, err := os.ReadFile(s.stderrFile)
	if err != nil {
		return fmt.Sprintf("failed to read stderr file: %v", err)
	}
	return string(r)
}

func (s *App) StdOutFile() string {
	return s.stdoutFile
}

func (s *App) StdErrFile() string {
	return s.stderrFile
}

package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework/utils"
)

// Logf logs a test progress message.
func Logf(format string, args ...interface{}) {
	logger.Default.Logf(ginkgo.GinkgoT(), format, args...)
}

type App struct {
	cmd     *exec.Cmd
	stdin   bytes.Buffer
	stdout  bytes.Buffer
	stderr  bytes.Buffer
	port    string
	logFile *os.File
}

func NewApp(appName string, logsPath string, verbose bool, port string, envMap map[string]string, args []string) *App {
	app := &App{
		port: port,
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
	err := os.MkdirAll(logsPath, os.ModePerm)
	if err != nil {
		panic(errors.Wrap(err, "could not create /tmp/e2e"))
	}
	logFileName := path.Join(logsPath, appName)
	app.logFile, err = os.OpenFile(logFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0o660)
	if err != nil {
		panic(errors.Wrap(err, "could not create "+logFileName))
	}
	outWriters := []io.Writer{&app.stdout, app.logFile}
	errWriters := []io.Writer{&app.stderr, app.logFile}

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
	s.logFile.Close()
}

func (s *App) Run() error {
	defer s.done()
	Logf("Running %v", s.cmd)
	return s.cmd.Run()
}

func (s *App) Signal(signal syscall.Signal, wait bool) error {
	defer s.done()
	Logf("Signaling %s %v", signal, s.cmd)
	if err := s.cmd.Process.Signal(signal); err != nil {
		return err
	}
	if wait {
		Logf("Waiting %v", s.cmd)
		_, err := s.cmd.Process.Wait()
		return err
	}
	return nil
}

func (s *App) Start() error {
	Logf("Starting %v", s.cmd)
	return s.cmd.Start()
}

func (s *App) Out() string {
	return s.stdout.String()
}

func (s *App) Err() string {
	return s.stderr.String()
}

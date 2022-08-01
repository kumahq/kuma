package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

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

func NewApp(appName string, verbose bool, port string, envMap map[string]string, args []string) *App {
	app := &App{
		port: port,
	}
	var env []string
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s='%s'", k, utils.ShellEscape(v)))
	}
	sshArgs := append(
		[]string{
			"-q", "-tt",
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"root@localhost", "-p", port,
		}, env...,
	)
	sshArgs = append(sshArgs, args...)
	app.cmd = exec.Command("ssh", sshArgs...)

	inWriters := []io.Reader{&app.stdin}
	outWriters := []io.Writer{&app.stdout}
	errWriters := []io.Writer{&app.stderr}

	logDirName := "/tmp/e2e"
	err := os.MkdirAll(logDirName, os.ModePerm)
	if err != nil {
		panic(errors.Wrap(err, "could not create /tmp/e2e"))
	}
	logFileName := logDirName + "/" + appName
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		panic(errors.Wrap(err, "could not create "+logFileName))
	}
	outWriters = append(outWriters, f)
	errWriters = append(errWriters, f)

	if verbose {
		outWriters = append(outWriters, os.Stdout)
		errWriters = append(errWriters, os.Stderr)
	}
	app.cmd.Stdout = io.MultiWriter(outWriters...)
	app.cmd.Stderr = io.MultiWriter(errWriters...)
	app.cmd.Stdin = io.MultiReader(inWriters...)
	return app
}

func (s *App) Run() error {
	Logf("Running %v", s.cmd)
	return s.cmd.Run()
}

func (s *App) Kill() error {
	Logf("Killing %v", s.cmd)
	return s.cmd.Process.Kill()
}

func (s *App) ProcessWait() error {
	Logf("Waiting %v", s.cmd)

	_, err := s.cmd.Process.Wait()

	return err
}

func (s *App) Start() error {
	Logf("Starting %v", s.cmd)
	return s.cmd.Start()
}

func (s *App) Stop() error {
	if err := s.cmd.Process.Kill(); err != nil {
		return err
	}
	if _, err := s.cmd.Process.Wait(); err != nil {
		return err
	}
	if s.logFile != nil {
		if err := s.logFile.Close(); err != nil {
			return errors.Wrapf(err, "could not close the file %s", s.logFile.Name())
		}
	}
	return nil
}

func (s *App) Wait() error {
	return s.cmd.Wait()
}

func (s *App) Out() string {
	return s.stdout.String()
}

func (s *App) Err() string {
	return s.stderr.String()
}

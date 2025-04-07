package ssh

import (
	"fmt"
	"io"
	"os"
	"path"

	"golang.org/x/crypto/ssh"

	"github.com/kumahq/kuma/test/framework/report"
)

type Session struct {
	reportPath string
	session    *ssh.Session
	name       string
	stderrFile string
	stdoutFile string
	cmd        string
}

func NewSession(client *ssh.Client, reportPath string, name string, verbose bool, cmd string) (*Session, error) {
	session, err := client.NewSession()
	if err != nil {
		panic(err)
	}
	sout, _ := os.CreateTemp(os.TempDir(), "ssh_out")
	serr, _ := os.CreateTemp(os.TempDir(), "ssh_err")

	outWriters := []io.Writer{sout}
	errWriters := []io.Writer{serr}

	if verbose {
		outWriters = append(outWriters, os.Stdout)
		errWriters = append(errWriters, os.Stderr)
	}
	session.Stderr = io.MultiWriter(errWriters...)
	session.Stdout = io.MultiWriter(outWriters...)

	return &Session{
		reportPath: reportPath,
		name:       name,
		session:    session,
		stderrFile: serr.Name(),
		stdoutFile: sout.Name(),
		cmd:        cmd,
	}, nil
}

func (s *Session) Close() error {
	return s.session.Close()
}

func (s *Session) Start() error {
	err := s.session.Start(s.cmd)
	if err != nil {
		return err
	}
	base := path.Join(s.reportPath, s.name)
	report.AddFileToReportEntry(path.Join(base, "cmd-debug.txt"), fmt.Sprintf("cmd:%s\n", s.cmd))
	report.AddFileToReportEntry(path.Join(base, "std-out.log"), s.stdoutFile)
	report.AddFileToReportEntry(path.Join(base, "std-err.log"), s.stderrFile)
	return nil
}

func (s *Session) Wait() error {
	return s.session.Wait()
}

func (s *Session) Signal(signal ssh.Signal, wait bool) error {
	if err := s.session.Signal(signal); err != nil {
		if err != io.EOF {
			return err
		}
		// EOF is expected when the process is already finished
		return nil
	}
	if wait {
		return s.Wait()
	}
	return nil
}

func (s *Session) Run() error {
	if err := s.Start(); err != nil {
		return err
	}
	return s.Wait()
}

func (s *Session) StdOutFile() string {
	return s.stdoutFile
}

func (s *Session) StdErrFile() string {
	return s.stderrFile
}

package universal

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	kssh "github.com/kumahq/kuma/test/framework/ssh"
)

type NetworkingState struct {
	ZoneEgress  Networking `json:"zoneEgress"`
	ZoneIngress Networking `json:"zoneIngress"`
	KumaCp      Networking `json:"kumaCp"`
}

type Networking struct {
	IP            string `json:"ip"` // IP inside a docker network
	ApiServerPort string `json:"apiServerPort"`
	SshPort       string `json:"sshPort"`
	StdOutFile    string `json:"stdOutFile"`
	StdErrFile    string `json:"stdErrFile"`
	sshClient     *ssh.Client
	id            int
	sync.Mutex
}

func (s *Networking) initSSH() (int, error) {
	s.Lock()
	defer s.Unlock()
	s.id++
	if s.sshClient == nil {
		for i := 0; ; i++ {
			if i == 10 {
				return s.id, errors.New("failed to connect to container")
			}
			client, err := ssh.Dial("tcp", net.JoinHostPort("localhost", s.SshPort), &ssh.ClientConfig{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(), //#nosec G106 // skip for tests
				User:            "root",
			})
			if err == nil {
				s.sshClient = client
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	return s.id, nil
}

func (s *Networking) RunCommand(cmd string) (string, string, error) {
	_, err := s.initSSH()
	if err != nil {
		return "", "", err
	}

	session, err := s.sshClient.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	// Set up command output
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Execute complex command - no need to worry about escaping
	// since you're passing directly to the remote shell
	err = session.Run(cmd)
	return stdout.String(), stderr.String(), err
}

func (s *Networking) Close() error {
	return s.sshClient.Close()
}

func (s *Networking) NewSession(exportPath string, cmdName string, verbose bool, cmd string) (*kssh.Session, error) {
	id, err := s.initSSH()
	if err != nil {
		return nil, err
	}
	session, err := kssh.NewSession(s.sshClient, exportPath, fmt.Sprintf("%s-%d", cmdName, id), verbose, cmd)
	if err != nil {
		return nil, err
	}
	s.StdErrFile = session.StdErrFile()
	s.StdOutFile = session.StdOutFile()
	return session, nil
}

func (u *Networking) BootstrapAddress() string {
	if u.ApiServerPort == "" {
		panic("ApiServerPort is not set, this networking is not for a CP")
	}
	return "https://" + net.JoinHostPort(u.IP, "5678")
}

package universal

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
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
	SshPort       string `json:"sshPort"` // SshPort is the local port that is forwarded into the docker container
	StdOutFile    string `json:"stdOutFile"`
	StdErrFile    string `json:"stdErrFile"`
	sshClient     *ssh.Client
	id            int

	RemoteHost *kssh.Host `json:"-"` // RemoteHost is a remote SSH target that can be directly connected to
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

			var client *ssh.Client
			var err error

			if s.RemoteHost != nil {
				var pemBytes []byte
				if len(s.RemoteHost.PrivateKeyData) > 0 {
					pemBytes = s.RemoteHost.PrivateKeyData
				} else if s.RemoteHost.PrivateKeyFile != "" {
					privKeyBytes, err := os.ReadFile(s.RemoteHost.PrivateKeyFile)
					if err != nil {
						return s.id, fmt.Errorf("failed to read ssh private key file: %w", err)
					}
					pemBytes = privKeyBytes
				}
				signer, e := ssh.ParsePrivateKey(pemBytes)
				if e != nil {
					return s.id, fmt.Errorf("failed to parse ssh private key: %w", err)
				}

				configCfg := &ssh.ClientConfig{
					User: s.RemoteHost.User,
					Auth: []ssh.AuthMethod{
						ssh.PublicKeys(signer),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //#nosec G106 // skip for tests
				}

				client, err = ssh.Dial("tcp", net.JoinHostPort(s.RemoteHost.Address,
					strconv.Itoa(s.RemoteHost.Port)), configCfg)
			} else {
				client, err = ssh.Dial("tcp", net.JoinHostPort("localhost", s.SshPort), &ssh.ClientConfig{
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //#nosec G106 // skip for tests
					User:            "root",
				})
			}
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

func (s *Networking) PortForward(local, remote string, stopChan <-chan struct{}) (net.Addr, error) {
	errorChan := make(chan error)
	readyChan := make(chan net.Addr)

	go func() {
		err := kssh.Tunnel(s.sshClient, local, remote, stopChan, readyChan)
		if err != nil {
			errorChan <- err
		}
		close(errorChan)
	}()

	select {
	case err := <-errorChan:
		return nil, err
	case listenAddr := <-readyChan:
		return listenAddr, nil
	}
}

func (u *Networking) BootstrapAddress() string {
	if u.ApiServerPort == "" {
		panic("ApiServerPort is not set, this networking is not for a CP")
	}
	return "https://" + net.JoinHostPort(u.IP, "5678")
}

package ssh

import (
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func Tunnel(sshClient *ssh.Client, local, remote string, stopChan <-chan struct{}, readyChan chan<- net.Addr) error {
	pipe := func(writer, reader net.Conn) {
		defer writer.Close()
		defer reader.Close()

		_, err := io.Copy(writer, reader)
		if err != nil {
			log.Printf("port forward failed: %s", err)
		}
	}

	listener, err := net.Listen("tcp", local)
	if err != nil {
		return err
	}

	if readyChan != nil {
		readyChan <- listener.Addr()
	}

	for {
		select {
		case <-stopChan:
			return nil
		default:
			localConnection, err := listener.Accept()
			if err != nil {
				return err
			}

			go func(localConn net.Conn) {
				remoteConn, err := sshClient.Dial("tcp", remote)
				if err != nil {
					log.Fatalf("failed to dial to ssh remote %q on host %s: %q",
						remote, sshClient.RemoteAddr().String(), err)
				}

				go pipe(remoteConn, localConn)
				go pipe(localConn, remoteConn)
			}(localConnection)
		}
	}
}

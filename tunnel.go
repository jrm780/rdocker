package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

const dockerSocket = "/var/run/docker.sock"

// SSHTunnel forwards connections to a remote host through SSH
type SSHTunnel struct {
	Endpoint string
	Config   *ssh.ClientConfig
}

// Start opens the ssh tunnel and forwards connections through it
func (tunnel *SSHTunnel) Start(listener net.Listener) error {
	serverConn, err := ssh.Dial("tcp", tunnel.Endpoint, tunnel.Config)
	if err != nil {
		return fmt.Errorf("server dial error: %v", err)
	}
	defer serverConn.Close()
	for {
		localConn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %v", err)
		}

		go tunnel.handleConnection(localConn, serverConn)
	}
}

func (tunnel *SSHTunnel) handleConnection(localConn net.Conn, s *ssh.Client) {
	defer localConn.Close()

	remoteConn, err := s.Dial("unix", dockerSocket)
	if err != nil {
		fmt.Printf("remote dial error: %v", err)
		return
	}
	defer remoteConn.Close()

	tunnel.forward(localConn, remoteConn)
}

func (tunnel *SSHTunnel) forward(localConn, remoteConn net.Conn) error {
	var g errgroup.Group
	g.Go(pipe(localConn, remoteConn))
	g.Go(pipe(remoteConn, localConn))
	return g.Wait()
}

func pipe(w io.Writer, r io.Reader) func() error {
	return func() error {
		_, err := io.Copy(w, r)
		if err != io.EOF {
			return err
		}
		return nil
	}
}

func getSSHKey(key []byte) ssh.AuthMethod {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	return ssh.PublicKeys(signer)
}

func getSSHKeyFromFile(filename string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	return getSSHKey(key)
}

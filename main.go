package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path"

	"github.com/ianschenck/envflag"
	"golang.org/x/crypto/ssh"
)

func main() {

	u, err := user.Current()
	if err != nil {
		u = &user.User{Username: "root"}
	}
	defaultKeyPath := path.Join(u.HomeDir, ".ssh", "id_rsa")

	var (
		host        = envflag.String("RD_HOST", "", "The remote host with which to establish a tunnel (REQUIRED)")
		user        = envflag.String("RD_USER", "root", "The user with which to establish a tunnel")
		privKey     = envflag.String("RD_PRIV_KEY", "", "The SSH private key used to establish a tunnel")
		privKeyFile = envflag.String("RD_PRIV_KEY_FILE", defaultKeyPath, "The path to the SSH private key used to establish a tunnel")
	)
	envflag.Parse()

	if *host == "" {
		termUsage()
	}

	var auth ssh.AuthMethod
	if *privKey == "" {
		auth = getSSHKeyFromFile(*privKeyFile)
	} else {
		auth = getSSHKey([]byte(*privKey))
	}

	tunnel := &SSHTunnel{
		Endpoint: *host,
		Config: &ssh.ClientConfig{
			User: *user,
			Auth: []ssh.AuthMethod{
				auth,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
	defer listener.Close()

	go func() {
		err := tunnel.Start(listener)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}()

	err = runDocker(listener.Addr().String(), os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}

func runDocker(dockerHost string, arguments []string) error {
	cmd := exec.Command("docker", arguments...)
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+dockerHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func termUsage() {
	fmt.Println("The following environment variables may be set:")
	envflag.PrintDefaults()
	os.Exit(1)
}

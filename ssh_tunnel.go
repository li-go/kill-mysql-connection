package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHConfig struct {
	UseTunnel bool `toml:"use_tunnel"`

	Host     string
	Port     int
	Username string
	Password string

	PrivateKey    string `toml:"private_key"`
	KeyPassphrase string `toml:"key_passphrase"`
}

func OpenSSHClient(config SSHConfig) (*ssh.Client, error) {
	authMethods, err := sshAuthMethods(config)
	if err != nil {
		return nil, err
	}
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	log.Printf("Dialing SSH server - %s@%s:%d ...", config.Username, config.Host, config.Port)
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("fail to dial ssh host: %w", err)
	}
	return sshClient, nil
}

func sshAuthMethods(config SSHConfig) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}
	if config.PrivateKey != "" {
		key, err := readFile(config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("fail to read ssh private key: %w", err)
		}
		var signer ssh.Signer
		if config.KeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(config.KeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, fmt.Errorf("fail to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if config.Password == "" && config.PrivateKey == "" {
		sshAgent, err := sshAgent()
		if err != nil {
			return nil, fmt.Errorf("fail to access ssh sshAgent: %w", err)
		}
		authMethods = append(authMethods, sshAgent)
	}
	return authMethods, nil
}

func sshAgent() (ssh.AuthMethod, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers), nil
}

func readFile(path string) ([]byte, error) {
	path, err := absPath(path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(path)
}

func absPath(path string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", nil
	}
	return strings.Replace(path, "~", u.HomeDir, 1), nil
}

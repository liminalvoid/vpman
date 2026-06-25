package sshclient

import (
	"fmt"
	"time"
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Client struct {
	conn *ssh.Client
}

type Config struct {
	Host           string
	Port           int
	User           string
	PrivateKey     []byte // PEM encoded
	KnownHostsPath string
	Timeout        time.Duration
}

func (c *Client) New(cfg Config) (*Client, error) {
	signer, err := ssh.ParsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	hostKeyCallback, err := knownhosts.New(cfg.KnownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts: %w", err)
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port)), &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout: cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", cfg.Host, err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

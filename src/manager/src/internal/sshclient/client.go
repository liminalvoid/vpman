package sshclient

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"
	"errors"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Client struct {
	Host string
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

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func New(cfg Config) (*Client, error) {
	signer, err := ssh.ParsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	hostKeyCallback, err := knownhosts.New(cfg.KnownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts: %w", err)
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port)), &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", cfg.Host, err)
	}

	return &Client{conn: conn, Host: cfg.Host}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Run(ctx context.Context, cmd Command) (Result, error) {
	session, err := c.conn.NewSession()
	if err != nil {
		return Result{}, fmt.Errorf("new session: %w", err)
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Start(cmd.String()); err != nil {
		return Result{}, fmt.Errorf("start: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- session.Wait() }()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)

		return Result{}, ctx.Err()
	case err := <-done:
		res := Result{Stdout: stdout.String(), Stderr: stderr.String()}

		if exitErr, ok := errors.AsType[*ssh.ExitError](err); ok {
			res.ExitCode = exitErr.ExitStatus()

			return res, nil
		}

		if err != nil {
			return res, fmt.Errorf("wait: %w", err)
		}

		return res, nil
	}
}

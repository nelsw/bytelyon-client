package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

var (
	client *ssh.Client
	mu     sync.Mutex
)

// conn wraps an ssh channel, which returns "deadline not supported" errors, so clients that set deadlines still work.
type conn struct{ net.Conn }

func (conn) SetDeadline(time.Time) error      { return nil }
func (conn) SetReadDeadline(time.Time) error  { return nil }
func (conn) SetWriteDeadline(time.Time) error { return nil }

// connect lazily dials the SSH server on first use and reuses the connection thereafter.
func connect() (*ssh.Client, error) {
	mu.Lock()
	defer mu.Unlock()

	if client != nil {
		return client, nil
	}

	pemBytes, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	var signer ssh.Signer
	if signer, err = ssh.ParsePrivateKey(pemBytes); err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	cfg := ssh.ClientConfig{
		User:            os.Getenv("SERVER_USER"),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}
	if client, err = ssh.Dial("tcp", os.Getenv("SERVER_ADDR"), &cfg); err != nil {
		return nil, fmt.Errorf("failed to dial SSH server: %w", err)
	}
	return client, nil
}

// DialFunc tunnels connections through the SSH server at SERVER_ADDR, or dials directly when it is unset.
func DialFunc() func(context.Context, string, string) (net.Conn, error) {
	if os.Getenv("SERVER_ADDR") == "" {
		return (&net.Dialer{}).DialContext
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, err := connect()
		if err != nil {
			return nil, err
		}
		ch, err := c.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		return conn{ch}, nil
	}
}

func Close() {
	mu.Lock()
	defer mu.Unlock()

	if client == nil {
		return
	}
	if err := client.Close(); err != nil {
		log.Err(err).Send()
	}
	client = nil
}

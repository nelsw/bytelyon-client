package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

var client *ssh.Client

// conn wraps an ssh channel, which returns "deadline not supported" errors, so clients that set deadlines still work.
type conn struct{ net.Conn }

func (conn) SetDeadline(time.Time) error      { return nil }
func (conn) SetReadDeadline(time.Time) error  { return nil }
func (conn) SetWriteDeadline(time.Time) error { return nil }

func init() {

	pemBytes, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa")
	if err != nil {
		panic(fmt.Sprintf("Failed to read private key: %v", err))
	}

	var signer ssh.Signer
	if signer, err = ssh.ParsePrivateKey(pemBytes); err != nil {
		panic(fmt.Sprintf("Failed to parse private key: %v", err))
	}

	cfg := ssh.ClientConfig{
		User:            os.Getenv("SERVER_USER"),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}
	if client, err = ssh.Dial("tcp", os.Getenv("SERVER_ADDR"), &cfg); err != nil {
		panic(fmt.Sprintf("Failed to dial SSH server: %v", err))
	}

}

func DialFunc() func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, err := client.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		return conn{c}, nil
	}
}

func Close() {
	if err := client.Close(); err != nil {
		log.Err(err).Send()
	}
}

package ssh

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// listen returns a TCP listener on a random local port that is closed when the test ends.
func listen(t *testing.T) net.Listener {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l
}

// echoServer accepts connections and writes back whatever it reads.
func echoServer(t *testing.T) string {
	l := listen(t)
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go func() { _, _ = io.Copy(c, c); _ = c.Close() }()
		}
	}()
	return l.Addr().String()
}

// sshServer starts an SSH server that accepts any public key and forwards direct-tcpip channels.
func sshServer(t *testing.T) string {
	t.Helper()
	_, hostKey, _ := ed25519.GenerateKey(rand.Reader)
	signer, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &ssh.ServerConfig{
		PublicKeyCallback: func(ssh.ConnMetadata, ssh.PublicKey) (*ssh.Permissions, error) { return nil, nil },
	}
	cfg.AddHostKey(signer)

	l := listen(t)
	go func() {
		for {
			nc, err := l.Accept()
			if err != nil {
				return
			}
			go serve(nc, cfg)
		}
	}()
	return l.Addr().String()
}

func serve(nc net.Conn, cfg *ssh.ServerConfig) {
	_, chans, reqs, err := ssh.NewServerConn(nc, cfg)
	if err != nil {
		return
	}
	go ssh.DiscardRequests(reqs)
	for nch := range chans {
		var target struct {
			Host       string
			Port       uint32
			OriginHost string
			OriginPort uint32
		}
		if err = ssh.Unmarshal(nch.ExtraData(), &target); err != nil {
			_ = nch.Reject(ssh.ConnectionFailed, err.Error())
			continue
		}
		up, err := net.Dial("tcp", net.JoinHostPort(target.Host, strconv.Itoa(int(target.Port))))
		if err != nil {
			_ = nch.Reject(ssh.ConnectionFailed, err.Error())
			continue
		}
		ch, chReqs, err := nch.Accept()
		if err != nil {
			_ = up.Close()
			continue
		}
		go ssh.DiscardRequests(chReqs)
		go func() { _, _ = io.Copy(ch, up); _ = ch.Close() }()
		go func() { _, _ = io.Copy(up, ch); _ = up.Close() }()
	}
}

// writeKey writes a client private key to $HOME/.ssh/id_rsa under a temporary HOME.
func writeKey(t *testing.T, body []byte) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".ssh"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".ssh", "id_rsa"), body, 0o600); err != nil {
		t.Fatal(err)
	}
}

func clientKey(t *testing.T) []byte {
	t.Helper()
	_, key, _ := ed25519.GenerateKey(rand.Reader)
	block, err := ssh.MarshalPrivateKey(key, "")
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(block)
}

func reset(t *testing.T) {
	t.Helper()
	Close()
	t.Cleanup(Close)
}

func TestDialFuncDirect(t *testing.T) {
	reset(t)
	t.Setenv("SERVER_ADDR", "")

	c, err := DialFunc()(context.Background(), "tcp", echoServer(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()

	if _, ok := c.(conn); ok {
		t.Error("direct connections should not be wrapped")
	}
}

func TestDialFuncTunnel(t *testing.T) {
	reset(t)
	writeKey(t, clientKey(t))
	t.Setenv("SERVER_USER", "test")
	t.Setenv("SERVER_ADDR", sshServer(t))

	dial := DialFunc()
	target := echoServer(t)

	for range 2 { // the second dial reuses the ssh connection
		c, err := dial(context.Background(), "tcp", target)
		if err != nil {
			t.Fatal(err)
		}

		if err = c.SetDeadline(time.Now()); err != nil {
			t.Error(err)
		} else if err = c.SetReadDeadline(time.Now()); err != nil {
			t.Error(err)
		} else if err = c.SetWriteDeadline(time.Now()); err != nil {
			t.Error(err)
		}

		if _, err = c.Write([]byte("ping")); err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4)
		if _, err = io.ReadFull(c, buf); err != nil || string(buf) != "ping" {
			t.Errorf("read %q, %v", buf, err)
		}
		_ = c.Close()
	}

	// the server rejects channels to unreachable targets
	if _, err := dial(context.Background(), "tcp", "127.0.0.1:1"); err == nil {
		t.Error("expected channel rejection")
	}
}

func TestDialFuncErrors(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
		addr string
	}{
		{"missing key", nil, "127.0.0.1:1"},
		{"invalid key", []byte("not a key"), "127.0.0.1:1"},
		{"unreachable server", clientKey(t), "127.0.0.1:1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reset(t)
			if tt.key != nil {
				writeKey(t, tt.key)
			} else {
				t.Setenv("HOME", t.TempDir())
			}
			t.Setenv("SERVER_ADDR", tt.addr)

			if _, err := DialFunc()(context.Background(), "tcp", "127.0.0.1:1"); err == nil {
				t.Error("expected error")
			}
		})
	}
}

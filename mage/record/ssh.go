package record

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient is a thin wrapper around golang.org/x/crypto/ssh for the few
// shell-out tasks the recorder needs on the outer Proxmox host:
// re-mastering the installer ISO with proxmox-auto-install-assistant.
type SSHClient struct {
	host string
	cfg  *ssh.ClientConfig
}

// NewSSHClient builds an SSH client from a recorder Config.
func NewSSHClient(c *Config) (*SSHClient, error) {
	keyBytes, err := os.ReadFile(c.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("read ssh key %q: %w", c.SSHKey, err)
	}
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse ssh key %q: %w", c.SSHKey, err)
	}

	cfg := &ssh.ClientConfig{
		User:            c.SSHUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         15 * time.Second,
	}

	host := c.SSHHost
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "22")
	}
	return &SSHClient{host: host, cfg: cfg}, nil
}

// Run executes a shell command on the outer host, returning combined stdout
// + stderr. Non-zero exit codes produce an error containing the captured
// output for diagnosis.
func (s *SSHClient) Run(cmd string) (string, error) {
	c, err := ssh.Dial("tcp", s.host, s.cfg)
	if err != nil {
		return "", fmt.Errorf("ssh dial %s: %w", s.host, err)
	}
	defer func() { _ = c.Close() }()

	sess, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh new session: %w", err)
	}
	defer func() { _ = sess.Close() }()

	var buf bytes.Buffer
	sess.Stdout = &buf
	sess.Stderr = &buf
	if err := sess.Run(cmd); err != nil {
		return buf.String(), fmt.Errorf("ssh run %q: %w\noutput: %s", cmd, err, buf.String())
	}
	return buf.String(), nil
}

// WriteFile writes data to a file on the outer host via SFTP-style cat.
// Avoids a hard sftp dependency by streaming over stdin to `cat`.
func (s *SSHClient) WriteFile(remotePath string, data []byte) error {
	c, err := ssh.Dial("tcp", s.host, s.cfg)
	if err != nil {
		return fmt.Errorf("ssh dial %s: %w", s.host, err)
	}
	defer func() { _ = c.Close() }()

	sess, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("ssh new session: %w", err)
	}
	defer func() { _ = sess.Close() }()

	stdin, err := sess.StdinPipe()
	if err != nil {
		return fmt.Errorf("ssh stdin pipe: %w", err)
	}

	var stderr bytes.Buffer
	sess.Stderr = &stderr

	if err := sess.Start(fmt.Sprintf("cat > %q", remotePath)); err != nil {
		return fmt.Errorf("ssh start cat: %w", err)
	}

	if _, err := io.Copy(stdin, bytes.NewReader(data)); err != nil {
		_ = stdin.Close()
		return fmt.Errorf("ssh write: %w", err)
	}
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("ssh close stdin: %w", err)
	}
	if err := sess.Wait(); err != nil {
		return fmt.Errorf("ssh wait cat: %w\nstderr: %s", err, stderr.String())
	}
	return nil
}

// TempPath returns a unique remote path under /tmp suitable for short-lived
// scratch files like the answer.toml.
func TempPath(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return path.Join("/tmp", fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b)))
}

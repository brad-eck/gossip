package session

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

// Session represents once connected SSH host with interactive shell
type Session struct {
	Host       string
	Client     *ssh.Client
	SSHSession *ssh.Session
	Stdin      io.WriteCloser
	Output     chan []byte
	Done       chan struct{}
	mu         sync.Mutex
	active     bool
	err        error
}

func NewSession(host string) *Session {
	s := &Session{
		Host:   host,
		Output: make(chan []byte, 256),
		Done:   make(chan struct{}),
		active: true,
	}

	go s.connectAndRun()

	return s
}

func (s *Session) connectAndRun() {
	defer close(s.Done)

	// TODO: Improve auth - agent, keys, config file
	config := &ssh.ClientConfig{
		User:            parseUser(s.Host),
		Auth:            []ssh.AuthMethod{ssh.PublicKeysCallback(sshAgentSigners)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // This is insecure, for testing only
		Timeout:         10 * time.Second,
	}

	if config.User == "" {
		config.User = "root" // fallback user
	}

	addr := parseAddr(s.Host)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		s.setError(fmt.Errorf("dial failed: %w", err))
		return
	}
	s.Client = client

	sess, err := client.NewSession()
	if err != nil {
		s.setError(fmt.Errorf("session failed: %w", err))
		return
	}
	s.SSHSession = sess

	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	if err := sess.RequestPty("xterm-256color", h, w, ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		s.setError(fmt.Errorf("pty request failed: %w", err))
		return
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		s.setError(fmt.Errorf("stdin pipe failed: %w", err))
		return
	}
	s.Stdin = stdin

	stdout, err := sess.StdoutPipe()
	if err != nil {
		s.setError(fmt.Errorf("stdout pipe failed: %w", err))
		return
	}

	stderr, err := sess.StderrPipe()
	if err != nil {
		s.setError(fmt.Errorf("stderr pipe failed: %w", err))
		return
	}

	// Starting shell
	if err := sess.Shell(); err != nil {
		s.setError(fmt.Errorf("shell failed: %w", err))
		return
	}

	go s.readOutput(io.MultiReader(stdout, stderr))
}

// Sends input to shell
func (s *Session) Write(b []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active || s.Stdin == nil {
		return fmt.Errorf("session inactive")
	}
	_, err := s.Stdin.Write(b)
	return err
}

func (s *Session) readOutput(r io.Reader) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			s.Output <- data
		}
		if err != nil {
			if err != io.EOF {
				s.setError(fmt.Errorf("read error: %w", err))
			}
			return
		}
	}
}

func (s *Session) setError(err error) {
	s.mu.Lock()
	s.err = err
	s.active = false
	s.mu.Unlock()
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
	if s.SSHSession != nil {
		_ = s.SSHSession.Close()
	}
	if s.Client != nil {
		_ = s.Client.Close()
	}
}

// Helpers
func parseUser(host string) string {
	if i := strings.Index(host, "@"); i > 0 {
		return host[:i]
	}
	return ""
}

func parseAddr(host string) string {
	if i := strings.Index(host, "@"); i >= 0 {
		host = host[i+1:]
	}
	if !strings.Contains(host, ":") {
		host += ":22"
	}
	return host
}

func sshAgentSigners() ([]ssh.Signer, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("no SSH agent running (SSH_AUTH_SOCK not set)")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH agent socket: %w", err)
	}

	// Note: we don't close conn here – the agent client manages it
	agentClient := agent.NewClient(conn)

	signers, err := agentClient.Signers()
	if err != nil {
		conn.Close() // clean up on failure
		return nil, fmt.Errorf("failed to get signers from agent: %w", err)
	}

	return signers, nil
}

package session

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Session represents once connected SSH host with interactive shell
type Session struct {
	Host	   string
	Client	   *ssh.Client
	SSHSession *ssh.Session
	Stdin	   io.WriteCloser
	Output	   chan []byte
	Done	   chan struct{}
	mu	       sync.Mutex
	active	   bool
	err		   error
}

func NewSession(host string) *Session {
	s := &Session{
		Host:	host,
		Ouput:  make(chan []byte, 256),
		Done:	make(chan struct{}),
		active: true,
	}

	go s.connectAndRun()

	return s
}

func (s *Session) connectAndRun() {
	defer close(s.Done)

	// TODO: Improve auth - agent, keys, config file
	config := &ssh.ClientConfig{
		User:			 parseUser(s.Host),
		Auth:			 []ssh.AuthMethod{ssh.PublicKeysCallback(sshAgentSigners)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // This is insecure, for testing only
		Timeout:		 10 * time.Second,
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
		ssh.ECHO:		   0,
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




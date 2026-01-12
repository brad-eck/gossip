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


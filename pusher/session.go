package pusher

import (
	"encoding/gob"
	//"encoding/json"
	"errors"
	//"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// Session represents a connection between server and client.
type Session interface {
	// Id returns a session id
	ID() string
	// Recv reads one text frame from session
	Recv() (string, error)
	// Send sends one text frame to session
	Send(string) error
	// Close closes the session with provided code and reason.
	Close(status uint32, reason string) error
	// Hack.. add a method to return the path
	Path() string
}

type sessionState uint32

const (
	// brand new session, need to send "h" to receiver
	sessionOpening sessionState = iota
	// active session
	sessionActive
	// session being closed, sending "closeFrame" to receivers
	sessionClosing
	// closed session, no activity at all, should be removed from handler completely and not reused
	sessionClosed
)

var (
	// ErrSessionNotOpen error is used to denote session not in open state.
	// Recv() and Send() operations are not suppored if session is closed.
	ErrSessionNotOpen          = errors.New("sockjs: session not in open state")
	errSessionReceiverAttached = errors.New("sockjs: another receiver already attached")
)

type session struct {
	sync.Mutex
	id    string
	state sessionState
	// protocol dependent receiver (xhr, eventsource, ...)
	recv receiver
	// messages to be sent to client
	sendBuffer []string
	// messages received from client to be consumed by application
	// receivedBuffer chan string
	msgReader  *io.PipeReader
	msgWriter  *io.PipeWriter
	msgEncoder *gob.Encoder
	msgDecoder *gob.Decoder

	// closeFrame to send after session is closed
	closeFrame string

	// internal timer used to handle session expiration if no receiver is attached, or heartbeats if recevier is attached
	sessionTimeoutInterval time.Duration
	heartbeatInterval      time.Duration
	timer                  *time.Timer
	// once the session timeouts this channel also closes
	closeCh chan struct{}
	// hack, add path
	path string
}

type receiver interface {
	// sendBulk send multiple data messages in frame frame in format: a["msg 1", "msg 2", ....]
	sendBulk(...string)
	// sendFrame sends given frame over the wire (with possible chunking depending on receiver)
	sendFrame(string)
	// close closes the receiver in a "done" way (idempotent)
	close()
	// done notification channel gets closed whenever receiver ends
	doneNotify() <-chan struct{}
	// interrupted channel gets closed whenever receiver is interrupted (i.e. http connection drops,...)
	interruptedNotify() <-chan struct{}
}

// Session is a central component that handles receiving and sending frames. It maintains internal state
func newSession(sessionID string, sessionTimeoutInterval, heartbeatInterval time.Duration, path string) *session {
	r, w := io.Pipe()
	s := &session{
		id:                     sessionID,
		path:                   path,
		msgReader:              r,
		msgWriter:              w,
		msgEncoder:             gob.NewEncoder(w),
		msgDecoder:             gob.NewDecoder(r),
		sessionTimeoutInterval: sessionTimeoutInterval,
		heartbeatInterval:      heartbeatInterval,
		closeCh:                make(chan struct{})}
	s.Lock() // "go test -race" complains if ommited, not sure why as no race can happen here
	log.Println("session timeout interval", sessionTimeoutInterval)
	s.timer = time.AfterFunc(sessionTimeoutInterval, s.timeout)
	s.Unlock()
	return s
}

func (s *session) sendMessage(msg string) error {
	s.Lock()
	defer s.Unlock()
	if s.state > sessionActive {
		return ErrSessionNotOpen
	}
	s.sendBuffer = append(s.sendBuffer, msg)
	if s.recv != nil {
		s.recv.sendBulk(s.sendBuffer...)
		s.sendBuffer = nil
	}
	return nil
}

func (s *session) attachReceiver(recv receiver) error {
	s.Lock()
	defer s.Unlock()
	log.Println("attachReceiver")
	if s.recv != nil {
		return errSessionReceiverAttached
	}
	s.recv = recv
	go func(r receiver) {
		select {
		case <-r.doneNotify():
			s.detachReceiver()
		case <-r.interruptedNotify():
			s.detachReceiver()
			log.Println("interruptedNotify, calling close")
			s.close()
		}
	}(recv)

	if s.state == sessionClosing {
		// s.recv.sendFrame(s.closeFrame)
		log.Println("closing would have sent close frame")
		s.recv.close()
		return nil
	}
	if s.state == sessionOpening {
		// s.recv.sendFrame("o")
		s.state = sessionActive
	}
	log.Println("sending buffer")
	s.recv.sendBulk(s.sendBuffer...)
	s.sendBuffer = nil
	log.Println("resetting timer to heartbeat ", s.heartbeatInterval)
	s.timer.Stop()

	s.timer = time.AfterFunc(s.heartbeatInterval, s.heartbeat)
	return nil
}

func (s *session) detachReceiver() {
	s.Lock()
	defer s.Unlock()
	s.timer.Stop()
	s.timer = time.AfterFunc(s.sessionTimeoutInterval, s.timeout)
	s.recv = nil
}

const EVENT_PING_RAW = "{\"event\":\"pusher:ping\",\"data\":\"{}\"}"

func (s *session) heartbeat() {
	// todo: refactor this
	s.Lock()
	defer s.Unlock()
	log.Println("heartbeat")
	if s.recv != nil { // timer could have fired between Lock and timer.Stop in detachReceiver
		// s.recv.sendFrame("h")
		log.Println("heartbeat would have been sent here")
		s.recv.sendFrame(EVENT_PING_RAW)
		s.timer = time.AfterFunc(s.heartbeatInterval, s.heartbeat)
	}
}

func (s *session) accept(messages ...string) error {
	for _, msg := range messages {
		if err := s.msgEncoder.Encode(msg); err != nil {
			return err
		}
	}
	return nil
}

// idempotent operation
func (s *session) closing() {
	s.Lock()
	defer s.Unlock()
	if s.state < sessionClosing {
		s.msgReader.Close()
		s.msgWriter.Close()
		s.state = sessionClosing
		if s.recv != nil {
			log.Println("would have closed with send frame")
			// s.recv.sendFrame(s.closeFrame)
			s.recv.close()
		}
	}
}

func (s *session) timeout() {
	log.Println("timeout called !")
	s.close()
}

// idempotent operation
func (s *session) close() {
	log.Println("close called")
	s.closing()
	s.Lock()
	defer s.Unlock()
	if s.state < sessionClosed {
		s.state = sessionClosed
		s.timer.Stop()
		close(s.closeCh)
	}
}

func (s *session) closedNotify() <-chan struct{} { return s.closeCh }

// Conn interface implementation
func (s *session) Close(status uint32, reason string) error {
	s.Lock()
	if s.state < sessionClosing {
		// s.closeFrame = closeFrame(status, reason)
		log.Println("close frame would have been sent here")
		s.Unlock()
		s.closing()
		return nil
	}
	s.Unlock()
	return ErrSessionNotOpen
}

func (s *session) Recv() (string, error) {
	var msg string
	err := s.msgDecoder.Decode(&msg)
	if err == io.ErrClosedPipe {
		err = ErrSessionNotOpen
	}
	return msg, err
}

func (s *session) Send(msg string) error {
	return s.sendMessage(msg)
}

func (s *session) ID() string { return s.id }

func (s *session) Path() string { return s.path }

package server

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"fmt"
	"github.com/igm/sockjs-go/sockjs"
	"github.com/screencloud/subhub/pubsub"
	"github.com/screencloud/subhub/pusher"
	"log"
	"net/http"
	"sync"
)

type server struct {
	lock sync.RWMutex

	id   string   // unique id for this server
	opts *Options // command opts see below

	pubsub pubsub.PubSub

	// state
	sockets  map[string]*socket  // Socket.Id => Socket
	sessions map[Session]*socket // Session   => Socket

	// sublist *sublist.Sublist // Pattern => Sockets
}

type Event struct {
	Event    string                 `json:"event"`
	Channel  string                 `json:"channel,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	SocketId string                 `json:"socket_id,omitempty"`
	// ServerId string `json:"server_id,omitempty"`
}

type Options struct {
	PubSub           pubsub.Options `json:"pubsub"`
	WebSocketAddress string         `json:"websocket_address"`
	Debug            bool           `json:"debug"`
}

var DefaultOptions = Options{
	PubSub:           pubsub.DefaultOptions,
	WebSocketAddress: "0.0.0.0:8080",
}

func New(opts *Options) *server {
	log.Println("new server called")
	s := &server{
		opts:    opts,
		sockets: make(map[string]*socket),
		pubsub:  pubsub.New(&opts.PubSub),
	}
	return s
}

func (s *server) Start() error {
	var err error = nil
	err = s.pubsub.Start()
	if err != nil {
		return err
	}
	err = s.bind()
	return err
}

func (s *server) bind() error {
	// primary transport is websockets, pusher uses its own websocket endpoint
	// what ive done is hack the sock js code a little so it has similar interface
	// this can certainly be improved, but for now it works ok
	http.Handle("/app/", pusher.NewHandler("/app", sockjs.DefaultOptions, s.newPusherWSHandlerFunc()))
	// fallback transports to sockjs, this does xhr-streaming, polling, iframes, etc
	http.Handle("/pusher/", sockjs.NewHandler("/pusher", sockjs.DefaultOptions, s.newSockJSHandlerFunc()))
	// lastly bind web folder for static files
	http.Handle("/", http.FileServer(http.Dir("web/")))

	return http.ListenAndServe(s.opts.WebSocketAddress, nil)
}

type Session interface {
	// Id returns a session id
	ID() string
	// Recv reads one text frame from session
	Recv() (string, error)
	// Send sends one text frame to session
	Send(string) error
	// Close closes the session with provided code and reason.
	Close(status uint32, reason string) error
}

type socket struct {
	// Id returns a session id
	id string
	// Recv reads one text frame from session
	session Session
	// Path which contains the app id / client token
	path string
}

func (sock *socket) ID() string { return sock.id }
func (sock *socket) Receive(channel string, msg *pubsub.Message) {
	log.Println("got message!!!", msg)
	packet := fmt.Sprintf(RAW_CHANNEL_EVENT, msg.Name, channel, msg.Data)
	log.Println("packet", packet)
	sock.session.Send(packet)
}

func (s *server) newSocket(session Session, path string) *socket {
	log.Println("new socket %s with path: %s", session.ID(), path)
	id := uuid.NewRandom().String()
	sock := &socket{
		id:      id,
		session: session,
		path:    path,
	}
	return sock
}

type pusherWSHandlerFunc func(session pusher.Session)

func (s *server) newPusherWSHandlerFunc() pusherWSHandlerFunc {
	handler := func(session pusher.Session) {
		socket := s.newSocket(session, session.Path())
		s.handleSocket(socket)
	}
	return handler
}

type sockJSHandlerFunc func(session sockjs.Session)

func (s *server) newSockJSHandlerFunc() sockJSHandlerFunc {
	handler := func(session sockjs.Session) {
		for {
			// first message should be path
			if msg, err := session.Recv(); err == nil {
				// decode that first message
				vals := make(map[string]interface{})
				err = json.Unmarshal([]byte(msg), &vals)
				if err != nil {
					log.Println("problem decoding first packet, should be path")
					break
				}
				path, ok := vals["path"]
				if !ok {
					log.Println("missing path in first packet")
					break
				}
				// connect session / socket
				socket := s.newSocket(session, path.(string))
				s.handleSocket(socket)
			} else {
				log.Println("unable to read packet")
				// something went wrong close
			}
			break
		}
		// disconnect
	}
	return handler
}

func (s *server) handleSocket(sock *socket) {
	log.Println("start of socket loop")
	// send connection established
	sock.session.Send(fmt.Sprintf(RAW_CONNECTION_ESTABLISHED, sock.id))
	// recv loop
	for {
		if msg, err := sock.session.Recv(); err == nil {
			// decode event
			log.Println("got msg", msg)
			event := &Event{}
			if err = json.Unmarshal([]byte(msg), event); err == nil {
				s.handleEvent(sock, event)
			} else {
				log.Println("Error decoding event", err.Error())
			}
		} else {
			log.Println("error", err.Error())
			break
		}
	}
	// unregister
	log.Println("end of socket, unsubscribe all")
	s.pubsub.UnsubscribeAll(sock)
}

func (s *server) handleEvent(sock *socket, event *Event) {
	log.Println("event", event)
	switch event.Event {
	case EVENT_PING:
		// respond with pong
		log.Println("replying to client ping")
		sock.session.Send(RAW_PONG)
	case EVENT_PONG:
		// noop
		log.Println("got a pong back, all is ok")
	case EVENT_SUBSCRIBE:
		// call pubsub subscribe
		channel, _ := event.Data["channel"].(string)
		auth, _ := event.Data["auth"].(string)
		s.handleSubscribe(sock, channel, auth)
		log.Println("subscribe event")
	case EVENT_UNSUBSCRIBE:
		// call pubsub unsubscribe
		log.Println("unsubscribe event")
		channel, _ := event.Data["channel"].(string)
		s.handleUnsubscribe(sock, channel)
	case EVENT_ERROR:
		// client sent us an error, print it out
		log.Println("got an error from client", event)
	default:
		// unsure how to handle this packet
		log.Println("got an unexpected event", event)
	}
}

func (s *server) handleSubscribe(sock *socket, channel string, auth string) {
	// todo: check auth
	s.pubsub.Subscribe(sock, channel)
	// todo: check return val
	sock.session.Send(fmt.Sprintf(RAW_SUBSCRIPTION_SUCCEEDED, channel))
}

func (s *server) handleUnsubscribe(sock *socket, channel string) {
	s.pubsub.Unsubscribe(sock, channel)
}

package server

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"fmt"
	"github.com/igm/sockjs-go/sockjs"
	"github.com/screencloud/subhub/pubsub"
	"github.com/screencloud/subhub/pusher"
	"github.com/screencloud/subhub/xredis"
	"github.com/xuyu/goredis"
	"log"
	"net/http"
	"strings"
	"sync"
)

type server struct {
	lock sync.RWMutex

	id   string   // unique id for this server
	opts *Options // command opts see below

	pubsub pubsub.PubSub

	redis *xredis.Redis
	//redisMaster *goredis.Redis // used for write
	//redisSlave  *goredis.Redis // used for reads

	// todo: maintain list / map of sockets
}

type Event struct {
	Event   string `json:"event"`
	Channel string `json:"channel,omitempty"`
	//Data     string `json:"data,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	SocketId  string                 `json:"socket_id,omitempty"`
	Timestamp int64                  `json:"timestamp,omitempty"`
}

type Options struct {
	PubSub             pubsub.Options `json:"pubsub"`
	RedisMasterAddress string         `json:"redis_master"`
	RedisSlaveAddress  string         `json:"redis_slave"`
	WebSocketAddress   string         `json:"websocket_address"`
	Debug              bool           `json:"debug"`
}

var DefaultRedisAddress = "127.0.0.1:6379"
var DefaultOptions = Options{
	PubSub:             pubsub.DefaultOptions,
	RedisMasterAddress: DefaultRedisAddress,
	RedisSlaveAddress:  DefaultRedisAddress,
	WebSocketAddress:   "0.0.0.0:8080",
}

func New(opts *Options) *server {
	s := &server{
		opts: opts,
		// sockets: make(map[string]*socket),
		pubsub: pubsub.New(&opts.PubSub),
	}
	return s
}

func (s *server) Start() error {
	var err error = nil
	err = s.connectRedis()
	if err != nil {
		return err
	}
	err = s.pubsub.Start()
	if err != nil {
		return err
	}
	s.createApp("foooooo")
	settings := s.loadApp("foooooo")
	log.Printf("%+v", settings)
	err = s.bind()

	return err
}

func (s *server) connectRedis() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Println("server connect redis")
	redis, err := xredis.Connect(&goredis.DialConfig{
		Address: s.opts.RedisMasterAddress}, &goredis.DialConfig{
		Address: s.opts.RedisSlaveAddress})
	if err != nil {
		log.Fatal("Unable to connect to redis master", err)
		return err
	}
	s.redis = redis
	return nil
}

func (s *server) bind() error {
	// primary transport is websockets, pusher uses its own websocket endpoint
	// what ive done is hack the sock js code a little so it has similar interface
	// this can certainly be improved, but for now it works ok
	http.Handle("/app/", pusher.NewHandler("/app", sockjs.DefaultOptions, s.newPusherWSHandlerFunc()))
	// fallback transports to sockjs, this does xhr-streaming, polling, iframes, etc
	http.Handle("/pusher/", sockjs.NewHandler("/pusher", sockjs.DefaultOptions, s.newSockJSHandlerFunc()))
	// add an auth endpoint for generating the signatures used in private and presence channels
	http.HandleFunc("/auth", s.newAuthHandlerFunc())
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
	// map of subscribed presence-channels to user_ids
	presense map[string]string
	// hack for now to access server
	server *server
}

func (sock *socket) ID() string { return sock.id }
func (sock *socket) Receive(channel string, msg *pubsub.Message) {

	// todo: this should be refactored, use diff return paths for diff types of subscriptions
	// perhaps we pass in a callback function when subscribing..

	// for now, this fugly hack is used to handle keyspace / object notifications
	if strings.HasPrefix(channel, KEYSPACE_NOTIFICATION_PREFIX) {
		channel = strings.TrimPrefix(channel, KEYSPACE_NOTIFICATION_PREFIX)
		sock.server.handleNotifyObjectChange(sock, channel, msg.Name)
		return
	}

	data, _ := json.Marshal(msg.Data)
	packet := fmt.Sprintf(RAW_CHANNEL_EVENT, msg.Name, channel, data, msg.Timestamp)
	log.Println("packet", packet)
	sock.session.Send(packet)
}

func (s *server) newSocket(session Session, path string) *socket {
	log.Println("new socket %s with path: %s", session.ID(), path)
	id := uuid.NewRandom().String()
	sock := &socket{
		id:       id,
		session:  session,
		path:     path,
		presense: make(map[string]string),
		server:   s,
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
		log.Println("sockjs session start")
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
			}
			break
		}
		log.Println("sockjs session end")
	}
	return handler
}

func (s *server) handleSocket(sock *socket) {
	log.Println("socket loop start")

	// check the path is a valid app id

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
	log.Println("socket closing, unsubscribe all")
	s.pubsub.UnsubscribeAll(sock)

	// presense: trigger member removed for each presence channel
	for presenseChannel, userId := range sock.presense {
		s.presenseMemberRemoved(sock, presenseChannel, userId)
	}
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
		//var data map[string]interface{}
		//err := json.Unmarshal([]byte(event.Data), &data)
		//if err != nil {
		//	log.Println("error decoding event data", err)
		//	return
		//}
		channel := event.Data["channel"].(string)

		switch {
		// private-
		case strings.HasPrefix(channel, CHANNEL_PREFIX_PRIVATE):
			auth := event.Data["auth"].(string)
			message := fmt.Sprintf("%s:%s", sock.ID(), channel)
			if ok := s.verifyAuth(auth, message); !ok {
				log.Println("auth not ok, return some error")
				return
			}

			// channel = strings.TrimPrefix(channel, CHANNEL_PREFIX_PRIVATE)

			s.handleSubscribe(sock, channel)
		// presence-
		case strings.HasPrefix(channel, CHANNEL_PREFIX_PRESENSE):

			auth := event.Data["auth"].(string)
			channelData := event.Data["channel_data"].(string)
			message := fmt.Sprintf("%s:%s:%s", sock.ID(), channel, channelData)
			if ok := s.verifyAuth(auth, message); !ok {
				log.Println("auth not ok, return some error")
				return
			}

			s.handleSubscribePresense(sock, channel, channelData)
		// any other, its public
		default:
			s.handleSubscribe(sock, channel)
		}

		// auth := event.Data["auth"].(string)
		// channel_data := event.Data["channel_data"].(string)
		// check for ?auth= in the channel name
		// s.handleSubscribe(sock, channel, auth)
		log.Println("subscribe event")
	case EVENT_UNSUBSCRIBE:
		// call pubsub unsubscribe
		//var data map[string]interface{}
		//err := json.Unmarshal([]byte(event.Data), &data)
		//if err != nil {
		//	log.Println("error decoding event data", err)
		//	return
		//}
		log.Println("unsubscribe event")
		channel, _ := event.Data["channel"].(string)
		s.handleUnsubscribe(sock, channel)
	case EVENT_ERROR:
		// client sent us an error, print it out
		log.Println("got an error from client", event)
	default:
		// check if this is a client event
		if event.Channel != "" && strings.HasPrefix(event.Event, "client-") {
			s.handleClientEvent(sock, event)
			return
		}
		// if we got to here, we are unsure how to handle this packet
		log.Println("got an unexpected event", event)
	}
}

const (
	CHANNEL_PREFIX_PRIVATE  = "private-"
	CHANNEL_PREFIX_PRESENSE = "presence-"
	CHANNEL_PREFIX_KEYSPACE = "keyspace-"
	CHANNEL_PREFIX_OBJECT   = "object-"
	CHANNEL_PREFIX_TOKEN    = "token-"
)

func (s *server) handleSubscribe(sock *socket, channel string) {
	s.pubsub.Subscribe(sock, channel)
	// todo: check this actually subscribed, if already subed do we send success?
	sock.session.Send(fmt.Sprintf(RAW_SUBSCRIPTION_SUCCEEDED, channel, "\"\""))
}

type RawEvent struct {
	Event    string `json:"event"`
	Channel  string `json:"channel,omitempty"`
	Data     string `json:"data,omitempty"`
	SocketId string `json:"socket_id,omitempty"`
}

func (s *server) handleUnsubscribe(sock *socket, channel string) {

	// if object channel change the prefix
	if strings.HasPrefix(channel, CHANNEL_PREFIX_OBJECT) {
		channel = strings.TrimPrefix(channel, CHANNEL_PREFIX_OBJECT)
		channel = KEYSPACE_NOTIFICATION_PREFIX + channel
	}

	// if keyspace channel change the prefix
	if strings.HasPrefix(channel, CHANNEL_PREFIX_KEYSPACE) {
		channel = strings.TrimPrefix(channel, CHANNEL_PREFIX_KEYSPACE)
		channel = KEYSPACE_NOTIFICATION_PREFIX + channel
	}

	if strings.HasPrefix(channel, CHANNEL_PREFIX_PRESENSE) {
		// remove from hash.
		s.handleUnsubscribePresense(sock, channel)
		return
	}

	s.pubsub.Unsubscribe(sock, channel)

}

func (s *server) handleClientEvent(sock *socket, event *Event) {
	// check we are actually subscribed to the channel in question
	if s.pubsub.IsSubscribed(sock, event.Channel) {
		// hmm this feels wrong having to marshal it again
		// perhaps we can get away with not decoding it in first place
		// need to re-read the pusher spec
		if data, err := json.Marshal(event.Data); err == nil {
			msg := &pubsub.Message{
				Name: event.Event,
				Data: string(data),
			}
			s.pubsub.Publish(sock, event.Channel, msg)
		} else {
			log.Println("unable to marshal data into string", err)
		}
	} else {
		log.Println("not publishing to channel, sock isnt subscribed")
	}
}

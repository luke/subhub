package pubsub

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"github.com/apcera/gnatsd/sublist"
	"github.com/xuyu/goredis"
	"gopkg.in/fatih/set.v0"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	PubSubModeNormal int = 1 << iota
	PubSubModeFirehose
)

type Options struct {
	PubSubNodeId    string `json:"pubsub_node_id"`
	PubSubMode      int    `json:"pubsub_mode"`
	RedisPubAddress string `json:"redis_pub_address"`
	RedisSubAddress string `json:"redis_sub_address"`
}

var DefaultRedisAddress = "127.0.0.1:6379"
var DefaultOptions = Options{
	RedisPubAddress: DefaultRedisAddress,
	RedisSubAddress: DefaultRedisAddress,
}

type Message struct {
	Name      string `json:"name"`
	Data      string `json:"data"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	NodeId    string `json:"node_id"`
}

type pubsub struct {
	lock sync.RWMutex
	opts *Options

	// redis clients - todo: switch over to using xredis wrapper
	redisPub        *goredis.Redis  // used for pub
	redisSub        *goredis.Redis  // used for sub
	redisSubscriber *goredis.PubSub // used for subscribe, unsubscribe

	sublist *sublist.Sublist
	topics  map[string]*set.Set
	subs    map[Subscriber]*set.Set
}

type Subscriber interface {
	ID() string
	Receive(string, *Message)
}

type Publisher interface {
	ID() string
}

type PubSub interface {
	Subscribe(Subscriber, string)
	Unsubscribe(Subscriber, string)
	UnsubscribeAll(Subscriber)
	IsSubscribed(Subscriber, string) bool
	SubscribedList(Subscriber) []interface{} // todo: cast to []Subscriber
	SubscriberList(string) []interface{}     // todo: cast to []string
	Publish(Publisher, string, *Message) (int64, error)
	// Publish(string, *Message) (int64, error)
	Start() error
}

func New(opts *Options) PubSub { // todo: this should return PubSub iface
	ps := &pubsub{
		opts:    opts,
		sublist: sublist.New(),
		subs:    make(map[Subscriber]*set.Set),
		topics:  make(map[string]*set.Set),
	}
	// set to random id if not set
	if opts.PubSubNodeId == "" {
		log.Println("pub sub node id not set, setting to uuid")
		// todo: truncate this to say 16 chars?
		opts.PubSubNodeId = uuid.NewRandom().String()
	}
	log.Println("pub sub node id:", opts.PubSubNodeId)
	if opts.PubSubMode != PubSubModeNormal && opts.PubSubMode != PubSubModeFirehose {
		log.Println("pub sub mode not set using normal mode")
		opts.PubSubMode = PubSubModeNormal
	}
	// Ensure interface compliance
	var _ PubSub = ps
	return ps
}

func (ps *pubsub) Start() error {
	err := ps.connectRedis()
	if err != nil {
		return err
	}
	if ps.opts.PubSubMode == PubSubModeFirehose {
		// turn on the firehose
		ps.redisSubscriber.PSubscribe("*")
	}
	go ps.subLoop()
	return nil
}

func (ps *pubsub) connectRedis() error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	log.Println("pubsub connect redis")

	redisPub, err := goredis.Dial(&goredis.DialConfig{
		Address: ps.opts.RedisPubAddress})
	if err != nil {
		log.Fatal("Unable to connect to redis pub", err)
		return err
	}
	ps.redisPub = redisPub

	redisSub, err := goredis.Dial(&goredis.DialConfig{
		Address: ps.opts.RedisSubAddress})
	if err != nil {
		log.Fatal("Unable to connect to redis sub", err)
		return err
	}
	ps.redisSub = redisSub

	redisSubscriber, err := redisSub.PubSub()
	if err != nil {
		log.Fatal("Unable to create redis subscriber", err)
		return err
	}
	ps.redisSubscriber = redisSubscriber
	return nil
}

const (
	EVENT_MESSAGE      = "message"
	EVENT_SUBSCRIBE    = "subscribe"
	EVENT_UNSUBSCRIBE  = "unsubscribe"
	EVENT_PMESSAGE     = "pmessage"
	EVENT_PSUBSCRIBE   = "psubscribe"
	EVENT_PUNSUBSCRIBE = "punsubscribe"
)

const KEYSPACE_NOTIFICATION_PREFIX = "__keyspace@0__:"

func (ps *pubsub) subLoop() {
	for {
		log.Println("trying to recieve")
		list, err := ps.redisSubscriber.Receive()
		if err != nil {
			log.Fatal("Unable to recieve from redis", err)
			break
		}
		log.Println("list", list)

		idx := 0
		switch list[0] {
		case EVENT_MESSAGE:
			idx++
		case EVENT_PMESSAGE:
			idx += 2
		case EVENT_SUBSCRIBE:
			continue
		case EVENT_UNSUBSCRIBE:
			continue
		case EVENT_PSUBSCRIBE:
			continue
		case EVENT_PUNSUBSCRIBE:
			continue
		default:
			log.Println("unespected pubsub type", list)
		}
		channel := list[idx]
		idx++
		payload := list[idx]
		msg := &Message{}

		// special case to deal with keyspace notifications
		if strings.HasPrefix(channel, KEYSPACE_NOTIFICATION_PREFIX) {
			msg.Name = payload
			// set the timestamp here... does it need .UTC(). ?
			msg.Timestamp = time.Now().UnixNano()
			ps.forwardToLocal(channel, msg)
			continue
		}

		if err := json.Unmarshal([]byte(payload), msg); err == nil {
			if msg.NodeId == ps.opts.PubSubNodeId {
				log.Println("skip, same node id")
				continue
			}
			ps.forwardToLocal(channel, msg)
		} else {
			log.Println("error decoding json %s", err.Error())
		}

	}
	log.Println("subloop end")
}

func (ps *pubsub) Subscribe(sub Subscriber, topic string) {
	log.Println("subscribe", topic)
	topics := ps.topicsSet(sub)
	if topics.Has(topic) {
		log.Println("already subscribed, return")
		return // dont sub again
	}
	topics.Add(topic)
	subs := ps.subsSet(topic)
	numSubs := subs.Size()
	subs.Add(sub)
	numSubs++
	ps.sublist.Insert([]byte(topic), sub)
	log.Println("numSubs", numSubs)
	if numSubs == 1 {
		if ps.opts.PubSubMode == PubSubModeNormal {
			// assuming redis pubsub client is also thread safe
			log.Println("subscribe redis to", topic)
			ps.redisSubscriber.Subscribe(topic)
		} else {
			log.Println("firehose mode?")
		}
	}
}

func (ps *pubsub) Unsubscribe(sub Subscriber, topic string) {
	log.Println("unsubscribe %s", topic)
	subs := ps.subsSet(topic)
	topics := ps.topicsSet(sub)
	if topics.Has(topic) == false {
		return // not subscribed
	}
	topics.Remove(topic)
	subs.Remove(sub)
	ps.sublist.Remove([]byte(topic), sub)
	if subs.Size() == 0 {
		if ps.opts.PubSubMode == PubSubModeNormal {
			log.Println("unsubscribe redis from", topic)
			ps.redisSubscriber.UnSubscribe(topic)
		}
		ps.lock.Lock()
		delete(ps.topics, topic)
		ps.lock.Unlock()
	}
	if topics.Size() == 0 {
		ps.lock.Lock()
		delete(ps.subs, sub)
		ps.lock.Unlock()
	}
}

func (ps *pubsub) UnsubscribeAll(sub Subscriber) {
	// iterate over subs and unsub each one
	log.Println("unsubscribe all")
	topics := ps.topicsSet(sub).List()
	for _, topic := range topics {
		ps.Unsubscribe(sub, topic.(string))
	}
}

var emptyList = make([]interface{}, 0)

func (ps *pubsub) IsSubscribed(sub Subscriber, topic string) bool {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	if s, ok := ps.subs[sub]; ok {
		return s.Has(topic)
	}
	return false
}

func (ps *pubsub) SubscribedList(sub Subscriber) []interface{} {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	if s, ok := ps.subs[sub]; ok {
		return s.List()
	}
	return emptyList
}

func (ps *pubsub) SubscriberList(topic string) []interface{} {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	if s, ok := ps.topics[topic]; ok {
		return s.List()
	}
	return emptyList
}

func (ps *pubsub) Publish(pub Publisher, channel string, msg *Message) (int64, error) {
	if pub != nil {
		msg.Sender = pub.ID()
	}
	msg.NodeId = ps.opts.PubSubNodeId
	msg.Timestamp = time.Now().UnixNano() // utc?
	var count int64 = 0
	var num int64 = 0
	var err error = nil
	num, err = ps.forwardToLocal(channel, msg)
	if err != nil {
		return count, err
	} else {
		count += num
	}
	num, err = ps.forwardToRemote(channel, msg)
	count += num
	return count, err
}

func (ps *pubsub) forwardToLocal(channel string, msg *Message) (int64, error) {
	// use sublist to find local subs, send them a copy
	list := ps.sublist.Match([]byte(channel))
	var err error = nil
	var num int64 = int64(len(list))
	if num == 0 {
		return 0, nil // nothing to do
	}
	for index, item := range list {
		log.Printf("match: %i, %+v", index, item)
		sub := item.(Subscriber)
		if sub.ID() == msg.Sender {
			log.Println("skip, dont deliever msg to sender")
			num--
			continue // dont send to self
		}
		sub.Receive(channel, msg)
	}
	return num, err
}

func (ps *pubsub) forwardToRemote(channel string, msg *Message) (int64, error) {
	// set the via/hub then
	// serialize to json
	buf, err := json.Marshal(msg)
	if err != nil {
		log.Println("unalbe to marshal json")
	}
	// push to redis
	num, err := ps.redisPub.Publish(channel, string(buf))
	return num, err
}

func (ps *pubsub) topicsSet(sub Subscriber) *set.Set {
	// var s *set.Set = nil
	ps.lock.RLock()
	s, ok := ps.subs[sub]
	ps.lock.RUnlock()
	if !ok {
		s = set.New()
		ps.lock.Lock()
		ps.subs[sub] = s
		ps.lock.Unlock()
	}
	return s
}

func (ps *pubsub) subsSet(topic string) *set.Set {
	// var s *set.Set = nil
	ps.lock.RLock()
	s, ok := ps.topics[topic]
	ps.lock.RUnlock()
	if !ok {
		s = set.New()
		ps.lock.Lock()
		ps.topics[topic] = s
		ps.lock.Unlock()
	}
	return s
}

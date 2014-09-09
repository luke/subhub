package pubsub

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"github.com/apcera/gnatsd/sublist"
	"github.com/xuyu/goredis"
	"gopkg.in/fatih/set.v0"
	"log"
	"sync"
)

const (
	PubSubModeNormal int = 1 << iota
	PubSubModeFirehose
)

type Options struct {
	PubSubNodeId       string `json:"pubsub_node_id"`
	PubSubMode         int    `json:"pubsub_mode"`
	RedisMasterAddress string `json:"redis_master_address"`
	RedisSlaveAddress  string `json:"redis_slave_address"`
	RedisSubAddress    string `json:"redis_sub_address"`
}

var DefaultRedisAddress = "127.0.0.1:6379"
var DefaultOptions = Options{
	RedisMasterAddress: DefaultRedisAddress,
	RedisSlaveAddress:  DefaultRedisAddress,
	RedisSubAddress:    DefaultRedisAddress,
}

type Message struct {
	Name   string `json:"name"`
	Data   string `json:"data"`
	Sender string `json:"sender"`
	NodeId string `json:"node_id"`
}

type pubsub struct {
	lock sync.RWMutex
	opts *Options

	// redis clients
	redisMaster *goredis.Redis  // used for writes
	redisSlave  *goredis.Redis  // used for reads
	redisSub    *goredis.Redis  // used for sub
	redisPubSub *goredis.PubSub // used for subscribe, unsubscribe

	sublist *sublist.Sublist
	topics  map[string]*set.Set
	subs    map[Subscriber]*set.Set
}

type Subscriber interface {
	ID() string
	Receive(string, *Message)
}

type PubSub interface {
	Subscribe(Subscriber, string)
	Unsubscribe(Subscriber, string)
	UnsubscribeAll(Subscriber)
	Publish(string, *Message) (int64, error)
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
		ps.redisPubSub.PSubscribe("*")
	}
	go ps.subLoop()
	return nil
}

func (ps *pubsub) connectRedis() error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	log.Println("connect redis")

	redisMaster, err := goredis.Dial(&goredis.DialConfig{
		Address: ps.opts.RedisMasterAddress})
	if err != nil {
		log.Fatal("Unable to connect to redis master", err)
		return err
	}
	ps.redisMaster = redisMaster

	redisSlave, err := goredis.Dial(&goredis.DialConfig{
		Address: ps.opts.RedisSlaveAddress})
	if err != nil {
		log.Fatal("Unable to connect to redis slave", err)
		return err
	}
	ps.redisSlave = redisSlave

	redisSub, err := goredis.Dial(&goredis.DialConfig{
		Address: ps.opts.RedisSubAddress})
	if err != nil {
		log.Fatal("Unable to connect to redis sub", err)
		return err
	}
	ps.redisSub = redisSub

	redisPubSub, err := redisSub.PubSub()
	if err != nil {
		log.Fatal("Unable to create redis pubsub", err)
		return err
	}
	ps.redisPubSub = redisPubSub
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

func (ps *pubsub) subLoop() {
	for {
		log.Println("trying to recieve")
		list, err := ps.redisPubSub.Receive()
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
		// log.Println("raw json  %s", payload)
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
			ps.redisPubSub.Subscribe(topic)
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
			ps.redisPubSub.UnSubscribe(topic)
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

func (ps *pubsub) Publish(channel string, msg *Message) (int64, error) {
	msg.NodeId = ps.opts.PubSubNodeId
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
	num, err := ps.redisMaster.Publish(channel, string(buf))
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

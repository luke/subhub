package server

import (
	"encoding/json"
	"fmt"
	"log"
)

const KEYSPACE_NOTIFICATION_PREFIX = "__keyspace@0__:"

func (s *server) handleSubscribeKeyspace(sock *socket, channel string, auth string) {
	s.pubsub.Subscribe(sock, KEYSPACE_NOTIFICATION_PREFIX+channel)
	sock.session.Send(fmt.Sprintf(RAW_SUBSCRIPTION_SUCCEEDED, channel))
}

func (s *server) handleNotifyObjectChange(sock *socket, channel string, keyspaceEvent string) {

	data, err := redisGetKeyData(s.redis.Slave(), channel)
	if err != nil {
		log.Println("problem subscribing to object", err.Error())
		return
	}

	event := &RawEvent{
		Channel: CHANNEL_PREFIX_OBJECT + channel,
		Event:   "change",
		Data:    data,
	}

	packet, err := json.Marshal(event)

	if err != nil {
		log.Println("problem encoding object load")
		return
	}

	sock.session.Send(string(packet))
}

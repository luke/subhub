package server

import (
	"encoding/json"
	"fmt"
	"github.com/xuyu/goredis"
	"log"
)

const (
	OBJECT_TYPE_STRING = "string"
	OBJECT_TYPE_LIST   = "list"
	OBJECT_TYPE_HASH   = "hash"
	OBJECT_TYPE_SET    = "set"
	OBJECT_TYPE_ZSET   = "zset"
	OBJECT_TYPE_NONE   = "none"
)

func redisGetKeyData(r *goredis.Redis, key string) (string, error) {
	objectType, err := r.Type(key)
	var data string = ""
	if err != nil {
		log.Println("Error getting redis type", err)
		return data, err
	}
	log.Println("redis type response", objectType)
	var resp interface{} = nil
	switch objectType {
	case OBJECT_TYPE_STRING:
		resp, err = r.Get(key)
	case OBJECT_TYPE_HASH:
		resp, err = r.HGetAll(key)
	case OBJECT_TYPE_LIST:
		resp, err = r.LRange(key, 0, -1)
	case OBJECT_TYPE_SET:
		resp, err = r.SMembers(key)
	case OBJECT_TYPE_ZSET:
		resp, err = r.ZRange(key, 0, -1, true)
	case OBJECT_TYPE_NONE:
		// key doesnt exist, return empty response
		return data, err
	default:
		// unknow type return empty response
		return data, err
	}
	if err != nil {
		log.Println("Error getting redis value", err)
		return data, err
	}
	if resp != nil {
		log.Printf("got resp to encode %+v", resp)
		// if we have a resp try decoding it
		var encoded []byte
		encoded, err = json.Marshal(resp)
		if err == nil {
			data = string(encoded)
			log.Printf("after encoding %+v", string(data))
		}
	}
	if err != nil {
		log.Println("Error encoding value to json string", err)
	}
	return data, err
}

func (s *server) handleSubscribeObject(sock *socket, channel string, auth string) {
	// panic("object channels not yet implemented")

	// check the key type, is it a hash or a key

	data, err := redisGetKeyData(s.redis.Slave(), channel)
	if err != nil {
		log.Println("problem subscribing to object", err.Error())
		return
	}

	event := &RawEvent{
		Channel: CHANNEL_PREFIX_OBJECT + channel,
		Event:   "load",
		Data:    data,
	}

	packet, err := json.Marshal(event)

	if err != nil {
		log.Println("problem encoding object load")
		return
	}

	// now subscribe to keyspace notifications
	s.pubsub.Subscribe(sock, KEYSPACE_NOTIFICATION_PREFIX+channel)

	sock.session.Send(fmt.Sprintf(RAW_SUBSCRIPTION_SUCCEEDED, channel))
	sock.session.Send(string(packet))
}

package xredis

import (
	"github.com/xuyu/goredis"
)

// Publish posts a message to the given channel.
// Integer reply: the number of clients that received the message.
func (r *Redis) Publish(channel, message string) (int64, error) {
	return r.redisMaster.Publish(channel, message)
}

// PubSub new a PubSub from *redis.
func (r *Redis) PubSub() (*goredis.PubSub, error) {
	return r.redisSlave.PubSub()
}

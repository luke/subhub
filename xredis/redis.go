package xredis

import (
	// "github.com/fatih/structs"
	"github.com/xuyu/goredis"
	"log"
)

type Redis struct {

	// lock sync.RWMutex

	redisMaster *goredis.Redis // used for write
	redisSlave  *goredis.Redis // used for reads
}

func Connect(masterConfig *goredis.DialConfig, slaveConfig *goredis.DialConfig) (*Redis, error) {
	r := &Redis{}

	log.Println("server connect redis")

	redisMaster, err := goredis.Dial(masterConfig)
	if err != nil {
		log.Fatal("Unable to connect to redis master", err)
		return r, err
	}
	r.redisMaster = redisMaster

	redisSlave, err := goredis.Dial(slaveConfig)
	if err != nil {
		log.Fatal("Unable to connect to redis sub", err)
		return r, err
	}
	r.redisSlave = redisSlave

	return r, err
}

func (r *Redis) Master() *goredis.Redis {
	return r.redisMaster
}

func (r *Redis) Slave() *goredis.Redis {
	return r.redisSlave
}

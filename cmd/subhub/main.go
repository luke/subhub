package main

import (
	"flag"
	"github.com/screencloud/subhub/server"
	"log"
)

func main() {

	opts := server.DefaultOptions
	psOpts := opts.PubSub

	// Parse flags
	flag.StringVar(&opts.WebSocketAddress, "addr", "0.0.0.0:8081", "Address to bind http for ws and sockjs")
	flag.StringVar(&psOpts.RedisMasterAddress, "master", "127.0.0.1:6379", "Address of redis master, writes go to master")
	flag.StringVar(&psOpts.RedisSlaveAddress, "slave", "127.0.0.1:6379", "Address of redis slave, reads go to slave")
	flag.StringVar(&psOpts.RedisSubAddress, "sub", "127.0.0.1:6379", "Address of redis pub/sub server, used only for subscribe")
	flag.IntVar(&psOpts.PubSubMode, "psmode", 1, "Pub sub mode 1: normal (default) 2: firehose")
	flag.StringVar(&psOpts.PubSubNodeId, "psid", "", "Pub sub node id. Auto generated if not set")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable debug logging.")

	flag.Parse()

	log.Printf("options: %+v", opts)

	// opts.PubSub.PubSubMode = 2 // firehose!

	s := server.New(&opts)
	err := s.Start()
	if err != nil {
		log.Fatal(err.Error())
	}
}

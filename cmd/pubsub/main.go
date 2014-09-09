package main

import (
	"github.com/screencloud/subhub/pubsub"
	"log"
)

type Sub struct {
}

func (s *Sub) ID() string {
	return "testsub"
}

func (s *Sub) Receive(channel string, msg *pubsub.Message) {
	log.Println("got message!", channel, msg)
}

func main() {

	log.Println("testing pubsub")

	var ps pubsub.PubSub
	ps = pubsub.New(pubsub.DefaultOptions)
	ps.Start()

	sub := &Sub{}
	ps.Subscribe(sub, "baz")

	msg := &pubsub.Message{
		Name:   "foo",
		Data:   "bar",
		Sender: "me",
	}
	num, err := ps.Publish("baz", msg)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("send ", num)
	}

	msg2 := &pubsub.Message{
		Name:   "foo2",
		Data:   "bar2",
		Sender: "testsub",
	}
	num, err = ps.Publish("baz", msg2)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("send ", num)
	}

	foo := make(chan bool)
	<-foo

}

package main

import (
	"bytes"
	"flag"
	"github.com/screencloud/subhub/server"
	"log"
	"os"
	"strings"
	"text/template"
)

type Inventory struct {
	Material string
	Count    uint
}

func getEnv() map[string]string {
	getenvironment := func(data []string, getkeyval func(item string) (key, val string)) map[string]string {
		items := make(map[string]string)
		for _, item := range data {
			key, val := getkeyval(item)
			items[key] = val
		}
		return items
	}
	environment := getenvironment(os.Environ(), func(item string) (key, val string) {
		splits := strings.Split(item, "=")
		key = splits[0]
		val = strings.Join(splits[1:], "=")
		return
	})
	// log.Println("env", environment)
	return environment
}

func templateArg(arg string, env map[string]string) string {
	b := new(bytes.Buffer)
	tmpl, err := template.New("test").Parse(arg)
	if err != nil {
		log.Println(err)
		return arg
	}
	err = tmpl.Execute(b, env)
	if err != nil {
		log.Println(err)
		return arg
	}
	return b.String()
}

func templateArgs(args []string) []string {
	newArgs := make([]string, len(args))
	env := getEnv()
	for idx, arg := range args {
		newArgs[idx] = templateArg(arg, env)
	}
	return newArgs
}

func main() {

	opts := server.DefaultOptions
	psOpts := opts.PubSub

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Parse flags
	f.StringVar(&opts.WebSocketAddress, "http", "0.0.0.0:8081", "Address to bind http for ws and sockjs")
	f.StringVar(&opts.RedisMasterAddress, "master", "127.0.0.1:6379", "Address of redis master, writes go to master")
	f.StringVar(&opts.RedisSlaveAddress, "slave", "127.0.0.1:6379", "Address of redis slave, reads go to slave")
	f.StringVar(&psOpts.RedisPubAddress, "pub", "127.0.0.1:6379", "Address of redis pub server, used only for publish")
	f.StringVar(&psOpts.RedisSubAddress, "sub", "127.0.0.1:6379", "Address of redis sub server, used only for subsciptions")
	f.IntVar(&psOpts.PubSubMode, "psmode", 1, "Pub sub mode 1: normal (default) 2: firehose")
	f.StringVar(&psOpts.PubSubNodeId, "psid", "", "Pub sub node id. Auto generated if not set")
	f.BoolVar(&opts.Debug, "debug", false, "Enable debug logging.")

	// log.Println("args", os.Args)
	err := f.Parse(templateArgs(os.Args[1:]))
	if err != nil {
		log.Println("problem", err)
	}

	log.Printf("options: %+v", opts)

	// opts.PubSub.PubSubMode = 2 // firehose!

	s := server.New(&opts)
	err = s.Start()
	if err != nil {
		log.Fatal(err.Error())
	}
}

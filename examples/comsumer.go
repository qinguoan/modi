package main

import (
	"github.com/nats-io/nats"
	"log"
	"os"
	"runtime"
	"strings"
)

func usage() {
	log.Fatalf("Usage: nats-sub [-s server] [--ssl] [-t] <subject> \n")
}

var index = 0

func printMsg(m *nats.Msg, i int) {
	index += 1
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
}

func main() {

	log.SetFlags(0)

	opts := nats.DefaultOptions
	opts.Servers = strings.Split(os.Args[1], ",")

	nc, err := opts.Connect()
	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}

	subj, i := os.Args[2], 0

	nc.Subscribe(subj, func(msg *nats.Msg) {
		i += 1
		printMsg(msg, i)
	})

	log.Printf("Listening on [%s]\n", subj)

	runtime.Goexit()
}

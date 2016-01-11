/*
---
desctiprion:
  general:
    - "This Appalication is used to forward nginx access.log to gnatsd."
    - "This module used to monitor file logs refresh and redirect new lines to gnatsd."
    - "Fashion like tail -F, performance much more better.modi source won't maintain any more."
  changlog:
    - "automataclly discover new log files and deal with new lines."
    - "update golang nats package to latest version."
    - "develop log calculate module in golang."
  author: qinguoan@wandoujia.com
  commit: 2015-08-14
*/
package main

import (
	"fmt"
	"github.com/nats-io/nats"
	"os"
	"path"
	"regexp"
	"strings"
	"taild/tail"
	"time"
)

var (
	newFileQueue chan string    = make(chan string)
	fileExistMap map[string]int = make(map[string]int)
)

type dataInfo struct {
	Filename string
	Hostname string
	Text     string
}

func getFile(p string) {
	fd, err := os.Open(p)
	if err != nil {
		os.Exit(1)
	}
	defer fd.Close()

	for {

		files, _ := fd.Readdir(0)

		for _, fh := range files {
			name := fh.Name()
			_, ok := fileExistMap[name]
			if match, _ := regexp.MatchString(".*.access.log$", name); match && !ok {
				newFileQueue <- name
				fileExistMap[name] = 1
			}
		}
		time.Sleep(3 * time.Second)

	}
}

func (d *dataInfo) tailFile(nc *nats.Conn) {
	t, err := tail.NewTail(d.Filename, nc)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = t.Wait()

	if err != nil {
		fmt.Println(err)
	}

}

func main() {
	borker := os.Args[1]
	paths := os.Args[2]

	fmt.Printf("borker:%s, paths: %s\n", borker, paths)
	servers := strings.Split(borker, ",")
	// check whether new log file has been created.
	go getFile(paths)

	hostname, _ := os.Hostname()
	opts := nats.DefaultOptions
	opts.Servers = servers
	opts.MaxReconnect = -1
	opts.ReconnectWait = 1 * time.Second
	opts.PingInterval = 15 * time.Second
	nc, err := opts.Connect()
	if err != nil {
		fmt.Printf("Can't connect: %v\n", err)
	}
	defer nc.Close()
	nc.Opts.DisconnectedCB = func(_ *nats.Conn) {
		fmt.Printf("Got disconnected!\n")
	}

	nc.Opts.ReconnectedCB = func(nc *nats.Conn) {
		fmt.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
	}

	nc.Opts.ClosedCB = func(nc *nats.Conn) {
		fmt.Printf("Nats connection closed!! err: %s \n", nc.LastError())
		os.Exit(1)
	}

	for {
		fmt.Println("wait for queue")
		filename := <-newFileQueue
		fmt.Println("get " + filename + " from queue")
		filename = path.Join(paths, filename)
		fmt.Println(filename)
		fi := &dataInfo{
			Filename: filename,
			Hostname: hostname,
		}
		fmt.Printf("%s is started\n", filename)
		go fi.tailFile(nc)
	}

}

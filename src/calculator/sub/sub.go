package sub

import (
	"calculator/config"
	"fmt"
	"github.com/nats-io/nats"
	//json "github.com/pquerna/ffjson/ffjson"
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"time"
	"utils/logger"
)

type Subscriber struct {
	NatsConn *nats.Conn
	Group    string
	Topic    string
	LogsChan chan LogJsonObj
	logsChan chan []byte
}

/*
	Json data from gnatsd.
*/
type LogJsonObj struct {
	HostName string
	FileName string
	LineText string
}

func NewSub() *Subscriber {
	opts := nats.DefaultOptions
	for _, v := range strings.Split(config.NatsServers, ",") {
		opts.Servers = append(opts.Servers, fmt.Sprintf("nats://%s", v))
	}
	opts.MaxReconnect = -1
	opts.ReconnectWait = 5 * time.Second
	opts.PingInterval = 15 * time.Second

	nc, err := opts.Connect()
	if err != nil {
		logger.Printf("Can't connect: %v\n", err)
	}
	nc.Opts.DisconnectedCB = func(_ *nats.Conn) {
		logger.Printf("Got disconnected! %v\n", nc.LastError())
	}

	nc.Opts.ReconnectedCB = func(nc *nats.Conn) {
		logger.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
	}

	nc.Opts.ClosedCB = func(nc *nats.Conn) {
		logger.Printf("Nats connection closed!! err: %+v\n", nc.IsClosed())
		os.Exit(1)

	}

	return &Subscriber{
		NatsConn: nc,
		Group:    config.SubGroupName,
		Topic:    config.SubTopicName,
	}

}

func (s *Subscriber) GetMessage() {
	logger.Println("start to receive message from gnatsd")
	s.LogsChan = make(chan LogJsonObj, 4096)
	s.logsChan = make(chan []byte, 4096)
	go s.unmarshal()
	s.NatsConn.QueueSubscribe(s.Topic, s.Group, func(m *nats.Msg) {
		s.logsChan <- m.Data
	})

}

func (s *Subscriber) unmarshal() {
	for i := 0; i < runtime.NumCPU(); i++ {
		jsonObj := LogJsonObj{}
		go func() {
			for {
				select {
				case b := <-s.logsChan:
					err := json.Unmarshal(b, &jsonObj)
					if err == nil {
						s.LogsChan <- jsonObj
					}
				}

			}
		}()
	}
}

func (s *Subscriber) Close() {
	s.NatsConn.Close()
}

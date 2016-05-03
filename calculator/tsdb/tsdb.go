package tsdb

import (
	"fmt"
	"modi/calculator/config"
	"modi/utils/logger"
	"net"
	"sync"
	"time"
)

type TsdbPipeLine struct {
	sync.RWMutex
	closed      chan string
	MessageChan chan string
	Opts        Options
	hosts       chan string
	pool        map[string]*Client
}

type Client struct {
	lastAttempt time.Time
	connStats   Status
	url         string
	conn        net.Conn
	wg          sync.WaitGroup
	sync.RWMutex
}

type Status int

const (
	DISCONNECT = Status(iota)
	CONNECTED
)

var DefaultOpts = Options{
	ConnectTimeout: 3 * time.Second,
	ReconnectWait:  3 * time.Second,
}

type Options struct {
	ConnectTimeout time.Duration
	ReconnectWait  time.Duration
}

func NewPipeLine() *TsdbPipeLine {

	hosts, _ := net.LookupHost(config.TsdbHost)
	tpl := &TsdbPipeLine{
		MessageChan: make(chan string, 4096),
		pool:        make(map[string]*Client),
		Opts:        DefaultOpts,
		closed:      make(chan string),
	}

	tpl.hosts = make(chan string, len(hosts))
	for _, host := range hosts {
		tpl.hosts <- fmt.Sprintf("%s:%s", host, config.TsdbPort)
	}

	go tpl.run()

	return tpl
}

func (t *TsdbPipeLine) run() {
	for {
		select {
		case url := <-t.hosts:
			go func() {
				c := t.newConn(url)
				c.wg.Add(2)
				go t.upload(c)
				go t.watcher(c)
			}()
		case <-t.closed:
			return
		}
	}

}

func (t *TsdbPipeLine) newConn(url string) *Client {

	c := &Client{}

	t.RLock()
	client, ok := t.pool[url]
	t.RUnlock()

	if ok {
		client.wg.Wait()
		c = client
		if past := time.Since(c.lastAttempt); past < t.Opts.ReconnectWait {
			time.Sleep(t.Opts.ReconnectWait - past)
		}
	} else {
		c.url = url
		c.setDisconnect()
	}

	conn, err := net.DialTimeout("tcp", url, t.Opts.ConnectTimeout)

	if err == nil {
		logger.Printf("connect to tsdb server: %v ok!\n", url)
		c.conn = conn
		c.setConnected()
	}

	c.lastAttempt = time.Now()

	t.Lock()
	defer t.Unlock()
	t.pool[url] = c

	return c
}

func (t *TsdbPipeLine) upload(c *Client) {
	defer c.wg.Done()

	if c.getStatus() == DISCONNECT {
		return
	}

	for {
		select {
		case s := <-t.MessageChan:
			if c.getStatus() == DISCONNECT {
				t.MessageChan <- s
				return
			}
			_, err := c.conn.Write([]byte(s))
			if err != nil {
				t.MessageChan <- s
				logger.Printf("write:%s to tsdb:%s error:%#v\n", s, c.url, err)
			}
		}
	}

}

func (t *TsdbPipeLine) watcher(c *Client) {
	defer c.wg.Done()

	if c.getStatus() == DISCONNECT {
		t.hosts <- c.url
		return
	}

	for {
		read := make([]byte, 1024)
		_, err := c.conn.Read(read)

		if t.isClosed() {
			return
		}

		if err != nil {
			logger.Printf("read from tsdb server error: %v => %v, close it", c.conn, err)
			c.conn.Close()
			c.setDisconnect()
			t.hosts <- c.url
			break
		} else {
			logger.Printf("receive from tsdb server: %s => %s", c.url, string(read))
		}
	}
}

func (t *TsdbPipeLine) isClosed() bool {
	select {
	case <-t.closed:
		return true
	default:
		return false
	}
}

func (c *Client) getStatus() Status {
	c.RLock()
	defer c.RUnlock()
	return c.connStats
}

func (c *Client) setDisconnect() {
	c.Lock()
	defer c.Unlock()
	c.connStats = DISCONNECT
}

func (c *Client) setConnected() {
	c.Lock()
	defer c.Unlock()
	c.connStats = CONNECTED
}

func (t *TsdbPipeLine) Close() {
	close(t.closed)
	for _, c := range t.pool {
		if c.getStatus() == CONNECTED {
			c.conn.Close()
		}
	}
}

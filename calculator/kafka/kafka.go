package kafka

import (
	"github.com/Shopify/sarama"
	"modi/calculator/config"
	"modi/utils/logger"
	"strings"
	"sync"
)

type Producer struct {
	Worker      sarama.AsyncProducer
	MessageChan chan []byte
}

/*
type LogJsonObj struct {
	HostName string
	FileName string
	LineText string
}
*/

func NewPub() *Producer {
	kafkaservers := strings.Split(config.KafkaServers, ",")
	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true
	work, err := sarama.NewAsyncProducer(kafkaservers, conf)
	if err != nil {
		logger.Printf("Can't connet to kafka broker: %s", err)
	}

	return &Producer{
		Worker:      work,
		MessageChan: make(chan []byte, 4096),
	}

}

func (p *Producer) PublishToKafka() {
	var (
		wg                sync.WaitGroup
		successes, errors int
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _ = range p.Worker.Successes() {
			successes++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range p.Worker.Errors() {
			logger.Printf("error occured when send message to kafka: %s", err)
			errors++
		}
	}()

	for {
		select {
		case m := <-p.MessageChan:
			go p.sendMsg(m)
		}
	}

}

func (p *Producer) sendMsg(data []byte) {
	message := &sarama.ProducerMessage{Topic: config.KafkaTopicName, Value: sarama.StringEncoder(data)}
	p.Worker.Input() <- message
}

package kafka

import (
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"strings"
)

type Consumer struct {
	kafkaConsumer *kafka.Consumer
	Message       chan *kafka.Message
	IsOpen        bool
	address       []string
	group         string
	topic         []string
}

func createConsumerCluster(kafkaAddrs []string, kafkaGroup string) *kafka.Consumer {
	config := kafka.ConfigMap{
		"bootstrap.servers":       strings.Join(kafkaAddrs, ","),
		"group.id":                kafkaGroup,
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 1000,
		"session.timeout.ms":      30000,
		"socket.keepalive.enable": true,
	}
	c, err := kafka.NewConsumer(&config)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func InitKakfaConsumer(kafkaAddrs []string, kafkaGroup string, topic []string) *Consumer {
	c := new(Consumer)
	c.address = kafkaAddrs
	c.group = kafkaGroup
	c.topic = topic
	c.kafkaConsumer = createConsumerCluster(kafkaAddrs, kafkaGroup)
	return c
}

func (c *Consumer) MarkOffset(msg *kafka.Message) {
	if c.IsOpen == false {
		return
	}
	c.kafkaConsumer.CommitMessage(msg)
}

func (c *Consumer) runPooler() {
	for c.IsOpen == true {
		ev := c.kafkaConsumer.Poll(100)
		if ev == nil {
			continue
		}
		switch msg := ev.(type) {
		case *kafka.Message:
			if strings.HasPrefix(*msg.TopicPartition.Topic, "_") == true {
				continue
			} else {
				c.Message <- msg
			}
		case kafka.Error:
			log.Printf("%% Error: %v\n", msg)
			c.IsOpen = false
			var count = 0
			for c.kafkaConsumer.Poll(10) != nil && count < 100 {
				count++
			}
			if count == 100 {
				log.Fatalln("Error: Cannot drain pool to close consumer, hard stop.")
			}
			c.kafkaConsumer.Close()
			c.kafkaConsumer = createConsumerCluster(c.address, c.group)
			err := c.kafkaConsumer.SubscribeTopics(c.topic, nil)
			if err != nil {
				log.Fatalln("Fatal, cannot recover from", err)
			}
			c.IsOpen = true
			log.Printf("Recovered, resuming listening.")
		default:
			// do nothing, ignore the message.
		}
	}
}

func (c *Consumer) Refresh() error {
	if c.IsOpen == false {
		return nil
	}
	err := c.kafkaConsumer.SubscribeTopics(c.topic, nil)
	if err != nil {
		log.Println("Error listening to  topic", err)
	}
	return err
}

func (c *Consumer) Open() error {
	if c.IsOpen == true {
		return errors.New("Unable to open consumer, its already open.")
	}
	if c.address == nil {
		return errors.New("invalid address")
	}
	if c.group == "" {
		return errors.New("invalid group")
	}
	err := c.kafkaConsumer.SubscribeTopics(c.topic, nil)
	if err != nil {
		log.Println("Error listening to topics", err)
	}
	c.Message = make(chan *kafka.Message)
	c.IsOpen = true
	go c.runPooler()
	return nil
}

func (c *Consumer) Close() {
	if c.IsOpen == false {
		return
	}
	c.IsOpen = false
	c.kafkaConsumer.Close()
}

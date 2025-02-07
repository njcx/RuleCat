package kafka

import (
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/json-iterator/go"
	"log"
	log2 "rulecat/utils/log"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type DataProducer struct {
	IsOpen   bool
	address  []string
	group    string
	topic    string
	user     string
	password string

	producer *kafka.Producer
}

func CreateProducer(kafkaAddrs []string, kafkaGroup string, user string, password string) *kafka.Producer {
	c, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(kafkaAddrs, ","),
		"group.id":           kafkaGroup,
		"session.timeout.ms": 30000,
		"sasl.mechanisms":    "PLAIN",
		"security.protocol":  "SASL_PLAINTEXT",
		"sasl.username":      user,
		"sasl.password":      password,
	})
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func (pd *DataProducer) AddMessage(message []byte) error {
	var err error
	go func() {
		for e := range pd.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log2.Error.Printf("Kafka Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log2.Info.Printf("Kafka Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	err = pd.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &pd.topic, Partition: kafka.PartitionAny},
		Value:          message,
		Headers:        []kafka.Header{},
	}, nil)

	pd.producer.Flush(10 * 1000)
	return err

}

func InitKafkaProducer(kafkaAddrs []string, kafkaGroup string, topic string, user string, password string) *DataProducer {

	pd := new(DataProducer)
	pd.address = kafkaAddrs
	pd.group = kafkaGroup
	pd.topic = topic

	pd.user = user
	pd.password = password
	pd.Open()
	return pd
}

func (pd *DataProducer) Open() error {
	if pd.IsOpen == true {
		return errors.New("Unable to open log consumer, its already open.")
	}
	if pd.address == nil || len(pd.address) == 0 {
		return errors.New("invalid address")
	}
	if pd.group == "" {
		return errors.New("invalid group")
	}
	pd.producer = CreateProducer(pd.address, pd.group, pd.user, pd.password)
	pd.IsOpen = true
	return nil
}

func (pd *DataProducer) Close() {
	if pd.IsOpen == false {
		return
	}

	pd.producer.Close()
	pd.IsOpen = false
}

package main

import (
	"fmt"
	"rule_engine_by_go/utils/kafka"
)

func main() {
	kafkaConsumer := kafka.InitKakfaConsumer([]string{"172.21.129.2:9092"}, "test", []string{"nids-conn"})
	kafkaConsumer.Open()

	for {

		message := <-kafkaConsumer.Message
		fmt.Println(string(message.Value))
	}

}

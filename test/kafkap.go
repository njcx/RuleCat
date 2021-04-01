package main

import (
	"fmt"
	"rule_engine_by_go/utils/kafka"
)

type Employee struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Age       int      `json:"age"`
	About     string   `json:"about"`
	Interests []string `json:"interests"`
}

func main() {
	kafkaP := kafka.InitKafkaProducer([]string{"172.21.129.2:9092"}, "test", "test1")

	e1 := Employee{"Jane", "Smith", 32, "I like to collect rock albums", []string{"music"}}

	for i := 0; i < 100; i++ {
		fmt.Println(i)
		err := kafkaP.AddMessage(e1)
		if err != nil {
			fmt.Println(err)
		}
	}

}

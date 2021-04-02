package main

import (
	"fmt"
	"rulecat/app"
	"rulecat/utils/kafka"
	"time"
)

func main() {

	topic := [...]string{"conn", "ssh", "redis", "mysql", "mongodb", "icmp", "dns", "http"}

	outPut := make(chan string, 48)
	for _, topicItem := range topic {

		kafkaC := kafka.InitKakfaConsumer([]string{"172.21.129.2:9092"}, "test1", []string{"nids-" + topicItem})
		kafkaC.Open()

		e := app.NewEngine(topicItem)
		e.ReadRules()

		for i := 0; i <= 5; i++ {
			go func() {
				for {
					message := <-kafkaC.Message
					e.InPutC <- string(message.Value)
				}
			}()
		}
		go e.ResCheck(128, outPut)
		go func() {
			for {
				fmt.Println(<-outPut)
			}
		}()

	}
	time.Sleep(1000000000 * time.Second)
}

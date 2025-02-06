package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"rulecat/app"
	"rulecat/utils"
	"rulecat/utils/kafka"
	log2 "rulecat/utils/log"
	"sync"
	"syscall"
)

func main() {
	topic := [...]string{"topic_tpl", "topic_tpl1"}
	outPut := make(chan *sync.Map, 48)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConsumers := make([]*kafka.Consumer, len(topic))
	defer func() {
		for _, kafkaC := range kafkaConsumers {
			if kafkaC != nil {
				kafkaC.Close()
			}
		}
	}()

	for i, topicItem := range topic {
		wg.Add(7)
		kafkaC := kafka.InitKakfaConsumer(app.ConfigG.InPut.Kafka.Server, app.ConfigG.InPut.Kafka.GroupId,
			[]string{"nids-" + topicItem})
		kafkaConsumers[i] = kafkaC
		if err := kafkaC.Open(); err != nil {
			log2.Error.Printf("Failed to open Kafka consumer for topic %s: %v", topicItem, err)
			wg.Done()
			continue
		}
		e := app.NewEngine(topicItem)
		e.ReadRules()

		for j := 0; j < 5; j++ {
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
						message := <-kafkaC.Message
						e.InPutC <- string(message.Value)
					}
				}
			}()
		}
		go func() {
			defer wg.Done()
			e.ResCheck(128, outPut)

		}()
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case dataStr := <-outPut:
					JsonByte, _ := utils.MarshalSMapToJSON(dataStr)
					app.SendKafka(JsonByte)
					app.SendEs("nids", "alert", string(JsonByte))
					app.SendMail(dataStr)
					app.SendJson(JsonByte)
				}
			}
		}()

	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case killSignal := <-interrupt:
			fmt.Println("Main app got signal:", killSignal)
			log2.Info.Printf("Main app is shutting down due to signal: %v", killSignal)
			cancel()
			wg.Wait()
			return
		}
	}
}

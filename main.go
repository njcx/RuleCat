package main

import (
	"context"
	"os"
	"os/signal"
	"rulecat/app"
	"rulecat/utils"
	"rulecat/utils/kafka"
	log2 "rulecat/utils/log"
	"sync"
	"syscall"
	"time"
)

func main() {
	topic := [...]string{"topic_tpl", "topic_tpl2"}
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

		kafkaC := kafka.InitKakfaConsumer(app.ConfigG.InPut.Kafka.Server, app.ConfigG.InPut.Kafka.GroupId,
			[]string{topicItem}, app.ConfigG.InPut.Kafka.User, app.ConfigG.InPut.Kafka.Passwd)
		kafkaConsumers[i] = kafkaC
		if err := kafkaC.Open(); err != nil {
			log2.Error.Printf("Failed to open Kafka consumer for topic %s: %v", topicItem, err)
			wg.Done()
			continue
		}
		e := app.NewEngine(topicItem)
		e.ReadRules()
		wg.Add(7)

		for j := 0; j < 5; j++ {
			go func(j int) {
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
			}(j)
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
					app.SendEs("_doc", "index_tpl", string(JsonByte))
					app.SendMail(JsonByte)
					app.SendJson(JsonByte)
				}
			}
		}()

	}
	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	defer func() {
		signal.Stop(interrupt)
	}()

	for {
		select {
		case killSignal := <-interrupt:
			log2.Info.Printf("Main app got signal: %v", killSignal)
			log2.Info.Printf("Main app is shutting down due to signal: %v", killSignal)
			cancel()
			timeout := time.After(5 * time.Second)
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()
			select {
			case <-done:
				return
			case <-timeout:
				log2.Warning.Println("Shutdown timeout exceeded, forcing exit.")
				return
			}
		}
	}
}

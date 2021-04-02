package main

import (
	"fmt"
	"os"
	"os/signal"
	"rule_engine_by_go/app"
	"rule_engine_by_go/utils"
	"rule_engine_by_go/utils/kafka"
	log2 "rule_engine_by_go/utils/log"
	"sync"
	"syscall"
)

func main() {
	topic := [...]string{"conn", "ssh", "redis", "mysql", "mongodb", "icmp", "dns", "http"}
	outPut := make(chan *sync.Map, 48)
	for _, topicItem := range topic {

		kafkaC := kafka.InitKakfaConsumer(app.ConfigG.InPut.Kafka.Server, app.ConfigG.InPut.Kafka.GroupId,
			[]string{"nids-" + topicItem})
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
				data := <-outPut
				sjson, _ := utils.MarshalSMapToJSON(data)
				app.SendKafka(sjson)
				app.SendEs("nids", "alert", string(sjson))
				app.SendMail(data)
				app.SendJson(sjson)
			}
		}()

	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	for {
		select {
		case killSignal := <-interrupt:
			fmt.Println("Main app got signal:", killSignal)
			if killSignal == os.Interrupt {
				err := "Main app was interruped by system signal"
				log2.Error.Fatalln(err)
			}
		}
	}
}

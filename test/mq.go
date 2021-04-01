package main

import (
"fmt"
//"rule_engine_by_go/utils/mq"
)

type TestPro struct {
	msgContent   string
}

// 实现发送者
func (t *TestPro) MsgContent() string {
	return t.msgContent
}

// 实现接收者
func (t *TestPro) Consumer(dataByte []byte) error {
	fmt.Println(string(dataByte))
	return nil
}

func main() {
	msg := fmt.Sprintf("这是测试任务")
	t := &TestPro{
		msg,
	}
	queueExchange := &rabbitmq.QueueExchange{
		"test.rabbit",
		"rabbit.key",
		"test.rabbit.mq",
		"direct",
	}
	mq := rabbitmq.New(queueExchange)
	mq.RegisterProducer(t)
	mq.RegisterReceiver(t)
	mq.Start()
}

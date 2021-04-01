package main

import "rule_engine_by_go/utils/email"

func main() {
	emailSender := email.New("smtp.qq.com", 465, "1484703183@qq.com", "123456")
	to := email.ToSomeBody{To: []string{"njcx91@tom.com"}, Cc: []string{"njcx91@tom.com"}}
	emailSender.SendEmail(&to, "test", "test")

}

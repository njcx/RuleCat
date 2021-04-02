package email

import (
	"crypto/tls"
	"github.com/go-gomail/gomail"
	log2 "rulecat/utils/log"
)

type EmailConf struct {
	from string
	d    *gomail.Dialer
}

type ToSomeBody struct {
	To []string
	Cc []string
}

func New(ServerHost string, ServerPort int, EmailAddr string, Passwd string) (*EmailConf, error) {
	ec := &EmailConf{}
	d := gomail.NewDialer(ServerHost, ServerPort, EmailAddr, Passwd)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	_, err := d.Dial()
	if err != nil {
		log2.Error.Println(err)
		return nil, err
	}
	ec.d = d
	ec.from = EmailAddr
	return ec, nil
}

func (ec *EmailConf) SendEmail(to *ToSomeBody, Subject string, content string) error {

	m := gomail.NewMessage()
	m.SetHeader("From", ec.from)
	m.SetHeader("To", to.To...)
	m.SetHeader("Cc", to.Cc...)
	m.SetHeader("Subject", Subject)
	m.SetBody("text/html", content)

	if err := ec.d.DialAndSend(m); err != nil {
		log2.Error.Println(err)
		return err
	}
	return nil
}

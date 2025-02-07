package email

import (
	"crypto/tls"
	"fmt"
	"github.com/go-gomail/gomail"
	"regexp"
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

func New(ServerHost string, ServerPort int, EmailFrom string, EmailUser string, Passwd string) (*EmailConf, error) {
	ec := &EmailConf{}
	d := gomail.NewDialer(ServerHost, ServerPort, EmailUser, Passwd)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	_, err := d.Dial()
	if err != nil {
		log2.Error.Println(err)
		return nil, err
	}
	ec.d = d
	ec.from = EmailFrom
	return ec, nil
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (ec *EmailConf) SendEmail(to *ToSomeBody, Subject string, content string) error {
	if !isValidEmail(ec.from) {
		return fmt.Errorf("invalid email address in From field: %s", ec.from)
	}

	for _, addr := range to.To {
		if !isValidEmail(addr) {
			return fmt.Errorf("invalid email address in To field: %s", addr)
		}
	}

	for _, addr := range to.Cc {
		if !isValidEmail(addr) {
			return fmt.Errorf("invalid email address in Cc field: %s", addr)
		}
	}

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

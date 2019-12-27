package email

import (
	"log"
	"net/smtp"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/matba/slyde-server/internals/utils"
)

func newGmailEmailSender() *gmailEmailSender {
	c := gmailEmailSender{}
	return &c
}

// GmailEmailSender Sends email from Gmail server
type gmailEmailSender struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (c *gmailEmailSender) initialize() error {
	f, err := os.Open(utils.GetConfigPath() + "emailSender.yaml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&c)
	if err != nil {
		return err
	}

	return nil
}

func (c *gmailEmailSender) SendEmail(email string, subject string, body string) error {

	msg := "From: " + c.Username + "\n" +
		"To: " + email + "\n" +
		"Subject:  " + subject + "\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", c.Username, c.Password, "smtp.gmail.com"),
		c.Username, []string{email}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return err
	}

	return nil
}

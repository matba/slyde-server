package email

import (
	"log"
	"sync"
)

var emailSender EmailSender
var mux sync.Mutex

// GetEmailSender Get an instance of email sender
func GetEmailSender() EmailSender {
	mux.Lock()
	if emailSender == nil {
		es := newGmailEmailSender()
		err := es.initialize()
		if err != nil {
			log.Fatal(err)
		}
		emailSender = es
	}
	mux.Unlock()
	return emailSender
}

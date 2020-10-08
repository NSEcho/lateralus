package email

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	mail "github.com/xhit/go-simple-mail/v2"
	"time"
)

// SMTP struct SMTP server configuration
type SMTP struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSL      bool   `json:"useSsl"`
	TLS      bool   `json:"useTls"`
}

// SendMails method sends mails to targets
func (m *SMTP) SendMails(to, bodies []string, attackerName, subject string) {
	client := mail.NewSMTPClient()

	client.Host = m.Host
	client.Port = m.Port
	client.Username = m.Username
	client.Password = m.Password
	if m.SSL && m.TLS {
		log.Fatalf("Cannot use both SSL and TLS for mail")
	}

	if m.SSL && !m.TLS {
		client.Encryption = mail.EncryptionSSL
	}

	if m.TLS && !m.SSL {
		client.Encryption = mail.EncryptionTLS
	}

	if !m.SSL && !m.TLS {
		client.Encryption = mail.EncryptionNone
	}

	client.ConnectTimeout = 10 * time.Second
	client.SendTimeout = 10 * time.Second

	client.KeepAlive = true

	smtpClient, err := client.Connect()
	if err != nil {
		log.Fatalf("Error creating SMTP client: %v\n", err)
	}

	for i := range to {
		email := mail.NewMSG()

		from := fmt.Sprintf("%s <%s>", attackerName, m.Username)

		email.SetFrom(from).
			AddTo(to[i]).
			SetSubject(subject)

		//Get from each mail
		email.GetFrom()
		email.SetBody(mail.TextPlain, bodies[i])

		//Send with high priority
		email.SetPriority(mail.PriorityHigh)

		//Pass the client to the email message to send it
		err := email.Send(smtpClient)

		if err != nil {
			log.Fatalf("Error sending mail: %v\n", err)
		} else {
			log.Infof("Email sent to %s\n", to[i])
		}
	}
}

package email

import (
	"fmt"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	mail "github.com/xhit/go-simple-mail/v2"

	"github.com/cheggaaa/pb/v3"
)

// SMTP struct SMTP server configuration
type SMTP struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	SSL       bool   `json:"useSsl"`
	TLS       bool   `json:"useTls"`
	Priority  string
	Signature string
}

var barTmpl = `{{ green "Sending mails:" }} {{ counters .}} {{ bar . "[" "=" (cycle . "=>") "_" "]"}} {{speed . "%s mail/s" | green }} {{percent . | blue}}`

// SendMails method sends mails to targets
func (m *SMTP) SendMails(names, to, bodies []string, attackerName, subject string, delay int) {
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

	log.Info("Sending mails")

	bar := pb.ProgressBarTemplate(barTmpl).Start64(int64(len(to)))
	//bar := pb.StartNew(len(to))

	for i := range to {
		bar.Increment()
		email := mail.NewMSG()

		from := fmt.Sprintf("%s <%s>", attackerName, m.Username)

		email.SetFrom(from).
			AddTo(to[i]).
			SetSubject(subject)

		// Set priority
		if m.Priority == "high" {
			email.SetPriority(mail.PriorityHigh)
		} else {
			email.SetPriority(mail.PriorityLow)
		}

		// If signature file is passed (template file should be in html)
		if m.Signature != "" {
			sig, err := ioutil.ReadFile(m.Signature)
			if err != nil {
				log.Fatalf("Error opening signature file: %v\n", err)
			}
			email.SetBody(mail.TextHTML, bodies[i]+string(sig))
		} else { // Signature is not passed, treat template as regular text file
			email.SetBody(mail.TextPlain, bodies[i])
		}

		//Pass the client to the email message to send it
		err := email.Send(smtpClient)

		if err != nil {
			log.Fatalf("Error sending mail: %v\n", err)
		}
		// If we need to sleep and it is not last item in targets
		if delay > 0 && i != len(to)-1 {
			log.Infof("Sleeping for %d\n", delay)
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}
}

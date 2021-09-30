package email

import (
	"fmt"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/lateralusd/lateralus/config"
	"github.com/lateralusd/lateralus/logging"
	"github.com/lateralusd/lateralus/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

const (
	ConnectTimeout = 10
	SendTimeout    = 10
)

func SendMails(targets []models.SendingData, opt *config.Options) error {
	singleMailDelay := opt.General.Delay
	bulkTimeout := 0

	client := mail.NewSMTPClient()

	client.Host = opt.MailServer.Host
	client.Port = opt.MailServer.Port
	client.Username = opt.MailServer.Username
	client.Password = opt.MailServer.Password

	switch opt.MailServer.Encryption {
	case "tls":
		client.Encryption = mail.EncryptionTLS
	case "ssl":
		client.Encryption = mail.EncryptionSSL
	default:
		client.Encryption = mail.EncryptionNone
	}

	client.ConnectTimeout = ConnectTimeout * time.Second
	client.SendTimeout = SendTimeout * time.Second

	client.KeepAlive = true

	smtpClient, err := client.Connect()
	if err != nil {
		return fmt.Errorf("SendMails: %v", err)
	}
	defer smtpClient.Close()

	var chunks [][]models.SendingData
	if opt.General.Bulk {
		chunks = createBulks(targets, &opt.General)
		logging.Infof("Created %d chunks with size %d", len(chunks), opt.General.BulkSize)
		bulkTimeout = opt.General.BulkDelay
	} else {
		chunks = append(chunks, targets)
	}

	barTmpl := `{{ green "Sending mails:" }} {{ counters .}} {{ bar . "[" "=" (cycle . "=>") "_" "]"}} {{speed . "%s mail/s" | green }} {{percent . | blue}}`

	bar := pb.ProgressBarTemplate(barTmpl).Start64(int64(len(targets)))

	for _, chunk := range chunks {
		for _, tgt := range chunk {
			bar.Increment()

			email := createMail(opt.Mail.Name, opt.MailServer.Username)
			email.SetPriority(mail.PriorityHigh)
			email.AddTo(tgt.Email).
				SetSubject(opt.Mail.Subject)
			email.SetBody(mail.TextHTML, tgt.Body)

			err := email.Send(smtpClient)
			if err != nil {
				return fmt.Errorf("SendMails: %v", err)
			}
			logging.Infof("Sent mail to %s => %s", tgt.Email, tgt.URL)
			<-time.After(time.Duration(singleMailDelay) * time.Second)

		}
		<-time.After(time.Duration(bulkTimeout) * time.Second)
	}

	bar.Finish()

	return nil

}

func createMail(attackerName, username string) *mail.Email {
	mail := mail.NewMSG()
	mail.SetFrom(fmt.Sprintf("%s <%s>", attackerName, username))
	return mail
}

func createBulks(targets []models.SendingData, general *config.General) [][]models.SendingData {
	chunkSize := general.BulkSize

	var ret [][]models.SendingData
	for i := 0; i < len(targets); i += chunkSize {
		batch := targets[i:min(i+chunkSize, len(targets))]
		ret = append(ret, batch)
	}

	return ret
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

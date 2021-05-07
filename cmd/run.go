package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/lateralusd/lateralus/logging"
	"github.com/lateralusd/lateralus/util"
	"github.com/spf13/cobra"
	mail "github.com/xhit/go-simple-mail/v2"
	"gopkg.in/yaml.v2"

	"github.com/cheggaaa/pb/v3"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the campaign",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		logging.Infof("Starting campaign at %s", start.Format("2006-01-02 15:04:05"))

		config, err := cmd.Flags().GetString("config")
		if err != nil {
			logging.Fatalf("Error ocurred: %v", err)
		}

		if config == "" {
			logging.Fatalf("You need to provide config filename")
		}

		template, err := cmd.Flags().GetString("template")
		if err != nil {
			logging.Fatalf("Error ocurred: %v", err)
		}

		if template == "" {
			logging.Infof("Template not provided, using default template")
		} else {
			logging.Infof("Using template from \"%s\"", template)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logging.Fatalf("Error ocurred: %v", err)
		}

		logging.Infof("Parsing config from \"%s\"", config)
		opts, err := parseConfig(config)
		if err != nil {
			logging.Fatalf("Error parsing configuration: %v", err)
		}

		if output == "" {
			logging.Infof("Output not provided, will use default output (Subject_startTime)")
			output = strings.ReplaceAll(fmt.Sprintf("%s_%s", opts.Mail.Subject, start.Format("2006-01-02 15:04:05")), " ", "")
		}

		logging.Infof("Output filename will be \"%s\"", output)

		logging.Infof("Parsing targets from \"%s\"", opts.Attack.Targets)
		targets, err := parseTargets(opts.Attack.Targets, opts.General.Separator)
		if err != nil {
			logging.Fatalf("Error parsing targets: %v", err)
		}

		sendingData, err := prepareTemplates(targets, opts)
		if err != nil {
			logging.Fatalf("Error preparing templates: %v", err)
		}

		logging.Infof("Starting to send the mails. Hope for the best")

		if err := sendEmails(sendingData, opts); err != nil {
			logging.Infof("Error sending mails: %v", err)
		}

		end := time.Now()
		logging.Infof("Finished sending mails at %s (%s)", end.Format("2006-01-02 15:04:05"), end.Sub(start))

		res := Result{
			StartTime:    start.Format("2006-01-02 15:04:05"),
			EndTime:      end.Format("2006-01-02 15:04:05"),
			Subject:      opts.Mail.Subject,
			From:         fmt.Sprintf("%s <%s>", opts.Mail.Name, opts.MailServer.Username),
			AttackerName: opts.Mail.Name,
			URL:          opts.Url.Link,
			Custom:       opts.Mail.Custom,
			Targets:      sendingData,
		}

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			logging.Fatalf("Error ocurred: %v", err)
		}

		if err := createReport(output, "", format, &res); err != nil {
			logging.Errorf("Error creating report: %v", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("config", "c", "", "config filename")
	runCmd.Flags().StringP("template", "t", "", "template to use for report generation")
	runCmd.Flags().StringP("output", "o", "", "where to store output")
	runCmd.Flags().StringP("format", "f", "tpl", "tpl, xml, json")
}

// Options struct holds all options inside of it
type Options struct {
	Mail       Mail       `yaml:"mail"`
	Attack     Attack     `yaml:"attack"`
	MailServer MailServer `yaml:"mailServer"`
	Url        Url        `yaml:"url"`
	General    General    `yaml:"general"`
}

// Mail struct holds information that will be used to populate mails
type Mail struct {
	Name    string `yaml:"name"`
	From    string `yaml:"from"`
	Subject string `yaml:"subject"`
	Custom  string `yaml:"custom"`
}

// Attack struct holds template targets and mail template used to send mails
type Attack struct {
	Targets  string `yaml:"targets"`
	Template string `yaml:"template"`
}

// MailServer struct holds information needed for mail server loging
type MailServer struct {
	UseSsl   bool   `yaml:"useSsl"`
	UseTls   bool   `yaml:"useTls"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Url struct holds information for mail generation
type Url struct {
	Generate bool   `yaml:"generate"`
	Link     string `yaml:"link"`
	Length   int    `yaml:"length"`
}

type Target struct {
	Name  string
	Email string
}

// Generator struct holds general information
type General struct {
	Bulk      bool   `yaml:"bulk"`
	BulkDelay int    `yaml:"bulkDelay"`
	BulkSize  int    `yaml:"bulkSize"`
	Delay     int    `yaml:"delay"`
	Separator string `yaml:"separator"`
}

type SendingMail struct {
	Target
	Body         string
	AttackerName string
	URL          string
	Custom       string
}

func parseConfig(filename string) (*Options, error) {
	opts := &Options{}

	f, err := os.Open(filename)
	if err != nil {
		return &Options{}, fmt.Errorf("parseConfig: %v", err)
	}
	defer f.Close()

	d := yaml.NewDecoder(f)

	if err := d.Decode(&opts); err != nil {
		return &Options{}, fmt.Errorf("parseConfig: %v", err)
	}

	return opts, nil
}

func parseTargets(filename string, sep string) ([]Target, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []Target{}, fmt.Errorf("parseTargets: %v", err)
	}
	defer f.Close()

	var targets []Target

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			splitted := strings.Split(line, sep)
			if len(splitted) < 2 {
				return []Target{}, errors.New("parseTargets: length of line is not 2, is separator ok?")
			}
			targets = append(targets, Target{
				Name:  splitted[0],
				Email: splitted[1],
			})
		}
	}

	return targets, nil
}

func prepareTemplates(targets []Target, opts *Options) ([]SendingMail, error) {
	t, err := template.ParseFiles(opts.Attack.Template)
	if err != nil {
		return []SendingMail{}, fmt.Errorf("prepareTemplates: %v", err)
	}

	var mails []SendingMail
	for _, tgt := range targets {
		var buf bytes.Buffer
		m := SendingMail{
			AttackerName: opts.Mail.Name,
			URL:          createUserURL(opts),
			Custom:       opts.Mail.Custom,
			Target:       tgt,
		}
		if err := t.Execute(&buf, &m); err != nil {
			return []SendingMail{}, fmt.Errorf("prepareTemplates: for \"%+v\": %v", tgt, err)
		}
		m.Body = buf.String()
		mails = append(mails, m)
	}

	return mails, nil
}

func sendEmails(mails []SendingMail, opts *Options) error {
	client := mail.NewSMTPClient()

	client.Host = opts.MailServer.Host
	client.Port = opts.MailServer.Port
	client.Username = opts.MailServer.Username
	client.Password = opts.MailServer.Password

	if opts.MailServer.UseSsl && opts.MailServer.UseTls {
		return fmt.Errorf("sendEmails: You cannot use both ssl and tls")
	}

	if opts.MailServer.UseSsl && !opts.MailServer.UseTls {
		client.Encryption = mail.EncryptionSSL
	}

	if !opts.MailServer.UseSsl && opts.MailServer.UseTls {
		client.Encryption = mail.EncryptionTLS
	}

	if !opts.MailServer.UseSsl && !opts.MailServer.UseTls {
		client.Encryption = mail.EncryptionNone
	}

	client.ConnectTimeout = 10 * time.Second
	client.SendTimeout = 10 * time.Second

	client.KeepAlive = true

	smtpClient, err := client.Connect()
	if err != nil {
		return fmt.Errorf("sendEmails: %v\n", err)
	}

	chunks := [][]SendingMail{mails}
	singleTimeout := opts.General.Delay
	bulkTimeout := 0
	if opts.General.Bulk {
		chunks = createBulks(mails, &opts.General)
		logging.Infof("Created %d chunks with size %d", len(chunks), opts.General.BulkSize)
		bulkTimeout = opts.General.BulkDelay
	}

	barTmpl := `{{ green "Sending mails:" }} {{ counters .}} {{ bar . "[" "=" (cycle . "=>") "_" "]"}} {{speed . "%s mail/s" | green }} {{percent . | blue}}`

	bar := pb.ProgressBarTemplate(barTmpl).Start64(int64(len(mails)))

	for _, chunk := range chunks {
		for _, tgt := range chunk {
			bar.Increment()
			email := mail.NewMSG()

			from := fmt.Sprintf("%s <%s>", tgt.AttackerName, opts.MailServer.Username)

			email.SetFrom(from).
				AddTo(tgt.Email).
				SetSubject(opts.Mail.Subject)

			email.SetBody(mail.TextHTML, tgt.Body)

			err := email.Send(smtpClient)
			if err != nil {
				return fmt.Errorf("sendEmails: %v", err)
			}
			<-time.After(time.Duration(singleTimeout) * time.Second)
		}
		<-time.After(time.Duration(bulkTimeout) * time.Second)
	}

	return nil
}

func createBulks(targets []SendingMail, general *General) [][]SendingMail {
	chunkSize := general.BulkSize

	var ret [][]SendingMail
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

func createUserURL(urlOpts *Options) string {
	confUrl := urlOpts.Url.Link
	if !urlOpts.Url.Generate {
		return confUrl
	}
	url := confUrl[:strings.Index(confUrl, "<CHANGE>")] + util.GenerateUUID(urlOpts.Url.Length)
	return url
}

var tpl = `Start time:     {{ .StartTime }}
End time:       {{ .EndTime }}

Mail data:
========================================
Mail Subject: 	{{ .Subject }}
From field: 	{{ .From }}
AttackerName: 	{{ .AttackerName }}
URL: 		{{ .URL }}
Custom: 	{{ .Custom }}

Targets:
========================================
Total: 			{{ len .Targets }}
Table in format NAME, EMAIL, URL
----------------------------------------{{ range .Targets }}
{{ .Name | printf "%-20s"}} | {{ .Email | printf "%-50s"}} | {{ .URL }}
{{end}}`

type Result struct {
	StartTime    string
	EndTime      string
	Subject      string
	From         string
	AttackerName string
	URL          string
	Custom       string
	Targets      []SendingMail
}

func createReport(output, templatePath, format string, res *Result) error {
	var err error
	switch format {
	case "json":
		err = createJson(output, res)
	case "xml":
		err = createXml(output, res)
	default:
		err = createTemplate(output, templatePath, res)
	}

	if err != nil {
		return fmt.Errorf("createReport: %v", err)
	}

	return nil
}

func createTemplate(output, templatePath string, res *Result) error {
	var t *template.Template
	var err error

	if templatePath == "" {
		t, err = template.New("").Parse(tpl)
		if err != nil {
			return fmt.Errorf("createTemplate: %v", err)
		}
	} else {
		t, err = template.ParseFiles(templatePath)
		if err != nil {
			return fmt.Errorf("createTemplate: %v", err)
		}
	}

	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("createTemplate: %v", err)
	}

	if err := t.Execute(f, res); err != nil {
		return fmt.Errorf("createTemplate: %v", err)
	}

	return nil
}

func createJson(output string, res *Result) error {
	d, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("createJson: %v", err)
	}

	if err := ioutil.WriteFile(output, d, 0600); err != nil {
		return fmt.Errorf("createJson: %v", err)
	}

	return nil
}

func createXml(output string, res *Result) error {
	d, err := xml.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("createXml: %v", err)
	}

	if err := ioutil.WriteFile(output, d, 0600); err != nil {
		return fmt.Errorf("createXml: %v", err)
	}

	return nil
}

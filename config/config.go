package config

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"github.com/XdaemonX/lateralus/email"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/XdaemonX/lateralus/util"
)

// TemplateFields keep fields that will be used in templates
type TemplateFields struct {
	Name         string `json:"templateName"`
	AttackerName string `json:"attackerName"`
	URL          string `json:"url"`
	Custom       string `json:"custom"`
}

// User struct gets populated from .csv file. These are the targets
type User struct {
	Name  string
	Email string
	URL   string
}

// Options is the main configuration structure
type Options struct {
	SingleURL      *bool   `json:"singleUrl"`
	ConfigFile     *string `json:"config"`
	TemplateName   *string `json:"template"`
	TargetsFile    *string `json:"targets"`
	Generate       *bool   `json:"generateUrl"`
	GenerateLength *int    `json:"generateLength"`
	SMTPConfig     *string `json:"smtpconfig"`
	Subject        *string `json:"subject"`
	From           *string `json:"from"`
	Targets        []User
	*TemplateFields
}

var (
	options = Options{
		SingleURL:      flag.Bool("singleUrl", true, "Use the same URL for all targets"),
		ConfigFile:     flag.String("config", "", "Config file to read parameters from"),
		TemplateName:   flag.String("template", "", "Email template from templates/ directory"),
		TargetsFile:    flag.String("targets", "", "File consisting of targets data (name, lastname, email, url)"),
		Generate:       flag.Bool("generate", false, "If set to true, parameter url needs to have <CHANGE> part"),
		GenerateLength: flag.Int("generateLength", 8, "Length of variable part of url with maximum of 36"),
		SMTPConfig:     flag.String("smtpConfig", "conf/smtp.conf", "SMTP config file"),
		Subject:        flag.String("subject", "Mail Subject", "Subject that will be used for emails"),
		From:           flag.String("from", "", "From field for an email. If not provided, will be the same as attackerName"),
	}
	s        = TemplateFields{}
	csvLines [][]string
)

// SMTPServer is object that will be used for sending mails (Client)
var SMTPServer *email.SMTP

// ParseConfiguration is the main function that will be called from main binary to initialize all flags and parse all config files.
func ParseConfiguration() *Options {
	SMTPServer = &email.SMTP{}

	flag.StringVar(&s.Name, "templateName", "", "Email template name")
	flag.StringVar(&s.AttackerName, "attackerName", "", "Attacker name to use in template")
	flag.StringVar(&s.URL, "url", "", "Single url to include in emails")
	flag.StringVar(&s.Custom, "custom", "", "Custom words to include in template")

	flag.Parse()

	options.TemplateFields = &s

	// If JSON config file is in use
	if *options.ConfigFile != "" {
		options.parseJSON(*options.ConfigFile)
	}

	/*if *options.From == "" {
		*options.From = options.AttackerName
	}*/

	// Parse targets from csv file
	parseCSV(*options.TargetsFile)

	log.Infof("Read %d targets from %s\n", len(options.Targets), *options.TargetsFile)

	// Fill user URL field in case of single field

	// Url param is passed, we have to do something with it
	if options.TemplateFields.URL != "" {
		// Fill every user url with the same field
		if *options.SingleURL {
			for i := range options.Targets {
				options.Targets[i].URL = options.TemplateFields.URL
			}
		} else { // Substitute <CHANGE> part of url with UUID of *options.GenerateLength length
			if strings.Contains(options.TemplateFields.URL, "<CHANGE>") {
				url := options.TemplateFields.URL
				for i := range options.Targets {
					userURL := url[:strings.Index(url, "<CHANGE>")] + util.GenerateUUID(*options.GenerateLength)
					options.Targets[i].URL = userURL
				}
			}
		}

	}

	// Parse smtp configuration
	options.parseSMTP()

	return &options
}

/*
ParseTemplate is method that for each target creates an email body.
First parameter it returns are slice of targets emails.
Second parameter are slices of email bodies for each user.
*/
func (c *Options) ParseTemplate() ([]string, []string) {
	t, err := template.ParseFiles(*c.TemplateName)
	if err != nil {
		log.Fatalf("Error parsing template: %v\n", err)
	}

	var to, bodies []string

	for _, user := range c.Targets {
		var buf bytes.Buffer
		tData := TemplateFields{
			Name:         user.Name,
			AttackerName: c.TemplateFields.AttackerName,
			URL:          user.URL,
			Custom:       c.TemplateFields.Custom,
		}
		_ = t.Execute(&buf, tData)
		to = append(to, user.Email)
		bodies = append(bodies, buf.String())
	}

	return to, bodies
}

func (c *Options) parseSMTP() {
	if len(*c.SMTPConfig) > 1 {
		file, err := os.Open(*options.SMTPConfig)
		defer file.Close()
		data, _ := ioutil.ReadAll(file)
		if err != nil {
			log.Fatalf("Error opening SMTP config file %s: %v\n", *options.SMTPConfig, err)
		}
		err = json.Unmarshal(data, SMTPServer)
		if err != nil {
			log.Fatalf("Error parsing SMTP configuration %v\n", err)
		}
	}
}

func (c *Options) parseJSON(file string) {
	ct, err := os.Open(file)
	defer ct.Close()
	if err != nil {
		log.Fatalf("Error opening JSON configuration (%s): %s . Terminating.", file, err)
	}

	ctb, _ := ioutil.ReadAll(ct)
	err = json.Unmarshal(ctb, &c)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON configuration (%s): %s . Terminating.", file, err)
	}

	err = json.Unmarshal(ctb, &s)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON configuration (%s): %s . Terminating.", file, err)
	}

	options.TemplateFields = &s
}

func parseCSV(file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Error opening target file %s: %v\n", file, err)
	}

	csvLines, err = csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV file: %v\n", err)
	}

	for _, line := range csvLines {
		options.Targets = append(options.Targets, User{Name: line[0], Email: line[1], URL: ""})
	}
}

package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/lateralusd/lateralus/models"
	"github.com/lateralusd/lateralus/util"
	"gopkg.in/yaml.v2"
)

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
	Targets   string `yaml:"targets"`
	Template  string `yaml:"template"`
	Signature string `yaml:"signature"`
}

// MailServer struct holds information needed for mail server loging
type MailServer struct {
	Encryption string `yaml:"encryption"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

// Url struct holds information for mail generation
type Url struct {
	Generate bool   `yaml:"generate"`
	Link     string `yaml:"link"`
	Length   int    `yaml:"length"`
}

// General struct holds general information
type General struct {
	Bulk      bool   `yaml:"bulk"`
	BulkDelay int    `yaml:"bulkDelay"`
	BulkSize  int    `yaml:"bulkSize"`
	Delay     int    `yaml:"delay"`
	Separator string `yaml:"separator"`
}

func ParseConfig(filename string) (*Options, error) {
	opts := &Options{}

	f, err := os.Open(filename)
	if err != nil {
		return &Options{}, fmt.Errorf("ParseConfig: %v", err)
	}
	defer f.Close()

	d := yaml.NewDecoder(f)

	if err := d.Decode(&opts); err != nil {
		return &Options{}, fmt.Errorf("ParseConfig: %v", err)
	}

	return opts, nil
}

func (opt Options) ParseTargets() ([]models.Target, error) {
	f, err := os.Open(opt.Attack.Targets)
	if err != nil {
		return []models.Target{}, fmt.Errorf("parseTargets: %v", err)
	}
	defer f.Close()

	var targets []models.Target

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			splitted := strings.Split(line, opt.General.Separator)
			if len(splitted) < 2 {
				return []models.Target{}, errors.New("parseTargets: length of line is not 2, is separator ok?")
			}
			targets = append(targets, models.Target{
				Name:  splitted[0],
				Email: splitted[1],
			})
		}
	}

	return targets, nil
}

func (opt Options) PrepareMails(targets []models.Target) ([]models.SendingData, error) {
	var mails []models.SendingData
	for _, tgt := range targets {
		url := opt.createUserURL()
		m := models.SendingData{
			AttackerName: opt.Mail.Name,
			URL:          url,
			Custom:       opt.Mail.Custom,
			Target:       tgt,
		}
		body, err := opt.parseBody(tgt.Name, url)
		if err != nil {
			return nil, fmt.Errorf("PrepareMails: %v", err)
		}
		m.Body = body
		mails = append(mails, m)
	}

	return mails, nil
}

func (opt Options) parseBody(targetName, url string) (string, error) {
	t, err := template.ParseFiles(opt.Attack.Template)
	if err != nil {
		return "", fmt.Errorf("parseBody: %v", err)
	}

	data := models.SendingData{
		AttackerName: opt.Mail.Name,
		Custom:       opt.Mail.Custom,
		URL:          url,
		Target: models.Target{
			Name: targetName,
		},
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("parseBody: %v", err)
	}

	f, err := ioutil.ReadFile(opt.Attack.Signature)
	if err != nil {
		return "", fmt.Errorf("parseBody: %v", err)
	}

	buf.Write(f)

	return buf.String(), nil
}

func (opt Options) createUserURL() string {
	confURL := opt.Url.Link
	if !opt.Url.Generate {
		return confURL
	}

	url := confURL[:strings.Index(confURL, "<CHANGE>")] + util.GenerateUUID(opt.Url.Length)
	return url
}

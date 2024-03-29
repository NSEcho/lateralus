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

var (
	//trackingPart = `<img src="%s/?id=%s" style="height:1px !important; width:1px !important; border: 0 !important; margin: 0 !important; padding: 0 !important" width="1" height="1" border="0">`
	trackingPart = `<img src="%s/?id=%s" width="1" height="1" border="0" />`
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
	Targets        string `yaml:"targets"`
	Template       string `yaml:"template"`
	Signature      string `yaml:"signature"`
	TrackingServer string `yaml:"tracking_server"`
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

func (opt Options) ParseTargets(tgtFile string) ([]models.Target, error) {
	f, err := os.Open(tgtFile)
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

	if scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

func (opt Options) PrepareMails(targets []models.Target) ([]models.SendingData, error) {
	var mails []models.SendingData
	for _, tgt := range targets {
		uuid := util.GenerateUUID(opt.Url.Length)
		url := opt.createUserURL(uuid)
		m := models.SendingData{
			AttackerName: opt.Mail.Name,
			URL:          url,
			Custom:       opt.Mail.Custom,
			Target:       tgt,
		}
		body, err := opt.parseBody(tgt.Name, tgt.Email, url, uuid)
		if err != nil {
			return nil, fmt.Errorf("PrepareMails: %v", err)
		}
		m.Body = body
		mails = append(mails, m)
	}

	return mails, nil
}

func (opt Options) parseBody(targetName, email, url, uuid string) (string, error) {
	tPath := "templates/sample"
	if opt.Attack.Template != "" {
		tPath = opt.Attack.Template
	}

	t, err := template.ParseFiles(tPath)
	if err != nil {
		return "", fmt.Errorf("parseBody: %v", err)
	}

	data := models.SendingData{
		AttackerName: opt.Mail.Name,
		Custom:       opt.Mail.Custom,
		URL:          url,
		Target: models.Target{
			Name:  targetName,
			Email: email,
		},
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("parseBody: %v", err)
	}

	// Parse signature if provided
	if opt.Attack.Signature != "" {
		f, err := ioutil.ReadFile(opt.Attack.Signature)
		if err != nil {
			return "", fmt.Errorf("parseBody: %v", err)
		}

		buf.Write(f)
	}

	// Create tracking pixel
	if opt.Attack.TrackingServer != "" {
		pixel := fmt.Sprintf(trackingPart, opt.Attack.TrackingServer, uuid)
		buf.WriteString(pixel)
	}

	return buf.String(), nil
}

func (opt Options) createUserURL(uuid string) string {
	confURL := opt.Url.Link
	if !opt.Url.Generate {
		return confURL
	}

	url := confURL[:strings.Index(confURL, "<CHANGE>")] + uuid
	return url
}

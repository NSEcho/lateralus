package main

import (
	"github.com/lateralusd/lateralus/config"
	"github.com/lateralusd/lateralus/util"
	log "github.com/sirupsen/logrus"
	"os"
	"text/template"
	"time"
)

func initLogging() {
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02.01.2006 15:05:04"
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)
}

func createReport(cfg *config.Options) {
	endTime := time.Now().Format("01-02-2006 15:04:05")
	cfg.EndTime = endTime
	templatePath := "templates/report_template"
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Fatalf("Error parsing template: %v\n", err)
	}

	var reportName string
	if *cfg.ReportName == "" {
		reportName = "report_" + endTime + ".txt"
	} else {
		reportName = *cfg.ReportName
	}
	file, err := os.Create(reportName)
	if err != nil {
		log.Fatalf("Error creating report time: %v\n", err)
	}

	if err = t.Execute(file, *cfg); err != nil {
		log.Fatalf("Error parsing report template: %v\n", err)
	}

	log.Infof("Report created at %s\n", reportName)
}

func main() {
	startTime := time.Now().Format("01-02-2006 15:04:05")

	cfg := config.ParseConfiguration(startTime)

	initLogging()
	log.Info("lateralus started")

	log.Infof("Generating uuids for %d users with uuid length: %d\n", len(cfg.Targets), *cfg.GenerateLength)

	// Parsing email template
	names, to, bodies := cfg.ParseTemplate()

	// Send emails
	config.SMTPServer.SendMails(names, to, bodies, *cfg.From, *cfg.Subject, *cfg.Delay)

	// Write to file
	var users, emails, urls []string
	for _, user := range cfg.Targets {
		users = append(users, user.Name)
		urls = append(urls, user.URL)
		emails = append(emails, user.Email)
	}
	createReport(cfg)
	util.WriteToFile(users, emails, urls)
}

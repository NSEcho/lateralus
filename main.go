package main

import (
	"github.com/XdaemonX/lateralus/config"
	"github.com/XdaemonX/lateralus/util"
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

	reportName := "report_" + endTime + ".txt"
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
	initLogging()
	log.Info("lateralus started")
	cfg := config.ParseConfiguration(startTime)

	log.Infof("Generating uuids for %d users with uuid length: %d\n", len(cfg.Targets), *cfg.GenerateLength)

	// Parsing email template
	names, to, bodies := cfg.ParseTemplate()

	// Send emails
	config.SMTPServer.SendMails(names, to, bodies, *cfg.From, *cfg.Subject)

	// Write to file
	var users, urls []string
	for _, user := range cfg.Targets {
		users = append(users, user.Name)
		urls = append(urls, user.URL)
	}
	createReport(cfg)
	util.WriteToFile(users, urls)
}

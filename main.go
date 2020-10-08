package main

import (
	"github.com/XdaemonX/lateralus/config"
	log "github.com/sirupsen/logrus"
)

func initLogging() {
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02.01.2006 15:05:04"
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)
}

func main() {
	initLogging()
	log.Info("lateralus started")
	cfg := config.ParseConfiguration()

	log.Infof("Generating uuids for %d users with uuid length: %d\n", len(cfg.Targets), *cfg.GenerateLength)

	// Parsing email template
	to, bodies := cfg.ParseTemplate()

	// Send emails
	config.SMTPServer.SendMails(to, bodies, *cfg.From, *cfg.Subject)
}

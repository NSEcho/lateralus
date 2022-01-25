package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lateralusd/lateralus/config"
	"github.com/lateralusd/lateralus/email"
	"github.com/lateralusd/lateralus/logging"
	"github.com/lateralusd/lateralus/models"
	"github.com/lateralusd/lateralus/reports"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the campaign",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		logging.Infof("Starting campaign at %s", start.Format("2006-01-02 15:04:05"))

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		go func() {
			for {
				select {
				case <-c:
					os.Exit(1)
				}
			}
		}()

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logging.Fatalf("Error occurred: %v", err)
		}

		if configPath == "" {
			logging.Fatalf("You need to provide config filename")
		}

		template, err := cmd.Flags().GetString("template")
		if err != nil {
			logging.Fatalf("Error occurred: %v", err)
		}

		if template == "" {
			logging.Infof("Template not provided, using default template")
		} else {
			logging.Infof("Using template from \"%s\"", template)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logging.Fatalf("Error occurred: %v", err)
		}

		logging.Infof("Parsing config from \"%s\"", configPath)
		opts, err := config.ParseConfig(configPath)
		if err != nil {
			logging.Fatalf("Error parsing configuration: %v", err)
		}

		if output == "" {
			logging.Infof("Output not provided, will use default output (Subject_startTime)")
			output = strings.ReplaceAll(fmt.Sprintf("%s_%s", opts.Mail.Subject, start.Format("2006-01-02 15:04:05")), " ", "")
		}

		logging.Infof("Output filename will be \"%s.rep\"", output)

		logging.Infof("Parsing targets from \"%s\"", opts.Attack.Targets)
		targets, err := opts.ParseTargets()
		if err != nil {
			logging.Fatalf("Error parsing targets: %v", err)
		}

		sendingData, err := opts.PrepareMails(targets)
		if err != nil {
			logging.Fatalf("Error preparing templates: %v", err)
		}

		logging.Infof("Starting to send the mails. Hope for the best")

		if err := email.SendMails(sendingData, opts); err != nil {
			logging.Infof("Error sending mails: %v", err)
		}

		end := time.Now()
		logging.Infof("Finished sending mails at %s (%s)", end.Format("2006-01-02 15:04:05"), end.Sub(start))

		/*c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)*/

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			logging.Fatalf("Error occurred: %v", err)
		}

		res := models.Result{
			StartTime:    start.Format("2006-01-02 15:04:05"),
			EndTime:      end.Format("2006-01-02 15:04:05"),
			Subject:      opts.Mail.Subject,
			From:         fmt.Sprintf("%s <%s>", opts.Mail.Name, opts.MailServer.Username),
			AttackerName: opts.Mail.Name,
			URL:          opts.Url.Link,
			Custom:       opts.Mail.Custom,
			Targets:      sendingData,
		}

		if err := reports.CreateReport(output, "", format, &res); err != nil {
			logging.Errorf("Error creating report: %v", err)
		}

		/* will be used with tracking mails
		for {
			select {
			case <-c:
				os.Exit(1)
			}
		}*/
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("config", "c", "", "config filename")
	runCmd.Flags().StringP("template", "t", "", "template to use for report generation")
	runCmd.Flags().StringP("output", "o", "", "where to store output")
	runCmd.Flags().StringP("format", "f", "tpl", "tpl, xml, json")
}

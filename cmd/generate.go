package cmd

import (
	"os"

	"github.com/lateralusd/lateralus/config"
	"github.com/lateralusd/lateralus/logging"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate sample config",
	Run: func(cmd *cobra.Command, args []string) {
		filename, err := cmd.Flags().GetString("name")
		if err != nil {
			logging.Fatalf("Error occurred: %v", err)
		}

		c := config.Options{
			Url: config.Url{
				Generate: true,
				Link:     "https://www.google.com/?ident=<CHANGE>",
				Length:   10,
			},
			Mail: config.Mail{
				Name:    "Attacker",
				From:    "Not Attacker",
				Subject: "Not phishing mail",
			},
			Attack: config.Attack{
				Targets:        "targets.csv",
				Template:       "templates/sample",
				TrackingServer: "https://10.10.10.1",
			},
			MailServer: config.MailServer{
				Host:       "smtp.gmail.com",
				Port:       587,
				Username:   "testusername@gmail.com",
				Password:   "testPassword",
				Encryption: "tls",
			},
			General: config.General{
				Bulk:      false,
				BulkDelay: 60,
				BulkSize:  3,
				Delay:     5,
				Separator: ",",
			},
		}

		f, err := os.Create(filename)
		if err != nil {
			logging.Fatalf("Error occurred: %v", err)
		}
		defer f.Close()

		if err := yaml.NewEncoder(f).Encode(&c); err != nil {
			logging.Fatalf("Error occurred: %v", err)
		}
		logging.Infof("Configuration saved in \"%s\"", filename)
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("name", "n", "config.yaml", "filename where to generate config")
}

package cmd

import (
	"io/ioutil"

	"github.com/lateralusd/lateralus/logging"
	"github.com/spf13/cobra"
)

var sampleConfig = `url:
  generate: True
  link: "https://www.google.com/?ident=<CHANGE>"
  length: 10
  
mail:
  name: Attacker
  from: Not Attacker
  subject: Not phishing mail
  custom: ""
  
attack:
  targets: targets.csv
  template: ./template
  
mailServer:
  host: smtp.gmail.com
  port: 587
  username: "testusername@gmail.com"
  password: ""
  encryption: tls

general:
  bulk: False
  bcc: True
  bulkDelay: 60
  bulkSize: 3
  delay: 5
  separator: ";"
`

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate sample config",
	Run: func(cmd *cobra.Command, args []string) {
		filename, err := cmd.Flags().GetString("name")
		if err != nil {
			logging.Fatalf("Error ocurred: %v", err)
		}
		err = ioutil.WriteFile(filename, []byte(sampleConfig), 0600)
		if err != nil {
			logging.Fatalf("Error ocurred: %v", err)
		}
		logging.Infof("Configuration saved in \"%s\"", filename)
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("name", "n", "config.yaml", "filename where to generate config")
}

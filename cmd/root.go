package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "lateralus",
	Short: "terminal-based phishing-campaing tool",
	Long: `Simple to use terminal-based phishing campaign tool
				with a lot of customization options and report generations.
				Provides integration with modlishka.
		`,
}

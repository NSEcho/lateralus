package cmd

import (
	"os"
	"path/filepath"

	"github.com/lateralusd/lateralus/logging"
	"github.com/spf13/cobra"
)

var types = []string{
	"*.yaml",
	"*.csv",
	"*.rep",
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean project",
	Long:  "Delete all .yaml, .csv and .rep files",
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := cmd.Flags().GetString("dir")
		if err != nil {
			logging.Fatalf("Error occurred: %v", err)
		}

		var files []string
		for _, t := range types {
			f, err := getFiles(dir, t)
			if err != nil {
				logging.Fatalf("Error ocurred: %v", err)
			}
			files = append(files, f...)
		}

		if err := deleteFiles(files); err != nil {
			logging.Fatalf("Error ocurred: %v", err)
		}
	},
}

func getFiles(dir, t string) ([]string, error) {
	return filepath.Glob(dir + "/" + t)
}

func deleteFiles(files []string) error {
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	RootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringP("dir", "d", ".", "which project to clear")
}

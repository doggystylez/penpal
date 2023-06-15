package main

import (
	"github.com/doggystylez/penpal/config"
	"github.com/doggystylez/penpal/scan"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "penpal",
		Short: "monitor validator signing",
		Run: func(cmd *cobra.Command, args []string) {
			path, err := cmd.Flags().GetString("config")
			if err != nil {
				panic(err)
			}
			cfg, err := config.Load(path)
			if err != nil {
				panic(err)
			}
			scan.Monitor(cfg)
		},
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.Flags().StringP("config", "c", "", "path to config")
	err := cmd.MarkFlagRequired("config")
	if err != nil {
		return
	}
	err = cmd.Execute()
	if err != nil {
		return
	}
}

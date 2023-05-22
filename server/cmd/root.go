// Package cmd to make it cmd app
/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/codescalers/cloud4students/app"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloud4Students",
	Short: "Cloud helps students to deploy their projects",
	Long: `Cloud for students helps them to deploy their projects with 
	applying for a voucher, For example :
		They can deploy virtual machine that can be small, medium or large.
		They can deploy Kubernetes that can be small, medium, or large.
		The Amount of resources available will depend on their voucher.
		`,

	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		app, err := app.NewApp(cmd.Context(), configFile)
		if err != nil {
			return fmt.Errorf("failed to create new app: %w", err)
		}

		err = app.Start(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to start app: %w", err)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	err := rootCmd.Execute()
	if err != nil {
		log.Err(err).Send()
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("config", "c", "./config.json", "Enter your configurations path")
}

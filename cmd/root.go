/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/mercedes-benz/gitflow-cli/cmd/hotfix"
	"github.com/mercedes-benz/gitflow-cli/cmd/release"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Configuration file of the workflow automation command line tool.
var cfgFile string

// RootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Args: cobra.NoArgs,
	Use:  "gitflow-cli",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}

// Initialize Cobra flags and configuration settings.
func init() {
	// sets the passed functions to be run when each command's Execute method is called
	cobra.OnInitialize(initConfiguration)

	// add subcommands to the root command
	rootCmd.AddCommand(release.ReleaseCmd, hotfix.HotfixCmd)

	// persistent flags, which, if defined here, will be global for the application
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.gitflow-cli.yaml)")
}

// Read in Viper config file and environment variables if set.
func initConfiguration() {
	if cfgFile != "" {
		// use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// search config in home directory with name ".gitflow-cli" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gitflow-cli")
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

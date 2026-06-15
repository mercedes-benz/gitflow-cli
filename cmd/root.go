/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mercedes-benz/gitflow-cli/cmd/hotfix"
	"github.com/mercedes-benz/gitflow-cli/cmd/release"
	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
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
func Execute() error {
	return rootCmd.Execute()
}

// Initialize Cobra flags and configuration settings.
func init() {
	rootCmd.Version = buildVersion()
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	// sets the passed functions to be run when each command's ExecuteHook method is called
	cobra.OnInitialize(initConfiguration)

	// set up interactive prompts (branch resolver, docker fallback)
	initPrompts()

	// add subcommands to the root command
	rootCmd.AddCommand(release.ReleaseCmd, hotfix.HotfixCmd)

	// persistent flags, which, if defined here, will be global for the application
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.gitflow-cli.yaml)")
	rootCmd.PersistentFlags().StringVarP(&core.ProjectPath, "path", "p", ".", "path to git repository (default is current directory)")
	rootCmd.PersistentFlags().Bool("docker-mode", false, "run plugin commands inside a Docker container")
	rootCmd.PersistentFlags().Bool("native-mode", false, "run plugin commands natively on the host (default)")
	rootCmd.PersistentFlags().Bool("no-push", false, "do not push changes to remote repository")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "automatically confirm all interactive prompts")
	rootCmd.MarkFlagsMutuallyExclusive("docker-mode", "native-mode")
}

// Read in Viper config file and environment variables if set.
func initConfiguration() {
	if docker, _ := rootCmd.Flags().GetBool("docker-mode"); docker {
		plugin.ExecutorModeOverride = plugin.ModeDocker
	} else if native, _ := rootCmd.Flags().GetBool("native-mode"); native {
		plugin.ExecutorModeOverride = plugin.ModeNative
	}

	if noPush, _ := rootCmd.Flags().GetBool("no-push"); noPush {
		viper.Set("core.push", false)
	}

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
	} else if cfgFile == "" {
		if err := initDefaultConfig(); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not create default config:", err)
		} else {
			_ = viper.ReadInConfig()
		}
	}
}

const defaultConfig = `core:
  # Branch names
  production: main
  development: develop
  release: release
  hotfix: hotfix

  # Behavior
  push: true
  undo: false
  docker-fallback: false

  # Logging: off, stdout, stderr, cmdline, output (combinable)
  logging: "off"
`

func initDefaultConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".gitflow-cli.yaml")

	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Created default config file:", configPath)
	return nil
}

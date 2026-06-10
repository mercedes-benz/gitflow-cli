/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package main

import (
	"os"

	"github.com/mercedes-benz/gitflow-cli/cmd"

	// import the plugin package so that init functions for all plugins are executed automatically
	_ "github.com/mercedes-benz/gitflow-cli/plugin"
)

// Entry point of the workflow automation command line tool.
func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

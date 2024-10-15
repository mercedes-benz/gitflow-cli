/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package main

import (
	"github.com/mercedes-benz/gitflow-cli/cmd"

	// import the plugin package so that init functions for all plugins are executed automatically
	_ "github.com/mercedes-benz/gitflow-cli/plugin"
)

// Entry point of the workflow automation command line tool.
func main() {
	// execute the Cobra root command
	cmd.Execute()
}

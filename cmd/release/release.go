/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package release

import (
	"os"

	"github.com/mercedes-benz/gitflow-cli/plugin/core"
	"github.com/spf13/cobra"
)

// ReleaseCmd represents the release subcommand of RootCmd.
var ReleaseCmd = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "release",
	Short: "Prepare a new production release",

	Long: `Prepare a new production release.

Release is a stage of the software development process where the code in the develop
branch has reached a stable point and is ready to be released into the master branch.

When the develop branch has acquired enough features for a release, a new 
branch is created. The name of the branch typically starts with 'release/' 
followed by a version number and an optional brief description of the release.
This branch is used to prepare for a new production release. It allows for 
last-minute dotting of i's and crossing t's: minor bug fixes, preparing 
meta-data like version number, build dates etc.

Once the team is satisfied with the state of the release branch, it is merged
into master and tagged with a version number. In addition, it should be merged
back into develop, which may have progressed since the release was initiated.

By doing this, the master branch always reflects the latest released and 
production-ready state of the software.`,
}

// Required for all plugin operations that execute workflow automation commands in a project directory.
var projectPath string

// Version increase flags for the start subcommand.
var major, minor bool

// StartCmd represents the start subcommand of ReleaseCmd.
var startCmd = &cobra.Command{
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "start",
	Short:        "Create a new production release branch",

	Long: `Create a new production release branch.

When the develop branch has acquired enough features for a release, a new 
branch is created. This branch is used to prepare for a new production 
release.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// ensure that only one version increase flag is set
		if major {
			minor = false
		}

		if minor {
			major = false
		}

		return core.Start(core.Release, projectPath, major, minor)
	},
}

// FinishCmd represents the finish subcommand of ReleaseCmd.
var finishCmd = &cobra.Command{
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "finish",
	Short:        "Finish the current production release branch",

	Long: `Finish the current production release branch.

Once the team is satisfied with the state of the release branch, it is merged
into master and tagged with a version number.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return core.Finish(core.Release, projectPath)
	},
}

// Initialize Cobra flags for the release subcommand.
func init() {
	// current working directory as default project path
	defaultPath, _ := os.Getwd()

	// add subcommands to the release command
	ReleaseCmd.AddCommand(startCmd, finishCmd)

	// persistent flags, which, if defined here, will be global for this command and all subcommands
	ReleaseCmd.PersistentFlags().
		StringVarP(&projectPath, "path", "p", defaultPath, "project path for workflow automation commands")

	// flags which will only run when this command is called directly
	startCmd.Flags().BoolVarP(&major, "major", "j", false, "increase major version number for a new release")
	startCmd.Flags().BoolVarP(&minor, "minor", "n", true, "increase minor version number for a new release")

	// enforce rules for the flags
	startCmd.MarkPersistentFlagDirname("path")
	startCmd.MarkFlagsMutuallyExclusive("major", "minor")
}

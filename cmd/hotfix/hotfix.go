/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package hotfix

import (
	"github.com/mercedes-benz/gitflow-cli/core"

	"github.com/spf13/cobra"
)

// HotfixCmd represents the hotfix subcommand of RootCmd.
var HotfixCmd = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "hotfix",
	Short: "Quickly patch a production release",

	Long: `Quickly patch a production release.

Hotfix is a type of branch used to quickly patch a production release. Hotfix branches 
are very much like release branches except they're based on master instead of develop.

Hotfix branches are created when there's a need to quickly fix an issue in the
production version of the software. The name of the branch typically starts 
with 'hotfix/' followed by a version number and an optional brief description 
of the fix.

Once the fix is complete, the hotfix branch is merged back into both master 
and develop (or the current release branch), so that the fix is included in the
next release as well. The master branch is then tagged with the updated 
production version number.

This way, the Gitflow model ensures that fixes for urgent production bugs can
be delivered quickly, without interrupting ongoing development work.`,
}

// StartCmd represents the start subcommand of HotfixCmd.
var startCmd = &cobra.Command{
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "start",
	Short:        "Create a new hotfix branch",

	Long: `Create a new hotfix branch.

Hotfix branches are created when there's a need to quickly fix an issue in the
production version of the software.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return core.Start(core.Hotfix, core.ProjectPath)
	},
}

// FinishCmd represents the finish subcommand of HotfixCmd.
var finishCmd = &cobra.Command{
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "finish",
	Short:        "Finish the current hotfix branch",

	Long: `Finish the current hotfix branch.

Once the fix is complete, the hotfix branch is merged back into both master and
develop (or the current release branch)`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return core.Finish(core.Hotfix, core.ProjectPath)
	},
}

// Initialize Cobra flags for the hotfix subcommand.
func init() {
	// add subcommands to the hotfix command
	HotfixCmd.AddCommand(startCmd, finishCmd)
}

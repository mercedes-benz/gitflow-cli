/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package fallback

import (
	"github.com/mercedes-benz/gitflow-cli/test/helper"
	"testing"
)

// TestHotfixStart without version file and fallback to standard plugin
func TestHotfixStart(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// main -> no version file
	// develop -> no version file

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// check hotfix branch state
	// standard plugin creates version file in main
	env.AssertFileEquals("version.txt", "1.0.0", "main")
	env.AssertCommitMessageEquals("Create versions file", "main")

	// check hotfix branch state
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertFileEquals("version.txt", "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Set next hotfix version.", "hotfix/1.0.1")

	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package fallback

import (
	"github.com/mercedes-benz/gitflow-cli/test/helper"
	"testing"
)

// TestHotfixStartMaven with Maven plugin and Maven preconditions
func TestHotfixStartMaven(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// main -> version.txt (1.0.0)
	// develop -> version.txt (1.1.0-dev)

	env.CommitFile("version.txt", "1.0.0", "Set up test precondition for main branch", "main")
	env.CommitFile("version.txt", "1.1.0-dev", "Set up test precondition for develop branch", "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// check hotfix branch state
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertFileEquals("version.txt", "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Set next hotfix version.", "hotfix/1.0.1")

	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

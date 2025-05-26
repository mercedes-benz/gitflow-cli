/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package fallback

import (
	"github.com/mercedes-benz/gitflow-cli/test/helper"
	"testing"
)

// TestReleaseStartStandard with standard plugin and standard preconditions
func TestReleaseStartStandard(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// main -> version.txt (1.0.0)
	// develop -> version.txt (1.1.0-dev)

	env.CommitFile("version.txt", "1.0.0", "Set up test precondition for main branch", "main")
	env.CommitFile("version.txt", "1.1.0-dev", "Set up test precondition for develop branch", "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "start")

	// THEN:
	// check release branch state
	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchExists("origin/release/1.1.0")

	env.AssertFileEquals("version.txt", "1.1.0", "release/1.1.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.1.0")

	env.AssertCurrentBranchEquals("release/1.1.0")
}

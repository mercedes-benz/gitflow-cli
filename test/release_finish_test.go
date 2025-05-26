/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package test

import (
	"github.com/mercedes-benz/gitflow-cli/test/base"
	"testing"
)

// TestReleaseStart checks if the command "release finish" performs the correct Git graph manipulation
func TestReleaseFinish(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := base.SetupTestEnv(t)

	env.CommitFile("version.txt", "1.0.0-dev", "Set up test precondition for develop branch", "develop")

	env.CreateBranch("release/1.0.0", "develop")
	env.CommitFile("version.txt", "1.0.0", "Set up test precondition for release branch.", "release/1.0.0")

	// WHEN
	env.ExecuteGitflow("release", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0'", "main")
	env.AssertTagEquals("1.0.0", "main")
	env.AssertFileEquals("version.txt", "1.0.0", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Set next minor project version.", "develop")
	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0' into develop", "develop", 1)
	env.AssertFileEquals("version.txt", "1.1.0-dev", "develop")

	env.AssertBranchDoesNotExist("release/1.0.0")
	env.AssertCurrentBranchEquals("develop")
}

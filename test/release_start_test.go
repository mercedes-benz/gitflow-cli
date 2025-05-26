/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package test

import (
	"github.com/mercedes-benz/gitflow-cli/test/base"
	"testing"
)

// TestReleaseStart checks if the command "release start" performs the correct Git graph manipulation
func TestReleaseStart(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := base.SetupTestEnv(t)

	env.CommitFile("version.txt", "1.0.0-dev", "Add version file", "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "start")

	// THEN:
	// check release branch state
	env.AssertBranchExists("release/1.0.0")
	env.AssertBranchExists("origin/release/1.0.0")

	env.AssertFileEquals("version.txt", "1.0.0", "release/1.0.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.0.0")

	env.AssertCurrentBranchEquals("release/1.0.0")
}

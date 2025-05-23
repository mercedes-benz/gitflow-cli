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
func TestReleaseFinish(t *testing.T) {
	// GIVEN
	env := base.SetupTestEnv(t)

	env.CommitFile("develop", "version.txt", "1.0.0-dev", "Add version file")
	env.CommitFile("release/1.0.0", "version.txt", "1.0.0", "Remove qualifier from project version.")

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
}

/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package fallback

import (
	"github.com/mercedes-benz/gitflow-cli/test/helper"
	"testing"
)

// TestReleaseStartFallback without version file and fallback to standard plugin
func TestReleaseStartFallback(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// main -> no version file
	// develop -> no version file

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "start")

	// THEN:
	// check develop branch
	// standard plugin creates version file in develop
	env.AssertFileEquals("version.txt", "1.0.0-dev", "develop")
	env.AssertCommitMessageEquals("Create versions file", "develop")

	// check release branch state
	env.AssertBranchExists("release/1.0.0")
	env.AssertBranchExists("origin/release/1.0.0")

	env.AssertFileEquals("version.txt", "1.0.0", "release/1.0.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.0.0")

	env.AssertCurrentBranchEquals("release/1.0.0")
}

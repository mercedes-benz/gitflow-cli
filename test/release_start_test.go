/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package test

import (
	"github.com/mercedes-benz/gitflow-cli/test/base"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestReleaseStart checks if the command "release start" performs the correct Git graph manipulation
func TestReleaseStart(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := base.SetupTestEnv(t)

	env.CommitFile("develop", "version.txt", "1.0.0-dev", "Add version file")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "start")

	// THEN: The release branch should have been created
	releaseBranch := "release/1.0.0"

	env.AssertBranchExists(releaseBranch)
	env.AssertBranchExists("origin/" + releaseBranch)

	assert.Equal(t, releaseBranch, env.GetCurrentBranch(), "Current branch should be the release branch")

	commitMessage := env.GetCommitMessage(releaseBranch)
	assert.Equal(t, "Remove qualifier from project version.", commitMessage, "Commit message should indicate removing qualifier from project version.'")

	env.AssertFileInBranchEquals(releaseBranch, "version.txt", "1.0.0")
}

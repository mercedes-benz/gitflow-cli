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

	// Create initial commit on production branch
	env.CreateFile("README.md", "# Temporary Test Repository")
	env.ExecuteGit("add", "README.md")
	env.ExecuteGit("commit", "-m", "Initial commit")
	env.ExecuteGit("branch", "-m", "main")

	// Push to remote
	env.ExecuteGit("push", "-u", "origin", "main")

	// Create development branch
	env.ExecuteGit("checkout", "-b", "develop")
	env.CreateFile("version.txt", "1.0.0-dev")
	env.ExecuteGit("add", "version.txt")
	env.ExecuteGit("commit", "-m", "Add version file")
	env.ExecuteGit("push", "-u", "origin", "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	output := env.ExecuteGitflow("release", "start")
	t.Logf("Command output: %s", output)

	// THEN: The release branch should have been created
	releaseBranch := "release/1.0.0"

	env.AssertBranchExists(releaseBranch)
	env.AssertBranchExists("origin/" + releaseBranch)

	currentBranch := env.GetCurrentBranch()
	assert.Equal(t, releaseBranch, currentBranch, "Current branch should be the release branch")

	env.AssertCommitsAhead(releaseBranch, "develop", 1)
	commitMessage := env.GetCommitMessage(releaseBranch)
	assert.Equal(t, "Remove qualifier from project version.", commitMessage, "Commit message should be 'Remove qualifier from project version.'")

	// The version.txt in the release branch should be correctly updated
	env.AssertFileInBranchEquals(releaseBranch, "version.txt", "1.0.0")
}

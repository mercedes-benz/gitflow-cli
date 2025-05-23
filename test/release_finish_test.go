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
func TestReleaseFinish(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := base.SetupTestEnv(t)

	// develop branch has version file with 1.0.0-dev
	env.ExecuteGit("checkout", "develop")
	env.WriteFile("version.txt", "1.0.0-dev")
	env.ExecuteGit("add", "version.txt")
	env.ExecuteGit("commit", "-m", "Add version file")
	env.ExecuteGit("push", "-u", "origin", "develop")

	// create release branch
	env.ExecuteGit("checkout", "-b", "release/1.0.0")
	env.WriteFile("version.txt", "1.0.0")
	env.ExecuteGit("add", "version.txt")
	env.ExecuteGit("commit", "-m", "Remove qualifier from project version.")
	env.ExecuteGit("push", "-u", "origin", "release/1.0.0")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "finish")

	// THEN: The release branch should be merged into main
	env.AssertBranchExists("main")
	env.ExecuteGit("checkout", "main")
	env.ExecuteGit("pull", "origin", "main")

	// Check that the release branch is merged into main
	assert.Equal(t, "Merge branch 'release/1.0.0'", env.GetCommitMessage("main"), "")

	// Check that the commit in main is tagged with 1.0.0
	// todo: add branch als parameter
	assert.Equal(t, "1.0.0", env.GetTag())

	// todo: check commit message (into develop must be deleted)
	assert.Equal(t, "Merge branch 'release/1.0.0' into develop", env.GetCommitMessage("develop", 1))
	//assert.Equal(t, env.GetCommitMessage("develop"), "Set next minor project version.")

	// Check that a commit was created in develop to update the version to 1.0.0-dev
	env.AssertFileInBranchEquals("develop", "version.txt", "1.1.0-dev")

	// Verify the release branch was deleted
	branches := env.ExecuteGit("branch", "-a")
	assert.NotContains(t, branches, "release/1.0.0", "Release branch should be deleted")

}

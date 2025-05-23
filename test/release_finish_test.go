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

	// Create initial commit on production branch
	env.CreateFile("README.md", "# Temporary Test Repository")
	env.ExecuteGit("add", "README.md")
	env.ExecuteGit("commit", "-m", "Initial commit")
	env.ExecuteGit("branch", "-m", "main")
	env.ExecuteGit("push", "-u", "origin", "main")

	// Create development branch
	env.ExecuteGit("checkout", "-b", "develop")
	env.CreateFile("version.txt", "1.0.0-dev")
	env.ExecuteGit("add", "version.txt")
	env.ExecuteGit("commit", "-m", "Add version file")
	env.ExecuteGit("push", "-u", "origin", "develop")

	// create release branch
	env.ExecuteGit("checkout", "-b", "release/1.0.0")
	env.CreateFile("version.txt", "1.0.0")
	env.ExecuteGit("add", "version.txt")
	env.ExecuteGit("commit", "-m", "Change version file")
	env.ExecuteGit("push", "-u", "origin", "release/1.0.0")

	// WHEN: The command "gitflow-cli release start" is executed
	output := env.ExecuteGitflow("release", "finish")
	t.Logf("Command output: %s", output)

	// THEN: The release branch should have been created

	// THEN: The release branch should be merged into main
	env.AssertBranchExists("main")
	env.ExecuteGit("checkout", "main")
	env.ExecuteGit("pull", "origin", "main")

	// Check that the release branch is merged into main
	commitMessage := env.GetCommitMessage("main")
	assert.Contains(t, commitMessage, "Merge branch 'release/1.0.0'", "Release branch should be merged into main")

	// Check that the commit in main is tagged with 1.0.0
	tags := env.ExecuteGit("tag", "--points-at", "HEAD")
	assert.Contains(t, tags, "1.0.0", "The commit in main should be tagged with 1.0.0")

	// Check that the release branch is merged into develop
	env.ExecuteGit("checkout", "develop")
	env.ExecuteGit("pull", "origin", "develop")
	developCommitMessage := env.GetCommitMessage("develop")
	assert.Contains(t, developCommitMessage, "Set next minor project version.")

	// Check that a commit was created in develop to update the version to 1.0.0-dev
	env.AssertFileInBranchEquals("develop", "version.txt", "1.1.0-dev")

	// Verify the release branch was deleted
	branches := env.ExecuteGit("branch", "-a")
	assert.NotContains(t, branches, "release/1.0.0", "Release branch should be deleted")

}

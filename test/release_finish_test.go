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
	// GIVEN
	env := base.SetupTestEnv(t)

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

	// WHEN
	env.ExecuteGitflow("release", "finish")

	// THEN
	assert.Equal(t, "Merge branch 'release/1.0.0'", env.GetCommitMessage("main"), "Commit message in main branch should indicate merge from release branch")

	assert.Equal(t, "1.0.0", env.GetTag("main"), "Latest commit in main should be tagged with 1.0.0")

	assert.Equal(t, "Set next minor project version.", env.GetCommitMessage("develop"), "Latest commit in develop should update version for next development cycle")
	assert.Equal(t, "Merge branch 'release/1.0.0' into develop", env.GetCommitMessage("develop", 1), "Second-to-last commit in develop should be the merge from release branch")

	env.AssertFileInBranchEquals("develop", "version.txt", "1.1.0-dev")

	branches := env.ExecuteGit("branch", "-a")
	assert.NotContains(t, branches, "release/1.0.0", "Release branch should be deleted")
}

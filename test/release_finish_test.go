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

	env.CommitFile("develop", "version.txt", "1.0.0-dev", "Add version file")
	env.CommitFile("release/1.0.0", "version.txt", "1.0.0", "Remove qualifier from project version.")

	// WHEN
	env.ExecuteGitflow("release", "finish")

	// THEN

	// main branch
	assert.Equal(t, "Merge branch 'release/1.0.0'", env.GetCommitMessage("main"))
	assert.Equal(t, "1.0.0", env.GetTag("main"))

	// develop branch
	assert.Equal(t, "Set next minor project version.", env.GetCommitMessage("develop"))
	assert.Equal(t, "Merge branch 'release/1.0.0' into develop", env.GetCommitMessage("develop", 1))

	env.AssertFileInBranchEquals("develop", "version.txt", "1.1.0-dev")

	env.AssertBranchDoesNotExist("release/1.0.0")
}

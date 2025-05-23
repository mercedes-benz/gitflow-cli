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

	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0'", "main", 0)
	assert.Equal(t, "1.0.0", env.GetTag("main"), "Latest commit in main should be tagged with release version")
	env.AssertFileInBranchEquals("main", "version.txt", "1.0.0")

	// Check develop branch state
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0' into develop", "develop", 1)
	env.AssertFileInBranchEquals("develop", "version.txt", "1.1.0-dev")

	env.AssertBranchDoesNotExist("release/1.0.0")
}

/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package fallback

import (
	"github.com/mercedes-benz/gitflow-cli/test/helper"
	"path/filepath"
	"testing"
)

// TestReleaseFinishFallback without version file and fallback to standard plugin
func TestReleaseFinishFallback(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Path to the version file template
	versionFileTemplate := filepath.Join("../..", "helper", "templates", "version.txt.tpl")

	// main -> no version file
	// develop -> version.txt (1.0.0-dev)
	// release/1.0.0 -> version.txt (1.0.0)

	env.CommitFileFromTemplate(versionFileTemplate, "1.0.0-dev", "develop")
	env.CreateBranch("release/1.0.0", "develop")
	env.CommitFileFromTemplate(versionFileTemplate, "1.0.0", "release/1.0.0")

	// WHEN
	env.ExecuteGitflow("release", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0'", "main")
	env.AssertTagEquals("1.0.0", "main")
	env.AssertVersionEquals(versionFileTemplate, "1.0.0", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0' into develop", "develop", 1)
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
	env.AssertVersionEquals(versionFileTemplate, "1.1.0-dev", "develop")

	env.AssertBranchDoesNotExist("release/1.0.0")
	env.AssertCurrentBranchEquals("develop")
}

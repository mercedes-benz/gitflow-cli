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

// TestReleaseFinishStandard with standard plugin and standard preconditions
func TestReleaseFinishStandard(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Path to the version file template
	versionTemplate := filepath.Join("..", "..", "templates", "version.txt.tpl")

	// main -> version.txt (1.0.0)
	// develop -> version.txt (1.1.0-dev)
	// release/1.1.0 -> version.txt (1.1.0)

	env.CommitFileFromTemplate(versionTemplate, "1.0.0", "main")
	env.CommitFileFromTemplate(versionTemplate, "1.1.0-dev", "develop")
	env.CreateBranch("release/1.1.0", "develop")
	env.CommitFileFromTemplate(versionTemplate, "1.1.0", "release/1.1.0")

	// WHEN
	env.ExecuteGitflow("release", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'release/1.1.0'", "main")
	env.AssertTagEquals("1.1.0", "main")
	env.AssertFileEquals("version.txt", "1.1.0", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'release/1.1.0' into develop", "develop", 1)
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
	env.AssertFileEquals("version.txt", "1.2.0-dev", "develop")

	env.AssertBranchDoesNotExist("release/1.1.0")
	env.AssertCurrentBranchEquals("develop")
}

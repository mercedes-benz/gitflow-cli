/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package fallback

import (
	"github.com/mercedes-benz/gitflow-cli/test/helper"
	"testing"
)

// TestHotfixFinish without version file and fallback to standard plugin
func TestHotfixFinishFallback(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// main -> version.txt (1.0.0)
	// develop -> version.txt (1.0.0-dev)
	// release/1.0.0 -> version.txt (1.0.0)

	env.CommitFile("version.txt", "1.0.0", "Set up test precondition for release branch.", "main")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitFile("version.txt", "1.0.1", "Set up test precondition for release branch.", "hotfix/1.0.1")

	// WHEN
	env.ExecuteGitflow("hotfix", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertFileEquals("version.txt", "1.0.1", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into develop", "develop", 1)
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)

	env.AssertFileEquals("version.txt", "1.1.0-dev", "develop")

	env.AssertBranchDoesNotExist("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

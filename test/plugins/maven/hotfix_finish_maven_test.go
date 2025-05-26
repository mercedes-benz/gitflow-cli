/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package fallback

import (
	"github.com/mercedes-benz/gitflow-cli/test/helper"
	"testing"
)

// TestHotfixFinishMaven with Maven plugin and Maven preconditions
func TestHotfixFinishMaven(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// main -> version.txt (1.0.0)
	// develop -> version.txt (1.1.0-dev)
	// hotfix/1.0.1 -> version.txt (1.0.1)

	env.CommitFile("version.txt", "1.0.0", "Set up test precondition for main branch", "main")
	env.CommitFile("version.txt", "1.1.0-dev", "Set up test precondition for develop branch", "develop")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitFile("version.txt", "1.0.1", "Set up test precondition for hotfix branch", "hotfix/1.0.1")

	// WHEN
	env.ExecuteGitflow("hotfix", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertFileEquals("version.txt", "1.0.1", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into develop\n\n# Conflicts:\n#\tversion.txt", "develop", 0)
	env.AssertFileEquals("version.txt", "1.1.0-dev", "develop")

	env.AssertBranchDoesNotExist("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

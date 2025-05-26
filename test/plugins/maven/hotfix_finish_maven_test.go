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

// TestHotfixFinishStandard with standard plugin and standard preconditions
func TestHotfixFinishStandard(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Path to the templates
	pomTemplate := filepath.Join("..", "..", "templates", "pom.xml.tpl")

	// main -> pom.xml (1.0.0)
	// develop -> pom.xml (1.1.0-SNAPSHOT)
	// hotfix/1.0.1 -> pom.xml (1.0.1)

	env.CommitFileFromTemplate(pomTemplate, "1.0.0", "main")
	env.CommitFileFromTemplate(pomTemplate, "1.1.0-SNAPSHOT", "develop")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitFileFromTemplate(pomTemplate, "1.0.1", "hotfix/1.0.1")

	// WHEN
	env.ExecuteGitflow("hotfix", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertVersionEquals(pomTemplate, "1.0.1", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into develop", "develop", 0)
	env.AssertVersionEquals(pomTemplate, "1.1.0-SNAPSHOT", "develop")

	env.AssertBranchDoesNotExist("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

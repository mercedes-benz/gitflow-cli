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

// TestHotfixStartStandard with standard plugin and standard preconditions
func TestHotfixStartStandard(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Path to the version file template
	versionTemplate := filepath.Join("..", "..", "templates", "version.txt.tpl")

	// main -> version.txt (1.0.0)
	// develop -> version.txt (1.1.0-dev)

	env.CommitFileFromTemplate(versionTemplate, "1.0.0", "main")
	env.CommitFileFromTemplate(versionTemplate, "1.1.0-dev", "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// check hotfix branch state
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertVersionEquals(versionTemplate, "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Set next hotfix version.", "hotfix/1.0.1")

	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

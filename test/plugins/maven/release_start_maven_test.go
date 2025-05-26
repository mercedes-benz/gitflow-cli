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

// TestReleaseStartStandard with standard plugin and standard preconditions
func TestReleaseStartStandard(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Path to the version file template
	versionTemplate := filepath.Join("..", "..", "templates", "version.txt.tpl")

	// main -> version.txt (1.0.0)
	// develop -> version.txt (1.1.0-dev)

	env.CommitFileFromTemplate(versionTemplate, "1.0.0", "main")
	env.CommitFileFromTemplate(versionTemplate, "1.1.0-dev", "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "start")

	// THEN:
	// check release branch state
	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchExists("origin/release/1.1.0")

	env.AssertVersionEquals(versionTemplate, "1.1.0", "release/1.1.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.1.0")

	env.AssertCurrentBranchEquals("release/1.1.0")
}

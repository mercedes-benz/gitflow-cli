/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"github.com/mercedes-benz/gitflow-cli/test/helper"
	"path/filepath"
	"testing"
)

// TestHotfixStart tests Hotfix Start with different templates
func TestHotfixStart(t *testing.T) {
	// Test with version.txt template
	t.Run("Test Standard Plugin", func(t *testing.T) {
		testHotfixStart(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("Test Maven Plugin", func(t *testing.T) {
		testHotfixStart(t, "pom.xml.tpl", "SNAPSHOT")
	})
}

// testHotfixStart runs the test with the specified template
func testHotfixStart(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	versionFileTemplate := filepath.Join("../..", "helper", "templates", templateName)

	// main -> template file (1.0.0)
	// develop -> template file (1.1.0-dev/1.1.0-SNAPSHOT)

	env.CommitFileFromTemplate(versionFileTemplate, "1.0.0", "main")
	env.CommitFileFromTemplate(versionFileTemplate, "1.1.0-"+versionQualifier, "develop")

	// WHEN: The command "gitflow-cli hotfix start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// check hotfix branch state
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertVersionEquals(versionFileTemplate, "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Set next hotfix version.", "hotfix/1.0.1")

	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

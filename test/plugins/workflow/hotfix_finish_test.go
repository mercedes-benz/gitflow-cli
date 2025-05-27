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

// TestHotfixFinish tests Hotfix Finish with different templates
func TestHotfixFinish(t *testing.T) {
	// Test with version.txt template
	t.Run("Test Standard Plugin", func(t *testing.T) {
		testHotfixFinish(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("Test Maven Plugin", func(t *testing.T) {
		testHotfixFinish(t, "pom.xml.tpl", "SNAPSHOT")
	})
}

// testHotfixFinish runs the test with the specified template
func testHotfixFinish(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	versionFileTemplate := filepath.Join("../..", "helper", "templates", templateName)

	// main -> template file (1.0.0)
	// develop -> template file (1.1.0-dev/1.1.0-SNAPSHOT)
	// hotfix/1.0.1 -> template file (1.0.1)

	env.CommitFileFromTemplate(versionFileTemplate, "1.0.0", "main")
	env.CommitFileFromTemplate(versionFileTemplate, "1.1.0-"+versionQualifier, "develop")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitFileFromTemplate(versionFileTemplate, "1.0.1", "hotfix/1.0.1")

	// WHEN: The command "gitflow-cli hotfix finish" is executed
	env.ExecuteGitflow("hotfix", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertVersionEquals(versionFileTemplate, "1.0.1", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into develop", "develop", 0)
	env.AssertVersionEquals(versionFileTemplate, "1.1.0-"+versionQualifier, "develop")

	env.AssertBranchDoesNotExist("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

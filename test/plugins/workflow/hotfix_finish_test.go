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

// Test hotfix finish job
func TestHotfixFinish(t *testing.T) {
	// Test with version.txt template
	t.Run("StandardPlugin", func(t *testing.T) {
		testHotfixFinish(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("MavenPlugin", func(t *testing.T) {
		testHotfixFinish(t, "pom.xml.tpl", "SNAPSHOT")
	})

	// Test fallback without versioning file
	t.Run("StandardPluginFallback", func(t *testing.T) {
		testHotfixFinishFallback(t)
	})
}

// testHotfixFinish runs the test with the specified template
func testHotfixFinish(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	template := filepath.Join("../..", "helper", "templates", templateName)

	// main -> version file (1.0.0)
	// develop -> version file (1.1.0-{qualifier})
	// hotfix/1.0.1 -> version file (1.0.1)

	env.CommitFileFromTemplate(template, "1.0.0", "main")
	env.CommitFileFromTemplate(template, "1.1.0-"+versionQualifier, "develop")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitFileFromTemplate(template, "1.0.1", "hotfix/1.0.1")

	// WHEN: The command "gitflow-cli hotfix finish" is executed
	env.ExecuteGitflow("hotfix", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertVersionEquals(template, "1.0.1", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into develop", "develop", 0)
	env.AssertVersionEquals(template, "1.1.0-"+versionQualifier, "develop")

	env.AssertBranchDoesNotExist("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

// TestHotfixFinishFallback (test standard plugin with additional functionality)
func testHotfixFinishFallback(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Path to the templates
	template := filepath.Join("../..", "helper", "templates", "version.txt.tpl")

	// main -> version.txt (1.0.0)
	// develop -> no version file
	// hotfix/1.0.1 -> version.txt (1.0.1)

	env.CommitFileFromTemplate(template, "1.0.0", "main")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitFileFromTemplate(template, "1.0.1", "hotfix/1.0.1")

	// WHEN: The command "gitflow-cli hotfix finish" is executed
	env.ExecuteGitflow("hotfix", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertVersionEquals(template, "1.0.1", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into develop", "develop", 1)
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
	env.AssertVersionEquals(template, "1.1.0-dev", "develop")

	env.AssertBranchDoesNotExist("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

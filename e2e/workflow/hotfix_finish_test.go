/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"path/filepath"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/e2e/helper"
)

// Test hotfix finish job
func TestHotfixFinish(t *testing.T) {
	// Test with version.txt template
	t.Run("StandardPlugin", func(t *testing.T) {
		testHotfixFinish(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("MvnPlugin", func(t *testing.T) {
		testHotfixFinish(t, "pom.xml.tpl", "SNAPSHOT")
	})

	// Test with package.json template
	t.Run("NpmPlugin", func(t *testing.T) {
		testHotfixFinish(t, "package.json.tpl", "dev")
	})

	// Test with pyproject.toml template
	t.Run("PythonPlugin_Pyproject", func(t *testing.T) {
		testHotfixFinish(t, "pyproject.toml.tpl", "dev")
	})

	// Test with setup.py template
	t.Run("PythonPlugin_SetupPy", func(t *testing.T) {
		testHotfixFinish(t, "setup.py.tpl", "dev")
	})

	// Test with composer.json template
	t.Run("ComposerPlugin", func(t *testing.T) {
		testHotfixFinish(t, "composer.json.tpl", "dev")
	})

	// Test with road.yaml template
	t.Run("RoadPlugin", func(t *testing.T) {
		testHotfixFinish(t, "road.yaml.tpl", "dev")
	})

	// Test fallback without versioning file
	t.Run("NoPluginFallback", func(t *testing.T) {
		testHotfixFinishFallback(t)
	})
}

// testHotfixFinish runs the test with the specified template
func testHotfixFinish(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	template := filepath.Join("..", "helper", "templates", templateName)

	// main -> version file (1.0.0)
	// develop -> version file (1.1.0-{qualifier})
	// release/1.1.0 -> version file (1.1.0)
	// hotfix/1.0.1 -> version file (1.0.1)

	env.CommitFileFromTemplate(template, "1.0.0", "main")
	env.CommitFileFromTemplate(template, "1.1.0-"+versionQualifier, "develop")

	env.CreateBranch("release/1.1.0", "develop")
	env.CommitFileFromTemplate(template, "1.1.0", "release/1.1.0")

	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitFileFromTemplate(template, "1.0.1", "hotfix/1.0.1")

	// WHEN: The command "gitflow-cli hotfix finish" is executed
	env.ExecuteGitflow("hotfix", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertVersionEquals(template, "1.0.1", "main")

	// Check release branch state - should be merged but version stays as it was
	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into release/1.1.0", "release/1.1.0", 0)
	env.AssertVersionEquals(template, "1.1.0", "release/1.1.0")

	// Check develop branch state - should be merged but version stays as it was
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
	template := filepath.Join("..", "helper", "templates", "version.txt.tpl")

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

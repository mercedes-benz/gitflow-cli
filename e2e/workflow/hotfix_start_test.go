/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"github.com/mercedes-benz/gitflow-cli/e2e/helper"
	"path/filepath"
	"testing"
)

// test Hotfix Start Job
func TestHotfixStart(t *testing.T) {
	// Test with version.txt template
	t.Run("StandardPlugin", func(t *testing.T) {
		testHotfixStart(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("MvnPlugin", func(t *testing.T) {
		testHotfixStart(t, "pom.xml.tpl", "SNAPSHOT")
	})

	// Test with package.json template
	t.Run("NpmPlugin", func(t *testing.T) {
		testHotfixStart(t, "package.json.tpl", "dev")
	})

	// Test with road.yaml template
	t.Run("RoadPlugin", func(t *testing.T) {
		testHotfixStart(t, "road.yaml.tpl", "dev")
	})

	// Test fallback without versioning file
	t.Run("NoPluginFallback", func(t *testing.T) {
		testHotfixStartFallback(t)
	})
}

// testHotfixStart runs the test with the specified template
func testHotfixStart(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	template := filepath.Join("..", "helper", "templates", templateName)

	// main -> version file (1.0.0)
	// develop -> version file (1.1.0-"+versionQualifier)

	env.CommitFileFromTemplate(template, "1.0.0", "main")
	env.CommitFileFromTemplate(template, "1.1.0-"+versionQualifier, "develop")

	// WHEN: The command "gitflow-cli hotfix start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// check hotfix branch state
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertVersionEquals(template, "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Increment patch version for hotfix.", "hotfix/1.0.1")

	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

// TestHotfixStartFallback (test standard plugin with additional functionality)
func testHotfixStartFallback(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Path to the templates
	template := filepath.Join("..", "helper", "templates", "version.txt.tpl")

	// main -> no version file
	// develop -> no version file

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// standard plugin creates version file in main
	env.AssertVersionEquals(template, "1.0.0", "main")
	env.AssertCommitMessageEquals("Create versions file", "main")

	// check hotfix branch state
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertVersionEquals(template, "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Increment patch version for hotfix.", "hotfix/1.0.1")

	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

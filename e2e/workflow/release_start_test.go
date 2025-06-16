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

// Test release start job
func TestReleaseStart(t *testing.T) {
	// Test with version.txt template
	t.Run("StandardPlugin", func(t *testing.T) {
		testReleaseStart(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("MvnPlugin", func(t *testing.T) {
		testReleaseStart(t, "pom.xml.tpl", "SNAPSHOT")
	})

	// Test with package.json template
	t.Run("NpmPlugin", func(t *testing.T) {
		testReleaseStart(t, "package.json.tpl", "dev")
	})

	// Test with road.yaml template
	t.Run("RoadPlugin", func(t *testing.T) {
		testReleaseStart(t, "road.yaml.tpl", "dev")
	})

	// Test fallback without versioning file
	t.Run("NoPluginFallback", func(t *testing.T) {
		testReleaseStartFallback(t)
	})
}

// testReleaseStart runs the test with the specified template
func testReleaseStart(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	template := filepath.Join("..", "helper", "templates", templateName)

	// main -> version file (1.0.0)
	// develop -> version file (1.1.0-{qualifier})
	env.CommitFileFromTemplate(template, "1.0.0", "main")
	env.CommitFileFromTemplate(template, "1.1.0-"+versionQualifier, "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "start")

	// THEN:
	// check release branch state
	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchExists("origin/release/1.1.0")

	env.AssertVersionEquals(template, "1.1.0", "release/1.1.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.1.0")

	env.AssertCurrentBranchEquals("release/1.1.0")
}

// TestReleaseStartFallback (test standard plugin with additional functionality)
func testReleaseStartFallback(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Path to the templates
	template := filepath.Join("..", "helper", "templates", "version.txt.tpl")

	// main -> no version file
	// develop -> no version file

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "start")

	// THEN:
	// standard plugin creates version file in develop
	env.AssertVersionEquals(template, "1.0.0-dev", "develop")
	env.AssertCommitMessageEquals("Create versions file", "develop")

	// check release branch state
	env.AssertBranchExists("release/1.0.0")
	env.AssertBranchExists("origin/release/1.0.0")

	env.AssertVersionEquals(template, "1.0.0", "release/1.0.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.0.0")

	env.AssertCurrentBranchEquals("release/1.0.0")
}

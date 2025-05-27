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

// TestReleaseStart tests Release Start with different templates
func TestReleaseStart(t *testing.T) {
	// Test with version.txt template
	t.Run("Test Standard Plugin", func(t *testing.T) {
		testReleaseStart(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("Test Maven Plugin", func(t *testing.T) {
		testReleaseStart(t, "pom.xml.tpl", "SNAPSHOT")
	})
}

// testReleaseStart runs the test with the specified template
func testReleaseStart(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	templatePath := filepath.Join("../..", "helper", "templates", templateName)

	// main -> template file (1.0.0)
	// develop -> template file (1.1.0-{qualifier})
	env.CommitFileFromTemplate(templatePath, "1.0.0", "main")
	env.CommitFileFromTemplate(templatePath, "1.1.0-"+versionQualifier, "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("release", "start")

	// THEN:
	// check release branch state
	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchExists("origin/release/1.1.0")

	env.AssertVersionEquals(templatePath, "1.1.0", "release/1.1.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.1.0")

	env.AssertCurrentBranchEquals("release/1.1.0")
}

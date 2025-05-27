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
	t.Run("TestStandardPlugin", func(t *testing.T) {
		testHotfixStart(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("TestMavenPlugin", func(t *testing.T) {
		testHotfixStart(t, "pom.xml.tpl", "SNAPSHOT")
	})
}

// testHotfixStart runs the test with the specified template
func testHotfixStart(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	template := filepath.Join("../..", "helper", "templates", templateName)

	// main -> version file (1.0.0)
	// develop -> version file (1.1.0-{qualifier})

	env.CommitFileFromTemplate(template, "1.0.0", "main")
	env.CommitFileFromTemplate(template, "1.1.0-"+versionQualifier, "develop")

	// WHEN: The command "gitflow-cli hotfix start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// check hotfix branch state
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertVersionEquals(template, "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Set next hotfix version.", "hotfix/1.0.1")

	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

// TestHotfixStartWithoutVersionFile (test fallback to standard plugin with additional functionality)
//func TestHotfixStartWithoutVersionFile(t *testing.T) {
//	// GIVEN: a Git repository with production and development branch
//	env := helper.SetupTestEnv(t)
//
//	// Path to the templates
//	template := filepath.Join("../..", "helper", "templates", "version.txt.tpl")
//
//	// main -> no version file
//	// develop -> no version file
//
//	// WHEN: The command "gitflow-cli release start" is executed
//	env.ExecuteGitflow("hotfix", "start")
//
//	// THEN:
//	// standard plugin creates version file in main
//	env.AssertVersionEquals(template, "1.0.0", "main")
//	env.AssertCommitMessageEquals("Create versions file", "main")
//
//	// check hotfix branch state
//	env.AssertBranchExists("hotfix/1.0.1")
//	env.AssertBranchExists("origin/hotfix/1.0.1")
//
//	env.AssertVersionEquals(template, "1.0.1", "hotfix/1.0.1")
//	env.AssertCommitMessageEquals("Set next hotfix version.", "hotfix/1.0.1")
//
//	env.AssertCurrentBranchEquals("hotfix/1.0.1")
//}

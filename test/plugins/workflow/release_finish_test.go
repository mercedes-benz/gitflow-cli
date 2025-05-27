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

// TestReleaseFinish tests Release Finish with different templates
func TestReleaseFinish(t *testing.T) {
	// Test with version.txt template
	t.Run("TestStandardPlugin", func(t *testing.T) {
		testReleaseFinish(t, "version.txt.tpl", "dev")
	})

	// Test with pom.xml template
	t.Run("TestMavenPlugin", func(t *testing.T) {
		testReleaseFinish(t, "pom.xml.tpl", "SNAPSHOT")
	})
}

// testReleaseFinish runs the test with the specified template
func testReleaseFinish(t *testing.T, templateName string, versionQualifier string) {
	// GIVEN: a Git repository with production and development branch
	env := helper.SetupTestEnv(t)

	// Create template path from template name
	template := filepath.Join("../..", "helper", "templates", templateName)

	// main -> version file (1.0.0)
	// develop -> version file (1.1.0-{qualifier})
	// release/1.1.0 -> version file (1.1.0)

	env.CommitFileFromTemplate(template, "1.0.0", "main")
	env.CommitFileFromTemplate(template, "1.1.0-"+versionQualifier, "develop")
	env.CreateBranch("release/1.1.0", "develop")
	env.CommitFileFromTemplate(template, "1.1.0", "release/1.1.0")

	// WHEN: The command "gitflow-cli release finish" is executed
	env.ExecuteGitflow("release", "finish")

	// THEN
	// Check main branch state
	env.AssertCommitMessageEquals("Merge branch 'release/1.1.0'", "main")
	env.AssertTagEquals("1.1.0", "main")
	env.AssertVersionEquals(template, "1.1.0", "main")

	// Check develop branch state
	env.AssertCommitMessageEquals("Merge branch 'release/1.1.0' into develop", "develop", 1)
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
	env.AssertVersionEquals(template, "1.2.0-"+versionQualifier, "develop")

	env.AssertBranchDoesNotExist("release/1.1.0")
	env.AssertCurrentBranchEquals("develop")
}

// TestReleaseFinishWithoutVersionFileInMain (test fallback to standard plugin with additional functionality)
//func TestReleaseFinishWithoutVersionFileInMain(t *testing.T) {
//	// GIVEN: a Git repository with production and development branch
//	env := helper.SetupTestEnv(t)
//
//	// Path to the version file template
//	template := filepath.Join("../..", "helper", "templates", "version.txt.tpl")
//
//	// main -> no version file
//	// develop -> version.txt (1.0.0-dev)
//	// release/1.0.0 -> version.txt (1.0.0)
//
//	env.CommitFileFromTemplate(template, "1.0.0-dev", "develop")
//	env.CreateBranch("release/1.0.0", "develop")
//	env.CommitFileFromTemplate(template, "1.0.0", "release/1.0.0")
//
//	// WHEN
//	env.ExecuteGitflow("release", "finish")
//
//	// THEN
//	// Check main branch state
//	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0'", "main")
//	env.AssertTagEquals("1.0.0", "main")
//	env.AssertVersionEquals(template, "1.0.0", "main")
//
//	// Check develop branch state
//	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0' into develop", "develop", 1)
//	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
//	env.AssertVersionEquals(template, "1.1.0-dev", "develop")
//
//	env.AssertBranchDoesNotExist("release/1.0.0")
//	env.AssertCurrentBranchEquals("develop")
//}

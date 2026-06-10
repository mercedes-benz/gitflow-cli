/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/e2e/helper"
	"github.com/mercedes-benz/gitflow-cli/plugin/standard"
)

// Constants for custom branch names used in all tests
const (
	productionBranch  = "custom-production"
	developmentBranch = "custom-develop"
	releaseBranch     = "custom-release"
	hotfixBranch      = "custom-hotfix"
)

var tc = standard.E2ETestConfig

// setupCustomBranchTest creates a test environment with custom branch names
func setupCustomBranchTest(t *testing.T) (*helper.GitTestEnv, string) {
	// GIVEN: a Git repository with custom branch names
	env := helper.SetupTestEnv(t,
		helper.WithProductionBranch(productionBranch),
		helper.WithDevelopmentBranch(developmentBranch),
		helper.WithReleaseBranch(releaseBranch),
		helper.WithHotfixBranch(hotfixBranch),
	)

	// Path to the predefined config file
	configPath := filepath.Join("helper", ".gitflow-test-config.yaml")

	return env, configPath
}

// TestReleaseStartWithConfigFile tests the release start workflow with a custom configuration file
func TestReleaseStartWithConfigFile(t *testing.T) {
	env, configPath := setupCustomBranchTest(t)

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", productionBranch)
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-dev", developmentBranch)

	env.ExecuteGitflow("release", "start", "--config", configPath)

	customReleaseBranch := releaseBranch + "/1.1.0"
	env.AssertBranchExists(customReleaseBranch)
	env.AssertBranchExists("origin/" + customReleaseBranch)
	env.AssertCurrentBranchEquals(customReleaseBranch)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.1.0", customReleaseBranch)
	env.AssertCommitMessageEquals("Remove qualifier from project version.", customReleaseBranch)
}

// TestReleaseFinishWithConfigFile tests the release finish workflow with a custom configuration file
func TestReleaseFinishWithConfigFile(t *testing.T) {
	env, configPath := setupCustomBranchTest(t)

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", productionBranch)
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-dev", developmentBranch)

	customReleaseBranch := releaseBranch + "/1.1.0"
	env.CreateBranch(customReleaseBranch, developmentBranch)
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0", customReleaseBranch)

	env.ExecuteGitflow("release", "finish", "--config", configPath)

	env.AssertCommitMessageEquals("Merge branch '"+customReleaseBranch+"' into "+productionBranch, productionBranch)
	env.AssertTagEquals("1.1.0", productionBranch)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.1.0", productionBranch)

	env.AssertCommitMessageEquals("Merge branch '"+customReleaseBranch+"' into "+developmentBranch, developmentBranch, 1)
	env.AssertCommitMessageEquals("Set next minor project version.", developmentBranch, 0)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.2.0-dev", developmentBranch)

	env.AssertBranchDoesNotExist(customReleaseBranch)
	env.AssertCurrentBranchEquals(developmentBranch)
}

// TestHotfixStartWithConfigFile tests the hotfix start workflow with a custom configuration file
func TestHotfixStartWithConfigFile(t *testing.T) {
	env, configPath := setupCustomBranchTest(t)

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", productionBranch)
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-dev", developmentBranch)

	env.ExecuteGitflow("hotfix", "start", "--config", configPath)

	customHotfixBranch := hotfixBranch + "/1.0.1"
	env.AssertBranchExists(customHotfixBranch)
	env.AssertBranchExists("origin/" + customHotfixBranch)
	env.AssertCurrentBranchEquals(customHotfixBranch)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.0.1", customHotfixBranch)
	env.AssertCommitMessageEquals("Increment patch version for hotfix.", customHotfixBranch)
}

// TestHotfixFinishWithConfigFile tests the hotfix finish workflow with a custom configuration file
func TestHotfixFinishWithConfigFile(t *testing.T) {
	env, configPath := setupCustomBranchTest(t)

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", productionBranch)
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-dev", developmentBranch)

	customHotfixBranch := hotfixBranch + "/1.0.1"
	env.CreateBranch(customHotfixBranch, productionBranch)
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.1", customHotfixBranch)

	env.ExecuteGitflow("hotfix", "finish", "--config", configPath)

	env.AssertCommitMessageEquals("Merge branch '"+customHotfixBranch+"' into "+productionBranch, productionBranch)
	env.AssertTagEquals("1.0.1", productionBranch)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.0.1", productionBranch)

	env.AssertCommitMessageEquals("Merge branch '"+customHotfixBranch+"' into "+developmentBranch, developmentBranch)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.1.0-dev", developmentBranch)

	env.AssertBranchDoesNotExist(customHotfixBranch)
	env.AssertCurrentBranchEquals(developmentBranch)
}

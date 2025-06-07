/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package e2e

import (
	"github.com/mercedes-benz/gitflow-cli/e2e/helper"
	"path/filepath"
	"testing"
)

// Constants for custom branch names used in all tests
const (
	productionBranch  = "custom-production"
	developmentBranch = "custom-develop"
	releaseBranch     = "custom-release"
	hotfixBranch      = "custom-hotfix"
)

// setupCustomBranchTest creates a test environment with custom branch names
// and returns the paths to configuration files
func setupCustomBranchTest(t *testing.T) (*helper.GitTestEnv, string, string) {
	// GIVEN: a Git repository with custom branch names
	env := helper.SetupTestEnv(t,
		helper.WithProductionBranch(productionBranch),
		helper.WithDevelopmentBranch(developmentBranch),
		helper.WithReleaseBranch(releaseBranch),
		helper.WithHotfixBranch(hotfixBranch),
	)

	// Path to the predefined config file
	configPath := filepath.Join("helper", ".gitflow-test-config.yaml")

	// Create the version file template path
	versionTemplate := filepath.Join("helper", "templates", "version.txt.tpl")

	return env, configPath, versionTemplate
}

// TestReleaseStartWithConfigFile tests the release start workflow with a custom configuration file
func TestReleaseStartWithConfigFile(t *testing.T) {
	// Set up the test environment with custom branch names
	env, configPath, versionTemplate := setupCustomBranchTest(t)

	// custom-production -> version file (1.0.0)
	// custom-develop -> version file (1.1.0-dev)
	env.CommitFileFromTemplate(versionTemplate, "1.0.0", productionBranch)
	env.CommitFileFromTemplate(versionTemplate, "1.1.0-dev", developmentBranch)

	// WHEN: Execute release start with the config file
	env.ExecuteGitflow("release", "start", "--config", configPath)

	// THEN:
	customReleaseBranch := releaseBranch + "/1.1.0"
	env.AssertBranchExists(customReleaseBranch)
	env.AssertBranchExists("origin/" + customReleaseBranch)

	env.AssertCurrentBranchEquals(customReleaseBranch)

	env.AssertVersionEquals(versionTemplate, "1.1.0", customReleaseBranch)
	env.AssertCommitMessageEquals("Remove qualifier from project version.", customReleaseBranch)
}

// TestReleaseFinishWithConfigFile tests the release finish workflow with a custom configuration file
func TestReleaseFinishWithConfigFile(t *testing.T) {
	// Set up the test environment with custom branch names
	env, configPath, versionTemplate := setupCustomBranchTest(t)

	// custom-production -> version file (1.0.0)
	// custom-develop -> version file (1.1.0-dev)
	// custom-release/1.1.0 -> version file (1.1.0)
	env.CommitFileFromTemplate(versionTemplate, "1.0.0", productionBranch)
	env.CommitFileFromTemplate(versionTemplate, "1.1.0-dev", developmentBranch)

	// Create a release branch and set its version
	customReleaseBranch := releaseBranch + "/1.1.0"
	env.CreateBranch(customReleaseBranch, developmentBranch)
	env.CommitFileFromTemplate(versionTemplate, "1.1.0", customReleaseBranch)

	// WHEN: Execute release finish with the config file
	env.ExecuteGitflow("release", "finish", "--config", configPath)

	// THEN:
	// Check production branch state
	env.AssertCommitMessageEquals("Merge branch '"+customReleaseBranch+"' into "+productionBranch+"", productionBranch)
	env.AssertTagEquals("1.1.0", productionBranch)
	env.AssertVersionEquals(versionTemplate, "1.1.0", productionBranch)

	// Check development branch state
	env.AssertCommitMessageEquals("Merge branch '"+customReleaseBranch+"' into "+developmentBranch, developmentBranch, 1)
	env.AssertCommitMessageEquals("Set next minor project version.", developmentBranch, 0)
	env.AssertVersionEquals(versionTemplate, "1.2.0-dev", developmentBranch)

	// Release branch should be deleted
	env.AssertBranchDoesNotExist(customReleaseBranch)

	// Current branch should be development
	env.AssertCurrentBranchEquals(developmentBranch)
}

// TestHotfixStartWithConfigFile tests the hotfix start workflow with a custom configuration file
func TestHotfixStartWithConfigFile(t *testing.T) {
	// Setup the test environment with custom branch names
	env, configPath, versionTemplate := setupCustomBranchTest(t)

	// custom-production -> version file (1.0.0)
	// custom-develop -> version file (1.1.0-dev)
	env.CommitFileFromTemplate(versionTemplate, "1.0.0", productionBranch)
	env.CommitFileFromTemplate(versionTemplate, "1.1.0-dev", developmentBranch)

	// WHEN: Execute hotfix start with the config file
	env.ExecuteGitflow("hotfix", "start", "--config", configPath)

	// THEN:
	customHotfixBranch := hotfixBranch + "/1.0.1"
	env.AssertBranchExists(customHotfixBranch)
	env.AssertBranchExists("origin/" + customHotfixBranch)

	// We should be on the hotfix branch
	env.AssertCurrentBranchEquals(customHotfixBranch)

	// The version in hotfix branch should be incremented patch version without qualifier
	env.AssertVersionEquals(versionTemplate, "1.0.1", customHotfixBranch)
	env.AssertCommitMessageEquals("Increment patch version for hotfix.", customHotfixBranch)
}

// TestHotfixFinishWithConfigFile tests the hotfix finish workflow with a custom configuration file
func TestHotfixFinishWithConfigFile(t *testing.T) {
	// Setup the test environment with custom branch names
	env, configPath, versionTemplate := setupCustomBranchTest(t)

	// custom-production -> version file (1.0.0)
	// custom-develop -> version file (1.1.0-dev)
	// custom-hotfix/1.0.1 -> version file (1.0.1)
	env.CommitFileFromTemplate(versionTemplate, "1.0.0", productionBranch)
	env.CommitFileFromTemplate(versionTemplate, "1.1.0-dev", developmentBranch)

	// Create a hotfix branch and increment the patch version
	customHotfixBranch := hotfixBranch + "/1.0.1"
	env.CreateBranch(customHotfixBranch, productionBranch)
	env.CommitFileFromTemplate(versionTemplate, "1.0.1", customHotfixBranch)

	// WHEN: Execute hotfix finish with the config file
	env.ExecuteGitflow("hotfix", "finish", "--config", configPath)

	// THEN:
	// Check production branch state
	env.AssertCommitMessageEquals("Merge branch '"+customHotfixBranch+"' into "+productionBranch+"", productionBranch)
	env.AssertTagEquals("1.0.1", productionBranch)
	env.AssertVersionEquals(versionTemplate, "1.0.1", productionBranch)

	// Check development branch state
	env.AssertCommitMessageEquals("Merge branch '"+customHotfixBranch+"' into "+developmentBranch, developmentBranch)
	env.AssertVersionEquals(versionTemplate, "1.1.0-dev", developmentBranch)

	// Hotfix branch should be deleted
	env.AssertBranchDoesNotExist(customHotfixBranch)

	// Current branch should be development
	env.AssertCurrentBranchEquals(developmentBranch)
}

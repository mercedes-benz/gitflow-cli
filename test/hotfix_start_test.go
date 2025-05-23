/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package test

import (
	"github.com/mercedes-benz/gitflow-cli/test/base"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestHotfixStart checks if the command "hotfix start" performs the correct Git graph manipulation
func TestHotfixStart(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := base.SetupTestEnv(t)

	env.CommitFile("version.txt", "1.0.0", "Merge branch 'release/1.0.0", "main")
	env.CommitFile("version.txt", "1.1.0-dev", "Set next minor project version.", "develop")

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// check hotfix branch state
	hotfixBranch := "hotfix/1.0.1"

	env.AssertBranchExists(hotfixBranch)
	env.AssertBranchExists("origin/" + hotfixBranch)

	env.AssertFileEquals("version.txt", "1.0.1", hotfixBranch)
	env.AssertCommitMessageEquals("Set next hotfix version.", hotfixBranch)

	assert.Equal(t, hotfixBranch, env.GetCurrentBranch(), "Current branch should be the hotfix branch")
}

/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package test

import (
	"github.com/mercedes-benz/gitflow-cli/test/base"
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
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertFileEquals("version.txt", "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Set next hotfix version.", "hotfix/1.0.1")

	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

func TestHotfixStartWithoutVersionFile(t *testing.T) {
	// GIVEN: a Git repository with production and development branch
	env := base.SetupTestEnv(t)

	// WHEN: The command "gitflow-cli release start" is executed
	env.ExecuteGitflow("hotfix", "start")

	// THEN:
	// check hotfix branch state
	env.AssertFileEquals("version.txt", "1.0.0", "main")
	env.AssertCommitMessageEquals("Create versions file", "main")

	// check hotfix branch state
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")

	env.AssertFileEquals("version.txt", "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Set next hotfix version.", "hotfix/1.0.1")
	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

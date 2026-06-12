/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/e2e"
)

func RunHotfixStart(t *testing.T, tc plugin.TestConfig) {
	t.Helper()
	env := e2e.SetupTestEnv(t, e2e.WithDockerMode(tc.DockerImage != ""))

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", "main")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-"+tc.VersionQualifier, "develop")

	env.ExecuteGitflow("hotfix", "start")

	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Increment patch version for hotfix.", "hotfix/1.0.1")
	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

func RunHotfixStartFallback(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.ExecuteGitflow("hotfix", "start")

	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.0", "main")
	env.AssertCommitMessageEquals("Create versions file", "main")
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")
	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Increment patch version for hotfix.", "hotfix/1.0.1")
	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

func RunBeforeHotfixStartHook(t *testing.T, tc plugin.TestConfig) {
	t.Helper()
	env := e2e.SetupTestEnv(t, e2e.WithDockerMode(tc.DockerImage != ""))

	env.CommitFile(tc.VersionFileName, tc.EmptyContent, "main")

	env.ExecuteGitflow("hotfix", "start")

	env.AssertCommitMessageEquals("Set initial project version.", "main")
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")
}

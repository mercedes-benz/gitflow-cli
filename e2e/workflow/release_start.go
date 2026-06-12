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

func RunReleaseStart(t *testing.T, tc plugin.TestConfig) {
	t.Helper()
	env := e2e.SetupTestEnv(t, e2e.WithDockerMode(tc.DockerImage != ""))

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", "main")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-"+tc.VersionQualifier, "develop")

	env.ExecuteGitflow("release", "start")

	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchExists("origin/release/1.1.0")
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.1.0", "release/1.1.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.1.0")
	env.AssertCurrentBranchEquals("release/1.1.0")
}

func RunReleaseStartFallback(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.ExecuteGitflow("release", "start")

	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.0-dev", "develop")
	env.AssertCommitMessageEquals("Create versions file", "develop")
	env.AssertBranchExists("release/1.0.0")
	env.AssertBranchExists("origin/release/1.0.0")
	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.0", "release/1.0.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.0.0")
	env.AssertCurrentBranchEquals("release/1.0.0")
}

func RunBeforeReleaseStartHook(t *testing.T, tc plugin.TestConfig, emptyContent []byte) {
	t.Helper()
	env := e2e.SetupTestEnv(t, e2e.WithDockerMode(tc.DockerImage != ""))

	env.CommitFile(tc.VersionFileName, emptyContent, "develop")

	env.ExecuteGitflow("release", "start")

	env.AssertCommitMessageEquals("Set initial project version.", "develop")
	env.AssertBranchExists("release/1.0.0")
	env.AssertBranchExists("origin/release/1.0.0")
}

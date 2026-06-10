/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/e2e/helper"
)

func TestReleaseStart(t *testing.T) {
	for _, tc := range pluginTestConfigs {
		t.Run(tc.Name+"Plugin", func(t *testing.T) {
			testReleaseStart(t, tc)
		})

		if tc.HasBeforeStartHook && tc.EmptyFileContent != nil {
			t.Run(tc.Name+"Plugin_BeforeReleaseStartHook", func(t *testing.T) {
				testBeforeReleaseStartHook(t, tc)
			})
		}
	}

	t.Run("NoPluginFallback", func(t *testing.T) {
		testReleaseStartFallback(t)
	})
}

func testReleaseStart(t *testing.T, tc plugin.TestConfig) {
	env := helper.SetupTestEnv(t)
	helper.SetupPluginContainer(t, tc, env.LocalPath)

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", "main")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-"+tc.VersionQualifier, "develop")

	env.ExecuteGitflow("release", "start")

	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchExists("origin/release/1.1.0")
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.1.0", "release/1.1.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.1.0")
	env.AssertCurrentBranchEquals("release/1.1.0")
}

func testReleaseStartFallback(t *testing.T) {
	env := helper.SetupTestEnv(t)

	env.ExecuteGitflow("release", "start")

	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.0-dev", "develop")
	env.AssertCommitMessageEquals("Create versions file", "develop")
	env.AssertBranchExists("release/1.0.0")
	env.AssertBranchExists("origin/release/1.0.0")
	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.0", "release/1.0.0")
	env.AssertCommitMessageEquals("Remove qualifier from project version.", "release/1.0.0")
	env.AssertCurrentBranchEquals("release/1.0.0")
}

func testBeforeReleaseStartHook(t *testing.T, tc plugin.TestConfig) {
	env := helper.SetupTestEnv(t)
	helper.SetupPluginContainer(t, tc, env.LocalPath)

	env.CommitFile(tc.VersionFileName, tc.EmptyFileContent, "develop")

	env.ExecuteGitflow("release", "start")

	env.AssertCommitMessageEquals("Set initial project version.", "develop")
	env.AssertBranchExists("release/1.0.0")
	env.AssertBranchExists("origin/release/1.0.0")
}

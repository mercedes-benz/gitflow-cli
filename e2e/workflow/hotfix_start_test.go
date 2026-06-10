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

func TestHotfixStart(t *testing.T) {
	for _, tc := range pluginTestConfigs {
		t.Run(tc.Name+"Plugin", func(t *testing.T) {
			testHotfixStart(t, tc)
		})

		if tc.HasBeforeStartHook && tc.EmptyFileContent != nil {
			t.Run(tc.Name+"Plugin_BeforeHotfixStartHook", func(t *testing.T) {
				testBeforeHotfixStartHook(t, tc)
			})
		}
	}

	t.Run("NoPluginFallback", func(t *testing.T) {
		testHotfixStartFallback(t)
	})
}

func testHotfixStart(t *testing.T, tc plugin.TestConfig) {
	env := helper.SetupTestEnv(t)
	helper.SetupPluginContainer(t, tc, env.LocalPath)

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", "main")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-"+tc.VersionQualifier, "develop")

	env.ExecuteGitflow("hotfix", "start")

	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Increment patch version for hotfix.", "hotfix/1.0.1")
	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

func testHotfixStartFallback(t *testing.T) {
	env := helper.SetupTestEnv(t)

	env.ExecuteGitflow("hotfix", "start")

	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.0", "main")
	env.AssertCommitMessageEquals("Create versions file", "main")
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")
	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.1", "hotfix/1.0.1")
	env.AssertCommitMessageEquals("Increment patch version for hotfix.", "hotfix/1.0.1")
	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

func testBeforeHotfixStartHook(t *testing.T, tc plugin.TestConfig) {
	env := helper.SetupTestEnv(t)
	helper.SetupPluginContainer(t, tc, env.LocalPath)

	env.CommitFile(tc.VersionFileName, tc.EmptyFileContent, "main")

	env.ExecuteGitflow("hotfix", "start")

	env.AssertCommitMessageEquals("Set initial project version.", "main")
	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchExists("origin/hotfix/1.0.1")
}

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

func TestHotfixFinish(t *testing.T) {
	for _, tc := range pluginTestConfigs {
		t.Run(tc.Name+"Plugin", func(t *testing.T) {
			testHotfixFinish(t, tc)
		})
	}

	t.Run("NoPluginFallback", func(t *testing.T) {
		testHotfixFinishFallback(t)
	})
}

func testHotfixFinish(t *testing.T, tc plugin.TestConfig) {
	env := helper.SetupTestEnv(t)
	helper.SetupPluginContainer(t, tc, env.LocalPath)

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", "main")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-"+tc.VersionQualifier, "develop")

	env.CreateBranch("release/1.1.0", "develop")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0", "release/1.1.0")

	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.1", "hotfix/1.0.1")

	env.ExecuteGitflow("hotfix", "finish")

	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.0.1", "main")

	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into release/1.1.0", "release/1.1.0", 0)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.1.0", "release/1.1.0")

	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into develop", "develop", 0)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.1.0-"+tc.VersionQualifier, "develop")

	env.AssertBranchDoesNotExist("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

func testHotfixFinishFallback(t *testing.T) {
	env := helper.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.1", "hotfix/1.0.1")

	env.ExecuteGitflow("hotfix", "finish")

	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1'", "main")
	env.AssertTagEquals("1.0.1", "main")
	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.1", "main")

	env.AssertCommitMessageEquals("Merge branch 'hotfix/1.0.1' into develop", "develop", 1)
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	env.AssertBranchDoesNotExist("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

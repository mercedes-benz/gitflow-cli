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

func RunHotfixFinish(t *testing.T, tc plugin.TestConfig) {
	t.Helper()
	env := e2e.SetupTestEnv(t, e2e.WithDockerMode(tc.DockerImage != ""))

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

func RunHotfixFinishFallback(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

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

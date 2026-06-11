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

func RunReleaseFinish(t *testing.T, tc plugin.TestConfig) {
	t.Helper()
	env := helper.SetupTestEnv(t, helper.WithDockerMode(tc.DockerImage != ""))

	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.0.0", "main")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0-"+tc.VersionQualifier, "develop")
	env.CreateBranch("release/1.1.0", "develop")
	env.CommitTemplateContent(tc.Template, tc.VersionFileName, "1.1.0", "release/1.1.0")

	env.ExecuteGitflow("release", "finish")

	env.AssertCommitMessageEquals("Merge branch 'release/1.1.0'", "main")
	env.AssertTagEquals("1.1.0", "main")
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.1.0", "main")

	env.AssertCommitMessageEquals("Merge branch 'release/1.1.0' into develop", "develop", 1)
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
	env.AssertTemplateVersionEquals(tc.Template, tc.VersionFileName, "1.2.0-"+tc.VersionQualifier, "develop")

	env.AssertBranchDoesNotExist("release/1.1.0")
	env.AssertCurrentBranchEquals("develop")
}

func RunReleaseFinishFallback(t *testing.T) {
	t.Helper()
	env := helper.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0-dev", "develop")
	env.CreateBranch("release/1.0.0", "develop")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "release/1.0.0")

	env.ExecuteGitflow("release", "finish")

	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0'", "main")
	env.AssertTagEquals("1.0.0", "main")
	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.0.0", "main")

	env.AssertCommitMessageEquals("Merge branch 'release/1.0.0' into develop", "develop", 1)
	env.AssertCommitMessageEquals("Set next minor project version.", "develop", 0)
	env.AssertTemplateVersionEquals("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	env.AssertBranchDoesNotExist("release/1.0.0")
	env.AssertCurrentBranchEquals("develop")
}

/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"os"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/e2e"
	"github.com/stretchr/testify/assert"
)

// --- Push disabled tests ---

func RunReleaseStartNoPush(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	configPath := env.WriteConfig("workflow:\n  push: false\n")
	env.ExecuteGitflow("release", "start", "--config", configPath)

	env.AssertBranchExists("release/1.1.0")
	env.AssertBranchNotOnRemote("release/1.1.0")
	env.AssertCurrentBranchEquals("release/1.1.0")
}

func RunReleaseFinishNoPush(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("release/1.1.0", "develop")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0", "release/1.1.0")

	configPath := env.WriteConfig("workflow:\n  push: false\n")
	env.ExecuteGitflow("release", "finish", "--config", configPath)

	env.AssertTagNotOnRemote("1.1.0")
	env.AssertCurrentBranchEquals("develop")
}

func RunHotfixStartNoPush(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	configPath := env.WriteConfig("workflow:\n  push: false\n")
	env.ExecuteGitflow("hotfix", "start", "--config", configPath)

	env.AssertBranchExists("hotfix/1.0.1")
	env.AssertBranchNotOnRemote("hotfix/1.0.1")
	env.AssertCurrentBranchEquals("hotfix/1.0.1")
}

func RunHotfixFinishNoPush(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("hotfix/1.0.1", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.1", "hotfix/1.0.1")

	configPath := env.WriteConfig("workflow:\n  push: false\n")
	env.ExecuteGitflow("hotfix", "finish", "--config", configPath)

	env.AssertTagNotOnRemote("1.0.1")
	env.AssertCurrentBranchEquals("develop")
}

// --- Rollback tests ---

func RunRollbackPreservesExistingBranches(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// Try release finish without a release branch — triggers an error
	configPath := env.WriteConfig("workflow:\n  rollback: true\n")
	errMsg := env.ExecuteGitflowExpectError("release", "finish", "--config", configPath)

	assert.Contains(t, errMsg, "'release'")

	// develop branch must still exist after rollback
	env.AssertBranchExists("develop")
}

func RunRollbackDisabledLeavesState(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	configPath := env.WriteConfig("workflow:\n  rollback: false\n")
	errMsg := env.ExecuteGitflowExpectError("release", "finish", "--config", configPath)

	assert.Contains(t, errMsg, "'release'")

	// develop branch must still exist (rollback disabled = no cleanup at all)
	env.AssertBranchExists("develop")
}

// --- Branch sync tests ---

func RunReleaseStartCreatesDevBranch(t *testing.T) {
	t.Helper()

	env := e2e.SetupTestEnvWithoutDevelop(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")

	// Set up sync that creates develop and sets a qualified version
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Development {
			// After syncBranch creates the branch, we need a qualified version on develop.
			// We handle this by creating the branch ourselves with the right version.
			if err := req.Repository.CheckoutBranch(req.CreateFrom); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.CreateBranch(req.Configured); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.WriteFile("version.txt", "1.1.0-dev"); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.AddFile("version.txt"); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.CommitChanges("Set development version"); err != nil {
				return core.BranchSyncResult{}, err
			}
			if err := req.Repository.PushChanges(req.Configured); err != nil {
				return core.BranchSyncResult{}, err
			}
			// Return Created: false because we already created it ourselves
			return core.BranchSyncResult{ResolvedName: req.Configured}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	env.ExecuteGitflow("release", "start")

	env.AssertBranchExists("develop")
	env.AssertBranchExists("release/1.1.0")
}

func RunReleaseStartDeclinedCreatesDev(t *testing.T) {
	t.Helper()

	env := e2e.SetupTestEnvWithoutDevelop(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")

	// Set up declining sync
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Development {
			return core.BranchSyncResult{}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	errMsg := env.ExecuteGitflowExpectError("release", "start")

	assert.Contains(t, errMsg, "required but was not resolved")
}

// --- Production branch sync tests ---

func RunReleaseStartWithMasterBranch(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { core.ResetBranchNames() })

	// Repo has 'master' as production branch, but config says 'main' (default)
	env := e2e.SetupTestEnv(t, e2e.WithProductionBranch("master"))

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "master")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// Sync callback: resolve 'main' → 'master' (found as candidate)
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Production && req.Configured == "main" {
			return core.BranchSyncResult{ResolvedName: "master"}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	env.ExecuteGitflow("release", "start")

	env.AssertBranchExists("release/1.1.0")
}

func RunReleaseStartWithMasterDeclined(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { core.ResetBranchNames() })

	env := e2e.SetupTestEnv(t, e2e.WithProductionBranch("master"))

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "master")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// Sync callback: decline resolution
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Production {
			return core.BranchSyncResult{}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	errMsg := env.ExecuteGitflowExpectError("release", "start")

	assert.Contains(t, errMsg, "required but was not resolved")
}

func RunReleaseStartWithDevBranch(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { core.ResetBranchNames() })

	// Repo has 'dev' instead of 'develop'
	env := e2e.SetupTestEnv(t, e2e.WithDevelopmentBranch("dev"))

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "dev")

	// Sync callback: resolve 'develop' → 'dev'
	oldSync := core.BranchSync
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		if req.BranchType == core.Development && req.Configured == "develop" {
			return core.BranchSyncResult{ResolvedName: "dev"}, nil
		}
		return core.BranchSyncResult{ResolvedName: req.Configured}, nil
	}
	t.Cleanup(func() { core.BranchSync = oldSync })

	env.ExecuteGitflow("release", "start")

	env.AssertBranchExists("release/1.1.0")
}

// --- Robustness tests ---

func RunReleaseStartDirtyRepo(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")

	// Make the repo dirty
	env.ExecuteGit("checkout", "develop")
	dirtyFile := env.LocalPath + "/dirty.txt"
	_ = os.WriteFile(dirtyFile, []byte("uncommitted"), 0644)
	env.ExecuteGit("add", dirtyFile)

	errMsg := env.ExecuteGitflowExpectError("release", "start")

	assert.Contains(t, errMsg, "not clean")
}

func RunReleaseStartDuplicateRelease(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("release/1.1.0", "develop")

	errMsg := env.ExecuteGitflowExpectError("release", "start")

	assert.Contains(t, errMsg, "already has")
}

func RunHotfixStartDuplicateHotfix(t *testing.T) {
	t.Helper()
	env := e2e.SetupTestEnv(t)

	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.0.0", "main")
	env.CommitTemplateContent("{{.Version}}", "version.txt", "1.1.0-dev", "develop")
	env.CreateBranch("hotfix/1.0.1", "main")

	errMsg := env.ExecuteGitflowExpectError("hotfix", "start")

	assert.Contains(t, errMsg, "already has")
}

/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package standard

import (
	_ "embed"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/e2e/workflow"
)

//go:embed testdata/e2e/version.txt.tpl
var versionTemplate string

var testConfig = plugin.TestConfig{
	Name:             "standard",
	DockerImage:      pluginConfig.DockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "version.txt",
	Template:         versionTemplate,
}

func TestReleaseStart(t *testing.T) {
	workflow.RunReleaseStart(t, testConfig)
}

func TestReleaseStartFallback(t *testing.T) {
	workflow.RunReleaseStartFallback(t)
}

func TestReleaseFinish(t *testing.T) {
	workflow.RunReleaseFinish(t, testConfig)
}

func TestReleaseFinishFallback(t *testing.T) {
	workflow.RunReleaseFinishFallback(t)
}

func TestHotfixStart(t *testing.T) {
	workflow.RunHotfixStart(t, testConfig)
}

func TestHotfixStartFallback(t *testing.T) {
	workflow.RunHotfixStartFallback(t)
}

func TestHotfixFinish(t *testing.T) {
	workflow.RunHotfixFinish(t, testConfig)
}

func TestHotfixFinishFallback(t *testing.T) {
	workflow.RunHotfixFinishFallback(t)
}

// --- Edge case tests ---

func TestReleaseStartNoPush(t *testing.T) {
	workflow.RunReleaseStartNoPush(t)
}

func TestReleaseFinishNoPush(t *testing.T) {
	workflow.RunReleaseFinishNoPush(t)
}

func TestHotfixStartNoPush(t *testing.T) {
	workflow.RunHotfixStartNoPush(t)
}

func TestHotfixFinishNoPush(t *testing.T) {
	workflow.RunHotfixFinishNoPush(t)
}

func TestRollbackPreservesExistingBranches(t *testing.T) {
	workflow.RunRollbackPreservesExistingBranches(t)
}

func TestRollbackDisabledLeavesState(t *testing.T) {
	workflow.RunRollbackDisabledLeavesState(t)
}

func TestReleaseStartCreatesDevBranch(t *testing.T) {
	workflow.RunReleaseStartCreatesDevBranch(t)
}

func TestReleaseStartDeclinedCreatesDev(t *testing.T) {
	workflow.RunReleaseStartDeclinedCreatesDev(t)
}

func TestReleaseStartWithMasterBranch(t *testing.T) {
	workflow.RunReleaseStartWithMasterBranch(t)
}

func TestReleaseStartWithMasterDeclined(t *testing.T) {
	workflow.RunReleaseStartWithMasterDeclined(t)
}

func TestReleaseStartWithDevBranch(t *testing.T) {
	workflow.RunReleaseStartWithDevBranch(t)
}

func TestReleaseStartDirtyRepo(t *testing.T) {
	workflow.RunReleaseStartDirtyRepo(t)
}

func TestReleaseStartDuplicateRelease(t *testing.T) {
	workflow.RunReleaseStartDuplicateRelease(t)
}

func TestHotfixStartDuplicateHotfix(t *testing.T) {
	workflow.RunHotfixStartDuplicateHotfix(t)
}

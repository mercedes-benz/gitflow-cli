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

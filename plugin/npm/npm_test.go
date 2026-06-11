/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package npm

import (
	_ "embed"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/e2e/workflow"
)

//go:embed testdata/e2e/package.json.tpl
var packageTemplate string

var e2eConfig = plugin.TestConfig{
	Name:             "npm",
	DockerImage:      pluginConfig.DockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "package.json",
	Template:         packageTemplate,
}

var emptyFileContent = []byte("{}")

func TestReleaseStart(t *testing.T) {
	workflow.RunReleaseStart(t, e2eConfig)
}

func TestReleaseStart_BeforeHook(t *testing.T) {
	workflow.RunBeforeReleaseStartHook(t, e2eConfig, emptyFileContent)
}

func TestReleaseFinish(t *testing.T) {
	workflow.RunReleaseFinish(t, e2eConfig)
}

func TestHotfixStart(t *testing.T) {
	workflow.RunHotfixStart(t, e2eConfig)
}

func TestHotfixStart_BeforeHook(t *testing.T) {
	workflow.RunBeforeHotfixStartHook(t, e2eConfig, emptyFileContent)
}

func TestHotfixFinish(t *testing.T) {
	workflow.RunHotfixFinish(t, e2eConfig)
}

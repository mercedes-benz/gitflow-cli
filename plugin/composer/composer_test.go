/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package composer

import (
	_ "embed"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/e2e/workflow"
)

//go:embed testdata/e2e/composer.json.tpl
var composerTemplate string

var testConfig = plugin.TestConfig{
	Name:             "composer",
	DockerImage:      pluginConfig.DockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "composer.json",
	Template:         composerTemplate,
	EmptyContent:     []byte("{}"),
}

func TestReleaseStart(t *testing.T) {
	workflow.RunReleaseStart(t, testConfig)
}

func TestReleaseStart_BeforeHook(t *testing.T) {
	workflow.RunBeforeReleaseStartHook(t, testConfig)
}

func TestReleaseFinish(t *testing.T) {
	workflow.RunReleaseFinish(t, testConfig)
}

func TestHotfixStart(t *testing.T) {
	workflow.RunHotfixStart(t, testConfig)
}

func TestHotfixStart_BeforeHook(t *testing.T) {
	workflow.RunBeforeHotfixStartHook(t, testConfig)
}

func TestHotfixFinish(t *testing.T) {
	workflow.RunHotfixFinish(t, testConfig)
}

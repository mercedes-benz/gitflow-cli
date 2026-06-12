/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package mvn

import (
	_ "embed"
	"testing"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/e2e/workflow"
)

//go:embed testdata/e2e/pom.xml.tpl
var pomTemplate string

var testConfig = plugin.TestConfig{
	Name:             "mvn",
	DockerImage:      pluginConfig.DockerImage,
	VersionQualifier: "SNAPSHOT",
	VersionFileName:  "pom.xml",
	Template:         pomTemplate,
}

func TestReleaseStart(t *testing.T) {
	workflow.RunReleaseStart(t, testConfig)
}

func TestReleaseFinish(t *testing.T) {
	workflow.RunReleaseFinish(t, testConfig)
}

func TestHotfixStart(t *testing.T) {
	workflow.RunHotfixStart(t, testConfig)
}

func TestHotfixFinish(t *testing.T) {
	workflow.RunHotfixFinish(t, testConfig)
}

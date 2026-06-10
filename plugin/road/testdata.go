/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package road

import (
	_ "embed"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

//go:embed testdata/e2e/road.yaml.tpl
var roadTemplate string

var E2ETestConfig = plugin.TestConfig{
	Name:               "road",
	DockerImage:        "",
	VersionQualifier:   "dev",
	VersionFileName:    "road.yaml",
	Template:           roadTemplate,
	EmptyFileContent:   nil,
	HasBeforeStartHook: false,
}

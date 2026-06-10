/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package standard

import (
	_ "embed"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

//go:embed testdata/e2e/version.txt.tpl
var versionTemplate string

var E2ETestConfig = plugin.TestConfig{
	Name:               "standard",
	DockerImage:        "",
	VersionQualifier:   "dev",
	VersionFileName:    "version.txt",
	Template:           versionTemplate,
	EmptyFileContent:   nil,
	HasBeforeStartHook: true,
}

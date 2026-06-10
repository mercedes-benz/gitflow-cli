/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package npm

import (
	_ "embed"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

//go:embed testdata/e2e/package.json.tpl
var packageTemplate string

var E2ETestConfig = plugin.TestConfig{
	Name:               "npm",
	DockerImage:        dockerImage,
	VersionQualifier:   "dev",
	VersionFileName:    "package.json",
	Template:           packageTemplate,
	EmptyFileContent:   []byte("{}"),
	HasBeforeStartHook: true,
}

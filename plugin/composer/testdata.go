/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package composer

import (
	_ "embed"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

//go:embed testdata/e2e/composer.json.tpl
var composerTemplate string

var E2ETestConfig = plugin.TestConfig{
	Name:               "composer",
	DockerImage:        dockerImage,
	VersionQualifier:   "dev",
	VersionFileName:    "composer.json",
	Template:           composerTemplate,
	EmptyFileContent:   []byte("{}"),
	HasBeforeStartHook: true,
}

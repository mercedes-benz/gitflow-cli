/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package mvn

import (
	_ "embed"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

//go:embed testdata/e2e/pom.xml.tpl
var pomTemplate string

var E2ETestConfig = plugin.TestConfig{
	Name:               "mvn",
	DockerImage:        dockerImage,
	VersionQualifier:   "SNAPSHOT",
	VersionFileName:    "pom.xml",
	Template:           pomTemplate,
	EmptyFileContent:   nil,
	HasBeforeStartHook: false,
}

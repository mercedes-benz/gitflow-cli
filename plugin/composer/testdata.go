/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package composer

import "github.com/mercedes-benz/gitflow-cli/core/plugin"

var E2ETestConfig = plugin.TestConfig{
	Name:             "composer",
	DockerImage:      dockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "composer.json",
	Template: `{
    "name": "mercedes-benz/gitflow-cli-test",
    "description": "Test for gitflow-cli",
    "type": "gitflow-cli-test",
    "license": "MIT",
    "version": "{{.Version}}",
    "minimum-stability": "dev",
    "prefer-stable": true
}
`,
	EmptyFileContent:   []byte("{}"),
	HasBeforeStartHook: true,
}

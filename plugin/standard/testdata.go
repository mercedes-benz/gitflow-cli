/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package standard

import "github.com/mercedes-benz/gitflow-cli/core/plugin"

var E2ETestConfig = plugin.TestConfig{
	Name:             "standard",
	DockerImage:      "",
	VersionQualifier: "dev",
	VersionFileName:  "version.txt",
	Template:         "{{.Version}}",
	EmptyFileContent: nil,
	HasBeforeStartHook: true,
}

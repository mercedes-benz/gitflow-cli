/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package npm

import "github.com/mercedes-benz/gitflow-cli/core/plugin"

var E2ETestConfig = plugin.TestConfig{
	Name:             "npm",
	DockerImage:      dockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "package.json",
	Template: `{
  "name": "gitflow-cli-test",
  "version": "{{.Version}}",
  "description": "Test package for gitflow-cli",
  "main": "index.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "license": "MIT"
}
`,
	EmptyFileContent:   []byte("{}"),
	HasBeforeStartHook: true,
}

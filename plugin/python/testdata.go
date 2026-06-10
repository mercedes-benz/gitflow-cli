/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package python

import (
	_ "embed"

	"github.com/mercedes-benz/gitflow-cli/core/plugin"
)

//go:embed testdata/e2e/pyproject_pep621.toml.tpl
var pyprojectTemplate string

//go:embed testdata/e2e/pyproject_poetry.toml.tpl
var poetryTemplate string

//go:embed testdata/e2e/setup.cfg.tpl
var setupCfgTemplate string

//go:embed testdata/e2e/setup.py.tpl
var setupPyTemplate string

var pythonSetupCommands = [][]string{
	{"pip", "install", "--quiet", "toml-cli"},
}

var E2ETestConfigPyproject = plugin.TestConfig{
	Name:               "python_pyproject",
	PluginName:         "python",
	DockerImage:        dockerImage,
	VersionQualifier:   "dev",
	VersionFileName:    "pyproject.toml",
	Template:           pyprojectTemplate,
	EmptyFileContent:   []byte{},
	HasBeforeStartHook: true,
	SetupCommands:      pythonSetupCommands,
}

var E2ETestConfigPoetry = plugin.TestConfig{
	Name:               "python_poetry",
	PluginName:         "python",
	DockerImage:        dockerImage,
	VersionQualifier:   "dev",
	VersionFileName:    "pyproject.toml",
	Template:           poetryTemplate,
	EmptyFileContent:   nil,
	HasBeforeStartHook: false,
	SetupCommands:      pythonSetupCommands,
}

var E2ETestConfigSetupCfg = plugin.TestConfig{
	Name:               "python_setup_cfg",
	PluginName:         "python",
	DockerImage:        dockerImage,
	VersionQualifier:   "dev",
	VersionFileName:    "setup.cfg",
	Template:           setupCfgTemplate,
	EmptyFileContent:   []byte{},
	HasBeforeStartHook: true,
	SetupCommands:      pythonSetupCommands,
}

var E2ETestConfigSetupPy = plugin.TestConfig{
	Name:               "python_setup_py",
	PluginName:         "python",
	DockerImage:        dockerImage,
	VersionQualifier:   "dev",
	VersionFileName:    "setup.py",
	Template:           setupPyTemplate,
	EmptyFileContent:   []byte{},
	HasBeforeStartHook: true,
	SetupCommands:      pythonSetupCommands,
}

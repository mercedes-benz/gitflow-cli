/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package python

import "github.com/mercedes-benz/gitflow-cli/core/plugin"

var pythonSetupCommands = [][]string{
	{"pip", "install", "--quiet", "toml-cli"},
}

var E2ETestConfigPyproject = plugin.TestConfig{
	Name:             "python_pyproject",
	PluginName:       "python",
	DockerImage:      dockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "pyproject.toml",
	Template: `[project]
name = "test-python-project"
version = "{{.Version}}"
description = "Test Python project"
requires-python = ">=3.8"
`,
	EmptyFileContent:   []byte{},
	HasBeforeStartHook: true,
	SetupCommands:      pythonSetupCommands,
}

var E2ETestConfigPoetry = plugin.TestConfig{
	Name:             "python_poetry",
	PluginName:       "python",
	DockerImage:      dockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "pyproject.toml",
	Template: `[tool.poetry]
name = "test-python-project"
version = "{{.Version}}"
description = "Test Python project"
authors = ["Test Author <test@example.com>"]

[tool.poetry.dependencies]
python = ">=3.8"

[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"
`,
	EmptyFileContent:   nil,
	HasBeforeStartHook: false,
	SetupCommands:      pythonSetupCommands,
}

var E2ETestConfigSetupCfg = plugin.TestConfig{
	Name:             "python_setup_cfg",
	PluginName:       "python",
	DockerImage:      dockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "setup.cfg",
	Template: `[metadata]
name = test-python-project
version = {{.Version}}
description = Test Python project
author = Test Author
author_email = test@example.com

[options]
python_requires = >=3.8
packages = find:
`,
	EmptyFileContent:   []byte{},
	HasBeforeStartHook: true,
	SetupCommands:      pythonSetupCommands,
}

var E2ETestConfigSetupPy = plugin.TestConfig{
	Name:             "python_setup_py",
	PluginName:       "python",
	DockerImage:      dockerImage,
	VersionQualifier: "dev",
	VersionFileName:  "setup.py",
	Template: `from setuptools import setup

setup(
    name="test-python-project",
    version="{{.Version}}",
    description="Test Python project",
    author="Test Author",
    author_email="test@example.com",
    py_modules=["mymodule"],
    python_requires=">=3.8",
)
`,
	EmptyFileContent:   []byte{},
	HasBeforeStartHook: true,
	SetupCommands:      pythonSetupCommands,
}

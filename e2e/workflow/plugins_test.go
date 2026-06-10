/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/mercedes-benz/gitflow-cli/plugin/composer"
	"github.com/mercedes-benz/gitflow-cli/plugin/mvn"
	"github.com/mercedes-benz/gitflow-cli/plugin/npm"
	"github.com/mercedes-benz/gitflow-cli/plugin/python"
	"github.com/mercedes-benz/gitflow-cli/plugin/road"
	"github.com/mercedes-benz/gitflow-cli/plugin/standard"
)

// pluginTestConfigs lists all plugins that should be tested in e2e workflows.
// Each plugin provides its own TestConfig with template, image, qualifier, etc.
var pluginTestConfigs = []plugin.TestConfig{
	standard.E2ETestConfig,
	mvn.E2ETestConfig,
	npm.E2ETestConfig,
	python.E2ETestConfigPyproject,
	python.E2ETestConfigPoetry,
	python.E2ETestConfigSetupCfg,
	python.E2ETestConfigSetupPy,
	composer.E2ETestConfig,
	road.E2ETestConfig,
}

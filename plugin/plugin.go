/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package plugin

import (
	// import all plugins here to make them available to the plugin registry
	_ "github.com/mercedes-benz/gitflow-cli/plugin/maven"
	_ "github.com/mercedes-benz/gitflow-cli/plugin/standard"
)

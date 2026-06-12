/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package workflow

import (
	"github.com/mercedes-benz/gitflow-cli/cmd"
	"github.com/mercedes-benz/gitflow-cli/e2e"
)

func init() {
	e2e.ExecuteFunc = cmd.Execute
}

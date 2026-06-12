/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package cmd

import (
	"runtime/debug"
	"time"
)

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	var revision, vcsTime, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.time":
			vcsTime = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				modified = " | dirty"
			}
		}
	}
	if revision == "" {
		return "dev"
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	t, err := time.Parse(time.RFC3339, vcsTime)
	if err != nil {
		return revision + modified
	}
	local := t.Local()
	formatted := local.Format("2006-01-02, 15:04")
	return revision + " (" + formatted + ")" + modified
}

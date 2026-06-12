/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package cmd

import "runtime/debug"

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	var revision, time, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.time":
			time = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				modified = " (dirty)"
			}
		}
	}
	if revision == "" {
		return "dev"
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	return revision + " " + time + modified
}

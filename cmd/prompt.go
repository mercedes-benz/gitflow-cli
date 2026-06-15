/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mercedes-benz/gitflow-cli/core"
	"github.com/mercedes-benz/gitflow-cli/core/plugin"
	"github.com/spf13/viper"
)

func initPrompts() {
	initBranchSync()
	initToolFallback()
}

func initBranchSync() {
	core.BranchSync = func(req core.BranchSyncRequest) (core.BranchSyncResult, error) {
		autoConfirm, _ := rootCmd.Flags().GetBool("yes")
		return handleBranchSync(req, autoConfirm)
	}
}

func handleBranchSync(req core.BranchSyncRequest, autoConfirm bool) (core.BranchSyncResult, error) {
	canCreate := req.CreateFrom != ""

	if autoConfirm {
		if canCreate {
			fmt.Fprintf(os.Stderr, "INFO: creating %s branch '%s' from '%s'\n",
				req.BranchType, req.Configured, req.CreateFrom)
			return core.BranchSyncResult{ResolvedName: req.Configured, Created: true}, nil
		}
		if len(req.Candidates) > 0 {
			chosen := req.Candidates[0]
			fmt.Fprintf(os.Stderr, "INFO: %s branch '%s' not found, using '%s'\n",
				req.BranchType, req.Configured, chosen)
			persistBranchToConfig(req.BranchType, chosen)
			return core.BranchSyncResult{ResolvedName: chosen, Persist: true}, nil
		}
		return core.BranchSyncResult{}, nil
	}

	fmt.Fprintf(os.Stderr, "%s branch '%s' not found.\n", req.BranchType, req.Configured)

	var input string
	if canCreate {
		fmt.Fprintf(os.Stderr, "Enter branch name or press Enter to create '%s': ", req.Configured)
		input = readLine()
		if input == "" {
			fmt.Fprintf(os.Stderr, "Creating '%s' from '%s'...\n", req.Configured, req.CreateFrom)
			return core.BranchSyncResult{ResolvedName: req.Configured, Created: true}, nil
		}
	} else if len(req.Candidates) > 0 {
		fmt.Fprintf(os.Stderr, "Enter branch name [%s]: ", req.Candidates[0])
		input = readLine()
		if input == "" {
			input = req.Candidates[0]
		}
	} else {
		fmt.Fprintf(os.Stderr, "Enter existing branch name: ")
		input = readLine()
		if input == "" {
			return core.BranchSyncResult{}, nil
		}
	}

	exists := branchExistsOnRemote(req, input)

	if exists {
		if input != req.Configured {
			persistBranchToConfig(req.BranchType, input)
		}
		return core.BranchSyncResult{ResolvedName: input, Persist: input != req.Configured}, nil
	}

	if !canCreate {
		fmt.Fprintf(os.Stderr, "Branch '%s' does not exist on remote.\n", input)
		return core.BranchSyncResult{}, nil
	}

	fmt.Fprintf(os.Stderr, "Creating '%s' from '%s'...\n", input, req.CreateFrom)
	if input != req.Configured {
		persistBranchToConfig(req.BranchType, input)
	}
	return core.BranchSyncResult{ResolvedName: input, Created: true, Persist: input != req.Configured}, nil
}

func branchExistsOnRemote(req core.BranchSyncRequest, name string) bool {
	for _, c := range req.Candidates {
		if c == name {
			return true
		}
	}
	if req.Repository != nil {
		if found, err := req.Repository.HasRemoteBranch(name); err == nil && found {
			return true
		}
	}
	return false
}

func persistBranchToConfig(branchType core.Branch, name string) {
	key := "core." + branchType.ConfigKey()
	viper.Set(key, name)
	if err := viper.WriteConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: could not save config: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Configured '%s: %s'\n", key, name)
	}
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func initToolFallback() {
	plugin.ToolFallbackFunc = func(tool string, image string) (bool, error) {
		autoConfirm, _ := rootCmd.Flags().GetBool("yes")

		if autoConfirm {
			fmt.Fprintf(os.Stderr, "INFO: %s not found, using Docker (%s)\n", tool, image)
			return true, nil
		}

		fmt.Fprintf(os.Stderr, "%s not found. Use Docker (%s) instead? [Y/n] ", tool, image)

		answer := readLine()
		if answer != "" && answer != "y" && answer != "yes" {
			return false, nil
		}

		return true, nil
	}
}

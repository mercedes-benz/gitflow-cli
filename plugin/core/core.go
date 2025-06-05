/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Tools and names required for the workflow automation commands.
const (
	Git    = "git"
	Remote = "origin"
)

// Branch types for the Gitflow model on which the workflow automation commands operate.
const (
	_ Branch = iota
	Production
	Development
	Release
	Hotfix
)

// Merge types for repository merging operations.
const (
	_ MergeType = iota
	Squash
	NoFastForward
	FastForward
)

type (
	// Plugins is the list of all registered plugins.
	Plugins []Plugin

	// Branch represents branch types in the Gitflow model.
	Branch int

	// MergeType represents merge types for repository merging operations.
	MergeType int

	// Plugin is the interface for all workflow automation plugins.
	Plugin interface {
		VersionFileName() string
		VersionQualifier() string
		RequiredTools() []string
		ReadVersion(repository Repository) (Version, error)
		WriteVersion(repository Repository, version Version) error
		fmt.Stringer
	}
)

// Settings group for the core package.
const settingsGroup = "core"

// UndoSetting controls undo-behavior for all local changes in a repository.
const undoSetting = "undo"

// Git version control system tool commands.
const (
	status        = "status"
	fetch         = "fetch"
	pull          = "pull"
	switch_       = "switch"
	merge         = "merge"
	add           = "add"
	commit        = "commit"
	branch        = "branch"
	tag           = "tag"
	push          = "push"
	clean         = "clean"
	reset         = "reset"
	create        = "-c"
	forcedelete   = "-D"
	dir           = "-d"
	ignored       = "-x"
	porcelain     = "--porcelain"
	upstream      = "--set-upstream"
	all           = "--all"
	tags          = "--tags"
	prune         = "--prune"
	delete        = "--delete"
	remotes       = "--remotes"
	message       = "--message"
	squash        = "--squash"
	nofastforward = "--no-ff"
	fastforwad    = "--ff-only"
	force         = "--force"
	hard          = "--hard"
)

// BranchNames maps branch types to their names.
var branchNames = map[Branch]string{
	Production:  "main",
	Development: "develop",
	Release:     "release",
	Hotfix:      "hotfix",
}

// BranchSettings maps settings to branch names.
var branchSettings = map[string]Branch{
	"production":  Production,
	"development": Development,
	"release":     Release,
	"hotfix":      Hotfix,
}

var undoChanges = false

// PluginRegistry is the global list of all registered plugins.
var pluginRegistry Plugins
var pluginRegistryLock sync.Mutex
var fallbackPlugin Plugin

// RegisterPlugin adds a plugin to the global list of all registered plugins.
func RegisterPlugin(plugin Plugin) {
	pluginRegistryLock.Lock()
	defer pluginRegistryLock.Unlock()
	pluginRegistry = append(pluginRegistry, plugin)
}

// RegisterFallbackPlugin RegisterPlugin adds a fallback plugin
func RegisterFallbackPlugin(plugin Plugin) {
	fallbackPlugin = plugin
}

// CheckVersionFile checks if version file is found
func CheckVersionFile(projectPath string, versionFile string) bool {
	_, err := os.Stat(filepath.Join(projectPath, versionFile))
	return !os.IsNotExist(err)
}

// ValidateToolsAvailability Check if some tools are available in the system.
func ValidateToolsAvailability(tools ...string) error {
	for _, tool := range append(tools, Git) {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("tool '%v' is not available on the system", tool)
		}
	}

	return nil
}

// String representation of a branch type.
func (b Branch) String() string {
	return branchNames[b]
}

// Apply suitable settings from the global configuration to the core package.
func applySettings() {
	log.SetOutput(os.Stdout)
	all := viper.AllSettings()

	if settings, ok := all[settingsGroup].(map[string]any); !ok {
		return
	} else {
		for key, value := range settings {
			if key == loggingSetting {
				// configure logging behavior for all repository operations
				if v, ok := value.(string); ok {
					// first reset logging flags to off if configuration is found
					loggingFlags = 0

					// logging output goes to standard output
					if strings.Contains(v, StdErr.String()) {
						loggingFlags |= StdErr
					}

					// logging output goes to standard error
					if strings.Contains(v, StdOut.String()) {
						loggingFlags |= StdOut
					}

					// log command line with all arguments
					if strings.Contains(v, CmdLine.String()) {
						loggingFlags |= CmdLine
					}

					// log output of all command lines
					if strings.Contains(v, Output.String()) {
						loggingFlags |= Output
					}

					// turn off logging must be the last option
					if strings.Contains(v, Off.String()) {
						loggingFlags = 0
					}
				}
			} else if key == undoSetting {
				// configure undo-behavior for all local changes in a repository
				if v, ok := value.(bool); ok {
					undoChanges = v
				}
			} else if b, ok := branchSettings[key]; ok {
				// configure branch names for the Gitflow model
				if v, ok := value.(string); ok || len(v) > 0 {
					branchNames[b] = v
				}
			}
		}
	}
}

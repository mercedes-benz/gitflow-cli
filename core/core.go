/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"fmt"
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

	// Plugin is the fundamental interface that must be implemented by all workflow automation plugins.
	Plugin interface {
		// VersionFileName returns the name of the file that contains the version information in the project.
		// For example: "pom.xml" for Maven, etc.
		VersionFileName() string

		// SetVersionFileName sets the name of the file that contains the version information.
		SetVersionFileName(fileName string)

		// VersionFileNames returns an optional list of file names that contain version information.
		// This is an alternative to VersionFileName for plugins that support multiple version files.
		VersionFileNames() []string

		// VersionQualifier returns the suffix that is appended to SNAPSHOT versions.
		// For example: "SNAPSHOT" for Maven, etc.
		VersionQualifier() string

		// RequiredTools returns a list of command-line tools needed to run the plugin.
		RequiredTools() []string

		// ReadVersion reads the current version from the project file.
		ReadVersion(repository Repository) (Version, error)

		// WriteVersion writes the provided version to the project file.
		WriteVersion(repository Repository, version Version) error

		// Stringer returns the human-readable name of the plugin.
		fmt.Stringer
	}
)

// Configuration groups.
const (
	branchesGroup = "branches"
	workflowGroup = "workflow"
	loggingKey    = "logging"
	legacyGroup   = "core"
)

// Workflow settings keys.
const rollbackSetting = "rollback"
const pushSetting = "push"
const dockerFallbackSetting = "docker-fallback"

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

var rollbackChanges = false
var pushChanges = true

// DockerFallback indicates whether to automatically fall back to Docker when a native tool is missing.
var DockerFallback = false

// ProjectPath holds the path to the Git repository
var ProjectPath = "."

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
func CheckVersionFile(plugin Plugin) bool {
	// If plugin supports multiple version files, detect the correct one for the current project
	if versionFileNames := plugin.VersionFileNames(); len(versionFileNames) > 0 {
		for _, versionFile := range versionFileNames {
			if _, err := os.Stat(filepath.Join(ProjectPath, versionFile)); !os.IsNotExist(err) {
				plugin.SetVersionFileName(versionFile)
				return true
			}
		}
		return false
	}

	// If VersionFileName is set, use it directly
	if versionFileName := plugin.VersionFileName(); versionFileName != "" {
		if _, err := os.Stat(filepath.Join(ProjectPath, versionFileName)); !os.IsNotExist(err) {
			return true
		}
	}

	return false
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

// ResetBranchNames restores default branch names. Used by tests to prevent state leakage.
func ResetBranchNames() {
	branchNames[Production] = "main"
	branchNames[Development] = "develop"
	branchNames[Release] = "release"
	branchNames[Hotfix] = "hotfix"
}

// branchConfigKeys maps Branch constants to their config key names.
var branchConfigKeys = map[Branch]string{
	Production:  "production",
	Development: "development",
	Release:     "release",
	Hotfix:      "hotfix",
}

// ConfigKey returns the config key name for this branch type.
func (b Branch) ConfigKey() string {
	return branchConfigKeys[b]
}

// Apply suitable settings from the global configuration to the core package.
func applySettings() {
	all := viper.AllSettings()

	if branches, ok := all[branchesGroup].(map[string]any); ok {
		applyBranchSettings(branches)
	} else if legacy, ok := all[legacyGroup].(map[string]any); ok {
		applyBranchSettings(legacy)
	}

	if wf, ok := all[workflowGroup].(map[string]any); ok {
		applyWorkflowSettings(wf)
	} else if legacy, ok := all[legacyGroup].(map[string]any); ok {
		applyWorkflowSettings(legacy)
	}

	if v, ok := all[loggingKey].(string); ok {
		applyLoggingSettings(v)
	} else if legacy, ok := all[legacyGroup].(map[string]any); ok {
		if v, ok := legacy[loggingSetting].(string); ok {
			applyLoggingSettings(v)
		}
	}
}

func applyBranchSettings(settings map[string]any) {
	for key, value := range settings {
		if b, ok := branchSettings[key]; ok {
			if v, ok := value.(string); ok && len(v) > 0 {
				branchNames[b] = v
			}
		}
	}
}

func applyWorkflowSettings(settings map[string]any) {
	if v, ok := settings[rollbackSetting].(bool); ok {
		rollbackChanges = v
	}
	// Legacy: accept "undo" as alias for "rollback"
	if v, ok := settings["undo"].(bool); ok {
		rollbackChanges = v
	}
	if v, ok := settings[pushSetting].(bool); ok {
		pushChanges = v
	}
	if v, ok := settings[dockerFallbackSetting].(bool); ok {
		DockerFallback = v
	}
}

func applyLoggingSettings(v string) {
	loggingFlags = 0
	if strings.Contains(v, StdErr.String()) {
		loggingFlags |= StdErr
	}
	if strings.Contains(v, StdOut.String()) {
		loggingFlags |= StdOut
	}
	if strings.Contains(v, CmdLine.String()) {
		loggingFlags |= CmdLine
	}
	if strings.Contains(v, Output.String()) {
		loggingFlags |= Output
	}
	if strings.Contains(v, Off.String()) {
		loggingFlags = 0
	}
}

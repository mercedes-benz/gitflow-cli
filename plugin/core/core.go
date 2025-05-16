/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Tools and names required for the workflow automation commands.
const (
	Git    = "git"
	Maven  = "mvn"
	Remote = "origin"
)

// Logging bit flags for controlling logging behavior for all repository operations.
const (
	_ Logging = 1 << iota
	Off
	StdErr
	StdOut
	CmdLine
	Output
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

// Version increment types for the workflow automation commands.
const (
	None VersionIncrement = iota
	Major
	Minor
	Incremental
)

// todo: should be solved as a hook
//const defaultVersionFile = "version.txt"

type (
	// Plugins is the list of all registered plugins.
	Plugins []Plugin

	// Logging controls logging behavior for all repository operations.
	Logging int

	// Branch represents branch types in the Gitflow model.
	Branch int

	// MergeType represents merge types for repository merging operations.
	MergeType int

	// VersionIncrement Type of version increment.
	VersionIncrement int

	// todo: check if this is still needed
	//// StartCallback Default callback functions that run custom business logic for release and hotfix branches.
	//StartCallback  func(repo Repository, args ...any) error
	//FinishCallback func(repo Repository) error

	// Plugin is the interface for all workflow automation plugins.
	Plugin interface {
		Precondition
		Name() string // todo: replace with String()
		SnapshotQualifier() string
		UpdateProjectVersion(next Version) error
	}

	// Precondition is the interface for checking if a plugin can be executed in a project directory.
	Precondition interface {
		Check(projectPath string) bool
		Version(projectPath string, major, minor, incremental bool) (Version, Version, error)
	}

	// Version represents a version-stamp with a major, minor, incremental part, and optionally empty qualifier.
	Version struct {
		VersionIncrement                     VersionIncrement
		Major, Minor, Incremental, Qualifier string
	}
)

// NoVersion is the default version without any parts.
var NoVersion Version

// Settings group for the core package.
const settingsGroup = "core"

// LoggingSetting controls logging behavior for all repository operations.
const loggingSetting = "logging"

// UndoSetting controls undo-behavior for all local changes in a repository.
const undoSetting = "undo"

// VersionStamp is the format for version strings.
const versionStamp = "%v.%v.%v"

// VersionStampWithQualifier is the format for version strings with a qualifier.
const versionStampWithQualifier = "%v.%v.%v-%v"

// VersionExpression is the regular expression for version strings with optional qualifier.
const versionExpression = `(\d+)\.(\d+)\.(\d+)(?:-(\w+))?$`

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

// LoggingNames maps logging flags to their names.
var loggingNames = map[Logging]string{
	Off:     "off",
	StdErr:  "stderr",
	StdOut:  "stdout",
	CmdLine: "cmdline",
	Output:  "output",
}

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

// Internal flags for controlling core package behavior.
var loggingFlags = StdOut | CmdLine | Output
var undoChanges = false

// PlugInRegistry is the global list of all registered plugins.
var pluginRegistry Plugins
var pluginRegistryLock sync.Mutex

// todo: what is this?
// NoQualifier is the default empty qualifier for versions.
var noQualifier = ""

// NewVersion Create new version with major, minor, incremental, and qualifier.
func NewVersion(major, minor, incremental string, args ...any) Version {
	var version Version

	// look for qualifier and version increment type in the arguments
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			version.Qualifier = arg

		case VersionIncrement:
			version.VersionIncrement = arg
		}
	}

	// set major, minor, and incremental version parts
	version.Major = major
	version.Minor = minor
	version.Incremental = incremental
	return version
}

// ParseVersion Parse a version string with major, minor, incremental, and optional qualifier.
func ParseVersion(version string) (Version, error) {
	var v Version

	// match a version string with optional qualifier
	matches := regexp.MustCompile(versionExpression).FindStringSubmatch(version)

	// check if the version string matches the regular expression
	if matches == nil {
		return v, fmt.Errorf("invalid version string: %v", version)
	}

	// set the major, minor, and incremental version parts
	v.Major = matches[1]
	v.Minor = matches[2]
	v.Incremental = matches[3]

	// check if the version string has a qualifier
	if len(matches) == 5 {
		v.Qualifier = matches[4]
	}

	return v, nil
}

// Register adds a plugin to the global list of all registered plugins.
func Register(plugin Plugin) {
	pluginRegistryLock.Lock()
	defer pluginRegistryLock.Unlock()
	pluginRegistry = append(pluginRegistry, plugin)
}

// ValidateArgumentsLength Check if the number of arguments matches the expected number.
func ValidateArgumentsLength(expected int, args ...any) error {
	if len(args) != expected {
		return fmt.Errorf("expected %v arguments, but got %v", expected, len(args))
	}

	return nil
}

// ValidateArgumentsType Check if all arguments are of a specific type.
func ValidateArgumentsType(t reflect.Type, args ...any) error {
	for _, arg := range args {
		if reflect.TypeOf(arg) != t {
			return fmt.Errorf("expected arguments of type %T, but got %T", t, reflect.TypeOf(arg))
		}
	}

	return nil
}

// ValidateToolsAvailability Check if some tools are available in the system.
// todo: should be implemented by each plugin
func ValidateToolsAvailability(tools ...string) error {
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("tool '%v' is not available on the system", tool)
		}
	}

	return nil
}

// Log a message to Go standard logging based on logging flags and variadic arguments.
func Log(message ...any) {
	println := func() {
		for _, msg := range message {
			switch msg := msg.(type) {
			case string:
				if len(msg) > 0 && (loggingFlags&CmdLine != 0 || loggingFlags&Output != 0) {
					log.Println(msg)
				}

			case *exec.Cmd:
				if msg != nil && len(msg.String()) > 0 && loggingFlags&CmdLine != 0 {
					log.Println(msg.String())
				}

			case []byte:
				if len(msg) > 0 && loggingFlags&Output != 0 {
					output := strings.TrimRight(string(msg), "\n\r")
					log.Println(output)
				}

			case error:
				if msg != nil && len(msg.Error()) > 0 && loggingFlags&Output != 0 {
					log.Println(msg.Error())
				}

			default:
				if msg != nil && len(fmt.Sprintf("%v", msg)) > 0 && loggingFlags&Output != 0 {
					log.Println(msg)
				}
			}
		}
	}

	if loggingFlags&StdErr != 0 {
		log.SetOutput(os.Stderr)
		println()
	}

	if loggingFlags&StdOut != 0 {
		log.SetOutput(os.Stdout)
		println()
	}
}

// String representation of a logging flag (only one allowed at a time).
func (l Logging) String() string {
	return loggingNames[l]
}

// String representation of a branch type.
func (b Branch) String() string {
	return branchNames[b]
}

// Format a version string with major, minor, incremental, and optionally empty qualifier.
func (v Version) String() string {
	if v.Qualifier == noQualifier {
		return fmt.Sprintf(versionStamp, v.Major, v.Minor, v.Incremental)
	}

	return fmt.Sprintf(versionStampWithQualifier, v.Major, v.Minor, v.Incremental, v.Qualifier)
}

// BranchName Create a branch name with a specific version and branch type.
func (v Version) BranchName(branch Branch) string {
	return fmt.Sprintf("%v/%v", branch, v)
}

// Increment Determine next version based on version increment type and next major, minor, and incremental version strings.
func (current Version) Increment(increment VersionIncrement, nextMajor, nextMinor, nextIncremental string) (Version, error) {
	switch increment {
	case Major:
		return NewVersion(nextMajor, "0", "0", current.Qualifier, increment), nil

	case Minor:
		return NewVersion(current.Major, nextMinor, "0", current.Qualifier, increment), nil

	case Incremental:
		return NewVersion(current.Major, current.Minor, nextIncremental, current.Qualifier, increment), nil

	default:
		return NoVersion, fmt.Errorf("unsupported version increment type: %v", increment)
	}
}

// Next Determine the next version based on the current version and the version increment type.
func (current Version) Next(increment VersionIncrement) (Version, error) {
	nextMajor, errMajor := strconv.Atoi(current.Major)
	nextMinor, errMinor := strconv.Atoi(current.Minor)
	nextIncremental, errIncremental := strconv.Atoi(current.Incremental)

	if errMajor != nil || errMinor != nil || errIncremental != nil {
		return NoVersion, errors.Join(fmt.Errorf("invalid version parts: %v", current), errMajor, errMinor, errIncremental)
	}

	nextMajor++
	nextMinor++
	nextIncremental++
	return current.Increment(increment, strconv.Itoa(nextMajor), strconv.Itoa(nextMinor), strconv.Itoa(nextIncremental))
}

// AddQualifier Add a qualifier to the version.
func (v Version) AddQualifier(qualifier string) Version {
	return NewVersion(v.Major, v.Minor, v.Incremental, qualifier, v.VersionIncrement)
}

// RemoveQualifier Remove the qualifier from the version.
func (v Version) RemoveQualifier() Version {
	return NewVersion(v.Major, v.Minor, v.Incremental, noQualifier, v.VersionIncrement)
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

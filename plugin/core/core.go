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

	// StartCallback Default callback functions that run custom business logic for release and hotfix branches.
	StartCallback  func(repo Repository, args ...any) error
	FinishCallback func(repo Repository) error

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

	// Repository represents a version control system repository.
	Repository interface {
		Local() string
		IsClean() error
		HasBranch(branch Branch) (bool, []string, error)
		CheckoutBranch(branchName string) error
		CreateBranch(branchName string) error
		MergeBranch(branchName string, mergeType MergeType) error
		PullBranch(branchName string) error
		DeleteBranch(branchName string) error
		AddFile(file string) error
		CommitChanges(message string) error
		TagCommit(tagName string) error
		PushChanges(branchName string) error
		PushAllChanges() error
		PushAllTags() error
		PushDeletion(branchName string) error
		UndoAllChanges(cause error) error
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

// Implementation of the Repository interface.
type repository struct {
	projectPath, remote string
	statusClean         []string
	fetchAll            []string
	allRemotes          []string
	allLocals           []string
	switchBranch        []string
	createBranch        []string
	mergeBranch         []string
	pullBranch          []string
	deleteBranch        []string
	forceDeleteBranch   []string
	addFile             []string
	commitAll           []string
	tagCommit           []string
	pushBranch          []string
	pushAll             []string
	pushTags            []string
	pushDeletion        []string
	cleanAll            []string
	resetBranch         []string
}

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
var loggingFlags Logging = StdOut | CmdLine | Output
var undoChanges bool = false

// PlugInRegistry is the global list of all registered plugins.
var pluginRegistry Plugins
var pluginRegistryLock sync.Mutex

// NoQualifier is the default empty qualifier for versions.
var noQualifier = ""

// NewRepository enables access to a version control system repository.
func NewRepository(projectPath, remote string) Repository {
	return &repository{
		projectPath:       projectPath,
		remote:            remote,
		statusClean:       []string{status, porcelain},
		fetchAll:          []string{fetch, all, prune},
		allRemotes:        []string{branch, remotes},
		allLocals:         []string{branch},
		switchBranch:      []string{switch_},
		createBranch:      []string{switch_, create},
		mergeBranch:       []string{merge},
		pullBranch:        []string{pull, remote},
		deleteBranch:      []string{branch, delete},
		forceDeleteBranch: []string{branch, forcedelete},
		addFile:           []string{add},
		commitAll:         []string{commit, all, message},
		tagCommit:         []string{tag},
		pushBranch:        []string{push, upstream, remote},
		pushAll:           []string{push, all, remote},
		pushTags:          []string{push, tags, remote},
		pushDeletion:      []string{push, delete, remote},
		cleanAll:          []string{clean, force, dir, ignored},
		resetBranch:       []string{reset, hard},
	}
}

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

// Start executes the first plugin that meets the precondition.
func Start(branch Branch, projectPath string, args ...any) error {
	pluginRegistryLock.Lock()
	defer pluginRegistryLock.Unlock()

	// apply suitable settings from the global configuration to the core package
	applySettings()

	// check if project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path '%v' does not exist", projectPath)
	}

	// execute the first plugin that meets the precondition
	for _, plugin := range pluginRegistry {
		if plugin.Check(projectPath) {
			// todo: this part can be generalized for both tasks (start and finish)
			// get access to the local version control system
			repo := NewRepository(projectPath, Remote)

			// check if required tools are available
			// todo: solve similar to another hook direct in plugin
			//if err := ValidateToolsAvailability(tools...); err != nil {
			//	return err
			//}

			// check if the repository prerequisites are met
			if err := repo.IsClean(); err != nil {
				return err
			}

			// format start command messages
			prefix := fmt.Sprintf("%v Plugin Start on branch", plugin.Name()) // todo: replace with String()
			called := fmt.Sprintf("%v %v called: %v", prefix, branch.String(), repo.Local())
			completed := fmt.Sprintf("%v %v completed: %v", prefix, branch, repo.Local())
			failed := fmt.Sprintf("%v %v failed: %v", prefix, branch, repo.Local())

			switch branch {
			case Release:
				fmt.Println(called)

				// start command requires two arguments 'major' and 'minor'
				if err := ValidateArgumentsLength(2, args...); err != nil {
					return err
				}

				// start command requires all arguments to be of type bool
				if err := ValidateArgumentsType(reflect.TypeOf(true), args...); err != nil {
					return err
				}

				// run the release start command
				// todo: do args optional and generic
				//if err := start(repo, args[0].(bool), args[1].(bool)); err != nil {
				if err := releaseStart(repo, plugin, args[0].(bool), args[1].(bool)); err != nil {
					fmt.Println(failed)
					return err
				}

				fmt.Println(completed)
				return nil

			case Hotfix:
				fmt.Println(called)

				// run the hotfix start command
				if err := hotfixStart(repo, plugin); err != nil {
					fmt.Println(failed)
					return err
				}

				fmt.Println(completed)
				return nil

			default:
				return fmt.Errorf("unsupported branch: %v", branch)
			}
		}
	}

	// todo: solve as a beforeStartHook
	//if !pluginMatched {
	//	repo := NewRepository(projectPath, Remote)
	//	if err := repo.CheckoutBranch(Development.String()); err != nil {
	//		return repo.UndoAllChanges(err)
	//	}
	//
	//	initVersion := NewVersion("1", "0", "0", "dev")
	//	if err := os.WriteFile(defaultVersionFile, []byte(initVersion.String()), 0644); err != nil {
	//		return repo.UndoAllChanges(err)
	//	}
	//
	//	if err := repo.AddFile(defaultVersionFile); err != nil {
	//		return repo.UndoAllChanges(err)
	//	}
	//
	//	if err := repo.CommitChanges("Create versions file"); err != nil {
	//		return repo.UndoAllChanges(err)
	//	}
	//
	//	for _, plugin := range pluginRegistry {
	//		if plugin.Check(projectPath) {
	//			pluginMatched = true
	//			if err := plugin.Start(branch, projectPath, args...); err != nil {
	//				return err
	//			}
	//		}
	//	}
	//}
	return nil
}

// Finish executes the first plugin that meets the precondition.
func Finish(branch Branch, projectPath string) error {

	// todo: begin: maybe generalize as well
	pluginRegistryLock.Lock()
	defer pluginRegistryLock.Unlock()

	// apply suitable settings from the global configuration to the core package
	applySettings()

	// check if project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path '%v' does not exist", projectPath)
	}
	// todo: end: maybe generalize as well

	// execute the first plugin that meets the precondition
	for _, plugin := range pluginRegistry {
		if plugin.Check(projectPath) {

			// finish the workflow with the selected release business logic
			repo := NewRepository(projectPath, Remote)

			// check if required tools are available
			// todo: fix it (create a hook)
			//if err := ValidateToolsAvailability(tools...); err != nil {
			//	return err
			//}

			// check if the repository prerequisites are met
			if err := repo.IsClean(); err != nil {
				return err
			}

			// format finish command messages
			// todo: check if plugin returns a text
			prefix := fmt.Sprintf("%v Plugin Finish on branch", plugin.Name()) // todo: replace with String()
			called := fmt.Sprintf("%v %v called: %v", prefix, branch.String(), repo.Local())
			completed := fmt.Sprintf("%v %v completed: %v", prefix, branch, repo.Local())
			failed := fmt.Sprintf("%v %v failed: %v", prefix, branch, repo.Local())

			fmt.Println(called)

			// select suitable business logic for the branch
			switch branch {
			case Release:

				// run the release finish command
				if err := releaseFinish(repo, plugin); err != nil {
					fmt.Println(failed)
					return err
				}

				fmt.Println(completed)
				return nil

			case Hotfix:

				// run the hotfix finish command
				if err := hotfixFinish(repo, plugin); err != nil {
					fmt.Println(failed)
					return err
				}

				fmt.Println(completed)
				return nil

			default:
				return fmt.Errorf("unsupported branch: %v", branch)
			}
		}
	}

	return fmt.Errorf("no plugin meets the precondition for branch '%v' and project path '%v'", branch, projectPath)
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

// Local Return the local path of the repository.
func (r *repository) Local() string {
	return r.projectPath
}

// IsClean Check if the repository under the project path is clean.
func (r *repository) IsClean() error {
	var err error
	var status *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(status, output, err) }()

	// get the status of the repository
	status = exec.Command(Git, r.statusClean...)
	status.Dir = r.projectPath

	// run git command to get the status
	if output, err = status.CombinedOutput(); err != nil {
		return fmt.Errorf("git 'status' failed with %v: %s", err, output)
	} else if len(output) != 0 {
		return fmt.Errorf("repository under project path '%v' is not clean", status.Dir)
	}

	return nil
}

// HasBranch Check if a branch exists in the repository.
func (r *repository) HasBranch(branch Branch) (bool, []string, error) {
	var remotes []string
	var logs []any = make([]any, 0)

	// log human-readable description of the git command
	defer func() { Log(logs...) }()

	// fetch and prune all remote branches
	fetch := exec.Command(Git, r.fetchAll...)
	fetch.Dir = r.projectPath

	// run git command to fetch all remotes
	if output, err := fetch.CombinedOutput(); err != nil {
		logs = append(logs, fetch, output, err)
		return false, nil, fmt.Errorf("fetching all remotes failed with %v: %s", err, output)
	} else {
		logs = append(logs, fetch, output)
	}

	// list all remotes of the repository
	all := exec.Command(Git, r.allRemotes...)
	all.Dir = r.projectPath

	// run git command to list all remotes
	if output, err := all.CombinedOutput(); err != nil {
		logs = append(logs, all, output, err)
		return false, nil, fmt.Errorf("getting all remotes failed with %v: %s", err, output)
	} else {
		logs = append(logs, all, output)

		// check every line of the output for the branch name
		for _, remote := range strings.Split(string(output), "\n") {
			if remote = strings.TrimSpace(remote); strings.Contains(remote, branch.String()) {
				if strings.HasPrefix(remote, r.remote) {
					remotes = append(remotes, remote)
				}
			}
		}
	}

	return len(remotes) > 0, remotes, nil
}

// CheckoutBranch Checkout a specific branch in the repository.
func (r *repository) CheckoutBranch(branchName string) error {
	var err error
	var checkout *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(checkout, output, err) }()

	// checkout branch
	checkout = exec.Command(Git, append(r.switchBranch, branchName)...)
	checkout.Dir = r.projectPath

	// run git command to checkout branch
	if output, err = checkout.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' '%v' failed with %v: %s", checkout, branchName, err, output)
	}

	return nil
}

// CreateBranch Create a new branch in the repository with a specific name.
func (r *repository) CreateBranch(branchName string) error {
	var err error
	var create *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(create, output, err) }()

	// create a new branch with the specific name
	create = exec.Command(Git, append(r.createBranch, branchName)...)
	create.Dir = r.projectPath

	// run git command to create a new branch
	if output, err = create.CombinedOutput(); err != nil {
		return fmt.Errorf("git create new '%v' failed with %v: %s", branchName, err, output)
	}

	return nil
}

// MergeBranch Merge a branch into the current branch in the repository with a specific merge type.
func (r *repository) MergeBranch(branchName string, mergeType MergeType) error {
	var option string
	var err error
	var merge *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(merge, output, err) }()

	// determine the merge option string based on the merge type
	switch mergeType {
	case Squash:
		option = squash

	case NoFastForward:
		option = nofastforward

	case FastForward:
		option = fastforwad

	default:
		err = fmt.Errorf("unsupported merge type: %v", mergeType)
		return err
	}

	// merge branch into the current branch
	merge = exec.Command(Git, append(r.mergeBranch, option, branchName)...)
	merge.Dir = r.projectPath

	// run git command to merge branch
	if output, err = merge.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' '%v' failed with %v: %s", merge, branchName, err, output)
	}

	return nil
}

// PullBranch Pull changes in a branch from the remote repository.
func (r *repository) PullBranch(branchName string) error {
	var err error
	var pull *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(pull, output, err) }()

	// pull changes from the remote repository
	pull = exec.Command(Git, append(r.pullBranch, branchName)...)
	pull.Dir = r.projectPath

	// run git command to pull changes
	if output, err = pull.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' failed with %v: %s", pull, err, output)
	}

	return nil
}

// DeleteBranch Delete a local branch in the repository with a specific name.
func (r *repository) DeleteBranch(branchName string) error {
	var err error
	var delete *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(delete, output, err) }()

	// delete the branch with the specific name
	delete = exec.Command(Git, append(r.deleteBranch, branchName)...)
	delete.Dir = r.projectPath

	// run git command to delete the branch
	if output, err = delete.CombinedOutput(); err != nil {
		return fmt.Errorf("git delete '%v' failed with %v: %s", branchName, err, output)
	}

	return nil
}

// AddFile Add file to git
func (r *repository) AddFile(file string) error {
	var err error
	var commit *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(commit, output, err) }()

	// automatically stage all modified and deleted files and do the commit
	commit = exec.Command(Git, append(r.addFile, file)...)
	commit.Dir = r.projectPath

	// run git command to stage and commit changes
	if output, err = commit.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' failed with %v: %s", commit, err, output)
	}

	return nil
}

// CommitChanges Stage and commit changes in the repository with a specific message.
func (r *repository) CommitChanges(message string) error {
	var err error
	var commit *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(commit, output, err) }()

	// automatically stage all modified and deleted files and do the commit
	commit = exec.Command(Git, append(r.commitAll, fmt.Sprintf("%v", message))...)
	commit.Dir = r.projectPath

	// run git command to stage and commit changes
	if output, err = commit.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' failed with %v: %s", commit, err, output)
	}

	return nil
}

// TagCommit Tag the latest commit in the repository with a specific tag name.
func (r *repository) TagCommit(tagName string) error {
	var err error
	var tag *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(tag, output, err) }()

	// tag the latest commit with the specific tag name
	tag = exec.Command(Git, append(r.tagCommit, tagName)...)
	tag.Dir = r.projectPath

	// run git command to tag the latest commit
	if output, err = tag.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' failed with %v: %s", tag, err, output)
	}

	return nil
}

// PushChanges Push changes in a branch to the remote repository.
func (r *repository) PushChanges(branchName string) error {
	var err error
	var push *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(push, output, err) }()

	// push changes to the remote repository
	push = exec.Command(Git, append(r.pushBranch, branchName)...)
	push.Dir = r.projectPath

	// run git command to push changes
	if output, err = push.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' failed with %v: %s", push, err, output)
	}

	return nil
}

// PushAllChanges Push all local changes in the repository to the remote repository.
func (r *repository) PushAllChanges() error {
	var err error
	var push *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(push, output, err) }()

	// push all changes to the remote repository
	push = exec.Command(Git, r.pushAll...)
	push.Dir = r.projectPath

	// run git command to push all changes
	if output, err = push.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' failed with %v: %s", push, err, output)
	}

	return nil
}

// PushAllTags Push all local tags in the repository to the remote repository.
func (r *repository) PushAllTags() error {
	var err error
	var push *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(push, output, err) }()

	// push all tags to the remote repository
	push = exec.Command(Git, r.pushTags...)
	push.Dir = r.projectPath

	// run git command to push all tags
	if output, err = push.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' failed with %v: %s", push, err, output)
	}

	return nil
}

// PushDeletion Push a local branch deletion in the repository to the remote repository.
func (r *repository) PushDeletion(branchName string) error {
	var err error
	var push *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(push, output, err) }()

	// push the branch deletion to the remote repository
	push = exec.Command(Git, append(r.pushDeletion, branchName)...)
	push.Dir = r.projectPath

	// run git command to push the branch deletion
	if output, err = push.CombinedOutput(); err != nil {
		return fmt.Errorf("git '%v' failed with %v: %s", push, err, output)
	}

	return nil
}

// UndoAllChanges Undo all local changes in the repository and synchronize with the remote repository.
func (r *repository) UndoAllChanges(cause error) error {
	var logs []any = make([]any, 0)

	// log human-readable description of the git command
	defer func() { Log(logs...) }()

	// just return the cause if undo changes is disabled
	if !undoChanges {
		return cause
	}

	// fetch and prune all remote branches
	fetch := exec.Command(Git, r.fetchAll...)
	fetch.Dir = r.projectPath

	// clean all files and directories in the working directory
	clean := exec.Command(Git, r.cleanAll...)
	clean.Dir = r.projectPath

	// checkout the production branch
	checkout := exec.Command(Git, append(r.switchBranch, Production.String())...)
	checkout.Dir = r.projectPath

	// reset the production branch to the remote production branch
	reset := exec.Command(Git, append(r.resetBranch, fmt.Sprintf("%v/%v", r.remote, Production))...)
	reset.Dir = r.projectPath

	// list all locals of the repository
	all := exec.Command(Git, r.allLocals...)
	all.Dir = r.projectPath

	// run git command to fetch all branches
	if output, err := fetch.CombinedOutput(); err != nil {
		logs = append(logs, fetch, output, err)
		return errors.Join(cause, fmt.Errorf("fetching all branches failed with %v: %s", err, output))
	} else {
		logs = append(logs, fetch, output)
	}

	// run git command to check out branch
	if output, err := checkout.CombinedOutput(); err != nil {
		logs = append(logs, checkout, output, err)
		return errors.Join(cause, fmt.Errorf("git '%v' '%v' failed with %v: %s", checkout, Production, err, output))
	} else {
		logs = append(logs, checkout, output)
	}

	// run git command to reset branch
	if output, err := reset.CombinedOutput(); err != nil {
		logs = append(logs, reset, output, err)
		return errors.Join(cause, fmt.Errorf("resetting production branch failed with %v: %s", err, output))
	} else {
		logs = append(logs, reset, output)
	}

	// run git command to clean all files and directories
	if output, err := clean.CombinedOutput(); err != nil {
		logs = append(logs, clean, output, err)
		return errors.Join(cause, fmt.Errorf("cleaning all files and directories failed with %v: %s", err, output))
	} else {
		logs = append(logs, clean, output)
	}

	// run git command to list all locals
	if output, err := all.CombinedOutput(); err != nil {
		logs = append(logs, all, output, err)
		return errors.Join(cause, fmt.Errorf("getting all locals failed with %v: %s", err, output))
	} else {
		logs = append(logs, all, output)

		// check every line of the output for the branch name
		for _, local := range strings.Split(string(output), "\n") {
			local = strings.Trim(local, "* \n\r")

			// check if the local branch is not the production branch
			if len(local) > 0 && local != Production.String() {
				// force-delete the local branch
				delete := exec.Command(Git, append(r.forceDeleteBranch, local)...)
				delete.Dir = r.projectPath

				// run git command to delete the local branch
				if output, err := delete.CombinedOutput(); err != nil {
					logs = append(logs, delete, output, err)
					return errors.Join(cause, fmt.Errorf("deleting local branch '%v' failed with %v: %s", local, err, output))
				} else {
					logs = append(logs, delete, output)
				}
			}
		}
	}

	// always return the original cause if no error occurred
	return cause
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

func releaseStart(repo Repository, p Plugin, major, minor bool) error {
	// check if the repository already has a release branch
	if found, _, err := repo.HasBranch(Release); err != nil {
		return err
	} else if found {
		return fmt.Errorf(
			"repository already has a '%v' branch and only one '%v' branch is allowed at a time",
			Release, Release)
	}

	// check if the repository has a develop branch // todo: has remote branch?
	if found, _, err := repo.HasBranch(Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to start a new '%v' branch from",
			Development, Release)
	}

	// checkout develop branch
	if err := repo.CheckoutBranch(Development.String()); err != nil {
		return err
	}

	// read out the current and next project version ${major}.${minor}.${increment}-${qualifier}
	current, next, err := p.Version(repo.Local(), major, minor, false)

	if err != nil {
		return err
	}

	// if --major Flag only
	//   set the version of project to (${major}+1).0.0-${qualifier}
	//   perform a git commit with a commit message
	if next.VersionIncrement == Major {
		if err := p.UpdateProjectVersion(next.AddQualifier(p.SnapshotQualifier())); err != nil {
			return repo.UndoAllChanges(err)
		}

		if err := repo.CommitChanges("Set next major project version."); err != nil {
			return repo.UndoAllChanges(err)
		}

		current = next
	}

	// create branch release/x.y.z based on the current develop branch without qualifier
	// checkout release/x.y.z branch
	if err := repo.CreateBranch(current.RemoveQualifier().BranchName(Release)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// remove qualifier from the project version (change POM file)
	if err := p.UpdateProjectVersion(current.RemoveQualifier()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repo.CommitChanges("Remove qualifier from project version."); err != nil {
		return repo.UndoAllChanges(err)
	}

	// todo: add hook here
	// todo: content for mvn plugin
	/**
	// execute https://www.mojohaus.org/versions/versions-maven-plugin/use-releases-mojo.html
	if err := p.updateProjectObjectModelReleases(repo.Local()); err != nil {
		return repo.UndoAllChanges(err)
	}
	*/

	// if not clean: perform a git commit with a commit message because the previous step changed the POM file
	if err := repo.IsClean(); err != nil {
		if err := repo.CommitChanges("Update project dependencies with corresponding releases."); err != nil {
			return repo.UndoAllChanges(err)
		}
	}

	// checkout production branch (just for consistency that commands always end on production branch)
	if err := repo.CheckoutBranch(Production.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repo.PushAllChanges(); err != nil {
		return err
	}

	return nil
}

func hotfixStart(repo Repository, p Plugin) error {
	// check if the repository already has a hotfix branch
	if found, _, err := repo.HasBranch(Hotfix); err != nil {
		return err
	} else if found {
		return fmt.Errorf(
			"repository already has a '%v' branch and only one '%v' branch is allowed at a time",
			Hotfix, Hotfix)
	}

	// checkout production branch
	if err := repo.CheckoutBranch(Production.String()); err != nil {
		return err
	}

	// read out the current and next project version ${major}.${minor}.${increment}-${qualifier}
	_, next, err := p.Version(repo.Local(), false, false, true)

	if err != nil {
		return err
	}

	// create branch hotfix/${major}.${minor}.${increment + 1} based on the current production branch
	// checkout hotfix/${major}.${minor}.${increment + 1} branch
	if err := repo.CreateBranch(next.BranchName(Hotfix)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// update project version to ${major}.${minor}.${increment + 1}
	if err := p.UpdateProjectVersion(next); err != nil {
		return repo.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repo.CommitChanges("Set next hotfix version."); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout production branch (just for consistency that commands always end on production branch)
	if err := repo.CheckoutBranch(Production.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repo.PushAllChanges(); err != nil {
		return err
	}

	return nil
}

// todo: p Plugin - all values are nil, why?
// todo: rename Plugin in Plugin
// Run the release finish command for the standard workflow.
func releaseFinish(repo Repository, p Plugin) error {
	var releaseVersion Version

	// check if the repository has a suitable release branch
	if found, remotes, err := repo.HasBranch(Release); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("repository does not have a '%v' branch to finish", Release)
	} else if len(remotes) > 1 {
		return fmt.Errorf("repository must not have multiple '%v' branches", Release)
	} else if version, err := ParseVersion(remotes[0]); err != nil {
		return err
	} else {
		releaseVersion = version
	}

	// check if the repository has a develop branch
	if found, _, err := repo.HasBranch(Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to finish and merge with a '%v' branch",
			Development, Release)
	}

	// checkout release branch
	if err := repo.CheckoutBranch(releaseVersion.BranchName(Release)); err != nil {
		return err
	}

	// checkout production branch
	if err := repo.CheckoutBranch(Production.String()); err != nil {
		return err
	}

	// merge release branch into current production branch (with merge commit --no-ff git flag)
	if err := repo.MergeBranch(releaseVersion.BranchName(Release), NoFastForward); err != nil {
		return repo.UndoAllChanges(err)
	}

	// tag last commit with the release version number
	if err := repo.TagCommit(releaseVersion.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout develop branch
	if err := repo.CheckoutBranch(Development.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// merge release branch into current develop branch (with merge commit --no-ff git flag)
	if err := repo.MergeBranch(releaseVersion.BranchName(Release), NoFastForward); err != nil {
		return repo.UndoAllChanges(err)
	}

	// set project version to the next develop version ${major}.(${minor}+1).0-${qualifier} (change POM file)
	if _, next, err := p.Version(repo.Local(), false, true, false); err != nil {
		return repo.UndoAllChanges(err)
	} else if err := p.UpdateProjectVersion(next.AddQualifier(p.SnapshotQualifier())); err != nil {
		return repo.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repo.CommitChanges("Set next minor project version."); err != nil {
		return repo.UndoAllChanges(err)
	}

	// delete the release branch locally
	if err := repo.DeleteBranch(releaseVersion.BranchName(Release)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout production branch (just for consistency that commands always end on production branch)
	if err := repo.CheckoutBranch(Production.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repo.PushAllChanges(); err != nil {
		return err
	}

	// push all tags to remotes
	if err := repo.PushAllTags(); err != nil {
		return err
	}

	// delete the release branch remotely
	if err := repo.PushDeletion(releaseVersion.BranchName(Release)); err != nil {
		return err
	}

	return nil
}

// todo: p Plugin - alle werte sind nil, warum?
// Run the release finish command for the standard workflow.
func hotfixFinish(repo Repository, p Plugin) error {
	var hotfixVersion Version

	// check if the repository has a suitable hotfix branch
	if found, remotes, err := repo.HasBranch(Hotfix); err != nil {
		return err
	} else if !found {
		return fmt.Errorf("repository does not have a '%v' branch to finish", Hotfix)
	} else if len(remotes) > 1 {
		return fmt.Errorf("repository must not have multiple '%v' branches", Hotfix)
	} else if version, err := ParseVersion(remotes[0]); err != nil {
		return err
	} else {
		hotfixVersion = version
	}

	// check if the repository has a develop branch
	if found, _, err := repo.HasBranch(Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to finish and merge with a '%v' branch",
			Development, Hotfix)
	}

	// checkout hotfix branch
	if err := repo.CheckoutBranch(hotfixVersion.BranchName(Hotfix)); err != nil {
		return err
	}

	// checkout production branch
	if err := repo.CheckoutBranch(Production.String()); err != nil {
		return err
	}

	// merge hotfix branch into current production branch (with merge commit --no-ff git flag)
	if err := repo.MergeBranch(hotfixVersion.BranchName(Hotfix), NoFastForward); err != nil {
		return repo.UndoAllChanges(err)
	}

	// tag last commit with the hotfix version number
	if err := repo.TagCommit(hotfixVersion.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout develop branch
	if err := repo.CheckoutBranch(Development.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// in order to avoid merge conflicts, set and commit pom.xml project version in develop branch equal
	// with current hotfix branch and remember its commit hash (or find a better solution)
	if currentVersion, _, err := p.Version(repo.Local(), false, false, false); err != nil {
		return repo.UndoAllChanges(err)
	} else {
		// update project version to ${major}.${minor}.${increment + 1} (means: hotfix branch version)
		if err := p.UpdateProjectVersion(hotfixVersion); err != nil {
			return repo.UndoAllChanges(err)
		}

		// perform a git commit with a commit message
		if err := repo.CommitChanges("Set hotfix version to avoid merge conflict."); err != nil {
			return repo.UndoAllChanges(err)
		}

		// merge hotfix branch into current develop branch (with merge commit --no-ff git flag)
		if err := repo.MergeBranch(hotfixVersion.BranchName(Hotfix), NoFastForward); err != nil {
			return repo.UndoAllChanges(err)
		}

		// remove previous commit with remembered commit hash, since it was committed just in order
		// to avoid merge conflicts (or find a better solution)
		// change version im develop Branch to the previous snapshot version (or find a better solution)
		// but at the end the project version in develop branch should remain the same as before hotfix merge
		if err := p.UpdateProjectVersion(currentVersion); err != nil {
			return repo.UndoAllChanges(err)
		}

		// perform a git commit with a commit message
		if err := repo.CommitChanges("Set version back to project version before hotfix merge."); err != nil {
			return repo.UndoAllChanges(err)
		}
	}

	// delete the release branch locally
	if err := repo.DeleteBranch(hotfixVersion.BranchName(Hotfix)); err != nil {
		return repo.UndoAllChanges(err)
	}

	// checkout production branch (just for consistency that commands always end on production branch)
	if err := repo.CheckoutBranch(Production.String()); err != nil {
		return repo.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repo.PushAllChanges(); err != nil {
		return err
	}

	// push all tags to remotes
	if err := repo.PushAllTags(); err != nil {
		return err
	}

	// delete the hotfix branch remotely
	if err := repo.PushDeletion(hotfixVersion.BranchName(Hotfix)); err != nil {
		return err
	}

	return nil
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

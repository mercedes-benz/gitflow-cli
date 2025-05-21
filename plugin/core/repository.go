/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type (
	// Repository represents a version control system repository.
	Repository interface {
		Local() string
		IsClean() error
		HasBranch(branch Branch) (bool, []string, error)
		CheckoutBranch(branchName string) error
		CheckoutFile(fileName string) error
		HasConflicts() bool
		ContinueMerge() error
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

// Local Return the local path of the repository.
func (r *repository) Local() string {
	return r.projectPath
}

func (r *repository) HasConflicts() bool {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = r.projectPath
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	return len(output) > 0
}

func (r *repository) CheckoutFile(fileName string) error {
	var err error
	var checkout *exec.Cmd
	var output []byte

	// log human-readable description of the git command
	defer func() { Log(checkout, output, err) }()

	checkout = exec.Command(Git, "checkout", "--ours", fileName)
	checkout.Dir = r.projectPath

	if output, err = checkout.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout file '%v' failed with %v: %s", fileName, err, output)
	}

	return nil
}

func (r *repository) ContinueMerge() error {
	cmd := exec.Command("git", "commit", "--no-edit")
	cmd.Dir = r.projectPath

	return cmd.Run()
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

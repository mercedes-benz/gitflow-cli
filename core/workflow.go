/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"fmt"
	"os"
)

// Start executes the first plugin that meets the precondition.
func Start(branch Branch, projectPath string) error {
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
		if CheckVersionFile(plugin.VersionFileName()) {
			return executePluginStart(plugin, branch, projectPath)
		}
	}
	// execute fallback plugin
	return executePluginStart(fallbackPlugin, branch, projectPath)
}

func executePluginStart(plugin Plugin, branch Branch, projectPath string) error {
	// get access to the local version control system
	repository := NewRepository(projectPath, Remote)

	// check if required tools are available
	if err := ValidateToolsAvailability(plugin.RequiredTools()...); err != nil {
		return err
	}

	// check if the repository prerequisites are met
	if err := repository.IsClean(); err != nil {
		return err
	}

	// format start command messages
	prefix := fmt.Sprintf("%v Plugin Start on branch", plugin.String())
	called := fmt.Sprintf("%v %v called: %v", prefix, branch.String(), repository.Local())
	completed := fmt.Sprintf("%v %v completed: %v", prefix, branch, repository.Local())
	failed := fmt.Sprintf("%v %v failed: %v", prefix, branch, repository.Local())

	switch branch {
	case Release:
		fmt.Println(called)

		// run the release start command
		if err := releaseStart(plugin, repository); err != nil {
			fmt.Println(failed)
			return err
		}

		fmt.Println(completed)
		return nil

	case Hotfix:
		fmt.Println(called)

		// run the hotfix start command
		if err := hotfixStart(plugin, repository); err != nil {
			fmt.Println(failed)
			return err
		}

		fmt.Println(completed)
		return nil

	default:
		return fmt.Errorf("unsupported branch: %v", branch)
	}
}

// Finish executes the first plugin that meets the precondition.
func Finish(branch Branch, projectPath string) error {

	pluginRegistryLock.Lock()
	defer pluginRegistryLock.Unlock()

	// apply suitable settings from the global configuration to the core package
	applySettings()

	// set path to execute plugin detection and workflow commands
	ProjectPath = projectPath

	// check if project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path '%v' does not exist", projectPath)
	}

	// execute the first plugin that meets the precondition
	for _, plugin := range pluginRegistry {
		if CheckVersionFile(plugin.VersionFileName()) {
			return executePluginFinish(plugin, branch, projectPath)
		}
	}
	// execute fallback plugin
	return executePluginFinish(fallbackPlugin, branch, projectPath)
}

func executePluginFinish(plugin Plugin, branch Branch, projectPath string) error {
	// finish the workflow with the selected release business logic
	repository := NewRepository(projectPath, Remote)

	// check if required tools are available
	if err := ValidateToolsAvailability(plugin.RequiredTools()...); err != nil {
		return err
	}

	// check if the repository prerequisites are met
	if err := repository.IsClean(); err != nil {
		return err
	}

	// format finish command messages
	prefix := fmt.Sprintf("%v Plugin Finish on branch", plugin.String())
	called := fmt.Sprintf("%v %v called: %v", prefix, branch.String(), repository.Local())
	completed := fmt.Sprintf("%v %v completed: %v", prefix, branch, repository.Local())
	failed := fmt.Sprintf("%v %v failed: %v", prefix, branch, repository.Local())

	fmt.Println(called)

	// select suitable business logic for the branch
	switch branch {
	case Release:

		// run the release finish command
		if err := releaseFinish(plugin, repository); err != nil {
			fmt.Println(failed)
			return err
		}

		fmt.Println(completed)
		return nil

	case Hotfix:

		// run the hotfix finish command
		if err := hotfixFinish(plugin, repository); err != nil {
			fmt.Println(failed)
			return err
		}

		fmt.Println(completed)
		return nil

	default:
		return fmt.Errorf("unsupported branch: %v", branch)
	}
}

func releaseStart(plugin Plugin, repository Repository) error {

	// check if the repository already has a release branch
	if found, _, err := repository.HasBranch(Release); err != nil {
		return err
	} else if found {
		return fmt.Errorf(
			"repository already has a '%v' branch and only one '%v' branch is allowed at a time",
			Release, Release)
	}

	// check if the repository has a develop branch
	if found, _, err := repository.HasBranch(Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to start a new '%v' branch from",
			Development, Release)
	}

	// checkout develop branch
	if err := repository.CheckoutBranch(Development.String()); err != nil {
		return err
	}

	if err := GlobalHooks.ExecuteHook(plugin, ReleaseStartHooks.BeforeReleaseStartHook, repository); err != nil {
		return repository.UndoAllChanges(err)
	}

	// read out the current project version
	current, err := plugin.ReadVersion(repository)
	if err != nil {
		return err
	}

	// create branch release/x.y.z based on the current develop branch without qualifier
	// checkout release/x.y.z branch
	if err := repository.CreateBranch(current.RemoveQualifier().BranchName(Release)); err != nil {
		return repository.UndoAllChanges(err)
	}

	// remove qualifier from the project version (change POM file)
	if err := plugin.WriteVersion(repository, current.RemoveQualifier()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repository.CommitChanges("Remove qualifier from project version."); err != nil {
		return repository.UndoAllChanges(err)
	}

	// After update project version hook
	if err := GlobalHooks.ExecuteHook(plugin, ReleaseStartHooks.AfterUpdateProjectVersionHook, repository); err != nil {
		return repository.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repository.PushAllChanges(); err != nil {
		return err
	}

	return nil
}

func hotfixStart(plugin Plugin, repository Repository) error {
	// check if the repository already has a hotfix branch
	if found, _, err := repository.HasBranch(Hotfix); err != nil {
		return err
	} else if found {
		return fmt.Errorf(
			"repository already has a '%v' branch and only one '%v' branch is allowed at a time",
			Hotfix, Hotfix)
	}

	// checkout production branch
	if err := repository.CheckoutBranch(Production.String()); err != nil {
		return err
	}

	if err := GlobalHooks.ExecuteHook(plugin, HotfixStartHooks.BeforeHotfixStartHook, repository); err != nil {
		return repository.UndoAllChanges(err)
	}

	// read out the current project version
	current, err := plugin.ReadVersion(repository)
	if err != nil {
		return err
	}

	// calculate the next incremental version
	next, err := current.Next(Incremental)
	if err != nil {
		return err
	}

	// create branch hotfix/${major}.${minor}.${increment + 1} based on the current production branch
	// checkout hotfix/${major}.${minor}.${increment + 1} branch
	if err := repository.CreateBranch(next.BranchName(Hotfix)); err != nil {
		return repository.UndoAllChanges(err)
	}

	// update project version to ${major}.${minor}.${increment + 1}
	if err := plugin.WriteVersion(repository, next); err != nil {
		return repository.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repository.CommitChanges("Increment patch version for hotfix."); err != nil {
		return repository.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repository.PushAllChanges(); err != nil {
		return err
	}

	return nil
}

// Run the release finish command for the standard workflow.
func releaseFinish(plugin Plugin, repository Repository) error {
	var releaseVersion Version

	// check if the repository has a suitable release branch
	if found, remotes, err := repository.HasBranch(Release); err != nil {
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
	if found, _, err := repository.HasBranch(Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to finish and merge with a '%v' branch",
			Development, Release)
	}

	// checkout release branch
	if err := repository.CheckoutBranch(releaseVersion.BranchName(Release)); err != nil {
		return err
	}

	// checkout production branch
	if err := repository.CheckoutBranch(Production.String()); err != nil {
		return err
	}

	// merge release branch into current production branch (with merge commit --no-ff git flag)
	if err := repository.MergeBranch(releaseVersion.BranchName(Release), NoFastForward); err != nil {
		mergeConflictsMap, err := repository.GetMergeConflicts()

		if err != nil {
			return repository.UndoAllChanges(err)
		}

		if len(mergeConflictsMap) == 1 && len(mergeConflictsMap[plugin.VersionFileName()]) == 1 {

			if err := repository.CheckoutFile(plugin.VersionFileName(), Theirs); err != nil {
				return repository.UndoAllChanges(err)
			}

			if err := repository.AddFile(plugin.VersionFileName()); err != nil {
				return repository.UndoAllChanges(err)
			}

			if err := repository.ContinueMerge(); err != nil {
				return repository.UndoAllChanges(err)
			}
		} else {
			return repository.UndoAllChanges(err)
		}
	}

	// tag last commit with the release version number
	if err := repository.TagCommit(releaseVersion.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// checkout develop branch
	if err := repository.CheckoutBranch(Development.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// merge release branch into current develop branch (with merge commit --no-ff git flag)
	if err := repository.MergeBranch(releaseVersion.BranchName(Release), NoFastForward); err != nil {
		return repository.UndoAllChanges(err)
	}

	// read the current version from the project
	current, err := plugin.ReadVersion(repository)
	if err != nil {
		return repository.UndoAllChanges(err)
	}

	// calculate the next minor version
	next, err := current.Next(Minor)
	if err != nil {
		return repository.UndoAllChanges(err)
	}

	// set project version to the next develop version ${major}.(${minor}+1).0-${qualifier}
	if err := plugin.WriteVersion(repository, next.AddQualifier(plugin.VersionQualifier())); err != nil {
		return repository.UndoAllChanges(err)
	}

	// perform a git commit with a commit message
	if err := repository.CommitChanges("Set next minor project version."); err != nil {
		return repository.UndoAllChanges(err)
	}

	// delete the release branch locally
	if err := repository.DeleteBranch(releaseVersion.BranchName(Release)); err != nil {
		return repository.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repository.PushAllChanges(); err != nil {
		return err
	}

	// push all tags to remotes
	if err := repository.PushAllTags(); err != nil {
		return err
	}

	// delete the release branch remotely
	if err := repository.PushDeletion(releaseVersion.BranchName(Release)); err != nil {
		return err
	}

	return nil
}

// Run the release finish command for the standard workflow.
func hotfixFinish(plugin Plugin, repository Repository) error {
	var hotfixVersion Version

	// check if the repository has a suitable hotfix branch
	if found, remotes, err := repository.HasBranch(Hotfix); err != nil {
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
	if found, _, err := repository.HasBranch(Development); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"repository does not have a '%v' branch to finish and merge with a '%v' branch",
			Development, Hotfix)
	}

	// checkout hotfix branch
	if err := repository.CheckoutBranch(hotfixVersion.BranchName(Hotfix)); err != nil {
		return err
	}

	// checkout production branch
	if err := repository.CheckoutBranch(Production.String()); err != nil {
		return err
	}

	// merge hotfix branch into current production branch (with merge commit --no-ff git flag)
	if err := repository.MergeBranch(hotfixVersion.BranchName(Hotfix), NoFastForward); err != nil {
		return repository.UndoAllChanges(err)
	}

	// tag last commit with the hotfix version number
	if err := repository.TagCommit(hotfixVersion.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// checkout develop branch
	if err := repository.CheckoutBranch(Development.String()); err != nil {
		return repository.UndoAllChanges(err)
	}

	// merge hotfix branch into current develop branch
	if err := repository.MergeBranch(hotfixVersion.BranchName(Hotfix), NoFastForward); err != nil {
		mergeConflictsMap, err := repository.GetMergeConflicts()

		if err != nil {
			return repository.UndoAllChanges(err)
		}

		if len(mergeConflictsMap) == 1 && len(mergeConflictsMap[plugin.VersionFileName()]) == 1 {

			if err := repository.CheckoutFile(plugin.VersionFileName(), Ours); err != nil {
				return repository.UndoAllChanges(err)
			}

			if err := repository.AddFile(plugin.VersionFileName()); err != nil {
				return repository.UndoAllChanges(err)
			}

			if err := repository.ContinueMerge(); err != nil {
				return repository.UndoAllChanges(err)
			}
		} else {
			return repository.UndoAllChanges(err)
		}
	}

	if err := GlobalHooks.ExecuteHook(plugin, HotfixFinishHooks.AfterMergeIntoDevelopmentHook, repository); err != nil {
		return repository.UndoAllChanges(err)
	}

	// delete the release branch locally
	if err := repository.DeleteBranch(hotfixVersion.BranchName(Hotfix)); err != nil {
		return repository.UndoAllChanges(err)
	}

	// push all branches to remotes
	if err := repository.PushAllChanges(); err != nil {
		return err
	}

	// push all tags to remotes
	if err := repository.PushAllTags(); err != nil {
		return err
	}

	// delete the hotfix branch remotely
	if err := repository.PushDeletion(hotfixVersion.BranchName(Hotfix)); err != nil {
		return err
	}

	return nil
}

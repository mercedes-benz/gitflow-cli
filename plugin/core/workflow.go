/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"fmt"
	"os"
	"reflect"
)

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
		if plugin.CheckRequiredFile(projectPath) {
			// get access to the local version control system
			repo := NewRepository(projectPath, Remote)

			// check if required tools are available
			if err := ValidateToolsAvailability(plugin.RequiredTools()...); err != nil {
				return err
			}

			// check if the repository prerequisites are met
			if err := repo.IsClean(); err != nil {
				return err
			}

			// format start command messages
			prefix := fmt.Sprintf("%v Plugin Start on branch", plugin.String())
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

	return nil
}

// Finish executes the first plugin that meets the precondition.
func Finish(branch Branch, projectPath string) error {

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
		if plugin.CheckRequiredFile(projectPath) {

			// finish the workflow with the selected release business logic
			repo := NewRepository(projectPath, Remote)

			// check if required tools are available
			if err := ValidateToolsAvailability(plugin.RequiredTools()...); err != nil {
				return err
			}

			// check if the repository prerequisites are met
			if err := repo.IsClean(); err != nil {
				return err
			}

			// format finish command messages
			prefix := fmt.Sprintf("%v Plugin Finish on branch", plugin.String())
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

func releaseStart(repo Repository, p Plugin, major, minor bool) error {

	if err := p.BeforeReleaseStartHook(); err != nil {
		return err
	}

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

	if err := p.AfterUpdateProjectVersionHook(); err != nil {
		return repo.UndoAllChanges(err)
	}

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

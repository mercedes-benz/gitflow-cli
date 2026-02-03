/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package manager

// VersionManager is the interface that all Python package managers must implement
type VersionManager interface {
	// GetVersion returns the current version string
	GetVersion() (string, error)

	// SetVersion sets the version to the provided string
	SetVersion(version string) error

	// GetName returns the name of the package manager
	GetName() string

	// GetFilePath returns the path to the version file
	GetFilePath() string
}

/*
SPDX-FileCopyrightText: 2024 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Version increment types for the workflow automation commands.
const (
	None VersionIncrement = iota
	Major
	Minor
	Incremental
)

type (
	// VersionIncrement Type of version increment.
	VersionIncrement int

	// Version represents a version-stamp with a major, minor, incremental part, and optionally empty qualifier.
	Version struct {
		VersionIncrement                     VersionIncrement
		Major, Minor, Incremental, Qualifier string
	}
)

// NoVersion is the default version without any parts.
var NoVersion Version

// VersionStamp is the format for version strings.
const versionStamp = "%v.%v.%v"

// VersionStampWithQualifier is the format for version strings with a qualifier.
const versionStampWithQualifier = "%v.%v.%v-%v"

// VersionExpression is the regular expression for version strings with optional qualifier.
const versionExpression = `(\d+)\.(\d+)\.(\d+)(?:-(\w+))?$`

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
func (v Version) Increment(increment VersionIncrement, nextMajor, nextMinor, nextIncremental string) (Version, error) {
	switch increment {
	case Major:
		return NewVersion(nextMajor, "0", "0", v.Qualifier, increment), nil

	case Minor:
		return NewVersion(v.Major, nextMinor, "0", v.Qualifier, increment), nil

	case Incremental:
		return NewVersion(v.Major, v.Minor, nextIncremental, v.Qualifier, increment), nil

	default:
		return NoVersion, fmt.Errorf("unsupported version increment type: %v", increment)
	}
}

// Next Determine the next version based on the current version and the version increment type.
func (v Version) Next(increment VersionIncrement) (Version, error) {
	nextMajor, errMajor := strconv.Atoi(v.Major)
	nextMinor, errMinor := strconv.Atoi(v.Minor)
	nextIncremental, errIncremental := strconv.Atoi(v.Incremental)

	if errMajor != nil || errMinor != nil || errIncremental != nil {
		return NoVersion, errors.Join(fmt.Errorf("invalid version parts: %v", v), errMajor, errMinor, errIncremental)
	}

	nextMajor++
	nextMinor++
	nextIncremental++
	return v.Increment(increment, strconv.Itoa(nextMajor), strconv.Itoa(nextMinor), strconv.Itoa(nextIncremental))
}

// AddQualifier Add a qualifier to the version.
func (v Version) AddQualifier(qualifier string) Version {
	return NewVersion(v.Major, v.Minor, v.Incremental, qualifier, v.VersionIncrement)
}

// RemoveQualifier Remove the qualifier from the version.
func (v Version) RemoveQualifier() Version {
	return NewVersion(v.Major, v.Minor, v.Incremental, noQualifier, v.VersionIncrement)
}

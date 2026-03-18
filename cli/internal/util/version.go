package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseVersion parses a semantic version string and returns the major, minor, and patch numbers.
// It accepts versions with or without a 'v' prefix, and handles versions with only major
// or major.minor components. Pre-release suffixes (e.g., -beta) and build metadata (+build)
// are stripped but don't cause errors.
func ParseVersion(version string) (major, minor, patch int, err error) {
	if version == "" {
		return 0, 0, 0, fmt.Errorf("empty version string")
	}

	// Strip 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Strip pre-release and build metadata
	// e.g., "1.2.3-beta+build" -> "1.2.3"
	if idx := strings.IndexAny(version, "-+"); idx != -1 {
		version = version[:idx]
	}

	// Validate that version contains only digits and dots
	validPattern := regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)
	if !validPattern.MatchString(version) {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	parts := strings.Split(version, ".")

	// Parse major
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	// Parse minor if present
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
		}
	}

	// Parse patch if present
	if len(parts) > 2 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return major, minor, patch, nil
}

// CompareVersions compares two semantic version strings.
// Returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
//
// If either version is invalid, returns 0.
func CompareVersions(v1, v2 string) int {
	major1, minor1, patch1, err1 := ParseVersion(v1)
	major2, minor2, patch2, err2 := ParseVersion(v2)

	if err1 != nil || err2 != nil {
		return 0
	}

	// Compare major
	if major1 > major2 {
		return 1
	}
	if major1 < major2 {
		return -1
	}

	// Compare minor
	if minor1 > minor2 {
		return 1
	}
	if minor1 < minor2 {
		return -1
	}

	// Compare patch
	if patch1 > patch2 {
		return 1
	}
	if patch1 < patch2 {
		return -1
	}

	return 0
}

// IsNewer returns true if v1 is newer (greater) than v2.
func IsNewer(v1, v2 string) bool {
	return CompareVersions(v1, v2) > 0
}

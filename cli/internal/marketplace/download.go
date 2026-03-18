// Package marketplace provides functionality for downloading and managing
// vibe releases from GitHub.
package marketplace

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/wmsimpson/claude-vibe/cli/internal/util"
)

// Release represents a GitHub release.
type Release struct {
	Tag         string
	Name        string
	PublishedAt time.Time
	IsDraft     bool
	IsPrerelease bool
	IsLatest    bool
}

// DefaultRepo is the default GitHub repository for vibe releases.
const DefaultRepo = "will-simpson/claude-vibe"

// ListReleases fetches available releases from GitHub using the gh CLI.
// Returns releases sorted by date (newest first).
func ListReleases(repo string) ([]Release, error) {
	if repo == "" {
		repo = DefaultRepo
	}

	// Use gh CLI to list releases.
	// Always specify --hostname github.com to avoid Databricks GHE interception.
	stdout, stderr, err := util.RunCommand("gh", "release", "list",
		"--repo", repo,
		"--limit", "20",
		"--hostname", "github.com",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w\n%s", err, stderr)
	}

	return parseReleaseList(stdout)
}

// parseReleaseList parses the output of `gh release list`.
// Format: TITLE\tTYPE\tTAG\tPUBLISHED_DATE
func parseReleaseList(output string) ([]Release, error) {
	var releases []Release
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			continue
		}

		title := fields[0]
		releaseType := fields[1]
		tag := fields[2]
		dateStr := fields[3]

		// Parse the date
		publishedAt, err := time.Parse("2006-01-02T15:04:05Z", dateStr)
		if err != nil {
			// Try alternative format
			publishedAt, err = time.Parse("Jan 2, 2006", dateStr)
			if err != nil {
				publishedAt = time.Now() // Fallback
			}
		}

		release := Release{
			Tag:          tag,
			Name:         title,
			PublishedAt:  publishedAt,
			IsDraft:      releaseType == "Draft",
			IsPrerelease: releaseType == "Pre-release",
			IsLatest:     i == 0 && releaseType == "Latest",
		}

		// Check for "Latest" tag
		if strings.Contains(releaseType, "Latest") {
			release.IsLatest = true
		}

		releases = append(releases, release)
	}

	return releases, nil
}

// DownloadRelease downloads a specific release tarball to the destination directory.
// Returns the path to the downloaded file.
func DownloadRelease(repo, tag, destDir string) (string, error) {
	if repo == "" {
		repo = DefaultRepo
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Determine download filename
	filename := fmt.Sprintf("vibe-%s.tar.gz", tag)
	destPath := filepath.Join(destDir, filename)

	// Download using gh CLI.
	// Always specify --hostname github.com to avoid Databricks GHE interception.
	args := []string{
		"release", "download", tag,
		"--repo", repo,
		"--archive=tar.gz",
		"--output", destPath,
		"--hostname", "github.com",
	}

	_, stderr, err := util.RunCommand("gh", args...)
	if err != nil {
		return "", fmt.Errorf("failed to download release %s: %w\n%s", tag, err, stderr)
	}

	return destPath, nil
}

// DownloadMainBranch downloads the latest source tarball from the main branch.
// Returns the path to the downloaded file.
func DownloadMainBranch(repo, destDir string) (string, error) {
	if repo == "" {
		repo = DefaultRepo
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	destPath := filepath.Join(destDir, "vibe-main.tar.gz")

	// Use gh api to download the tarball of the main branch.
	// gh api writes binary response to stdout, so capture it to a file.
	// Always specify --hostname github.com to avoid Databricks GHE interception.
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/tarball/main", repo),
		"--hostname", "github.com",
	)
	outFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	cmd.Stdout = outFile

	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	err = cmd.Run()
	outFile.Close()
	if err != nil {
		os.Remove(destPath)
		return "", fmt.Errorf("failed to download main branch: %w\n%s", err, errBuf.String())
	}

	return destPath, nil
}

// ExtractTarball extracts a .tar.gz file to the destination directory.
// Returns the path to the extracted directory (typically the first directory in the archive).
func ExtractTarball(tarballPath, destDir string) (string, error) {
	file, err := os.Open(tarballPath)
	if err != nil {
		return "", fmt.Errorf("failed to open tarball: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var extractedDir string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tarball: %w", err)
		}

		// Get the target path
		target := filepath.Join(destDir, header.Name)

		// Validate path (prevent directory traversal)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return "", fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
			// Track the first top-level directory
			if extractedDir == "" {
				parts := strings.Split(header.Name, "/")
				if len(parts) > 0 && parts[0] != "" {
					extractedDir = filepath.Join(destDir, parts[0])
				}
			}

		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", fmt.Errorf("failed to create parent directory: %w", err)
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return "", fmt.Errorf("failed to create file: %w", err)
			}

			// Limit copy size to prevent decompression bombs
			if _, err := io.Copy(outFile, io.LimitReader(tr, 100*1024*1024)); err != nil {
				outFile.Close()
				return "", fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()

		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Ignore symlink errors on some platforms
				continue
			}
		}
	}

	return extractedDir, nil
}

// isVersionTag returns true if the tag looks like a CLI version tag (e.g., "v1.0.16").
func isVersionTag(tag string) bool {
	re := regexp.MustCompile(`^v?\d+\.\d+\.\d+`)
	return re.MatchString(tag)
}

// GetLatestRelease returns the latest non-draft, non-prerelease CLI release.
// Only releases with version-like tags (e.g., "v1.0.16") are considered.
func GetLatestRelease(repo string) (*Release, error) {
	releases, err := ListReleases(repo)
	if err != nil {
		return nil, err
	}

	for _, r := range releases {
		if !r.IsDraft && !r.IsPrerelease && isVersionTag(r.Tag) {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("no CLI releases found")
}

// CompareVersions compares two version strings.
// Returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Clean up version strings (remove 'v' prefix if present)
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Handle special cases
	if v1 == v2 {
		return 0
	}
	if v1 == "dev" || v1 == "unknown" {
		return -1
	}
	if v2 == "dev" || v2 == "unknown" {
		return 1
	}

	// Parse version components
	v1Parts := parseVersion(v1)
	v2Parts := parseVersion(v2)

	// Compare each component
	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(v1Parts) {
			p1 = v1Parts[i]
		}
		if i < len(v2Parts) {
			p2 = v2Parts[i]
		}

		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}

	return 0
}

// parseVersion parses a version string into numeric components.
func parseVersion(v string) []int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(v, -1)

	parts := make([]int, len(matches))
	for i, m := range matches {
		var n int
		fmt.Sscanf(m, "%d", &n)
		parts[i] = n
	}

	return parts
}

// MarketplacePath returns the path to the vibe marketplace directory.
func MarketplacePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vibe", "marketplace")
}

// CopyDir recursively copies a directory tree.
func CopyDir(src, dst string) error {
	// Get source info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a single file.
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Release namespace for release-related targets
type Release mg.Namespace

// Build builds release binaries for all platforms using goreleaser
func (Release) Build() error {
	mg.Deps(Util.Check)
	fmt.Println("Building release binaries with goreleaser...")
	return sh.Run("goreleaser", "build", "--clean", "--single-target=false")
}

// BuildSingle builds release binary for current platform only
func (Release) BuildSingle() error {
	mg.Deps(Util.Check)
	fmt.Println("Building release binary for current platform...")
	return sh.Run("goreleaser", "build", "--clean", "--single-target=true")
}

// Release creates a full release using goreleaser
func (Release) Release() error {
	mg.Deps(Util.Check)
	fmt.Println("Creating release with goreleaser...")

	// Check if we're on a tag
	if tag, err := sh.Output("git", "describe", "--tags", "--exact-match", "HEAD"); err == nil && tag != "" {
		fmt.Printf("Creating release for tag: %s\n", strings.TrimSpace(tag))
		return sh.Run("goreleaser", "release", "--clean")
	} else {
		fmt.Println("Warning: Not on a git tag. Releases should be created from tagged commits.")
		fmt.Println("To create a release, tag your commit first:")
		fmt.Println("  git tag -a v1.2.3 -m 'Release v1.2.3'")
		fmt.Println("  git push origin v1.2.3")
		return fmt.Errorf("not on a git tag - cannot create release")
	}
}

// Snapshot creates a snapshot release (pre-release) using goreleaser
func (Release) Snapshot() error {
	mg.Deps(Util.Check)
	fmt.Println("Creating snapshot release with goreleaser...")
	return sh.Run("goreleaser", "release", "--clean", "--snapshot", "--skip-publish")
}

// CheckPrerelease verifies that the environment is ready for a release
func (Release) CheckPrerelease() error {
	fmt.Println("Checking release prerequisites...")

	// Check if goreleaser is installed
	if _, err := sh.Output("which", "goreleaser"); err != nil {
		return fmt.Errorf("goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest")
	}

	// Check git status
	if status, err := sh.Output("git", "status", "--porcelain"); err != nil {
		return fmt.Errorf("failed to check git status: %v", err)
	} else if status != "" {
		fmt.Println("Warning: Working directory is not clean, this may affect the release:")
		fmt.Println(status)
	}

	// Check if on a tagged commit
	if tag, err := sh.Output("git", "describe", "--tags", "--exact-match", "HEAD"); err == nil && tag != "" {
		fmt.Printf("✓ Ready to release from tag: %s\n", strings.TrimSpace(tag))
	} else {
		fmt.Println("Note: Not on a git tag. Use 'git tag -a v1.2.3 -m \"Release v1.2.3\"' to create a tag.")
	}

	// Check for GitHub token if publishing
	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Println("Warning: GITHUB_TOKEN not set. Release will not be published to GitHub.")
	}

	return nil
}

// DryRun performs a dry run of the release process
func (Release) DryRun() error {
	mg.Deps(Util.Check)
	fmt.Println("Running release dry run with goreleaser...")
	return sh.Run("goreleaser", "release", "--clean", "--snapshot")
}

// Manifest generates release manifest without publishing
func (Release) Manifest() error {
	mg.Deps(Util.Check)
	fmt.Println("Generating release manifest...")
	return sh.Run("goreleaser", "release", "--clean", "--skip-publish")
}

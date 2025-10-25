package main

import (
	"fmt"
	"os"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Build namespace for build-related targets
type Build mg.Namespace

// Build builds the binary for current platform
func (Build) Build() error {
	mg.Deps(Build.Deps)
	fmt.Println("Building binary...")

	ldflags := fmt.Sprintf("-s -w -X main.version=%s -X main.commit=%s -X main.date=%s",
		getVersion(), getCommit(), time.Now().Format(time.RFC3339))

	return sh.Run("go", "build", "-ldflags", ldflags, "-o", binaryName, packagePath)
}

// Deps downloads and installs dependencies
func (Build) Deps() error {
	fmt.Println("Installing dependencies...")
	return sh.Run("go", "mod", "download")
}

// Install installs the binary to $GOPATH/bin with version information
func (Build) Install() error {
	mg.Deps(Build.Deps)
	fmt.Println("Installing binary with version information...")

	ldflags := fmt.Sprintf("-s -w -X main.version=%s -X main.commit=%s -X main.date=%s",
		getVersion(), getCommit(), time.Now().Format(time.RFC3339))

	return sh.Run("go", "install", "-ldflags", ldflags, packagePath)
}

// Run builds and runs the application with help
func (Build) Run() error {
	mg.Deps(Build.Build)
	fmt.Println("Running application...")
	return sh.Run("./"+binaryName, "--help")
}

// Clean removes build artifacts
func (Build) Clean() error {
	fmt.Println("Cleaning build artifacts...")

	artifacts := []string{
		binaryName,
		binaryName + ".exe",
		"bin/",
		"dist/",
		"coverage/coverage.out",
		"coverage/coverage.html",
		"coverage.out",  // Clean up any old files in root
		"coverage.html", // Clean up any old files in root
	}

	for _, artifact := range artifacts {
		if err := os.RemoveAll(artifact); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	// Clean Go cache
	sh.Run("go", "clean", "-cache")
	sh.Run("go", "clean", "-testcache")

	fmt.Println("Clean completed!")
	return nil
}

// Version shows version information
func (Build) Version() error {
	fmt.Printf("Version: %s\n", getVersion())
	fmt.Printf("Commit: %s\n", getCommit())
	fmt.Printf("Date: %s\n", time.Now().Format(time.RFC3339))
	return nil
}

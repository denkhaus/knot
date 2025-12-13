package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindWorkspaceRoot(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()

	// Test case 1: Workspace root with .knot directory
	knotWorkspace := filepath.Join(tempDir, "workspace1")
	require.NoError(t, os.MkdirAll(filepath.Join(knotWorkspace, ".knot"), 0o700))

	subdir := filepath.Join(knotWorkspace, "subdir", "deep")
	require.NoError(t, os.MkdirAll(subdir, 0o700))

	// Change to subdirectory and test finding workspace root
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	require.NoError(t, os.Chdir(subdir))

	workspaceRoot, err := FindWorkspaceRoot()
	require.NoError(t, err)
	require.Equal(t, knotWorkspace, workspaceRoot)

	// Test case 2: Workspace root with .git directory (no .knot)
	gitWorkspace := filepath.Join(tempDir, "workspace2")
	require.NoError(t, os.MkdirAll(filepath.Join(gitWorkspace, ".git"), 0o700))

	subdir2 := filepath.Join(gitWorkspace, "src", "package")
	require.NoError(t, os.MkdirAll(subdir2, 0o700))

	require.NoError(t, os.Chdir(subdir2))

	workspaceRoot, err = FindWorkspaceRoot()
	require.NoError(t, err)
	require.Equal(t, gitWorkspace, workspaceRoot)

	// Test case 3: No workspace root found
	noWorkspaceDir := filepath.Join(tempDir, "no-workspace")
	require.NoError(t, os.MkdirAll(filepath.Join(noWorkspaceDir, "deep", "nested"), 0o700))

	require.NoError(t, os.Chdir(filepath.Join(noWorkspaceDir, "deep", "nested")))

	_, err = FindWorkspaceRoot()
	require.Error(t, err)
	require.Contains(t, err.Error(), "workspace root not found")

	// Test case 4: Prefer .knot over .git when both exist
	bothWorkspace := filepath.Join(tempDir, "workspace3")
	require.NoError(t, os.MkdirAll(filepath.Join(bothWorkspace, ".knot"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(bothWorkspace, ".git"), 0o700))

	subdir3 := filepath.Join(bothWorkspace, "subdir")
	require.NoError(t, os.MkdirAll(subdir3, 0o700))

	require.NoError(t, os.Chdir(subdir3))

	workspaceRoot, err = FindWorkspaceRoot()
	require.NoError(t, err)
	require.Equal(t, bothWorkspace, workspaceRoot)
}

func TestGetWorkspaceProjectDir(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Test case 1: .knot directory exists in workspace root
	workspace := filepath.Join(tempDir, "test-workspace")
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, ".knot"), 0o700))

	subdir := filepath.Join(workspace, "subdir")
	require.NoError(t, os.MkdirAll(subdir, 0o700))

	require.NoError(t, os.Chdir(subdir))

	projectDir, err := GetWorkspaceProjectDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(workspace, ".knot"), projectDir)

	// Test case 2: .knot directory doesn't exist but .git does
	workspace2 := filepath.Join(tempDir, "test-workspace2")
	require.NoError(t, os.MkdirAll(filepath.Join(workspace2, ".git"), 0o700))

	subdir2 := filepath.Join(workspace2, "src")
	require.NoError(t, os.MkdirAll(subdir2, 0o700))

	require.NoError(t, os.Chdir(subdir2))

	projectDir, err = GetWorkspaceProjectDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(workspace2, ".knot"), projectDir)
}

func TestEnsureWorkspaceProjectDir(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Test case 1: Create .knot directory in workspace root with .git
	workspace := filepath.Join(tempDir, "test-workspace")
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, ".git"), 0o700))

	subdir := filepath.Join(workspace, "subdir")
	require.NoError(t, os.MkdirAll(subdir, 0o700))

	require.NoError(t, os.Chdir(subdir))

	projectDir, err := EnsureWorkspaceProjectDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(workspace, ".knot"), projectDir)

	// Verify directory was created with correct permissions
	info, err := os.Stat(projectDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
	require.Equal(t, os.FileMode(0o700), info.Mode().Perm())

	// Test case 2: Directory already exists
	projectDir2, err := EnsureWorkspaceProjectDir()
	require.NoError(t, err)
	require.Equal(t, projectDir, projectDir2)
}

func TestGetWorkspaceDatabasePath(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Test case 1: Database path in workspace root
	workspace := filepath.Join(tempDir, "test-workspace")
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, ".knot"), 0o700))

	subdir := filepath.Join(workspace, "subdir")
	require.NoError(t, os.MkdirAll(subdir, 0o700))

	require.NoError(t, os.Chdir(subdir))

	dbPath, err := GetWorkspaceDatabasePath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(workspace, ".knot", "knot.db"), dbPath)
}

func TestGetWorkspaceSQLiteConnectionString(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Test case 1: Connection string for workspace database
	workspace := filepath.Join(tempDir, "test-workspace")
	require.NoError(t, os.MkdirAll(filepath.Join(workspace, ".knot"), 0o700))

	subdir := filepath.Join(workspace, "subdir")
	require.NoError(t, os.MkdirAll(subdir, 0o700))

	require.NoError(t, os.Chdir(subdir))

	connStr, err := GetWorkspaceSQLiteConnectionString()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(workspace, ".knot", "knot.db"), connStr)
}

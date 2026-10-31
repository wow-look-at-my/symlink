package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSymlink_TargetDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	nonexistent := filepath.Join(tempDir, "nonexistent")
	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, nonexistent, false, false, true)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "target does not exist")

}

func TestCreateSymlink_TypeMismatch_ExpectFileGotDir(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "targetdir")
	require.NoError(t, os.Mkdir(targetDir, 0755))

	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetDir, false, false, true)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "is a directory, not a file")

}

func TestCreateSymlink_TypeMismatch_ExpectDirGotFile(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "targetfile")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetFile, false, false, false)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "is a file, not a directory")

}

func TestCreateSymlink_ExistsWithoutForce(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	symlinkPath := filepath.Join(tempDir, "mylink")
	require.NoError(t, os.Symlink(targetFile, symlinkPath))

	err := createSymlink(symlinkPath, targetFile, false, false, true)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "already exists")

}

func TestCreateSymlink_ExistsWithForce(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	oldTarget := filepath.Join(tempDir, "oldtarget")
	require.NoError(t, os.WriteFile(oldTarget, []byte("old"), 0644))

	symlinkPath := filepath.Join(tempDir, "mylink")
	require.NoError(t, os.Symlink(oldTarget, symlinkPath))

	err := createSymlink(symlinkPath, targetFile, true, false, true)
	require.Nil(t, err)

	// Verify symlink was replaced
	linkDest, err := os.Readlink(symlinkPath)
	require.Nil(t, err)

	absTarget, _ := filepath.Abs(targetFile)
	assert.Equal(t, absTarget, linkDest)

}

func TestCreateSymlink_CannotOverwriteRegularFile(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	existingFile := filepath.Join(tempDir, "existingfile")
	require.NoError(t, os.WriteFile(existingFile, []byte("data"), 0644))

	err := createSymlink(existingFile, targetFile, true, false, true)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "not a symlink")

}

func TestCreateSymlink_ParentDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	symlinkPath := filepath.Join(tempDir, "nonexistent", "mylink")

	err := createSymlink(symlinkPath, targetFile, false, false, true)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "parent directory does not exist")

}

func TestCreateSymlink_ParentIsFile(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	parentFile := filepath.Join(tempDir, "parentfile")
	require.NoError(t, os.WriteFile(parentFile, []byte("data"), 0644))

	symlinkPath := filepath.Join(parentFile, "mylink")

	err := createSymlink(symlinkPath, targetFile, false, false, true)
	require.NotNil(t, err)

	// The error occurs when trying to lstat the symlink path (parent is a file, not a dir)
	assert.Contains(t, err.Error(), "cannot check symlink path")

}

func TestCreateSymlink_AbsolutePath_File(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetFile, false, false, true)
	require.Nil(t, err)

	linkDest, err := os.Readlink(symlinkPath)
	require.Nil(t, err)

	absTarget, _ := filepath.Abs(targetFile)
	assert.Equal(t, absTarget, linkDest)

	// Verify symlink works
	content, err := os.ReadFile(symlinkPath)
	require.Nil(t, err)

	assert.Equal(t, "content", string(content))

}

func TestCreateSymlink_AbsolutePath_Folder(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "targetdir")
	require.NoError(t, os.Mkdir(targetDir, 0755))

	testFile := filepath.Join(targetDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetDir, false, false, false)
	require.Nil(t, err)

	// Verify symlink works
	content, err := os.ReadFile(filepath.Join(symlinkPath, "test.txt"))
	require.Nil(t, err)

	assert.Equal(t, "hello", string(content))

}

func TestCreateSymlink_RelativePath_File(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetFile, false, true, true)
	require.Nil(t, err)

	linkDest, err := os.Readlink(symlinkPath)
	require.Nil(t, err)

	// Should be relative path
	assert.False(t, filepath.IsAbs(linkDest))

	assert.Equal(t, "target", linkDest)

	// Verify symlink works
	content, err := os.ReadFile(symlinkPath)
	require.Nil(t, err)

	assert.Equal(t, "content", string(content))

}

func TestCreateSymlink_RelativePath_Folder(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "targetdir")
	require.NoError(t, os.Mkdir(targetDir, 0755))

	testFile := filepath.Join(targetDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetDir, false, true, false)
	require.Nil(t, err)

	linkDest, err := os.Readlink(symlinkPath)
	require.Nil(t, err)

	assert.False(t, filepath.IsAbs(linkDest))

	assert.Equal(t, "targetdir", linkDest)

}

func TestCreateSymlink_RelativePath_DifferentDirectories(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "subdir", "target")
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	linkDir := filepath.Join(tempDir, "linkdir")
	require.NoError(t, os.Mkdir(linkDir, 0755))

	symlinkPath := filepath.Join(linkDir, "mylink")

	err := createSymlink(symlinkPath, targetDir, false, true, false)
	require.Nil(t, err)

	linkDest, err := os.Readlink(symlinkPath)
	require.Nil(t, err)

	assert.False(t, filepath.IsAbs(linkDest))

	// Should be something like ../subdir/target
	assert.True(t, strings.HasPrefix(linkDest, ".."))

	// Verify symlink works
	info, err := os.Stat(symlinkPath)
	require.Nil(t, err)

	assert.True(t, info.IsDir())

}

func TestCreateSymlink_SymlinkInCurrentDir(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	require.NoError(t, os.WriteFile(targetFile, []byte("content"), 0644))

	// Change to temp dir to test "." parent case
	oldWd, err := os.Getwd()
	require.Nil(t, err)

	defer os.Chdir(oldWd)
	require.NoError(t, os.Chdir(tempDir))

	err = createSymlink("mylink", "target", false, true, true)
	require.Nil(t, err)

	content, err := os.ReadFile("mylink")
	require.Nil(t, err)

	assert.Equal(t, "content", string(content))

}

func TestCLI_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.Nil(t, err)

}

func TestCLI_InvalidType(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("x"), 0644)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"invalid", "absolute", "--new-symlink", "/tmp/a", "--existing-target", target})
	err := cmd.Execute()
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "first argument must be 'file' or 'folder'")

}

func TestCLI_InvalidPathMode(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("x"), 0644)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "invalid", "--new-symlink", "/tmp/a", "--existing-target", target})
	err := cmd.Execute()
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "second argument must be 'relative' or 'absolute'")

}

func TestCLI_MissingNewSymlink(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("x"), 0644)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--existing-target", target})
	err := cmd.Execute()
	require.NotNil(t, err)

}

func TestCLI_MissingExistingTarget(t *testing.T) {
	tempDir := t.TempDir()
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--new-symlink", symlink})
	err := cmd.Execute()
	require.NotNil(t, err)

}

func TestCLI_Success(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--new-symlink", symlink, "--existing-target", target})
	err := cmd.Execute()
	require.Nil(t, err)

	// Verify symlink was created
	linkDest, err := os.Readlink(symlink)
	require.Nil(t, err)

	absTarget, _ := filepath.Abs(target)
	assert.Equal(t, absTarget, linkDest)

}

func TestCLI_SuccessWithForce(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")
	os.Symlink(target, symlink)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--new-symlink", symlink, "--existing-target", target, "--force"})
	err := cmd.Execute()
	require.Nil(t, err)

}

func TestCLI_SuccessRelative(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "relative", "--new-symlink", symlink, "--existing-target", target})
	err := cmd.Execute()
	require.Nil(t, err)

	linkDest, err := os.Readlink(symlink)
	require.Nil(t, err)

	assert.False(t, filepath.IsAbs(linkDest))

}

func TestCLI_SuccessFolder(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "targetdir")
	os.Mkdir(target, 0755)
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"folder", "absolute", "--new-symlink", symlink, "--existing-target", target})
	err := cmd.Execute()
	require.Nil(t, err)

}

func TestCLI_TargetDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--new-symlink", symlink, "--existing-target", "/nonexistent/path"})
	err := cmd.Execute()
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "target does not exist")

}

func TestCLI_WrongArgCount(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"file"})
	err := cmd.Execute()
	require.NotNil(t, err)

}

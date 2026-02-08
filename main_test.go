package main

import (
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
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}
	if !strings.Contains(err.Error(), "target does not exist") {
		t.Errorf("expected 'target does not exist' error, got: %v", err)
	}
}

func TestCreateSymlink_TypeMismatch_ExpectFileGotDir(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "targetdir")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetDir, false, false, true)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
	if !strings.Contains(err.Error(), "is a directory, not a file") {
		t.Errorf("expected type mismatch error, got: %v", err)
	}
}

func TestCreateSymlink_TypeMismatch_ExpectDirGotFile(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "targetfile")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetFile, false, false, false)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
	if !strings.Contains(err.Error(), "is a file, not a directory") {
		t.Errorf("expected type mismatch error, got: %v", err)
	}
}

func TestCreateSymlink_ExistsWithoutForce(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "mylink")
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Fatal(err)
	}

	err := createSymlink(symlinkPath, targetFile, false, false, true)
	if err == nil {
		t.Fatal("expected error when symlink exists without --force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestCreateSymlink_ExistsWithForce(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	oldTarget := filepath.Join(tempDir, "oldtarget")
	if err := os.WriteFile(oldTarget, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "mylink")
	if err := os.Symlink(oldTarget, symlinkPath); err != nil {
		t.Fatal(err)
	}

	err := createSymlink(symlinkPath, targetFile, true, false, true)
	if err != nil {
		t.Fatalf("expected success with --force, got: %v", err)
	}

	// Verify symlink was replaced
	linkDest, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatal(err)
	}
	absTarget, _ := filepath.Abs(targetFile)
	if linkDest != absTarget {
		t.Errorf("expected symlink to point to %s, got %s", absTarget, linkDest)
	}
}

func TestCreateSymlink_CannotOverwriteRegularFile(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	existingFile := filepath.Join(tempDir, "existingfile")
	if err := os.WriteFile(existingFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	err := createSymlink(existingFile, targetFile, true, false, true)
	if err == nil {
		t.Fatal("expected error when trying to overwrite regular file")
	}
	if !strings.Contains(err.Error(), "not a symlink") {
		t.Errorf("expected 'not a symlink' error, got: %v", err)
	}
}

func TestCreateSymlink_ParentDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "nonexistent", "mylink")

	err := createSymlink(symlinkPath, targetFile, false, false, true)
	if err == nil {
		t.Fatal("expected error when parent directory doesn't exist")
	}
	if !strings.Contains(err.Error(), "parent directory does not exist") {
		t.Errorf("expected 'parent directory does not exist' error, got: %v", err)
	}
}

func TestCreateSymlink_ParentIsFile(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	parentFile := filepath.Join(tempDir, "parentfile")
	if err := os.WriteFile(parentFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(parentFile, "mylink")

	err := createSymlink(symlinkPath, targetFile, false, false, true)
	if err == nil {
		t.Fatal("expected error when parent is a file")
	}
	// The error occurs when trying to lstat the symlink path (parent is a file, not a dir)
	if !strings.Contains(err.Error(), "cannot check symlink path") {
		t.Errorf("expected 'cannot check symlink path' error, got: %v", err)
	}
}

func TestCreateSymlink_AbsolutePath_File(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetFile, false, false, true)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	linkDest, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatal(err)
	}
	absTarget, _ := filepath.Abs(targetFile)
	if linkDest != absTarget {
		t.Errorf("expected absolute path %s, got %s", absTarget, linkDest)
	}

	// Verify symlink works
	content, err := os.ReadFile(symlinkPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}
}

func TestCreateSymlink_AbsolutePath_Folder(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "targetdir")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(targetDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetDir, false, false, false)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	// Verify symlink works
	content, err := os.ReadFile(filepath.Join(symlinkPath, "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello" {
		t.Errorf("expected 'hello', got %q", string(content))
	}
}

func TestCreateSymlink_RelativePath_File(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetFile, false, true, true)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	linkDest, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatal(err)
	}
	// Should be relative path
	if filepath.IsAbs(linkDest) {
		t.Errorf("expected relative path, got absolute: %s", linkDest)
	}
	if linkDest != "target" {
		t.Errorf("expected 'target', got %s", linkDest)
	}

	// Verify symlink works
	content, err := os.ReadFile(symlinkPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}
}

func TestCreateSymlink_RelativePath_Folder(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "targetdir")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(targetDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(tempDir, "mylink")

	err := createSymlink(symlinkPath, targetDir, false, true, false)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	linkDest, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(linkDest) {
		t.Errorf("expected relative path, got absolute: %s", linkDest)
	}
	if linkDest != "targetdir" {
		t.Errorf("expected 'targetdir', got %s", linkDest)
	}
}

func TestCreateSymlink_RelativePath_DifferentDirectories(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "subdir", "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(tempDir, "linkdir")
	if err := os.Mkdir(linkDir, 0755); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(linkDir, "mylink")

	err := createSymlink(symlinkPath, targetDir, false, true, false)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	linkDest, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(linkDest) {
		t.Errorf("expected relative path, got absolute: %s", linkDest)
	}
	// Should be something like ../subdir/target
	if !strings.HasPrefix(linkDest, "..") {
		t.Errorf("expected path starting with '..', got %s", linkDest)
	}

	// Verify symlink works
	info, err := os.Stat(symlinkPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected symlink to resolve to directory")
	}
}

func TestCreateSymlink_SymlinkInCurrentDir(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir to test "." parent case
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	err = createSymlink("mylink", "target", false, true, true)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	content, err := os.ReadFile("mylink")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}
}

func TestCLI_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("help should not error: %v", err)
	}
}

func TestCLI_InvalidType(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("x"), 0644)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"invalid", "absolute", "--new-symlink", "/tmp/a", "--existing-target", target})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "first argument must be 'file' or 'folder'") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_InvalidPathMode(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("x"), 0644)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "invalid", "--new-symlink", "/tmp/a", "--existing-target", target})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid path mode")
	}
	if !strings.Contains(err.Error(), "second argument must be 'relative' or 'absolute'") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_MissingNewSymlink(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("x"), 0644)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--existing-target", target})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --new-symlink")
	}
}

func TestCLI_MissingExistingTarget(t *testing.T) {
	tempDir := t.TempDir()
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--new-symlink", symlink})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --existing-target")
	}
}

func TestCLI_Success(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--new-symlink", symlink, "--existing-target", target})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	// Verify symlink was created
	linkDest, err := os.Readlink(symlink)
	if err != nil {
		t.Fatal(err)
	}
	absTarget, _ := filepath.Abs(target)
	if linkDest != absTarget {
		t.Errorf("expected %s, got %s", absTarget, linkDest)
	}
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
	if err != nil {
		t.Fatalf("expected success with --force, got: %v", err)
	}
}

func TestCLI_SuccessRelative(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "relative", "--new-symlink", symlink, "--existing-target", target})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	linkDest, err := os.Readlink(symlink)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(linkDest) {
		t.Errorf("expected relative path, got: %s", linkDest)
	}
}

func TestCLI_SuccessFolder(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "targetdir")
	os.Mkdir(target, 0755)
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"folder", "absolute", "--new-symlink", symlink, "--existing-target", target})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestCLI_TargetDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	symlink := filepath.Join(tempDir, "link")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"file", "absolute", "--new-symlink", symlink, "--existing-target", "/nonexistent/path"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}
	if !strings.Contains(err.Error(), "target does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_WrongArgCount(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"file"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}


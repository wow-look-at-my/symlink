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

func TestRun_Help(t *testing.T) {
	r, w, _ := os.Pipe()
	err := run([]string{"--help"}, w, os.Stderr)
	w.Close()

	if err != errHelp {
		t.Errorf("expected errHelp, got: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	if !strings.Contains(output, "symlink - create symlinks without the ln footguns") {
		t.Error("help output missing header")
	}
}

func TestRun_HelpShort(t *testing.T) {
	r, w, _ := os.Pipe()
	err := run([]string{"-h"}, w, os.Stderr)
	w.Close()

	if err != errHelp {
		t.Errorf("expected errHelp, got: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	if !strings.Contains(output, "symlink - create symlinks without the ln footguns") {
		t.Error("help output missing header")
	}
}

func TestRun_NoArgs(t *testing.T) {
	r, w, _ := os.Pipe()
	err := run([]string{}, w, os.Stderr)
	w.Close()

	if err != errUsage {
		t.Errorf("expected errUsage, got: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	if !strings.Contains(output, "symlink - create symlinks without the ln footguns") {
		t.Error("usage output missing header")
	}
}

func TestRun_InvalidType(t *testing.T) {
	err := run([]string{"invalid", "absolute"}, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "first argument must be 'file' or 'folder'") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_InvalidPathMode(t *testing.T) {
	err := run([]string{"file", "invalid"}, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for invalid path mode")
	}
	if !strings.Contains(err.Error(), "second argument must be 'relative' or 'absolute'") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_MissingNewSymlink(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("x"), 0644)

	err := run([]string{"file", "absolute", "-e", target}, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for missing --new-symlink")
	}
	if !strings.Contains(err.Error(), "--new-symlink (-n) is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_MissingExistingTarget(t *testing.T) {
	tempDir := t.TempDir()
	symlink := filepath.Join(tempDir, "link")

	err := run([]string{"file", "absolute", "-n", symlink}, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for missing --existing-target")
	}
	if !strings.Contains(err.Error(), "--existing-target (-e) is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_Success(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")

	r, w, _ := os.Pipe()
	err := run([]string{"file", "absolute", "-n", symlink, "-e", target}, w, os.Stderr)
	w.Close()

	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	if !strings.Contains(output, "created symlink") {
		t.Error("expected success message")
	}
}

func TestRun_SuccessLongFlags(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")

	err := run([]string{"file", "absolute", "--new-symlink", symlink, "--existing-target", target}, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestRun_SuccessWithForce(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")
	os.Symlink(target, symlink)

	err := run([]string{"file", "absolute", "-n", symlink, "-e", target, "--force"}, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("expected success with --force, got: %v", err)
	}
}

func TestRun_SuccessRelative(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target")
	os.WriteFile(target, []byte("content"), 0644)
	symlink := filepath.Join(tempDir, "link")

	err := run([]string{"file", "relative", "-n", symlink, "-e", target}, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestRun_SuccessFolder(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "targetdir")
	os.Mkdir(target, 0755)
	symlink := filepath.Join(tempDir, "link")

	err := run([]string{"folder", "absolute", "-n", symlink, "-e", target}, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestRun_TargetDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	symlink := filepath.Join(tempDir, "link")

	err := run([]string{"file", "absolute", "-n", symlink, "-e", "/nonexistent/path"}, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}
	if !strings.Contains(err.Error(), "target does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_InvalidFlag(t *testing.T) {
	err := run([]string{"file", "absolute", "--invalid-flag"}, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for invalid flag")
	}
}

func TestRun_HelpInMiddle(t *testing.T) {
	r, w, _ := os.Pipe()
	err := run([]string{"file", "--help", "absolute"}, w, os.Stderr)
	w.Close()

	if err != errHelp {
		t.Errorf("expected errHelp, got: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	if !strings.Contains(output, "symlink - create symlinks without the ln footguns") {
		t.Error("help output missing header")
	}
}

func TestRun_OneArg(t *testing.T) {
	r, w, _ := os.Pipe()
	err := run([]string{"file"}, w, os.Stderr)
	w.Close()

	if err != errUsage {
		t.Errorf("expected errUsage, got: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	if !strings.Contains(output, "symlink - create symlinks without the ln footguns") {
		t.Error("usage output missing header")
	}
}

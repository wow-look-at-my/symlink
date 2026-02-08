package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const usageText = `symlink - create symlinks without the ln footguns

Usage:
  symlink <type> <pathmode> --new-symlink <path> --existing-target <path> [--force]

Arguments:
  type      file or folder - what kind of target you're linking to
  pathmode  relative or absolute - how the symlink stores the path

Flags:
  -n, --new-symlink      path where the symlink will be created (required)
  -e, --existing-target  path to the existing file/folder to link to (required)
  -f, --force            overwrite if symlink already exists (cannot overwrite files/folders)

Examples:
  symlink file absolute -n /tmp/mylink -e /home/user/file.txt
  symlink folder relative -n ./link -e ../target-dir
  symlink file absolute -n /tmp/mylink -e /tmp/file.txt --force
`

// errHelp is returned when --help is requested (exit 0)
var errHelp = fmt.Errorf("help requested")

// errUsage is returned when usage should be shown due to bad args (exit 1)
var errUsage = fmt.Errorf("usage error")

func run(args []string, stdout, stderr *os.File) error {
	// Handle --help or -h anywhere in args
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, usageText)
			return errHelp
		}
	}

	// Check we have at least 2 positional args
	if len(args) < 2 {
		fmt.Fprint(stdout, usageText)
		return errUsage
	}

	linkType := args[0]
	pathMode := args[1]

	// Validate positional args
	if linkType != "file" && linkType != "folder" {
		return fmt.Errorf("first argument must be 'file' or 'folder', got '%s'", linkType)
	}
	if pathMode != "relative" && pathMode != "absolute" {
		return fmt.Errorf("second argument must be 'relative' or 'absolute', got '%s'", pathMode)
	}

	// Parse flags from remaining args
	fs := flag.NewFlagSet("symlink", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stdout, usageText) }

	var newSymlink, existingTarget string
	var force bool

	fs.StringVar(&newSymlink, "new-symlink", "", "path where the symlink will be created")
	fs.StringVar(&newSymlink, "n", "", "path where the symlink will be created (shorthand)")
	fs.StringVar(&existingTarget, "existing-target", "", "path to the existing file/folder to link to")
	fs.StringVar(&existingTarget, "e", "", "path to the existing file/folder to link to (shorthand)")
	fs.BoolVar(&force, "force", false, "overwrite if symlink already exists")
	fs.BoolVar(&force, "f", false, "overwrite if symlink already exists (shorthand)")

	if err := fs.Parse(args[2:]); err != nil {
		return err
	}

	// Validate required flags
	if newSymlink == "" {
		return fmt.Errorf("--new-symlink (-n) is required")
	}
	if existingTarget == "" {
		return fmt.Errorf("--existing-target (-e) is required")
	}

	isRelative := pathMode == "relative"
	expectFile := linkType == "file"

	if err := createSymlink(newSymlink, existingTarget, force, isRelative, expectFile); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "created symlink: %s -> %s\n", newSymlink, existingTarget)
	return nil
}

func main() {
	err := run(os.Args[1:], os.Stdout, os.Stderr)
	if err == nil {
		return
	}
	if err == errHelp {
		return
	}
	if err == errUsage {
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func createSymlink(newSymlink, existingTarget string, force, isRelative, expectFile bool) error {
	// 1. Check that existing target exists
	targetInfo, err := os.Stat(existingTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target does not exist: %s", existingTarget)
		}
		return fmt.Errorf("cannot access target: %v", err)
	}

	// 2. Validate target type matches command
	if expectFile && targetInfo.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", existingTarget)
	}
	if !expectFile && !targetInfo.IsDir() {
		return fmt.Errorf("%s is a file, not a directory", existingTarget)
	}

	// 3. Check if symlink location already exists
	symlinkInfo, err := os.Lstat(newSymlink)
	if err == nil {
		// Something exists at the symlink path
		isSymlink := symlinkInfo.Mode()&os.ModeSymlink != 0

		if !force {
			return fmt.Errorf("%s already exists, use --force to overwrite", newSymlink)
		}

		if !isSymlink {
			return fmt.Errorf("cannot overwrite %s: not a symlink", newSymlink)
		}

		// It's a symlink and --force is set, remove it
		if err := os.Remove(newSymlink); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot check symlink path: %v", err)
	}

	// 4. Check parent directory exists
	parentDir := filepath.Dir(newSymlink)
	if parentDir != "" && parentDir != "." {
		parentInfo, err := os.Stat(parentDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("parent directory does not exist: %s", parentDir)
			}
			return fmt.Errorf("cannot access parent directory: %v", err)
		}
		if !parentInfo.IsDir() {
			return fmt.Errorf("parent path is not a directory: %s", parentDir)
		}
	}

	// 5. Compute the target path for the symlink
	var target string
	if isRelative {
		// Compute relative path from symlink's parent directory to target
		symlinkDir := filepath.Dir(newSymlink)
		absSymlinkDir, err := filepath.Abs(symlinkDir)
		if err != nil {
			return fmt.Errorf("cannot resolve symlink directory: %v", err)
		}
		absTarget, err := filepath.Abs(existingTarget)
		if err != nil {
			return fmt.Errorf("cannot resolve target path: %v", err)
		}
		target, err = filepath.Rel(absSymlinkDir, absTarget)
		if err != nil {
			return fmt.Errorf("cannot compute relative path: %v", err)
		}
	} else {
		// Use absolute path
		target, err = filepath.Abs(existingTarget)
		if err != nil {
			return fmt.Errorf("cannot resolve absolute path: %v", err)
		}
	}

	// 6. Create the symlink
	if err := os.Symlink(target, newSymlink); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	return nil
}

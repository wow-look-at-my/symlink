package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	var newSymlink, existingTarget string
	var force bool

	cmd := &cobra.Command{
		Use:   "symlink <file|folder> <relative|absolute> --new-symlink <path> --existing-target <path>",
		Short: "Create symlinks without the ln footguns",
		Long: `symlink - create symlinks without the ln footguns

Arguments:
  type      file or folder - what kind of target you're linking to
  pathmode  relative or absolute - how the symlink stores the path

Flags:
  --new-symlink      path where the symlink will be created (required)
  --existing-target  path to the existing file/folder to link to (required)
  --force            overwrite if symlink already exists (cannot overwrite files/folders)

Examples:
  symlink file absolute --new-symlink /tmp/mylink --existing-target /home/user/file.txt
  symlink folder relative --new-symlink ./link --existing-target ../target-dir
  symlink file absolute --new-symlink /tmp/mylink --existing-target /tmp/file.txt --force`,
		Args:          cobra.ExactArgs(2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			linkType := args[0]
			pathMode := args[1]

			if linkType != "file" && linkType != "folder" {
				return fmt.Errorf("first argument must be 'file' or 'folder', got '%s'", linkType)
			}
			if pathMode != "relative" && pathMode != "absolute" {
				return fmt.Errorf("second argument must be 'relative' or 'absolute', got '%s'", pathMode)
			}

			isRelative := pathMode == "relative"
			expectFile := linkType == "file"

			if err := createSymlink(newSymlink, existingTarget, force, isRelative, expectFile); err != nil {
				return err
			}

			fmt.Printf("created symlink: %s -> %s\n", newSymlink, existingTarget)
			return nil
		},
	}

	cmd.Flags().StringVar(&newSymlink, "new-symlink", "", "path where the symlink will be created")
	cmd.Flags().StringVar(&existingTarget, "existing-target", "", "path to the existing file/folder to link to")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite if symlink already exists (cannot overwrite files/folders)")

	cmd.MarkFlagRequired("new-symlink")
	cmd.MarkFlagRequired("existing-target")

	return cmd
}

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
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

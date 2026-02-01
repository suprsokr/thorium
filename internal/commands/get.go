// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"thorium-cli/internal/config"
)

// Get clones a mod from a GitHub repository and installs it into the workspace
func Get(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	modNameOverride := fs.String("name", "", "Custom name for the mod (overrides auto-detected name)")
	updateExisting := fs.Bool("update", false, "Update/overwrite existing mod if it already exists")
	fs.Parse(args)

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: thorium get <github-url> [--name <custom-name>] [--update]\n\nExample: thorium get https://github.com/suprsokr/thorium-custom-packets\nExample: thorium get https://github.com/user/repo --name my-custom-name\nExample: thorium get https://github.com/user/repo --update")
	}

	githubURL := fs.Arg(0)

	// Validate GitHub URL format
	if !strings.Contains(githubURL, "github.com") {
		return fmt.Errorf("invalid GitHub URL: %s", githubURL)
	}

	fmt.Printf("Installing mod from: %s\n", githubURL)

	// Extract repo name from URL
	// Example: https://github.com/suprsokr/thorium-custom-packets -> thorium-custom-packets
	parts := strings.Split(strings.TrimSuffix(githubURL, ".git"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid GitHub URL format")
	}
	repoName := parts[len(parts)-1]

	// Create temporary directory for cloning
	tempDir, err := os.MkdirTemp("", "thorium-get-*")
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone the repository
	fmt.Printf("Cloning repository...\n")
	cloneDir := filepath.Join(tempDir, repoName)
	cmd := exec.Command("git", "clone", "--depth", "1", githubURL, cloneDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Check if the cloned repo has a valid mod structure
	// Look for at least one of: README.md, scripts/, server-patches/, binary-edits/, assets/, luaxml/, dbc_sql/, world_sql/
	modDirs := []string{"scripts", "server-patches", "binary-edits", "assets", "luaxml", "dbc_sql", "world_sql"}
	hasValidStructure := false
	hasReadme := false

	if _, err := os.Stat(filepath.Join(cloneDir, "README.md")); err == nil {
		hasReadme = true
	}

	for _, dir := range modDirs {
		if _, err := os.Stat(filepath.Join(cloneDir, dir)); err == nil {
			hasValidStructure = true
			break
		}
	}

	if !hasValidStructure {
		return fmt.Errorf("repository does not appear to be a valid Thorium mod (missing expected directories)")
	}

	// Determine mod name (use repo name by default)
	modName := repoName
	
	// Apply custom name override if provided
	if *modNameOverride != "" {
		modName = *modNameOverride
	}

	// Check if mod already exists
	modPath := filepath.Join(cfg.GetModsPath(), modName)
	if _, err := os.Stat(modPath); err == nil {
		if !*updateExisting {
			return fmt.Errorf("mod '%s' already exists at: %s\n\nUse --update to overwrite, or --name to install with a different name", modName, modPath)
		}
		
		// Update mode: remove existing mod
		fmt.Printf("Updating existing mod: %s\n", modName)
		if err := os.RemoveAll(modPath); err != nil {
			return fmt.Errorf("remove existing mod: %w", err)
		}
	}

	// Create mods directory if it doesn't exist
	if err := os.MkdirAll(cfg.GetModsPath(), 0755); err != nil {
		return fmt.Errorf("create mods directory: %w", err)
	}

	// Copy the mod to the workspace (excluding .git directory)
	if *updateExisting {
		fmt.Printf("Installing updated mod as: %s\n", modName)
	} else {
		fmt.Printf("Installing mod as: %s\n", modName)
	}
	if err := copyModDir(cloneDir, modPath); err != nil {
		return fmt.Errorf("copy mod files: %w", err)
	}

	if *updateExisting {
		fmt.Printf("\n✓ Mod updated successfully!\n\n")
	} else {
		fmt.Printf("\n✓ Mod installed successfully!\n\n")
	}
	fmt.Printf("Location: %s\n", modPath)
	
	if hasReadme {
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Review the mod documentation: mods/%s/README.md\n", modName)
		fmt.Printf("  2. Build the mod: thorium build\n")
	} else {
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Build the mod: thorium build\n")
	}

	return nil
}

// copyModDir recursively copies a directory, excluding .git
func copyModDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyModFile(path, dstPath, info.Mode())
	})
}

// copyModFile copies a single file with specific permissions
func copyModFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Preserve file permissions
	return os.Chmod(dst, mode)
}

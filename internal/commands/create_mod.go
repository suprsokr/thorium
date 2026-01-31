// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"thorium-cli/internal/config"
)

// CreateMod creates a new mod with the standard directory structure
func CreateMod(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("create-mod", flag.ExitOnError)
	noLuaXML := fs.Bool("no-luaxml", false, "Skip creating luaxml folder and copying source files")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("mod name required. Usage: thorium create-mod <mod-name>")
	}

	modName := remaining[0]

	// Validate mod name
	if err := validateModName(modName); err != nil {
		return err
	}

	modsPath := cfg.GetModsPath()
	modPath := filepath.Join(modsPath, modName)

	// Check if mod already exists
	if _, err := os.Stat(modPath); err == nil {
		return fmt.Errorf("mod already exists: %s", modPath)
	}

	fmt.Printf("Creating mod: %s\n", modName)

	// Create directory structure
	dirs := []string{
		modPath,
		filepath.Join(modPath, "dbc_sql"),
		filepath.Join(modPath, "world_sql"),
		filepath.Join(modPath, "scripts"),
	}

	// Only create luaxml folder if not skipped
	if !*noLuaXML {
		dirs = append(dirs, filepath.Join(modPath, "luaxml"))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
		fmt.Printf("  Created: %s\n", dir)
	}

	// Create .gitkeep files
	gitkeepDirs := []string{
		filepath.Join(modPath, "dbc_sql"),
		filepath.Join(modPath, "world_sql"),
		filepath.Join(modPath, "scripts"),
	}
	if !*noLuaXML {
		gitkeepDirs = append(gitkeepDirs, filepath.Join(modPath, "luaxml"))
	}
	for _, dir := range gitkeepDirs {
		gitkeep := filepath.Join(dir, ".gitkeep")
		os.WriteFile(gitkeep, []byte{}, 0644)
	}

	// Copy luaxml_source files to mod's luaxml folder
	if !*noLuaXML {
		luaxmlSource := cfg.GetLuaXMLSourcePath()
		modLuaXML := filepath.Join(modPath, "luaxml")
		
		if _, err := os.Stat(luaxmlSource); err == nil {
			count, err := copyLuaXMLSource(luaxmlSource, modLuaXML)
			if err != nil {
				fmt.Printf("  Warning: failed to copy luaxml source: %v\n", err)
			} else if count > 0 {
				fmt.Printf("  Copied %d LuaXML source files\n", count)
			}
		}
	}

	// Create README
	readme := fmt.Sprintf(`# %s

A Thorium mod for TrinityCore.

## Structure

- `+"`dbc_sql/`"+` - DBC database migrations
- `+"`world_sql/`"+` - World database migrations
- `+"`scripts/`"+` - TrinityCore ServerScripts
- `+"`luaxml/`"+` - Client-side Lua/XML modifications

## Creating Migrations

Create SQL files with the naming convention:
`+"```"+`
YYYYMMDD_HHMMSS_description.sql
YYYYMMDD_HHMMSS_description.rollback.sql
`+"```"+`

## Creating Scripts

`+"```"+`bash
thorium create-script --mod %s --type spell my_spell
`+"```"+`

## Building

`+"```"+`bash
thorium build --mod %s
`+"```"+`
`, modName, modName, modName)

	readmePath := filepath.Join(modPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("write README: %w", err)
	}
	fmt.Printf("  Created: %s\n", readmePath)

	fmt.Println()
	fmt.Printf("✓ Mod '%s' created successfully!\n", modName)

	return nil
}

// copyLuaXMLSource copies all files from source to destination
func copyLuaXMLSource(srcDir, dstDir string) (int, error) {
	count := 0

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}

		// Create destination path
		dstPath := filepath.Join(dstDir, relPath)

		// Ensure destination directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return err
		}

		count++
		return nil
	})

	return count, err
}

// validateModName checks if a mod name is valid
func validateModName(name string) error {
	if name == "" {
		return fmt.Errorf("mod name cannot be empty")
	}

	valid := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !valid.MatchString(name) {
		return fmt.Errorf("invalid mod name: must start with a letter and contain only letters, numbers, hyphens, and underscores")
	}

	reserved := []string{"shared", "mods", "thorium", "config", "build"}
	for _, r := range reserved {
		if strings.EqualFold(name, r) {
			return fmt.Errorf("invalid mod name: '%s' is reserved", name)
		}
	}

	return nil
}


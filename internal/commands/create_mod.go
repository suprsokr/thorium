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
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: thorium create-mod <mod-name>

Create a new mod with the standard directory structure.

Arguments:
  <mod-name>    Name of the mod to create (required)
                Must start with a letter and contain only letters, numbers, hyphens, and underscores
                Reserved names: shared, mods, thorium, config, build

The mod will be created with the following structure:
  mods/<mod-name>/
    ├── dbc_sql/         DBC database migrations
    ├── world_sql/       World database migrations
    ├── scripts/         TrinityCore C++ scripts
    ├── server-patches/  TrinityCore source patches (.patch files)
    ├── binary-edits/    Client binary patches (.json files)
    ├── assets/          Files to copy to client directory
    └── luaxml/          Client-side Lua/XML modifications

Example:
  thorium create-mod my-custom-mod
`)
	}
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		fs.Usage()
		return fmt.Errorf("mod name required")
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
		filepath.Join(modPath, "server-patches"),
		filepath.Join(modPath, "binary-edits"),
		filepath.Join(modPath, "assets"),
		filepath.Join(modPath, "luaxml"),
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
		filepath.Join(modPath, "server-patches"),
		filepath.Join(modPath, "binary-edits"),
		filepath.Join(modPath, "assets"),
		filepath.Join(modPath, "luaxml"),
	}
	for _, dir := range gitkeepDirs {
		gitkeep := filepath.Join(dir, ".gitkeep")
		os.WriteFile(gitkeep, []byte{}, 0644)
	}

	// Note: luaxml folder starts empty. Use 'thorium create-addon' to add custom addons,
	// or 'thorium extract --mod <mod>' to copy specific interface files to modify.

	// Create README
	readme := fmt.Sprintf(`# %s

A Thorium mod for TrinityCore.

## Structure

- `+"`dbc_sql/`"+` - DBC database migrations
- `+"`world_sql/`"+` - World database migrations
- `+"`scripts/`"+` - TrinityCore C++ scripts
- `+"`server-patches/`"+` - TrinityCore source patches (.patch files)
- `+"`binary-edits/`"+` - Client binary patches (.json files)
- `+"`assets/`"+` - Files to copy to client directory
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

// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"thorium-cli/internal/config"
	"thorium-cli/internal/custompackets"
)

// Init initializes a new Thorium workspace
func Init(configPath string, args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	force := fs.Bool("force", false, "Overwrite existing files")
	fs.Parse(args)

	workspaceRoot := filepath.Dir(configPath)
	if workspaceRoot == "." {
		var err error
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !*force {
		return fmt.Errorf("workspace already initialized (config.json exists). Use --force to reinitialize")
	}

	fmt.Println("Initializing Thorium workspace...")
	fmt.Println()

	// Create directory structure
	dirs := []struct {
		path string
		desc string
	}{
		{"shared", "Shared resources"},
		{"shared/dbc", "DBC files"},
		{"shared/dbc/dbc_source", "Original DBC files (extracted from client)"},
		{"shared/dbc/dbc_out", "Modified DBC files (exported from database)"},
		{"shared/luaxml", "LuaXML files"},
		{"shared/luaxml/luaxml_source", "Original LuaXML files (extracted from client)"},
		{"shared/migrations_applied", "Tracks applied SQL migrations"},
		{"mods", "Your mods go here"},
	}

	for _, d := range dirs {
		dirPath := filepath.Join(workspaceRoot, d.path)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", d.path, err)
		}
		fmt.Printf("  Created: %s/\n", d.path)

		// Add .gitkeep to empty directories
		gitkeep := filepath.Join(dirPath, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			os.WriteFile(gitkeep, []byte{}, 0644)
		}
	}

	// Create config.json
	cfg := config.DefaultConfig()
	if err := config.WriteConfig(cfg, configPath); err != nil {
		return err
	}
	fmt.Printf("  Created: config.json\n")

	// Create .gitignore
	gitignore := `# Thorium workspace
shared/dbc/dbc_out/*.dbc
shared/dbc/dbc_out/*.MPQ
shared/migrations_applied/*
!shared/migrations_applied/.gitkeep

# Build artifacts
*.MPQ

# OS files
.DS_Store
Thumbs.db
`
	gitignorePath := filepath.Join(workspaceRoot, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) || *force {
		if err := os.WriteFile(gitignorePath, []byte(gitignore), 0644); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
		fmt.Printf("  Created: .gitignore\n")
	}

	// Create CustomPackets addon in shared luaxml
	customPacketsDir := filepath.Join(workspaceRoot, "shared", "luaxml", "luaxml_source", "Interface", "AddOns", "CustomPackets")
	if err := os.MkdirAll(customPacketsDir, 0755); err != nil {
		return fmt.Errorf("create CustomPackets directory: %w", err)
	}

	// Generate CustomPackets.lua
	opts := custompackets.DefaultLuaGeneratorOptions()
	opts.OutputPath = filepath.Join(customPacketsDir, "CustomPackets.lua")
	if err := custompackets.GenerateLuaAPI(opts); err != nil {
		return fmt.Errorf("generate CustomPackets.lua: %w", err)
	}

	// Generate TOC file
	tocContent := `## Interface: 30300
## Title: CustomPackets
## Notes: Custom packet API for client-server communication
## Author: Thorium
## Version: 1.0.0

CustomPackets.lua
`
	tocPath := filepath.Join(customPacketsDir, "CustomPackets.toc")
	if err := os.WriteFile(tocPath, []byte(tocContent), 0644); err != nil {
		return fmt.Errorf("write CustomPackets.toc: %w", err)
	}
	fmt.Printf("  Created: shared/luaxml/luaxml_source/Interface/AddOns/CustomPackets/\n")

	// Create README
	readme := `# Thorium Workspace

A TrinityCore modding workspace.

## Structure

` + "```" + `
.
├── config.json           # Workspace configuration
├── shared/
│   ├── dbc/
│   │   ├── dbc_source/   # Original DBCs (extract with: thorium extract --dbc)
│   │   └── dbc_out/      # Modified DBCs (exported from database)
│   ├── luaxml/
│   │   └── luaxml_source/  # Baseline LuaXML (extract with: thorium extract --luaxml)
│   └── migrations_applied/ # Tracks applied migrations
└── mods/
    └── your-mod/
        ├── dbc_sql/      # DBC database migrations
        ├── world_sql/    # World database migrations
        └── luaxml/       # Client UI modifications
` + "```" + `

## Quick Start

` + "```" + `bash
# Configure (edit config.json with your paths)
vim config.json

# Extract original files from WoW client
thorium extract --dbc --luaxml

# Create a new mod
thorium create-mod my-first-mod

# Build all mods
thorium build

# Check status
thorium status
` + "```" + `

## Configuration

Edit ` + "`config.json`" + ` to set:

- ` + "`wotlk.path`" + `: Path to your WoW 3.3.5 client
- ` + "`databases`" + `: MySQL connection settings
- ` + "`server.dbc_path`" + `: Where to copy server DBCs

Environment variables can be used: ` + "`${VAR:-default}`" + `
`
	readmePath := filepath.Join(workspaceRoot, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) || *force {
		if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
			return fmt.Errorf("write README: %w", err)
		}
		fmt.Printf("  Created: README.md\n")
	}

	fmt.Println()
	fmt.Println("✓ Workspace initialized!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit config.json with your paths and database settings")
	fmt.Println("  2. Run: thorium extract --dbc --luaxml")
	fmt.Println("  3. Run: thorium create-mod my-first-mod")
	fmt.Println("  4. Run: thorium build")

	return nil
}

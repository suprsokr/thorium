// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"thorium-cli/internal/config"
	"thorium-cli/internal/database"
)

// Init initializes a new Thorium workspace
func Init(configPath string, args []string) error {
	// Check if this is a subcommand
	if len(args) > 0 && args[0] == "db" {
		return InitDB(configPath, args[1:])
	}

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

	// Create README
	readme := `# Thorium Workspace

A TrinityCore modding workspace created with [Thorium](https://github.com/suprsokr/thorium).

## Structure

` + "```" + `
.
├── config.json              # Workspace configuration
├── .gitignore               # Git ignore rules
├── shared/
│   ├── dbc/
│   │   ├── dbc_source/      # Original DBCs (if using DBC workflow)
│   │   └── dbc_out/         # Modified DBCs (if using DBC workflow)
│   ├── luaxml/
│   │   └── luaxml_source/   # Original LuaXML (if extracted)
│   └── migrations_applied/  # Tracks applied SQL migrations
└── mods/                    # Your mods go here
    └── your-mod/
        ├── dbc_sql/         # DBC database migrations (optional)
        ├── world_sql/       # World database migrations (optional)
        ├── scripts/         # TrinityCore C++ scripts (optional)
        ├── server-patches/  # TrinityCore source patches (optional)
        ├── binary-edits/    # Client binary patches (optional)
        ├── assets/          # Files to copy to client (optional)
        └── luaxml/          # Client UI modifications (optional)
` + "```" + `

## Quick Start

` + "```" + `bash
# 1. Configure paths (or use Peacebloom: https://github.com/suprsokr/peacebloom)
vim config.json

# 2. Create your first mod
thorium create-mod my-first-mod

# 3. Start modding!
# For addons:
thorium create-addon --mod my-first-mod MyAddon

# For World database changes:
thorium create-migration --mod my-first-mod --db world "add custom npc"

# For DBC changes (requires DBC workflow setup - see below):
thorium create-migration --mod my-first-mod --db dbc "custom spell"

# 4. Build and package
thorium build

# 5. Check status
thorium status

# 6. Create a distributable package
thorium dist --mod my-first-mod
` + "```" + `

## DBC Workflow (Optional)

Only set this up if your mod modifies client-side data (spells, items, creatures, zones, etc.).

Many mods (like custom addons or server-only changes) don't need this!

` + "```" + `bash
# One-command setup
thorium init db

# This automatically:
# - Creates the dbc and dbc_source databases
# - Extracts DBCs from your WoW client
# - Imports DBCs to dbc_source
# - Copies dbc_source to dbc
# (Note: world database is managed by TrinityCore)

# Now DBC migrations will work
thorium build dbc_sql
` + "```" + `

### Why Two DBC Databases?

- **dbc_source**: Pristine baseline, never modified. Used for creating clean distribution packages.
- **dbc**: Development database where your migrations are applied and tested.

During distribution, Thorium temporarily applies migrations to ` + "`dbc_source`" + `, exports only the changed DBCs, then rolls back to keep it pristine.

## Configuration

The ` + "`config.json`" + ` file uses environment variables with fallback defaults:

- ` + "`wotlk.path`" + `: Path to your WoW 3.3.5 client (default: ` + "`${WOTLK_PATH:-/wotlk}`" + `)
- ` + "`wotlk.locale`" + `: Client locale (default: ` + "`enUS`" + `)
- ` + "`databases.*`" + `: MySQL connection settings
- ` + "`server.dbc_path`" + `: Where to copy server DBCs (default: ` + "`${TC_SERVER_PATH}/bin/dbc`" + `)
- ` + "`trinitycore.source_path`" + `: TrinityCore source directory for patches
- ` + "`trinitycore.scripts_path`" + `: Where to deploy C++ scripts

Set environment variables in your shell or use [Peacebloom](https://github.com/suprsokr/peacebloom):

` + "```" + `bash
export WOTLK_PATH="/home/me/WoW-3.3.5"
export TC_SOURCE_PATH="/home/me/TrinityCore"
export TC_SERVER_PATH="/home/me/server"
export MYSQL_HOST="127.0.0.1"
` + "```" + `

## Documentation

Full documentation: https://github.com/suprsokr/thorium/tree/main/docs

- [Installation](https://github.com/suprsokr/thorium/blob/main/docs/install.md)
- [Initialization](https://github.com/suprsokr/thorium/blob/main/docs/init.md)
- [Mods](https://github.com/suprsokr/thorium/blob/main/docs/mods.md)
- [Commands](https://github.com/suprsokr/thorium/blob/main/docs/commands.md)
- [DBC Files](https://github.com/suprsokr/thorium/blob/main/docs/dbc.md)
- [SQL Migrations](https://github.com/suprsokr/thorium/blob/main/docs/sql-migrations.md)
- [Scripts](https://github.com/suprsokr/thorium/blob/main/docs/scripts.md)
- [Distribution](https://github.com/suprsokr/thorium/blob/main/docs/distribution.md)
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
	fmt.Println("  1. Edit config.json with your paths (or use Peacebloom: https://github.com/suprsokr/peacebloom)")
	fmt.Println("  2. Create your first mod: thorium create-mod my-mod")
	fmt.Println("  3. Start building!")
	fmt.Println()
	fmt.Println("Note: DBC databases will be automatically set up when you create your first DBC migration.")

	return nil
}

// InitDB initializes or recreates the Thorium databases and sets up the full DBC workflow
func InitDB(configPath string, args []string) error {
	fs := flag.NewFlagSet("init db", flag.ExitOnError)
	fs.Parse(args)

	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Initializing DBC Workflow                               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Step 1: Create databases
	fmt.Println("Step 1: Creating databases...")
	if err := initializeDatabases(cfg); err != nil {
		return err
	}
	fmt.Println("  ✓ Databases created")
	fmt.Println()

	// Step 2: Extract DBCs
	fmt.Println("Step 2: Extracting DBCs from WoW client...")
	if cfg.WoTLK.Path == "" || cfg.WoTLK.Path == "${WOTLK_PATH}" {
		fmt.Println()
		fmt.Println("  ⚠ wotlk.path not configured in config.json")
		fmt.Println("  Please set WOTLK_PATH environment variable or edit config.json")
		fmt.Println()
		fmt.Println("After configuring, complete the setup:")
		fmt.Println("  1. thorium extract --dbc")
		fmt.Println("  2. thorium import dbc --database dbc_source")
		fmt.Println("  3. mysqldump dbc_source | mysql dbc")
		return fmt.Errorf("wotlk.path not configured")
	}

	extractArgs := []string{"--dbc"}
	if err := Extract(cfg, extractArgs); err != nil {
		return fmt.Errorf("extract DBCs: %w", err)
	}
	fmt.Println()

	// Step 3: Import to dbc_source
	fmt.Println("Step 3: Importing DBCs to baseline database...")
	importArgs := []string{"dbc", "--database", "dbc_source"}
	if err := Import(cfg, importArgs); err != nil {
		return fmt.Errorf("import DBCs: %w", err)
	}
	fmt.Println()

	// Step 4: Copy dbc_source to dbc
	fmt.Println("Step 4: Copying baseline to development database...")
	if err := copyDBCSourceToDBC(cfg); err != nil {
		return fmt.Errorf("copy database: %w", err)
	}
	fmt.Println("  ✓ Copied dbc_source to dbc")
	fmt.Println()

	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              DBC Workflow Ready!                                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("You can now create DBC migrations and use thorium build")
	fmt.Println()

	return nil
}

// copyDBCSourceToDBC copies the dbc_source database to dbc using mysqldump
func copyDBCSourceToDBC(cfg *config.Config) error {
	srcDB := cfg.Databases.DBCSource
	dstDB := cfg.Databases.DBC

	// Build mysqldump command
	dumpCmd := fmt.Sprintf("mysqldump -h%s -P%s -u%s", srcDB.Host, srcDB.Port, srcDB.User)
	if srcDB.Password != "" {
		dumpCmd += fmt.Sprintf(" -p%s", srcDB.Password)
	}
	dumpCmd += fmt.Sprintf(" %s", srcDB.Name)

	// Build mysql import command
	importCmd := fmt.Sprintf("mysql -h%s -P%s -u%s", dstDB.Host, dstDB.Port, dstDB.User)
	if dstDB.Password != "" {
		importCmd += fmt.Sprintf(" -p%s", dstDB.Password)
	}
	importCmd += fmt.Sprintf(" %s", dstDB.Name)

	// Execute: mysqldump | mysql
	fullCmd := fmt.Sprintf("%s | %s", dumpCmd, importCmd)
	
	cmd := exec.Command("sh", "-c", fullCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("execute mysqldump: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// initializeDatabases creates DBC databases (dbc and dbc_source)
// Note: world database is created and managed by TrinityCore
func initializeDatabases(cfg *config.Config) error {
	databases := []struct {
		name   string
		config config.DBConfig
	}{
		{"dbc", cfg.Databases.DBC},
		{"dbc_source", cfg.Databases.DBCSource},
	}

	for _, db := range databases {
		fmt.Printf("  Creating database: %s\n", db.name)
		if err := database.CreateDatabase(db.config); err != nil {
			return fmt.Errorf("create %s database: %w", db.name, err)
		}
	}

	return nil
}

// checkDBCDatabasesSetup checks if DBC databases exist and have tables
func checkDBCDatabasesSetup(cfg *config.Config) error {
	// Check if dbc database exists and has tables
	count, err := database.QueryValue(cfg.Databases.DBC, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '"+cfg.Databases.DBC.Name+"'")
	if err != nil {
		return fmt.Errorf("DBC database not configured or unreachable")
	}
	
	if count == "0" {
		return fmt.Errorf("DBC database exists but has no tables (not yet imported)")
	}

	return nil
}

// printDBCSetupInstructions prints helpful instructions for setting up DBC workflow
func printDBCSetupInstructions() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              DBC Workflow Not Yet Initialized                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Your mod has DBC migrations, but the DBC databases haven't been set up yet.")
	fmt.Println()
	fmt.Println("To set up the DBC workflow:")
	fmt.Println()
	fmt.Println("  1. Ensure config.json has correct database settings:")
	fmt.Println("     - wotlk.path (path to your WoW 3.3.5 client)")
	fmt.Println("     - databases.dbc (development database)")
	fmt.Println("     - databases.dbc_source (pristine baseline)")
	fmt.Println()
	fmt.Println("  2. Run the setup command:")
	fmt.Println("     thorium init db")
	fmt.Println()
	fmt.Println("     This will:")
	fmt.Println("     - Create the databases")
	fmt.Println("     - Extract DBCs from your WoW client")
	fmt.Println("     - Import DBCs to dbc_source")
	fmt.Println("     - Copy dbc_source to dbc")
	fmt.Println()
	fmt.Println("  3. Try your command again")
	fmt.Println()
}

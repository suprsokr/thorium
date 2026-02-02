// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"thorium-cli/internal/config"
)

// CreateMigration creates a new migration file pair in a mod
func CreateMigration(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("create-migration", flag.ExitOnError)
	modName := fs.String("mod", "", "Mod name (required)")
	dbType := fs.String("db", "", "Database type: 'dbc', 'world', or 'dbc,world' (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: thorium create-migration [--mod <mod>] --db <dbc|world|dbc,world> <description>

Create a new migration file pair in a mod.

Flags:
  --mod <name>    Mod name (optional, inferred from current directory if not provided)
  --db <type>     Database type(s) (required)
                  Options: 'dbc', 'world', or 'dbc,world' to create both

Arguments:
  <description>   Description of the migration (required)
                  Spaces will be converted to underscores

Examples:
  # From inside a mod directory (mod name inferred):
  cd mods/my-mod
  thorium create-migration --db world add_custom_npc
  
  # Explicit mod name:
  thorium create-migration --mod my-mod --db dbc add_custom_item
  thorium create-migration --mod my-mod --db dbc,world add_custom_feature

The migration will be created with a timestamp prefix:
  YYYYMMDD_HHMMSS_description.sql
  YYYYMMDD_HHMMSS_description.rollback.sql
`)
	}
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		fs.Usage()
		return fmt.Errorf("migration description required")
	}

	description := strings.Join(remaining, "_")
	description = sanitizeName(description)

	// Validate required flags
	if *dbType == "" {
		return fmt.Errorf("--db flag is required (use 'dbc', 'world', or 'dbc,world')")
	}

	// Parse comma-separated database types
	dbTypes := parseDBTypes(*dbType)
	if len(dbTypes) == 0 {
		return fmt.Errorf("--db must be 'dbc', 'world', or 'dbc,world', got '%s'", *dbType)
	}

	// Infer mod name from current directory if not provided
	finalModName := *modName
	if finalModName == "" {
		inferred, err := inferModName(cfg)
		if err != nil {
			return fmt.Errorf("could not infer mod name: %v\nUse --mod flag to specify mod name explicitly", err)
		}
		finalModName = inferred
	}

	// Check mod exists
	modPath := filepath.Join(cfg.GetModsPath(), finalModName)
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return fmt.Errorf("mod not found: %s\nRun 'thorium create-mod %s' first", finalModName, finalModName)
	}

	// Generate timestamp and base name
	timestamp := time.Now().Format("20060102_150405")
	baseName := fmt.Sprintf("%s_%s", timestamp, description)

	// Create migrations for each database type
	var createdFiles []string
	for _, db := range dbTypes {
		sqlDir := filepath.Join(modPath, db+"_sql")
		applyFile := filepath.Join(sqlDir, baseName+".sql")
		rollbackFile := filepath.Join(sqlDir, baseName+".rollback.sql")

		// Ensure directory exists
		if err := os.MkdirAll(sqlDir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}

		// Create apply migration
		applyContent := generateApplyTemplate(db, description)
		if err := os.WriteFile(applyFile, []byte(applyContent), 0644); err != nil {
			return fmt.Errorf("write apply file: %w", err)
		}

		// Create rollback migration
		rollbackContent := generateRollbackTemplate(db, description)
		if err := os.WriteFile(rollbackFile, []byte(rollbackContent), 0644); err != nil {
			return fmt.Errorf("write rollback file: %w", err)
		}

		createdFiles = append(createdFiles, applyFile)
	}

	fmt.Printf("Created migration in %s:\n", finalModName)
	for _, file := range createdFiles {
		fmt.Printf("  %s\n", file)
	}
	fmt.Println()
	fmt.Printf("Edit your migrations:\n")
	for _, file := range createdFiles {
		fmt.Printf("  %s\n", file)
	}

	return nil
}

// parseDBTypes parses comma-separated database types and validates them
func parseDBTypes(dbStr string) []string {
	parts := strings.Split(dbStr, ",")
	var validTypes []string
	seen := make(map[string]bool)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "dbc" || part == "world" {
			if !seen[part] {
				validTypes = append(validTypes, part)
				seen[part] = true
			}
		}
	}

	return validTypes
}

// inferModName tries to infer the mod name from the current working directory
// It checks if we're inside a mods/<mod-name>/ directory structure
func inferModName(cfg *config.Config) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current directory: %w", err)
	}

	modsPath := cfg.GetModsPath()
	modsPathAbs, err := filepath.Abs(modsPath)
	if err != nil {
		return "", fmt.Errorf("resolve mods path: %w", err)
	}

	cwdAbs, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("resolve current directory: %w", err)
	}

	// Check if current directory is inside mods path
	relPath, err := filepath.Rel(modsPathAbs, cwdAbs)
	if err != nil {
		return "", fmt.Errorf("not inside mods directory")
	}

	// relPath should start with a mod name, e.g., "my-mod" or "my-mod/some/subdir"
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) == 0 || parts[0] == "." || parts[0] == ".." {
		return "", fmt.Errorf("not inside a mod directory")
	}

	modName := parts[0]
	
	// Verify the mod directory actually exists
	modPath := filepath.Join(modsPath, modName)
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return "", fmt.Errorf("mod directory not found: %s", modPath)
	}

	return modName, nil
}

// sanitizeName converts a description to a safe filename component
func sanitizeName(name string) string {
	// Replace spaces and special chars with underscores
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		if r == ' ' || r == '-' {
			return '_'
		}
		return -1 // Remove other characters
	}, name)

	// Convert to lowercase and collapse multiple underscores
	result = strings.ToLower(result)
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	result = strings.Trim(result, "_")

	return result
}

// generateApplyTemplate creates the apply migration template
func generateApplyTemplate(dbType, description string) string {
	if dbType == "dbc" {
		return fmt.Sprintf(`-- Migration: %s
-- Database: DBC
-- Created: %s

-- Add your DBC changes here
-- Example:
-- DELETE FROM `+"`Item`"+` WHERE `+"`id`"+` = 90001;
-- INSERT INTO `+"`Item`"+` (`+"`id`"+`, `+"`class`"+`, `+"`subclass`"+`, ...) VALUES (90001, 2, 7, ...);

`, description, time.Now().Format("2006-01-02 15:04:05"))
	}

	return fmt.Sprintf(`-- Migration: %s
-- Database: World
-- Created: %s

-- Add your World database changes here
-- Example:
-- DELETE FROM `+"`item_template`"+` WHERE `+"`entry`"+` = 90001;
-- INSERT INTO `+"`item_template`"+` (`+"`entry`"+`, `+"`name`"+`, ...) VALUES (90001, 'My Item', ...);

`, description, time.Now().Format("2006-01-02 15:04:05"))
}

// generateRollbackTemplate creates the rollback migration template
func generateRollbackTemplate(dbType, description string) string {
	if dbType == "dbc" {
		return fmt.Sprintf(`-- Rollback: %s
-- Database: DBC
-- Created: %s

-- Add your DBC rollback changes here
-- This should undo everything in the apply migration
-- Example:
-- DELETE FROM `+"`Item`"+` WHERE `+"`id`"+` = 90001;

`, description, time.Now().Format("2006-01-02 15:04:05"))
	}

	return fmt.Sprintf(`-- Rollback: %s
-- Database: World
-- Created: %s

-- Add your World database rollback changes here
-- This should undo everything in the apply migration
-- Example:
-- DELETE FROM `+"`item_template`"+` WHERE `+"`entry`"+` = 90001;

`, description, time.Now().Format("2006-01-02 15:04:05"))
}

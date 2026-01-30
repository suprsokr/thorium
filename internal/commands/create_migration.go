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
	dbType := fs.String("db", "", "Database type: 'dbc' or 'world' (required)")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("migration description required\nUsage: thorium create-migration --mod <mod> --db <dbc|world> <description>")
	}

	description := strings.Join(remaining, "_")
	description = sanitizeName(description)

	// Validate required flags
	if *modName == "" {
		return fmt.Errorf("--mod flag is required")
	}
	if *dbType == "" {
		return fmt.Errorf("--db flag is required (use 'dbc' or 'world')")
	}
	if *dbType != "dbc" && *dbType != "world" {
		return fmt.Errorf("--db must be 'dbc' or 'world', got '%s'", *dbType)
	}

	// Check mod exists
	modPath := filepath.Join(cfg.GetModsPath(), *modName)
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return fmt.Errorf("mod not found: %s\nRun 'thorium create-mod %s' first", *modName, *modName)
	}

	// Generate timestamp and filenames
	timestamp := time.Now().Format("20060102_150405")
	baseName := fmt.Sprintf("%s_%s", timestamp, description)

	sqlDir := filepath.Join(modPath, *dbType+"_sql")
	applyFile := filepath.Join(sqlDir, baseName+".sql")
	rollbackFile := filepath.Join(sqlDir, baseName+".rollback.sql")

	// Ensure directory exists
	if err := os.MkdirAll(sqlDir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Create apply migration
	applyContent := generateApplyTemplate(*dbType, description)
	if err := os.WriteFile(applyFile, []byte(applyContent), 0644); err != nil {
		return fmt.Errorf("write apply file: %w", err)
	}

	// Create rollback migration
	rollbackContent := generateRollbackTemplate(*dbType, description)
	if err := os.WriteFile(rollbackFile, []byte(rollbackContent), 0644); err != nil {
		return fmt.Errorf("write rollback file: %w", err)
	}

	fmt.Printf("Created migration in %s:\n", *modName)
	fmt.Printf("  Apply:    %s\n", filepath.Base(applyFile))
	fmt.Printf("  Rollback: %s\n", filepath.Base(rollbackFile))
	fmt.Println()
	fmt.Printf("Edit your migration:\n")
	fmt.Printf("  %s\n", applyFile)

	return nil
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

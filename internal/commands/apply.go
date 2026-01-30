// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"thorium-cli/internal/config"
	"thorium-cli/internal/database"
)

// Apply applies SQL migrations for mods
func Apply(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	modName := fs.String("mod", "", "Apply migrations for specific mod only")
	dbType := fs.String("db", "", "Apply only 'dbc' or 'world' migrations")
	fs.Parse(args)

	fmt.Println("=== Applying SQL Migrations ===")
	fmt.Println()

	// Get list of mods
	mods, err := listMods(cfg)
	if err != nil {
		return fmt.Errorf("list mods: %w", err)
	}

	// Filter to specific mod if requested
	if *modName != "" {
		found := false
		for _, m := range mods {
			if m == *modName {
				mods = []string{m}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("mod not found: %s", *modName)
		}
	}

	if len(mods) == 0 {
		fmt.Println("No mods found.")
		return nil
	}

	// Process each mod
	for _, mod := range mods {
		if *dbType == "" || *dbType == "dbc" {
			if err := applyMigrations(cfg, mod, "dbc"); err != nil {
				return err
			}
		}
		if *dbType == "" || *dbType == "world" {
			if err := applyMigrations(cfg, mod, "world"); err != nil {
				return err
			}
		}
	}

	fmt.Println("\n=== Migrations Complete ===")
	return nil
}

// Rollback rolls back SQL migrations for mods
func Rollback(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("rollback", flag.ExitOnError)
	modName := fs.String("mod", "", "Rollback migrations for specific mod only")
	dbType := fs.String("db", "", "Rollback only 'dbc' or 'world' migrations")
	all := fs.Bool("all", false, "Rollback all migrations (not just last)")
	fs.Parse(args)

	fmt.Println("=== Rolling Back SQL Migrations ===")
	fmt.Println()

	// Get list of mods
	mods, err := listMods(cfg)
	if err != nil {
		return fmt.Errorf("list mods: %w", err)
	}

	// Filter to specific mod if requested
	if *modName != "" {
		found := false
		for _, m := range mods {
			if m == *modName {
				mods = []string{m}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("mod not found: %s", *modName)
		}
	}

	// Process each mod in reverse order
	for i := len(mods) - 1; i >= 0; i-- {
		mod := mods[i]
		if *dbType == "" || *dbType == "world" {
			if err := rollbackMigrations(cfg, mod, "world", *all); err != nil {
				return err
			}
		}
		if *dbType == "" || *dbType == "dbc" {
			if err := rollbackMigrations(cfg, mod, "dbc", *all); err != nil {
				return err
			}
		}
	}

	fmt.Println("\n=== Rollback Complete ===")
	return nil
}

// applyMigrations applies migrations for a single mod and db type
func applyMigrations(cfg *config.Config, mod, dbType string) error {
	migrationDir := filepath.Join(cfg.GetModsPath(), mod, dbType+"_sql")
	appliedDir := filepath.Join(cfg.GetAppliedMigrationsPath(), mod, dbType)

	migrations, err := listMigrations(migrationDir)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return nil
	}

	fmt.Printf("[%s] Processing %s migrations...\n", mod, dbType)

	// Ensure applied directory exists
	if err := os.MkdirAll(appliedDir, 0755); err != nil {
		return fmt.Errorf("create applied dir: %w", err)
	}

	// Get database config
	var db config.DBConfig
	if dbType == "dbc" {
		db = cfg.Databases.DBC
	} else {
		db = cfg.Databases.World
	}

	applied := 0
	skipped := 0

	for _, migration := range migrations {
		sqlFile := filepath.Join(migrationDir, migration)
		appliedMarker := filepath.Join(appliedDir, migration+".applied")

		// Check if already applied
		markerInfo, err := os.Stat(appliedMarker)
		if err == nil {
			// Check if migration was modified after being applied
			sqlInfo, err := os.Stat(sqlFile)
			if err == nil && sqlInfo.ModTime().After(markerInfo.ModTime()) {
				// Need to rollback and re-apply
				fmt.Printf("  [modified] %s - rolling back and re-applying\n", migration)

				rollbackFile := strings.TrimSuffix(sqlFile, ".sql") + ".rollback.sql"
				if _, err := os.Stat(rollbackFile); err == nil {
					if err := runSQLFile(db, rollbackFile); err != nil {
						return fmt.Errorf("rollback %s: %w", migration, err)
					}
				}
				os.Remove(appliedMarker)
			} else {
				skipped++
				continue
			}
		}

		// Apply migration
		fmt.Printf("  [apply] %s\n", migration)
		if err := runSQLFile(db, sqlFile); err != nil {
			return fmt.Errorf("apply %s: %w", migration, err)
		}

		// Mark as applied
		if err := os.WriteFile(appliedMarker, []byte(time.Now().Format(time.RFC3339)), 0644); err != nil {
			return fmt.Errorf("write marker: %w", err)
		}

		applied++
	}

	if applied > 0 || skipped > 0 {
		fmt.Printf("  Applied: %d, Skipped: %d\n", applied, skipped)
	}

	return nil
}

// rollbackMigrations rolls back migrations for a single mod and db type
func rollbackMigrations(cfg *config.Config, mod, dbType string, all bool) error {
	migrationDir := filepath.Join(cfg.GetModsPath(), mod, dbType+"_sql")
	appliedDir := filepath.Join(cfg.GetAppliedMigrationsPath(), mod, dbType)

	migrations, err := listMigrations(migrationDir)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return nil
	}

	// Reverse order for rollback
	sort.Sort(sort.Reverse(sort.StringSlice(migrations)))

	// Get database config
	var db config.DBConfig
	if dbType == "dbc" {
		db = cfg.Databases.DBC
	} else {
		db = cfg.Databases.World
	}

	rolledBack := 0

	for _, migration := range migrations {
		appliedMarker := filepath.Join(appliedDir, migration+".applied")

		// Check if applied
		if _, err := os.Stat(appliedMarker); os.IsNotExist(err) {
			continue // Not applied, skip
		}

		rollbackFile := filepath.Join(migrationDir, strings.TrimSuffix(migration, ".sql")+".rollback.sql")
		if _, err := os.Stat(rollbackFile); os.IsNotExist(err) {
			return fmt.Errorf("rollback file not found: %s", rollbackFile)
		}

		fmt.Printf("  [rollback] %s\n", migration)
		if err := runSQLFile(db, rollbackFile); err != nil {
			return fmt.Errorf("rollback %s: %w", migration, err)
		}

		os.Remove(appliedMarker)
		rolledBack++

		if !all {
			break // Only rollback the last one
		}
	}

	if rolledBack > 0 {
		fmt.Printf("[%s] Rolled back %d %s migration(s)\n", mod, rolledBack, dbType)
	}

	return nil
}

// runSQLFile executes a SQL file
func runSQLFile(db config.DBConfig, sqlFile string) error {
	content, err := os.ReadFile(sqlFile)
	if err != nil {
		return fmt.Errorf("read SQL: %w", err)
	}

	return database.Execute(db, string(content))
}

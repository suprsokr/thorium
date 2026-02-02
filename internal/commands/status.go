// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"thorium-cli/internal/config"
)

// Status shows the status of migrations and mods
func Status(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	modName := fs.String("mod", "", "Show status for specific mod only")
	fs.Parse(args)

	fmt.Println("=== Thorium Status ===")
	fmt.Println()
	fmt.Printf("Workspace: %s\n", cfg.WorkspaceRoot)
	fmt.Println()

	// List mods
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
		fmt.Println("No mods found in", cfg.GetModsPath())
		fmt.Println("Run 'thorium create-mod <name>' to create one.")
		return nil
	}

	// Only show mods list if not filtering to a specific mod
	if *modName == "" {
		fmt.Printf("Found %d mod(s):\n", len(mods))
		for _, mod := range mods {
			fmt.Printf("  - %s\n", mod)
		}
		fmt.Println()
	}

	// Show migration status for each mod
	appliedPath := cfg.GetAppliedMigrationsPath()

	for _, mod := range mods {
		fmt.Printf("=== %s ===\n", mod)

		// DBC migrations
		dbcPath := filepath.Join(cfg.GetModsPath(), mod, "dbc_sql")
		if migrations, _ := listMigrations(dbcPath); len(migrations) > 0 {
			fmt.Println("  DBC Migrations:")
			for _, m := range migrations {
				status := "pending"
				appliedFile := filepath.Join(appliedPath, mod, "dbc", m+".applied")
				if _, err := os.Stat(appliedFile); err == nil {
					status = "applied"
				}
				fmt.Printf("    [%s] %s\n", status, m)
			}
		}

		// World migrations
		worldPath := filepath.Join(cfg.GetModsPath(), mod, "world_sql")
		if migrations, _ := listMigrations(worldPath); len(migrations) > 0 {
			fmt.Println("  World Migrations:")
			for _, m := range migrations {
				status := "pending"
				appliedFile := filepath.Join(appliedPath, mod, "world", m+".applied")
				if _, err := os.Stat(appliedFile); err == nil {
					status = "applied"
				}
				fmt.Printf("    [%s] %s\n", status, m)
			}
		}

		// LuaXML files
		luaxmlPath := filepath.Join(cfg.GetModsPath(), mod, "luaxml")
		if count := countFiles(luaxmlPath); count > 0 {
			fmt.Printf("  LuaXML Files: %d\n", count)
		}

		fmt.Println()
	}

	return nil
}

// listMods returns a sorted list of mod directories
func listMods(cfg *config.Config) ([]string, error) {
	modsPath := cfg.GetModsPath()

	entries, err := os.ReadDir(modsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var mods []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			mods = append(mods, e.Name())
		}
	}

	sort.Strings(mods)
	return mods, nil
}

// listMigrations returns SQL migration files (excluding rollbacks)
func listMigrations(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var migrations []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".sql") && !strings.HasSuffix(name, ".rollback.sql") {
			migrations = append(migrations, name)
		}
	}

	sort.Strings(migrations)
	return migrations, nil
}

// countFiles counts non-hidden files in a directory recursively
func countFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
			count++
		}
		return nil
	})
	return count
}

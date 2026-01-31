// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"thorium-cli/internal/config"
	"thorium-cli/internal/dbc"
	"thorium-cli/internal/mpq"
	"thorium-cli/internal/scripts"
)

// Build performs a full build: apply migrations, export DBCs, overlay LuaXML, package MPQs
func Build(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	modName := fs.String("mod", "", "Build specific mod only")
	skipMigrations := fs.Bool("skip-migrations", false, "Skip SQL migrations")
	skipExport := fs.Bool("skip-export", false, "Skip DBC export")
	skipPackage := fs.Bool("skip-package", false, "Skip MPQ packaging")
	skipServer := fs.Bool("skip-server", false, "Skip copying to server")
	fs.Parse(args)

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        Thorium Build System              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
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
		fmt.Println("No mods found in", cfg.GetModsPath())
		return nil
	}

	fmt.Printf("Building %d mod(s): %v\n\n", len(mods), mods)

	// Step 1: Apply SQL migrations
	if !*skipMigrations {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 1: Applying SQL Migrations         │")
		fmt.Println("└──────────────────────────────────────────┘")

		for _, mod := range mods {
			if err := applyMigrations(cfg, mod, "dbc"); err != nil {
				return fmt.Errorf("apply dbc migrations for %s: %w", mod, err)
			}
			if err := applyMigrations(cfg, mod, "world"); err != nil {
				return fmt.Errorf("apply world migrations for %s: %w", mod, err)
			}
		}
		fmt.Println()
	}

	// Step 2: Export modified DBCs
	if !*skipExport {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 2: Exporting Modified DBCs         │")
		fmt.Println("└──────────────────────────────────────────┘")

		exporter := dbc.NewExporter(cfg)
		tables, err := exporter.Export()
		if err != nil {
			return fmt.Errorf("export DBCs: %w", err)
		}
		if len(tables) > 0 {
			fmt.Printf("Exported %d DBC table(s)\n", len(tables))
		}
		fmt.Println()
	}

	// Step 3: Check for LuaXML modifications
	fmt.Println("┌──────────────────────────────────────────┐")
	fmt.Println("│  Step 3: Checking LuaXML Modifications   │")
	fmt.Println("└──────────────────────────────────────────┘")

	// Collect modified LuaXML files from all mods
	var allModifiedLuaXML []mpq.ModifiedLuaXMLFile
	for _, mod := range mods {
		modFiles, err := findModifiedLuaXMLInMod(cfg, mod)
		if err != nil {
			return fmt.Errorf("check luaxml for %s: %w", mod, err)
		}
		if len(modFiles) > 0 {
			fmt.Printf("[%s] Found %d modified LuaXML file(s)\n", mod, len(modFiles))
			// Convert to mpq type
			for _, f := range modFiles {
				allModifiedLuaXML = append(allModifiedLuaXML, mpq.ModifiedLuaXMLFile{
					ModName:  f.ModName,
					FilePath: f.FilePath,
					RelPath:  f.RelPath,
				})
			}
		}
	}
	fmt.Println()

	// Step 4: Deploy Scripts to TrinityCore
	if cfg.TrinityCore.ScriptsPath != "" {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 4: Deploying Scripts               │")
		fmt.Println("└──────────────────────────────────────────┘")

		if err := scripts.DeployScripts(cfg, mods); err != nil {
			return fmt.Errorf("deploy scripts: %w", err)
		}
		fmt.Println()
	}

	// Step 5: Package and distribute
	if !*skipPackage {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 5: Packaging and Distributing      │")
		fmt.Println("└──────────────────────────────────────────┘")

		builder := mpq.NewBuilder(cfg)

		// Copy to server
		if !*skipServer && cfg.Server.DBCPath != "" {
			count, err := builder.CopyToServer()
			if err != nil {
				return fmt.Errorf("copy to server: %w", err)
			}
			if count > 0 {
				fmt.Printf("Copied %d DBC file(s) to server\n", count)
			}
		}

		// Package DBC MPQ
		dbcCount, err := builder.PackageDBCs()
		if err != nil {
			return fmt.Errorf("package DBCs: %w", err)
		}
		if dbcCount > 0 {
			fmt.Printf("Created DBC MPQ with %d file(s)\n", dbcCount)
		}

		// Package LuaXML MPQ from modified files
		if len(allModifiedLuaXML) > 0 {
			luaxmlCount, err := builder.PackageLuaXMLFromMods(allModifiedLuaXML)
			if err != nil {
				return fmt.Errorf("package LuaXML: %w", err)
			}
			if luaxmlCount > 0 {
				fmt.Printf("Created LuaXML MPQ with %d file(s)\n", luaxmlCount)
			}
		}
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║           Build Complete!                ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// Print summary
	fmt.Println("Output locations:")
	if cfg.Server.DBCPath != "" {
		fmt.Printf("  Server DBCs: %s\n", cfg.Server.DBCPath)
	}
	fmt.Printf("  Client DBC MPQ: %s/%s\n", cfg.GetClientDataPath(), cfg.Output.DBCMPQ)
	fmt.Printf("  Client LuaXML MPQ: %s/%s\n", cfg.GetClientLocalePath(), cfg.GetMPQName(cfg.Output.LuaXMLMPQ))

	return nil
}

// ModifiedLuaXMLFile represents a modified LuaXML file from a mod
type modifiedLuaXMLFile struct {
	ModName  string // Which mod it's from
	FilePath string // Absolute path to the file
	RelPath  string // Relative path (for MPQ)
}

// findModifiedLuaXMLInMod finds LuaXML files in a mod that differ from source
func findModifiedLuaXMLInMod(cfg *config.Config, mod string) ([]modifiedLuaXMLFile, error) {
	modLuaXML := filepath.Join(cfg.GetModsPath(), mod, "luaxml")
	luaxmlSource := cfg.GetLuaXMLSourcePath()

	// Check if mod has luaxml directory
	if _, err := os.Stat(modLuaXML); os.IsNotExist(err) {
		return nil, nil // No luaxml folder, skip
	}

	var modified []modifiedLuaXMLFile

	err := filepath.Walk(modLuaXML, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// Skip hidden files and common junk files
		name := info.Name()
		if strings.HasPrefix(name, ".") || name == "Thumbs.db" {
			return nil
		}

		// Get relative path
		relPath, _ := filepath.Rel(modLuaXML, path)

		// Compare with source file
		sourcePath := filepath.Join(luaxmlSource, relPath)

		// Check if files are different
		if !filesAreIdentical(path, sourcePath) {
			modified = append(modified, modifiedLuaXMLFile{
				ModName:  mod,
				FilePath: path,
				RelPath:  relPath,
			})
		}

		return nil
	})

	return modified, err
}

// filesAreIdentical checks if two files have identical content
func filesAreIdentical(file1, file2 string) bool {
	data1, err1 := os.ReadFile(file1)
	data2, err2 := os.ReadFile(file2)

	if err1 != nil || err2 != nil {
		return false // If either can't be read, consider different
	}

	if len(data1) != len(data2) {
		return false
	}

	for i := range data1 {
		if data1[i] != data2[i] {
			return false
		}
	}

	return true
}

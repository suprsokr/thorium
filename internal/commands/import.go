// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"

	"thorium-cli/internal/config"
	"thorium-cli/internal/dbc"
)

// Import imports data files into the MySQL database
func Import(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing import type\nUsage: thorium import <dbc> [flags]\n\nExample:\n  thorium import dbc --source /path/to/dbc")
	}

	importType := args[0]
	subArgs := args[1:]

	switch importType {
	case "dbc":
		return importDBC(cfg, subArgs)
	default:
		return fmt.Errorf("unknown import type: %s\nSupported types: dbc", importType)
	}
}

// importDBC imports DBC files into the MySQL database
func importDBC(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("import dbc", flag.ExitOnError)
	sourcePath := fs.String("source", "", "Path to DBC files (default: shared/dbc/dbc_source)")
	database := fs.String("database", "", "Target database name (default: from config)")
	skipExisting := fs.Bool("skip-existing", false, "Skip tables that already exist")
	fs.Parse(args)

	fmt.Println("=== Importing DBC Files to Database ===")
	fmt.Println()

	// Determine source path
	dbcSource := cfg.GetDBCSourcePath()
	if *sourcePath != "" {
		dbcSource = *sourcePath
	}

	// Override database if specified
	dbConfig := cfg.Databases.DBC
	if *database != "" {
		dbConfig.Name = *database
	}

	// Check if source exists
	if _, err := os.Stat(dbcSource); os.IsNotExist(err) {
		return fmt.Errorf("DBC source directory does not exist: %s\nRun 'thorium extract --dbc' or specify --source path", dbcSource)
	}

	// Check if there are DBC files
	entries, err := os.ReadDir(dbcSource)
	if err != nil {
		return fmt.Errorf("read source directory: %w", err)
	}

	dbcCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[len(entry.Name())-4:] == ".dbc" {
			dbcCount++
		}
	}

	if dbcCount == 0 {
		return fmt.Errorf("no DBC files found in %s", dbcSource)
	}

	fmt.Printf("Source: %s\n", dbcSource)
	fmt.Printf("Found %d DBC files\n", dbcCount)
	fmt.Printf("Database: %s@%s:%s/%s\n", dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.Name)
	fmt.Println()

	// Create importer with custom source path and database config
	importer := dbc.NewImporterWithDB(cfg, dbcSource, dbConfig, *skipExisting)
	tables, err := importer.Import()
	if err != nil {
		return fmt.Errorf("import DBCs: %w", err)
	}

	fmt.Printf("\nImported %d DBC tables to database\n", len(tables))
	fmt.Println("\n=== Import Complete ===")
	return nil
}

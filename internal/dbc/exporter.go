// Copyright (c) 2025 Thorium

package dbc

import (
	"database/sql"
	"fmt"
	"os"

	appconfig "thorium-cli/internal/config"
)

// Exporter exports DBC files from the database
type Exporter struct {
	cfg    *appconfig.Config
	dbcCfg *Config
}

// NewExporter creates a new DBC exporter
func NewExporter(cfg *appconfig.Config) *Exporter {
	return NewExporterWithDB(cfg, cfg.Databases.DBC)
}

// NewExporterWithDB creates a new DBC exporter with a custom database config
func NewExporterWithDB(cfg *appconfig.Config, dbConfig appconfig.DBConfig) *Exporter {
	// Create DBC tool config from thorium config
	dbcCfg := &Config{
		DBC: DBConfig{
			User:     dbConfig.User,
			Password: dbConfig.Password,
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Name:     dbConfig.Name,
		},
		Paths: PathConfig{
			Base:   cfg.GetDBCSourcePath(),
			Export: cfg.GetDBCOutPath(),
			Meta:   cfg.GetDBCMetaPath(),
		},
		Options: OptionConfig{
			UseVersioning: true,
		},
	}

	return &Exporter{
		cfg:    cfg,
		dbcCfg: dbcCfg,
	}
}

// Export exports all modified DBC tables to files
// Returns the list of table names that were actually exported (not all files in dir)
func (e *Exporter) Export() ([]string, error) {
	// Ensure directories exist
	if err := os.MkdirAll(e.dbcCfg.Paths.Export, 0755); err != nil {
		return nil, fmt.Errorf("create export dir: %w", err)
	}

	// Connect to database
	db, err := openDB(e.dbcCfg.DBC)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	// Export all modified DBCs - returns only the tables that were actually exported
	exported, err := ExportDBCs(db, e.dbcCfg)
	if err != nil {
		return nil, fmt.Errorf("export DBCs: %w", err)
	}

	return exported, nil
}

// Import imports DBC files into the database
func (e *Exporter) Import() ([]string, error) {
	// Ensure source directory exists
	if _, err := os.Stat(e.dbcCfg.Paths.Base); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist: %s", e.dbcCfg.Paths.Base)
	}

	// Connect to database
	db, err := openDB(e.dbcCfg.DBC)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	// Import all DBCs (false = don't skip existing)
	if err := ImportDBCs(db, false, e.dbcCfg); err != nil {
		return nil, fmt.Errorf("import DBCs: %w", err)
	}

	// List imported tables
	return getImportedTables(db)
}

// getImportedTables returns a list of DBC tables in the database
func getImportedTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			continue
		}
		// Skip internal tables
		if table != "dbc_version" && table != "dbc_checksums" {
			tables = append(tables, table)
		}
	}
	return tables, nil
}

// Importer imports DBC files into the database
type Importer struct {
	cfg          *appconfig.Config
	dbcCfg       *Config
	skipExisting bool
}

// NewImporter creates a new DBC importer with custom source path
func NewImporter(cfg *appconfig.Config, sourcePath string, skipExisting bool) *Importer {
	return NewImporterWithDB(cfg, sourcePath, cfg.Databases.DBC, skipExisting)
}

// NewImporterWithDB creates a new DBC importer with custom source path and database config
func NewImporterWithDB(cfg *appconfig.Config, sourcePath string, dbConfig appconfig.DBConfig, skipExisting bool) *Importer {
	dbcCfg := &Config{
		DBC: DBConfig{
			User:     dbConfig.User,
			Password: dbConfig.Password,
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Name:     dbConfig.Name,
		},
		Paths: PathConfig{
			Base:     sourcePath,
			Export:   cfg.GetDBCOutPath(),
			Meta:     cfg.GetDBCMetaPath(),
			Baseline: cfg.GetDBCSourcePath(), // Store baseline DBCs here for later comparison
		},
		Options: OptionConfig{
			UseVersioning: true,
		},
	}

	return &Importer{
		cfg:          cfg,
		dbcCfg:       dbcCfg,
		skipExisting: skipExisting,
	}
}

// Import imports DBC files into the database
func (i *Importer) Import() ([]string, error) {
	// Ensure source directory exists
	if _, err := os.Stat(i.dbcCfg.Paths.Base); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist: %s", i.dbcCfg.Paths.Base)
	}

	// Connect to database
	db, err := openDB(i.dbcCfg.DBC)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	// Import all DBCs
	if err := ImportDBCs(db, i.skipExisting, i.dbcCfg); err != nil {
		return nil, fmt.Errorf("import DBCs: %w", err)
	}

	// List imported tables
	return getImportedTables(db)
}

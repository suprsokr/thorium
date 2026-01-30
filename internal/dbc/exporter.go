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
	// Create DBC tool config from thorium config
	dbcCfg := &Config{
		DBC: DBConfig{
			User:     cfg.Databases.DBC.User,
			Password: cfg.Databases.DBC.Password,
			Host:     cfg.Databases.DBC.Host,
			Port:     cfg.Databases.DBC.Port,
			Name:     cfg.Databases.DBC.Name,
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

	// Export all DBCs
	if err := ExportDBCs(db, e.dbcCfg); err != nil {
		return nil, fmt.Errorf("export DBCs: %w", err)
	}

	// List exported files
	var exported []string
	entries, _ := os.ReadDir(e.dbcCfg.Paths.Export)
	for _, entry := range entries {
		if len(entry.Name()) > 4 && entry.Name()[len(entry.Name())-4:] == ".dbc" {
			exported = append(exported, entry.Name())
		}
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

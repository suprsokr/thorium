// Copyright (c) 2025 Thorium

package dbc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"thorium-cli/internal/config"
	"thorium-cli/internal/mpq"
)

// Extractor extracts DBC files from the WoW client
type Extractor struct {
	cfg *config.Config
}

// NewExtractor creates a new DBC extractor
func NewExtractor(cfg *config.Config) *Extractor {
	return &Extractor{cfg: cfg}
}

// Extract extracts DBC files from client MPQs
func (e *Extractor) Extract() (int, error) {
	clientData := e.cfg.GetClientDataPath()
	outputDir := e.cfg.GetDBCSourcePath()

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return 0, fmt.Errorf("create output dir: %w", err)
	}

	// Find MPQ files containing DBCs (typically locale-*.MPQ and dbc.MPQ)
	mpqFiles, err := findDBCMPQs(clientData)
	if err != nil {
		return 0, fmt.Errorf("find MPQ files: %w", err)
	}

	if len(mpqFiles) == 0 {
		return 0, fmt.Errorf("no MPQ files found in %s", clientData)
	}

	count := 0

	// Extract DBFilesClient from each MPQ
	for _, mpqFile := range mpqFiles {
		fmt.Printf("  Extracting from: %s\n", filepath.Base(mpqFile))

		archive, err := mpq.Open(mpqFile)
		if err != nil {
			fmt.Printf("    Warning: %v\n", err)
			continue
		}

		files, err := archive.Extract("DBFilesClient\\*.dbc", outputDir)
		if err != nil {
			fmt.Printf("    Warning: %v\n", err)
			continue
		}

		count += len(files)
	}

	return count, nil
}

// ExtractToDatabase extracts and imports DBCs into the database
func (e *Extractor) ExtractToDatabase() (int, error) {
	// First extract to files
	count, err := e.Extract()
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	// Then import to database
	exporter := NewExporter(e.cfg)
	imported, err := exporter.Import()
	if err != nil {
		return count, fmt.Errorf("import to database: %w", err)
	}

	return len(imported), nil
}

// findDBCMPQs finds MPQ files that contain DBC data
func findDBCMPQs(dataDir string) ([]string, error) {
	var mpqFiles []string

	// Look for common DBC-containing MPQs
	patterns := []string{
		"dbc.MPQ",
		"locale-*.MPQ",
		"patch*.MPQ",
	}

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".mpq" {
			continue
		}

		// Check if matches any pattern
		for _, pattern := range patterns {
			matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(name))
			if matched {
				mpqFiles = append(mpqFiles, filepath.Join(dataDir, name))
				break
			}
		}
	}

	// Sort by name to ensure consistent order (patches should come last)
	// This is a simple sort - patches override base files
	return mpqFiles, nil
}

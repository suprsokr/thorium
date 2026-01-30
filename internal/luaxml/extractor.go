// Copyright (c) 2025 Thorium

package luaxml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"thorium-cli/internal/config"
	"thorium-cli/internal/mpq"
)

// Extractor extracts LuaXML files from the WoW client
type Extractor struct {
	cfg *config.Config
}

// NewExtractor creates a new LuaXML extractor
func NewExtractor(cfg *config.Config) *Extractor {
	return &Extractor{cfg: cfg}
}

// Extract extracts LuaXML files from client MPQs
func (e *Extractor) Extract() (int, error) {
	clientLocale := e.cfg.GetClientLocalePath()
	outputDir := e.cfg.GetLuaXMLSourcePath()

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return 0, fmt.Errorf("create output dir: %w", err)
	}

	// Find locale MPQ files
	mpqFiles, err := findLocaleMPQs(clientLocale)
	if err != nil {
		return 0, fmt.Errorf("find MPQ files: %w", err)
	}

	if len(mpqFiles) == 0 {
		// Also check main Data directory
		clientData := e.cfg.GetClientDataPath()
		mpqFiles, err = findInterfaceMPQs(clientData)
		if err != nil {
			return 0, fmt.Errorf("find MPQ files: %w", err)
		}
	}

	if len(mpqFiles) == 0 {
		return 0, fmt.Errorf("no MPQ files found in %s", clientLocale)
	}

	count := 0

	// Extract Interface files from each MPQ
	for _, mpqFile := range mpqFiles {
		fmt.Printf("  Extracting from: %s\n", filepath.Base(mpqFile))

		archive, err := mpq.Open(mpqFile)
		if err != nil {
			fmt.Printf("    Warning: %v\n", err)
			continue
		}

		// Extract Interface folder
		files, err := archive.Extract("Interface\\*", outputDir)
		if err != nil {
			fmt.Printf("    Warning: %v\n", err)
			continue
		}

		count += len(files)
	}

	return count, nil
}

// findLocaleMPQs finds MPQ files in the locale directory
func findLocaleMPQs(localeDir string) ([]string, error) {
	var mpqFiles []string

	entries, err := os.ReadDir(localeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".mpq" {
			mpqFiles = append(mpqFiles, filepath.Join(localeDir, entry.Name()))
		}
	}

	return mpqFiles, nil
}

// findInterfaceMPQs finds MPQ files that contain Interface data
func findInterfaceMPQs(dataDir string) ([]string, error) {
	var mpqFiles []string

	patterns := []string{
		"interface.MPQ",
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

		for _, pattern := range patterns {
			matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(name))
			if matched {
				mpqFiles = append(mpqFiles, filepath.Join(dataDir, name))
				break
			}
		}
	}

	return mpqFiles, nil
}

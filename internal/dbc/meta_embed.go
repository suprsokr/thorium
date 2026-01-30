// Copyright (c) 2025 Thorium

package dbc

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

//go:embed meta/*.meta.json
var embeddedMeta embed.FS

// GetEmbeddedMetaFiles returns a list of all embedded meta file names
func GetEmbeddedMetaFiles() ([]string, error) {
	entries, err := embeddedMeta.ReadDir("meta")
	if err != nil {
		return nil, fmt.Errorf("read embedded meta dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".meta.json") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// LoadEmbeddedMeta loads a meta file from embedded FS
func LoadEmbeddedMeta(name string) (*MetaFile, error) {
	// Ensure the name has the right format
	if !strings.HasSuffix(name, ".meta.json") {
		name = name + ".meta.json"
	}

	data, err := embeddedMeta.ReadFile(filepath.Join("meta", name))
	if err != nil {
		return nil, fmt.Errorf("read embedded meta %s: %w", name, err)
	}

	var meta MetaFile
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse meta %s: %w", name, err)
	}

	return &meta, nil
}

// GetMetaForTable returns the meta file for a given table name
func GetMetaForTable(tableName string) (*MetaFile, error) {
	// Try direct match first
	meta, err := LoadEmbeddedMeta(strings.ToLower(tableName))
	if err == nil {
		return meta, nil
	}

	// Scan all metas to find matching table
	files, err := GetEmbeddedMetaFiles()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		meta, err := LoadEmbeddedMeta(file)
		if err != nil {
			continue
		}
		if strings.EqualFold(meta.TableName, tableName) {
			return meta, nil
		}
		// Check filename without extension
		baseName := strings.TrimSuffix(file, ".meta.json")
		if strings.EqualFold(baseName, tableName) {
			return meta, nil
		}
	}

	return nil, fmt.Errorf("no meta found for table: %s", tableName)
}

// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"thorium-cli/internal/config"
	"thorium-cli/internal/dbc"
	"thorium-cli/internal/luaxml"
)

// Extract extracts DBC/LuaXML files from the WoW client or copies to a mod
func Extract(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("extract", flag.ExitOnError)
	extractDBC := fs.Bool("dbc", false, "Extract DBC files")
	extractLuaXML := fs.Bool("luaxml", false, "Extract LuaXML files")
	modName := fs.String("mod", "", "Copy extracted files to a mod's luaxml/ directory")
	destPath := fs.String("dest", "", "Destination path within the mod (e.g., Interface/FrameXML/UnitFrame.lua)")
	filter := fs.String("filter", "", "Only extract files matching this path prefix (e.g., Interface/FrameXML)")
	fs.Parse(args)

	// If --mod is specified, copy specific files to the mod
	if *modName != "" {
		return extractToMod(cfg, *modName, *destPath)
	}

	// Store filter for use by extractors
	if *filter != "" {
		cfg.ExtractFilter = *filter
	}

	// Default to extracting both if neither specified
	if !*extractDBC && !*extractLuaXML {
		*extractDBC = true
		*extractLuaXML = true
	}

	if cfg.WoTLK.Path == "" || cfg.WoTLK.Path == "${WOTLK_PATH}" {
		return fmt.Errorf("wotlk.path not configured in config.json\nSet WOTLK_PATH environment variable or edit config.json")
	}

	fmt.Println("=== Extracting from WoW Client ===")
	fmt.Printf("Client path: %s\n", cfg.WoTLK.Path)
	fmt.Printf("Locale: %s\n", cfg.WoTLK.Locale)
	fmt.Println()

	if *extractDBC {
		fmt.Println("Extracting DBC files...")
		extractor := dbc.NewExtractor(cfg)
		count, err := extractor.Extract()
		if err != nil {
			return fmt.Errorf("extract DBC: %w", err)
		}
		fmt.Printf("  Extracted %d DBC files to %s\n", count, cfg.GetDBCSourcePath())
	}

	if *extractLuaXML {
		fmt.Println("Extracting LuaXML files...")
		extractor := luaxml.NewExtractor(cfg)
		count, err := extractor.Extract()
		if err != nil {
			return fmt.Errorf("extract LuaXML: %w", err)
		}
		fmt.Printf("  Extracted %d files to %s\n", count, cfg.GetLuaXMLSourcePath())
	}

	fmt.Println("\n=== Extraction Complete ===")
	return nil
}

// extractToMod copies files from luaxml_source to a mod's luaxml directory
func extractToMod(cfg *config.Config, modName, destPath string) error {
	if destPath == "" {
		return fmt.Errorf("--dest is required when using --mod\nExample: thorium extract --mod my-mod --dest Interface/FrameXML/UnitFrame.lua")
	}

	// Normalize path separators
	destPath = filepath.FromSlash(destPath)

	// Source: shared/luaxml/luaxml_source/<destPath>
	sourcePath := filepath.Join(cfg.GetLuaXMLSourcePath(), destPath)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s\nMake sure you've run 'thorium extract --luaxml' first", sourcePath)
	}

	// Destination: mods/<modName>/luaxml/<destPath>
	modLuaXMLDir := filepath.Join("mods", modName, "luaxml")
	destFullPath := filepath.Join(modLuaXMLDir, destPath)

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	// Check if source is a file or directory
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if info.IsDir() {
		// Copy entire directory
		count, err := copyDir(sourcePath, destFullPath)
		if err != nil {
			return fmt.Errorf("copy directory: %w", err)
		}
		fmt.Printf("Copied %d files to %s\n", count, destFullPath)
	} else {
		// Copy single file
		if err := copyFile(sourcePath, destFullPath); err != nil {
			return fmt.Errorf("copy file: %w", err)
		}
		fmt.Printf("Copied %s to %s\n", destPath, destFullPath)
	}

	fmt.Println("\nYou can now edit the files in your mod's luaxml/ directory.")
	fmt.Println("Run 'thorium build' to package your changes.")
	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyDir recursively copies a directory
func copyDir(src, dst string) (int, error) {
	count := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Skip non-lua/xml files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".lua" && ext != ".xml" && ext != ".toc" {
			return nil
		}

		if err := copyFile(path, dstPath); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

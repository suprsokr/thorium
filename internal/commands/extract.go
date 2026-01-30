// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"

	"thorium-cli/internal/config"
	"thorium-cli/internal/dbc"
	"thorium-cli/internal/luaxml"
)

// Extract extracts DBC/LuaXML files from the WoW client
func Extract(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("extract", flag.ExitOnError)
	extractDBC := fs.Bool("dbc", false, "Extract DBC files")
	extractLuaXML := fs.Bool("luaxml", false, "Extract LuaXML files")
	fs.Parse(args)

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

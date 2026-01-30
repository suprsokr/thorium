// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"

	"thorium-cli/internal/config"
	"thorium-cli/internal/dbc"
)

// Export exports modified DBCs from the database
func Export(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	_ = fs.Bool("dbc", true, "Export DBC files (default: true)")
	fs.Parse(args)

	fmt.Println("=== Exporting Modified DBCs ===")
	fmt.Println()

	exporter := dbc.NewExporter(cfg)
	tables, err := exporter.Export()
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		fmt.Println("No modified DBCs to export.")
	} else {
		fmt.Printf("\nExported %d DBC file(s)\n", len(tables))
	}

	fmt.Println("\n=== Export Complete ===")
	return nil
}

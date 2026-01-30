// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"

	"thorium-cli/internal/config"
	"thorium-cli/internal/mpq"
)

// Package packages files into MPQ archives
func Package(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("package", flag.ExitOnError)
	dbcFlag := fs.Bool("dbc", false, "Package DBC files only")
	luaxmlFlag := fs.Bool("luaxml", false, "Package LuaXML files only")
	client := fs.Bool("client", false, "Package for client (both DBC and LuaXML)")
	server := fs.Bool("server", false, "Copy DBCs to server only")
	fs.Parse(args)

	// Default to all if nothing specified
	if !*dbcFlag && !*luaxmlFlag && !*client && !*server {
		*client = true
		*server = true
	}

	if *client {
		*dbcFlag = true
		*luaxmlFlag = true
	}

	fmt.Println("=== Packaging Files ===")
	fmt.Println()

	builder := mpq.NewBuilder(cfg)

	// Package DBCs
	if *dbcFlag {
		fmt.Println("Packaging DBC files...")
		count, err := builder.PackageDBCs()
		if err != nil {
			return fmt.Errorf("package DBCs: %w", err)
		}
		if count > 0 {
			fmt.Printf("  Packaged %d DBC file(s) into %s\n", count, cfg.Output.DBCMPQ)
		} else {
			fmt.Println("  No modified DBC files to package")
		}
	}

	// Copy to server
	if *server && cfg.Server.DBCPath != "" {
		fmt.Println("Copying DBCs to server...")
		count, err := builder.CopyToServer()
		if err != nil {
			return fmt.Errorf("copy to server: %w", err)
		}
		if count > 0 {
			fmt.Printf("  Copied %d DBC file(s) to %s\n", count, cfg.Server.DBCPath)
		}
	}

	// Package LuaXML
	if *luaxmlFlag {
		fmt.Println("Packaging LuaXML files...")
		count, err := builder.PackageLuaXML()
		if err != nil {
			return fmt.Errorf("package LuaXML: %w", err)
		}
		if count > 0 {
			mpqName := cfg.GetMPQName(cfg.Output.LuaXMLMPQ)
			fmt.Printf("  Packaged %d file(s) into %s\n", count, mpqName)
		} else {
			fmt.Println("  No modified LuaXML files to package")
		}
	}

	fmt.Println("\n=== Packaging Complete ===")
	return nil
}

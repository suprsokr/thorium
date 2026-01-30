// Copyright (c) 2025 Thorium
// Thorium is a modding framework for TrinityCore WoW servers.

package main

import (
	"fmt"
	"os"
	"strings"

	"thorium-cli/internal/commands"
	"thorium-cli/internal/config"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	// Parse global flags
	configPath := "./config.json"
	var argsWithoutGlobal []string

	i := 1
	for i < len(os.Args) {
		arg := os.Args[i]

		// Handle --config flag
		if arg == "--config" || arg == "-c" {
			if i+1 < len(os.Args) {
				configPath = os.Args[i+1]
				i += 2
				continue
			}
			fmt.Println("Error: --config requires a value")
			os.Exit(1)
		}
		if strings.HasPrefix(arg, "--config=") {
			configPath = strings.TrimPrefix(arg, "--config=")
			i++
			continue
		}

		argsWithoutGlobal = append(argsWithoutGlobal, arg)
		i++
	}

	if len(argsWithoutGlobal) == 0 {
		printUsage()
		os.Exit(0)
	}

	cmd := argsWithoutGlobal[0]
	subArgs := argsWithoutGlobal[1:]

	// Handle help and version before loading config
	switch cmd {
	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)
	case "version", "-v", "--version":
		fmt.Printf("thorium version %s\n", version)
		os.Exit(0)
	case "init":
		if err := commands.Init(configPath, subArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Load config for other commands
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'thorium init' to create a workspace.\n")
		os.Exit(1)
	}

	// Execute command
	var cmdErr error
	switch cmd {
	case "build":
		cmdErr = commands.Build(cfg, subArgs)
	case "apply":
		cmdErr = commands.Apply(cfg, subArgs)
	case "rollback":
		cmdErr = commands.Rollback(cfg, subArgs)
	case "export":
		cmdErr = commands.Export(cfg, subArgs)
	case "package":
		cmdErr = commands.Package(cfg, subArgs)
	case "status":
		cmdErr = commands.Status(cfg, subArgs)
	case "create-mod":
		cmdErr = commands.CreateMod(cfg, subArgs)
	case "create-migration":
		cmdErr = commands.CreateMigration(cfg, subArgs)
	case "extract":
		cmdErr = commands.Extract(cfg, subArgs)
	case "patch":
		cmdErr = commands.Patch(cfg, subArgs)
	case "dist":
		cmdErr = commands.Dist(cfg, subArgs)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if cmdErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Thorium - TrinityCore Modding Framework

Usage: thorium [global-flags] <command> [command-flags]

Global Flags:
  --config <path>    Path to config.json (default: ./config.json)

Commands:
  init               Initialize a new Thorium workspace
  create-mod <name>  Create a new mod with standard structure
  create-migration   Create a new SQL migration in a mod
  build              Full build: apply migrations, export DBCs, package MPQs
  apply              Apply SQL migrations for mods
  rollback           Rollback SQL migrations
  export             Export modified DBCs from database
  extract            Extract DBC/LuaXML from client MPQs
  package            Package files into MPQ archives
  dist               Create distributable zip with client MPQs and server SQL
  patch              Apply patches to WoW client executable
  status             Show status of migrations and mods
  version            Show version information
  help               Show this help message

Examples:
  thorium init                          # Create new workspace in current directory
  thorium create-mod my-mod             # Create a new mod
  thorium create-migration --mod my-mod --db world add_custom_npc
  thorium create-migration --mod my-mod --db dbc add_custom_item
  thorium build                         # Full build all mods
  thorium build --mod custom-weapon     # Build specific mod
  thorium apply --mod custom-weapon     # Apply migrations only
  thorium export                        # Export DBCs only
  thorium extract --dbc                 # Extract DBCs from client
  thorium package --client              # Package client MPQs only
  thorium dist                          # Create distributable zip of all mods
  thorium dist --mod my-mod             # Create zip for specific mod
  thorium status                        # Show migration status

Environment Variables:
  WOTLK_PATH           Path to WoW 3.3.5 client directory
  TC_SERVER_PATH       Path to TrinityCore server directory
  MYSQL_HOST           MySQL host (default: 127.0.0.1)
  MYSQL_PORT           MySQL port (default: 3306)
`)
}

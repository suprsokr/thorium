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

const version = "1.5.0"

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
	case "rollback":
		cmdErr = commands.Rollback(cfg, subArgs)
	case "export":
		cmdErr = commands.Export(cfg, subArgs)
	case "status":
		cmdErr = commands.Status(cfg, subArgs)
	case "get":
		cmdErr = commands.Get(cfg, subArgs)
	case "search":
		cmdErr = commands.Search(cfg, subArgs)
	case "create-mod":
		cmdErr = commands.CreateMod(cfg, subArgs)
	case "create-migration":
		cmdErr = commands.CreateMigration(cfg, subArgs)
	case "create-script":
		cmdErr = commands.CreateScript(cfg, subArgs)
	case "create-addon":
		cmdErr = commands.CreateAddon(cfg, subArgs)
	case "extract":
		cmdErr = commands.Extract(cfg, subArgs)
	case "import":
		cmdErr = commands.Import(cfg, subArgs)
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
	fmt.Println(`Thorium - Unified Workflow for WoW server and client mods.

Usage: thorium [global-flags] <command> [command-flags]

Global Flags:
  --config <path>    Path to config.json (default: ./config.json)

Commands:
  build              Full build: migrations, patches, DBCs, scripts, assets, MPQs
  create-addon       Create a new WoW addon in a mod's luaxml folder
  create-migration   Create a new SQL migration in a mod
  create-mod <name>  Create a new mod with standard structure
  create-script      Create a new TrinityCore script in a mod
  dist               Create player distribution zip (client MPQs + exe)
  export             Export modified DBCs from database
  extract            Extract DBC/LuaXML from client MPQs
  get <url>          Install a mod from a GitHub repository
  help               Show this help message
  import             Import DBC files into database
  init               Initialize a new Thorium workspace
  rollback           Rollback SQL migrations
  search [query]     Search the mod registry for available mods
  status             Show status of migrations and mods
  version            Show version information

Mod Structure:
  mods/<mod>/dbc_sql/         SQL migrations for DBC database
  mods/<mod>/world_sql/       SQL migrations for World database
  mods/<mod>/scripts/         TrinityCore C++ scripts
  mods/<mod>/server-patches/  TrinityCore source patches (.patch files)
  mods/<mod>/binary-edits/    Client binary patches (.json files)
  mods/<mod>/assets/          Files to copy to client (with config.json)
  mods/<mod>/luaxml/          Client-side Lua/XML modifications
  shared/                     Shared resources (extracted client data, migration tracking, applied patches)`)
}

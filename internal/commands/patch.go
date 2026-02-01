// Copyright (c) 2025 Thorium

package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"thorium-cli/internal/config"
	"thorium-cli/internal/patcher"
)

// Patch applies patches to the WoW client
func Patch(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("patch", flag.ExitOnError)
	listPatches := fs.Bool("list", false, "List available patches")
	dryRun := fs.Bool("dry-run", false, "Show what would be patched without applying")
	verbose := fs.Bool("verbose", false, "Verbose output")
	restore := fs.Bool("restore", false, "Restore from backup")
	patchesFlag := fs.String("patches", "", "Comma-separated list of patch names to apply (default: all)")
	fs.Parse(args)

	var wowExe string

	// Check if a direct path was provided as a positional argument
	if fs.NArg() > 0 {
		// Direct path to exe provided - use it without requiring config.json
		wowExe = fs.Arg(0)
		if _, err := os.Stat(wowExe); os.IsNotExist(err) {
			return fmt.Errorf("WoW executable not found: %s", wowExe)
		}
	} else {
		// No direct path - use config.json
		if cfg == nil {
			return fmt.Errorf("no WoW executable path provided and no config.json found\nUsage: thorium patch [/path/to/WoW.exe]")
		}
		if cfg.WoTLK.Path == "" || cfg.WoTLK.Path == "${WOTLK_PATH}" {
			return fmt.Errorf("wotlk.path not configured in config.json\nAlternatively, provide path directly: thorium patch /path/to/WoW.exe")
		}

		// Find WoW executable in the configured directory
		wowExe = findWoWExecutable(cfg.WoTLK.Path)
		if wowExe == "" {
			return fmt.Errorf("WoW executable not found in %s", cfg.WoTLK.Path)
		}
	}

	fmt.Println("=== Client Patcher ===")
	fmt.Printf("Client: %s\n", wowExe)
	fmt.Println()

	// List available patches
	if *listPatches {
		fmt.Println("Available patches:")
		patches := patcher.GetClientPatches()
		for _, cat := range patches {
			fmt.Printf("  - %s\n", cat.Name)
			fmt.Printf("    %s\n", cat.Description)
		}
		return nil
	}

	// Restore from backup
	if *restore {
		if err := patcher.RestoreFromBackup(wowExe, ""); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
		fmt.Println("✓ Restored successfully")
		return nil
	}

	// Dry run just shows what would happen
	if *dryRun {
		patches := patcher.GetClientPatches()
		fmt.Println("Would apply these patches:")
		for _, cat := range patches {
			fmt.Printf("  - %s (%d byte edits)\n", cat.Name, len(cat.Patches))
		}
		return nil
	}

	// Parse selected patches
	var selectedPatches []string
	if *patchesFlag != "" {
		// Parse comma-separated list
		for _, p := range splitAndTrim(*patchesFlag, ",") {
			selectedPatches = append(selectedPatches, p)
		}
	}

	// Apply patches
	opts := patcher.PatchOptions{
		WowExePath:      wowExe,
		BackupPath:      wowExe + ".clean",
		OutputPath:      wowExe,
		Verbose:         *verbose,
		SelectedPatches: selectedPatches,
	}

	if err := patcher.ApplyPatches(opts); err != nil {
		return fmt.Errorf("apply patches: %w", err)
	}

	fmt.Println("\n✓ Patches applied successfully")
	return nil
}

// splitAndTrim splits a string by delimiter and trims each part
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := []string{}
	for _, p := range splitString(s, sep) {
		trimmed := trimString(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	current := ""
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// findWoWExecutable finds the WoW executable in the client directory
func findWoWExecutable(clientPath string) string {
	candidates := []string{
		"WoW.exe",
		"Wow.exe",
		"wow.exe",
		"World of Warcraft.app/Contents/MacOS/World of Warcraft",
		"WoW",
		"wow",
	}

	for _, candidate := range candidates {
		path := filepath.Join(clientPath, candidate)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

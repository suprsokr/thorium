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
	fs.Parse(args)

	if cfg.WoTLK.Path == "" || cfg.WoTLK.Path == "${WOTLK_PATH}" {
		return fmt.Errorf("wotlk.path not configured in config.json")
	}

	// Find WoW executable
	wowExe := findWoWExecutable(cfg.WoTLK.Path)
	if wowExe == "" {
		return fmt.Errorf("WoW executable not found in %s", cfg.WoTLK.Path)
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

	// Apply all patches
	opts := patcher.PatchOptions{
		WowExePath: wowExe,
		BackupPath: wowExe + ".clean",
		OutputPath: wowExe,
		Verbose:    *verbose,
	}

	if err := patcher.ApplyPatches(opts); err != nil {
		return fmt.Errorf("apply patches: %w", err)
	}

	fmt.Println("\n✓ Patches applied successfully")
	return nil
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

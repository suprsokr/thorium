// Copyright (c) 2025 Thorium

package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"thorium-cli/internal/config"
)

// ServerPatch represents a server patch that can be applied to TrinityCore
type ServerPatch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PatchFile   string `json:"patch_file"`
}

// AppliedPatch tracks a patch that has been applied
type AppliedPatch struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	AppliedAt string `json:"applied_at"`
	AppliedBy string `json:"appliedBy"`
}

// PatchTracker stores the state of applied patches
type PatchTracker struct {
	Applied []AppliedPatch `json:"applied"`
}

const (
	patchTrackerFile = "server_patches_applied.json"
	thoriumVersion   = "1.5.0"
)

// Available server patches (embedded in thorium)
var availablePatches = []ServerPatch{
	{
		Name:        "custom-packets",
		Description: "Adds custom packet support (opcode 0x51F) for addon-server communication",
		PatchFile:   "custom-packets/custom-packets.patch",
	},
}

// PatchServer manages TrinityCore source patches
func PatchServer(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("patch-server", flag.ExitOnError)
	listPatches := fs.Bool("list", false, "List available server patches")
	statusFlag := fs.Bool("status", false, "Show status of applied patches")
	dryRun := fs.Bool("dry-run", false, "Show what would be done without applying")
	verbose := fs.Bool("verbose", false, "Verbose output")
	fs.Parse(args)

	// List available patches
	if *listPatches {
		fmt.Println("Available server patches:")
		fmt.Println()
		for _, p := range availablePatches {
			fmt.Printf("  %s\n", p.Name)
			fmt.Printf("    %s\n", p.Description)
			fmt.Println()
		}
		fmt.Println("Usage:")
		fmt.Println("  thorium patch-server apply <patch-name>")
		fmt.Println("  thorium patch-server revert <patch-name>")
		fmt.Println("  thorium patch-server --status")
		return nil
	}

	// Get TrinityCore source path
	tcPath := ""
	if cfg != nil && cfg.TrinityCore.SourcePath != "" && cfg.TrinityCore.SourcePath != "${TC_SOURCE_PATH}" {
		tcPath = cfg.TrinityCore.SourcePath
	}
	if tcPath == "" {
		tcPath = os.Getenv("TC_SOURCE_PATH")
	}
	if tcPath == "" {
		return fmt.Errorf("TrinityCore source path not configured\nSet trinitycore.source_path in config.json or TC_SOURCE_PATH environment variable")
	}

	// Verify it's a valid TC directory
	if !isTrinityCorePath(tcPath) {
		return fmt.Errorf("invalid TrinityCore path: %s\nExpected to find src/server/game/ directory", tcPath)
	}

	// Get workspace root for tracking (may be empty if no config)
	workspaceRoot := ""
	if cfg != nil {
		workspaceRoot = cfg.WorkspaceRoot
	}

	// Show status
	if *statusFlag {
		return showPatchStatus(tcPath, workspaceRoot)
	}

	// Parse subcommand
	if fs.NArg() < 1 {
		fmt.Println("Usage:")
		fmt.Println("  thorium patch-server --list")
		fmt.Println("  thorium patch-server --status")
		fmt.Println("  thorium patch-server apply <patch-name>")
		fmt.Println("  thorium patch-server revert <patch-name>")
		return nil
	}

	subCmd := fs.Arg(0)
	switch subCmd {
	case "apply":
		if fs.NArg() < 2 {
			return fmt.Errorf("usage: thorium patch-server apply <patch-name>")
		}
		patchName := fs.Arg(1)
		return applyServerPatch(tcPath, workspaceRoot, patchName, *dryRun, *verbose)

	case "revert":
		if fs.NArg() < 2 {
			return fmt.Errorf("usage: thorium patch-server revert <patch-name>")
		}
		patchName := fs.Arg(1)
		return revertServerPatch(tcPath, workspaceRoot, patchName, *dryRun, *verbose)

	default:
		return fmt.Errorf("unknown subcommand: %s\nUse 'apply' or 'revert'", subCmd)
	}
}

func isTrinityCorePath(path string) bool {
	// Check for typical TC directory structure
	gameDir := filepath.Join(path, "src", "server", "game")
	if _, err := os.Stat(gameDir); err == nil {
		return true
	}
	return false
}

func showPatchStatus(tcPath, workspaceRoot string) error {
	fmt.Println("=== Server Patch Status ===")
	fmt.Printf("TrinityCore: %s\n", tcPath)
	if workspaceRoot != "" {
		fmt.Printf("Workspace: %s\n", workspaceRoot)
	}
	fmt.Println()

	tracker, err := loadPatchTracker(workspaceRoot)
	if err != nil {
		fmt.Println("No patches tracked (first time running patch-server)")
		return nil
	}

	if len(tracker.Applied) == 0 {
		fmt.Println("No patches applied")
		return nil
	}

	fmt.Println("Applied patches:")
	for _, p := range tracker.Applied {
		fmt.Printf("  ✓ %s\n", p.Name)
		fmt.Printf("    Applied: %s\n", p.AppliedAt)
		fmt.Printf("    By: %s\n", p.AppliedBy)
	}

	return nil
}

func findPatch(name string) *ServerPatch {
	for _, p := range availablePatches {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

func applyServerPatch(tcPath, workspaceRoot, patchName string, dryRun, verbose bool) error {
	patch := findPatch(patchName)
	if patch == nil {
		return fmt.Errorf("unknown patch: %s\nUse 'thorium patch-server --list' to see available patches", patchName)
	}

	// Check if already applied
	tracker, _ := loadPatchTracker(workspaceRoot)
	for _, p := range tracker.Applied {
		if p.Name == patchName {
			return fmt.Errorf("patch '%s' is already applied\nUse 'thorium patch-server revert %s' to remove it first", patchName, patchName)
		}
	}

	// Get the patch file path (embedded in thorium binary directory or assets)
	patchFile, err := getPatchFilePath(patch.PatchFile)
	if err != nil {
		return fmt.Errorf("patch file not found: %w", err)
	}

	fmt.Printf("=== Applying Server Patch: %s ===\n", patchName)
	fmt.Printf("TrinityCore: %s\n", tcPath)
	fmt.Printf("Patch: %s\n\n", patchFile)

	if dryRun {
		fmt.Println("[DRY RUN] Would apply patch with: git apply --check")
		// Check if patch would apply cleanly
		cmd := exec.Command("git", "apply", "--check", patchFile)
		cmd.Dir = tcPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Patch would NOT apply cleanly:\n%s\n", string(output))
			return fmt.Errorf("patch check failed")
		}
		fmt.Println("Patch would apply cleanly")
		return nil
	}

	// Apply the patch
	if verbose {
		fmt.Println("Running: git apply " + patchFile)
	}

	cmd := exec.Command("git", "apply", patchFile)
	cmd.Dir = tcPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply patch: %s\n%s", err, string(output))
	}

	// Track the applied patch (only if we have a workspace)
	if workspaceRoot != "" {
		tracker.Applied = append(tracker.Applied, AppliedPatch{
			Name:      patchName,
			Version:   "1.0.0",
			AppliedAt: time.Now().Format(time.RFC3339),
			AppliedBy: "thorium " + thoriumVersion,
		})
		if err := savePatchTracker(workspaceRoot, tracker); err != nil {
			fmt.Printf("Warning: could not save patch tracker: %v\n", err)
		}
	} else {
		fmt.Println("Note: No workspace found, patch application not tracked.")
	}

	fmt.Println("✓ Patch applied successfully")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Rebuild TrinityCore:")
	fmt.Println("     cd " + filepath.Join(tcPath, "build"))
	fmt.Println("     make -j$(nproc)")
	fmt.Println()
	fmt.Println("  2. Restart your server")
	fmt.Println()
	fmt.Println("To revert this patch:")
	fmt.Printf("  thorium patch-server revert %s\n", patchName)

	return nil
}

func revertServerPatch(tcPath, workspaceRoot, patchName string, dryRun, verbose bool) error {
	patch := findPatch(patchName)
	if patch == nil {
		return fmt.Errorf("unknown patch: %s", patchName)
	}

	// Check if actually applied (only if we have tracking)
	tracker, _ := loadPatchTracker(workspaceRoot)
	if workspaceRoot != "" {
		found := false
		for _, p := range tracker.Applied {
			if p.Name == patchName {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Warning: patch '%s' is not tracked as applied (attempting revert anyway)\n", patchName)
		}
	}

	patchFile, err := getPatchFilePath(patch.PatchFile)
	if err != nil {
		return fmt.Errorf("patch file not found: %w", err)
	}

	fmt.Printf("=== Reverting Server Patch: %s ===\n", patchName)
	fmt.Printf("TrinityCore: %s\n\n", tcPath)

	if dryRun {
		fmt.Println("[DRY RUN] Would revert with: git apply -R --check")
		cmd := exec.Command("git", "apply", "-R", "--check", patchFile)
		cmd.Dir = tcPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Patch would NOT revert cleanly:\n%s\n", string(output))
			return fmt.Errorf("revert check failed")
		}
		fmt.Println("Patch would revert cleanly")
		return nil
	}

	// Revert the patch
	if verbose {
		fmt.Println("Running: git apply -R " + patchFile)
	}

	cmd := exec.Command("git", "apply", "-R", patchFile)
	cmd.Dir = tcPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to revert patch: %s\n%s", err, string(output))
	}

	// Update tracker (only if we have a workspace)
	if workspaceRoot != "" {
		newApplied := []AppliedPatch{}
		for _, p := range tracker.Applied {
			if p.Name != patchName {
				newApplied = append(newApplied, p)
			}
		}
		tracker.Applied = newApplied
		if err := savePatchTracker(workspaceRoot, tracker); err != nil {
			fmt.Printf("Warning: could not save patch tracker: %v\n", err)
		}
	}

	fmt.Println("✓ Patch reverted successfully")
	fmt.Println()
	fmt.Println("Don't forget to rebuild TrinityCore:")
	fmt.Println("  cd " + filepath.Join(tcPath, "build"))
	fmt.Println("  make -j$(nproc)")

	return nil
}

func getTrackerPath(workspaceRoot string) string {
	if workspaceRoot == "" {
		return ""
	}
	return filepath.Join(workspaceRoot, "shared", patchTrackerFile)
}

func loadPatchTracker(workspaceRoot string) (PatchTracker, error) {
	trackerPath := getTrackerPath(workspaceRoot)
	if trackerPath == "" {
		return PatchTracker{}, fmt.Errorf("no workspace")
	}
	data, err := os.ReadFile(trackerPath)
	if err != nil {
		return PatchTracker{}, err
	}
	var tracker PatchTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return PatchTracker{}, err
	}
	return tracker, nil
}

func savePatchTracker(workspaceRoot string, tracker PatchTracker) error {
	trackerPath := getTrackerPath(workspaceRoot)
	if trackerPath == "" {
		return fmt.Errorf("no workspace to save tracker")
	}
	// Ensure shared directory exists
	sharedDir := filepath.Dir(trackerPath)
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tracker, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(trackerPath, data, 0644)
}

func getPatchFilePath(relativePath string) (string, error) {
	// Try to find the patch file
	// 1. Check if running from source (development)
	// 2. Check relative to executable
	// 3. Check in known locations

	candidates := []string{}

	// Development: relative to working directory
	candidates = append(candidates, filepath.Join("assets", "server-patches", relativePath))

	// Relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(exeDir, "assets", "server-patches", relativePath))
		candidates = append(candidates, filepath.Join(exeDir, "server-patches", relativePath))
	}

	// Common install locations
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		candidates = append(candidates, filepath.Join(homeDir, ".thorium", "server-patches", relativePath))
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}

	// For embedded patches, we might need to extract them
	// For now, return error with helpful message
	return "", fmt.Errorf("patch file not found: %s\nTried: %s", relativePath, strings.Join(candidates, ", "))
}

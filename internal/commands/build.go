// Copyright (c) 2025 Thorium

package commands

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"thorium-cli/internal/config"
	"thorium-cli/internal/dbc"
	"thorium-cli/internal/mpq"
	"thorium-cli/internal/scripts"
)

// Build performs a full build: apply migrations, patches, export DBCs, overlay LuaXML, package MPQs
func Build(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	modName := fs.String("mod", "", "Build specific mod only")
	skipDBCSQL := fs.Bool("skip-dbc-sql", false, "Skip DBC SQL migrations")
	skipWorldSQL := fs.Bool("skip-world-sql", false, "Skip World SQL migrations")
	skipBinaryEdits := fs.Bool("skip-binary-edits", false, "Skip binary edits")
	skipServerPatches := fs.Bool("skip-server-patches", false, "Skip server patches")
	skipAssets := fs.Bool("skip-assets", false, "Skip copying assets")
	skipExportDBC := fs.Bool("skip-export-dbc", false, "Skip DBC export")
	skipLuaXML := fs.Bool("skip-luaxml", false, "Skip LuaXML processing")
	skipScripts := fs.Bool("skip-scripts", false, "Skip script deployment")
	skipPackage := fs.Bool("skip-package", false, "Skip MPQ packaging")
	skipServerDBC := fs.Bool("skip-server-dbc", false, "Skip copying DBCs to server")
	force := fs.Bool("force", false, "Force reapply all tracked items (migrations, binary-edits, server-patches, assets, scripts)")
	forceDBCSQL := fs.Bool("force-dbc-sql", false, "Force reapply DBC SQL migrations even if already applied")
	forceWorldSQL := fs.Bool("force-world-sql", false, "Force reapply World SQL migrations even if already applied")
	forceBinaryEdits := fs.Bool("force-binary-edits", false, "Force reapply binary edits even if already applied")
	forceServerPatches := fs.Bool("force-server-patches", false, "Force reapply server patches even if already applied")
	forceAssets := fs.Bool("force-assets", false, "Force recopy assets even if unchanged")
	forceScripts := fs.Bool("force-scripts", false, "Force redeploy scripts even if unchanged")
	fs.Parse(args)

	// Parse positional arguments for component selection
	components := fs.Args()
	validComponents := map[string]bool{
		"dbc":          true,
		"dbc_sql":      true,
		"world_sql":    true,
		"binary":       true,
		"server-patches": true,
		"assets":       true,
		"luaxml":       true,
		"scripts":      true,
	}
	
	componentSet := make(map[string]bool)
	for _, comp := range components {
		compLower := strings.ToLower(comp)
		if !validComponents[compLower] {
			return fmt.Errorf("unknown component: %s\nValid components: dbc, dbc_sql, world_sql, binary, server-patches, assets, luaxml, scripts", comp)
		}
		componentSet[compLower] = true
	}
	buildAll := len(componentSet) == 0

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        Thorium Build System              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// Get list of mods
	mods, err := listMods(cfg)
	if err != nil {
		return fmt.Errorf("list mods: %w", err)
	}

	// Filter to specific mod if requested
	if *modName != "" {
		found := false
		for _, m := range mods {
			if m == *modName {
				mods = []string{m}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("mod not found: %s", *modName)
		}
	}

	if len(mods) == 0 {
		fmt.Println("No mods found in", cfg.GetModsPath())
		return nil
	}

	fmt.Printf("Building %d mod(s): %v\n\n", len(mods), mods)

	// Determine which components to build
	// Note: "dbc" implies DBC SQL migrations, export, and packaging
	buildDBCSQL := buildAll || componentSet["dbc"] || componentSet["dbc_sql"]
	buildWorldSQL := buildAll || componentSet["world_sql"]
	buildBinary := buildAll || componentSet["binary"]
	buildServerPatches := buildAll || componentSet["server-patches"]
	buildAssets := buildAll || componentSet["assets"]
	buildDBCExport := buildAll || componentSet["dbc"]
	buildLuaXML := buildAll || componentSet["luaxml"]
	buildScripts := buildAll || componentSet["scripts"]
	buildPackage := buildAll || componentSet["dbc"] || componentSet["luaxml"]

	if !buildAll {
		fmt.Printf("Building components: %v\n\n", components)
	}

	// Step 1: Apply SQL migrations
	shouldRunDBCMigrations := buildDBCSQL && !*skipDBCSQL
	shouldRunWorldMigrations := buildWorldSQL && !*skipWorldSQL
	
	if shouldRunDBCMigrations || shouldRunWorldMigrations {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 1: Applying SQL Migrations         │")
		fmt.Println("└──────────────────────────────────────────┘")

		migrationsFound := false
		dbcMigrationsFound := false
		
		for _, mod := range mods {
			// Check if mod has any migration directories with files
			if shouldRunDBCMigrations {
				dbcDir := filepath.Join(cfg.GetModsPath(), mod, "dbc_sql")
				if entries, _ := os.ReadDir(dbcDir); len(entries) > 0 {
					migrationsFound = true
					dbcMigrationsFound = true
				}
			}
			if shouldRunWorldMigrations {
				worldDir := filepath.Join(cfg.GetModsPath(), mod, "world_sql")
				if entries, _ := os.ReadDir(worldDir); len(entries) > 0 {
					migrationsFound = true
				}
			}
		}
		
		// Check DBC setup only if we have DBC migrations
		if dbcMigrationsFound {
			if err := checkDBCDatabasesSetup(cfg); err != nil {
				printDBCSetupInstructions()
				return fmt.Errorf("DBC databases not initialized")
			}
		}
		
		for _, mod := range mods {
			if shouldRunDBCMigrations {
				forceDBC := *force || *forceDBCSQL
				if err := applyMigrationsWithForce(cfg, mod, "dbc", forceDBC); err != nil {
					return fmt.Errorf("apply dbc migrations for %s: %w", mod, err)
				}
			}
			if shouldRunWorldMigrations {
				forceWorld := *force || *forceWorldSQL
				if err := applyMigrationsWithForce(cfg, mod, "world", forceWorld); err != nil {
					return fmt.Errorf("apply world migrations for %s: %w", mod, err)
				}
			}
		}
		if !migrationsFound {
			fmt.Println("  No new SQL migrations to apply")
		}
		fmt.Println()
	}

	// Step 2: Apply binary edits to client
	if buildBinary && !*skipBinaryEdits && cfg.WoTLK.Path != "" {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 2: Applying Binary Edits           │")
		fmt.Println("└──────────────────────────────────────────┘")

		editsApplied, err := applyModBinaryEdits(cfg, mods, *force || *forceBinaryEdits)
		if err != nil {
			return fmt.Errorf("apply binary edits: %w", err)
		}
		if editsApplied > 0 {
			fmt.Printf("Applied %d new binary edit(s)\n", editsApplied)
		} else {
			fmt.Println("  No new binary edits to apply")
		}
		fmt.Println()
	}

	// Step 3: Apply server patches from mods
	if buildServerPatches && !*skipServerPatches && cfg.TrinityCore.SourcePath != "" {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 3: Applying Server Patches         │")
		fmt.Println("└──────────────────────────────────────────┘")

		patchesApplied, err := applyModServerPatches(cfg, mods, *force || *forceServerPatches)
		if err != nil {
			return fmt.Errorf("apply server patches: %w", err)
		}
		if patchesApplied > 0 {
			fmt.Printf("Applied %d new server patch(es)\n", patchesApplied)
			fmt.Println("  Note: Rebuild TrinityCore to apply changes")
		} else {
			fmt.Println("  No new server patches to apply")
		}
		fmt.Println()
	}

	// Step 4: Copy mod assets to client
	if buildAssets && !*skipAssets && cfg.WoTLK.Path != "" {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 4: Copying Mod Assets              │")
		fmt.Println("└──────────────────────────────────────────┘")

		assetsCopied, err := copyModAssets(cfg, mods, *force || *forceAssets)
		if err != nil {
			return fmt.Errorf("copy mod assets: %w", err)
		}
		if assetsCopied > 0 {
			fmt.Printf("Copied %d asset(s) to client\n", assetsCopied)
		} else {
			fmt.Println("  No new/changed assets to copy")
		}
		fmt.Println()
	}

	// Step 5: Export modified DBCs
	if buildDBCExport && !*skipExportDBC {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 5: Exporting Modified DBCs         │")
		fmt.Println("└──────────────────────────────────────────┘")

		exporter := dbc.NewExporter(cfg)
		tables, err := exporter.Export()
		if err != nil {
			return fmt.Errorf("export DBCs: %w", err)
		}
		if len(tables) > 0 {
			fmt.Printf("Exported %d DBC table(s)\n", len(tables))
		} else {
			fmt.Println("  No modified DBCs found")
		}
		fmt.Println()
	}

	// Step 6: Check for LuaXML modifications
	var allModifiedLuaXML []mpq.ModifiedLuaXMLFile
	if buildLuaXML && !*skipLuaXML {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 6: Checking LuaXML Modifications   │")
		fmt.Println("└──────────────────────────────────────────┘")

		// Collect modified LuaXML files from all mods
		for _, mod := range mods {
		modFiles, err := findModifiedLuaXMLInMod(cfg, mod)
		if err != nil {
			return fmt.Errorf("check luaxml for %s: %w", mod, err)
		}
			if len(modFiles) > 0 {
				fmt.Printf("[%s] Found %d modified LuaXML file(s)\n", mod, len(modFiles))
				// Convert to mpq type
				for _, f := range modFiles {
					allModifiedLuaXML = append(allModifiedLuaXML, mpq.ModifiedLuaXMLFile{
						ModName:  f.ModName,
						FilePath: f.FilePath,
						RelPath:  f.RelPath,
					})
				}
			}
		}
		if len(allModifiedLuaXML) == 0 {
			fmt.Println("  No LuaXML modifications found in mods")
		}
		fmt.Println()
	}

	// Step 7: Deploy Scripts to TrinityCore
	if buildScripts && !*skipScripts && cfg.TrinityCore.ScriptsPath != "" {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 7: Deploying Scripts               │")
		fmt.Println("└──────────────────────────────────────────┘")

		if _, err := scripts.DeployScripts(cfg, mods, *force || *forceScripts); err != nil {
			return fmt.Errorf("deploy scripts: %w", err)
		}
		fmt.Println()
	}

	// Step 8: Package and distribute
	var dbcCount, luaxmlCount int
	if buildPackage && !*skipPackage {
		fmt.Println("┌──────────────────────────────────────────┐")
		fmt.Println("│  Step 8: Packaging and Distributing      │")
		fmt.Println("└──────────────────────────────────────────┘")

		builder := mpq.NewBuilder(cfg)

		// Copy to server (only if building DBCs)
		if buildDBCExport && !*skipServerDBC && cfg.Server.DBCPath != "" {
			count, err := builder.CopyToServer()
			if err != nil {
				return fmt.Errorf("copy to server: %w", err)
			}
			if count > 0 {
				fmt.Printf("Copied %d DBC file(s) to server\n", count)
			}
		}

		// Package DBC MPQ (only if building DBCs)
		if buildDBCExport {
			count, err := builder.PackageDBCs()
			if err != nil {
				return fmt.Errorf("package DBCs: %w", err)
			}
			dbcCount = count
		}

		// Package LuaXML MPQ from modified files (only if building LuaXML)
		if buildLuaXML && len(allModifiedLuaXML) > 0 {
			count, err := builder.PackageLuaXMLFromMods(allModifiedLuaXML)
			if err != nil {
				return fmt.Errorf("package LuaXML: %w", err)
			}
			luaxmlCount = count
		}

		// Print a nice summary of what was packaged
		if dbcCount > 0 && luaxmlCount > 0 {
			fmt.Printf("Created MPQ with DBC and LuaXML files and copied to client\n")
		} else if dbcCount > 0 {
			fmt.Printf("Created MPQ with DBC files and copied to client\n")
		} else if luaxmlCount > 0 {
			fmt.Printf("Created MPQ with LuaXML files and copied to client\n")
		}
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║           Build Complete!                ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// Print summary - only show what was actually built
	if dbcCount > 0 || luaxmlCount > 0 {
		fmt.Println("Output locations:")
		if cfg.Server.DBCPath != "" && dbcCount > 0 {
			fmt.Printf("  Server DBCs: %s\n", cfg.Server.DBCPath)
		}
		if dbcCount > 0 {
			fmt.Printf("  Client DBC MPQ: %s/%s\n", cfg.GetClientDataPath(), cfg.Output.DBCMPQ)
		}
		if luaxmlCount > 0 {
			fmt.Printf("  Client LuaXML MPQ: %s/%s\n", cfg.GetClientLocalePath(), cfg.GetMPQName(cfg.Output.LuaXMLMPQ))
		}
	}

	return nil
}

// ModifiedLuaXMLFile represents a modified LuaXML file from a mod
type modifiedLuaXMLFile struct {
	ModName  string // Which mod it's from
	FilePath string // Absolute path to the file
	RelPath  string // Relative path (for MPQ)
}

// findModifiedLuaXMLInMod finds LuaXML files in a mod that differ from source
func findModifiedLuaXMLInMod(cfg *config.Config, mod string) ([]modifiedLuaXMLFile, error) {
	modLuaXML := filepath.Join(cfg.GetModsPath(), mod, "luaxml")
	luaxmlSource := cfg.GetLuaXMLSourcePath()

	// Check if mod has luaxml directory
	if _, err := os.Stat(modLuaXML); os.IsNotExist(err) {
		return nil, nil // No luaxml folder, skip
	}

	var modified []modifiedLuaXMLFile

	err := filepath.Walk(modLuaXML, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// Skip hidden files and common junk files
		name := info.Name()
		if strings.HasPrefix(name, ".") || name == "Thumbs.db" {
			return nil
		}

		// Get relative path
		relPath, _ := filepath.Rel(modLuaXML, path)

		// Compare with source file
		sourcePath := filepath.Join(luaxmlSource, relPath)

		// Check if files are different
		if !filesAreIdentical(path, sourcePath) {
			modified = append(modified, modifiedLuaXMLFile{
				ModName:  mod,
				FilePath: path,
				RelPath:  relPath,
			})
		}

		return nil
	})

	return modified, err
}

// filesAreIdentical checks if two files have identical content
func filesAreIdentical(file1, file2 string) bool {
	data1, err1 := os.ReadFile(file1)
	data2, err2 := os.ReadFile(file2)

	if err1 != nil || err2 != nil {
		return false // If either can't be read, consider different
	}

	if len(data1) != len(data2) {
		return false
	}

	for i := range data1 {
		if data1[i] != data2[i] {
			return false
		}
	}

	return true
}

// ModServerPatch represents a server patch from a mod
type modServerPatch struct {
	ModName   string
	PatchName string
	PatchPath string
}

// applyModServerPatches discovers and applies server patches from mods
func applyModServerPatches(cfg *config.Config, mods []string, force bool) (int, error) {
	tcPath := cfg.TrinityCore.SourcePath
	if tcPath == "" {
		return 0, nil // No TC source configured
	}

	// Load tracker
	tracker, _ := loadServerPatchTracker(cfg.WorkspaceRoot)

	// Discover patches from all mods
	var patches []modServerPatch
	for _, mod := range mods {
		modPatches, err := findModServerPatches(cfg, mod)
		if err != nil {
			return 0, fmt.Errorf("find patches in %s: %w", mod, err)
		}
		patches = append(patches, modPatches...)
	}

	if len(patches) == 0 {
		return 0, nil
	}

	// Apply any patches that haven't been applied yet
	applied := 0
	for _, patch := range patches {
		patchID := fmt.Sprintf("%s/%s", patch.ModName, patch.PatchName)

		// Check if already applied (skip unless force)
		alreadyApplied := false
		for _, p := range tracker.Applied {
			if p.Name == patchID {
				alreadyApplied = true
				break
			}
		}
		if alreadyApplied && !force {
			continue
		}

		// Try to apply the patch
		fmt.Printf("[%s] Applying %s...\n", patch.ModName, patch.PatchName)

		cmd := exec.Command("git", "apply", "--check", patch.PatchPath)
		cmd.Dir = tcPath
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("  Warning: patch may not apply cleanly: %s\n", strings.TrimSpace(string(output)))
			fmt.Printf("  Skipping %s (use 'git apply' manually if needed)\n", patch.PatchName)
			continue
		}

		// Actually apply it
		cmd = exec.Command("git", "apply", patch.PatchPath)
		cmd.Dir = tcPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return applied, fmt.Errorf("apply %s: %s\n%s", patch.PatchName, err, string(output))
		}

		// Track it (update if already exists due to force)
		if !alreadyApplied {
			tracker.Applied = append(tracker.Applied, serverPatchApplied{
				Name:      patchID,
				Version:   "1.0.0",
				AppliedAt: time.Now().Format(time.RFC3339),
				AppliedBy: "thorium build",
			})
		}

		fmt.Printf("  ✓ Applied %s\n", patch.PatchName)
		applied++
	}

	// Save tracker
	if applied > 0 {
		if err := saveServerPatchTracker(cfg.WorkspaceRoot, tracker); err != nil {
			fmt.Printf("Warning: could not save patch tracker: %v\n", err)
		}
	}

	return applied, nil
}

// findModServerPatches finds .patch files in a mod's server-patches folder
func findModServerPatches(cfg *config.Config, mod string) ([]modServerPatch, error) {
	patchDir := filepath.Join(cfg.GetModsPath(), mod, "server-patches")

	if _, err := os.Stat(patchDir); os.IsNotExist(err) {
		return nil, nil
	}

	var patches []modServerPatch

	entries, err := os.ReadDir(patchDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".patch") {
			continue
		}

		patches = append(patches, modServerPatch{
			ModName:   mod,
			PatchName: entry.Name(),
			PatchPath: filepath.Join(patchDir, entry.Name()),
		})
	}

	return patches, nil
}

// Server patch tracking (shares format with patch_server.go)
type serverPatchApplied struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	AppliedAt string `json:"applied_at"`
	AppliedBy string `json:"applied_by"`
}

type serverPatchTracker struct {
	Applied []serverPatchApplied `json:"applied"`
}

func loadServerPatchTracker(workspaceRoot string) (serverPatchTracker, error) {
	trackerPath := filepath.Join(workspaceRoot, "shared", "server_patches_applied.json")
	data, err := os.ReadFile(trackerPath)
	if err != nil {
		return serverPatchTracker{}, err
	}
	var tracker serverPatchTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return serverPatchTracker{}, err
	}
	return tracker, nil
}

func saveServerPatchTracker(workspaceRoot string, tracker serverPatchTracker) error {
	sharedDir := filepath.Join(workspaceRoot, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return err
	}
	trackerPath := filepath.Join(sharedDir, "server_patches_applied.json")
	data, err := json.MarshalIndent(tracker, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(trackerPath, data, 0644)
}

// ============================================================================
// Binary Edits
// ============================================================================

// Clean WoW 3.3.5a (12340) client MD5
const CLEAN_CLIENT_MD5 = "45892bdedd0ad70aed4ccd22d9fb5984"

// BinaryEditFile represents a binary edit JSON file
type binaryEditFile struct {
	Patches []binaryPatch `json:"patches"`
}

type binaryPatch struct {
	Address string   `json:"address"`
	Bytes   []string `json:"bytes"`
}

type binaryEditApplied struct {
	Name      string `json:"name"`
	AppliedAt string `json:"applied_at"`
	AppliedBy string `json:"applied_by"`
}

type binaryEditTracker struct {
	Applied []binaryEditApplied `json:"applied"`
}

func loadBinaryEditTracker(workspaceRoot string) (binaryEditTracker, error) {
	trackerPath := filepath.Join(workspaceRoot, "shared", "binary_edits_applied.json")
	data, err := os.ReadFile(trackerPath)
	if err != nil {
		return binaryEditTracker{}, err
	}
	var tracker binaryEditTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return binaryEditTracker{}, err
	}
	return tracker, nil
}

func saveBinaryEditTracker(workspaceRoot string, tracker binaryEditTracker) error {
	sharedDir := filepath.Join(workspaceRoot, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return err
	}
	trackerPath := filepath.Join(sharedDir, "binary_edits_applied.json")
	data, err := json.MarshalIndent(tracker, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(trackerPath, data, 0644)
}

// applyModBinaryEdits discovers and applies binary edits from mods
func applyModBinaryEdits(cfg *config.Config, mods []string, force bool) (int, error) {
	wowExePath := filepath.Join(cfg.WoTLK.Path, "Wow.exe")
	if _, err := os.Stat(wowExePath); os.IsNotExist(err) {
		return 0, nil // No Wow.exe found
	}

	// Load tracker
	tracker, _ := loadBinaryEditTracker(cfg.WorkspaceRoot)

	// Discover binary edits from all mods
	type modBinaryEdit struct {
		ModName  string
		EditName string
		EditPath string
	}
	var edits []modBinaryEdit

	for _, mod := range mods {
		editDir := filepath.Join(cfg.GetModsPath(), mod, "binary-edits")
		if _, err := os.Stat(editDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(editDir)
		if err != nil {
			return 0, fmt.Errorf("read binary-edits dir for %s: %w", mod, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			edits = append(edits, modBinaryEdit{
				ModName:  mod,
				EditName: entry.Name(),
				EditPath: filepath.Join(editDir, entry.Name()),
			})
		}
	}

	if len(edits) == 0 {
		return 0, nil
	}

	// Read Wow.exe
	wowBin, err := os.ReadFile(wowExePath)
	if err != nil {
		return 0, fmt.Errorf("read Wow.exe: %w", err)
	}

	// Create backup if it doesn't exist
	backupPath := wowExePath + ".clean"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		if err := os.WriteFile(backupPath, wowBin, 0644); err != nil {
			return 0, fmt.Errorf("create backup: %w", err)
		}
		fmt.Printf("  Created backup: %s\n", backupPath)
	}

	// Verify clean backup MD5
	backupBin, err := os.ReadFile(backupPath)
	if err != nil {
		return 0, fmt.Errorf("read backup: %w", err)
	}
	
	backupHash := md5.Sum(backupBin)
	backupMD5 := hex.EncodeToString(backupHash[:])
	
	if backupMD5 == CLEAN_CLIENT_MD5 {
		fmt.Printf("  ✓ Clean WoW 3.3.5a client verified (MD5: %s)\n", backupMD5)
	} else {
		fmt.Printf("  ⚠ Warning: Backup MD5 %s does not match expected clean client MD5 %s\n", backupMD5, CLEAN_CLIENT_MD5)
		fmt.Printf("  Expected: %s\n", CLEAN_CLIENT_MD5)
		fmt.Printf("  Got:      %s\n", backupMD5)
		fmt.Printf("  This may indicate a non-standard WoW 3.3.5a (12340) client.\n")
		fmt.Printf("  Patches may not apply correctly. Proceed with caution.\n")
	}

	// Apply edits
	applied := 0
	modified := false

	for _, edit := range edits {
		editID := fmt.Sprintf("%s/%s", edit.ModName, edit.EditName)

		// Check if already applied (skip unless force)
		alreadyApplied := false
		for _, e := range tracker.Applied {
			if e.Name == editID {
				alreadyApplied = true
				break
			}
		}
		if alreadyApplied && !force {
			continue
		}

		// Parse the edit file
		editData, err := os.ReadFile(edit.EditPath)
		if err != nil {
			return applied, fmt.Errorf("read %s: %w", edit.EditPath, err)
		}

		var editFile binaryEditFile
		if err := json.Unmarshal(editData, &editFile); err != nil {
			return applied, fmt.Errorf("parse %s: %w", edit.EditPath, err)
		}

		fmt.Printf("[%s] Applying %s...\n", edit.ModName, edit.EditName)

		// Apply each patch
		for _, patch := range editFile.Patches {
			// Parse address (hex string like "0x28e19c")
			var address uint32
			_, err := fmt.Sscanf(patch.Address, "0x%x", &address)
			if err != nil {
				_, err = fmt.Sscanf(patch.Address, "%x", &address)
				if err != nil {
					return applied, fmt.Errorf("invalid address %s in %s: %w", patch.Address, edit.EditName, err)
				}
			}

			// Apply bytes
			for i, byteStr := range patch.Bytes {
				var b uint8
				_, err := fmt.Sscanf(byteStr, "0x%x", &b)
				if err != nil {
					_, err = fmt.Sscanf(byteStr, "%x", &b)
					if err != nil {
						return applied, fmt.Errorf("invalid byte %s in %s: %w", byteStr, edit.EditName, err)
					}
				}

				offset := int(address) + i
				if offset >= len(wowBin) {
					return applied, fmt.Errorf("address 0x%x exceeds file size in %s", offset, edit.EditName)
				}
				wowBin[offset] = b
			}
		}

		// Track it
		if !alreadyApplied {
			tracker.Applied = append(tracker.Applied, binaryEditApplied{
				Name:      editID,
				AppliedAt: time.Now().Format(time.RFC3339),
				AppliedBy: "thorium build",
			})
		}

		fmt.Printf("  ✓ Applied %s (%d patches)\n", edit.EditName, len(editFile.Patches))
		applied++
		modified = true
	}

	// Write modified Wow.exe
	if modified {
		if err := os.WriteFile(wowExePath, wowBin, 0755); err != nil {
			return applied, fmt.Errorf("write Wow.exe: %w", err)
		}
	}

	// Save tracker
	if applied > 0 {
		if err := saveBinaryEditTracker(cfg.WorkspaceRoot, tracker); err != nil {
			fmt.Printf("Warning: could not save binary edit tracker: %v\n", err)
		}
	}

	return applied, nil
}

// ============================================================================
// Assets
// ============================================================================

// AssetsConfig represents assets/config.json
type assetsConfig struct {
	Files []assetFile `json:"files"`
}

type assetFile struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// Asset tracking
type assetApplied struct {
	Name      string `json:"name"`
	MD5       string `json:"md5"`
	AppliedAt string `json:"applied_at"`
	AppliedBy string `json:"applied_by"`
}

type assetTracker struct {
	Applied []assetApplied `json:"applied"`
}

func loadAssetTracker(workspaceRoot string) (assetTracker, error) {
	trackerPath := filepath.Join(workspaceRoot, "shared", "assets_applied.json")
	data, err := os.ReadFile(trackerPath)
	if err != nil {
		return assetTracker{}, err
	}
	var tracker assetTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return assetTracker{}, err
	}
	return tracker, nil
}

func saveAssetTracker(workspaceRoot string, tracker assetTracker) error {
	sharedDir := filepath.Join(workspaceRoot, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return err
	}
	trackerPath := filepath.Join(sharedDir, "assets_applied.json")
	data, err := json.MarshalIndent(tracker, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(trackerPath, data, 0644)
}

func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// copyModAssets copies asset files from mods to client directory
func copyModAssets(cfg *config.Config, mods []string, force bool) (int, error) {
	clientPath := cfg.WoTLK.Path
	if clientPath == "" {
		return 0, nil
	}

	// Load tracker
	tracker, _ := loadAssetTracker(cfg.WorkspaceRoot)

	copied := 0
	trackerModified := false

	for _, mod := range mods {
		assetsDir := filepath.Join(cfg.GetModsPath(), mod, "assets")
		configPath := filepath.Join(assetsDir, "config.json")

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue // No assets config
		}

		// Parse config
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return copied, fmt.Errorf("read assets config for %s: %w", mod, err)
		}

		var config assetsConfig
		if err := json.Unmarshal(configData, &config); err != nil {
			return copied, fmt.Errorf("parse assets config for %s: %w", mod, err)
		}

		// Copy each file
		for _, file := range config.Files {
			srcPath := filepath.Join(assetsDir, file.Source)
			destPath := filepath.Join(clientPath, file.Destination, file.Source)

			// If destination is ".", just use the filename
			if file.Destination == "." {
				destPath = filepath.Join(clientPath, file.Source)
			}

			// Read source
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return copied, fmt.Errorf("read asset %s: %w", srcPath, err)
			}

			// Calculate MD5 of source
			srcMD5 := calculateMD5(data)
			assetID := fmt.Sprintf("%s/%s", mod, file.Source)

			// Check if already applied with same hash (skip unless force)
			if !force {
				alreadyApplied := false
				for _, a := range tracker.Applied {
					if a.Name == assetID && a.MD5 == srcMD5 {
						alreadyApplied = true
						break
					}
				}
				if alreadyApplied {
					continue // Same file already copied
				}
			}

			// Ensure destination directory exists
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return copied, fmt.Errorf("create asset dir: %w", err)
			}

			// Write destination
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return copied, fmt.Errorf("write asset %s: %w", destPath, err)
			}

			// Update tracker (replace existing entry or add new)
			found := false
			for i, a := range tracker.Applied {
				if a.Name == assetID {
					tracker.Applied[i].MD5 = srcMD5
					tracker.Applied[i].AppliedAt = time.Now().Format(time.RFC3339)
					tracker.Applied[i].AppliedBy = "thorium build"
					found = true
					break
				}
			}
			if !found {
				tracker.Applied = append(tracker.Applied, assetApplied{
					Name:      assetID,
					MD5:       srcMD5,
					AppliedAt: time.Now().Format(time.RFC3339),
					AppliedBy: "thorium build",
				})
			}
			trackerModified = true

			fmt.Printf("[%s] Copied %s -> %s\n", mod, file.Source, destPath)
			copied++
		}
	}

	// Save tracker
	if trackerModified {
		if err := saveAssetTracker(cfg.WorkspaceRoot, tracker); err != nil {
			fmt.Printf("Warning: could not save asset tracker: %v\n", err)
		}
	}

	return copied, nil
}

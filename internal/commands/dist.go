// Copyright (c) 2025 Thorium

package commands

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	gompq "github.com/suprsokr/go-mpq"

	"thorium-cli/internal/config"
	"thorium-cli/internal/dbc"
	"thorium-cli/internal/mpq"
)

// Dist creates a distributable package for players (client files only)
func Dist(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("dist", flag.ExitOnError)
	modName := fs.String("mod", "", "Package specific mod only")
	outputPath := fs.String("output", "", "Output zip file path (default: mods/<mod>/dist/<timestamp>.zip)")
	noExe := fs.Bool("no-exe", false, "Skip including wow.exe even if binary edits were applied")
	fs.Parse(args)

	fmt.Println("=== Creating Client Distribution Package ===")
	fmt.Println()
	fmt.Println("This creates a player-ready package with MPQs and optional wow.exe.")
	fmt.Println("For mod source distribution, host your mod on GitHub.")
	fmt.Println()

	// Get list of mods
	mods, err := listMods(cfg)
	if err != nil {
		return fmt.Errorf("list mods: %w", err)
	}

	// Infer mod name from current directory if not provided
	if *modName == "" {
		cwd, err := os.Getwd()
		if err == nil {
			// Check if we're in a mod directory (mods/<mod>)
			modsPath := cfg.GetModsPath()
			if strings.HasPrefix(cwd, modsPath) {
				relPath, err := filepath.Rel(modsPath, cwd)
				if err == nil {
					parts := strings.Split(relPath, string(filepath.Separator))
					if len(parts) > 0 && parts[0] != "." && parts[0] != ".." {
						// Check if this mod exists
						for _, m := range mods {
							if m == parts[0] {
								*modName = parts[0]
								fmt.Printf("Inferred mod name from directory: %s\n", *modName)
								break
							}
						}
					}
				}
			}
		}
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
		fmt.Println("No mods found.")
		return nil
	}

	// For dist, we only support single mod
	if len(mods) > 1 {
		return fmt.Errorf("dist command requires --mod flag when multiple mods exist")
	}
	targetMod := mods[0]

	// Create temp directory for building
	tempDir, err := os.MkdirTemp("", "thorium-dist-*")
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Building distribution for mod: %s\n", targetMod)
	fmt.Println()

	// Step 1: Apply mod's DBC migrations to dbc_source (will rollback after export)
	fmt.Println("Applying mod's DBC migrations to dbc_source...")
	if err := applyMigrationsToDatabase(cfg, targetMod, "dbc", &cfg.Databases.DBCSource, false); err != nil {
		return fmt.Errorf("apply dbc migrations to dbc_source: %w", err)
	}

	// Ensure we rollback migrations even if export fails
	defer func() {
		fmt.Println("Rolling back DBC migrations from dbc_source...")
		if err := rollbackMigrationsFromDatabase(cfg, targetMod, "dbc", &cfg.Databases.DBCSource, true); err != nil {
			fmt.Printf("Warning: failed to rollback migrations: %v\n", err)
		}
	}()

	// Step 2: Export DBCs from dbc_source
	fmt.Println("Exporting DBCs from dbc_source...")
	tempDBCSource := filepath.Join(tempDir, "dbc_source")
	tempDBCOut := filepath.Join(tempDir, "dbc_out")
	if err := os.MkdirAll(tempDBCSource, 0755); err != nil {
		return fmt.Errorf("create temp dbc source dir: %w", err)
	}
	if err := os.MkdirAll(tempDBCOut, 0755); err != nil {
		return fmt.Errorf("create temp dbc out dir: %w", err)
	}

	// Copy baseline source DBCs to temp (we need them for comparison)
	// Note: We need the ORIGINAL baseline files, not the current dbc_source state
	dbcSourceFiles := cfg.GetDBCSourcePath()
	if err := copyDirContents(dbcSourceFiles, tempDBCSource); err != nil {
		return fmt.Errorf("copy dbc source files: %w", err)
	}

	// Export DBCs from dbc_source (which now has the mod's migrations applied)
	exporter := dbc.NewExporterWithDB(cfg, cfg.Databases.DBCSource)
	tables, err := exporter.Export()
	if err != nil {
		return fmt.Errorf("export DBCs: %w", err)
	}

	// Copy exported DBCs to temp
	dbcOut := cfg.GetDBCOutPath()
	if err := copyDirContents(dbcOut, tempDBCOut); err != nil {
		return fmt.Errorf("copy dbc out: %w", err)
	}

	if len(tables) > 0 {
		fmt.Printf("  Exported %d DBC file(s)\n", len(tables))
	} else {
		fmt.Println("  No modified DBCs found")
	}

	// Step 3: Collect LuaXML files from mod
	fmt.Println("Collecting LuaXML files...")
	modLuaXMLFiles, err := findModifiedLuaXMLInMod(cfg, targetMod)
	if err != nil {
		return fmt.Errorf("find luaxml files: %w", err)
	}
	if len(modLuaXMLFiles) > 0 {
		fmt.Printf("  Found %d LuaXML file(s)\n", len(modLuaXMLFiles))
	} else {
		fmt.Println("  No LuaXML modifications found")
	}

	// Step 4: Collect SQL files from mod
	fmt.Println("Collecting SQL files...")
	sqlFiles, err := collectModSQLFiles(cfg, targetMod)
	if err != nil {
		return fmt.Errorf("collect sql files: %w", err)
	}
	if len(sqlFiles) > 0 {
		fmt.Printf("  Found %d SQL file(s)\n", len(sqlFiles))
	} else {
		fmt.Println("  No SQL files found")
	}

	// Step 5: Build MPQs in temp directory
	fmt.Println("Building MPQs...")
	tempMPQDir := filepath.Join(tempDir, "mpqs")
	if err := os.MkdirAll(tempMPQDir, 0755); err != nil {
		return fmt.Errorf("create temp mpq dir: %w", err)
	}

	var dbcMPQPath, luaxmlMPQPath string
	filesAdded := 0

	// Build DBC MPQ if we have DBCs
	if len(tables) > 0 {
		dbcMPQPath = filepath.Join(tempMPQDir, cfg.Output.DBCMPQ)
		if err := buildTempDBCMPQ(tempDBCOut, tempDBCSource, dbcMPQPath); err != nil {
			return fmt.Errorf("build dbc mpq: %w", err)
		}
		fmt.Printf("  Created: %s\n", filepath.Base(dbcMPQPath))
		filesAdded++
	}

	// Build LuaXML MPQ if we have LuaXML files
	if len(modLuaXMLFiles) > 0 {
		luaxmlMPQName := cfg.GetMPQName(cfg.Output.LuaXMLMPQ)
		luaxmlMPQPath = filepath.Join(tempMPQDir, luaxmlMPQName)
		
		// Convert to mpq.ModifiedLuaXMLFile format
		var mpqFiles []mpq.ModifiedLuaXMLFile
		for _, f := range modLuaXMLFiles {
			mpqFiles = append(mpqFiles, mpq.ModifiedLuaXMLFile{
				ModName:  f.ModName,
				FilePath: f.FilePath,
				RelPath:  f.RelPath,
			})
		}

		if err := buildTempLuaXMLMPQ(mpqFiles, luaxmlMPQPath); err != nil {
			return fmt.Errorf("build luaxml mpq: %w", err)
		}
		fmt.Printf("  Created: %s\n", filepath.Base(luaxmlMPQPath))
		filesAdded++
	}

	// Step 6: Determine output path
	zipPath := *outputPath
	if zipPath == "" {
		modDistDir := filepath.Join(cfg.GetModsPath(), targetMod, "dist")
		if err := os.MkdirAll(modDistDir, 0755); err != nil {
			return fmt.Errorf("create mod dist directory: %w", err)
		}
		timestamp := time.Now().Format("20060102_150405")
		zipPath = filepath.Join(modDistDir, fmt.Sprintf("%s_%s.zip", targetMod, timestamp))
	}

	// Step 7: Create zip file
	fmt.Println()
	fmt.Println("Creating distribution package...")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add MPQs to zip
	if dbcMPQPath != "" {
		if err := addFileToZip(zipWriter, dbcMPQPath, filepath.Base(dbcMPQPath)); err != nil {
			return fmt.Errorf("add dbc mpq to zip: %w", err)
		}
		fmt.Printf("  Added: %s\n", filepath.Base(dbcMPQPath))
	}
	if luaxmlMPQPath != "" {
		if err := addFileToZip(zipWriter, luaxmlMPQPath, filepath.Base(luaxmlMPQPath)); err != nil {
			return fmt.Errorf("add luaxml mpq to zip: %w", err)
		}
		fmt.Printf("  Added: %s\n", filepath.Base(luaxmlMPQPath))
	}

	// Add SQL files to zip
	for _, sqlFile := range sqlFiles {
		zipPath := filepath.Join("sql", sqlFile.relPath)
		if err := addFileToZip(zipWriter, sqlFile.absPath, zipPath); err != nil {
			return fmt.Errorf("add sql file to zip: %w", err)
		}
		fmt.Printf("  Added: %s\n", zipPath)
		filesAdded++
	}

	// Check if any binary edits have been applied
	hasBinaryEdits, err := checkBinaryEditsApplied(cfg)
	if err != nil {
		return fmt.Errorf("check binary edits: %w", err)
	}

	// Add wow.exe if binary edits were applied (unless --no-exe flag)
	if hasBinaryEdits && !*noExe {
		fmt.Println("Binary edits detected, collecting wow.exe...")
		exeFiles, err := collectWowExe(cfg)
		if err != nil {
			return fmt.Errorf("collect wow.exe: %w", err)
		}
		if len(exeFiles) == 0 {
			fmt.Println("  Warning: wow.exe not found in WoTLK path. Skipping.")
			fmt.Println("  Make sure binary edits have been applied with 'thorium build'")
		}
		for _, ef := range exeFiles {
			if err := addFileToZip(zipWriter, ef.srcPath, ef.zipPath); err != nil {
				return fmt.Errorf("add %s to zip: %w", ef.srcPath, err)
			}
			fmt.Printf("  Added: %s\n", ef.zipPath)
			filesAdded++
		}
	} else if *noExe && hasBinaryEdits {
		fmt.Println("Skipping wow.exe (--no-exe flag used)")
	}

	// Add a README
	var clientFiles []distFile
	if dbcMPQPath != "" {
		clientFiles = append(clientFiles, distFile{zipPath: filepath.Base(dbcMPQPath)})
	}
	if luaxmlMPQPath != "" {
		clientFiles = append(clientFiles, distFile{zipPath: filepath.Base(luaxmlMPQPath)})
	}
	readmeContent := generateDistReadme([]string{targetMod}, clientFiles, hasBinaryEdits)
	readmeWriter, err := zipWriter.Create("README.txt")
	if err != nil {
		return fmt.Errorf("create README in zip: %w", err)
	}
	if _, err := readmeWriter.Write([]byte(readmeContent)); err != nil {
		return fmt.Errorf("write README: %w", err)
	}
	fmt.Printf("  Added: README.txt\n")
	filesAdded++

	if filesAdded == 0 {
		fmt.Println()
		fmt.Println("No files to package.")
		os.Remove(zipPath)
		return nil
	}

	fmt.Println()
	fmt.Printf("Created: %s\n", zipPath)
	fmt.Printf("Total files: %d\n", filesAdded)
	fmt.Println("\n=== Distribution Complete ===")

	return nil
}

// distFile represents a file to add to the distribution
type distFile struct {
	srcPath string
	zipPath string
}

// collectClientFiles finds client MPQ files to distribute
func collectClientFiles(cfg *config.Config) ([]distFile, error) {
	var files []distFile

	// Check for DBC MPQ
	dbcMPQ := cfg.Output.DBCMPQ
	if dbcMPQ != "" {
		// Resolve path
		if !filepath.IsAbs(dbcMPQ) {
			dbcMPQ = filepath.Join(cfg.WorkspaceRoot, dbcMPQ)
		}
		if _, err := os.Stat(dbcMPQ); err == nil {
			files = append(files, distFile{
				srcPath: dbcMPQ,
				zipPath: filepath.Base(dbcMPQ),
			})
		}
	}

	// Check for LuaXML MPQ
	luaxmlMPQ := cfg.Output.LuaXMLMPQ
	if luaxmlMPQ != "" {
		if !filepath.IsAbs(luaxmlMPQ) {
			luaxmlMPQ = filepath.Join(cfg.WorkspaceRoot, luaxmlMPQ)
		}
		if _, err := os.Stat(luaxmlMPQ); err == nil {
			files = append(files, distFile{
				srcPath: luaxmlMPQ,
				zipPath: filepath.Base(luaxmlMPQ),
			})
		}
	}

	return files, nil
}

// binaryEditsTracker represents the binary edits tracker file structure
type binaryEditsTracker struct {
	Applied []struct {
		Name      string `json:"name"`
		AppliedAt string `json:"applied_at"`
		AppliedBy string `json:"applied_by"`
	} `json:"applied"`
}

// checkBinaryEditsApplied checks if any binary edits have been applied
func checkBinaryEditsApplied(cfg *config.Config) (bool, error) {
	trackerPath := filepath.Join(cfg.GetSharedPath(), "binary_edits_applied.json")
	
	// If the file doesn't exist, no edits have been applied
	if _, err := os.Stat(trackerPath); os.IsNotExist(err) {
		return false, nil
	}

	// Read and parse the tracker file
	data, err := os.ReadFile(trackerPath)
	if err != nil {
		return false, err
	}

	var tracker binaryEditsTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return false, err
	}

	// Check if any edits have been applied
	return len(tracker.Applied) > 0, nil
}

// collectWowExe finds the wow.exe to distribute
func collectWowExe(cfg *config.Config) ([]distFile, error) {
	var files []distFile

	// Look for wow.exe in the WoTLK path
	wowPath := cfg.WoTLK.Path
	if wowPath == "" {
		return files, nil
	}

	// Expand environment variables in the path
	wowPath = os.ExpandEnv(wowPath)

	// Check for wow.exe (we look for wow.exe, Wow.exe, or WoW.exe)
	possibleNames := []string{"wow.exe", "Wow.exe", "WoW.exe"}
	for _, name := range possibleNames {
		exePath := filepath.Join(wowPath, name)
		if _, err := os.Stat(exePath); err == nil {
			files = append(files, distFile{
				srcPath: exePath,
				zipPath: name,
			})
			break
		}
	}

	return files, nil
}


// addFileToZip adds a file to the zip archive
func addFileToZip(zw *zip.Writer, srcPath, zipPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipPath
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// generateDistReadme creates a README for the distribution
func generateDistReadme(mods []string, clientFiles []distFile, includeExe bool) string {
	var sb strings.Builder

	sb.WriteString("Thorium Client Distribution Package\n")
	sb.WriteString("====================================\n\n")
	sb.WriteString("This package contains client files for players.\n")
	sb.WriteString("Connect to a server that has the server-side modifications installed.\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Mods included: %s\n\n", strings.Join(mods, ", ")))

	sb.WriteString("Contents\n")
	sb.WriteString("--------\n\n")

	if includeExe {
		sb.WriteString("- wow.exe: Patched game executable\n")
	}
	if len(clientFiles) > 0 {
		sb.WriteString("- MPQ files: Game data patches\n")
		for _, cf := range clientFiles {
			sb.WriteString(fmt.Sprintf("  - %s\n", cf.zipPath))
		}
	}

	sb.WriteString("\nInstallation\n")
	sb.WriteString("------------\n\n")

	sb.WriteString("IMPORTANT: Backup your existing WoW installation before proceeding!\n\n")

	if includeExe {
		sb.WriteString("1. Copy wow.exe to your WoW 3.3.5a folder (replace existing file)\n")
	}

	if len(clientFiles) > 0 {
		if includeExe {
			sb.WriteString("2. Copy MPQ files to your WoW Data/ folder:\n")
		} else {
			sb.WriteString("1. Copy MPQ files to your WoW Data/ folder:\n")
		}
		sb.WriteString("   - patch-T.MPQ goes in Data/\n")
		sb.WriteString("   - patch-enUS-T.MPQ (or your locale) goes in Data/enUS/ (or your locale)\n\n")
		sb.WriteString("Example structure:\n")
		sb.WriteString("  WoW 3.3.5a/\n")
		if includeExe {
			sb.WriteString("  ├── wow.exe\n")
		}
		sb.WriteString("  └── Data/\n")
		sb.WriteString("      ├── patch-T.MPQ\n")
		sb.WriteString("      └── enUS/\n")
		sb.WriteString("          └── patch-enUS-T.MPQ\n")
	}

	sb.WriteString("\nNotes\n")
	sb.WriteString("-----\n\n")
	sb.WriteString("- These modifications only work with WoW 3.3.5a (12340)\n")
	sb.WriteString("- You must connect to a server with matching server-side mods\n")
	sb.WriteString("- If addons don't load, check that you're using the correct locale folder\n")

	return sb.String()
}

// sqlFile represents a SQL file to include in distribution
type sqlFile struct {
	absPath string
	relPath string
}

// collectModSQLFiles collects SQL files from a mod's dbc_sql and world_sql directories
func collectModSQLFiles(cfg *config.Config, mod string) ([]sqlFile, error) {
	var files []sqlFile

	// Collect DBC SQL files
	dbcSQLDir := filepath.Join(cfg.GetModsPath(), mod, "dbc_sql")
	if entries, err := os.ReadDir(dbcSQLDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".sql") && !strings.HasSuffix(name, ".rollback.sql") {
				absPath := filepath.Join(dbcSQLDir, name)
				relPath := filepath.Join("dbc_sql", name)
				files = append(files, sqlFile{
					absPath: absPath,
					relPath: relPath,
				})
			}
		}
	}

	// Collect World SQL files
	worldSQLDir := filepath.Join(cfg.GetModsPath(), mod, "world_sql")
	if entries, err := os.ReadDir(worldSQLDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".sql") && !strings.HasSuffix(name, ".rollback.sql") {
				absPath := filepath.Join(worldSQLDir, name)
				relPath := filepath.Join("world_sql", name)
				files = append(files, sqlFile{
					absPath: absPath,
					relPath: relPath,
				})
			}
		}
	}

	return files, nil
}

// copyDirContents copies all files from srcDir to dstDir
func copyDirContents(srcDir, dstDir string) error {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil // Source doesn't exist, nothing to copy
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDirContents(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// buildTempDBCMPQ builds a DBC MPQ from temp directories
func buildTempDBCMPQ(dbcOutDir, dbcSourceDir, outputPath string) error {
	// Find modified DBCs by comparing outDir with sourceDir
	var modified []string

	entries, err := os.ReadDir(dbcOutDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".dbc" {
			continue
		}

		outFile := filepath.Join(dbcOutDir, e.Name())
		srcFile := filepath.Join(dbcSourceDir, e.Name())

		if !filesAreIdentical(outFile, srcFile) {
			modified = append(modified, e.Name())
		}
	}

	if len(modified) == 0 {
		return nil // No modified DBCs
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	os.Remove(outputPath)

	// Build MPQ using go-mpq
	archive, err := gompq.CreateV2(outputPath, len(modified)+10)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range modified {
		srcPath := filepath.Join(dbcOutDir, file)
		mpqPath := "DBFilesClient\\" + file
		if err := archive.AddFile(srcPath, mpqPath); err != nil {
			return fmt.Errorf("add %s: %w", file, err)
		}
	}

	return nil
}

// buildTempLuaXMLMPQ builds a LuaXML MPQ from mod files
func buildTempLuaXMLMPQ(files []mpq.ModifiedLuaXMLFile, outputPath string) error {
	if len(files) == 0 {
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	os.Remove(outputPath)

	// Build MPQ using go-mpq
	archive, err := gompq.CreateV2(outputPath, len(files)+10)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range files {
		mpqPath := strings.ReplaceAll(file.RelPath, "/", "\\")
		if err := archive.AddFile(file.FilePath, mpqPath); err != nil {
			return fmt.Errorf("add %s: %w", file.RelPath, err)
		}
	}

	return nil
}

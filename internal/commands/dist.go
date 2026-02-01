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

	"thorium-cli/internal/config"
)

// Dist creates a distributable package for players (client files only)
func Dist(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("dist", flag.ExitOnError)
	modName := fs.String("mod", "", "Package specific mod only")
	outputPath := fs.String("output", "", "Output zip file path (default: dist/<timestamp>.zip)")
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

	// Determine output path
	distDir := filepath.Join(cfg.WorkspaceRoot, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("create dist directory: %w", err)
	}

	zipPath := *outputPath
	if zipPath == "" {
		timestamp := time.Now().Format("20060102_150405")
		if *modName != "" {
			zipPath = filepath.Join(distDir, fmt.Sprintf("%s_%s.zip", *modName, timestamp))
		} else {
			zipPath = filepath.Join(distDir, fmt.Sprintf("thorium_dist_%s.zip", timestamp))
		}
	}

	// Create zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	filesAdded := 0

	// Add client MPQs
	fmt.Println("Collecting client files...")
	clientFiles, err := collectClientFiles(cfg)
	if err != nil {
		return fmt.Errorf("collect client files: %w", err)
	}
	for _, cf := range clientFiles {
		if err := addFileToZip(zipWriter, cf.srcPath, cf.zipPath); err != nil {
			return fmt.Errorf("add %s to zip: %w", cf.srcPath, err)
		}
		fmt.Printf("  Added: %s\n", cf.zipPath)
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
	readmeContent := generateDistReadme(mods, clientFiles, hasBinaryEdits)
	readmeWriter, err := zipWriter.Create("README.txt")
	if err != nil {
		return fmt.Errorf("create README in zip: %w", err)
	}
	if _, err := readmeWriter.Write([]byte(readmeContent)); err != nil {
		return fmt.Errorf("write README: %w", err)
	}
	fmt.Println("  Added: README.txt")
	filesAdded++

	if filesAdded <= 1 { // Only README
		fmt.Println()
		fmt.Println("No client files to package.")
		fmt.Println("Run 'thorium build' first to create MPQ files.")
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

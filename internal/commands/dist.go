// Copyright (c) 2025 Thorium

package commands

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"thorium-cli/internal/config"
)

// Dist creates a distributable package containing client MPQs and server SQL
func Dist(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("dist", flag.ExitOnError)
	modName := fs.String("mod", "", "Package specific mod only")
	outputPath := fs.String("output", "", "Output zip file path (default: dist/<timestamp>.zip)")
	fs.Parse(args)

	fmt.Println("=== Creating Distribution Package ===")
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
		if err := addFileToZip(zipWriter, cf.srcPath, filepath.Join("client", cf.zipPath)); err != nil {
			return fmt.Errorf("add %s to zip: %w", cf.srcPath, err)
		}
		fmt.Printf("  Added: client/%s\n", cf.zipPath)
		filesAdded++
	}

	// Add world SQL migrations (apply and rollback)
	fmt.Println("Collecting server SQL...")
	for _, mod := range mods {
		sqlFiles, err := collectWorldSQL(cfg, mod)
		if err != nil {
			return fmt.Errorf("collect SQL for %s: %w", mod, err)
		}
		for _, sf := range sqlFiles {
			zipDest := filepath.Join("server", "sql", mod, sf.zipPath)
			if err := addFileToZip(zipWriter, sf.srcPath, zipDest); err != nil {
				return fmt.Errorf("add %s to zip: %w", sf.srcPath, err)
			}
			fmt.Printf("  Added: server/sql/%s/%s\n", mod, sf.zipPath)
			filesAdded++
		}
	}

	// Add a README
	readmeContent := generateDistReadme(mods, clientFiles)
	readmeWriter, err := zipWriter.Create("README.txt")
	if err != nil {
		return fmt.Errorf("create README in zip: %w", err)
	}
	if _, err := readmeWriter.Write([]byte(readmeContent)); err != nil {
		return fmt.Errorf("write README: %w", err)
	}
	filesAdded++

	if filesAdded == 0 {
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

// collectWorldSQL finds world SQL migrations to distribute
func collectWorldSQL(cfg *config.Config, mod string) ([]distFile, error) {
	var files []distFile

	worldSQLDir := filepath.Join(cfg.GetModsPath(), mod, "world_sql")
	if _, err := os.Stat(worldSQLDir); os.IsNotExist(err) {
		return files, nil
	}

	entries, err := os.ReadDir(worldSQLDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		files = append(files, distFile{
			srcPath: filepath.Join(worldSQLDir, entry.Name()),
			zipPath: entry.Name(),
		})
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
func generateDistReadme(mods []string, clientFiles []distFile) string {
	var sb strings.Builder

	sb.WriteString("Thorium Distribution Package\n")
	sb.WriteString("============================\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Mods included: %s\n\n", strings.Join(mods, ", ")))

	sb.WriteString("Contents\n")
	sb.WriteString("--------\n\n")

	sb.WriteString("client/\n")
	sb.WriteString("  MPQ files to copy to your WoW Data/ folder.\n")
	for _, cf := range clientFiles {
		sb.WriteString(fmt.Sprintf("  - %s\n", cf.zipPath))
	}

	sb.WriteString("\nserver/sql/\n")
	sb.WriteString("  SQL migrations organized by mod.\n")
	sb.WriteString("  Apply: Run *.sql files (not *.rollback.sql) against your world database.\n")
	sb.WriteString("  Rollback: Run *.rollback.sql files to undo changes.\n")

	sb.WriteString("\nInstallation\n")
	sb.WriteString("------------\n\n")
	sb.WriteString("1. Client: Copy all files from client/ to your WoW Data/ folder.\n")
	sb.WriteString("2. Server: Run the SQL files in server/sql/ against your world database.\n")
	sb.WriteString("   Order matters - apply in alphabetical/timestamp order.\n")

	return sb.String()
}

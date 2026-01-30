// Copyright (c) 2025 Thorium

package mpq

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"thorium-cli/internal/config"
)

// Builder builds MPQ archives
type Builder struct {
	cfg *config.Config
}

// NewBuilder creates a new MPQ builder
func NewBuilder(cfg *config.Config) *Builder {
	return &Builder{cfg: cfg}
}

// PackageDBCs packages DBC files into an MPQ
func (b *Builder) PackageDBCs() (int, error) {
	dbcSource := b.cfg.GetDBCSourcePath()
	dbcOut := b.cfg.GetDBCOutPath()

	// Find modified DBCs
	modified, err := findModifiedFiles(dbcSource, dbcOut, ".dbc")
	if err != nil {
		return 0, err
	}

	if len(modified) == 0 {
		return 0, nil
	}

	// Build MPQ
	outputPath := filepath.Join(b.cfg.GetClientDataPath(), b.cfg.Output.DBCMPQ)

	if err := b.buildMPQ(dbcOut, modified, "DBFilesClient", outputPath); err != nil {
		return 0, err
	}

	return len(modified), nil
}

// ModifiedLuaXMLFile represents a modified LuaXML file from a mod
type ModifiedLuaXMLFile struct {
	ModName  string // Which mod it's from
	FilePath string // Absolute path to the file
	RelPath  string // Relative path (for MPQ)
}

// PackageLuaXMLFromMods packages modified LuaXML files from mods into an MPQ
func (b *Builder) PackageLuaXMLFromMods(files []ModifiedLuaXMLFile) (int, error) {
	if len(files) == 0 {
		return 0, nil
	}

	// Build MPQ
	mpqName := b.cfg.GetMPQName(b.cfg.Output.LuaXMLMPQ)
	outputPath := filepath.Join(b.cfg.GetClientLocalePath(), mpqName)

	// Ensure output directory exists
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	// Try StormLib first (built-in)
	if stormLibAvailable {
		return b.packageLuaXMLFromModsStorm(files, outputPath)
	}

	// Fall back to external tool
	return b.packageLuaXMLFromModsExternal(files, outputPath)
}

// packageLuaXMLFromModsStorm packages using StormLib
func (b *Builder) packageLuaXMLFromModsStorm(files []ModifiedLuaXMLFile, outputPath string) (int, error) {
	os.Remove(outputPath)

	archive, err := CreateWithStorm(outputPath, len(files)+10)
	if err != nil {
		return 0, err
	}
	defer archive.Close()

	for _, file := range files {
		mpqPath := strings.ReplaceAll(file.RelPath, "/", "\\")
		if err := archive.AddFile(file.FilePath, mpqPath); err != nil {
			return 0, fmt.Errorf("add %s: %w", file.RelPath, err)
		}
	}

	return len(files), nil
}

// packageLuaXMLFromModsExternal packages using external tool
func (b *Builder) packageLuaXMLFromModsExternal(files []ModifiedLuaXMLFile, outputPath string) (int, error) {
	mpqbuilder, err := b.findMPQBuilder()
	if err != nil {
		return 0, err
	}

	// Create listfile
	listfile, err := os.CreateTemp("", "thorium_luaxml_*.txt")
	if err != nil {
		return 0, fmt.Errorf("create listfile: %w", err)
	}
	defer os.Remove(listfile.Name())

	for _, file := range files {
		mpqPath := strings.ReplaceAll(file.RelPath, "/", "\\")
		fmt.Fprintf(listfile, "%s\t%s\n", file.FilePath, mpqPath)
	}
	listfile.Close()

	// Run mpqbuilder
	cmd := exec.Command(mpqbuilder, listfile.Name(), outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("mpqbuilder: %w\n%s", err, stderr.String())
	}

	return len(files), nil
}

// PackageLuaXML packages LuaXML files into an MPQ (legacy - uses luaxml_out vs luaxml_source)
func (b *Builder) PackageLuaXML() (int, error) {
	luaxmlSource := b.cfg.GetLuaXMLSourcePath()
	luaxmlOut := b.cfg.GetLuaXMLOutPath()

	// Find modified files
	modified, err := findModifiedFilesRecursive(luaxmlSource, luaxmlOut)
	if err != nil {
		return 0, err
	}

	if len(modified) == 0 {
		return 0, nil
	}

	// Build MPQ
	mpqName := b.cfg.GetMPQName(b.cfg.Output.LuaXMLMPQ)
	outputPath := filepath.Join(b.cfg.GetClientLocalePath(), mpqName)

	if err := b.buildMPQWithPaths(luaxmlOut, modified, outputPath); err != nil {
		return 0, err
	}

	return len(modified), nil
}

// CopyToServer copies modified DBCs to the server
func (b *Builder) CopyToServer() (int, error) {
	if b.cfg.Server.DBCPath == "" {
		return 0, nil
	}

	dbcSource := b.cfg.GetDBCSourcePath()
	dbcOut := b.cfg.GetDBCOutPath()
	serverPath := b.cfg.Server.DBCPath

	// Find modified DBCs
	modified, err := findModifiedFiles(dbcSource, dbcOut, ".dbc")
	if err != nil {
		return 0, err
	}

	if len(modified) == 0 {
		return 0, nil
	}

	// Create server directory
	if err := os.MkdirAll(serverPath, 0755); err != nil {
		return 0, fmt.Errorf("create server dir: %w", err)
	}

	// Copy files
	for _, file := range modified {
		src := filepath.Join(dbcOut, file)
		dst := filepath.Join(serverPath, file)

		if err := copyFile(src, dst); err != nil {
			return 0, fmt.Errorf("copy %s: %w", file, err)
		}
	}

	return len(modified), nil
}

// buildMPQ builds an MPQ with files in a single directory
func (b *Builder) buildMPQ(sourceDir string, files []string, mpqPrefix, outputPath string) error {
	// Ensure output directory exists
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	// Try StormLib first (built-in)
	if stormLibAvailable {
		return b.buildMPQWithStorm(sourceDir, files, mpqPrefix, outputPath)
	}

	// Fall back to external mpqbuilder
	return b.buildMPQExternal(sourceDir, files, mpqPrefix, outputPath)
}

// buildMPQWithStorm builds an MPQ using the built-in StormLib
func (b *Builder) buildMPQWithStorm(sourceDir string, files []string, mpqPrefix, outputPath string) error {
	// Remove existing file
	os.Remove(outputPath)

	archive, err := CreateWithStorm(outputPath, len(files)+10)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range files {
		srcPath := filepath.Join(sourceDir, file)
		mpqPath := mpqPrefix + "\\" + file

		if err := archive.AddFile(srcPath, mpqPath); err != nil {
			return fmt.Errorf("add %s: %w", file, err)
		}
	}

	return nil
}

// buildMPQExternal builds an MPQ using external mpqbuilder tool
func (b *Builder) buildMPQExternal(sourceDir string, files []string, mpqPrefix, outputPath string) error {
	mpqbuilder, err := b.findMPQBuilder()
	if err != nil {
		return err
	}

	// Create listfile
	listfile, err := os.CreateTemp("", "thorium_mpq_*.txt")
	if err != nil {
		return fmt.Errorf("create listfile: %w", err)
	}
	defer os.Remove(listfile.Name())

	for _, file := range files {
		srcPath := filepath.Join(sourceDir, file)
		mpqPath := mpqPrefix + "\\" + file
		fmt.Fprintf(listfile, "%s\t%s\n", srcPath, mpqPath)
	}
	listfile.Close()

	// Run mpqbuilder
	cmd := exec.Command(mpqbuilder, listfile.Name(), outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mpqbuilder: %w\n%s", err, stderr.String())
	}

	return nil
}

// buildMPQWithPaths builds an MPQ preserving directory structure
func (b *Builder) buildMPQWithPaths(sourceDir string, files []string, outputPath string) error {
	// Ensure output directory exists
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	// Try StormLib first (built-in)
	if stormLibAvailable {
		return b.buildMPQWithPathsStorm(sourceDir, files, outputPath)
	}

	// Fall back to external mpqbuilder
	return b.buildMPQWithPathsExternal(sourceDir, files, outputPath)
}

// buildMPQWithPathsStorm builds an MPQ using StormLib
func (b *Builder) buildMPQWithPathsStorm(sourceDir string, files []string, outputPath string) error {
	os.Remove(outputPath)

	archive, err := CreateWithStorm(outputPath, len(files)+10)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range files {
		srcPath := filepath.Join(sourceDir, file)
		mpqPath := strings.ReplaceAll(file, "/", "\\")

		if err := archive.AddFile(srcPath, mpqPath); err != nil {
			return fmt.Errorf("add %s: %w", file, err)
		}
	}

	return nil
}

// buildMPQWithPathsExternal builds an MPQ using external tool
func (b *Builder) buildMPQWithPathsExternal(sourceDir string, files []string, outputPath string) error {
	mpqbuilder, err := b.findMPQBuilder()
	if err != nil {
		return err
	}

	// Create listfile
	listfile, err := os.CreateTemp("", "thorium_mpq_*.txt")
	if err != nil {
		return fmt.Errorf("create listfile: %w", err)
	}
	defer os.Remove(listfile.Name())

	for _, file := range files {
		srcPath := filepath.Join(sourceDir, file)
		mpqPath := strings.ReplaceAll(file, "/", "\\")
		fmt.Fprintf(listfile, "%s\t%s\n", srcPath, mpqPath)
	}
	listfile.Close()

	// Run mpqbuilder
	cmd := exec.Command(mpqbuilder, listfile.Name(), outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mpqbuilder: %w\n%s", err, stderr.String())
	}

	return nil
}

// findMPQBuilder locates the mpqbuilder binary
func (b *Builder) findMPQBuilder() (string, error) {
	// Check PATH first
	if path, err := exec.LookPath("mpqbuilder"); err == nil {
		return path, nil
	}

	// Check tools directory
	execPath, _ := os.Executable()
	toolsPath := filepath.Join(filepath.Dir(execPath), "tools")

	candidates := []string{
		filepath.Join(toolsPath, "mpqbuilder", "mpqbuilder"),
		filepath.Join(toolsPath, "mpqbuilder"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("mpqbuilder not found. Install it or add to PATH")
}

// findModifiedFiles finds files that differ between source and output directories
func findModifiedFiles(sourceDir, outDir, ext string) ([]string, error) {
	var modified []string

	entries, err := os.ReadDir(outDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if ext != "" && filepath.Ext(e.Name()) != ext {
			continue
		}

		outFile := filepath.Join(outDir, e.Name())
		srcFile := filepath.Join(sourceDir, e.Name())

		if !filesEqual(outFile, srcFile) {
			modified = append(modified, e.Name())
		}
	}

	return modified, nil
}

// findModifiedFilesRecursive finds files that differ, preserving paths
func findModifiedFilesRecursive(sourceDir, outDir string) ([]string, error) {
	var modified []string

	err := filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		relPath, _ := filepath.Rel(outDir, path)
		srcPath := filepath.Join(sourceDir, relPath)

		if !filesEqual(path, srcPath) {
			modified = append(modified, relPath)
		}

		return nil
	})

	return modified, err
}

// filesEqual checks if two files have identical content
func filesEqual(file1, file2 string) bool {
	f1, err := os.Open(file1)
	if err != nil {
		return false
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false
	}
	defer f2.Close()

	s1, _ := f1.Stat()
	s2, _ := f2.Stat()
	if s1.Size() != s2.Size() {
		return false
	}

	buf1 := make([]byte, 4096)
	buf2 := make([]byte, 4096)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true
		}
		if err1 != nil || err2 != nil {
			return false
		}
	}
}

// copyFile copies a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

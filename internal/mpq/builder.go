// Copyright (c) 2025 Thorium

package mpq

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"thorium-cli/internal/config"

	gompq "github.com/suprsokr/go-mpq"
)

// Archive wraps go-mpq Archive for reading MPQ files
type Archive struct {
	*gompq.Archive
}

// Open opens an MPQ archive for reading
func Open(path string) (*Archive, error) {
	a, err := gompq.Open(path)
	if err != nil {
		return nil, err
	}
	return &Archive{a}, nil
}

// Extract extracts files matching a pattern from the archive
func (a *Archive) Extract(pattern, outputDir string) ([]string, error) {
	files, err := a.ListFiles()
	if err != nil {
		return nil, err
	}

	var extracted []string
	for _, file := range files {
		if matchPattern(file, pattern) {
			destPath := filepath.Join(outputDir, strings.ReplaceAll(file, "\\", string(os.PathSeparator)))
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return extracted, err
			}
			if err := a.ExtractFile(file, destPath); err != nil {
				return extracted, fmt.Errorf("extract %s: %w", file, err)
			}
			extracted = append(extracted, file)
		}
	}
	return extracted, nil
}

// matchPattern performs simple wildcard matching
func matchPattern(name, pattern string) bool {
	name = strings.ToLower(strings.ReplaceAll(name, "\\", "/"))
	pattern = strings.ToLower(strings.ReplaceAll(pattern, "\\", "/"))

	if pattern == "*" || pattern == "" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(name, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(name, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	}
	return name == pattern
}

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
	os.Remove(outputPath)

	// Use V2 format for WotLK (3.x) and later compatibility
	archive, err := gompq.CreateV2(outputPath, len(files)+10)
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
	os.Remove(outputPath)

	// Use V2 format for WotLK (3.x) and later compatibility
	archive, err := gompq.CreateV2(outputPath, len(files)+10)
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

// buildMPQWithPaths builds an MPQ preserving directory structure
func (b *Builder) buildMPQWithPaths(sourceDir string, files []string, outputPath string) error {
	// Ensure output directory exists
	os.MkdirAll(filepath.Dir(outputPath), 0755)
	os.Remove(outputPath)

	// Use V2 format for WotLK (3.x) and later compatibility
	archive, err := gompq.CreateV2(outputPath, len(files)+10)
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

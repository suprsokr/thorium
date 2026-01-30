// Copyright (c) 2025 Thorium
// MPQ archive reading and writing.
// Uses StormLib via cgo or external tools as fallback.

package mpq

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Archive represents an MPQ archive
type Archive struct {
	path   string
	mode   string // "r" for read, "w" for write
	files  map[string]string // mpqPath -> localPath for writing
}

// Open opens an MPQ archive for reading
func Open(path string) (*Archive, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("archive not found: %s", path)
	}
	
	return &Archive{
		path: path,
		mode: "r",
	}, nil
}

// Create creates a new MPQ archive for writing
func Create(path string) (*Archive, error) {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	
	return &Archive{
		path:  path,
		mode:  "w",
		files: make(map[string]string),
	}, nil
}

// AddFile adds a file to be written to the archive
func (a *Archive) AddFile(localPath, mpqPath string) error {
	if a.mode != "w" {
		return fmt.Errorf("archive not opened for writing")
	}
	
	// Normalize MPQ path (use backslashes)
	mpqPath = strings.ReplaceAll(mpqPath, "/", "\\")
	a.files[mpqPath] = localPath
	return nil
}

// Close closes the archive, writing it if in write mode
func (a *Archive) Close() error {
	if a.mode == "w" && len(a.files) > 0 {
		return a.write()
	}
	return nil
}

// write writes the archive using available tools
func (a *Archive) write() error {
	// Try to find mpqbuilder or other MPQ tools
	tool, err := findMPQTool()
	if err != nil {
		return err
	}
	
	// Create listfile
	listfile, err := os.CreateTemp("", "thorium_mpq_*.txt")
	if err != nil {
		return fmt.Errorf("create listfile: %w", err)
	}
	defer os.Remove(listfile.Name())
	
	for mpqPath, localPath := range a.files {
		fmt.Fprintf(listfile, "%s\t%s\n", localPath, mpqPath)
	}
	listfile.Close()
	
	// Run mpqbuilder
	cmd := exec.Command(tool, listfile.Name(), a.path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mpqbuilder failed: %w\n%s", err, string(output))
	}
	
	return nil
}

// Extract extracts files from the archive matching a pattern
func (a *Archive) Extract(pattern, outputDir string) ([]string, error) {
	if a.mode != "r" {
		return nil, fmt.Errorf("archive not opened for reading")
	}
	
	tool, err := findMPQExtractor()
	if err != nil {
		return nil, err
	}
	
	// Different tools have different syntax
	var cmd *exec.Cmd
	switch filepath.Base(tool) {
	case "mpqextract", "MPQExtractor":
		cmd = exec.Command(tool, "-e", a.path, "-o", outputDir, "-p", pattern)
	default:
		cmd = exec.Command(tool, "extract", a.path, pattern, outputDir)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w\n%s", err, string(output))
	}
	
	// List extracted files
	var files []string
	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(outputDir, path)
			files = append(files, rel)
		}
		return nil
	})
	
	return files, nil
}

// ListFiles lists files in the archive
func (a *Archive) ListFiles() ([]string, error) {
	if a.mode != "r" {
		return nil, fmt.Errorf("archive not opened for reading")
	}
	
	tool, err := findMPQExtractor()
	if err != nil {
		return nil, err
	}
	
	cmd := exec.Command(tool, "list", a.path)
	output, err := cmd.Output()
	if err != nil {
		// Try alternative syntax
		cmd = exec.Command(tool, "-l", a.path)
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("list failed: %w", err)
		}
	}
	
	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			files = append(files, line)
		}
	}
	
	return files, nil
}

// ExtractFile extracts a single file from the archive
func (a *Archive) ExtractFile(mpqPath string, w io.Writer) error {
	// Create temp dir, extract, read, cleanup
	tmpDir, err := os.MkdirTemp("", "thorium_extract_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	
	_, err = a.Extract(mpqPath, tmpDir)
	if err != nil {
		return err
	}
	
	// Find the extracted file
	var foundPath string
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			foundPath = path
			return filepath.SkipAll
		}
		return nil
	})
	
	if foundPath == "" {
		return fmt.Errorf("file not found in archive: %s", mpqPath)
	}
	
	f, err := os.Open(foundPath)
	if err != nil {
		return err
	}
	defer f.Close()
	
	_, err = io.Copy(w, f)
	return err
}

// findMPQTool finds an MPQ builder tool
func findMPQTool() (string, error) {
	tools := []string{"mpqbuilder", "StormLib"}
	
	for _, tool := range tools {
		if path, err := exec.LookPath(tool); err == nil {
			return path, nil
		}
	}
	
	// Check common locations
	home, _ := os.UserHomeDir()
	locations := []string{
		"/usr/local/bin/mpqbuilder",
		"/usr/bin/mpqbuilder",
		filepath.Join(home, ".local/bin/mpqbuilder"),
	}
	
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}
	
	return "", fmt.Errorf("mpqbuilder not found. Install it or add to PATH")
}

// findMPQExtractor finds an MPQ extraction tool
func findMPQExtractor() (string, error) {
	tools := []string{"mpqextract", "MPQExtractor", "mpq"}
	
	for _, tool := range tools {
		if path, err := exec.LookPath(tool); err == nil {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("MPQ extractor not found. Install mpqextract or MPQExtractor")
}

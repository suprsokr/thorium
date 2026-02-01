// Copyright (c) 2025 Thorium

package scripts

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"thorium-cli/internal/config"
)

// scriptDeployed represents a single deployed script in the tracker
type scriptDeployed struct {
	Name       string `json:"name"`        // mod/filename.cpp
	MD5        string `json:"md5"`         // MD5 hash of source file
	AppliedAt  string `json:"applied_at"`  // Timestamp
	AppliedBy  string `json:"applied_by"`  // "thorium build"
}

// scriptTracker tracks which scripts have been deployed
type scriptTracker struct {
	Applied []scriptDeployed `json:"applied"`
}

// loadScriptTracker loads the script deployment tracker
func loadScriptTracker(workspaceRoot string) (*scriptTracker, error) {
	trackerPath := filepath.Join(workspaceRoot, "shared", "scripts_deployed.json")
	
	data, err := os.ReadFile(trackerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &scriptTracker{Applied: []scriptDeployed{}}, nil
		}
		return nil, err
	}
	
	var tracker scriptTracker
	if err := json.Unmarshal(data, &tracker); err != nil {
		return nil, err
	}
	return &tracker, nil
}

// saveScriptTracker saves the script deployment tracker
func saveScriptTracker(workspaceRoot string, tracker *scriptTracker) error {
	trackerPath := filepath.Join(workspaceRoot, "shared", "scripts_deployed.json")
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(trackerPath), 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(tracker, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(trackerPath, data, 0644)
}

// calculateFileMD5 calculates MD5 hash of a file
func calculateFileMD5(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:]), nil
}

// DeployScripts copies all script files from mods to TrinityCore source and generates CMakeLists.txt
// Returns the number of scripts deployed (new or changed)
func DeployScripts(cfg *config.Config, mods []string, force bool) (int, error) {
	// Check if TrinityCore scripts path is configured
	if cfg.TrinityCore.ScriptsPath == "" {
		fmt.Println("  TrinityCore scripts path not configured, skipping script deployment")
		return 0, nil
	}

	// Ensure Custom scripts directory exists
	if err := os.MkdirAll(cfg.TrinityCore.ScriptsPath, 0755); err != nil {
		return 0, fmt.Errorf("create scripts directory: %w", err)
	}

	// Get workspace root from mods path
	workspaceRoot := filepath.Dir(cfg.GetModsPath())

	// Load tracker
	tracker, err := loadScriptTracker(workspaceRoot)
	if err != nil {
		return 0, fmt.Errorf("load script tracker: %w", err)
	}

	// Collect all scripts from mods
	var scriptFiles []ScriptFile
	for _, mod := range mods {
		modScripts := filepath.Join(cfg.GetModsPath(), mod, "scripts")
		if _, err := os.Stat(modScripts); os.IsNotExist(err) {
			continue // Skip mods without scripts
		}

		files, err := collectScriptFiles(modScripts, mod)
		if err != nil {
			return 0, fmt.Errorf("collect scripts from %s: %w", mod, err)
		}
		scriptFiles = append(scriptFiles, files...)
	}

	if len(scriptFiles) == 0 {
		fmt.Println("  No scripts found in mods")
		return 0, nil
	}

	// Copy scripts to TrinityCore (only new or changed)
	deployed := 0
	for _, script := range scriptFiles {
		scriptID := fmt.Sprintf("%s/%s", script.ModName, script.FileName)
		
		// Calculate MD5 of source
		srcMD5, err := calculateFileMD5(script.SourcePath)
		if err != nil {
			return deployed, fmt.Errorf("calculate MD5 for %s: %w", script.FileName, err)
		}

		// Check if already deployed with same hash
		if !force {
			alreadyDeployed := false
			for _, s := range tracker.Applied {
				if s.Name == scriptID && s.MD5 == srcMD5 {
					alreadyDeployed = true
					break
				}
			}
			if alreadyDeployed {
				continue // Skip - same file already deployed
			}
		}

		// Deploy the script
		destPath := filepath.Join(cfg.TrinityCore.ScriptsPath, script.FileName)
		if err := copyFile(script.SourcePath, destPath); err != nil {
			return deployed, fmt.Errorf("copy %s: %w", script.FileName, err)
		}

		// Update tracker (replace existing entry or add new)
		found := false
		for i, s := range tracker.Applied {
			if s.Name == scriptID {
				tracker.Applied[i].MD5 = srcMD5
				tracker.Applied[i].AppliedAt = time.Now().Format(time.RFC3339)
				tracker.Applied[i].AppliedBy = "thorium build"
				found = true
				break
			}
		}
		if !found {
			tracker.Applied = append(tracker.Applied, scriptDeployed{
				Name:      scriptID,
				MD5:       srcMD5,
				AppliedAt: time.Now().Format(time.RFC3339),
				AppliedBy: "thorium build",
			})
		}

		fmt.Printf("[%s] Deployed %s\n", script.ModName, script.FileName)
		deployed++
	}

	// Check for removed scripts (in tracker but not in current list)
	currentScripts := make(map[string]bool)
	for _, script := range scriptFiles {
		scriptID := fmt.Sprintf("%s/%s", script.ModName, script.FileName)
		currentScripts[scriptID] = true
	}
	
	removed := 0
	var updatedApplied []scriptDeployed
	for _, s := range tracker.Applied {
		if currentScripts[s.Name] {
			updatedApplied = append(updatedApplied, s)
		} else {
			// Script was removed from mods
			fmt.Printf("  Removed %s (no longer in mods)\n", s.Name)
			removed++
		}
	}
	tracker.Applied = updatedApplied

	// Save tracker
	if err := saveScriptTracker(workspaceRoot, tracker); err != nil {
		return deployed, fmt.Errorf("save script tracker: %w", err)
	}

	if deployed > 0 {
		fmt.Printf("  Deployed %d new/changed script(s) to TrinityCore\n", deployed)
	} else if removed == 0 {
		fmt.Println("  No new/changed scripts to deploy")
	}

	// Only regenerate loader if scripts changed
	// Note: We don't generate CMakeLists.txt because TrinityCore's CollectSourceFiles
	// automatically finds all .cpp files in the Custom directory
	if deployed > 0 || removed > 0 {
		if err := generateLoaderScript(cfg.TrinityCore.ScriptsPath, scriptFiles); err != nil {
			return deployed, fmt.Errorf("generate loader script: %w", err)
		}
		fmt.Println("  Generated custom_script_loader.cpp")
	}

	return deployed, nil
}

// ScriptFile represents a script file with metadata
type ScriptFile struct {
	ModName    string
	FileName   string
	SourcePath string
	AddSCFunc  string // Function name like AddSC_spell_fire_blast
}

// collectScriptFiles finds all .cpp files in a directory
func collectScriptFiles(dir, modName string) ([]ScriptFile, error) {
	var files []ScriptFile

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cpp") {
			continue
		}

		// Extract AddSC function name from file
		sourcePath := filepath.Join(dir, entry.Name())
		addSCFunc, err := extractAddSCFunc(sourcePath)
		if err != nil || addSCFunc == "" {
			// Skip files without AddSC function
			continue
		}

		files = append(files, ScriptFile{
			ModName:    modName,
			FileName:   entry.Name(),
			SourcePath: sourcePath,
			AddSCFunc:  addSCFunc,
		})
	}

	return files, nil
}

// extractAddSCFunc extracts the AddSC_* function name from a script file
func extractAddSCFunc(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	content := string(data)

	// Look for pattern: void AddSC_something()
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "void AddSC_") {
			// Extract function name
			if idx := strings.Index(line, "("); idx != -1 {
				funcDef := line[:idx]
				funcName := strings.TrimPrefix(funcDef, "void ")
				return strings.TrimSpace(funcName), nil
			}
		}
	}

	return "", fmt.Errorf("no AddSC function found in %s", filePath)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}

// generateLoaderScript creates the loader that calls all AddSC functions
func generateLoaderScript(scriptsPath string, scripts []ScriptFile) error {
	var forwardDecls []string
	var calls []string

	for _, script := range scripts {
		forwardDecls = append(forwardDecls, fmt.Sprintf("void %s();", script.AddSCFunc))
		calls = append(calls, fmt.Sprintf("    %s(); // %s", script.AddSCFunc, script.ModName))
	}

	content := fmt.Sprintf(`// Auto-generated by Thorium
// Do not edit manually

#include "ScriptMgr.h"

// Forward declarations
%s

// This function is called by ScriptMgr to load all custom scripts
void AddCustomScripts()
{
%s
}
`, strings.Join(forwardDecls, "\n"), strings.Join(calls, "\n"))

	loaderPath := filepath.Join(scriptsPath, "custom_script_loader.cpp")
	return os.WriteFile(loaderPath, []byte(content), 0644)
}

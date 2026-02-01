// Copyright (c) 2025 Client Patcher
//
// Client Patcher is licensed under the MIT License.
// See the LICENSE file for details.

package patcher

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"thorium-cli/assets"
)

// MD5 hash of clean WoW 3.3.5a (12340) client
const CleanClientMD5 = "45892bdedd0ad70aed4ccd22d9fb5984"

// PatchOptions contains options for applying patches
type PatchOptions struct {
	WowExePath      string
	BackupPath      string
	OutputPath      string
	SelectedPatches []string // If empty, applies all patches
	Verbose         bool
}

// copyOnNoTarget copies source to dest only if dest doesn't exist
func copyOnNoTarget(sourcePath, destPath string) error {
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		source, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("open source: %w", err)
		}
		defer source.Close()

		dest, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("create dest: %w", err)
		}
		defer dest.Close()

		if _, err := io.Copy(dest, source); err != nil {
			return fmt.Errorf("copy: %w", err)
		}
	}
	return nil
}

// calculateMD5 calculates MD5 hash of a file
func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// logVerbose prints message if verbose is enabled
func logVerbose(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Printf("[PATCHER] "+format+"\n", args...)
	}
}

// ApplyPatches applies binary patches to Wow.exe
func ApplyPatches(opts PatchOptions) error {
	// Set defaults
	if opts.BackupPath == "" {
		opts.BackupPath = opts.WowExePath + ".clean"
	}
	if opts.OutputPath == "" {
		opts.OutputPath = opts.WowExePath
	}

	logVerbose(opts.Verbose, "Applying client patches...")

	// Step 1: Verify Wow.exe exists
	if _, err := os.Stat(opts.WowExePath); os.IsNotExist(err) {
		return fmt.Errorf("Wow.exe not found at: %s", opts.WowExePath)
	}

	// Step 2: Create backup if it doesn't exist
	if err := copyOnNoTarget(opts.WowExePath, opts.BackupPath); err != nil {
		return fmt.Errorf("create backup: %w", err)
	}
	logVerbose(opts.Verbose, "Backup created/verified: %s", opts.BackupPath)

	// Step 3: Read clean source
	wowbin, err := os.ReadFile(opts.BackupPath)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	hash, err := calculateMD5(opts.BackupPath)
	if err != nil {
		return fmt.Errorf("calculate hash: %w", err)
	}

	// Step 4: Verify clean source
	if hash != CleanClientMD5 {
		logVerbose(opts.Verbose, "Warning: Backup hash (%s) doesn't match clean client", hash)

		// Check if Wow.exe itself is clean
		exeHash, err := calculateMD5(opts.WowExePath)
		if err != nil {
			return fmt.Errorf("calculate exe hash: %w", err)
		}

		if exeHash == CleanClientMD5 {
			logVerbose(opts.Verbose, "Wow.exe is clean, updating backup")
			wowbin, err = os.ReadFile(opts.WowExePath)
			if err != nil {
				return fmt.Errorf("read exe: %w", err)
			}
			if err := os.WriteFile(opts.BackupPath, wowbin, 0644); err != nil {
				return fmt.Errorf("update backup: %w", err)
			}
		} else {
			logVerbose(opts.Verbose, "Warning: Wow.exe hash (%s) also doesn't match clean client", exeHash)
			logVerbose(opts.Verbose, "Proceeding anyway, but patches may not work correctly")
		}
	} else {
		logVerbose(opts.Verbose, "Source client hash verified: %s", hash)
	}

	// Step 5: Get patches
	logVerbose(opts.Verbose, "Loading patch definitions...")
	allPatches := GetClientPatches()

	// Filter to requested patches (or use all if none specified)
	var patchesToApply []PatchCategory
	if len(opts.SelectedPatches) == 0 {
		patchesToApply = allPatches
	} else {
		for _, cat := range allPatches {
			for _, requestedName := range opts.SelectedPatches {
				if cat.Name == requestedName {
					patchesToApply = append(patchesToApply, cat)
					break
				}
			}
		}
		if len(patchesToApply) == 0 {
			return fmt.Errorf("no patches found for names: %v", opts.SelectedPatches)
		}
	}

	// Step 6: Apply patches
	logVerbose(opts.Verbose, "Applying %d patch(es)...", len(patchesToApply))
	for _, cat := range patchesToApply {
		logVerbose(opts.Verbose, "  Applying: %s", cat.Name)
		for _, patch := range cat.Patches {
			for offset, value := range patch.Values {
				address := int(patch.Address) + offset
				if address >= len(wowbin) {
					return fmt.Errorf("patch address 0x%x exceeds file size (0x%x)", address, len(wowbin))
				}
				wowbin[address] = value
			}
		}
	}

	// Step 7: Write patched file
	logVerbose(opts.Verbose, "Writing patched Wow.exe to: %s", opts.OutputPath)
	if err := os.WriteFile(opts.OutputPath, wowbin, 0755); err != nil {
		return fmt.Errorf("write patched exe: %w", err)
	}

	// Step 8: Copy ClientExtensions.dll if custom-packets patch was applied
	needsDLL := false
	if len(opts.SelectedPatches) == 0 {
		needsDLL = true // applying all patches
	} else {
		for _, name := range opts.SelectedPatches {
			if name == "custom-packets" {
				needsDLL = true
				break
			}
		}
	}

	if needsDLL {
		clientDir := filepath.Dir(opts.WowExePath)
		dllPath := filepath.Join(clientDir, "ClientExtensions.dll")
		
		logVerbose(opts.Verbose, "Copying ClientExtensions.dll to: %s", dllPath)
		
		// Get embedded DLL
		dllData := assets.GetClientExtensionsDLL()
		if len(dllData) == 0 {
			return fmt.Errorf("ClientExtensions.dll is empty in embedded assets")
		}
		
		logVerbose(opts.Verbose, "Using embedded ClientExtensions.dll (%d bytes)", len(dllData))
		
		if err := os.WriteFile(dllPath, dllData, 0644); err != nil {
			return fmt.Errorf("write ClientExtensions.dll: %w", err)
		}
		logVerbose(opts.Verbose, "ClientExtensions.dll installed successfully")
	}

	logVerbose(opts.Verbose, "Patches applied successfully!")
	return nil
}

// RestoreFromBackup restores Wow.exe from backup
func RestoreFromBackup(wowExePath, backupPath string) error {
	if backupPath == "" {
		backupPath = wowExePath + ".clean"
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	if err := os.WriteFile(wowExePath, data, 0755); err != nil {
		return fmt.Errorf("write exe: %w", err)
	}

	return nil
}

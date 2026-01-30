// Copyright (c) 2025 Thorium
//go:build cgo

package mpq

/*
#cgo CFLAGS: -I${SRCDIR}/../../../stormlib/src
#cgo LDFLAGS: -L${SRCDIR}/../../../stormlib/build -lstorm -lz -lbz2 -lstdc++ -lc++

#include <stdlib.h>
#include <stdbool.h>
#include "StormLib.h"

// Wrapper functions for Go compatibility
static int storm_create_archive(const char* path, unsigned int maxFiles, void** handle) {
    return SFileCreateArchive(path, 0, maxFiles, (HANDLE*)handle) ? 0 : -1;
}

static int storm_close_archive(void* handle) {
    return SFileCloseArchive((HANDLE)handle) ? 0 : -1;
}

static int storm_flush_archive(void* handle) {
    return SFileFlushArchive((HANDLE)handle) ? 0 : -1;
}

static int storm_add_file(void* handle, const char* srcPath, const char* mpqPath) {
    // Use simple SFileAddFile like tswow (no explicit compression flags)
    return SFileAddFile((HANDLE)handle, srcPath, mpqPath, 0) ? 0 : -1;
}

static int storm_open_archive(const char* path, void** handle) {
    return SFileOpenArchive(path, 0, 0, (HANDLE*)handle) ? 0 : -1;
}

static int storm_extract_file(void* handle, const char* mpqPath, const char* destPath) {
    return SFileExtractFile((HANDLE)handle, mpqPath, destPath, 0) ? 0 : -1;
}

static int storm_has_file(void* handle, const char* mpqPath) {
    return SFileHasFile((HANDLE)handle, mpqPath) ? 1 : 0;
}
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unsafe"
)

// stormLibAvailable indicates if StormLib is linked
const stormLibAvailable = true

// StormArchive wraps a StormLib MPQ handle
type StormArchive struct {
	handle     unsafe.Pointer
	path       string
	tempPath   string // Temp file path for safe writing (like tswow)
	mode       string
}

// CreateWithStorm creates an MPQ using StormLib
// Uses temp file approach like tswow for crash safety
func CreateWithStorm(path string, maxFiles int) (*StormArchive, error) {
	// Create temp file in same directory
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, "thorium_mpq_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()

	cPath := C.CString(tempPath)
	defer C.free(unsafe.Pointer(cPath))

	var handle unsafe.Pointer
	result := C.storm_create_archive(cPath, C.uint(maxFiles), &handle)
	if result != 0 {
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to create archive: %s", path)
	}

	return &StormArchive{
		handle:   handle,
		path:     path,
		tempPath: tempPath,
		mode:     "w",
	}, nil
}

// OpenWithStorm opens an MPQ using StormLib
func OpenWithStorm(path string) (*StormArchive, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var handle unsafe.Pointer
	result := C.storm_open_archive(cPath, &handle)
	if result != 0 {
		return nil, fmt.Errorf("failed to open archive: %s", path)
	}

	return &StormArchive{
		handle: handle,
		path:   path,
		mode:   "r",
	}, nil
}

// AddFile adds a file to the archive
func (a *StormArchive) AddFile(srcPath, mpqPath string) error {
	if a.mode != "w" {
		return fmt.Errorf("archive not opened for writing")
	}

	cSrc := C.CString(srcPath)
	cMpq := C.CString(mpqPath)
	defer C.free(unsafe.Pointer(cSrc))
	defer C.free(unsafe.Pointer(cMpq))

	result := C.storm_add_file(a.handle, cSrc, cMpq)
	if result != 0 {
		return fmt.Errorf("failed to add file: %s", srcPath)
	}

	return nil
}

// ExtractFile extracts a file from the archive
func (a *StormArchive) ExtractFile(mpqPath, destPath string) error {
	if a.mode != "r" {
		return fmt.Errorf("archive not opened for reading")
	}

	cMpq := C.CString(mpqPath)
	cDest := C.CString(destPath)
	defer C.free(unsafe.Pointer(cMpq))
	defer C.free(unsafe.Pointer(cDest))

	result := C.storm_extract_file(a.handle, cMpq, cDest)
	if result != 0 {
		return fmt.Errorf("failed to extract file: %s", mpqPath)
	}

	return nil
}

// HasFile checks if a file exists in the archive
func (a *StormArchive) HasFile(mpqPath string) bool {
	cMpq := C.CString(mpqPath)
	defer C.free(unsafe.Pointer(cMpq))

	return C.storm_has_file(a.handle, cMpq) == 1
}

// Close closes the archive
// For write mode: flushes, closes, and copies temp file to final path (like tswow)
func (a *StormArchive) Close() error {
	if a.handle == nil {
		return nil
	}

	// Flush and close
	if a.mode == "w" {
		C.storm_flush_archive(a.handle)
	}

	result := C.storm_close_archive(a.handle)
	a.handle = nil
	if result != 0 {
		if a.tempPath != "" {
			os.Remove(a.tempPath)
		}
		return fmt.Errorf("failed to close archive")
	}

	// For write mode: copy temp file to final destination (like tswow)
	if a.mode == "w" && a.tempPath != "" {
		defer os.Remove(a.tempPath)

		// Remove old output file if exists
		os.Remove(a.path)

		// Copy temp to final
		if err := stormCopyFile(a.tempPath, a.path); err != nil {
			return fmt.Errorf("failed to save archive: %w", err)
		}
	}

	return nil
}

// stormCopyFile copies src to dst
func stormCopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

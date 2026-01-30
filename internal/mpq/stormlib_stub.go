// Copyright (c) 2025 Thorium
//go:build !cgo

package mpq

import "fmt"

// stormLibAvailable indicates if StormLib is linked
const stormLibAvailable = false

// StormArchive stub for non-cgo builds
type StormArchive struct {
	handle   interface{}
	path     string
	tempPath string
	mode     string
}

// CreateWithStorm stub
func CreateWithStorm(path string, maxFiles int) (*StormArchive, error) {
	return nil, fmt.Errorf("StormLib not available (built without cgo)")
}

// OpenWithStorm stub
func OpenWithStorm(path string) (*StormArchive, error) {
	return nil, fmt.Errorf("StormLib not available (built without cgo)")
}

// AddFile stub
func (a *StormArchive) AddFile(srcPath, mpqPath string) error {
	return fmt.Errorf("StormLib not available")
}

// ExtractFile stub
func (a *StormArchive) ExtractFile(mpqPath, destPath string) error {
	return fmt.Errorf("StormLib not available")
}

// HasFile stub
func (a *StormArchive) HasFile(mpqPath string) bool {
	return false
}

// Close stub
func (a *StormArchive) Close() error {
	return nil
}

// Copyright (c) 2025 Thorium
//
// Assets package embeds static files into the binary

package assets

import _ "embed"

//go:embed ClientExtensions.dll
var ClientExtensionsDLL []byte

// GetClientExtensionsDLL returns the embedded ClientExtensions.dll
func GetClientExtensionsDLL() []byte {
	return ClientExtensionsDLL
}

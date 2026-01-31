// Copyright (c) 2025 Thorium
// Ported from TSWoW CustomPackets

package custompackets

// Opcode represents a custom packet opcode (uint16)
type Opcode uint16

// TotalSize represents the total size of a packet (uint32)
type TotalSize uint32

// ChunkSize represents the size of a single chunk (uint16)
type ChunkSize uint16

// ChunkCount represents the number of chunks (uint16)
type ChunkCount uint16

const (
	// TotalSizeNpos represents an invalid/unknown total size
	TotalSizeNpos TotalSize = 0xFFFF

	// MaxFragmentSize is the maximum size of a single packet fragment
	MaxFragmentSize ChunkSize = 30000

	// MinFragmentSize is the minimum size for packet fragments
	MinFragmentSize ChunkSize = 25000

	// BufferQuota is the default buffer quota (~8MB)
	BufferQuota TotalSize = 8000000

	// ServerToClientOpcode is the base opcode for server->client custom packets
	// Using CMSG_EMOTE (0x102) because client doesn't accept higher IDs
	ServerToClientOpcode uint16 = 0x102

	// ClientToServerOpcode is the base opcode for client->server custom packets
	ClientToServerOpcode uint16 = 0x51F
)

// CustomPacketHeader represents the header of a custom packet chunk
type CustomPacketHeader struct {
	FragmentID  ChunkCount // Current fragment ID
	TotalFrags  ChunkCount // Total number of fragments
	Opcode      Opcode     // Custom packet opcode
}

// HeaderSize returns the size of the header in bytes
func HeaderSize() ChunkSize {
	return ChunkSize(6) // 2 + 2 + 2 bytes
}

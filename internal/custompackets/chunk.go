// Copyright (c) 2025 Thorium
// Ported from TSWoW CustomPackets

package custompackets

import (
	"encoding/binary"
	"fmt"
)

// Chunk represents a single packet fragment
type Chunk struct {
	data []byte
}

// NewChunk creates a new chunk with the given size
func NewChunk(size ChunkSize) *Chunk {
	return &Chunk{
		data: make([]byte, size),
	}
}

// NewChunkFromData creates a chunk from existing data
func NewChunkFromData(data []byte) *Chunk {
	chunk := &Chunk{
		data: make([]byte, len(data)),
	}
	copy(chunk.data, data)
	return chunk
}

// Data returns the raw data of the chunk
func (c *Chunk) Data() []byte {
	return c.data
}

// Header returns the header of the chunk
func (c *Chunk) Header() *CustomPacketHeader {
	if len(c.data) < int(HeaderSize()) {
		return nil
	}
	return &CustomPacketHeader{
		FragmentID: ChunkCount(binary.LittleEndian.Uint16(c.data[0:2])),
		TotalFrags: ChunkCount(binary.LittleEndian.Uint16(c.data[2:4])),
		Opcode:     Opcode(binary.LittleEndian.Uint16(c.data[4:6])),
	}
}

// SetHeader sets the header of the chunk
func (c *Chunk) SetHeader(header *CustomPacketHeader) {
	if len(c.data) < int(HeaderSize()) {
		return
	}
	binary.LittleEndian.PutUint16(c.data[0:2], uint16(header.FragmentID))
	binary.LittleEndian.PutUint16(c.data[2:4], uint16(header.TotalFrags))
	binary.LittleEndian.PutUint16(c.data[4:6], uint16(header.Opcode))
}

// FullSize returns the total size of the chunk
func (c *Chunk) FullSize() ChunkSize {
	return ChunkSize(len(c.data))
}

// Size returns the payload size (excluding header)
func (c *Chunk) Size() ChunkSize {
	return ChunkSize(len(c.data)) - HeaderSize()
}

// RemBytes returns the remaining bytes from the given offset
func (c *Chunk) RemBytes(idx ChunkSize) ChunkSize {
	size := c.Size()
	if idx >= size {
		return 0
	}
	return size - idx
}

// Offset returns a pointer to data at the given offset (after header)
func (c *Chunk) Offset(offset ChunkSize) []byte {
	baseOffset := int(HeaderSize() + offset)
	if baseOffset >= len(c.data) {
		return nil
	}
	return c.data[baseOffset:]
}

// WriteBytes writes bytes at the given offset
func (c *Chunk) WriteBytes(idx ChunkSize, size ChunkSize, value []byte) error {
	offset := int(HeaderSize() + idx)
	if offset+int(size) > len(c.data) {
		return fmt.Errorf("write exceeds chunk size")
	}
	copy(c.data[offset:offset+int(size)], value)
	return nil
}

// ReadBytes reads bytes from the given offset
func (c *Chunk) ReadBytes(idx ChunkSize, size ChunkSize) ([]byte, error) {
	offset := int(HeaderSize() + idx)
	if offset+int(size) > len(c.data) {
		return nil, fmt.Errorf("read exceeds chunk size")
	}
	result := make([]byte, size)
	copy(result, c.data[offset:offset+int(size)])
	return result, nil
}

// Increase increases the size of the chunk
func (c *Chunk) Increase(size ChunkSize) {
	newData := make([]byte, len(c.data)+int(size))
	copy(newData, c.data)
	c.data = newData
}

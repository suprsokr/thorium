// Copyright (c) 2025 Thorium
// Ported from TSWoW CustomPackets

package custompackets

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Base provides base functionality for packet reading/writing
type Base struct {
	chunks       []*Chunk
	size         TotalSize
	maxChunkSize ChunkSize
	globalIdx    TotalSize  // global read index
	idx          ChunkSize  // chunk read index
	chunk        ChunkCount // chunk to read
	opcode       Opcode
}

// NewBase creates a new packet base
func NewBase(opcode Opcode, maxChunkSize ChunkSize, initialSize TotalSize) *Base {
	base := &Base{
		chunks:       make([]*Chunk, 0),
		size:         0,
		maxChunkSize: maxChunkSize,
		globalIdx:    0,
		idx:          0,
		chunk:        0,
		opcode:       opcode,
	}
	if initialSize > 0 {
		base.Increase(initialSize)
	}
	return base
}

// BuildMessages builds the message chunks
func (b *Base) BuildMessages() []*Chunk {
	return b.chunks
}

// Reset resets the read position
func (b *Base) Reset() {
	b.globalIdx = 0
	b.idx = 0
	b.chunk = 0
}

// Clear clears all chunks
func (b *Base) Clear() {
	b.chunks = make([]*Chunk, 0)
	b.size = 0
	b.Reset()
}

// Push adds a chunk to the base
func (b *Base) Push(chunk *Chunk) {
	b.chunks = append(b.chunks, chunk)
	b.size += TotalSize(chunk.Size())
}

// Size returns the total size
func (b *Base) Size() TotalSize {
	return b.size
}

// Chunk returns the chunk at the given index
func (b *Base) Chunk(index ChunkCount) *Chunk {
	if int(index) >= len(b.chunks) {
		return nil
	}
	return b.chunks[index]
}

// ChunkSize returns the size of the chunk at the given index
func (b *Base) ChunkSize(index ChunkCount) ChunkSize {
	chunk := b.Chunk(index)
	if chunk == nil {
		return 0
	}
	return chunk.Size()
}

// ChunkCount returns the number of chunks
func (b *Base) ChunkCount() ChunkCount {
	return ChunkCount(len(b.chunks))
}

// Opcode returns the opcode
func (b *Base) Opcode() Opcode {
	return b.opcode
}

// MaxWritableChunkSize returns the maximum writable size per chunk
func (b *Base) MaxWritableChunkSize() ChunkSize {
	return b.maxChunkSize - HeaderSize()
}

// Increase increases the total size by allocating more chunks
func (b *Base) Increase(increase TotalSize) {
	for increase > 0 {
		maxWritable := TotalSize(b.MaxWritableChunkSize())
		toAlloc := increase
		if toAlloc > maxWritable {
			toAlloc = maxWritable
		}

		chunk := NewChunk(HeaderSize() + ChunkSize(toAlloc))
		b.chunks = append(b.chunks, chunk)
		b.size += toAlloc
		increase -= toAlloc
	}

	// Set headers
	totalFrags := ChunkCount(len(b.chunks))
	for i, chunk := range b.chunks {
		chunk.SetHeader(&CustomPacketHeader{
			FragmentID: ChunkCount(i),
			TotalFrags: totalFrags,
			Opcode:     b.opcode,
		})
	}
}

// WriteBytes writes bytes to the packet
func (b *Base) WriteBytes(size TotalSize, bytes []byte) error {
	if TotalSize(len(bytes)) < size {
		return fmt.Errorf("byte array too small")
	}

	written := TotalSize(0)
	for written < size {
		if int(b.chunk) >= len(b.chunks) {
			return fmt.Errorf("write exceeds packet size")
		}

		currentChunk := b.chunks[b.chunk]
		rem := currentChunk.RemBytes(b.idx)
		toWrite := size - written
		if TotalSize(rem) < toWrite {
			toWrite = TotalSize(rem)
		}

		if err := currentChunk.WriteBytes(b.idx, ChunkSize(toWrite), bytes[written:written+toWrite]); err != nil {
			return err
		}

		written += toWrite
		b.incIdx(ChunkSize(toWrite))
	}

	return nil
}

// ReadBytes reads bytes from the packet
func (b *Base) ReadBytes(size TotalSize) ([]byte, error) {
	if b.size-b.globalIdx < size {
		return nil, fmt.Errorf("read exceeds packet size")
	}

	result := make([]byte, size)
	read := TotalSize(0)

	for read < size {
		if int(b.chunk) >= len(b.chunks) {
			return nil, fmt.Errorf("read exceeds chunks")
		}

		currentChunk := b.chunks[b.chunk]
		rem := currentChunk.RemBytes(b.idx)
		toRead := size - read
		if TotalSize(rem) < toRead {
			toRead = TotalSize(rem)
		}

		data, err := currentChunk.ReadBytes(b.idx, ChunkSize(toRead))
		if err != nil {
			return nil, err
		}

		copy(result[read:], data)
		read += toRead
		b.incIdx(ChunkSize(toRead))
	}

	return result, nil
}

// incIdx increments the read/write index
func (b *Base) incIdx(amount ChunkSize) {
	b.idx += amount
	b.globalIdx += TotalSize(amount)

	for int(b.chunk) < len(b.chunks) {
		currentChunk := b.chunks[b.chunk]
		if b.idx >= currentChunk.Size() {
			b.idx -= currentChunk.Size()
			b.chunk++
		} else {
			break
		}
	}
}

// Write writes a value of type T
func (b *Base) Write(value interface{}) error {
	switch v := value.(type) {
	case uint8:
		return b.WriteBytes(1, []byte{v})
	case int8:
		return b.WriteBytes(1, []byte{byte(v)})
	case uint16:
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, v)
		return b.WriteBytes(2, buf)
	case int16:
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(v))
		return b.WriteBytes(2, buf)
	case uint32:
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, v)
		return b.WriteBytes(4, buf)
	case int32:
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(v))
		return b.WriteBytes(4, buf)
	case uint64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, v)
		return b.WriteBytes(8, buf)
	case int64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(v))
		return b.WriteBytes(8, buf)
	case float32:
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, math.Float32bits(v))
		return b.WriteBytes(4, buf)
	case float64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
		return b.WriteBytes(8, buf)
	default:
		return fmt.Errorf("unsupported type")
	}
}

// Read reads a value with default
func (b *Base) Read(def interface{}) (interface{}, error) {
	switch def.(type) {
	case uint8:
		data, err := b.ReadBytes(1)
		if err != nil {
			return def, err
		}
		return data[0], nil
	case int8:
		data, err := b.ReadBytes(1)
		if err != nil {
			return def, err
		}
		return int8(data[0]), nil
	case uint16:
		data, err := b.ReadBytes(2)
		if err != nil {
			return def, err
		}
		return binary.LittleEndian.Uint16(data), nil
	case int16:
		data, err := b.ReadBytes(2)
		if err != nil {
			return def, err
		}
		return int16(binary.LittleEndian.Uint16(data)), nil
	case uint32:
		data, err := b.ReadBytes(4)
		if err != nil {
			return def, err
		}
		return binary.LittleEndian.Uint32(data), nil
	case int32:
		data, err := b.ReadBytes(4)
		if err != nil {
			return def, err
		}
		return int32(binary.LittleEndian.Uint32(data)), nil
	case uint64:
		data, err := b.ReadBytes(8)
		if err != nil {
			return def, err
		}
		return binary.LittleEndian.Uint64(data), nil
	case int64:
		data, err := b.ReadBytes(8)
		if err != nil {
			return def, err
		}
		return int64(binary.LittleEndian.Uint64(data)), nil
	case float32:
		data, err := b.ReadBytes(4)
		if err != nil {
			return def, err
		}
		return math.Float32frombits(binary.LittleEndian.Uint32(data)), nil
	case float64:
		data, err := b.ReadBytes(8)
		if err != nil {
			return def, err
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(data)), nil
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}

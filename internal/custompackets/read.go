// Copyright (c) 2025 Thorium
// Ported from TSWoW CustomPackets

package custompackets

import "fmt"

// Reader provides packet reading functionality
type Reader struct {
	*Base
}

// NewReader creates a new packet reader
func NewReader(opcode Opcode) *Reader {
	return &Reader{
		Base: NewBase(opcode, MaxFragmentSize, 0),
	}
}

// ReadUInt8 reads a uint8
func (r *Reader) ReadUInt8(def uint8) uint8 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(uint8)
}

// ReadInt8 reads an int8
func (r *Reader) ReadInt8(def int8) int8 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(int8)
}

// ReadUInt16 reads a uint16
func (r *Reader) ReadUInt16(def uint16) uint16 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(uint16)
}

// ReadInt16 reads an int16
func (r *Reader) ReadInt16(def int16) int16 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(int16)
}

// ReadUInt32 reads a uint32
func (r *Reader) ReadUInt32(def uint32) uint32 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(uint32)
}

// ReadInt32 reads an int32
func (r *Reader) ReadInt32(def int32) int32 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(int32)
}

// ReadUInt64 reads a uint64
func (r *Reader) ReadUInt64(def uint64) uint64 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(uint64)
}

// ReadInt64 reads an int64
func (r *Reader) ReadInt64(def int64) int64 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(int64)
}

// ReadFloat reads a float32
func (r *Reader) ReadFloat(def float32) float32 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(float32)
}

// ReadDouble reads a float64
func (r *Reader) ReadDouble(def float64) float64 {
	val, err := r.Read(def)
	if err != nil {
		return def
	}
	return val.(float64)
}

// ReadString reads a null-terminated string
func (r *Reader) ReadString() (string, error) {
	result := make([]byte, 0)
	for {
		if r.globalIdx >= r.size {
			break
		}
		b, err := r.ReadBytes(1)
		if err != nil {
			return string(result), err
		}
		if b[0] == 0 {
			break
		}
		result = append(result, b[0])
	}
	return string(result), nil
}

// ReadLengthString reads a length-prefixed string (uint32 length + string bytes)
// This matches TSWoW's ReadString format
func (r *Reader) ReadLengthString(def string) string {
	length := r.ReadUInt32(uint32(TotalSizeNpos))
	if length == uint32(TotalSizeNpos) {
		return def
	}
	if length == 0 {
		return ""
	}
	data, err := r.ReadBytes(TotalSize(length))
	if err != nil {
		return def
	}
	return string(data)
}

// ReadStringNullTerm reads a null-terminated string (alias for ReadString)
func (r *Reader) ReadStringNullTerm() (string, error) {
	return r.ReadString()
}

// ReadBytesN reads a specified number of bytes
func (r *Reader) ReadBytesN(size TotalSize) ([]byte, error) {
	return r.ReadBytes(size)
}

// Print prints the packet contents (for debugging)
func (r *Reader) Print() string {
	result := fmt.Sprintf("Reader{opcode=%d, size=%d, chunks=%d}\n", r.Opcode(), r.Size(), r.ChunkCount())
	for i := ChunkCount(0); i < r.ChunkCount(); i++ {
		chunk := r.Chunk(i)
		if chunk != nil {
			header := chunk.Header()
			result += fmt.Sprintf("  Chunk %d: frag=%d/%d, size=%d\n",
				i, header.FragmentID, header.TotalFrags, chunk.Size())
		}
	}
	return result
}

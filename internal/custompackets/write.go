// Copyright (c) 2025 Thorium
// Ported from TSWoW CustomPackets

package custompackets

import "fmt"

// Writer provides packet writing functionality
type Writer struct {
	*Base
}

// NewWriter creates a new packet writer
func NewWriter(opcode Opcode, size TotalSize) *Writer {
	return &Writer{
		Base: NewBase(opcode, MaxFragmentSize, size),
	}
}

// WriteUInt8 writes a uint8
func (w *Writer) WriteUInt8(value uint8) error {
	return w.Write(value)
}

// WriteInt8 writes an int8
func (w *Writer) WriteInt8(value int8) error {
	return w.Write(value)
}

// WriteUInt16 writes a uint16
func (w *Writer) WriteUInt16(value uint16) error {
	return w.Write(value)
}

// WriteInt16 writes an int16
func (w *Writer) WriteInt16(value int16) error {
	return w.Write(value)
}

// WriteUInt32 writes a uint32
func (w *Writer) WriteUInt32(value uint32) error {
	return w.Write(value)
}

// WriteInt32 writes an int32
func (w *Writer) WriteInt32(value int32) error {
	return w.Write(value)
}

// WriteUInt64 writes a uint64
func (w *Writer) WriteUInt64(value uint64) error {
	return w.Write(value)
}

// WriteInt64 writes an int64
func (w *Writer) WriteInt64(value int64) error {
	return w.Write(value)
}

// WriteFloat writes a float32
func (w *Writer) WriteFloat(value float32) error {
	return w.Write(value)
}

// WriteDouble writes a float64
func (w *Writer) WriteDouble(value float64) error {
	return w.Write(value)
}

// WriteString writes a null-terminated string
func (w *Writer) WriteString(value string) error {
	data := []byte(value)
	data = append(data, 0) // null terminator
	return w.WriteBytes(TotalSize(len(data)), data)
}

// WriteLengthString writes a length-prefixed string (uint32 length + string bytes, no null terminator)
// This matches TSWoW's WriteString format
func (w *Writer) WriteLengthString(value string) error {
	data := []byte(value)
	if err := w.WriteUInt32(uint32(len(data))); err != nil {
		return err
	}
	if len(data) > 0 {
		return w.WriteBytes(TotalSize(len(data)), data)
	}
	return nil
}

// WriteLengthStringN writes a length-prefixed string with a maximum length
// If the string exceeds maxLen, it is truncated
func (w *Writer) WriteLengthStringN(value string, maxLen uint32) error {
	data := []byte(value)
	if uint32(len(data)) > maxLen {
		data = data[:maxLen]
	}
	if err := w.WriteUInt32(uint32(len(data))); err != nil {
		return err
	}
	if len(data) > 0 {
		return w.WriteBytes(TotalSize(len(data)), data)
	}
	return nil
}

// WriteStringNullTerm writes a null-terminated string (alias for WriteString)
func (w *Writer) WriteStringNullTerm(value string) error {
	return w.WriteString(value)
}

// WriteStringNullTermN writes a null-terminated string with a maximum length (excluding null)
func (w *Writer) WriteStringNullTermN(value string, maxLen uint32) error {
	data := []byte(value)
	if uint32(len(data)) > maxLen {
		data = data[:maxLen]
	}
	data = append(data, 0) // null terminator
	return w.WriteBytes(TotalSize(len(data)), data)
}

// GetChunks returns all message chunks ready to send
func (w *Writer) GetChunks() []*Chunk {
	return w.BuildMessages()
}

// Print prints the packet contents (for debugging)
func (w *Writer) Print() string {
	result := fmt.Sprintf("Writer{opcode=%d, size=%d, chunks=%d}\n", w.Opcode(), w.Size(), w.ChunkCount())
	for i := ChunkCount(0); i < w.ChunkCount(); i++ {
		chunk := w.Chunk(i)
		if chunk != nil {
			header := chunk.Header()
			result += fmt.Sprintf("  Chunk %d: frag=%d/%d, size=%d\n",
				i, header.FragmentID, header.TotalFrags, chunk.Size())
		}
	}
	return result
}

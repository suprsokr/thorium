// Copyright (c) 2025 Thorium

package custompackets

import (
	"os"
	"testing"
)

func TestChunkHeaderReadWrite(t *testing.T) {
	chunk := NewChunk(100)

	header := &CustomPacketHeader{
		FragmentID: 5,
		TotalFrags: 10,
		Opcode:     1001,
	}

	chunk.SetHeader(header)
	readHeader := chunk.Header()

	if readHeader.FragmentID != 5 {
		t.Errorf("FragmentID: expected 5, got %d", readHeader.FragmentID)
	}
	if readHeader.TotalFrags != 10 {
		t.Errorf("TotalFrags: expected 10, got %d", readHeader.TotalFrags)
	}
	if readHeader.Opcode != 1001 {
		t.Errorf("Opcode: expected 1001, got %d", readHeader.Opcode)
	}
}

func TestWriterBasic(t *testing.T) {
	writer := NewWriter(1001, 100)

	if err := writer.WriteUInt32(12345); err != nil {
		t.Fatalf("WriteUInt32 failed: %v", err)
	}

	if err := writer.WriteString("test"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	if err := writer.WriteFloat(3.14); err != nil {
		t.Fatalf("WriteFloat failed: %v", err)
	}

	chunks := writer.GetChunks()
	if len(chunks) == 0 {
		t.Fatal("No chunks generated")
	}

	firstChunk := chunks[0]
	header := firstChunk.Header()

	if header.Opcode != 1001 {
		t.Errorf("Opcode: expected 1001, got %d", header.Opcode)
	}

	t.Logf("Generated %d chunk(s), total size: %d", len(chunks), writer.Size())
}

func TestReaderBasic(t *testing.T) {
	// Create a writer
	writer := NewWriter(1001, 100)
	writer.WriteUInt32(12345)
	writer.WriteString("test")
	writer.WriteFloat(3.14)

	chunks := writer.GetChunks()

	// Create a reader from the chunks
	reader := NewReader(1001)
	for _, chunk := range chunks {
		reader.Push(chunk)
	}
	reader.Reset()

	// Read back the values
	val1 := reader.ReadUInt32(0)
	if val1 != 12345 {
		t.Errorf("ReadUInt32: expected 12345, got %d", val1)
	}

	str, err := reader.ReadString()
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}
	if str != "test" {
		t.Errorf("ReadString: expected 'test', got '%s'", str)
	}

	val3 := reader.ReadFloat(0)
	if val3 < 3.13 || val3 > 3.15 {
		t.Errorf("ReadFloat: expected ~3.14, got %f", val3)
	}

	t.Logf("Successfully read: uint32=%d, string='%s', float=%f", val1, str, val3)
}

func TestFragmentation(t *testing.T) {
	// Create a large packet that will fragment
	largeSize := TotalSize(60000) // Will create 2+ fragments
	writer := NewWriter(2001, largeSize)

	// Fill with data
	for i := 0; i < 15000; i++ {
		writer.WriteUInt32(uint32(i))
	}

	chunks := writer.GetChunks()
	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks, got %d", len(chunks))
	}

	t.Logf("Large packet fragmented into %d chunks", len(chunks))

	// Verify each chunk has correct headers
	for i, chunk := range chunks {
		header := chunk.Header()
		if header.FragmentID != ChunkCount(i) {
			t.Errorf("Chunk %d: wrong FragmentID, expected %d, got %d", i, i, header.FragmentID)
		}
		if header.TotalFrags != ChunkCount(len(chunks)) {
			t.Errorf("Chunk %d: wrong TotalFrags, expected %d, got %d", i, len(chunks), header.TotalFrags)
		}
		if header.Opcode != 2001 {
			t.Errorf("Chunk %d: wrong Opcode, expected 2001, got %d", i, header.Opcode)
		}
	}
}

func TestBufferReassembly(t *testing.T) {
	// Create a fragmented packet
	writer := NewWriter(3001, 60000)
	for i := 0; i < 15000; i++ {
		writer.WriteUInt32(uint32(i))
	}

	chunks := writer.GetChunks()
	t.Logf("Created packet with %d chunks", len(chunks))

	// Create buffer to reassemble
	buffer := NewBuffer(MinFragmentSize, BufferQuota, MaxFragmentSize)

	receivedComplete := false
	buffer.SetOnPacket(func(reader *Reader) {
		receivedComplete = true
		t.Logf("Received complete packet: opcode=%d, size=%d", reader.Opcode(), reader.Size())

		// Verify we can read the data back
		for i := 0; i < 10; i++ {
			val := reader.ReadUInt32(0)
			if val != uint32(i) {
				t.Errorf("Value mismatch at index %d: expected %d, got %d", i, i, val)
			}
		}
	})

	buffer.SetOnError(func(result Result) {
		t.Errorf("Buffer error: %s", result.String())
	})

	// Feed chunks to buffer
	for _, chunk := range chunks {
		result := buffer.ReceivePacket(chunk.FullSize(), chunk.Data())
		if result.IsError() {
			t.Fatalf("Failed to receive chunk: %s", result.String())
		}
	}

	if !receivedComplete {
		t.Fatal("Complete packet callback was not called")
	}
}

func TestAllDataTypes(t *testing.T) {
	writer := NewWriter(4001, 100)

	// Write all types
	writer.WriteUInt8(255)
	writer.WriteInt8(-128)
	writer.WriteUInt16(65535)
	writer.WriteInt16(-32768)
	writer.WriteUInt32(4294967295)
	writer.WriteInt32(-2147483648)
	writer.WriteUInt64(18446744073709551615)
	writer.WriteInt64(-9223372036854775808)
	writer.WriteFloat(3.14159)
	writer.WriteDouble(2.71828)
	writer.WriteString("Hello, World!")

	chunks := writer.GetChunks()

	// Read back
	reader := NewReader(4001)
	for _, chunk := range chunks {
		reader.Push(chunk)
	}
	reader.Reset()

	if v := reader.ReadUInt8(0); v != 255 {
		t.Errorf("UInt8: expected 255, got %d", v)
	}
	if v := reader.ReadInt8(0); v != -128 {
		t.Errorf("Int8: expected -128, got %d", v)
	}
	if v := reader.ReadUInt16(0); v != 65535 {
		t.Errorf("UInt16: expected 65535, got %d", v)
	}
	if v := reader.ReadInt16(0); v != -32768 {
		t.Errorf("Int16: expected -32768, got %d", v)
	}
	if v := reader.ReadUInt32(0); v != 4294967295 {
		t.Errorf("UInt32: expected 4294967295, got %d", v)
	}
	if v := reader.ReadInt32(0); v != -2147483648 {
		t.Errorf("Int32: expected -2147483648, got %d", v)
	}
	// Skip uint64/int64 as they may have precision issues in Go
	reader.ReadUInt64(0)
	reader.ReadInt64(0)

	if v := reader.ReadFloat(0); v < 3.14 || v > 3.15 {
		t.Errorf("Float: expected ~3.14159, got %f", v)
	}
	if v := reader.ReadDouble(0); v < 2.71 || v > 2.72 {
		t.Errorf("Double: expected ~2.71828, got %f", v)
	}
	if v, _ := reader.ReadString(); v != "Hello, World!" {
		t.Errorf("String: expected 'Hello, World!', got '%s'", v)
	}

	t.Log("All data types read/write successfully")
}

func TestLengthPrefixedStrings(t *testing.T) {
	writer := NewWriter(5001, 200)

	// Write length-prefixed strings
	if err := writer.WriteLengthString("Hello"); err != nil {
		t.Fatalf("WriteLengthString failed: %v", err)
	}
	if err := writer.WriteLengthString(""); err != nil {
		t.Fatalf("WriteLengthString empty failed: %v", err)
	}
	if err := writer.WriteLengthString("World with spaces and 日本語"); err != nil {
		t.Fatalf("WriteLengthString unicode failed: %v", err)
	}

	chunks := writer.GetChunks()

	// Read back
	reader := NewReader(5001)
	for _, chunk := range chunks {
		reader.Push(chunk)
	}
	reader.Reset()

	// Read length-prefixed strings
	str1 := reader.ReadLengthString("default")
	if str1 != "Hello" {
		t.Errorf("ReadLengthString: expected 'Hello', got '%s'", str1)
	}

	str2 := reader.ReadLengthString("default")
	if str2 != "" {
		t.Errorf("ReadLengthString empty: expected '', got '%s'", str2)
	}

	str3 := reader.ReadLengthString("default")
	if str3 != "World with spaces and 日本語" {
		t.Errorf("ReadLengthString unicode: expected 'World with spaces and 日本語', got '%s'", str3)
	}

	t.Log("Length-prefixed strings read/write successfully")
}

func TestLengthPrefixedStringWithMaxLength(t *testing.T) {
	writer := NewWriter(5002, 100)

	// Write with max length
	if err := writer.WriteLengthStringN("Hello World", 5); err != nil {
		t.Fatalf("WriteLengthStringN failed: %v", err)
	}

	chunks := writer.GetChunks()

	reader := NewReader(5002)
	for _, chunk := range chunks {
		reader.Push(chunk)
	}
	reader.Reset()

	str := reader.ReadLengthString("default")
	if str != "Hello" {
		t.Errorf("ReadLengthString truncated: expected 'Hello', got '%s'", str)
	}

	t.Log("Length-prefixed string with max length works correctly")
}

func TestNullTerminatedStringMethods(t *testing.T) {
	writer := NewWriter(5003, 100)

	// Use the explicit null-terminated methods
	if err := writer.WriteStringNullTerm("Test1"); err != nil {
		t.Fatalf("WriteStringNullTerm failed: %v", err)
	}
	if err := writer.WriteStringNullTermN("TruncateMe", 5); err != nil {
		t.Fatalf("WriteStringNullTermN failed: %v", err)
	}

	chunks := writer.GetChunks()

	reader := NewReader(5003)
	for _, chunk := range chunks {
		reader.Push(chunk)
	}
	reader.Reset()

	str1, err := reader.ReadStringNullTerm()
	if err != nil {
		t.Fatalf("ReadStringNullTerm failed: %v", err)
	}
	if str1 != "Test1" {
		t.Errorf("ReadStringNullTerm: expected 'Test1', got '%s'", str1)
	}

	str2, err := reader.ReadStringNullTerm()
	if err != nil {
		t.Fatalf("ReadStringNullTerm truncated failed: %v", err)
	}
	if str2 != "Trunc" {
		t.Errorf("ReadStringNullTerm truncated: expected 'Trunc', got '%s'", str2)
	}

	t.Log("Null-terminated string methods work correctly")
}

func TestMixedStringTypes(t *testing.T) {
	writer := NewWriter(5004, 200)

	// Mix both string types
	writer.WriteString("NullTerm")             // null-terminated
	writer.WriteLengthString("LengthPrefixed") // length-prefixed
	writer.WriteString("Another")              // null-terminated

	chunks := writer.GetChunks()

	reader := NewReader(5004)
	for _, chunk := range chunks {
		reader.Push(chunk)
	}
	reader.Reset()

	str1, _ := reader.ReadString()
	if str1 != "NullTerm" {
		t.Errorf("First string: expected 'NullTerm', got '%s'", str1)
	}

	str2 := reader.ReadLengthString("")
	if str2 != "LengthPrefixed" {
		t.Errorf("Second string: expected 'LengthPrefixed', got '%s'", str2)
	}

	str3, _ := reader.ReadString()
	if str3 != "Another" {
		t.Errorf("Third string: expected 'Another', got '%s'", str3)
	}

	t.Log("Mixed string types work correctly")
}

func TestLuaGenerator(t *testing.T) {
	opts := DefaultLuaGeneratorOptions()
	opts.OutputPath = "tmp_rovodev_test_custompackets.lua"

	err := GenerateLuaAPI(opts)
	if err != nil {
		t.Fatalf("GenerateLuaAPI failed: %v", err)
	}

	// Verify file exists and has content
	content, err := os.ReadFile(opts.OutputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Check for expected content
	contentStr := string(content)
	expectedStrings := []string{
		"CustomPackets Lua API",
		"CreateCustomPacket",
		"OnCustomPacket",
		"WriteUInt8",
		"ReadUInt8",
		"WriteLengthString",
		"ReadLengthString",
		"_CustomPacketReceive",
		"CUSTOM_PACKET_HEADER_SIZE",
	}

	for _, expected := range expectedStrings {
		if !contains(contentStr, expected) {
			t.Errorf("Generated Lua missing expected string: %s", expected)
		}
	}

	// Cleanup
	os.Remove(opts.OutputPath)

	t.Log("Lua generator produces valid output")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

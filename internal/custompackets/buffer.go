// Copyright (c) 2025 Thorium
// Ported from TSWoW CustomPackets

package custompackets

import "fmt"

// Result represents the result of packet processing
type Result uint32

const (
	NoHeader          Result = 0x1   // 1
	HeaderMismatch    Result = 0x2   // 2
	InvalidFragCount  Result = 0x4   // 4
	InvalidFirstFrag  Result = 0x8   // 8
	InvalidFragID     Result = 0x10  // 16
	TooSmallFragment  Result = 0x20  // 32
	TooBigFragment    Result = 0x40  // 64
	OutOfSpace        Result = 0x80  // 128
	HandledFragment   Result = 0x100 // 256
	HandledMessage    Result = 0x200 // 512
)

// IsSuccess returns true if the result is a success
func (r Result) IsSuccess() bool {
	return r&(HandledFragment|HandledMessage) != 0
}

// IsError returns true if the result is an error
func (r Result) IsError() bool {
	return r&(NoHeader|HeaderMismatch|InvalidFragCount|InvalidFirstFrag|InvalidFragID|TooSmallFragment|TooBigFragment|OutOfSpace) != 0
}

// String returns a string representation of the result
func (r Result) String() string {
	switch r {
	case NoHeader:
		return "NoHeader"
	case HeaderMismatch:
		return "HeaderMismatch"
	case InvalidFragCount:
		return "InvalidFragCount"
	case InvalidFirstFrag:
		return "InvalidFirstFrag"
	case InvalidFragID:
		return "InvalidFragID"
	case TooSmallFragment:
		return "TooSmallFragment"
	case TooBigFragment:
		return "TooBigFragment"
	case OutOfSpace:
		return "OutOfSpace"
	case HandledFragment:
		return "HandledFragment"
	case HandledMessage:
		return "HandledMessage"
	default:
		return fmt.Sprintf("Unknown(%d)", r)
	}
}

// Buffer handles packet fragment reassembly
type Buffer struct {
	quota            TotalSize
	minFragmentSize  ChunkSize
	maxFragmentSize  ChunkSize
	current          *Reader
	onPacket         func(*Reader)
	onError          func(Result)
}

// NewBuffer creates a new packet buffer
func NewBuffer(minFragmentSize ChunkSize, quota TotalSize, maxFragmentSize ChunkSize) *Buffer {
	return &Buffer{
		quota:           quota,
		minFragmentSize: minFragmentSize,
		maxFragmentSize: maxFragmentSize,
		current:         nil,
		onPacket:        func(*Reader) {},
		onError:         func(Result) {},
	}
}

// SetOnPacket sets the callback for when a complete packet is received
func (b *Buffer) SetOnPacket(fn func(*Reader)) {
	b.onPacket = fn
}

// SetOnError sets the callback for when an error occurs
func (b *Buffer) SetOnError(fn func(Result)) {
	b.onError = fn
}

// ReceivePacket processes an incoming packet fragment
func (b *Buffer) ReceivePacket(size ChunkSize, data []byte) Result {
	// Validate minimum size
	if size < HeaderSize() {
		return b.handleError(NoHeader, data)
	}

	// Create chunk from data
	chunk := NewChunkFromData(data)
	header := chunk.Header()

	if header == nil {
		return b.handleError(NoHeader, data)
	}

	// Validate fragment size
	if size < b.minFragmentSize && header.FragmentID != header.TotalFrags-1 {
		return b.handleError(TooSmallFragment, data)
	}

	if size > b.maxFragmentSize {
		return b.handleError(TooBigFragment, data)
	}

	// Validate fragment count
	if header.TotalFrags == 0 {
		return b.handleError(InvalidFragCount, data)
	}

	// If this is the first fragment, create new reader
	if header.FragmentID == 0 {
		// Check quota
		if b.current != nil && TotalSize(b.current.Size()) > b.quota {
			return b.handleError(OutOfSpace, data)
		}

		b.current = NewReader(header.Opcode)
	}

	// Validate we have a current reader
	if b.current == nil {
		return b.handleError(InvalidFirstFrag, data)
	}

	// Validate header matches
	if b.current.Opcode() != header.Opcode {
		return b.handleError(HeaderMismatch, data)
	}

	// Validate fragment ID
	expectedFragID := b.current.ChunkCount()
	if header.FragmentID != expectedFragID {
		return b.handleError(InvalidFragID, data)
	}

	// Append fragment
	b.appendFragment(chunk, header.FragmentID == header.TotalFrags-1)

	// Check if message is complete
	if header.FragmentID == header.TotalFrags-1 {
		return b.handleSuccess()
	}

	return HandledFragment
}

// Size returns the current buffer size
func (b *Buffer) Size() TotalSize {
	if b.current == nil {
		return 0
	}
	return b.current.Size()
}

// appendFragment adds a fragment to the current reader
func (b *Buffer) appendFragment(chunk *Chunk, isLast bool) {
	b.current.Push(chunk)
}

// handleError handles an error
func (b *Buffer) handleError(result Result, data []byte) Result {
	b.current = nil
	b.onError(result)
	return result
}

// handleSuccess handles successful packet completion
func (b *Buffer) handleSuccess() Result {
	if b.current != nil {
		b.current.Reset() // Reset read position for consumer
		b.onPacket(b.current)
	}
	b.current = nil
	return HandledMessage
}

package zanarkand

import (
	"bufio"
	"bytes"
	"io"
	"testing"
	"time"
)

var headerTestBlob = []byte{
	0x52, 0x52, 0xA0, 0x41, 0xFF, 0x5D, 0x46, 0xE2, // magic
	0x7F, 0x2A, 0x64, 0x4D, 0x7B, 0x99, 0xC4, 0x75, // padding
	0x81, 0x48, 0x6E, 0xD6, 0x68, 0x01, 0x00, 0x00, // time
	0x5C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, // length, connection, count
	0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved, compressed, padding
}

var badJujuTestBlob = []byte{
	0x69, 0x69, 0xA0, 0x41, 0xFF, 0x5D, 0x46, 0xE2, // magic
	0x7F, 0x2A, 0x64, 0x4D, 0x7B, 0x99, 0xC4, 0x75, // padding
	0x81, 0x48, 0x6E, 0xD6, 0x68, 0x01, 0x00, 0x00, // time
	0x5C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, // length, connection, count
	0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved, compressed, padding
}

var zlibBodyTestBlob = []byte{
	0x78, 0x9C, 0x33, 0x60, 0x60, 0x60, 0x28, 0x8B,
	0x50, 0x13, 0x58, 0x33, 0x45, 0x51, 0x80, 0x19,
	0xC8, 0x16, 0x61, 0x70, 0x65, 0x64, 0x60, 0x60,
	0x65, 0xD8, 0x74, 0x2B, 0x3E, 0x06, 0xC8, 0x65,
	0x88, 0xD9, 0xC8, 0xC0, 0xC0, 0x61, 0xF2, 0x82,
	0xD9, 0x5F, 0xD4, 0x81, 0x11, 0xC4, 0x07, 0x00,
	0xCD, 0xC1, 0x08, 0x28,
}

var zlibFrameTestBlob = append(headerTestBlob[:], zlibBodyTestBlob[:]...)

func TestDecode(t *testing.T) {
	frame := new(Frame)
	frame.Decode(zlibFrameTestBlob)

	if frame.Length != 92 {
		t.Errorf("Expected frame length 92, got %v", frame.Length)
	}

	if frame.Connection > 0 {
		t.Errorf("Expected connection ID 0, got %v", frame.Connection)
	}

	if frame.Count != 1 {
		t.Errorf("Expected 1 message in this frame, got %v", frame.Count)
	}

	if !frame.Compressed {
		t.Error("Expected compressed frame, got uncompressed")
	}

	if frame.Timestamp != time.Unix(int64(1549785778), int64(305000000)) {
		t.Errorf("Expected frame timestamp to be 2019-02-10 08:02:58.305 GMT, got %v", frame.Timestamp.UnixNano())
	}

	if len(frame.Body) != int(frame.Length)-frameHeaderLength {
		t.Errorf("Expected frame payload to be 52 bytes, got %d", len(frame.Body))
	}
}

func TestDiscard(t *testing.T) {
	reader := bufio.NewReader(bytes.NewReader(headerTestBlob))
	err := discardUntilValid(reader)
	if err != nil {
		t.Error("Expected no errors with discarding")
	}

	reader = bufio.NewReader(bytes.NewReader(badJujuTestBlob))
	err = discardUntilValid(reader)
	if err != io.EOF {
		t.Error("Unexpected error with discarding")
	}
}

func TestValidatePredicate(t *testing.T) {
	valid := validateMagic(headerTestBlob)

	if !valid {
		t.Error("Expected valid predicate magic bytes")
	}

	invalid := validateMagic(badJujuTestBlob)
	if invalid {
		t.Error("Expected invalid predicate magic bytes to fail validation")
	}
}

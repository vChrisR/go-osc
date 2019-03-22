package osc

import (
	"bufio"
	"bytes"
	"testing"
)

func TestMessage_Append(t *testing.T) {
	oscAddress := "/address"
	message := NewMessage(oscAddress)
	if message.Address != oscAddress {
		t.Errorf("OSC address should be \"%s\" and is \"%s\"", oscAddress, message.Address)
	}

	message.Append("string argument")
	message.Append(123456789)
	message.Append(true)

	if len(message.Arguments) != 3 {
		t.Errorf("Number of arguments should be %d and is %d", 3, len(message.Arguments))
	}
}

func TestMessage_String(t *testing.T) {
	for _, tt := range []struct {
		desc string
		msg  *Message
		str  string
	}{
		{"nil", nil, ""},
		{"addr_only", NewMessage("/foo/bar"), "/foo/bar ,"},
		{"one_addr", NewMessage("/foo/bar", "123"), "/foo/bar ,s 123"},
		{"two_args", NewMessage("/foo/bar", "123", int32(456)), "/foo/bar ,si 123 456"},
	} {
		if got, want := tt.msg.String(), tt.str; got != want {
			t.Errorf("%s: String() = '%s', want = '%s'", tt.desc, got, want)
		}
	}
}

func TestReadPaddedString(t *testing.T) {
	for _, tt := range []struct {
		buf []byte // buffer
		n   int    // bytes needed
		s   string // resulting string
	}{
		{[]byte{'t', 'e', 's', 't', 's', 't', 'r', 'i', 'n', 'g', 0, 0}, 12, "teststring"},
		{[]byte{'t', 'e', 's', 't', 0, 0, 0, 0}, 8, "test"},
	} {
		buf := bytes.NewBuffer(tt.buf)
		s, n, err := readPaddedString(bufio.NewReader(buf))
		if err != nil {
			t.Errorf("%s: Error reading padded string: %s", s, err)
		}
		if got, want := n, tt.n; got != want {
			t.Errorf("%s: Bytes needed don't match; got = %d, want = %d", tt.s, got, want)
		}
		if got, want := s, tt.s; got != want {
			t.Errorf("%s: Strings don't match; got = %s, want = %s", tt.s, got, want)
		}
	}
}

func TestWritePaddedString(t *testing.T) {
	buf := []byte{}
	bytesBuffer := bytes.NewBuffer(buf)
	testString := "testString"
	expectedNumberOfWrittenBytes := len(testString) + padBytesNeeded(len(testString))

	n, err := writePaddedString(testString, bytesBuffer)
	if err != nil {
		t.Errorf(err.Error())
	}

	if n != expectedNumberOfWrittenBytes {
		t.Errorf("Expected number of written bytes should be \"%d\" and is \"%d\"", expectedNumberOfWrittenBytes, n)
	}
}

func TestPadBytesNeeded(t *testing.T) {
	var n int
	n = padBytesNeeded(4)
	if n != 4 {
		t.Errorf("Number of pad bytes should be 4 and is: %d", n)
	}

	n = padBytesNeeded(3)
	if n != 1 {
		t.Errorf("Number of pad bytes should be 1 and is: %d", n)
	}

	n = padBytesNeeded(1)
	if n != 3 {
		t.Errorf("Number of pad bytes should be 3 and is: %d", n)
	}

	n = padBytesNeeded(0)
	if n != 4 {
		t.Errorf("Number of pad bytes should be 4 and is: %d", n)
	}

	n = padBytesNeeded(32)
	if n != 4 {
		t.Errorf("Number of pad bytes should be 4 and is: %d", n)
	}

	n = padBytesNeeded(63)
	if n != 1 {
		t.Errorf("Number of pad bytes should be 1 and is: %d", n)
	}

	n = padBytesNeeded(10)
	if n != 2 {
		t.Errorf("Number of pad bytes should be 2 and is: %d", n)
	}
}

var parsedMessage *Message

func BenchmarkReadMessage(b *testing.B) {
	message := NewMessage("/text")
	message.Append("thisisastring")
	message.Append(true)
	message.Append(int32(42))
	data, _ := message.MarshalBinary()
	for n := 0; n < b.N; n++ {
		var start int = 0
		parsedMessage, _ = readMessage(bufio.NewReader(bytes.NewBuffer(data)), &start)
	}
}

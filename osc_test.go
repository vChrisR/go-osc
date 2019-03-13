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

/*
func TestMessage_TypeTags(t *testing.T) {
	for _, tt := range []struct {
		desc string
		msg  *Message
		tags string
		ok   bool
	}{
		{"addr_only", NewMessage("/"), ",", true},
		{"nil", NewMessage("/", nil), ",N", true},
		{"bool_true", NewMessage("/", true), ",T", true},
		{"bool_false", NewMessage("/", false), ",F", true},
		{"int32", NewMessage("/", int32(1)), ",i", true},
		{"int64", NewMessage("/", int64(2)), ",h", true},
		{"float32", NewMessage("/", float32(3.0)), ",f", true},
		{"float64", NewMessage("/", float64(4.0)), ",d", true},
		{"string", NewMessage("/", "5"), ",s", true},
		{"[]byte", NewMessage("/", []byte{'6'}), ",b", true},
		{"two_args", NewMessage("/", "123", int32(456)), ",si", true},
		{"invalid_msg", nil, "", false},
		{"invalid_arg", NewMessage("/foo/bar", 789), "", false},
	} {
		tags, err := tt.msg.TypeTags()
		if err != nil && tt.ok {
			t.Errorf("%s: TypeTags() unexpected error: %s", tt.desc, err)
			continue
		}
		if err == nil && !tt.ok {
			t.Errorf("%s: TypeTags() expected an error", tt.desc)
			continue
		}
		if !tt.ok {
			continue
		}
		if got, want := tags, tt.tags; got != want {
			t.Errorf("%s: TypeTags() = '%s', want = '%s'", tt.desc, got, want)
		}
	}
}
*/
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

func TestClientSetLocalAddr(t *testing.T) {
	client := NewClient("localhost", 8967)
	err := client.SetLocalAddr("localhost", 41789)
	if err != nil {
		t.Error(err.Error())
	}
	expectedAddr := "127.0.0.1:41789"
	if client.laddr.String() != expectedAddr {
		t.Errorf("Expected laddr to be %s but was %s", expectedAddr, client.laddr.String())
	}
}

const zero = string(byte(0))

// nulls returns a string of `i` nulls.
func nulls(i int) string {
	s := ""
	for j := 0; j < i; j++ {
		s += zero
	}
	return s
}

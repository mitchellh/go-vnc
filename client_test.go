package vnc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func newMockServer(t *testing.T, version string) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("error listening: %s", err)
	}

	go func() {
		defer ln.Close()
		c, err := ln.Accept()
		if err != nil {
			t.Fatalf("error accepting conn: %s", err)
		}
		defer c.Close()

		_, err = c.Write([]byte(fmt.Sprintf("RFB %s\n", version)))
		if err != nil {
			t.Fatal("failed writing version")
		}
	}()

	return ln.Addr().String()
}

func TestClient_LowMajorVersion(t *testing.T) {
	nc, err := net.Dial("tcp", newMockServer(t, "002.009"))
	if err != nil {
		t.Fatalf("error connecting to mock server: %s", err)
	}

	_, err = Client(nc, &ClientConfig{})
	if err == nil {
		t.Fatal("error expected")
	}

	if err.Error() != "unsupported server ProtocolVersion 'RFB 002.009\n'" {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestClient_LowMinorVersion(t *testing.T) {
	nc, err := net.Dial("tcp", newMockServer(t, "003.007"))
	if err != nil {
		t.Fatalf("error connecting to mock server: %s", err)
	}

	_, err = Client(nc, &ClientConfig{})
	if err == nil {
		t.Fatal("error expected")
	}

	if err.Error() != "unsupported server ProtocolVersion 'RFB 003.007\n'" {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestParseProtocolVersion(t *testing.T) {
	tests := []struct {
		proto        []byte
		major, minor uint
		isErr        bool
	}{
		// Valid ProtocolVersion messages.
		{[]byte{82, 70, 66, 32, 48, 48, 51, 46, 48, 48, 56, 10}, 3, 8, false},   // RFB 003.008\n
		{[]byte{82, 70, 66, 32, 48, 48, 51, 46, 56, 56, 57, 10}, 3, 889, false}, // RFB 003.889\n -- OS X 10.10.3
		{[]byte{82, 70, 66, 32, 48, 48, 48, 46, 48, 48, 48, 10}, 0, 0, false},   // RFB 000.0000\n
		// Invalid messages.
		{[]byte{82, 70, 66, 32, 51, 46, 56, 10}, 0, 0, true}, // RFB 3.8\n -- too short; not zero padded
		{[]byte{82, 70, 66, 10}, 0, 0, true},                 // RFB\n -- too short
		{[]byte{}, 0, 0, true},                               // (empty) -- too short
	}

	for _, tt := range tests {
		major, minor, err := parseProtocolVersion(tt.proto)
		if err != nil && !tt.isErr {
			t.Fatalf("parseProtocolVersion(%v) unexpected error %v", tt.proto, err)
		}
		if err == nil && tt.isErr {
			t.Fatalf("parseProtocolVersion(%v) expected error", tt.proto)
		}
		if major != tt.major {
			t.Errorf("parseProtocolVersion(%v) major = %v, want %v", tt.proto, major, tt.major)
		}
		if major != tt.major {
			t.Errorf("parseProtocolVersion(%v) minor = %v, want %v", tt.proto, minor, tt.minor)
		}
	}
}

func TestSecurityResultHandshake(t *testing.T) {
	tests := []struct {
		result uint32
		ok     bool
		reason string
	}{
		{0, true, ""},
		{1, false, "SecurityResult error"},
	}

	mockConn := &MockConn{}
	conn := &ClientConn{
		c:      mockConn,
		config: &ClientConfig{},
	}

	for _, tt := range tests {
		mockConn.Reset()
		if err := binary.Write(conn.c, binary.BigEndian, tt.result); err != nil {
			t.Fatal(err)
		}
		if err := binary.Write(conn.c, binary.BigEndian, uint32(len(tt.reason))); err != nil {
			t.Fatal(err)
		}
		if err := binary.Write(conn.c, binary.BigEndian, []byte(tt.reason)); err != nil {
			t.Fatal(err)
		}

		err := conn.securityResultHandshake()
		if err == nil && !tt.ok {
			t.Fatalf("securityResultHandshake() expected error for result %v", tt.result)
		}
		if err != nil {
			if verr, ok := err.(*VNCError); !ok {
				t.Errorf("securityResultHandshake() unexpected %v error: %v", reflect.TypeOf(err), verr)
			}
		}
	}
}

func TestClientInit(t *testing.T) {
	tests := []struct {
		exclusive bool
		shared    uint8
	}{
		{true, 0},
		{false, 1},
	}

	mockConn := &MockConn{}
	conn := &ClientConn{
		c:      mockConn,
		config: &ClientConfig{},
	}

	for _, tt := range tests {
		mockConn.Reset()
		conn.config.Exclusive = tt.exclusive

		shared, err := conn.clientInit()
		if err != nil {
			t.Fatalf("clientInit() error %v", err)
		}
		if shared != tt.shared {
			t.Errorf("clientInit() got = %v, want %v", shared, tt.shared)
		}
	}
}

// MockConn implements the net.Conn interface.
type MockConn struct {
	b bytes.Buffer
}

func (m *MockConn) Read(b []byte) (int, error) {
	return m.b.Read(b)
}
func (m *MockConn) Write(b []byte) (int, error) {
	return m.b.Write(b)
}
func (m *MockConn) Close() error                       { return nil }
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }
// Implement additional buffer.Buffer functions.
func (m *MockConn) Reset() {
	m.b.Reset()
}

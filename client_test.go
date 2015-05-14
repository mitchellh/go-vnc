package vnc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
	if err != nil {
		if verr, ok := err.(*VNCError); !ok {
			t.Errorf("Client() unexpected %v error: %v", reflect.TypeOf(err), verr)
		}
	}
}

func TestClient_LowMinorVersion(t *testing.T) {
	nc, err := net.Dial("tcp", newMockServer(t, "003.002"))
	if err != nil {
		t.Fatalf("error connecting to mock server: %s", err)
	}

	_, err = Client(nc, &ClientConfig{})
	if err == nil {
		t.Fatal("error expected")
	}
	if err != nil {
		if verr, ok := err.(*VNCError); !ok {
			t.Errorf("Client() unexpected %v error: %v", reflect.TypeOf(err), verr)
		}
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
		// TODO(kward): validate VNCError thrown.
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

func TestProtocolVersionHandshake(t *testing.T) {
	tests := []struct {
		server string
		client string
		ok     bool
	}{
		// Supported versions.
		{"RFB 003.003\n", "RFB 003.003\n", true},
		{"RFB 003.006\n", "RFB 003.003\n", true},
		{"RFB 003.008\n", "RFB 003.008\n", true},
		{"RFB 003.389\n", "RFB 003.008\n", true},
		// Unsupported versions.
		{server: "RFB 002.009\n", ok: false},
	}

	mockConn := &MockConn{}
	conn := &ClientConn{
		c:      mockConn,
		config: &ClientConfig{},
	}

	for _, tt := range tests {
		mockConn.Reset()
		if err := binary.Write(conn.c, binary.BigEndian, []byte(tt.server)); err != nil {
			t.Fatal(err)
		}

		// Validate server message handling.
		err := conn.protocolVersionHandshake()
		if err == nil && !tt.ok {
			t.Fatalf("protocolVersionHandshake() expected error for server protocol version %v", tt.server)
		}
		if err != nil {
			if verr, ok := err.(*VNCError); !ok {
				t.Errorf("protocolVersionHandshake() unexpected %v error: %v", reflect.TypeOf(err), verr)
			}
		}

		// Validate client response.
		var client [pvLen]byte
		err = binary.Read(conn.c, binary.BigEndian, &client)
		if err == nil && !tt.ok {
			t.Fatalf("protocolVersionHandshake() unexpected error: %v", err)
		}
		if string(client[:]) != tt.client && tt.ok {
			t.Errorf("protocolVersionHandshake() client version: got = %v, want = %v", string(client[:]), tt.client)
		}
	}
}

func writeVNCAuthChallenge(w io.Writer) error {
	var c [vncAuthChallengeSize]uint8
	for i := 0; i < vncAuthChallengeSize; i++ {
		c[i] = uint8(i)
	}
	if err := binary.Write(w, binary.BigEndian, c); err != nil {
		return err
	}
	return nil
}

func readVNCAuthChallenge(r io.Reader) error {
	var c [vncAuthChallengeSize]uint8
	if err := binary.Read(r, binary.BigEndian, &c); err != nil {
		return fmt.Errorf("error reading back VNCAuth challenge")
	}
	return nil
}

func TestSecurityHandshake33(t *testing.T) {
	tests := []struct {
		server uint32
		ok     bool
		reason string
	}{
		//-- Supported security types. --
		// Server supports None.
		{secTypeNone, true, ""},
		// Server supports VNCAuth.
		{secTypeVNCAuth, true, ""},
		//-- Unsupported security types. --
		{secTypeInvalid, false, "some reason"},
		{255, false, ""},
	}

	mockConn := &MockConn{}
	conn := &ClientConn{
		c:               mockConn,
		config:          NewClientConfig("."),
		protocolVersion: PROTO_VERS_3_3,
	}

	for _, tt := range tests {
		mockConn.Reset()
		if err := binary.Write(conn.c, binary.BigEndian, tt.server); err != nil {
			t.Fatal(err)
		}
		if len(tt.reason) > 0 {
			if err := binary.Write(conn.c, binary.BigEndian, uint32(len(tt.reason))); err != nil {
				t.Fatal(err)
			}
			if err := binary.Write(conn.c, binary.BigEndian, []byte(tt.reason)); err != nil {
				t.Fatal(err)
			}
		}
		if tt.server == secTypeVNCAuth {
			if err := writeVNCAuthChallenge(conn.c); err != nil {
				t.Fatal(err)
			}
		}

		// Validate server message handling.
		err := conn.securityHandshake()
		if err == nil && !tt.ok {
			t.Fatalf("securityHandshake() expected error for server auth %v", tt.server)
		}
		if err != nil {
			if verr, ok := err.(*VNCError); !ok {
				t.Errorf("securityHandshake() unexpected %v error: %v", reflect.TypeOf(err), verr)
			}
		}
		if !tt.ok {
			continue
		}

		// Validate client response.
		if tt.server == secTypeVNCAuth {
			if err := readVNCAuthChallenge(conn.c); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSecurityHandshake38(t *testing.T) {
	tests := []struct {
		server  []uint8
		client  []ClientAuth
		secType uint8
		ok      bool
		reason  string
	}{
		//-- Supported security types. --
		// Server and client support None.
		{[]uint8{secTypeNone}, []ClientAuth{&ClientAuthNone{}}, secTypeNone, true, ""},
		// Server and client support VNCAuth.
		{[]uint8{secTypeVNCAuth}, []ClientAuth{&ClientAuthVNC{"."}}, secTypeVNCAuth, true, ""},
		// Server and client both support VNCAuth and None.
		{[]uint8{secTypeVNCAuth, secTypeNone}, []ClientAuth{&ClientAuthVNC{"."}, &ClientAuthNone{}}, secTypeVNCAuth, true, ""},
		// Server supports unknown #255, VNCAuth and None.
		{[]uint8{255, secTypeVNCAuth, secTypeNone}, []ClientAuth{&ClientAuthVNC{"."}, &ClientAuthNone{}}, secTypeVNCAuth, true, ""},
		//-- Unsupported security types. --
		// Server provided no valid security types.
		{[]uint8{secTypeInvalid}, []ClientAuth{}, secTypeInvalid, false, "some reason"},
		// Client and server don't support same security types.
		{[]uint8{secTypeVNCAuth}, []ClientAuth{&ClientAuthNone{}}, secTypeInvalid, false, ""},
		// Server supports only unknown #255.
		{[]uint8{255}, []ClientAuth{&ClientAuthNone{}}, secTypeInvalid, false, ""},
	}

	mockConn := &MockConn{}
	conn := &ClientConn{
		c:               mockConn,
		config:          &ClientConfig{},
		protocolVersion: PROTO_VERS_3_8,
	}

	for _, tt := range tests {
		mockConn.Reset()
		if err := binary.Write(conn.c, binary.BigEndian, uint8(len(tt.server))); err != nil {
			t.Fatal(err)
		}
		if err := binary.Write(conn.c, binary.BigEndian, []byte(tt.server)); err != nil {
			t.Fatal(err)
		}
		if len(tt.reason) > 0 {
			if err := binary.Write(conn.c, binary.BigEndian, uint32(len(tt.reason))); err != nil {
				t.Fatal(err)
			}
			if err := binary.Write(conn.c, binary.BigEndian, []byte(tt.reason)); err != nil {
				t.Fatal(err)
			}
		}
		if tt.secType == secTypeVNCAuth {
			if err := writeVNCAuthChallenge(conn.c); err != nil {
				t.Fatal(err)
			}
		}
		conn.config.Auth = tt.client

		// Validate server message handling.
		err := conn.securityHandshake()
		if err == nil && !tt.ok {
			t.Fatalf("securityHandshake() expected error for server auth %v", tt.server)
		}
		if err != nil {
			if verr, ok := err.(*VNCError); !ok {
				t.Errorf("securityHandshake() unexpected %v error: %v", reflect.TypeOf(err), verr)
			}
		}
		if !tt.ok {
			continue
		}

		// Validate client response.
		var secType uint8
		err = binary.Read(conn.c, binary.BigEndian, &secType)
		if secType != tt.secType {
			t.Errorf("securityHandshake() secType: got = %v, want = %v", secType, tt.secType)
		}
		if tt.secType == secTypeVNCAuth {
			if err := readVNCAuthChallenge(conn.c); err != nil {
				t.Fatal(err)
			}
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

		// Validate server message handling.
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

		// Validate client response.
		err := conn.clientInit()
		if err != nil {
			t.Fatalf("clientInit() unexpected error %v", err)
		}
		var shared uint8
		err = binary.Read(conn.c, binary.BigEndian, &shared)
		if shared != tt.shared {
			t.Errorf("clientInit() shared: got = %v, want = %v", shared, tt.shared)
		}
	}
}

func TestServerInit(t *testing.T) {
	const (
		none = iota
		fbw
		fbh
		pf
		dn
	)
	tests := []struct {
		eof               int
		fbWidth, fbHeight uint16
		pixelFormat       [16]byte // TODO(kward): replace with PixelFormat
		desktopName       string
	}{
		// Valid protocol.
		{dn, 100, 200, [16]byte{}, "foo"},
		// Invalid protocol (missing fields).
		{eof: none},
		{eof: fbw, fbWidth: 1},
		{eof: fbh, fbWidth: 2, fbHeight: 1},
		{eof: pf, fbWidth: 3, fbHeight: 2, pixelFormat: [16]byte{}},
	}

	mockConn := &MockConn{}
	conn := &ClientConn{
		c:      mockConn,
		config: &ClientConfig{},
	}

	for _, tt := range tests {
		mockConn.Reset()
		if tt.eof >= fbw {
			if err := binary.Write(conn.c, binary.BigEndian, tt.fbWidth); err != nil {
				t.Fatal(err)
			}
		}
		if tt.eof >= fbh {
			if err := binary.Write(conn.c, binary.BigEndian, tt.fbHeight); err != nil {
				t.Fatal(err)
			}
		}
		if tt.eof >= pf {
			if err := binary.Write(conn.c, binary.BigEndian, tt.pixelFormat); err != nil {
				t.Fatal(err)
			}
		}
		if tt.eof >= dn {
			if err := binary.Write(conn.c, binary.BigEndian, uint32(len(tt.desktopName))); err != nil {
				t.Fatal(err)
			}
			if err := binary.Write(conn.c, binary.BigEndian, []byte(tt.desktopName)); err != nil {
				t.Fatal(err)
			}
		}

		// Validate server message handling.
		err := conn.serverInit()
		if tt.eof < dn && err == nil {
			t.Fatalf("serverInit() expected error")
		}
		if tt.eof < dn {
			// The protocol was incomplete; no point in checking values.
			continue
		}
		if err != nil {
			t.Fatalf("serverInit() error %v", err)
		}
		if conn.FrameBufferWidth != tt.fbWidth {
			t.Errorf("serverInit() FrameBufferWidth: got = %v, want = %v", conn.FrameBufferWidth, tt.fbWidth)
		}
		if conn.FrameBufferHeight != tt.fbHeight {
			t.Errorf("serverInit() FrameBufferHeight: got = %v, want = %v", conn.FrameBufferHeight, tt.fbHeight)
		}
		// TODO(kward): add test for PixelFormat.
		if conn.DesktopName != tt.desktopName {
			t.Errorf("serverInit() DesktopName: got = %v, want = %v", conn.DesktopName, tt.desktopName)
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

// Package vnc implements a VNC client.
//
// References:
//   [PROTOCOL]: http://tools.ietf.org/html/rfc6143
package vnc

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type ClientConn struct {
	c net.Conn
	config *ClientConfig

	// Width of the frame buffer in pixels, sent from the server.
	FrameBufferWidth uint16

	// Height of the frame buffer in pixels, sent from the server.
	FrameBufferHeight uint16

	// Name associated with the desktop, sent from the server.
	DesktopName string
}

// A ClientConfig structure is used to configure a ClientConn. After
// one has been passed to initialize a connection, it must not be modified.
type ClientConfig struct {
	// A slice of ClientAuth methods. Only the first instance that is
	// suitable by the server will be used to authenticate.
	Auth []ClientAuth

	// Exclusive determines whether the connection is shared with other
	// clients. If true, then all other clients connected will be
	// disconnected when a connection is established to the VNC server.
	Exclusive bool
}

func Client(c net.Conn, cfg *ClientConfig) (*ClientConn, error) {
	conn := &ClientConn{
		c: c,
		config: cfg,
	}

	if err := conn.handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func (c *ClientConn) Close() error {
	return c.c.Close()
}

func (c *ClientConn) handshake() error {
	var protocolVersion [12]byte

	// 7.1.1, read the ProtocolVersion message sent by the server.
	if _, err := io.ReadFull(c.c, protocolVersion[:]); err != nil {
		return err
	}

	var maxMajor, maxMinor uint8
	_, err := fmt.Sscanf(string(protocolVersion[:]), "RFB %d.%d\n", &maxMajor, &maxMinor)
	if err != nil {
		return err
	}

	if maxMajor < 3 {
		return fmt.Errorf("unsupported major version, less than 3: %d", maxMajor)
	}

	if maxMinor < 8 {
		return fmt.Errorf("unsupported minor version, less than 8: %d", maxMinor)
	}

	// Respond with the version we will support
	if _, err = c.c.Write([]byte("RFB 003.008\n")); err != nil {
		return err
	}

	// 7.1.2 Security Handshake from server
	var numSecurityTypes uint8
	if err = binary.Read(c.c, binary.BigEndian, &numSecurityTypes); err != nil {
		return err
	}

	if numSecurityTypes == 0 {
		return fmt.Errorf("no security types: %s", c.readErrorReason())
	}

	securityTypes := make([]uint8, numSecurityTypes)
	if err = binary.Read(c.c, binary.BigEndian, &securityTypes); err != nil {
		return err
	}

	var auth ClientAuth
FindAuth:
	for _, curAuth := range c.config.Auth {
		for _, securityType := range securityTypes {
			if curAuth.SecurityType() == securityType {
				// We use the first matching supported authentication
				auth = curAuth
				break FindAuth
			}
		}
	}

	if auth == nil {
		return fmt.Errorf("no suitable auth schemes found. server supported: %#v", securityTypes)
	}

	// Respond back with the security type we'll use
	if err = binary.Write(c.c, binary.BigEndian, auth.SecurityType()); err != nil {
		return err
	}

	if err = auth.Handshake(c.c); err != nil {
		return err
	}

	// 7.1.3 SecurityResult Handshake
	var securityResult uint32
	if err = binary.Read(c.c, binary.BigEndian, &securityResult); err != nil {
		return err
	}

	if securityResult == 1 {
		return fmt.Errorf("security handshake failed: %s", c.readErrorReason())
	}

	// 7.3.1 ClientInit
	var sharedFlag uint8 = 1
	if c.config.Exclusive {
		sharedFlag = 0
	}

	if err = binary.Write(c.c, binary.BigEndian, sharedFlag); err != nil {
		return err
	}

	// 7.3.2 ServerInit
	if err = binary.Read(c.c, binary.BigEndian, &c.FrameBufferWidth); err != nil {
		return err
	}

	if err = binary.Read(c.c, binary.BigEndian, &c.FrameBufferHeight); err != nil {
		return err
	}

	// TODO(mitchellh): Store or use this information somehow
	var pixelFormat [16]uint8
	if err = binary.Read(c.c, binary.BigEndian, pixelFormat[:]); err != nil {
		return err
	}

	var nameLength uint32
	if err = binary.Read(c.c, binary.BigEndian, &nameLength); err != nil {
		return err
	}

	nameBytes := make([]uint8, nameLength)
	if err = binary.Read(c.c, binary.BigEndian, &nameBytes); err != nil {
		return err
	}

	c.DesktopName = string(nameBytes)

	return nil
}

func (c *ClientConn) readErrorReason() string {
	var reasonLen uint32
	if err := binary.Read(c.c, binary.BigEndian, &reasonLen); err != nil {
		return "<error>"
	}

	reason := make([]uint8, reasonLen)
	if err := binary.Read(c.c, binary.BigEndian, &reason); err != nil {
		return "<error>"
	}

	return string(reason)
}

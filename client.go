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
}

// A ClientConfig structure is used to configure a ClientConn. After
// one has been passed to initialize a connection, it must not be modified.
type ClientConfig struct {
	// A slice of ClientAuth methods. Only the first instance that is
	// suitable by the server will be used to authenticate.
	Auth []ClientAuth
}

func Client(c net.Conn, cfg *ClientConfig) (*ClientConn, error) {
	conn := &ClientConn{
		c,
		cfg,
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

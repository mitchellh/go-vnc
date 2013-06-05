// Package vnc implements a VNC client.
//
// References:
//   [PROTOCOL]: http://tools.ietf.org/html/rfc6143
package vnc

import (
	"fmt"
	"io"
	"net"
)

type ClientConn struct {
	net.Conn
}

func Client(c net.Conn) (*ClientConn, error) {
	conn := &ClientConn{
		c,
	}

	if err := conn.handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func (c *ClientConn) handshake() error {
	var protocolVersion [12]byte

	// 7.1.1, read the ProtocolVersion message sent by the server.
	if _, err := io.ReadFull(c, protocolVersion[:]); err != nil {
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

	return nil
}

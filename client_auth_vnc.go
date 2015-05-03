/*
ClientAuthVNC implements the ClientAuth interface to provide support for
VNC Authentication.

See http://tools.ietf.org/html/rfc6143#section-7.2.2 for more info.
*/
package vnc

import (
	"crypto/des"
	"encoding/binary"
	"net"
)

// ClientAuthVNC is the standard password authentication
type ClientAuthVNC struct {
	Password string
}

func (*ClientAuthVNC) SecurityType() uint8 {
	return 2
}

// 7.2.2. VNC Authentication uses a 16-byte challenge.
const challengeSize = 16

func (auth *ClientAuthVNC) Handshake(conn net.Conn) error {
	// Read challenge block
	var challenge [challengeSize]byte
	if err := binary.Read(conn, binary.BigEndian, &challenge); err != nil {
		return err
	}

	auth.encode(&challenge)

	// Send the encrypted challenge back to server
	if err := binary.Write(conn, binary.BigEndian, challenge); err != nil {
		return err
	}

	return nil
}

func (auth *ClientAuthVNC) encode(c *[challengeSize]byte) error {
	// Copy password string to 8 byte 0-padded slice
	key := make([]byte, 8)
	copy(key, auth.Password)

	// Each byte of the password needs to be reversed. This is a
	// non RFC-documented behaviour of VNC clients and servers
	for i := range key {
		key[i] = (key[i]&0x55)<<1 | (key[i]&0xAA)>>1 // Swap adjacent bits
		key[i] = (key[i]&0x33)<<2 | (key[i]&0xCC)>>2 // Swap adjacent pairs
		key[i] = (key[i]&0x0F)<<4 | (key[i]&0xF0)>>4 // Swap the 2 halves
	}

	// Encrypt challenge with key.
	cipher, err := des.NewCipher(key)
	if err != nil {
		return err
	}
	for i := 0; i < challengeSize; i += cipher.BlockSize() {
		cipher.Encrypt(c[i:i+cipher.BlockSize()], c[i:i+cipher.BlockSize()])
	}

	return nil
}

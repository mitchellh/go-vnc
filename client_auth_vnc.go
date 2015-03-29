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

func (auth *ClientAuthVNC) Handshake(conn net.Conn) error {
	// Read challenge block
	var challenge [16]byte
	if err := binary.Read(conn, binary.BigEndian, &challenge); err != nil {
		return err
	}

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

	cipher, err := des.NewCipher(key)
	if err != nil {
		return err
	}

	// Encrypt the challenge low 8 bytes then high 8 bytes
	challengeLow := challenge[0:8]
	challengeHigh := challenge[8:16]
	cipher.Encrypt(challengeLow, challengeLow)
	cipher.Encrypt(challengeHigh, challengeHigh)

	// Send the encrypted challenge back to server
	err = binary.Write(conn, binary.BigEndian, challenge)
	if err != nil {
		return err
	}

	return nil
}

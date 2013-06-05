package vnc

import (
	"bytes"
	"encoding/binary"
	"io"
)

// PixelFormat describes the way a pixel is formatted for a VNC connection.
//
// See RFC 6143 Section 7.4 for information on each of the fields.
type PixelFormat struct {
	BPP        uint8
	Depth      uint8
	BigEndian  bool
	TrueColor  bool
	RedMax     uint16
	GreenMax   uint16
	BlueMax    uint16
	RedShift   uint8
	GreenShift uint8
	BlueShift  uint8
}

func readPixelFormat(r io.Reader, result *PixelFormat) error {
	var rawPixelFormat [16]byte
	if _, err := io.ReadFull(r, rawPixelFormat[:]); err != nil {
		return err
	}

	var pfBoolByte uint8
	brPF := bytes.NewReader(rawPixelFormat[:])
	if err := binary.Read(brPF, binary.BigEndian, &result.BPP); err != nil {
		return err
	}

	if err := binary.Read(brPF, binary.BigEndian, &result.Depth); err != nil {
		return err
	}

	if err := binary.Read(brPF, binary.BigEndian, &pfBoolByte); err != nil {
		return err
	}

	if pfBoolByte != 0 {
		// Big endian is true
		result.BigEndian = true
	}

	if err := binary.Read(brPF, binary.BigEndian, &pfBoolByte); err != nil {
		return err
	}

	if pfBoolByte != 0 {
		// True Color is true. So we also have to read all the color max & shifts.
		result.TrueColor = true

		if err := binary.Read(brPF, binary.BigEndian, &result.RedMax); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.GreenMax); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.BlueMax); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.RedShift); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.GreenShift); err != nil {
			return err
		}

		if err := binary.Read(brPF, binary.BigEndian, &result.BlueShift); err != nil {
			return err
		}
	}

	return nil
}

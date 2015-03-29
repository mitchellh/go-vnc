package vnc

import (
	"encoding/binary"
	"fmt"
	"io"
	"image"
	"image/color"
)

// A ServerMessage implements a message sent from the server to the client.
type ServerMessage interface {
	// The type of the message that is sent down on the wire.
	Type() uint8

	// Read reads the contents of the message from the reader. At the point
	// this is called, the message type has already been read from the reader.
	// This should return a new ServerMessage that is the appropriate type.
	Read(*ClientConn, io.Reader) (ServerMessage, error)
}

// FramebufferUpdateMessage consists of a sequence of images that the
// client should combine into its framebuffer.
type FramebufferUpdateMessage struct {
	Rectangles []image.Image
}

// Rectangle represents a rectangle of pixel data.
//
// See RFC 6143 Section 7.6.1
type rectangleHeader struct {
	X      uint16
	Y      uint16
	Width  uint16
	Height uint16
	Enc    int32
}

func (*FramebufferUpdateMessage) Type() uint8 {
	return 0
}

func (*FramebufferUpdateMessage) Read(c *ClientConn, r io.Reader) (ServerMessage, error) {
	// Read off the padding
	var padding [1]byte
	if _, err := io.ReadFull(r, padding[:]); err != nil {
		return nil, err
	}

	var numRects uint16
	if err := binary.Read(r, binary.BigEndian, &numRects); err != nil {
		return nil, err
	}

	// Build the map of encodings supported
	encMap := make(map[int32]Encoding)
	for _, enc := range c.Encs {
		encMap[enc.Type()] = enc
	}

	// We must always support the raw encoding
	rawEnc := new(RawEncoding)
	encMap[rawEnc.Type()] = rawEnc

	rectangles := make([]image.Image, 0, numRects)
	for i := uint16(0); i < numRects; i++ {
		var hdr rectangleHeader
		if err := binary.Read(r, binary.BigEndian, &hdr); err != nil {
			return nil, err
		}
		rectangle := image.Rectangle{
			image.Point{int(hdr.X), int(hdr.Y)},
			image.Point{int(hdr.X + hdr.Width), int(hdr.Y + hdr.Height)},
		}

		enc, ok := encMap[hdr.Enc]
		if !ok {
			return nil, fmt.Errorf("unsupported encoding type: %d", hdr.Enc)
		}

		if img, err := enc.Read(c, rectangle, r); err != nil {
			return nil, err
		} else {
			rectangles = append(rectangles, img)
		}
	}

	return &FramebufferUpdateMessage{rectangles}, nil
}

// SetColorMapEntriesMessage is sent by the server to set values into
// the color map. This message will automatically update the color map
// for the associated connection, but contains the color change data
// if the consumer wants to read it.
//
// See RFC 6143 Section 7.6.2
type SetColorMapEntriesMessage struct {
	FirstColor uint16
	Colors     []color.RGBA
}

func (*SetColorMapEntriesMessage) Type() uint8 {
	return 1
}

func (*SetColorMapEntriesMessage) Read(c *ClientConn, r io.Reader) (ServerMessage, error) {
	// Read off the padding
	var padding [1]byte
	if _, err := io.ReadFull(r, padding[:]); err != nil {
		return nil, err
	}

	var result SetColorMapEntriesMessage
	if err := binary.Read(r, binary.BigEndian, &result.FirstColor); err != nil {
		return nil, err
	}

	var numColors uint16
	if err := binary.Read(r, binary.BigEndian, &numColors); err != nil {
		return nil, err
	}

	result.Colors = make([]color.RGBA, numColors)
	for i := uint16(0); i < numColors; i++ {

		color := &result.Colors[i]
		data := []interface{}{
			&color.R,
			&color.G,
			&color.B,
		}

		for _, val := range data {
			if err := binary.Read(r, binary.BigEndian, val); err != nil {
				return nil, err
			}
		}

		// Update the connection's color map
		c.ColorMap[result.FirstColor+i] = *color
	}

	return &result, nil
}

// Bell signals that an audible bell should be made on the client.
//
// See RFC 6143 Section 7.6.3
type BellMessage byte

func (*BellMessage) Type() uint8 {
	return 2
}

func (*BellMessage) Read(*ClientConn, io.Reader) (ServerMessage, error) {
	return new(BellMessage), nil
}

// ServerCutTextMessage indicates the server has new text in the cut buffer.
//
// See RFC 6143 Section 7.6.4
type ServerCutTextMessage struct {
	Text string
}

func (*ServerCutTextMessage) Type() uint8 {
	return 3
}

func (*ServerCutTextMessage) Read(c *ClientConn, r io.Reader) (ServerMessage, error) {
	// Read off the padding
	var padding [1]byte
	if _, err := io.ReadFull(r, padding[:]); err != nil {
		return nil, err
	}

	var textLength uint32
	if err := binary.Read(r, binary.BigEndian, &textLength); err != nil {
		return nil, err
	}

	textBytes := make([]uint8, textLength)
	if err := binary.Read(r, binary.BigEndian, &textBytes); err != nil {
		return nil, err
	}

	return &ServerCutTextMessage{string(textBytes)}, nil
}

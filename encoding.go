package vnc

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"io"
)

// An Encoding implements a method for encoding pixel data that is
// sent by the server to the client.
type Encoding interface {
	// The number that uniquely identifies this encoding type.
	Type() int32

	// Read reads the contents of the encoded pixel data from the reader.
	// This should return a new Encoding implementation that contains
	// the proper data.
	Read(*ClientConn, image.Rectangle, io.Reader) (image.Image, error)
}

// RawEncoding is raw pixel data sent by the server.
//
// See RFC 6143 Section 7.7.1
type RawEncoding struct{}

func (*RawEncoding) Type() int32 {
	return 0
}

func (*RawEncoding) Read(c *ClientConn, rect image.Rectangle, r io.Reader) (image.Image, error) {
	bytesPerPixel := c.PixelFormat.BPP / 8
	pixelBytes := make([]uint8, bytesPerPixel)

	var byteOrder binary.ByteOrder = binary.LittleEndian
	if c.PixelFormat.BigEndian {
		byteOrder = binary.BigEndian
	}

	var img draw.Image
	if c.PixelFormat.TrueColor {
		img = image.NewRGBA(rect)
	} else {
		var palette [256]color.Color
		for i := 0; i < 256; i++ {
			palette[i] = c.ColorMap[i]
		}
		img = image.NewPaletted(rect, palette[:])
	}

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if _, err := io.ReadFull(r, pixelBytes); err != nil {
				return nil, err
			}

			var rawPixel uint32
			if c.PixelFormat.BPP == 8 {
				rawPixel = uint32(pixelBytes[0])
			} else if c.PixelFormat.BPP == 16 {
				rawPixel = uint32(byteOrder.Uint16(pixelBytes))
			} else if c.PixelFormat.BPP == 32 {
				rawPixel = byteOrder.Uint32(pixelBytes)
			}

			var pixel color.RGBA
			if c.PixelFormat.TrueColor {
				pixel.R = uint8(((uint16(rawPixel>>c.PixelFormat.RedShift) &
					c.PixelFormat.RedMax) * 255) / c.PixelFormat.RedMax)
				pixel.G = uint8(((uint16(rawPixel>>c.PixelFormat.BlueShift) &
					c.PixelFormat.BlueMax) * 255) / c.PixelFormat.BlueMax)
				pixel.B = uint8(((uint16(rawPixel>>c.PixelFormat.GreenShift) &
					c.PixelFormat.GreenMax) * 255) / c.PixelFormat.GreenMax)
				pixel.A = 255
			} else {
				pixel = c.ColorMap[rawPixel]
			}
			img.Set(x, y, pixel)
		}
	}

	return img, nil
}

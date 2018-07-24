package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"net"

	"github.com/mitchellh/go-vnc"
)

func TakeScreenshot(address, password string) (*image.RGBA, error) {
	nc, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	defer nc.Close()

	serverMessageChannel := make(chan vnc.ServerMessage)

	vncClient, err := vnc.Client(nc, &vnc.ClientConfig{
		Auth: []vnc.ClientAuth{
			&vnc.PasswordAuth{Password: password},
		},
		ServerMessages: []vnc.ServerMessage{
			&vnc.FramebufferUpdateMessage{},
		},
		ServerMessageCh: serverMessageChannel,
	})
	if err != nil {
		return nil, err
	}
	defer vncClient.Close()

	err = vncClient.FramebufferUpdateRequest(false, 0, 0,
		vncClient.FrameBufferWidth, vncClient.FrameBufferHeight)
	if err != nil {
		return nil, err
	}

	serverMessage := <-serverMessageChannel

	rects := serverMessage.(*vnc.FramebufferUpdateMessage).Rectangles
	if len(rects) == 0 {
		panic("vnc: framebuffer rects length")
	}

	w := int(rects[0].Width)
	h := int(rects[0].Height)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	enc := rects[0].Enc.(*vnc.RawEncoding)
	for i, c := range enc.Colors {
		x, y := i%w, i/w
		r, g, b := uint8(c.R), uint8(c.G), uint8(c.B)

		img.Set(x, y, color.RGBA{r, g, b, 255})
	}

	return img, nil
}

func main() {
	img, err := TakeScreenshot("server:port", "password")
	if err != nil {
		panic(err)
	}

	data := bytes.Buffer{}
	pngEncoder := png.Encoder{CompressionLevel: png.NoCompression}

	err = pngEncoder.Encode(io.Writer(&data), img)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("screenshot.png", data.Bytes(), 0600)
}

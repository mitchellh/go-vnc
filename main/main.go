package main

import (
	"flag"
	"fmt"
	"github.com/mitchellh/go-vnc"
	"github.com/veandco/go-sdl2/sdl"
	"net"
	"sync"
	"time"
)

var (
	addr  string
	mutex sync.Mutex
)

func main() {
	flag.StringVar(&addr, "a", "webrtc.touchiot.top:60001", "VNC Address")
	//flag.StringVar(&addr, "a", "10.190.50.76:5900", "VNC Address")
	flag.Parse()
	fmt.Println(addr)
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()
	conn, _ := net.Dial("tcp", addr)
	c := make(chan vnc.ServerMessage, 1)
	conf := vnc.ClientConfig{ServerMessageCh: c}

	client, err := vnc.Client(conn, &conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("h:%d,w:%d\n", client.FrameBufferHeight, client.FrameBufferWidth)
	fmt.Println("Begin-CreateWindow")
	window, err := sdl.CreateWindow("SDL", 0, 20,
		int32(client.FrameBufferWidth), int32(client.FrameBufferHeight), sdl.WINDOW_SHOWN)
	fmt.Println("End-CreateWindow")
	//fullRect :=sdl.Rect{0,0,int32(client.FrameBufferWidth), int32(client.FrameBufferHeight)}
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Created Window")
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer renderer.Destroy()
	fmt.Println("Created Renderer")
	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, int32(client.FrameBufferWidth), int32(client.FrameBufferHeight))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer texture.Destroy()
	fmt.Println("Created Texture")
	go func() {
		fmt.Println("Begin Recv Message")
		client.FramebufferUpdateRequest(true, 0, 0, client.FrameBufferWidth, client.FrameBufferHeight)
		for {

			select {
			case msg := <-c:


				switch t := msg.(type) {
				case *vnc.FramebufferUpdateMessage:
					//fmt.Println("FramebufferUpdateMessage")
					for _, rect := range t.Rectangles {
						//fmt.Printf("%v", rect)
						raw := rect.Enc.(*vnc.RawEncoding)
						sdlRect := sdl.Rect{X: int32(rect.X), Y: int32(rect.Y), W: int32(rect.Width), H: int32(rect.Height)}
						err := texture.UpdateRGBA(&sdlRect, raw.RawPixel, int(rect.Width))
						if err != nil {
							fmt.Println(err)
							continue
						}
						W, H, _ := renderer.GetOutputSize()
						renderRect := sdl.Rect{X: 0, Y: 0, W: W, H: H}
						//fmt.Printf("-%v ", renderRect)
						err = renderer.Copy(texture, nil, &renderRect)
						if err != nil {
							fmt.Println(err)
							continue
						}

					}
					//fmt.Println(" - ")
					renderer.Present()
					time.Sleep(50 * time.Millisecond)
					mutex.Lock()
					client.FramebufferUpdateRequest(true, 0, 0, client.FrameBufferWidth, client.FrameBufferHeight)
					mutex.Unlock()
					break
				case *vnc.SetColorMapEntriesMessage:
					fmt.Println("SetColorMapEntriesMessage")
					break
				case *vnc.BellMessage:
					fmt.Println("BellMessage")
					break
				case *vnc.ServerCutTextMessage:

					fmt.Println("ServerCutTextMessage=", t.Text)
					break
				default:

				}

				//case <-time.After(3 * time.Second):

			}
		}
	}()
	//client.FramebufferUpdateRequest(true, 0, 0, client.FrameBufferWidth, client.FrameBufferHeight)
	//for  {
	//time.Sleep(500 * time.Millisecond)

	//client.PointerEvent (vnc.ButtonMiddle,100,100)
	//client.KeyEvent("\r",true)
	//fmt.Printf("====\n")
	//}
	//| SDL_WINDOWEVENT_CLOSE -> 14
	//| SDL_WINDOWEVENT_FOCUS_LOST -> 13
	//| SDL_WINDOWEVENT_FOCUS_GAINED -> 12
	//| SDL_WINDOWEVENT_LEAVE -> 11
	//| SDL_WINDOWEVENT_ENTER -> 10
	//| SDL_WINDOWEVENT_RESTORED -> 9
	//| SDL_WINDOWEVENT_MAXIMIZED -> 8
	//| SDL_WINDOWEVENT_MINIMIZED -> 7
	//| SDL_WINDOWEVENT_SIZE_CHANGED -> 6
	//| SDL_WINDOWEVENT_RESIZED -> 5
	//| SDL_WINDOWEVENT_MOVED -> 4
	//| SDL_WINDOWEVENT_EXPOSED -> 3
	//| SDL_WINDOWEVENT_HIDDEN -> 2
	//| SDL_WINDOWEVENT_SHOWN -> 1
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				//fmt.Println("KeyboardEvent",e.State,e.Keysym.Sym)
				client.KeyEvent(uint32(e.Keysym.Sym), e.State > 0)
				break
			case *sdl.MouseMotionEvent:
				W, H, _ := renderer.GetOutputSize()
				x := e.X * int32(client.FrameBufferWidth) / W
				y := e.Y * int32(client.FrameBufferHeight) / H
				//fmt.Println(x,y)
				mutex.Lock()
				client.PointerEvent(0, uint16(x), uint16(y))
				mutex.Unlock()
				break
			case *sdl.MouseButtonEvent:
				W, H, _ := renderer.GetOutputSize()
				x := e.X * int32(client.FrameBufferWidth) / W
				y := e.Y * int32(client.FrameBufferHeight) / H
				//fmt.Println("MouseButtonEvent",e.Button,e.State,x,y)
				mutex.Lock()
				if e.State == 0 {
					client.PointerEvent(vnc.ButtonMask(e.Button), uint16(x), uint16(y))
				}
				mutex.Unlock()
				break
			case *sdl.WindowEvent:
				break

			}
		}
		sdl.Delay(16)
	}

}

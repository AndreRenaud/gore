package main

import (
	"image"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/AndreRenaud/gore"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type DoomGame struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture

	events      []gore.DoomEvent
	lock        sync.Mutex
	terminating bool
	running     bool

	// Channel for frame updates from DOOM goroutine
	frameChannel chan *image.RGBA
	titleChannel chan string
}

func NewDoomGame() (*DoomGame, error) {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	window, err := sdl.CreateWindow("SDL DOOM", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		screenWidth, screenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		window.Destroy()
		return nil, err
	}

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, screenWidth, screenHeight)
	if err != nil {
		renderer.Destroy()
		window.Destroy()
		return nil, err
	}

	return &DoomGame{
		window:       window,
		renderer:     renderer,
		texture:      texture,
		running:      true,
		frameChannel: make(chan *image.RGBA, 1),
		titleChannel: make(chan string, 1),
	}, nil
}

func (g *DoomGame) Close() {
	if g.texture != nil {
		g.texture.Destroy()
	}
	if g.renderer != nil {
		g.renderer.Destroy()
	}
	if g.window != nil {
		g.window.Destroy()
	}
	sdl.Quit()
}

func (g *DoomGame) mapSDLKeyToDoom(key sdl.Scancode) uint8 {
	keyMap := map[sdl.Scancode]uint8{
		sdl.SCANCODE_SPACE:     gore.KEY_USE1,
		sdl.SCANCODE_ESCAPE:    gore.KEY_ESCAPE,
		sdl.SCANCODE_UP:        gore.KEY_UPARROW1,
		sdl.SCANCODE_DOWN:      gore.KEY_DOWNARROW1,
		sdl.SCANCODE_LEFT:      gore.KEY_LEFTARROW1,
		sdl.SCANCODE_RIGHT:     gore.KEY_RIGHTARROW1,
		sdl.SCANCODE_RETURN:    gore.KEY_ENTER,
		sdl.SCANCODE_LCTRL:     gore.KEY_FIRE1,
		sdl.SCANCODE_RCTRL:     gore.KEY_FIRE1,
		sdl.SCANCODE_LSHIFT:    0x80 + 0x36,
		sdl.SCANCODE_RSHIFT:    0x80 + 0x36,
		sdl.SCANCODE_BACKSPACE: gore.KEY_BACKSPACE3,
		sdl.SCANCODE_Y:         'y',
		sdl.SCANCODE_N:         'n',
		sdl.SCANCODE_I:         'i',
		sdl.SCANCODE_D:         'd',
		sdl.SCANCODE_F:         'f',
		sdl.SCANCODE_A:         'a',
		sdl.SCANCODE_E:         'e',
		sdl.SCANCODE_R:         'r',
		sdl.SCANCODE_V:         'v',
		sdl.SCANCODE_C:         'c',
		sdl.SCANCODE_L:         'l',
		sdl.SCANCODE_Q:         'q',
		sdl.SCANCODE_1:         '1',
		sdl.SCANCODE_2:         '2',
		sdl.SCANCODE_3:         '3',
		sdl.SCANCODE_4:         '4',
		sdl.SCANCODE_5:         '5',
		sdl.SCANCODE_6:         '6',
		sdl.SCANCODE_7:         '7',
		sdl.SCANCODE_8:         '8',
		sdl.SCANCODE_9:         '9',
		sdl.SCANCODE_0:         '0',
	}

	if doomKey, exists := keyMap[key]; exists {
		return doomKey
	}
	return 0
}

func (g *DoomGame) handleEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			g.terminating = true
			g.running = false

		case *sdl.KeyboardEvent:
			doomKey := g.mapSDLKeyToDoom(e.Keysym.Scancode)
			if doomKey != 0 {
				var doomEvent gore.DoomEvent
				if e.Type == sdl.KEYDOWN {
					doomEvent.Type = gore.Ev_keydown
				} else {
					doomEvent.Type = gore.Ev_keyup
				}
				doomEvent.Key = doomKey

				g.lock.Lock()
				g.events = append(g.events, doomEvent)
				g.lock.Unlock()
			}

		case *sdl.MouseMotionEvent:
			var mouseEvent gore.DoomEvent
			mouseEvent.Type = gore.Ev_mouse
			mouseEvent.Mouse.XPos = float64(e.X) / float64(screenWidth)
			mouseEvent.Mouse.YPos = float64(e.Y) / float64(screenHeight)

			g.lock.Lock()
			g.events = append(g.events, mouseEvent)
			g.lock.Unlock()

		case *sdl.MouseButtonEvent:
			var mouseEvent gore.DoomEvent
			mouseEvent.Type = gore.Ev_mouse

			// Get current mouse position
			x, y, _ := sdl.GetMouseState()
			mouseEvent.Mouse.XPos = float64(x) / float64(screenWidth)
			mouseEvent.Mouse.YPos = float64(y) / float64(screenHeight)

			if e.Button == sdl.BUTTON_LEFT {
				mouseEvent.Mouse.Button1 = (e.Type == sdl.MOUSEBUTTONDOWN)
			} else if e.Button == sdl.BUTTON_RIGHT {
				mouseEvent.Mouse.Button2 = (e.Type == sdl.MOUSEBUTTONDOWN)
			}

			g.lock.Lock()
			g.events = append(g.events, mouseEvent)
			g.lock.Unlock()
		}
	}
}

func (g *DoomGame) GetEvent(event *gore.DoomEvent) bool {
	g.lock.Lock()
	defer g.lock.Unlock()
	if len(g.events) > 0 {
		*event = g.events[0]
		g.events = g.events[1:]
		return true
	}
	return false
}

func (g *DoomGame) DrawFrame(frame *image.RGBA) {
	// Send frame to main thread via channel (non-blocking)
	select {
	case g.frameChannel <- frame:
	default:
		// Drop frame if channel is full to prevent blocking
	}
}

func (g *DoomGame) SetTitle(title string) {
	// Send title update to main thread via channel (non-blocking)
	select {
	case g.titleChannel <- title:
	default:
		// Drop title update if channel is full
	}
}

func (g *DoomGame) drawFrameOnMainThread(frame *image.RGBA) {
	// Ensure we're on the main thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	bounds := frame.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Validate frame dimensions
	if width <= 0 || height <= 0 {
		log.Printf("Invalid frame dimensions: %dx%d", width, height)
		return
	}

	// Recreate texture if dimensions changed
	if g.texture != nil {
		_, _, textureWidth, textureHeight, err := g.texture.Query()
		if err != nil || int(textureWidth) != width || int(textureHeight) != height {
			g.texture.Destroy()
			g.texture = nil
		}
	}

	if g.texture == nil {
		var err error
		g.texture, err = g.renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, int32(width), int32(height))
		if err != nil {
			log.Printf("Failed to create texture: %v", err)
			return
		}
	}

	// Copy frame data directly without extra conversion
	frameData := make([]uint32, width*height)
	for i := 0; i < width*height; i++ {
		pixPos := i * 4
		frameData[i] = uint32(frame.Pix[pixPos+0])<<24 | uint32(frame.Pix[pixPos+1])<<16 | uint32(frame.Pix[pixPos+2])<<8 | uint32(frame.Pix[pixPos+3])
	}

	// Update texture with error handling
	err := g.texture.UpdateRGBA(nil, frameData, width)
	if err != nil {
		log.Printf("Failed to update texture: %v", err)
		return
	}

	// Clear and render with error handling
	err = g.renderer.Clear()
	if err != nil {
		log.Printf("Failed to clear renderer: %v", err)
		return
	}

	err = g.renderer.Copy(g.texture, nil, nil)
	if err != nil {
		log.Printf("Failed to copy texture: %v", err)
		return
	}

	g.renderer.Present()
}

func (g *DoomGame) Run() {
	// Start DOOM in a separate goroutine
	go func() {
		gore.Run(g, os.Args[1:])
		g.terminating = true
		g.running = false
	}()

	// Main event loop - MUST run on main thread for macOS compatibility
	for g.running {
		g.handleEvents()

		select {
		case frame := <-g.frameChannel:
			g.drawFrameOnMainThread(frame)
		case title := <-g.titleChannel:
			g.window.SetTitle(title)
		default:
			// No frame available, continue
		}

		sdl.Delay(16) // ~60 FPS
	}
}

func main() {
	game, err := NewDoomGame()
	if err != nil {
		log.Fatal("Failed to initialize SDL DOOM:", err)
	}
	defer game.Close()

	log.Println("Starting SDL DOOM...")
	game.Run()
}

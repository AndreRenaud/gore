package main

import (
	"encoding/binary"
	"image"
	"io"
	"log"
	"os"
	"runtime"
	"slices"
	"sync"

	"github.com/AndreRenaud/gore"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	screenWidth  = 640
	screenHeight = 480

	audioChunkFrames = 512                      // Frames mixed at a time (~12ms)
	audioQueueBytes  = audioChunkFrames * 4 * 3 // Keep ~35ms of audio queued, so sound effects are prompt
	maxSounds        = 16                       // Sound effects that can play at once
)

// sound is a cached sound effect, as 8-bit unsigned mono samples
type sound struct {
	samples []byte
	rate    int
}

type playingSound struct {
	sound       *sound
	pos         uint64 // Position in the samples, as 16.16 fixed point
	left, right int32  // Volume of each side, 0-254
}

type DoomGame struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture

	audioLock sync.Mutex
	audio     sdl.AudioDeviceID // 0 if audio isn't available
	music     io.Reader
	sounds    map[string]*sound
	playing   []*playingSound

	events      []gore.DoomEvent
	lock        sync.Mutex
	terminating bool
	running     bool

	// Channel for frame updates from DOOM goroutine
	frameChannel chan *image.RGBA
	titleChannel chan string
}

func NewDoomGame() (*DoomGame, error) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
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

	g := &DoomGame{
		window:       window,
		renderer:     renderer,
		texture:      texture,
		running:      true,
		frameChannel: make(chan *image.RGBA, 1),
		titleChannel: make(chan string, 1),
		sounds:       map[string]*sound{},
	}

	spec := sdl.AudioSpec{Freq: gore.MusicSampleRate, Format: sdl.AUDIO_S16LSB, Channels: 2, Samples: audioChunkFrames}
	if g.audio, err = sdl.OpenAudioDevice("", false, &spec, nil, 0); err != nil {
		log.Printf("No audio available: %s", err)
	} else {
		sdl.PauseAudioDevice(g.audio, false)
		go g.mixAudio()
	}
	return g, nil
}

// mixAudio keeps the SDL audio queue topped up with the music & sound effects
func (g *DoomGame) mixAudio() {
	buf := make([]byte, audioChunkFrames*4)
	mix := make([]int32, audioChunkFrames*2)
	for {
		g.audioLock.Lock()
		if g.audio == 0 {
			g.audioLock.Unlock()
			return
		}
		if sdl.GetQueuedAudioSize(g.audio) >= audioQueueBytes {
			g.audioLock.Unlock()
			sdl.Delay(5)
			continue
		}

		clear(mix)
		if g.music != nil {
			n, err := io.ReadFull(g.music, buf)
			for i := range n / 2 {
				mix[i] = int32(int16(binary.LittleEndian.Uint16(buf[i*2:])))
			}
			if err != nil {
				g.music = nil // Reached the end of the song
			}
		}
		g.mixSounds(mix)
		for i, v := range mix {
			binary.LittleEndian.PutUint16(buf[i*2:], uint16(int16(max(-32768, min(32767, v)))))
		}
		sdl.QueueAudio(g.audio, buf)
		g.audioLock.Unlock()
	}
}

// mixSounds adds the playing sound effects to mix, resampling them to the
// output rate (as DOOM did, without interpolation)
func (g *DoomGame) mixSounds(mix []int32) {
	g.playing = slices.DeleteFunc(g.playing, func(p *playingSound) bool {
		step := uint64(p.sound.rate) << 16 / gore.MusicSampleRate
		for i := 0; i < len(mix); i += 2 {
			pos := int(p.pos >> 16)
			if pos >= len(p.sound.samples) {
				return true // Finished
			}
			sample := (int32(p.sound.samples[pos]) - 128) << 8
			mix[i] += sample * p.left / 255
			mix[i+1] += sample * p.right / 255
			p.pos += step
		}
		return false
	})
}

func (g *DoomGame) Close() {
	g.audioLock.Lock()
	if g.audio != 0 {
		sdl.CloseAudioDevice(g.audio)
		g.audio = 0
	}
	g.audioLock.Unlock()
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

// CacheSound stores a sound effect, from a DMX format lump: a header with
// the sample rate & length, then 8-bit unsigned mono samples
func (g *DoomGame) CacheSound(name string, data []byte) {
	if len(data) < 8 || binary.LittleEndian.Uint16(data) != 3 {
		return
	}
	rate := int(binary.LittleEndian.Uint16(data[2:]))
	length := int(binary.LittleEndian.Uint32(data[4:]))
	// DMX skips the 16 bytes of padding at either end of the samples
	if rate == 0 || length <= 32 || length > len(data)-8 {
		return
	}
	g.audioLock.Lock()
	defer g.audioLock.Unlock()
	g.sounds[name] = &sound{samples: data[8+16 : 8+length-16], rate: rate}
}

// PlaySound starts a sound effect. Sounds are mixed independently, rather
// than per channel, as the engine doesn't know when a sound has finished
// and so reuses channels while their sounds are still playing.
func (g *DoomGame) PlaySound(name string, channel, vol, sep int) {
	g.audioLock.Lock()
	defer g.audioLock.Unlock()
	s := g.sounds[name]
	if s == nil {
		return
	}
	if len(g.playing) >= maxSounds {
		g.playing = g.playing[1:] // Drop the oldest
	}
	// sep is 0 (left) to 254 (right), and vol is 0-127
	g.playing = append(g.playing, &playingSound{
		sound: s,
		left:  int32((254 - sep) * vol / 127),
		right: int32(sep * vol / 127),
	})
}

func (g *DoomGame) PlayMusic(music io.Reader) {
	g.audioLock.Lock()
	defer g.audioLock.Unlock()
	g.music = music
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

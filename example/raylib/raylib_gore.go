package main

import (
	"encoding/binary"
	"image"
	"io"
	"os"
	"sync"

	"github.com/AndreRenaud/gore"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 640
	screenHeight = 480

	doomMaxVolume = 127 // DOOM's maximum sound volume
	doomCenterSep = 127 // DOOM's stereo separation for the center
)

var keyMap = map[int32]uint8{
	rl.KeyUp:           gore.KEY_UPARROW1,
	rl.KeyDown:         gore.KEY_DOWNARROW1,
	rl.KeyLeft:         gore.KEY_LEFTARROW1,
	rl.KeyRight:        gore.KEY_RIGHTARROW1,
	rl.KeyEnter:        gore.KEY_ENTER,
	rl.KeyKpEnter:      gore.KEY_ENTER,
	rl.KeyEscape:       gore.KEY_ESCAPE,
	rl.KeyTab:          gore.KEY_TAB,
	rl.KeyBackspace:    gore.KEY_BACKSPACE3,
	rl.KeySpace:        gore.KEY_USE1,
	rl.KeyLeftControl:  gore.KEY_FIRE1,
	rl.KeyRightControl: gore.KEY_FIRE1,
	rl.KeyLeftShift:    gore.KEY_RSHIFT,
	rl.KeyRightShift:   gore.KEY_RSHIFT,
	rl.KeyLeftAlt:      gore.KEY_RALT,
	rl.KeyRightAlt:     gore.KEY_RALT,
	rl.KeyMinus:        gore.KEY_MINUS1,
	rl.KeyEqual:        gore.KEY_EQUALS1,
	rl.KeyComma:        ',',
	rl.KeyPeriod:       '.',
	rl.KeyPause:        gore.KEY_PAUSE1,
	rl.KeyF11:          gore.KEY_F11,
	rl.KeyF12:          gore.KEY_F12,
}

func init() {
	for k := int32(rl.KeyA); k <= rl.KeyZ; k++ {
		keyMap[k] = uint8('a' + k - rl.KeyA)
	}
	for k := int32(rl.KeyZero); k <= rl.KeyNine; k++ {
		keyMap[k] = uint8('0' + k - rl.KeyZero)
	}
	for k := int32(rl.KeyF1); k <= rl.KeyF10; k++ {
		keyMap[k] = uint8(gore.KEY_F1 + k - rl.KeyF1)
	}
}

type soundData struct {
	samples []byte
	rate    int
}

type soundRequest struct {
	name                 string
	channel, volume, sep int
}

// DoomGame implements gore.DoomFrontend
type DoomGame struct {
	lock       sync.Mutex
	frame      *image.RGBA
	frameReady bool
	title      string
	titleReady bool
	events     []gore.DoomEvent
	newSounds  map[string]soundData
	soundQueue []soundRequest

	musicLock   sync.Mutex
	musicReader io.Reader
	musicBuffer []byte

	texture  rl.Texture2D
	front    *image.RGBA
	sounds   map[string]rl.Sound
	channels map[int]rl.Sound

	mouseCaptured bool
	mouseX        float64
	mouseY        float64
	mouseButton1  bool
	mouseButton2  bool
}

func (g *DoomGame) DrawFrame(frame *image.RGBA) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.frame == nil || g.frame.Bounds() != frame.Bounds() {
		g.frame = image.NewRGBA(frame.Bounds())
	}
	copy(g.frame.Pix, frame.Pix)
	g.frameReady = true
}

func (g *DoomGame) SetTitle(title string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.title = title
	g.titleReady = true
}

func (g *DoomGame) GetEvent(event *gore.DoomEvent) bool {
	g.lock.Lock()
	defer g.lock.Unlock()

	if len(g.events) == 0 {
		return false
	}
	*event = g.events[0]
	g.events = g.events[1:]
	return true
}

func (g *DoomGame) CacheSound(name string, data []byte) {
	if len(data) < 8 || binary.LittleEndian.Uint16(data) != 3 {
		return
	}
	rate := int(binary.LittleEndian.Uint16(data[2:]))
	length := int(binary.LittleEndian.Uint32(data[4:]))
	if rate == 0 || length <= 32 || length > len(data)-8 {
		return
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	g.newSounds[name] = soundData{samples: data[8+16 : 8+length-16], rate: rate}
}

func (g *DoomGame) PlaySound(name string, channel, volume, sep int) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.soundQueue = append(g.soundQueue, soundRequest{name: name, channel: channel, volume: volume, sep: sep})
}

func (g *DoomGame) PlayMusic(music io.Reader) {
	g.musicLock.Lock()
	defer g.musicLock.Unlock()

	g.musicReader = music
}

func (g *DoomGame) streamMusic(data []float32, frames int) {
	g.musicLock.Lock()
	defer g.musicLock.Unlock()

	size := frames * 4
	if cap(g.musicBuffer) < size {
		g.musicBuffer = make([]byte, size)
	}
	buffer := g.musicBuffer[:size]
	clear(buffer)
	if g.musicReader != nil {
		if _, err := io.ReadFull(g.musicReader, buffer); err != nil {
			g.musicReader = nil
		}
	}
	for i := range data {
		data[i] = float32(int16(binary.LittleEndian.Uint16(buffer[i*2:]))) / 32768
	}
}

func (g *DoomGame) pushEvent(event gore.DoomEvent) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.events = append(g.events, event)
}

func (g *DoomGame) handleInput() {
	for key, doomKey := range keyMap {
		if rl.IsKeyPressed(key) {
			g.pushEvent(gore.DoomEvent{Type: gore.Ev_keydown, Key: doomKey})
		} else if rl.IsKeyReleased(key) {
			g.pushEvent(gore.DoomEvent{Type: gore.Ev_keyup, Key: doomKey})
		}
	}

	if !g.mouseCaptured {
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && rl.IsWindowFocused() {
			rl.DisableCursor()
			g.mouseCaptured = true
		}
		return
	}
	if rl.IsKeyPressed(rl.KeyEscape) || !rl.IsWindowFocused() {
		rl.EnableCursor()
		g.mouseCaptured = false
		return
	}

	delta := rl.GetMouseDelta()
	button1 := rl.IsMouseButtonDown(rl.MouseButtonLeft)
	button2 := rl.IsMouseButtonDown(rl.MouseButtonRight)
	if delta.X == 0 && delta.Y == 0 && button1 == g.mouseButton1 && button2 == g.mouseButton2 {
		return
	}

	g.mouseX += float64(delta.X) / screenWidth
	g.mouseY += float64(delta.Y) / screenHeight
	g.mouseButton1 = button1
	g.mouseButton2 = button2

	event := gore.DoomEvent{Type: gore.Ev_mouse}
	event.Mouse.XPos = g.mouseX
	event.Mouse.YPos = g.mouseY
	event.Mouse.Button1 = button1
	event.Mouse.Button2 = button2
	g.pushEvent(event)
}

func (g *DoomGame) updateAudio() {
	g.lock.Lock()
	newSounds := g.newSounds
	queue := g.soundQueue
	g.newSounds = map[string]soundData{}
	g.soundQueue = nil
	g.lock.Unlock()

	for name, data := range newSounds {
		if _, ok := g.sounds[name]; ok {
			continue
		}
		wave := rl.NewWave(uint32(len(data.samples)), uint32(data.rate), 8, 1, data.samples)
		g.sounds[name] = rl.LoadSoundFromWave(wave)
	}

	for channel, alias := range g.channels {
		if !rl.IsSoundPlaying(alias) {
			rl.UnloadSoundAlias(alias)
			delete(g.channels, channel)
		}
	}

	for _, request := range queue {
		sound, ok := g.sounds[request.name]
		if !ok {
			continue
		}
		if alias, ok := g.channels[request.channel]; ok {
			rl.StopSound(alias)
			rl.UnloadSoundAlias(alias)
		}
		alias := rl.LoadSoundAlias(sound)
		rl.SetSoundVolume(alias, min(float32(request.volume)/doomMaxVolume, 1))
		rl.SetSoundPan(alias, max(-1, min(float32(request.sep-doomCenterSep)/doomCenterSep, 1)))
		rl.PlaySound(alias)
		g.channels[request.channel] = alias
	}
}

func (g *DoomGame) updateVideo() {
	g.lock.Lock()
	title, titleReady := g.title, g.titleReady
	frameReady := g.frameReady
	if frameReady {
		g.front, g.frame = g.frame, g.front
	}
	g.titleReady = false
	g.frameReady = false
	g.lock.Unlock()

	if titleReady {
		rl.SetWindowTitle(title)
	}
	if !frameReady {
		return
	}

	bounds := g.front.Bounds()
	if g.texture.ID == 0 || g.texture.Width != int32(bounds.Dx()) || g.texture.Height != int32(bounds.Dy()) {
		if g.texture.ID != 0 {
			rl.UnloadTexture(g.texture)
		}
		blank := rl.GenImageColor(bounds.Dx(), bounds.Dy(), rl.Black)
		g.texture = rl.LoadTextureFromImage(blank)
		rl.UnloadImage(blank)
	}
	rl.UpdateTexture(g.texture, g.front)
}

func (g *DoomGame) draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	if g.texture.ID != 0 {
		width := float32(rl.GetScreenWidth())
		height := float32(rl.GetScreenHeight())
		scale := min(width/4, height/3)
		dest := rl.NewRectangle((width-4*scale)/2, (height-3*scale)/2, 4*scale, 3*scale)
		source := rl.NewRectangle(0, 0, float32(g.texture.Width), float32(g.texture.Height))
		rl.DrawTexturePro(g.texture, source, dest, rl.Vector2{}, 0, rl.White)
	}

	rl.EndDrawing()
}

func main() {
	game := &DoomGame{
		newSounds: map[string]soundData{},
		sounds:    map[string]rl.Sound{},
		channels:  map[int]rl.Sound{},
	}

	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagVsyncHint)
	rl.InitWindow(screenWidth, screenHeight, "raylib DOOM")
	defer rl.CloseWindow()
	rl.SetExitKey(rl.KeyNull)

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()
	musicStream := rl.LoadAudioStream(gore.MusicSampleRate, 32, 2)
	defer rl.UnloadAudioStream(musicStream)
	rl.SetAudioStreamCallback(musicStream, game.streamMusic)
	rl.PlayAudioStream(musicStream)

	done := make(chan struct{})
	go func() {
		gore.Run(game, os.Args[1:])
		close(done)
	}()

	for running := true; running; {
		select {
		case <-done:
			running = false
			continue
		default:
		}

		if rl.WindowShouldClose() {
			gore.Stop()
			<-done
			break
		}

		game.handleInput()
		game.updateAudio()
		game.updateVideo()
		game.draw()
	}

	for _, alias := range game.channels {
		rl.UnloadSoundAlias(alias)
	}
	for _, sound := range game.sounds {
		rl.UnloadSound(sound)
	}
	if game.texture.ID != 0 {
		rl.UnloadTexture(game.texture)
	}
}

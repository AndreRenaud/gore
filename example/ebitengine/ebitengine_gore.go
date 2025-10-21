package main

import (
	"image"
	"log"
	"os"
	"sync"

	"github.com/AndreRenaud/gore"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 640
	screenHeight = 480

	// Audio constants
	ebitengineAudioSampleRate = 44100 // Standard sample rate for Ebitengine
	doomAudioSampleRate       = 11025 // DOOM's original sample rate
	doomMaxVolume             = 127   // DOOM's maximum volume value
	audioVolumeScale          = 0.5   // Scale factor to prevent clipping
)

type DoomGame struct {
	lastFrame *ebiten.Image

	audio *audio.Context

	events      []gore.DoomEvent
	lock        sync.Mutex
	terminating bool

	soundCache map[string]*audio.Player
}

func (g *DoomGame) Update() error {
	keys := map[ebiten.Key]uint8{
		ebiten.KeySpace:     gore.KEY_USE1,
		ebiten.KeyEscape:    gore.KEY_ESCAPE,
		ebiten.KeyUp:        gore.KEY_UPARROW1,
		ebiten.KeyDown:      gore.KEY_DOWNARROW1,
		ebiten.KeyLeft:      gore.KEY_LEFTARROW1,
		ebiten.KeyRight:     gore.KEY_RIGHTARROW1,
		ebiten.KeyEnter:     gore.KEY_ENTER,
		ebiten.KeyControl:   gore.KEY_FIRE1,
		ebiten.KeyShift:     0x80 + 0x36,
		ebiten.KeyBackspace: gore.KEY_BACKSPACE3,
		ebiten.KeyY:         'y',
		ebiten.KeyN:         'n',
		ebiten.KeyI:         'i',
		ebiten.KeyD:         'd',
		ebiten.KeyF:         'f',
		ebiten.KeyA:         'a',
		ebiten.KeyE:         'e',
		ebiten.KeyR:         'r',
		ebiten.KeyV:         'v',
		ebiten.KeyC:         'c',
		ebiten.KeyL:         'l',
		ebiten.KeyQ:         'q',
		ebiten.Key1:         '1',
		ebiten.Key2:         '2',
		ebiten.Key3:         '3',
		ebiten.Key4:         '4',
		ebiten.Key5:         '5',
		ebiten.Key6:         '6',
		ebiten.Key7:         '7',
		ebiten.Key8:         '8',
		ebiten.Key9:         '9',
		ebiten.Key0:         '0',
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	for key, doomKey := range keys {
		if inpututil.IsKeyJustPressed(key) {
			var event gore.DoomEvent

			event.Type = gore.Ev_keydown
			event.Key = doomKey
			g.events = append(g.events, event)
		} else if inpututil.IsKeyJustReleased(key) {
			var event gore.DoomEvent
			event.Type = gore.Ev_keyup
			event.Key = doomKey
			g.events = append(g.events, event)
		}

		var mouseEvent gore.DoomEvent
		x, y := ebiten.CursorPosition()
		mouseEvent.Mouse.XPos = float64(x) / float64(screenWidth)
		mouseEvent.Mouse.YPos = float64(y) / float64(screenHeight)
		mouseEvent.Type = gore.Ev_mouse
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			mouseEvent.Mouse.Button1 = true
		}
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			mouseEvent.Mouse.Button2 = true
		}
		g.events = append(g.events, mouseEvent)
	}
	if g.terminating {
		return ebiten.Termination
	}
	return nil
}

func (g *DoomGame) Draw(screen *ebiten.Image) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.lastFrame == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	rect := g.lastFrame.Bounds()
	yScale := float64(screenHeight) / float64(rect.Dy())
	xScale := float64(screenWidth) / float64(rect.Dx())
	op.GeoM.Scale(xScale, yScale)
	screen.DrawImage(g.lastFrame, op)
}

func (g *DoomGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
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
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.lastFrame != nil {
		if g.lastFrame.Bounds().Dx() != frame.Bounds().Dx() || g.lastFrame.Bounds().Dy() != frame.Bounds().Dy() {
			g.lastFrame.Deallocate()
			g.lastFrame = nil
		}
	}
	if g.lastFrame == nil {
		g.lastFrame = ebiten.NewImage(frame.Bounds().Dx(), frame.Bounds().Dy())
	}
	g.lastFrame.WritePixels(frame.Pix)
}

func (g *DoomGame) SetTitle(title string) {
	ebiten.SetWindowTitle(title)
}

// convertAudioSample converts DOOM's 8-bit unsigned mono PCM to 16-bit signed stereo PCM
// DOOM audio format: 8-bit unsigned PCM (0-255, with 128 = silence)
// Ebitengine format: 16-bit signed PCM stereo (-32768 to 32767, with 0 = silence)
func convertAudioSample(data []byte) []byte {
	// Each 8-bit sample becomes two 16-bit samples (left/right channels)
	// Output size: input samples * 2 channels * 2 bytes per sample = input * 4
	// Plus we 4x it to go from 11025Hz to 44100Hz
	converted := make([]byte, len(data)*4*4)

	for i, sample8 := range data {
		leftHigh := (sample8 - 128)

		// Duplicate it 4x to upsample, and left/right for stereo
		// Bottom byte is zero since we've only got 8-bit input
		baseIdx := i * 4 * 4
		converted[baseIdx+1] = leftHigh
		converted[baseIdx+3] = leftHigh
		converted[baseIdx+5] = leftHigh
		converted[baseIdx+7] = leftHigh
		converted[baseIdx+9] = leftHigh
		converted[baseIdx+11] = leftHigh
		converted[baseIdx+13] = leftHigh
		converted[baseIdx+15] = leftHigh
	}

	return converted
}

func (g *DoomGame) CacheSound(name string, data []byte) {
	// Convert DOOM's 8-bit mono audio @11025Hz to 16-bit stereo @44100
	convertedData := convertAudioSample(data)
	// Create and configure the audio player
	player := g.audio.NewPlayerFromBytes(convertedData)
	if g.soundCache == nil {
		g.soundCache = make(map[string]*audio.Player)
	}
	g.soundCache[name] = player

}

func (g *DoomGame) PlaySound(name string, channel, volume, sep int) {
	player, ok := g.soundCache[name]
	if !ok {
		log.Printf("Sound %s not found in cache", name)
		return
	}

	volumeScale := float64(volume) / float64(doomMaxVolume)
	if volumeScale > 1.0 {
		volumeScale = 1.0
	}
	player.SetVolume(volumeScale * audioVolumeScale)
	// Start playback
	player.Rewind()
	player.Play()

	// Note: The player will be garbage collected when the sound finishes
	// For a production implementation, you might want to track active players
}

func main() {
	game := &DoomGame{}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine Doom")
	ebiten.SetFullscreen(true)
	game.audio = audio.NewContext(ebitengineAudioSampleRate)
	go func() {
		gore.Run(game, os.Args[1:])
		game.terminating = true
	}()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

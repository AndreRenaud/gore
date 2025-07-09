package main

import (
	"image"
	"log"
	"os"
	"sync"

	"github.com/AndreRenaud/gore"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type DoomGame struct {
	lastFrame *ebiten.Image

	events      []gore.DoomEvent
	lock        sync.Mutex
	terminating bool
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

func main() {
	game := &DoomGame{}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Gamepad (Ebitengine Demo)")
	ebiten.SetFullscreen(true)
	go func() {
		gore.Run(game, os.Args[1:])
		game.terminating = true
	}()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

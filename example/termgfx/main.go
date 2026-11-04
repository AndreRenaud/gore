package main

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/AndreRenaud/gore"
	"golang.org/x/term"
)

type graphicsDoom struct {
	*keyboard
	kitty        bool // kitty graphics protocol, otherwise sixel
	cellW, cellH int  // cell size in pixels

	last  layout
	out   bytes.Buffer
	sixel sixelEncoder
}

// layout is where a frame is drawn on the terminal.
type layout struct {
	col, row   int // 1-based cell of the image's top left corner
	cols, rows int // cells covered by the image
	w, h       int // image size in pixels
}

// layout fits a sw x sh frame to the terminal, preserving its aspect ratio
// and centring it.
func (g *graphicsDoom) layout(sw, sh int) layout {
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		cols, rows = 80, 24
	}
	if !g.kitty {
		// a sixel image leaves the cursor on the row below it, which must not scroll the screen
		rows--
	}
	w, h := cols*g.cellW, rows*g.cellH
	if w*sh > h*sw {
		w = h * sw / sh
	} else {
		h = w * sh / sw
	}
	l := layout{w: w, h: h, cols: (w + g.cellW/2) / g.cellW, rows: (h + g.cellH/2) / g.cellH}
	l.col = (cols-l.cols)/2 + 1
	l.row = (rows-l.rows)/2 + 1
	return l
}

// DrawFrame writes the frame to stdout as a kitty or sixel image.
func (g *graphicsDoom) DrawFrame(img *image.RGBA) {
	b := img.Bounds()
	l := g.layout(b.Dx(), b.Dy())

	g.out.Reset()
	if l != g.last {
		// the terminal has been resized, so clear the old frame
		g.out.WriteString("\x1b[2J")
		g.last = l
	}
	fmt.Fprintf(&g.out, "\x1b[%d;%dH", l.row, l.col)
	if g.kitty {
		writeKitty(&g.out, img, l.cols, l.rows)
	} else {
		g.sixel.Encode(&g.out, img, l.w, l.h)
	}
	_, _ = os.Stdout.Write(g.out.Bytes())
}

// SetTitle sets the terminal window title.
func (g *graphicsDoom) SetTitle(title string) {
	// OSC title
	fmt.Fprintf(os.Stdout, "\x1b]0;%s\x07", title)
}

func (g *graphicsDoom) CacheSound(name string, data []byte) {
}

func (g *graphicsDoom) PlaySound(name string, channel, vol, sep int) {
}

var (
	kittyKeyboardRe = regexp.MustCompile(`\x1b\[\?\d+u`)
	cellSizeRe      = regexp.MustCompile(`\x1b\[6;(\d+);(\d+)t`)
	da1Re           = regexp.MustCompile(`\x1b\[\?[\d;]*c`)
)

// queryTerminal asks the terminal whether it supports the kitty graphics and
// keyboard protocols, and for its cell size in pixels (XTWINOPS 16). Every
// terminal answers the device attributes query (DA1) that follows, so its
// reply marks the end of the responses.
func queryTerminal(keys <-chan byte) (kittyGraphics, kittyKeyboard bool, cellW, cellH int) {
	_, _ = os.Stdout.WriteString("\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\\x1b[?u\x1b[16t\x1b[c")
	var resp []byte
	timeout := time.After(time.Second)
	for !da1Re.Match(resp) {
		select {
		case b, ok := <-keys:
			if !ok {
				return
			}
			resp = append(resp, b)
		case <-timeout:
			return
		}
	}

	kittyGraphics = bytes.Contains(resp, []byte("\x1b_Gi=31;OK"))
	kittyKeyboard = kittyKeyboardRe.Match(resp)
	if m := cellSizeRe.FindSubmatch(resp); m != nil {
		cellH, _ = strconv.Atoi(string(m[1]))
		cellW, _ = strconv.Atoi(string(m[2]))
	}
	return
}

func main() {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "terminal raw mode:", err)
		return
	}
	defer term.Restore(fd, oldState)

	keys := keyReader(os.Stdin)
	kitty, kittyKeyboard, cellW, cellH := queryTerminal(keys)
	if cellW == 0 || cellH == 0 {
		// small enough that the image fits on most terminals
		cellW, cellH = 8, 16
	}

	// clear screen, move home, hide cursor
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
	defer fmt.Print("\x1b[0m\x1b[2J\x1b[H\x1b[?25h")
	if kitty {
		// delete our image
		defer fmt.Print("\x1b_Ga=d,d=A,q=2\x1b\\")
	}
	if kittyKeyboard {
		// push our keyboard mode, popping it again on exit
		fmt.Printf("\x1b[>%du", kittyKeyboardFlags)
		defer fmt.Print("\x1b[<u")
	}

	gore.Run(&graphicsDoom{
		keyboard: &keyboard{keys: keys, kitty: kittyKeyboard, held: make(map[uint8]time.Time)},
		kitty:    kitty,
		cellW:    cellW,
		cellH:    cellH,
	}, os.Args[1:])
}

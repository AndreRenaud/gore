package main

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AndreRenaud/gore"
	"golang.org/x/term"
)

type graphicsDoom struct {
	*keyboard
	kitty        bool
	scale        int // fixed integer scale, or 0 to fill the terminal
	cellW, cellH int // cell size in pixels, 0 if unknown

	last  layout
	out   bytes.Buffer
	sixel sixelEncoder
	kty   kittyEncoder
}

// layout is where a frame is drawn on the terminal.
type layout struct {
	col, row   int // 1-based cell of the image's top left corner
	cols, rows int // cells covered by the image
	w, h       int // image size in pixels
}

// layout fits a sw x sh frame to the current terminal size, preserving its
// aspect ratio and centring it.
func (g *graphicsDoom) layout(sw, sh int) layout {
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || cols < 1 || rows < 2 {
		cols, rows = 80, 24
	}
	if !g.kitty {
		// a sixel image leaves the cursor on the row below it, which must not scroll the screen
		rows--
	}
	cellW, cellH := g.cellW, g.cellH
	if cellW == 0 || cellH == 0 {
		if !g.kitty {
			// without the cell size there is no way to know how many pixels fit
			s := max(g.scale, 1)
			return layout{col: 1, row: 1, w: sw * s, h: sh * s}
		}
		// kitty scales the image to cover the cells, so only the cell aspect ratio matters
		cellW, cellH = 1, 2
	}

	var l layout
	availW, availH := cols*cellW, rows*cellH
	switch {
	case g.scale > 0:
		l.w, l.h = sw*g.scale, sh*g.scale
	case availW*sh <= availH*sw:
		l.w, l.h = availW, availW*sh/sw
	default:
		l.w, l.h = availH*sw/sh, availH
	}
	l.cols = min(cols, max(1, (l.w+cellW/2)/cellW))
	l.rows = min(rows, max(1, (l.h+cellH/2)/cellH))
	l.col = max(0, (cols-l.cols)/2) + 1
	l.row = max(0, (rows-l.rows)/2) + 1
	return l
}

// DrawFrame writes the frame to stdout as a kitty or sixel image.
func (g *graphicsDoom) DrawFrame(img *image.RGBA) {
	b := img.Bounds()
	l := g.layout(b.Dx(), b.Dy())

	g.out.Reset()
	if l != g.last {
		// the terminal has been resized, so remove the old frame
		if g.kitty {
			g.out.WriteString(kittyDeleteAll)
		}
		g.out.WriteString("\x1b[2J")
		g.last = l
	}
	fmt.Fprintf(&g.out, "\x1b[%d;%dH", l.row, l.col)
	switch {
	case !g.kitty:
		g.sixel.Encode(&g.out, img, l.w, l.h)
	case g.scale > 0:
		// scale it ourselves, as the terminal may not scale by an exact integer
		g.kty.Encode(&g.out, img, g.scale, 0, 0)
	default:
		g.kty.Encode(&g.out, img, 1, l.cols, l.rows)
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
	kittyOKRe       = regexp.MustCompile(`\x1b_Gi=31;OK\x1b\\`)
	kittyKeyboardRe = regexp.MustCompile(`\x1b\[\?\d+u`)
	cellSizeRe      = regexp.MustCompile(`\x1b\[6;(\d+);(\d+)t`)
	pixelSizeRe     = regexp.MustCompile(`\x1b\[4;(\d+);(\d+)t`)
	da1Re           = regexp.MustCompile(`\x1b\[\?([\d;]*)c`)
)

// termInfo is what the terminal reports about its graphics support.
type termInfo struct {
	kitty, sixel  bool
	kittyKeyboard bool
	cellW, cellH  int // 0 if unknown
}

// queryTerminal asks the terminal whether it supports the kitty graphics
// and keyboard protocols, for its cell and text area sizes in pixels (XTWINOPS 16 and 14),
// and for its device attributes (DA1). Every terminal answers DA1, so its
// reply marks the end of the responses.
func queryTerminal(keys <-chan byte) termInfo {
	_, _ = os.Stdout.WriteString("\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\\x1b[?u\x1b[16t\x1b[14t\x1b[c")
	var resp []byte
	timeout := time.After(time.Second)
loop:
	for {
		select {
		case b, ok := <-keys:
			if !ok {
				break loop
			}
			resp = append(resp, b)
			if b == 'c' && da1Re.Match(resp) {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	var info termInfo
	info.kitty = kittyOKRe.Match(resp)
	info.kittyKeyboard = kittyKeyboardRe.Match(resp)
	if m := cellSizeRe.FindSubmatch(resp); m != nil {
		info.cellH, _ = strconv.Atoi(string(m[1]))
		info.cellW, _ = strconv.Atoi(string(m[2]))
	} else if m := pixelSizeRe.FindSubmatch(resp); m != nil {
		cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && cols > 0 && rows > 0 {
			h, _ := strconv.Atoi(string(m[1]))
			w, _ := strconv.Atoi(string(m[2]))
			info.cellW, info.cellH = w/cols, h/rows
		}
	}
	if m := da1Re.FindSubmatch(resp); m != nil {
		for attr := range strings.SplitSeq(string(m[1]), ";") {
			if attr == "4" {
				info.sixel = true
			}
		}
	}
	return info
}

// parseArgs removes our own -scale and -sixel options, leaving the rest for Doom.
func parseArgs(args []string) (scale int, forceSixel bool, rest []string, err error) {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-scale" && i+1 < len(args):
			scale, err = strconv.Atoi(args[i+1])
			if err != nil || scale < 1 {
				return 0, false, nil, fmt.Errorf("invalid -scale %q", args[i+1])
			}
			i++
		case args[i] == "-sixel":
			forceSixel = true
		default:
			rest = append(rest, args[i])
		}
	}
	return scale, forceSixel, rest, nil
}

func main() {
	scale, forceSixel, args, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "terminal raw mode:", err)
		return
	}
	defer term.Restore(fd, oldState)

	keys := keyReader(os.Stdin)
	info := queryTerminal(keys)
	kitty := info.kitty && !forceSixel
	if !kitty && !info.sixel {
		fmt.Print("Warning: terminal does not report kitty graphics or sixel support, output may be garbled\r\n")
		time.Sleep(2 * time.Second)
	}

	// clear screen, move home, hide cursor
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
	defer fmt.Print("\x1b[0m\x1b[2J\x1b[H\x1b[?25h")
	if kitty {
		defer fmt.Print(kittyDeleteAll)
	}
	if info.kittyKeyboard {
		// push our keyboard mode, popping it again on exit
		fmt.Printf("\x1b[>%du", kittyKeyboardFlags)
		defer fmt.Print("\x1b[<u")
	}

	gd := &graphicsDoom{
		keyboard: newKeyboard(keys, info.kittyKeyboard),
		kitty:    kitty,
		scale:    scale,
		cellW:    info.cellW,
		cellH:    info.cellH,
	}
	gore.Run(gd, args)
}

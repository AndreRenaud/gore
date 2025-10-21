package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"time"

	"github.com/AndreRenaud/gore"
	"github.com/nfnt/resize"
	"golang.org/x/term"
)

// Characters from bright to dark
const brightChars = "$@B%#*+=\"~^;:..."

type termDoom struct {
	keys            <-chan byte
	outstandingDown map[uint8]time.Time
}

// DrawFrame converts the RGBA frame to ANSI colored ASCII and writes to stdout.
func (t *termDoom) DrawFrame(img *image.RGBA) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w < 20 || h < 10 {
		w, h = 80, 24
	}
	// leave one row for safety to avoid weird scrolling issues on some terminals
	h--

	target := resize.Resize(uint(w), uint(h), img, resize.NearestNeighbor)

	var b bytes.Buffer
	// move cursor home
	b.WriteString("\x1b[H")

	rgba, _ := ensureRGBA(target)
	toASCII(&b, rgba)
	_, _ = os.Stdout.Write(b.Bytes())
}

// SetTitle sets the terminal window title.
func (t *termDoom) SetTitle(title string) {
	// OSC title
	fmt.Fprintf(os.Stdout, "\x1b]0;%s\x07", title)
}

// GetEvent provides keydown/keyup events from stdin without unix/syscalls.
func (t *termDoom) GetEvent(ev *gore.DoomEvent) bool {
	// emit pending key-up after a short delay
	const upDelay = 60 * time.Millisecond
	now := time.Now()
	for k, ts := range t.outstandingDown {
		if now.Sub(ts) >= upDelay {
			delete(t.outstandingDown, k)
			ev.Type = gore.Ev_keyup
			ev.Key = k
			return true
		}
	}

	// try to read a byte non-blocking
	select {
	case b, ok := <-t.keys:
		if !ok {
			return false
		}
		seq := []byte{b}
		if b == 0x1b { // ESC sequence for arrows
			select {
			case b2 := <-t.keys:
				seq = append(seq, b2)
				select {
				case b3 := <-t.keys:
					seq = append(seq, b3)
				default:
				}
			default:
			}
		}
		if k, ok := mapKey(seq); ok {
			ev.Type = gore.Ev_keydown
			ev.Key = k
			t.outstandingDown[k] = now
			return true
		}
		return false
	default:
		return false
	}
}

// ensureRGBA guarantees we have *image.RGBA for fast pixel walks.
func ensureRGBA(img image.Image) (*image.RGBA, bool) {
	if r, ok := img.(*image.RGBA); ok {
		return r, true
	}
	b := img.Bounds()
	r := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r.Set(x, y, img.At(x, y))
		}
	}
	return r, false
}

// toASCII writes a full-frame ANSI image using ramp + 24-bit color.
func toASCII(w *bytes.Buffer, img *image.RGBA) {
	b := img.Bounds()
	last := color.RGBA{}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			o := (y-b.Min.Y)*img.Stride + (x-b.Min.X)*4
			r := img.Pix[o+0]
			g := img.Pix[o+1]
			bl := img.Pix[o+2]
			brightness := int(r+g+bl) * (len(brightChars) - 1) / (3 * 255) // Normalize to 0-1 range
			ch := brightChars[brightness]

			// emit color only if it changed
			if r != last.R || g != last.G || bl != last.B {
				fmt.Fprintf(w, "\x1b[38;2;%d;%d;%dm", r, g, bl)
				last = color.RGBA{r, g, bl, 255}
			}
			_, _ = w.Write([]byte{byte(ch)})
		}
		// reset at EOL
		_, _ = w.Write([]byte("\x1b[0m\r\n"))
		last = color.RGBA{}
	}
}

func mapKey(seq []byte) (uint8, bool) {
	s := string(seq)
	switch s {
	case "\x1b[A":
		return gore.KEY_UPARROW1, true
	case "\x1b[B":
		return gore.KEY_DOWNARROW1, true
	case "\x1b[C":
		return gore.KEY_RIGHTARROW1, true
	case "\x1b[D":
		return gore.KEY_LEFTARROW1, true
	case " ", "\x1bOP":
		return gore.KEY_USE1, true
	case "\r", "\n":
		return gore.KEY_ENTER, true
	case "\x1b":
		return gore.KEY_ESCAPE, true
	case "\t":
		return gore.KEY_TAB, true
	case ",":
		return gore.KEY_FIRE1, true
	}
	// direct digits and y/n
	if len(seq) == 1 {
		if seq[0] >= '0' && seq[0] <= '9' {
			return seq[0], true
		}
		if seq[0] == 'y' || seq[0] == 'n' || seq[0] == 'Y' || seq[0] == 'N' {
			return toLower(seq[0]), true
		}
	}
	return 0, false
}

func toLower(b byte) uint8 {
	if b >= 'A' && b <= 'Z' {
		return b - 'A' + 'a'
	}
	return b
}

// keyReader returns a non-blocking byte channel backed by a goroutine.
func keyReader(r io.Reader) <-chan byte {
	ch := make(chan byte, 128)
	br := bufio.NewReader(r)
	go func() {
		defer close(ch)
		for {
			b, err := br.ReadByte()
			if err != nil {
				return
			}
			ch <- b
		}
	}()
	return ch
}

func main() {
	// raw mode and initial clear
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "terminal raw mode:", err)
		return
	}
	defer term.Restore(fd, oldState)
	// clear screen, move home, hide cursor
	fmt.Print("\x1b[2J\x1b[H\x1b[?25l")
	defer fmt.Print("\x1b[0m\x1b[2J\x1b[H\x1b[?25h")

	td := &termDoom{
		keys:            keyReader(os.Stdin),
		outstandingDown: make(map[uint8]time.Time),
	}
	gore.Run(td, os.Args[1:])
}

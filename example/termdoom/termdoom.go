package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"time"

	"github.com/AndreRenaud/gore"
	"github.com/nfnt/resize"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type termDoom struct {
	title           string
	outstandingKeys map[uint8]time.Time
	width           uint
}

func ascii(img image.Image, writer io.Writer) {
	asciiChar := []byte("$@B%#*+=,.....")
	bound := img.Bounds()
	height, width := bound.Max.Y, bound.Max.X

	for y := bound.Min.Y; y < height; y++ {
		for x := bound.Min.X; x < width; x++ {
			pixelValue := img.At(x, y)
			r, g, b, _ := pixelValue.RGBA()
			writer.Write([]byte(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r>>8, g>>8, b>>8)))

			pixel := color.GrayModel.Convert(pixelValue).(color.Gray).Y
			asciiIndex := int(pixel) * (len(asciiChar) - 1) / 255
			writer.Write([]byte{asciiChar[asciiIndex]})
		}
		writer.Write([]byte("\r\n"))
	}
}

func (t *termDoom) DrawFrame(frame *image.RGBA) {
	height := t.width * 200 / 320
	smaller := resize.Resize(t.width, height, frame, resize.Lanczos3)
	fmt.Print("\033[H\033[2J")
	ascii(smaller, os.Stdout)
	os.Stdout.Sync()
}

func (t *termDoom) GetEvent(event *gore.DoomEvent) bool {
	for key, lastTime := range t.outstandingKeys {
		if time.Since(lastTime) > time.Second {
			delete(t.outstandingKeys, key)
			event.Type = gore.Ev_keyup
			event.Key = key
			return true
		}
	}
	var buf [3]byte
	fds := []unix.PollFd{
		{Fd: int32(os.Stdin.Fd()), Events: unix.POLLIN},
	}
	if n, err := unix.Poll(fds, 0); err != nil || n <= 0 {
		return false // No event handled
	}
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
	}
	log.Printf("Read %d bytes: %v\n", n, buf[:n])
	switch string(buf[:n]) {
	case "\x1b[A": // Up arrow
		event.Type = gore.Ev_keydown
		event.Key = gore.KEY_UPARROW1
	case "\x1b[B": // Down arrow
		event.Type = gore.Ev_keydown
		event.Key = gore.KEY_DOWNARROW1
	case "\x1b[C": // Right arrow
		event.Type = gore.Ev_keydown
		event.Key = gore.KEY_RIGHTARROW1
	case "\x1b[D": // Left arrow
		event.Type = gore.Ev_keydown
		event.Key = gore.KEY_LEFTARROW1
	case " ":
		event.Type = gore.Ev_keydown
		event.Key = gore.KEY_USE1
	case "\n", "\r":
		event.Type = gore.Ev_keydown
		event.Key = gore.KEY_ENTER
	case "\x1b": // Escape
		event.Type = gore.Ev_keydown
		event.Key = gore.KEY_ESCAPE
	case ",":
		event.Type = gore.Ev_keydown
		event.Key = gore.KEY_FIRE1
	case "y", "n", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
		event.Type = gore.Ev_keydown
		event.Key = buf[0]
	default:
		return false

	}
	t.outstandingKeys[event.Key] = time.Now()
	return true
}

func (t *termDoom) SetTitle(title string) {
	t.title = title
}

func main() {
	// Initialize termDoom and start the game loop
	termGame := &termDoom{
		width:           80,
		outstandingKeys: make(map[uint8]time.Time),
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	gore.Run(termGame, os.Args[1:])
}

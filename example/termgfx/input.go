package main

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/AndreRenaud/gore"
)

const (
	keySpeed  = 0x80 + 0x36 // shift
	keyStrafe = 0x80 + 0x38 // alt

	// Report press/repeat/release events and every key (including
	// modifiers) as an escape code.
	// See https://sw.kovidgoyal.net/kitty/keyboard-protocol/
	kittyKeyboardFlags = 1 | 2 | 8

	// Without the kitty keyboard protocol a terminal only sends key presses,
	// repeating them while held. A key is considered released once its
	// repeats stop arriving.
	initialRepeatGap = 500 * time.Millisecond
	repeatGap        = 150 * time.Millisecond
)

var (
	// CSI code[:alternates] ; modifiers[:event] [; text] u, or
	// CSI 1 ; modifiers[:event] A-D for the arrows
	kittyKeyRe = regexp.MustCompile(`^\x1b\[(\d*)[\d:]*(?:;\d*(?::(\d))?)?(?:;[\d:]*)?([uABCD])$`)
	kittyKeys  = map[int]uint8{
		27:    gore.KEY_ESCAPE,
		13:    gore.KEY_ENTER,
		9:     gore.KEY_TAB,
		127:   gore.KEY_BACKSPACE3,
		' ':   gore.KEY_USE1,
		57441: keySpeed, // left/right shift
		57447: keySpeed,
		57442: gore.KEY_FIRE1, // left/right control
		57448: gore.KEY_FIRE1,
		57443: keyStrafe, // left/right alt
		57449: keyStrafe,
	}
	arrowKeys = map[string]uint8{
		"A": gore.KEY_UPARROW1,
		"B": gore.KEY_DOWNARROW1,
		"C": gore.KEY_RIGHTARROW1,
		"D": gore.KEY_LEFTARROW1,
	}

	// Traditional terminal input has no modifier keys, so fire is remapped to ','
	legacyKeys = map[string]uint8{
		"\x1b[A": gore.KEY_UPARROW1,
		"\x1b[B": gore.KEY_DOWNARROW1,
		"\x1b[C": gore.KEY_RIGHTARROW1,
		"\x1b[D": gore.KEY_LEFTARROW1,
		" ":      gore.KEY_USE1,
		"\r":     gore.KEY_ENTER,
		"\x1b":   gore.KEY_ESCAPE,
		"\t":     gore.KEY_TAB,
		"\x7f":   gore.KEY_BACKSPACE3,
		",":      gore.KEY_FIRE1,
	}
)

// keyboard turns terminal input into Doom key events.
type keyboard struct {
	keys  <-chan byte
	kitty bool                // terminal reports press and release via the kitty keyboard protocol
	held  map[uint8]time.Time // otherwise, when each key is considered released
}

// GetEvent provides keydown/keyup events from stdin without unix/syscalls.
func (k *keyboard) GetEvent(ev *gore.DoomEvent) bool {
	now := time.Now()
	for key, release := range k.held {
		if now.After(release) {
			delete(k.held, key)
			ev.Type, ev.Key = gore.Ev_keyup, key
			return true
		}
	}

	for {
		// try to read a byte non-blocking
		var b byte
		select {
		case c, ok := <-k.keys:
			if !ok {
				return false
			}
			b = c
		default:
			return false
		}
		seq := k.readSeq(b)

		if k.kitty {
			if key, typ := kittyKey(seq); key != 0 {
				ev.Type, ev.Key = typ, key
				return true
			}
			continue
		}

		key := legacyKey(seq)
		if key == 0 {
			continue
		}
		if _, down := k.held[key]; down {
			// auto-repeat of a held key
			k.held[key] = now.Add(repeatGap)
			continue
		}
		k.held[key] = now.Add(initialRepeatGap)
		ev.Type, ev.Key = gore.Ev_keydown, key
		return true
	}
}

// readSeq returns the complete escape sequence starting with b, or just b
// if it does not start one.
func (k *keyboard) readSeq(b byte) string {
	seq := []byte{b}
	for incomplete(seq) {
		// the rest of a sequence arrives together, so a short wait is enough
		select {
		case c, ok := <-k.keys:
			if !ok {
				return string(seq)
			}
			seq = append(seq, c)
		case <-time.After(5 * time.Millisecond):
			return string(seq)
		}
	}
	return string(seq)
}

// incomplete reports whether seq is the start of an unfinished escape sequence.
func incomplete(seq []byte) bool {
	switch {
	case seq[0] != 0x1b:
		return false
	case len(seq) == 1:
		return true
	case seq[1] != '[': // alt+key
		return false
	case len(seq) == 2:
		return true
	}
	// a CSI sequence ends with a byte in the range 0x40-0x7e
	last := seq[len(seq)-1]
	return last < 0x40 || last > 0x7e
}

// kittyKey decodes a kitty keyboard protocol sequence, returning 0 for keys
// Doom doesn't use.
func kittyKey(seq string) (uint8, gore.Evtype_t) {
	m := kittyKeyRe.FindStringSubmatch(seq)
	if m == nil || m[2] == "2" {
		// Doom does its own auto-repeat, so ignore the terminal's
		return 0, 0
	}
	key := arrowKeys[m[3]]
	if m[3] == "u" {
		code, _ := strconv.Atoi(m[1])
		key = kittyKeys[code]
		if key == 0 {
			key = printableKey(code)
		}
	}
	if m[2] == "3" {
		return key, gore.Ev_keyup
	}
	return key, gore.Ev_keydown
}

func legacyKey(seq string) uint8 {
	if key, ok := legacyKeys[seq]; ok {
		return key
	}
	if len(seq) == 1 {
		return printableKey(int(seq[0]))
	}
	return 0
}

// printableKey maps printable ASCII to its lower case Doom key, so menus and
// cheat codes work.
func printableKey(c int) uint8 {
	if c <= ' ' || c > '~' {
		return 0
	}
	if c >= 'A' && c <= 'Z' {
		c += 'a' - 'A'
	}
	return uint8(c)
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

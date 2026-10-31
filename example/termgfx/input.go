package main

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/AndreRenaud/gore"
)

const (
	keySpeed  = 0x80 + 0x36 // shift
	keyStrafe = 0x80 + 0x38 // alt
	keyF1     = 0x80 + 0x3b // F1-F10 are consecutive
	keyF11    = 0x80 + 0x57
	keyF12    = 0x80 + 0x58

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

// keyboard turns terminal input into Doom key events.
type keyboard struct {
	keys  <-chan byte
	kitty bool                // terminal reports press and release via the kitty keyboard protocol
	held  map[uint8]time.Time // raw mode: when each key is considered released
}

func newKeyboard(keys <-chan byte, kitty bool) *keyboard {
	return &keyboard{keys: keys, kitty: kitty, held: make(map[uint8]time.Time)}
}

// GetEvent provides keydown/keyup events from stdin without unix/syscalls.
func (k *keyboard) GetEvent(ev *gore.DoomEvent) bool {
	now := time.Now()
	for key, release := range k.held {
		if now.After(release) {
			delete(k.held, key)
			ev.Type = gore.Ev_keyup
			ev.Key = key
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
			if key, typ, ok := kittyKey(seq); ok {
				ev.Type = typ
				ev.Key = key
				return true
			}
			continue
		}

		key, ok := legacyKey(seq)
		if !ok {
			continue
		}
		if _, down := k.held[key]; down {
			// auto-repeat of a held key
			k.held[key] = now.Add(repeatGap)
			continue
		}
		k.held[key] = now.Add(initialRepeatGap)
		ev.Type = gore.Ev_keydown
		ev.Key = key
		return true
	}
}

// readSeq returns the complete escape sequence starting with b, or just b
// if it does not start one.
func (k *keyboard) readSeq(b byte) []byte {
	seq := []byte{b}
	if b != 0x1b {
		return seq
	}
	// the rest of a sequence arrives together, so a short wait is enough
	next := func() (byte, bool) {
		select {
		case c, ok := <-k.keys:
			return c, ok
		case <-time.After(5 * time.Millisecond):
			return 0, false
		}
	}
	c, ok := next()
	if !ok {
		return seq
	}
	seq = append(seq, c)
	switch c {
	case '[':
		// CSI: parameters, ended by a byte in the range 0x40-0x7e
		for {
			c, ok := next()
			if !ok {
				return seq
			}
			seq = append(seq, c)
			if c >= 0x40 && c <= 0x7e {
				return seq
			}
		}
	case 'O':
		// SS3: one more byte
		if c, ok := next(); ok {
			seq = append(seq, c)
		}
	}
	return seq
}

// kittyKey decodes a kitty keyboard protocol sequence, of the form
// CSI code[:alternates] ; modifiers[:event] [; text] final
func kittyKey(seq []byte) (uint8, gore.Evtype_t, bool) {
	if len(seq) < 3 || !bytes.HasPrefix(seq, []byte("\x1b[")) {
		return 0, 0, false
	}
	final := seq[len(seq)-1]
	fields := strings.Split(string(seq[2:len(seq)-1]), ";")
	code, _ := strconv.Atoi(strings.Split(fields[0], ":")[0])
	event := 1
	if len(fields) > 1 {
		if mods := strings.Split(fields[1], ":"); len(mods) > 1 {
			event, _ = strconv.Atoi(mods[1])
		}
	}

	var key uint8
	switch final {
	case 'u':
		key = kittyCodeKey(code)
	case '~':
		key = tildeKey(code)
	default:
		key = finalKey(final)
	}
	if key == 0 {
		return 0, 0, false
	}
	switch event {
	case 1:
		return key, gore.Ev_keydown, true
	case 3:
		return key, gore.Ev_keyup, true
	}
	// Doom does its own auto-repeat, so ignore the terminal's
	return 0, 0, false
}

func kittyCodeKey(code int) uint8 {
	switch code {
	case 27:
		return gore.KEY_ESCAPE
	case 13:
		return gore.KEY_ENTER
	case 9:
		return gore.KEY_TAB
	case 127:
		return gore.KEY_BACKSPACE3
	case ' ':
		return gore.KEY_USE1
	case 57441, 57447: // left/right shift
		return keySpeed
	case 57442, 57448: // left/right control
		return gore.KEY_FIRE1
	case 57443, 57449: // left/right alt
		return keyStrafe
	}
	return printableKey(code)
}

// tildeKey maps the number of a CSI number ~ sequence to a function key.
func tildeKey(n int) uint8 {
	switch n {
	case 11, 12, 13, 14:
		return keyF1 + uint8(n-11)
	case 15:
		return keyF1 + 4
	case 17, 18, 19, 20, 21:
		return keyF1 + 5 + uint8(n-17)
	case 23:
		return keyF11
	case 24:
		return keyF12
	}
	return 0
}

// finalKey maps the final byte of a CSI or SS3 sequence for arrows and F1-F4.
func finalKey(final byte) uint8 {
	switch final {
	case 'A':
		return gore.KEY_UPARROW1
	case 'B':
		return gore.KEY_DOWNARROW1
	case 'C':
		return gore.KEY_RIGHTARROW1
	case 'D':
		return gore.KEY_LEFTARROW1
	case 'P', 'Q', 'R', 'S':
		return keyF1 + final - 'P'
	}
	return 0
}

// legacyKey maps traditional terminal input, which has no modifier keys of
// its own, so fire is remapped to ','.
func legacyKey(seq []byte) (uint8, bool) {
	switch s := string(seq); {
	case s == " ":
		return gore.KEY_USE1, true
	case s == "\r" || s == "\n":
		return gore.KEY_ENTER, true
	case s == "\x1b":
		return gore.KEY_ESCAPE, true
	case s == "\t":
		return gore.KEY_TAB, true
	case s == "\x7f" || s == "\b":
		return gore.KEY_BACKSPACE3, true
	case s == ",":
		return gore.KEY_FIRE1, true
	case len(seq) == 1:
		key := printableKey(int(seq[0]))
		return key, key != 0
	case len(seq) >= 3 && (seq[1] == '[' || seq[1] == 'O'):
		final := seq[len(seq)-1]
		key := finalKey(final)
		if final == '~' {
			n, _ := strconv.Atoi(strings.Split(string(seq[2:len(seq)-1]), ";")[0])
			key = tildeKey(n)
		}
		return key, key != 0
	}
	return 0, false
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

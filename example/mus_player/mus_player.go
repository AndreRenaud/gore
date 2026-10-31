// Package main is a simple MUS music player. It plays a MUS track from inside
// a WAD file through the system audio output, using the instruments from the
// WAD's GENMIDI lump.
//
//	mus_player -wad doom1.wad -track e1m1
//	mus_player -wad doom1.wad              # Lists the tracks in the WAD
//
// While playing, the left/right arrows seek 5 seconds, up/down seek 30
// seconds, space pauses, 'r' restarts and 'q' quits.
//
// It can also save the raw audio data to a file, which can be played with `sox`:
//
//	play -r 44100 -b 16 -c 2 -e signed-integer runnin.raw
package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/AndreRenaud/gore"
	"github.com/ebitengine/oto/v3"
	"golang.org/x/term"
)

type wadLump struct {
	Name string
	Data []byte
}

// readWadMusic returns all of the MUS format lumps in a WAD file, and its GENMIDI lump
func readWadMusic(filename string) (lumps []wadLump, genmidi []byte, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	if len(data) < 12 || (string(data[:4]) != "IWAD" && string(data[:4]) != "PWAD") {
		return nil, nil, fmt.Errorf("%s is not a WAD file", filename)
	}
	numLumps := int(binary.LittleEndian.Uint32(data[4:8]))
	dirOffset := int(binary.LittleEndian.Uint32(data[8:12]))
	if dirOffset < 0 || numLumps < 0 || dirOffset+numLumps*16 > len(data) {
		return nil, nil, fmt.Errorf("%s has an invalid lump directory", filename)
	}

	for i := range numLumps {
		entry := data[dirOffset+i*16 : dirOffset+(i+1)*16]
		pos := int(binary.LittleEndian.Uint32(entry[0:4]))
		size := int(binary.LittleEndian.Uint32(entry[4:8]))
		name, _, _ := strings.Cut(string(entry[8:16]), "\x00")
		if pos < 0 || size < 0 || pos+size > len(data) {
			continue
		}
		lump := data[pos : pos+size]
		if gore.IsMusData(lump) {
			lumps = append(lumps, wadLump{Name: strings.ToUpper(name), Data: lump})
		} else if strings.EqualFold(name, "GENMIDI") {
			genmidi = lump
		}
	}
	if genmidi == nil {
		return nil, nil, fmt.Errorf("%s has no GENMIDI lump", filename)
	}
	return lumps, genmidi, nil
}

// findTrack looks up a track by name, with or without the D_ prefix.
// Later lumps win, as they do when PWADs override an IWAD.
func findTrack(lumps []wadLump, track string) *wadLump {
	track = strings.ToUpper(track)
	for i := len(lumps) - 1; i >= 0; i-- {
		if lumps[i].Name == track || lumps[i].Name == "D_"+track {
			return &lumps[i]
		}
	}
	return nil
}

func listTracks(lumps []wadLump) {
	if len(lumps) == 0 {
		fmt.Println("No MUS tracks found")
		return
	}
	fmt.Printf("%d MUS tracks:\n", len(lumps))
	for _, l := range lumps {
		length := "invalid"
		if m, err := gore.NewMusMusic(l.Data); err == nil {
			length = formatDuration(m.Duration())
		}
		fmt.Printf("  %-8s %s\n", l.Name, length)
	}
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	return fmt.Sprintf("%d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}

type key int

const (
	keyLeft key = iota
	keyRight
	keyUp
	keyDown
	keyPause
	keyRestart
	keyQuit
)

func readKeys(r io.Reader, keys chan<- key) {
	buf := make([]byte, 32)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return
		}
		for i := 0; i < n; i++ {
			if buf[i] == 0x1b && i+2 < n && buf[i+1] == '[' {
				switch buf[i+2] {
				case 'A':
					keys <- keyUp
				case 'B':
					keys <- keyDown
				case 'C':
					keys <- keyRight
				case 'D':
					keys <- keyLeft
				}
				i += 2
				continue
			}
			switch buf[i] {
			case ' ':
				keys <- keyPause
			case 'r', 'R':
				keys <- keyRestart
			case 'q', 'Q', 3: // 3 = Ctrl-C, which raw mode no longer turns into a signal
				keys <- keyQuit
			}
		}
	}
}

func play(name string, synth *gore.MusSynth, sampleRate int, loop bool) error {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return err
	}
	<-ready

	player := ctx.NewPlayer(synth)
	// Keep the buffer small so seeking & the position display are responsive
	player.SetBufferSize(sampleRate * 4 / 10)
	player.Play()

	duration := synth.Duration()
	position := func() time.Duration {
		pos := synth.Position() - synth.OffsetToTime(int64(player.BufferedSize()))
		if pos < 0 && loop {
			pos += duration
		}
		return max(pos, 0)
	}
	seek := func(to time.Duration) error {
		to = max(0, min(to, duration))
		_, err := player.Seek(synth.TimeToOffset(to), io.SeekStart)
		return err
	}

	keys := make(chan key)
	interactive := term.IsTerminal(int(os.Stdin.Fd()))
	if interactive {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
		go readKeys(os.Stdin, keys)
		fmt.Print("←/→ seek 5s, ↑/↓ seek 30s, space pause, r restart, q quit\r\n")
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	paused := false
	for {
		select {
		case k := <-keys:
			var err error
			switch k {
			case keyLeft:
				err = seek(position() - 5*time.Second)
			case keyRight:
				err = seek(position() + 5*time.Second)
			case keyDown:
				err = seek(position() - 30*time.Second)
			case keyUp:
				err = seek(position() + 30*time.Second)
			case keyRestart:
				err = seek(0)
			case keyPause:
				paused = !paused
				if paused {
					player.Pause()
				} else {
					player.Play()
				}
			case keyQuit:
				fmt.Print("\r\n")
				return nil
			}
			if err != nil {
				return err
			}
		case <-ticker.C:
		}

		if err := player.Err(); err != nil {
			return err
		}
		if !paused && !player.IsPlaying() {
			// Reached the end of the song
			if interactive {
				fmt.Print("\r\n")
			}
			return nil
		}
		if interactive {
			state := "▶"
			if paused {
				state = "⏸"
			}
			fmt.Printf("\r%s %s  %s / %s\x1b[K", state, name, formatDuration(position()), formatDuration(duration))
		}
	}
}

func main() {
	wadFile := flag.String("wad", "", "Path to a WAD file to play music & take instruments from")
	track := flag.String("track", "", "Name of the music track within the WAD (eg: e1m1 or D_E1M1). Lists the tracks if not supplied")
	rawOutput := flag.String("raw", "", "Save raw audio (signed 16-bit stereo) instead of playing it")
	frequency := flag.Int("freq", 44100, "Sample frequency")
	loop := flag.Bool("loop", false, "Loop the song forever")
	volume := flag.Int("volume", 127, "Music volume, 0-127")
	opl3 := flag.Bool("opl3", false, "Emulate an OPL3 (18 voices, stereo) rather than an OPL2")
	flag.Parse()

	if *wadFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	lumps, genmidi, err := readWadMusic(*wadFile)
	if err != nil {
		log.Fatalf("Cannot read %s: %s", *wadFile, err)
	}

	if *track == "" {
		listTracks(lumps)
		return
	}
	lump := findTrack(lumps, *track)
	if lump == nil {
		fmt.Printf("Track %q not found in %s\n", *track, *wadFile)
		listTracks(lumps)
		os.Exit(1)
	}
	name := lump.Name

	m, err := gore.NewMusMusic(lump.Data)
	if err != nil {
		log.Fatalf("Cannot parse MUS track %s: %s", name, err)
	}
	fmt.Printf("%s: %s, %d events\n", name, formatDuration(m.Duration()), len(m.Events))

	synth, err := gore.NewMusSynth(m, genmidi, *frequency)
	if err != nil {
		log.Fatalf("Cannot create synthesizer: %s", err)
	}
	synth.SetOPL3(*opl3)
	synth.SetVolume(*volume)

	if *rawOutput != "" {
		output, err := os.Create(*rawOutput)
		if err != nil {
			log.Fatalf("Cannot create output file %s: %s", *rawOutput, err)
		}
		defer output.Close()
		if _, err := io.Copy(output, synth); err != nil {
			log.Fatalf("Cannot generate audio: %s", err)
		}
		return
	}

	synth.SetLoop(*loop)
	if err := play(name, synth, *frequency, *loop); err != nil && !errors.Is(err, io.EOF) {
		log.Fatalf("Playback failed: %s", err)
	}
}

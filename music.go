package gore

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"time"
)

// MusMusic represents the content of a MUS file.
// See https://doomwiki.org/wiki/MUS & https://moddingwiki.shikadi.net/wiki/MUS_Format
type MusMusic struct {
	Events []MusEvent // These are guaranteed to be in order of time
	Ticks  uint32     // Total length of the song, in ticks
}

type MusEventType byte

type MusEvent struct {
	Tick       uint32 // When does this event occur, in ticks (MusTicksPerSecond) from the start of the song
	Type       MusEventType
	Channel    byte
	Note       byte
	Controller byte
	Value      int // The note volume, pitch bend or controller value, depending on the event type
}

const (
	MusEventRelease MusEventType = iota
	MusEventPlay
	MusEventPitchBend
	MusEventSystem
	MusEventController
	MusEventEndOfMeasure
	MusEventFinish
	MusEventUnused
)

// MUS controller numbers, used by MusEventController & MusEventSystem.
// These are the only ones that DMX responds to.
const (
	MusControllerInstrument = 0
	MusControllerVolume     = 3
	MusSystemAllNotesOff    = 11
)

// MusTicksPerSecond is the playback rate of Doom MUS files
const MusTicksPerSecond = 140

// MusPercussionChannel is the MUS channel used for percussion (MIDI channel 10)
const MusPercussionChannel = 15

var (
	ErrInvalidMusFile = errors.New("invalid MUS file")

	musMagic = []byte{'M', 'U', 'S', 0x1a}
)

// IsMusData reports whether data looks like a MUS file
func IsMusData(data []byte) bool {
	return bytes.HasPrefix(data, musMagic)
}

func NewMusMusic(data []byte) (*MusMusic, error) {
	if len(data) < 16 || !IsMusData(data) {
		return nil, ErrInvalidMusFile
	}
	m := &MusMusic{}
	songLength := int(binary.LittleEndian.Uint16(data[4:6]))
	songOff := int(binary.LittleEndian.Uint16(data[6:8]))
	if songOff > len(data) {
		return nil, ErrInvalidMusFile
	}

	// Some tools write an incorrect song length, so be lenient and just
	// play to the end of the data
	song := data[songOff:]
	if songLength < len(song) {
		song = song[:songLength]
	}

	// Notes without a volume use the previous one played on their channel
	velocity := [16]int{}
	for i := range velocity {
		velocity[i] = 127
	}
	var err error
	pos := 0
	read := func() byte {
		if pos >= len(song) {
			err = ErrInvalidMusFile
			return 0
		}
		pos++
		return song[pos-1]
	}
	tick := uint32(0)
	for pos < len(song) && err == nil {
		desc := read()
		e := MusEvent{
			Tick:    tick,
			Type:    MusEventType((desc >> 4) & 0x7),
			Channel: desc & 0x0f,
		}
		switch e.Type {
		case MusEventRelease:
			e.Note = read() & 0x7f
		case MusEventPlay:
			b := read()
			e.Note = b & 0x7f
			if b&0x80 != 0 {
				velocity[e.Channel] = min(int(read()), 127)
			}
			e.Value = velocity[e.Channel]
		case MusEventPitchBend:
			e.Value = int(read())
		case MusEventSystem:
			e.Controller = read() & 0x7f
		case MusEventController:
			e.Controller = read() & 0x7f
			e.Value = min(int(read()), 127)
		case MusEventFinish:
			m.Ticks = tick
			return m, nil
		case MusEventUnused:
			read() // Has a single (meaningless) data byte
		}
		if e.Type != MusEventUnused {
			m.Events = append(m.Events, e)
		}

		if desc&0x80 != 0 {
			delay := uint32(0)
			for {
				b := read()
				delay = delay<<7 | uint32(b&0x7f)
				if b&0x80 == 0 {
					break
				}
			}
			tick += delay
		}
	}
	if err != nil {
		return nil, err
	}

	// No finish event, so the song ends at the end of the data
	m.Ticks = tick
	return m, nil
}

// Duration returns the total length of the song
func (m *MusMusic) Duration() time.Duration {
	return time.Duration(m.Ticks) * time.Second / MusTicksPerSecond
}

const (
	musBytesPerFrame = 4 // 16-bit stereo
	musMaxFrames     = 256
	musMaxTail       = 1500 * time.Millisecond // How long notes can ring on after the song ends
)

// MusSynth plays a MusMusic through an emulated OPL FM synthesizer, using
// the instruments from a GENMIDI lump, the same way Doom's DMX library did.
//
// It implements io.ReadSeeker, producing signed 16-bit little-endian stereo
// PCM. Offsets are measured in bytes of PCM output, so it can be handed
// straight to an audio player such as oto.
// It is safe to call its methods from multiple goroutines.
type MusSynth struct {
	mu         sync.Mutex
	music      *MusMusic
	sampleRate int
	loop       bool
	paused     bool
	opl        *oplDriver

	frame    int64 // current position in the song, in sample frames
	next     int   // index of the next event to process
	finished bool  // reached the end of the song, and not looping
	tail     int64 // frames rendered since the song finished
}

// NewMusSynth creates a synthesizer for the given song, producing samples
// at sampleRate. genmidi is the contents of the GENMIDI lump from a WAD.
func NewMusSynth(m *MusMusic, genmidi []byte, sampleRate int) (*MusSynth, error) {
	opl, err := newOPLDriver(genmidi, sampleRate)
	if err != nil {
		return nil, err
	}
	s := &MusSynth{music: m, sampleRate: sampleRate, opl: opl}
	s.seekFrame(0)
	return s, nil
}

// SetLoop controls whether the song restarts once it reaches the end
func (s *MusSynth) SetLoop(loop bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loop = loop
	if loop {
		s.finished = false
	}
}

// SetVolume sets the music volume, from 0 to 127 (the default). As with DMX,
// this limits the volume of each channel, rather than scaling the output.
func (s *MusSynth) SetVolume(volume int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.opl.setMusicVolume(max(0, min(127, volume)))
	if s.frame == 0 && s.next == 0 {
		// The song hasn't started yet, so start it at this volume
		s.seekFrame(0)
	}
}

// SetPaused pauses or resumes the song. As with DMX, pausing releases the
// notes being held, apart from percussion, and they're left to fade out.
func (s *MusSynth) SetPaused(paused bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if paused && !s.paused {
		s.opl.pause()
	}
	s.paused = paused
}

// Duration returns the length of the song
func (s *MusSynth) Duration() time.Duration {
	return s.music.Duration()
}

// Position returns how far through the song the synthesizer has rendered
func (s *MusSynth) Position() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.frameToTime(s.frame)
}

// TimeToOffset converts a time within the song to a byte offset suitable for Seek
func (s *MusSynth) TimeToOffset(t time.Duration) int64 {
	return int64(t) * int64(s.sampleRate) / int64(time.Second) * musBytesPerFrame
}

// OffsetToTime converts a byte offset of PCM output into a time
func (s *MusSynth) OffsetToTime(offset int64) time.Duration {
	return s.frameToTime(offset / musBytesPerFrame)
}

func (s *MusSynth) frameToTime(frame int64) time.Duration {
	return time.Duration(frame * int64(time.Second) / int64(s.sampleRate))
}

func (s *MusSynth) tickToFrame(tick uint32) int64 {
	return int64(tick) * int64(s.sampleRate) / MusTicksPerSecond
}

// Seek implements io.Seeker. The offset is in bytes of PCM output, and the
// resulting position is clamped to the length of the song.
func (s *MusSynth) Seek(offset int64, whence int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	end := s.tickToFrame(s.music.Ticks)
	frame := offset / musBytesPerFrame
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		frame += s.frame
	case io.SeekEnd:
		frame += end
	default:
		return 0, errors.New("MusSynth.Seek: invalid whence")
	}
	frame = max(0, min(frame, end))
	s.seekFrame(frame)
	return frame * musBytesPerFrame, nil
}

// seekFrame resets the synthesizer and fast-forwards to the given frame.
// Notes that would still be sounding are not restarted, but all the channel
// state (instruments, volumes, etc...) is brought up to date.
func (s *MusSynth) seekFrame(frame int64) {
	s.opl.reset()
	s.restart()
	for s.next < len(s.music.Events) && s.tickToFrame(s.music.Events[s.next].Tick) < frame {
		s.opl.event(&s.music.Events[s.next], false)
		s.next++
	}
	s.frame = frame
	s.finished = false
	s.tail = 0
}

// restart returns to the start of the song
func (s *MusSynth) restart() {
	s.opl.initChannels()
	s.frame = 0
	s.next = 0
}

// Read implements io.Reader, producing signed 16-bit little-endian stereo samples
func (s *MusSynth) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := len(p) / musBytesPerFrame
	if frames == 0 {
		return 0, io.ErrShortBuffer
	}
	if s.paused {
		s.opl.render(p, frames)
		return frames * musBytesPerFrame, nil
	}
	end := s.tickToFrame(s.music.Ticks)
	maxTail := int64(s.sampleRate) * int64(musMaxTail) / int64(time.Second)
	done := 0
	for done < frames {
		if !s.finished {
			for s.next < len(s.music.Events) && s.tickToFrame(s.music.Events[s.next].Tick) <= s.frame {
				s.opl.event(&s.music.Events[s.next], true)
				s.next++
			}
			if s.next >= len(s.music.Events) && s.frame >= end {
				if s.loop && end > 0 {
					s.restart()
					continue
				}
				// Let any notes still held fade out, rather than ringing forever
				s.opl.releaseAll()
				s.finished = true
			}
		}
		if s.finished && (s.opl.chip.silent() || s.tail >= maxTail) {
			break
		}

		chunk := min(frames-done, musMaxFrames)
		if !s.finished {
			target := end
			if s.next < len(s.music.Events) {
				target = s.tickToFrame(s.music.Events[s.next].Tick)
			}
			chunk = int(max(1, min(int64(chunk), target-s.frame)))
		}
		s.opl.render(p[done*musBytesPerFrame:], chunk)
		done += chunk
		if s.finished {
			s.tail += int64(chunk)
		} else {
			s.frame += int64(chunk)
		}
	}
	if done == 0 {
		return 0, io.EOF
	}
	return done * musBytesPerFrame, nil
}

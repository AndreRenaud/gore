package gore

// Doom's DMX OPL music driver, emulating the Doom 1.9 version of DMX.
// Ported from Chocolate Doom's i_oplmusic.c, but driven directly by MUS
// events rather than converting to MIDI first.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"slices"
)

const (
	genmidiHeader        = "#OPL_II#"
	genmidiNumInstrs     = 128
	genmidiNumPerc       = 47
	genmidiFlagFixed     = 0x0001 // fixed pitch
	genmidiFlag2Voice    = 0x0004 // double voice (OPL3)
	oplNumVoices         = 9
	oplNumOperators      = 21
	oplRegWaveformEnable = 0x01
	oplRegTimerCtrl      = 0x04
	oplRegFMMode         = 0x08
	oplRegNew            = 0x105
	oplRegsTremolo       = 0x20
	oplRegsLevel         = 0x40
	oplRegsAttack        = 0x60
	oplRegsSustain       = 0x80
	oplRegsWaveform      = 0xe0
	oplRegsFreq1         = 0xa0
	oplRegsFreq2         = 0xb0
	oplRegsFeedback      = 0xc0
)

var ErrInvalidGENMIDI = errors.New("invalid GENMIDI lump")

// GENMIDI lump structures, laid out to be read directly with encoding/binary

type genmidiOp struct {
	Tremolo, Attack, Sustain, Waveform, Scale, Level uint8
}

type genmidiVoice struct {
	Modulator      genmidiOp
	Feedback       uint8
	Carrier        genmidiOp
	_              uint8
	BaseNoteOffset int16
}

type genmidiInstr struct {
	Flags      uint16
	FineTuning uint8
	FixedNote  uint8
	Voices     [2]genmidiVoice
}

type oplChannel struct {
	num        int // The MUS channel number, which is also its priority
	instrument *genmidiInstr
	volume     int
	pan        int
	bend       int
}

type oplVoice struct {
	index    int // Index of this voice
	op1, op2 int // The operators used by this voice
	array    int // Register array (0x100 for the second half of an OPL3)

	instr      *genmidiInstr // Currently-loaded instrument data
	instrVoice int           // Which voice of a double voice instrument

	channel    *oplChannel // The channel currently using this voice, or nil
	key        int         // The MUS key that this voice is playing
	note       int         // The note being played. This is normally the same as the key, but not for fixed pitch instruments
	freq       int         // The frequency value being used
	noteVolume int         // The volume of the note being played on this channel
	carVolume  int         // The current volume (register value) that has been set for this channel
	modVolume  int
	regPan     int
}

// The OPL frequency value for each 1/32 of a semitone, starting from A#4 in
// block 5. DMX's tuning is very slightly flat of A440, so the base value is
// fitted to reproduce DMX's original table, rather than calculated.
var oplFrequencyCurve = func() (curve [668]uint16) {
	for i := range curve {
		curve[i] = uint16(307.2455 * math.Exp2(float64(i)/384))
	}
	// The last value in DMX's table is actually a buffer overrun and does
	// not fit with the other values.
	curve[len(curve)-1] = 0x36c
	return
}()

var oplVoiceOperators = [2][oplNumVoices]int{
	{0x00, 0x01, 0x02, 0x08, 0x09, 0x0a, 0x10, 0x11, 0x12},
	{0x03, 0x04, 0x05, 0x0b, 0x0c, 0x0d, 0x13, 0x14, 0x15},
}

type oplDriver struct {
	chip       opl3Chip
	sampleRate int
	opl3       bool
	instrs     []genmidiInstr
	voices     [oplNumVoices * 2]oplVoice
	free       []*oplVoice
	alloced    []*oplVoice
	channels   [16]oplChannel
}

func newOPLDriver(genmidi []byte, sampleRate int) (*oplDriver, error) {
	if !bytes.HasPrefix(genmidi, []byte(genmidiHeader)) {
		return nil, ErrInvalidGENMIDI
	}
	d := &oplDriver{
		sampleRate: sampleRate,
		instrs:     make([]genmidiInstr, genmidiNumInstrs+genmidiNumPerc),
	}
	r := bytes.NewReader(genmidi[len(genmidiHeader):])
	if err := binary.Read(r, binary.LittleEndian, d.instrs); err != nil {
		return nil, ErrInvalidGENMIDI
	}
	return d, nil
}

func (d *oplDriver) write(reg, value int) {
	d.chip.writeRegBuffered(uint16(reg), uint8(value))
}

// reset initializes the chip and frees all voices
func (d *oplDriver) reset() {
	d.chip.reset(d.sampleRate)

	initArray := func(array int) {
		for r := oplRegsLevel; r <= oplRegsLevel+oplNumOperators; r++ {
			d.write(r|array, 0x3f)
		}
		// These write to registers that actually don't exist, but this is what Doom does ...
		for r := oplRegsAttack; r <= oplRegsWaveform+oplNumOperators; r++ {
			d.write(r|array, 0)
		}
		for r := 1; r < oplRegsLevel; r++ {
			d.write(r|array, 0)
		}
	}
	initArray(0)
	// Reset both timers and enable interrupts
	d.write(oplRegTimerCtrl, 0x60)
	d.write(oplRegTimerCtrl, 0x80)
	// "Allow FM chips to control the waveform of each operator"
	d.write(oplRegWaveformEnable, 0x20)
	if d.opl3 {
		d.write(oplRegNew, 0x01)
		initArray(0x100)
	}
	// Keyboard split point on (?)
	d.write(oplRegFMMode, 0x40)
	if d.opl3 {
		d.write(oplRegNew, 0x01)
	}

	numVoices := oplNumVoices
	if d.opl3 {
		numVoices *= 2
	}
	d.free = d.free[:0]
	d.alloced = d.alloced[:0]
	for i := range numVoices {
		d.voices[i] = oplVoice{
			index: i % oplNumVoices,
			op1:   oplVoiceOperators[0][i%oplNumVoices],
			op2:   oplVoiceOperators[1][i%oplNumVoices],
			array: (i / oplNumVoices) << 8,
		}
		d.free = append(d.free, &d.voices[i])
	}
}

func (d *oplDriver) initChannels() {
	for i := range d.channels {
		d.channels[i] = oplChannel{num: i, instrument: &d.instrs[0], volume: 100, pan: 0x30}
	}
}

// getFreeVoice takes the next available voice from the freelist
func (d *oplDriver) getFreeVoice() *oplVoice {
	if len(d.free) == 0 {
		return nil
	}
	voice := d.free[0]
	d.free = slices.Delete(d.free, 0, 1)
	d.alloced = append(d.alloced, voice)
	return voice
}

// releaseVoice releases an allocated voice back to the freelist
func (d *oplDriver) releaseVoice(index int) {
	voice := d.alloced[index]
	d.voiceKeyOff(voice)
	voice.channel = nil
	voice.note = 0
	d.alloced = slices.Delete(d.alloced, index, index+1)
	// Add to the end of the freelist (This is how Doom behaves!)
	d.free = append(d.free, voice)
}

// loadOperatorData loads data to the specified operator, returning the level register value
func (d *oplDriver) loadOperatorData(op int, data *genmidiOp, maxLevel bool) int {
	// The scale and level fields must be combined for the level register.
	// For the carrier wave we always set the maximum level.
	level := int(data.Scale)
	if maxLevel {
		level |= 0x3f
	} else {
		level |= int(data.Level)
	}
	d.write(oplRegsLevel+op, level)
	d.write(oplRegsTremolo+op, int(data.Tremolo))
	d.write(oplRegsAttack+op, int(data.Attack))
	d.write(oplRegsSustain+op, int(data.Sustain))
	d.write(oplRegsWaveform+op, int(data.Waveform))
	return level
}

func (d *oplDriver) setVoiceInstrument(voice *oplVoice, instr *genmidiInstr, instrVoice int) {
	// Instrument already set for this channel?
	if voice.instr == instr && voice.instrVoice == instrVoice {
		return
	}
	voice.instr = instr
	voice.instrVoice = instrVoice
	data := &instr.Voices[instrVoice]

	// Doom loads the second operator first, then the first.
	// The carrier is set to minimum volume until the voice volume
	// is set in setVoiceVolume (below). If we are not using
	// modulating mode, we must set both to minimum volume.
	modulating := data.Feedback&0x01 == 0
	voice.carVolume = d.loadOperatorData(voice.op2|voice.array, &data.Carrier, true)
	voice.modVolume = d.loadOperatorData(voice.op1|voice.array, &data.Modulator, !modulating)

	// Set feedback register that control the connection between the two operators
	d.write((oplRegsFeedback+voice.index)|voice.array, int(data.Feedback)|voice.regPan)
}

func (d *oplDriver) setVoiceVolume(voice *oplVoice, volume int) {
	voice.noteVolume = volume
	data := &voice.instr.Voices[voice.instrVoice]

	// Multiply note volume and channel volume to get the actual volume.
	midiVolume := 2 * (oplVolumeMappingTable[voice.channel.volume] + 1)
	fullVolume := (oplVolumeMappingTable[volume] * midiVolume) >> 9

	// The volume value to use in the register:
	carVolume := 0x3f - fullVolume
	if carVolume == voice.carVolume&0x3f {
		return
	}
	voice.carVolume = carVolume | (voice.carVolume & 0xc0)
	d.write((oplRegsLevel+voice.op2)|voice.array, voice.carVolume)

	// If we are using non-modulated feedback mode, we must set the
	// volume for both voices.
	if data.Feedback&0x01 != 0 && data.Modulator.Level != 0x3f {
		modVolume := max(int(data.Modulator.Level), carVolume) | (voice.modVolume & 0xc0)
		if modVolume != voice.modVolume {
			voice.modVolume = modVolume
			d.write((oplRegsLevel+voice.op1)|voice.array, modVolume|int(data.Modulator.Scale&0xc0))
		}
	}
}

func (d *oplDriver) voiceKeyOff(voice *oplVoice) {
	d.write((oplRegsFreq2+voice.index)|voice.array, voice.freq>>8)
}

func (d *oplDriver) keyOff(channel *oplChannel, key int) {
	// Turn off voices being used to play this key.
	// If it is a double voice instrument there will be two.
	for i := 0; i < len(d.alloced); i++ {
		if d.alloced[i].channel == channel && d.alloced[i].key == key {
			d.releaseVoice(i)
			i--
		}
	}
}

// replaceExistingVoice discards an existing voice to make room for a new
// note when all voices are in use. Second voices of double voice
// instruments are non-essential, and higher numbered channels have a lower
// priority than lower numbered ones.
func (d *oplDriver) replaceExistingVoice() {
	result := 0
	for i, voice := range d.alloced {
		if voice.instrVoice != 0 || voice.channel.num >= d.alloced[result].channel.num {
			result = i
		}
	}
	d.releaseVoice(result)
}

func (d *oplDriver) frequencyForVoice(voice *oplVoice) int {
	note := voice.note

	// Apply note offset. Don't apply offset if the instrument is a fixed note instrument.
	if voice.instr.Flags&genmidiFlagFixed == 0 {
		note += int(voice.instr.Voices[voice.instrVoice].BaseNoteOffset)
	}

	// Avoid possible overflow due to base note offset:
	for note < 0 {
		note += 12
	}
	for note > 95 {
		note -= 12
	}

	freqIndex := 64 + 32*note + voice.channel.bend

	// If this is the second voice of a double voice instrument, the
	// frequency index can be adjusted by the fine tuning field.
	if voice.instrVoice != 0 {
		freqIndex += int(voice.instr.FineTuning)/2 - 64
	}
	freqIndex = max(freqIndex, 0)

	// The first 7 notes use the start of the table, while
	// consecutive notes loop around the latter part.
	if freqIndex < 284 {
		return int(oplFrequencyCurve[freqIndex])
	}
	subIndex := (freqIndex - 284) % (12 * 32)
	octave := min((freqIndex-284)/(12*32), 7)
	return int(oplFrequencyCurve[subIndex+284]) | octave<<10
}

func (d *oplDriver) updateVoiceFrequency(voice *oplVoice) {
	if freq := d.frequencyForVoice(voice); voice.freq != freq {
		d.write((oplRegsFreq1+voice.index)|voice.array, freq&0xff)
		d.write((oplRegsFreq2+voice.index)|voice.array, freq>>8|0x20)
		voice.freq = freq
	}
}

// voiceKeyOn programs a single voice for an instrument. For a double voice
// instrument (genmidiFlag2Voice), this is called twice for each key on event.
func (d *oplDriver) voiceKeyOn(channel *oplChannel, instr *genmidiInstr, instrVoice, note, key, volume int) {
	voice := d.getFreeVoice()
	if voice == nil {
		return
	}
	voice.channel = channel
	voice.key = key

	// Work out the note to use. This is normally the same as
	// the key, unless it is a fixed pitch instrument.
	voice.note = note
	if instr.Flags&genmidiFlagFixed != 0 {
		voice.note = int(instr.FixedNote)
	}
	voice.regPan = channel.pan

	d.setVoiceInstrument(voice, instr, instrVoice)
	d.setVoiceVolume(voice, volume)

	// Write the frequency value to turn the note on.
	voice.freq = 0
	d.updateVoiceFrequency(voice)
}

func (d *oplDriver) keyOn(channel *oplChannel, key, volume int) {
	// A volume of zero means key off.
	if volume <= 0 {
		d.keyOff(channel, key)
		return
	}
	note := key
	instr := channel.instrument
	// Percussion channel is treated differently.
	if channel.num == MusPercussionChannel {
		if key < 35 || key > 81 {
			return
		}
		instr = &d.instrs[genmidiNumInstrs+key-35]
		note = 60
	}

	if len(d.free) == 0 {
		d.replaceExistingVoice()
	}

	// Find and program a voice for this instrument. If this
	// is a double voice instrument, we must do this twice.
	d.voiceKeyOn(channel, instr, 0, note, key, volume)
	if instr.Flags&genmidiFlag2Voice != 0 {
		d.voiceKeyOn(channel, instr, 1, note, key, volume)
	}
}

func (d *oplDriver) setChannelVolume(channel *oplChannel, volume int) {
	channel.volume = volume
	for i := range d.voices {
		if d.voices[i].channel == channel {
			d.setVoiceVolume(&d.voices[i], d.voices[i].noteVolume)
		}
	}
}

func (d *oplDriver) setChannelPan(channel *oplChannel, pan int) {
	if !d.opl3 {
		return
	}
	// The DMX library has the stereo channels backwards, maybe because
	// Paul Radek had a Soundblaster card with the channels reversed, or
	// perhaps it was just a bug in the OPL3 support that was never
	// finished. We preserve this bug.
	regPan := 0x30
	if pan >= 96 {
		regPan = 0x10
	} else if pan <= 48 {
		regPan = 0x20
	}
	if channel.pan == regPan {
		return
	}
	channel.pan = regPan
	for i := range d.voices {
		if voice := &d.voices[i]; voice.channel == channel {
			voice.regPan = regPan
			d.write((oplRegsFeedback+voice.index)|voice.array, int(voice.instr.Voices[voice.instrVoice].Feedback)|regPan)
		}
	}
}

func (d *oplDriver) allNotesOff(channel *oplChannel) {
	for i := 0; i < len(d.alloced); i++ {
		if d.alloced[i].channel == channel {
			d.releaseVoice(i)
			i--
		}
	}
}

func (d *oplDriver) pitchBend(channel *oplChannel, bend int) {
	channel.bend = bend

	// Update all voices for this channel, which DMX also moves to the end
	// of the allocated list.
	var updated, others []*oplVoice
	for _, voice := range d.alloced {
		if voice.channel == channel {
			d.updateVoiceFrequency(voice)
			updated = append(updated, voice)
		} else {
			others = append(others, voice)
		}
	}
	d.alloced = append(others, updated...)
}

// event processes a MUS event. velocity is the note volume to use for
// MusEventPlay. When live is false the song is being fast-forwarded, so
// only channel state is updated and no notes are played.
func (d *oplDriver) event(e *MusEvent, velocity int, live bool) {
	channel := &d.channels[e.Channel]
	switch e.Type {
	case MusEventPlay:
		if live {
			d.keyOn(channel, int(e.Note), velocity)
		}
	case MusEventRelease:
		d.keyOff(channel, int(e.Note))
	case MusEventPitchBend:
		// Only the MSB of the MIDI pitch bend value is considered: this is what Doom does.
		d.pitchBend(channel, e.Value>>1-64)
	case MusEventController:
		switch e.Controller {
		case MusControllerInstrument:
			channel.instrument = &d.instrs[e.Value]
		case MusControllerVolume:
			d.setChannelVolume(channel, e.Value)
		case MusControllerPan:
			d.setChannelPan(channel, e.Value)
		}
	case MusEventSystem:
		if e.Controller == MusSystemAllNotesOff {
			d.allNotesOff(channel)
		}
	}
}

func (d *oplDriver) releaseAll() {
	for i := range d.channels {
		d.allNotesOff(&d.channels[i])
	}
}

// active reports whether any notes are still sounding
func (d *oplDriver) active() bool {
	return len(d.alloced) > 0 || !d.chip.silent()
}

// render produces frames of signed 16-bit little-endian stereo samples, scaled by gain
func (d *oplDriver) render(out []byte, frames int, gain float64) {
	var buf [2]int16
	for i := range frames {
		d.chip.generateResampled(&buf)
		for c, v := range buf {
			if gain != 1 {
				v = int16(max(-32768, min(32767, float64(v)*gain)))
			}
			binary.LittleEndian.PutUint16(out[i*4+c*2:], uint16(v))
		}
	}
}

// Mapping from MIDI volume level to OPL level value.
var oplVolumeMappingTable = [128]int{
	0, 1, 3, 5, 6, 8, 10, 11,
	13, 14, 16, 17, 19, 20, 22, 23,
	25, 26, 27, 29, 30, 32, 33, 34,
	36, 37, 39, 41, 43, 45, 47, 49,
	50, 52, 54, 55, 57, 59, 60, 61,
	63, 64, 66, 67, 68, 69, 71, 72,
	73, 74, 75, 76, 77, 79, 80, 81,
	82, 83, 84, 84, 85, 86, 87, 88,
	89, 90, 91, 92, 92, 93, 94, 95,
	96, 96, 97, 98, 99, 99, 100, 101,
	101, 102, 103, 103, 104, 105, 105, 106,
	107, 107, 108, 109, 109, 110, 110, 111,
	112, 112, 113, 113, 114, 114, 115, 115,
	116, 117, 117, 118, 118, 119, 119, 120,
	120, 121, 121, 122, 122, 123, 123, 123,
	124, 124, 125, 125, 126, 126, 127, 127,
}

package gore

import "math"

// Go port of the Nuked OPL3 emulator, version 1.8.
// See https://github.com/nukeykt/Nuked-OPL3
//
// Copyright (C) 2013-2020 Nuke.YKT
//
// Nuked OPL3 is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as
// published by the Free Software Foundation, either version 2.1
// of the License, or (at your option) any later version.
//
// Nuked OPL3 is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// Only the features used by Doom's DMX music driver are included: the 9
// 2-operator voices of OPL2 mode. The OPL3 extensions, 4-operator voices and
// rhythm mode are not emulated. For those register writes, the output is
// identical to the original.

const (
	opl3WriteBufSize  = 1024
	opl3WriteBufDelay = 2
	opl3RsmFrac       = 10
	opl3NativeRate    = 49716
	opl3TremoloShift  = 4 // Register 0xbd, which DMX never sets
	opl3VibShift      = 1
)

// Envelope generator states
const (
	opl3EgAttack = iota
	opl3EgDecay
	opl3EgSustain
	opl3EgRelease
)

// freq mult table multiplied by 2: 1/2, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 12, 12, 15, 15
var opl3MT = [16]uint32{1, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 20, 24, 24, 30, 30}

// The chip's logarithmic sine and exponent ROMs, which follow these formulas exactly
var opl3LogSinROM, opl3ExpROM = func() (logsin, exp [256]uint16) {
	for i := range 256 {
		logsin[i] = uint16(math.Round(-math.Log2(math.Sin((float64(i)+0.5)*math.Pi/512)) * 256))
		exp[i] = uint16(math.Round(math.Exp2(float64(255-i)/256) * 1024))
	}
	return
}()

var opl3KSLROM = [16]int16{0, 32, 40, 45, 48, 51, 53, 55, 56, 58, 59, 60, 61, 62, 63, 64}

var opl3KSLShift = [4]uint8{8, 1, 2, 0}

var opl3EgIncStep = [4][4]uint8{
	{0, 0, 0, 0},
	{1, 0, 0, 0},
	{1, 0, 1, 0},
	{1, 1, 1, 0},
}

type opl3Slot struct {
	chip       *opl3Chip
	channel    *opl3Channel
	out        int16
	fbmod      int16
	prout      int16
	mod        *int16 // Phase modulation input
	egRout     uint16
	egOut      uint16
	egGen      uint8
	egKsl      uint8
	trem       bool
	vib        uint8
	egType     uint8
	ksr        uint8
	mult       uint8
	ksl        uint8
	tl         uint8
	ar         uint8
	dr         uint8
	sl         uint8
	rr         uint8
	wf         uint8
	key        uint8
	pgReset    bool
	pgPhase    uint32
	pgPhaseOut uint16
}

type opl3Channel struct {
	slots [2]*opl3Slot
	fNum  uint16
	block uint8
	fb    uint8
	con   uint8
	ksv   uint8
}

type opl3WriteBuf struct {
	time      uint64
	reg, data uint8
	pending   bool
}

type opl3Chip struct {
	channel    [9]opl3Channel
	slot       [18]opl3Slot
	timer      uint16
	egTimer    uint64
	egState    uint8
	egAdd      uint8
	egTimerLo  uint8
	nts        uint8
	vibpos     uint8
	tremolo    uint8
	tremolopos uint8
	zeromod    int16
	mixbuff    [2]int32

	rateratio  int32
	samplecnt  int32
	oldsamples [2]int16
	samples    [2]int16

	writebufSamplecnt uint64
	writebufCur       uint32
	writebufLast      uint32
	writebufLasttime  uint64
	writebuf          [opl3WriteBufSize]opl3WriteBuf
}

func opl3EnvelopeCalcExp(level uint32) int16 {
	level = min(level, 0x1fff)
	return int16((int32(opl3ExpROM[level&0xff]) << 1) >> (level >> 8))
}

func opl3LogSin(phase uint16) uint16 {
	if phase&0x100 != 0 {
		return opl3LogSinROM[(phase&0xff)^0xff]
	}
	return opl3LogSinROM[phase&0xff]
}

// opl3Waveform calculates the output of one of the 4 OPL2 waveforms
func opl3Waveform(wf uint8, phase uint16, envelope uint16) int16 {
	var out uint16
	neg := false
	phase &= 0x3ff
	switch wf {
	case 0: // Sine
		neg = phase&0x200 != 0
		out = opl3LogSin(phase)
	case 1: // Half sine
		if phase&0x200 != 0 {
			out = 0x1000
		} else {
			out = opl3LogSin(phase)
		}
	case 2: // Absolute sine
		out = opl3LogSin(phase)
	case 3: // Quarter sine
		if phase&0x100 != 0 {
			out = 0x1000
		} else {
			out = opl3LogSinROM[phase&0xff]
		}
	}
	v := opl3EnvelopeCalcExp(uint32(out) + uint32(envelope)<<3)
	if neg {
		return ^v
	}
	return v
}

func (slot *opl3Slot) updateKSL() {
	ksl := opl3KSLROM[slot.channel.fNum>>6]<<2 - int16(0x08-slot.channel.block)<<5
	slot.egKsl = uint8(max(ksl, 0))
}

func (slot *opl3Slot) envelopeCalc() {
	chip := slot.chip
	slot.egOut = slot.egRout + uint16(slot.tl)<<2 + uint16(slot.egKsl>>opl3KSLShift[slot.ksl])
	if slot.trem {
		slot.egOut += uint16(chip.tremolo)
	}
	var regRate uint8
	reset := slot.key != 0 && slot.egGen == opl3EgRelease
	switch {
	case reset, slot.egGen == opl3EgAttack:
		regRate = slot.ar
	case slot.egGen == opl3EgDecay:
		regRate = slot.dr
	case slot.egGen == opl3EgRelease, slot.egType == 0:
		regRate = slot.rr
	}
	slot.pgReset = reset
	ks := slot.channel.ksv >> ((slot.ksr ^ 1) << 1)
	rate := ks + regRate<<2
	rateHi := min(rate>>2, 0x0f)
	rateLo := rate & 0x03
	shift := uint8(0)
	if regRate != 0 {
		if rateHi < 12 {
			if chip.egState != 0 {
				switch rateHi + chip.egAdd {
				case 12:
					shift = 1
				case 13:
					shift = (rateLo >> 1) & 0x01
				case 14:
					shift = rateLo & 0x01
				}
			}
		} else {
			shift = (rateHi & 0x03) + opl3EgIncStep[rateLo][chip.egTimerLo]
			if shift&0x04 != 0 {
				shift = 0x03
			}
			if shift == 0 {
				shift = chip.egState
			}
		}
	}
	egRout := slot.egRout
	egInc := int32(0)
	egOff := slot.egRout&0x1f8 == 0x1f8
	if reset && rateHi == 0x0f { // Instant attack
		egRout = 0
	}
	if slot.egGen != opl3EgAttack && !reset && egOff {
		egRout = 0x1ff
	}
	switch slot.egGen {
	case opl3EgAttack:
		if slot.egRout == 0 {
			slot.egGen = opl3EgDecay
		} else if slot.key != 0 && shift > 0 && rateHi != 0x0f {
			egInc = ^int32(slot.egRout) >> (4 - shift)
		}
	case opl3EgDecay:
		if slot.egRout>>4 == uint16(slot.sl) {
			slot.egGen = opl3EgSustain
		} else if !egOff && !reset && shift > 0 {
			egInc = 1 << (shift - 1)
		}
	default:
		if !egOff && !reset && shift > 0 {
			egInc = 1 << (shift - 1)
		}
	}
	slot.egRout = uint16(int32(egRout)+egInc) & 0x1ff
	if reset {
		slot.egGen = opl3EgAttack
	}
	if slot.key == 0 {
		slot.egGen = opl3EgRelease
	}
}

func (slot *opl3Slot) phaseGenerate() {
	fNum := slot.channel.fNum
	if slot.vib != 0 {
		rng := int32(fNum>>7) & 7
		vibpos := slot.chip.vibpos
		if vibpos&3 == 0 {
			rng = 0
		} else if vibpos&1 != 0 {
			rng >>= 1
		}
		rng >>= opl3VibShift
		if vibpos&4 != 0 {
			rng = -rng
		}
		fNum = uint16(int32(fNum) + rng)
	}
	basefreq := (uint32(fNum) << slot.channel.block) >> 1
	slot.pgPhaseOut = uint16(slot.pgPhase >> 9)
	if slot.pgReset {
		slot.pgPhase = 0
	}
	slot.pgPhase += (basefreq * opl3MT[slot.mult]) >> 1
}

func (slot *opl3Slot) process() {
	// Feedback
	if fb := slot.channel.fb; fb != 0 {
		slot.fbmod = int16((int32(slot.prout) + int32(slot.out)) >> (9 - fb))
	} else {
		slot.fbmod = 0
	}
	slot.prout = slot.out
	slot.envelopeCalc()
	slot.phaseGenerate()
	slot.out = opl3Waveform(slot.wf, uint16(int32(slot.pgPhaseOut)+int32(*slot.mod)), slot.egOut)
}

func (channel *opl3Channel) updateFrequency(nts uint8) {
	channel.ksv = channel.block<<1 | uint8(channel.fNum>>(9-nts))&0x01
	channel.slots[0].updateKSL()
	channel.slots[1].updateKSL()
}

// output returns the channel's current sample: either the carrier alone
// (FM), or the sum of both operators (additive)
func (channel *opl3Channel) output() int32 {
	out := channel.slots[1].out
	if channel.con != 0 {
		out += channel.slots[0].out
	}
	return int32(out)
}

func (chip *opl3Chip) mix() int32 {
	total := int32(0)
	for i := range chip.channel {
		total += chip.channel[i].output()
	}
	return total
}

func opl3ClipSample(sample int32) int16 {
	return int16(max(-32768, min(32767, sample)))
}

// generate produces one stereo sample at the chip's native rate. Channels
// 6-8 are output one sample later on the left side than the right, which is
// emulated by the order slots are processed and mixed.
func (chip *opl3Chip) generate(buf *[2]int16) {
	process := func(slots []opl3Slot) {
		for i := range slots {
			slots[i].process()
		}
	}
	buf[1] = opl3ClipSample(chip.mixbuff[1])
	process(chip.slot[:15])
	chip.mixbuff[0] = chip.mix()
	process(chip.slot[15:])
	buf[0] = opl3ClipSample(chip.mixbuff[0])
	chip.mixbuff[1] = chip.mix()

	if chip.timer&0x3f == 0x3f {
		chip.tremolopos = (chip.tremolopos + 1) % 210
	}
	if chip.tremolopos < 105 {
		chip.tremolo = chip.tremolopos >> opl3TremoloShift
	} else {
		chip.tremolo = (210 - chip.tremolopos) >> opl3TremoloShift
	}
	if chip.timer&0x3ff == 0x3ff {
		chip.vibpos = (chip.vibpos + 1) & 7
	}
	chip.timer++

	if chip.egState != 0 {
		shift := uint8(0)
		for shift < 13 && (chip.egTimer>>shift)&1 == 0 {
			shift++
		}
		if shift > 12 {
			chip.egAdd = 0
		} else {
			chip.egAdd = shift + 1
		}
		chip.egTimerLo = uint8(chip.egTimer & 0x3)
	}
	// The real timer is 36 bits, but only the bottom 13 are used
	if chip.egState != 0 {
		chip.egTimer++
	}
	chip.egState ^= 1

	for {
		wb := &chip.writebuf[chip.writebufCur]
		if wb.time > chip.writebufSamplecnt || !wb.pending {
			break
		}
		wb.pending = false
		chip.writeReg(wb.reg, wb.data)
		chip.writebufCur = (chip.writebufCur + 1) % opl3WriteBufSize
	}
	chip.writebufSamplecnt++
}

// generateResampled produces one stereo sample at the rate passed to reset
func (chip *opl3Chip) generateResampled(buf *[2]int16) {
	for chip.samplecnt >= chip.rateratio {
		chip.oldsamples = chip.samples
		chip.generate(&chip.samples)
		chip.samplecnt -= chip.rateratio
	}
	for i := range buf {
		buf[i] = int16((int32(chip.oldsamples[i])*(chip.rateratio-chip.samplecnt) +
			int32(chip.samples[i])*chip.samplecnt) / chip.rateratio)
	}
	chip.samplecnt += 1 << opl3RsmFrac
}

func (chip *opl3Chip) reset(sampleRate int) {
	*chip = opl3Chip{}
	for i := range chip.channel {
		channel := &chip.channel[i]
		for j := range channel.slots {
			// Channels' slots are in groups of 3, with the second slot 3 after the first
			slot := &chip.slot[i/3*6+i%3+j*3]
			slot.chip = chip
			slot.channel = channel
			slot.egRout = 0x1ff
			slot.egOut = 0x1ff
			slot.egGen = opl3EgRelease
			channel.slots[j] = slot
		}
		channel.slots[0].mod = &channel.slots[0].fbmod
		channel.slots[1].mod = &channel.slots[0].out
	}
	chip.rateratio = int32((sampleRate << opl3RsmFrac) / opl3NativeRate)
}

func (chip *opl3Chip) writeReg(reg uint8, v uint8) {
	switch {
	case reg == 0x08:
		chip.nts = (v >> 6) & 0x01
	case (reg >= 0x20 && reg < 0xa0) || reg >= 0xe0:
		// Slot registers are in groups of 6, 8 registers apart
		offset := reg & 0x1f
		if offset%8 >= 6 || offset >= 0x18 {
			return
		}
		slot := &chip.slot[offset/8*6+offset%8]
		switch reg & 0xe0 {
		case 0x20:
			slot.trem = v&0x80 != 0
			slot.vib = (v >> 6) & 0x01
			slot.egType = (v >> 5) & 0x01
			slot.ksr = (v >> 4) & 0x01
			slot.mult = v & 0x0f
		case 0x40:
			slot.ksl = (v >> 6) & 0x03
			slot.tl = v & 0x3f
			slot.updateKSL()
		case 0x60:
			slot.ar = (v >> 4) & 0x0f
			slot.dr = v & 0x0f
		case 0x80:
			slot.sl = (v >> 4) & 0x0f
			if slot.sl == 0x0f {
				slot.sl = 0x1f
			}
			slot.rr = v & 0x0f
		case 0xe0:
			slot.wf = v & 0x03
		}
	case reg&0x0f < 9:
		channel := &chip.channel[reg&0x0f]
		switch reg & 0xf0 {
		case 0xa0:
			channel.fNum = (channel.fNum & 0x300) | uint16(v)
			channel.updateFrequency(chip.nts)
		case 0xb0:
			channel.fNum = (channel.fNum & 0xff) | uint16(v&0x03)<<8
			channel.block = (v >> 2) & 0x07
			channel.updateFrequency(chip.nts)
			channel.slots[0].key = (v >> 5) & 0x01
			channel.slots[1].key = (v >> 5) & 0x01
		case 0xc0:
			channel.fb = (v & 0x0e) >> 1
			channel.con = v & 0x01
			if channel.con == 0 {
				channel.slots[1].mod = &channel.slots[0].out
			} else {
				channel.slots[1].mod = &chip.zeromod
			}
		}
	}
}

// writeRegBuffered queues a register write, spacing writes out in time the
// way the real chip requires. This matters when a key off is immediately
// followed by a key on, which would otherwise not re-trigger the envelope.
func (chip *opl3Chip) writeRegBuffered(reg uint8, v uint8) {
	last := chip.writebufLast
	wb := &chip.writebuf[last]
	if wb.pending {
		chip.writeReg(wb.reg, wb.data)
		chip.writebufCur = (last + 1) % opl3WriteBufSize
		chip.writebufSamplecnt = wb.time
	}
	wb.reg = reg
	wb.data = v
	wb.pending = true
	wb.time = max(chip.writebufLasttime+opl3WriteBufDelay, chip.writebufSamplecnt)
	chip.writebufLasttime = wb.time
	chip.writebufLast = (last + 1) % opl3WriteBufSize
}

// silent reports whether every operator's envelope has fully decayed
func (chip *opl3Chip) silent() bool {
	for i := range chip.slot {
		if chip.slot[i].egRout < 0x1f8 {
			return false
		}
	}
	return true
}

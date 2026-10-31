package gore

// Sound for the step API.
//
// Vanilla DOOM hands sounds to a platform "sound module" and asks it which
// channels are still playing, so it can free them and move positional
// sounds as their sources move. Without a module nothing ever finishes
// playing, so after snd_channels sounds every channel stays allocated and
// new sounds are dropped unless they outrank one already playing.
//
// In step mode this module plays nothing. It records what should be heard
// as a queue of SoundEvents, and works out how long each sound lasts from
// the sample's length and the tic counter, so the sound system stays
// deterministic along with everything else.

import (
	"encoding/binary"
	"strings"
)

// SoundEvent is something the player should hear.
type SoundEvent struct {
	// Tic is the tic whose simulation produced the event (the value of
	// Stats.Tic once that tic has run).
	Tic int32 `json:"tic"`
	// Kind is "start", "update" (volume/separation of a playing sound
	// changed, e.g. because its source moved) or "stop".
	Kind    string `json:"kind"`
	Channel int32  `json:"ch"`
	// Name is the sound effect, e.g. "pistol"; its samples are in the WAD
	// lump "DS" + upper(Name). Set for "start" only.
	Name string `json:"name,omitempty"`
	// Volume is 0-127. Sep is the stereo separation, 0 (left) to 254
	// (right), with 128 in the centre.
	Volume int32 `json:"vol,omitempty"`
	Sep    int32 `json:"sep,omitempty"`
}

// maxSoundEvents bounds the queue for hosts that never drain it.
const maxSoundEvents = 512

var (
	stepSoundEvents []SoundEvent
	stepSoundTic    int32
	stepChannelEnd  []int32              // per channel: tic at which the sound finishes
	stepSoundTics   = map[string]int32{} // sample name -> duration in tics
)

// StepSoundEvents returns the sound events since the last call.
func StepSoundEvents() []SoundEvent {
	ev := stepSoundEvents
	stepSoundEvents = nil
	return ev
}

func emitSound(ev SoundEvent) {
	ev.Tic = stepSoundTic
	if len(stepSoundEvents) >= maxSoundEvents {
		stepSoundEvents = stepSoundEvents[1:]
	}
	stepSoundEvents = append(stepSoundEvents, ev)
}

// sampleName returns the name of the sample a sound effect plays. Some
// effects are variations of another (e.g. the chaingun is the pistol at a
// higher pitch) and have no samples of their own.
func sampleName(sfx *sfxinfo_t) string {
	if sfx.Flink != nil {
		return sfx.Flink.Fname
	}
	return sfx.Fname
}

// soundTics returns how many tics a sample plays for, from its DMX header:
// format (u16), sample rate (u16), sample count (u32, including 16 bytes of
// padding at each end).
func soundTics(name string) int32 {
	if t, ok := stepSoundTics[name]; ok {
		return t
	}
	t := int32(1)
	if lump := w_CheckNumForName("ds" + name); lump >= 0 {
		data := w_ReadLumpBytes(lump)
		if len(data) >= 8 {
			rate := int32(binary.LittleEndian.Uint16(data[2:]))
			samples := int32(binary.LittleEndian.Uint32(data[4:])) - 32
			if rate > 0 && samples > 0 {
				t = max(1, (samples*TICRATE+rate-1)/rate)
			}
		}
	}
	stepSoundTics[name] = t
	return t
}

func channelEnd(ch int32) *int32 {
	for int(ch) >= len(stepChannelEnd) {
		stepChannelEnd = append(stepChannelEnd, 0)
	}
	return &stepChannelEnd[ch]
}

// stepSoundModule is the sound_module_t StepInit installs. It plays nothing:
// it turns each sound DOOM starts, moves or stops into a SoundEvent, so a
// host plays sounds by draining StepSoundEvents rather than in real time. It
// still tracks how long each sound lasts, so DOOM frees channels as sounds
// finish, exactly as it would with real audio.
var stepSoundModule = sound_module_t{
	FInit:     func(boolean) boolean { return 1 },
	FShutdown: func() {},
	FGetSfxLumpNum: func(sfx *sfxinfo_t) int32 {
		return w_CheckNumForName("ds" + sampleName(sfx))
	},
	FUpdate: func() {},
	FUpdateSoundParams: func(ch, vol, sep int32) {
		emitSound(SoundEvent{Kind: "update", Channel: ch, Volume: vol, Sep: sep})
	},
	FStartSound: func(sfx *sfxinfo_t, ch, vol, sep int32) int32 {
		name := sampleName(sfx)
		*channelEnd(ch) = gametic + soundTics(name)
		emitSound(SoundEvent{Kind: "start", Channel: ch, Name: strings.ToLower(name), Volume: vol, Sep: sep})
		return ch // the handle DOOM passes back to us is the channel
	},
	FStopSound: func(ch int32) {
		*channelEnd(ch) = 0
		emitSound(SoundEvent{Kind: "stop", Channel: ch})
	},
	FSoundIsPlaying: func(ch int32) boolean {
		return booluint32(gametic < *channelEnd(ch))
	},
}

// stepUpdateSounds frees finished channels and repositions playing sounds,
// as the regular main loop does once per frame.
func stepUpdateSounds() {
	if mo := players[consoleplayer].Fmo; mo != nil {
		s_UpdateSounds(&mo.degenmobj_t)
	}
}

package gore

// Step API: drive the engine one tic at a time from a deterministic host.
//
// The game simulation itself is already deterministic: it uses fixed-point
// math and its own random number table, and never reads the clock. That is
// why vanilla .lmp demos work: a demo is just the ticcmds of a game, and
// replaying them reproduces the game exactly.
//
// What isn't deterministic is the loop around the simulation. Run() reads
// the wall clock to decide when to run tics, and builds each ticcmd from
// keyboard and mouse events. (Demo playback only replaces the second part.)
// The step API replaces both: the host calls StepTic for every tic, with an
// explicit ticcmd, and game time is derived from the number of tics run. So
// the engine is purely deterministic: the same WAD, arguments and ticcmds
// always produce the same game.

import (
	"errors"
	"hash/fnv"
	"image"
	"io"
	"time"
)

// dg_step_mode is set when the engine is driven through the step API.
var dg_step_mode bool

// Button bits for TicCmd.Buttons. These are the engine's bt_* values from
// d_event.h, exported for step API hosts.
const (
	ButtonAttack      = 1
	ButtonUse         = 2
	ButtonChange      = 4    // change weapon; weapon number in ButtonWeaponMask
	ButtonWeaponMask  = 0x38 // (weapon number) << ButtonWeaponShift
	ButtonWeaponShift = 3
	ButtonSpecial     = 0x80 // pause/savegame; never valid from a remote player
)

// TicCmd is one tic of a player's input, as passed to StepTic. It is the
// part of the engine's internal ticcmd_t that a player controls; the rest
// of ticcmd_t is bookkeeping for the network code.
//
// For reference, the keyboard moves at Forward 25 (50 when running), and
// turns at AngleTurn 640 per tic (1280 when running).
type TicCmd struct {
	Forward int8 `json:"f,omitempty"` // forward (+) / backward (-) movement
	Side    int8 `json:"s,omitempty"` // strafe right (+) / left (-)
	// AngleTurn turns left (+) or right (-); 65536 is a full turn. It is
	// added to the top 16 bits of the 32-bit view angle.
	AngleTurn int16 `json:"t,omitempty"`
	Buttons   uint8 `json:"b,omitempty"` // Button* bits
}

// Stats is a snapshot of interesting game state, cheap to compute every tic.
// See GetStats.
type Stats struct {
	Tic          int32  `json:"tic"`
	LevelTime    int32  `json:"levelTime"`
	GameState    string `json:"gameState"`
	Skill        int32  `json:"skill"`
	Episode      int32  `json:"episode"`
	Map          int32  `json:"map"`
	Health       int32  `json:"health"`
	Armor        int32  `json:"armor"`
	Weapon       int32  `json:"weapon"`
	Ammo         int32  `json:"ammo"` // ammo for the ready weapon, -1 if it needs none
	Dead         bool   `json:"dead"`
	Kills        int32  `json:"kills"`
	TotalKills   int32  `json:"totalKills"`
	Items        int32  `json:"items"`
	TotalItems   int32  `json:"totalItems"`
	Secrets      int32  `json:"secrets"`
	TotalSecrets int32  `json:"totalSecrets"`
	Checksum     uint32 `json:"checksum"`
}

// nullFrontend is the frontend StepInit uses when the host passes nil. Step
// hosts usually have no use for one: they read frames with FrameBuffer and
// sounds with StepSoundEvents, and pass input to StepTic.
type nullFrontend struct{}

var _ DoomFrontend = nullFrontend{}

func (nullFrontend) DrawFrame(img *image.RGBA)                       {}
func (nullFrontend) SetTitle(title string)                           {}
func (nullFrontend) GetEvent(event *DoomEvent) bool                  { return false }
func (nullFrontend) CacheSound(name string, data []byte)             {}
func (nullFrontend) PlaySound(name string, channel, volume, sep int) {}

// StepInit initialises the engine for stepping. args are regular DOOM command
// line arguments, e.g. {"-iwad", "doom1.wad", "-warp", "1", "1", "-skill", "3"}.
// Only one engine may exist per process (or per WASM instance).
func StepInit(fe DoomFrontend, args []string) {
	if fe == nil {
		fe = nullFrontend{}
	}
	dg_frontend = fe
	dg_step_mode = true
	dg_exiting = false
	start_time = time.Time{}
	// Tics are only ever run by StepTic, one at a time, as with -timedemo.
	// This also stops netUpdate, which the renderer calls, from building
	// ticcmds out of keyboard events.
	singletics = 1
	sound_module = &stepSoundModule // see step_sound.go
	doomgeneric_Create(append([]string{"doom"}, args...))
	stepView = consoleplayer
}

// StepTic runs exactly one game tic, with cmd as the console player's input.
func StepTic(cmd TicCmd) {
	var cmds [NET_MAXPLAYERS]ticcmd_t
	var ingame [NET_MAXPLAYERS]boolean
	cmds[consoleplayer] = ticcmd_t{
		Fforwardmove: cmd.Forward,
		Fsidemove:    cmd.Side,
		Fangleturn:   cmd.AngleTurn,
		Fbuttons:     cmd.Buttons,
	}
	ingame[consoleplayer] = 1
	stepSoundTic = gametic + 1
	runTic(cmds[:], ingame[:])
	gametic++
	stepUpdateSounds()
}

// StepRender draws the current game state into the frame buffer. It only
// touches renderer state, never game state, so a host can render as often or
// as rarely as it likes without affecting determinism.
func StepRender() {
	if gamestate == gs_LEVEL && players[displayplayer].Fmo == nil {
		return // nobody to look through (see stepFixView); keep the last frame
	}
	d_Display()
}

// FrameBuffer returns the 320x200 palette-indexed frame buffer as last drawn
// by StepRender. The slice is owned by the engine.
func FrameBuffer() []byte {
	return I_VideoBuffer
}

// Palette returns the current 256 entry RGB palette (768 bytes). DOOM uses
// palette swaps for damage/pickup flashes, so this changes over time.
func Palette() []byte {
	pal := make([]byte, 768)
	for i := range 256 {
		pal[i*3] = colors[i].R
		pal[i*3+1] = colors[i].G
		pal[i*3+2] = colors[i].B
	}
	return pal
}

// GetStats returns a snapshot of the current game state, from the console
// player's point of view.
func GetStats() Stats {
	return GetPlayerStats(consoleplayer)
}

// GetPlayerStats returns a snapshot of the current game state, from player
// p's point of view.
func GetPlayerStats(p32 int32) Stats {
	p := &players[p32]
	s := Stats{
		Tic:          gametic,
		LevelTime:    leveltime,
		Skill:        int32(gameskill),
		Episode:      gameepisode,
		Map:          gamemap,
		Health:       p.Fhealth,
		Armor:        p.Farmorpoints,
		Weapon:       int32(p.Freadyweapon),
		Ammo:         -1,
		Dead:         p.Fplayerstate == Pst_DEAD,
		Kills:        p.Fkillcount,
		TotalKills:   totalkills,
		Items:        p.Fitemcount,
		TotalItems:   totalitems,
		Secrets:      p.Fsecretcount,
		TotalSecrets: totalsecret,
		Checksum:     Checksum(),
	}
	if int(p.Freadyweapon) < len(weaponinfo) {
		if am := weaponinfo[p.Freadyweapon].Fammo; am != am_noammo {
			s.Ammo = p.Fammo[am]
		}
	}
	switch gamestate {
	case gs_LEVEL:
		s.GameState = "level"
	case gs_INTERMISSION:
		s.GameState = "intermission"
	case gs_FINALE:
		s.GameState = "finale"
	default:
		s.GameState = "demoscreen"
	}
	return s
}

// Checksum hashes the simulation state: RNG index, level time and every map
// object's position, momentum, angle, health and state. Two engines that have
// consumed the same ticcmds from the same starting point return the same
// value; if they don't, they have desynced.
func Checksum() uint32 {
	h := fnv.New32a()
	var buf [4]byte
	put := func(v int32) {
		buf[0], buf[1], buf[2], buf[3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
		h.Write(buf[:])
	}
	put(gametic)
	put(leveltime)
	put(int32(gamestate))
	put(prndindex)
	put(gameepisode)
	put(gamemap)
	if gamestate == gs_LEVEL {
		for th := thinkercap.Fnext; th != nil && th != &thinkercap; th = th.Fnext {
			mo, ok := th.Ffunction.(*mobj_t)
			if !ok {
				continue
			}
			put(mo.Fx)
			put(mo.Fy)
			put(mo.Fz)
			put(int32(mo.Fangle))
			put(mo.Fmomx)
			put(mo.Fmomy)
			put(mo.Fmomz)
			put(int32(mo.Ftype1))
			put(mo.Fhealth)
			put(mo.Ftics)
			put(mo.Fflags)
		}
	}
	// Every player in the game, not just the console player: in a
	// multiplayer game each node has a different console player.
	if netgame != 0 {
		put(int32(PlayersInGame()))
	}
	for i := range players {
		if playeringame[i] == 0 {
			continue
		}
		p := &players[i]
		put(p.Fhealth)
		put(p.Farmorpoints)
		for _, a := range p.Fammo {
			put(a)
		}
		put(int32(p.Freadyweapon))
	}
	return h.Sum32()
}

// CanSave reports whether the game is in a state that SaveGameBytes can
// capture: in a level, with the player alive. In a netgame dead players are
// fine (they respawn when they press use), but everyone in the game must
// have spawned.
func CanSave() bool {
	if gamestate != gs_LEVEL || gameaction != ga_nothing {
		return false
	}
	if netgame == 0 {
		return players[consoleplayer].Fplayerstate != Pst_DEAD
	}
	for i := range players {
		if playeringame[i] != 0 && (players[i].Fmo == nil || players[i].Fplayerstate == Pst_REBORN) {
			return false
		}
	}
	return true
}

// SaveGameBytes serialises the current level into a vanilla-format savegame,
// followed by a small trailer with state vanilla doesn't save (RNG indices
// and the tic counter). Unlike g_DoSaveGame it does not print "game saved"
// or otherwise touch game state.
func SaveGameBytes() ([]byte, error) {
	if !CanSave() {
		return nil, errors.New("gore: can only save while alive in a level")
	}
	m := &memStream{}
	save_stream = m
	defer func() { save_stream = nil }()
	savegame_error = 0
	p_WriteSaveGameHeader("gore")
	p_ArchivePlayers()
	p_ArchiveWorld()
	p_ArchiveThinkers()
	p_ArchiveSpecials()
	p_WriteSaveGameEOF()
	saveg_write32(prndindex)
	saveg_write32(rndindex)
	saveg_write32(gametic)
	if savegame_error != 0 {
		return nil, errors.New("gore: error writing savegame")
	}
	return m.buf, nil
}

// LoadGameBytes restores a savegame produced by SaveGameBytes. It should be
// called on a freshly initialised engine.
func LoadGameBytes(data []byte) error {
	m := &memStream{buf: data}
	save_stream = m
	defer func() { save_stream = nil }()
	savegame_error = 0
	if p_ReadSaveGameHeader() == 0 {
		return errors.New("gore: bad savegame header")
	}
	savedleveltime := leveltime
	g_InitNew(gameskill, gameepisode, gamemap)
	leveltime = savedleveltime
	p_UnArchivePlayers()
	p_UnArchiveWorld()
	p_UnArchiveThinkers()
	p_UnArchiveSpecials()
	if p_ReadSaveGameEOF() == 0 {
		return errors.New("gore: bad savegame")
	}
	prndindex = saveg_read32()
	rndindex = saveg_read32()
	gametic = saveg_read32()
	if savegame_error != 0 {
		return errors.New("gore: truncated savegame")
	}
	gameaction = ga_nothing
	stepFixView() // the console player may not be in the saved game
	return nil
}

// memStream is an in-memory saveStream.
type memStream struct {
	buf []byte
	pos int
}

func (m *memStream) Read(p []byte) (int, error) {
	if m.pos >= len(m.buf) {
		return 0, io.EOF
	}
	n := copy(p, m.buf[m.pos:])
	m.pos += n
	return n, nil
}

func (m *memStream) Write(p []byte) (int, error) {
	if end := m.pos + len(p); end > len(m.buf) {
		m.buf = append(m.buf, make([]byte, end-len(m.buf))...)
	}
	n := copy(m.buf[m.pos:], p)
	m.pos += n
	return n, nil
}

func (m *memStream) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.pos = int(offset)
	case io.SeekCurrent:
		m.pos += int(offset)
	case io.SeekEnd:
		m.pos = len(m.buf) + int(offset)
	}
	return int64(m.pos), nil
}

func (m *memStream) Close() error { return nil }

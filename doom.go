package gore

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type DoomFrontend interface {
	DrawFrame(img *image.RGBA)
	SetTitle(title string)
	GetEvent(event *DoomEvent) bool
}

var dg_frontend DoomFrontend
var dg_run_full_speed bool = false // If true, don't ever sleep, and just tick up once-per-frame
var dg_fake_tics uint64
var dg_exiting bool
var start_time time.Time

type boolean = uint32

// LIBC functions
func xabs(j int32) int32 {
	if j < 0 {
		return -j
	}
	return j
}

func xtoupper(c int32) int32 {
	if c >= 'a' && c <= 'z' {
		return c - ('a' - 'A')
	}

	return c
}

func xmemcpy(dest, src uintptr, n uint64) uintptr {
	if n != 0 {
		srcSlice := unsafe.Slice((*byte)(unsafe.Pointer(src)), n)
		destSlice := unsafe.Slice((*byte)(unsafe.Pointer(dest)), n)
		copy(destSlice, srcSlice)
	}
	return dest
}

func boolint32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func booluint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func gostring_bytes(s []byte) string {
	if len(s) == 0 {
		return ""
	}
	var end int
	for ; end < len(s) && s[end] != 0; end++ {
	}
	return string(s[:end])
}

func gostring(s uintptr) string {
	if s == 0 {
		return ""
	}

	p := s
	for *(*byte)(unsafe.Pointer(p)) != 0 {
		p++
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(s)), p-s))
}

func gostring_n(s uintptr, n int) string {
	if s == 0 || n <= 0 {
		return ""
	}
	p := s
	for i := 0; i < n && *(*byte)(unsafe.Pointer(p)) != 0; i++ {
		p++
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(s)), p-s))
}

const AM_NUMMARKPOINTS = 10
const ANGLETOFINESHIFT = 19
const BACKUPTICS = 128
const BASETHRESHOLD = 100
const BLACK = 0
const BUTTONTIME = 35
const DOOM_191_VERSION = 111
const FASTDARK = 15
const FINEANGLES = 8192
const FRACBITS = 16
const FRACUNIT = 1 << FRACBITS
const F_PANINC = 4
const GLOWSPEED = 8
const GRAYSRANGE = 16
const GREENRANGE = 16
const INT_MAX1 = 2147483647
const ITEMQUESIZE = 128
const KEY_ENTER = 13
const KEY_ESCAPE = 27
const KEY_TAB = 9
const LIGHTLEVELS = 16
const LIGHTSCALESHIFT = 12
const LIGHTSEGSHIFT = 4
const LIGHTZSHIFT = 20
const MAPBLOCKUNITS = 128
const MAXBUTTONS = 16
const MAXCEILINGS = 30
const MAXDRAWSEGS = 256
const MAXHEALTH = 100
const MAXINTERCEPTS_ORIGINAL = 128
const MAXLIGHTSCALE = 48
const MAXLIGHTZ = 128
const MAXPLATS = 30
const MAXPLAYERS = 4
const MAXSPECIALCROSS_ORIGINAL = 8
const MAXSWITCHES = 50
const MAX_MOUSE_BUTTONS = 8
const ml_BLOCKING = 1
const ml_BLOCKMONSTERS = 2
const ml_DONTDRAW = 128
const ml_DONTPEGBOTTOM = 16
const ml_DONTPEGTOP = 8
const ml_MAPPED = 256
const ml_SECRET = 32
const ml_SOUNDBLOCK = 64
const ml_TWOSIDED = 4
const MTF_AMBUSH = 8
const NET_MAXPLAYERS = 8
const NUMCOLORMAPS = 32
const NUM_QUITMESSAGES = 8
const PLATWAIT = 3
const PT_ADDLINES = 1
const PT_ADDTHINGS = 2
const PT_EARLYOUT = 4
const REDRANGE = 16
const SCREENHEIGHT = 200
const SCREENHEIGHT_4_3 = 240
const SCREENWIDTH = 320
const SCREENWIDTH_4_3 = 256
const SIL_BOTH = 3
const SIL_BOTTOM = 1
const SIL_TOP = 2
const SLOPEBITS = 11
const SLOPERANGE = 2048
const SLOWDARK = 35
const STROBEBRIGHT = 5
const st_HEIGHT = 32
const TICRATE = 35
const VDOORWAIT = 150

type sha1_digest_t = [20]uint8

type gamemission_t int32

const doom gamemission_t = 0
const doom2 gamemission_t = 1
const pack_tnt gamemission_t = 2
const pack_plut gamemission_t = 3
const pack_chex gamemission_t = 4
const pack_hacx gamemission_t = 5
const heretic gamemission_t = 6
const hexen gamemission_t = 7
const strife gamemission_t = 8
const none gamemission_t = 9

type gamemode_t int32

const shareware gamemode_t = 0
const registered gamemode_t = 1
const commercial gamemode_t = 2
const retail gamemode_t = 3
const indetermined gamemode_t = 4

type gameversion_t int32

const exe_doom_1_2 gameversion_t = 0
const exe_doom_1_666 gameversion_t = 1
const exe_doom_1_7 gameversion_t = 2
const exe_doom_1_8 gameversion_t = 3
const exe_doom_1_9 gameversion_t = 4
const exe_hacx gameversion_t = 5
const exe_ultimate gameversion_t = 6
const exe_final gameversion_t = 7
const exe_final2 gameversion_t = 8
const exe_chex gameversion_t = 9

type skill_t int32

const sk_baby skill_t = 0
const sk_easy skill_t = 1
const sk_medium skill_t = 2
const sk_nightmare skill_t = 4

type gamestate_t int32

const gs_LEVEL gamestate_t = 0
const gs_INTERMISSION gamestate_t = 1
const gs_FINALE gamestate_t = 2
const gs_DEMOSCREEN gamestate_t = 3

type gameaction_t int32

const ga_nothing = 0
const ga_loadlevel = 1
const ga_newgame = 2
const ga_loadgame = 3
const ga_savegame = 4
const ga_playdemo = 5
const ga_completed = 6
const ga_victory = 7
const ga_worlddone = 8
const ga_screenshot = 9

type card_t int32

const it_bluecard = 0
const it_yellowcard = 1
const it_redcard = 2
const it_blueskull = 3
const it_yellowskull = 4
const it_redskull = 5
const NUMCARDS = 6

type weapontype_t int32

const wp_fist = 0
const wp_pistol = 1
const wp_shotgun = 2
const wp_chaingun = 3
const wp_missile = 4
const wp_plasma = 5
const wp_bfg = 6
const wp_chainsaw = 7
const wp_supershotgun = 8
const NUMWEAPONS = 9
const wp_nochange = 10

type ammotype_t int32

const am_clip = 0
const am_shell = 1
const am_cell = 2
const am_misl = 3
const NUMAMMO = 4
const am_noammo = 5

const pw_invulnerability = 0
const pw_strength = 1
const pw_invisibility = 2
const pw_ironfeet = 3
const pw_allmap = 4
const pw_infrared = 5
const NUMPOWERS = 6

const INVULNTICS = 1050
const INVISTICS = 2100
const INFRATICS = 4200
const IRONTICS = 2100

type Evtype_t int32

const Ev_keydown Evtype_t = 0
const Ev_keyup Evtype_t = 1
const Ev_mouse Evtype_t = 2
const Ev_joystick Evtype_t = 3
const Ev_quit Evtype_t = 4

type event_t struct {
	Ftype1 Evtype_t
	Fdata1 int32
	Fdata2 int32
	Fdata3 int32
	Fdata4 int32
}

const bt_ATTACK = 1
const bt_USE = 2
const bt_SPECIAL = 128
const bt_SPECIALMASK = 3
const bt_CHANGE = 4
const bt_WEAPONMASK = 56
const bt_WEAPONSHIFT = 3
const bts_PAUSE = 1
const bts_SAVEGAME = 2
const bts_SAVEMASK = 28
const bts_SAVESHIFT = 2

type lumpType interface {
	*patch_t
}

type cheatseq_t struct {
	Fsequence         string
	Fsequence_len     uint64
	Fparameter_chars  int32
	Fchars_read       uint64
	Fparam_chars_read int32
	Fparameter_buf    [5]byte
}

type fixed_t = int32

func float2fixed(f float32) fixed_t {
	return fixed_t(f * float32(FRACUNIT))
}

func float2fixedinv(f float32) fixed_t {
	return fixed_t(float32(FRACUNIT) / f)
}

type angle_t = uint32

type thinker_t struct {
	Fprev     *thinker_t
	Fnext     *thinker_t
	Ffunction thinker_func_t
}

type thinker_func_t interface {
	ThinkerFunc()
}

const ml_THINGS = 1
const ml_LINEDEFS = 2
const ml_SIDEDEFS = 3
const ml_VERTEXES = 4
const ml_SEGS = 5
const ml_SSECTORS = 6
const ml_NODES = 7
const ml_SECTORS = 8
const ml_REJECT = 9
const ml_BLOCKMAP = 10

type mapvertex_t struct {
	Fx int16
	Fy int16
}

type mapsidedef_t struct {
	Ftextureoffset int16
	Frowoffset     int16
	Ftoptexture    [8]byte
	Fbottomtexture [8]byte
	Fmidtexture    [8]byte
	Fsector        int16
}

type maplinedef_t struct {
	Fv1      int16
	Fv2      int16
	Fflags   int16
	Fspecial int16
	Ftag     int16
	Fsidenum [2]int16
}

type mapsector_t struct {
	Ffloorheight   int16
	Fceilingheight int16
	Ffloorpic      [8]byte
	Fceilingpic    [8]byte
	Flightlevel    int16
	Fspecial       int16
	Ftag           int16
}

type mapsubsector_t struct {
	Fnumsegs  int16
	Ffirstseg int16
}

type mapseg_t struct {
	Fv1      int16
	Fv2      int16
	Fangle   int16
	Flinedef int16
	Fside    int16
	Foffset  int16
}

type mapnode_t struct {
	Fx        int16
	Fy        int16
	Fdx       int16
	Fdy       int16
	Fbbox     [2][4]int16
	Fchildren [2]uint16
}

type mapthing_t struct {
	Fx       int16
	Fy       int16
	Fangle   int16
	Ftype1   int16
	Foptions int16
}

type spritenum_t = int32

const spr_SHTG = 1
const spr_PUNG = 2
const spr_PISG = 3
const spr_PISF = 4
const spr_SHTF = 5
const spr_SHT2 = 6
const spr_CHGG = 7
const spr_CHGF = 8
const spr_MISG = 9
const spr_MISF = 10
const spr_SAWG = 11
const spr_PLSG = 12
const spr_PLSF = 13
const spr_BFGG = 14
const spr_BFGF = 15
const spr_BLUD = 16
const spr_PUFF = 17
const spr_BAL1 = 18
const spr_BAL2 = 19
const spr_PLSS = 20
const spr_PLSE = 21
const spr_MISL = 22
const spr_BFS1 = 23
const spr_BFE1 = 24
const spr_BFE2 = 25
const spr_TFOG = 26
const spr_IFOG = 27
const spr_PLAY = 28
const spr_POSS = 29
const spr_SPOS = 30
const spr_VILE = 31
const spr_FIRE = 32
const spr_FATB = 33
const spr_FBXP = 34
const spr_SKEL = 35
const spr_MANF = 36
const spr_FATT = 37
const spr_CPOS = 38
const spr_SARG = 39
const spr_HEAD = 40
const spr_BAL7 = 41
const spr_BOSS = 42
const spr_BOS2 = 43
const spr_SKUL = 44
const spr_SPID = 45
const spr_BSPI = 46
const spr_APLS = 47
const spr_APBX = 48
const spr_CYBR = 49
const spr_PAIN = 50
const spr_SSWV = 51
const spr_KEEN = 52
const spr_BBRN = 53
const spr_BOSF = 54
const spr_ARM1 = 55
const spr_ARM2 = 56
const spr_BAR1 = 57
const spr_BEXP = 58
const spr_FCAN = 59
const spr_BON1 = 60
const spr_BON2 = 61
const spr_BKEY = 62
const spr_RKEY = 63
const spr_YKEY = 64
const spr_BSKU = 65
const spr_RSKU = 66
const spr_YSKU = 67
const spr_STIM = 68
const spr_MEDI = 69
const spr_SOUL = 70
const spr_PINV = 71
const spr_PSTR = 72
const spr_PINS = 73
const spr_MEGA = 74
const spr_SUIT = 75
const spr_PMAP = 76
const spr_PVIS = 77
const spr_CLIP = 78
const spr_AMMO = 79
const spr_ROCK = 80
const spr_BROK = 81
const spr_CELL = 82
const spr_CELP = 83
const spr_SHEL = 84
const spr_SBOX = 85
const spr_BPAK = 86
const spr_BFUG = 87
const spr_MGUN = 88
const spr_CSAW = 89
const spr_LAUN = 90
const spr_PLAS = 91
const spr_SHOT = 92
const spr_SGN2 = 93
const spr_COLU = 94
const spr_SMT2 = 95
const spr_GOR1 = 96
const spr_POL2 = 97
const spr_POL5 = 98
const spr_POL4 = 99
const spr_POL3 = 100
const spr_POL1 = 101
const spr_POL6 = 102
const spr_GOR2 = 103
const spr_GOR3 = 104
const spr_GOR4 = 105
const spr_GOR5 = 106
const spr_SMIT = 107
const spr_COL1 = 108
const spr_COL2 = 109
const spr_COL3 = 110
const spr_COL4 = 111
const spr_CAND = 112
const spr_CBRA = 113
const spr_COL6 = 114
const spr_TRE1 = 115
const spr_TRE2 = 116
const spr_ELEC = 117
const spr_CEYE = 118
const spr_FSKU = 119
const spr_COL5 = 120
const spr_TBLU = 121
const spr_TGRN = 122
const spr_TRED = 123
const spr_SMBT = 124
const spr_SMGT = 125
const spr_SMRT = 126
const spr_HDB1 = 127
const spr_HDB2 = 128
const spr_HDB3 = 129
const spr_HDB4 = 130
const spr_HDB5 = 131
const spr_HDB6 = 132
const spr_POB1 = 133
const spr_POB2 = 134
const spr_BRS1 = 135
const spr_TLMP = 136
const spr_TLP2 = 137

type statenum_t = int32

const s_NULL = 0
const s_LIGHTDONE = 1
const s_PUNCH = 2
const s_PUNCHDOWN = 3
const s_PUNCHUP = 4
const s_PUNCH1 = 5
const s_PUNCH2 = 6
const s_PUNCH3 = 7
const s_PUNCH4 = 8
const s_PUNCH5 = 9
const s_PISTOL = 10
const s_PISTOLDOWN = 11
const s_PISTOLUP = 12
const s_PISTOL1 = 13
const s_PISTOL2 = 14
const s_PISTOL3 = 15
const s_PISTOL4 = 16
const s_PISTOLFLASH = 17
const s_SGUN = 18
const s_SGUNDOWN = 19
const s_SGUNUP = 20
const s_SGUN1 = 21
const s_SGUN2 = 22
const s_SGUN3 = 23
const s_SGUN4 = 24
const s_SGUN5 = 25
const s_SGUN6 = 26
const s_SGUN7 = 27
const s_SGUN8 = 28
const s_SGUN9 = 29
const s_SGUNFLASH1 = 30
const s_SGUNFLASH2 = 31
const s_DSGUN = 32
const s_DSGUNDOWN = 33
const s_DSGUNUP = 34
const s_DSGUN1 = 35
const s_DSGUN2 = 36
const s_DSGUN3 = 37
const s_DSGUN4 = 38
const s_DSGUN5 = 39
const s_DSGUN6 = 40
const s_DSGUN7 = 41
const s_DSGUN8 = 42
const s_DSGUN9 = 43
const s_DSGUN10 = 44
const s_DSNR2 = 46
const s_DSGUNFLASH1 = 47
const s_DSGUNFLASH2 = 48
const s_CHAIN = 49
const s_CHAINDOWN = 50
const s_CHAINUP = 51
const s_CHAIN1 = 52
const s_CHAIN2 = 53
const s_CHAIN3 = 54
const s_CHAINFLASH1 = 55
const s_MISSILE = 57
const s_MISSILEDOWN = 58
const s_MISSILEUP = 59
const s_MISSILE1 = 60
const s_MISSILE2 = 61
const s_MISSILE3 = 62
const s_MISSILEFLASH1 = 63
const s_MISSILEFLASH2 = 64
const s_MISSILEFLASH3 = 65
const s_MISSILEFLASH4 = 66
const s_SAW = 67
const s_SAWB = 68
const s_SAWDOWN = 69
const s_SAWUP = 70
const s_SAW1 = 71
const s_SAW2 = 72
const s_SAW3 = 73
const s_PLASMA = 74
const s_PLASMADOWN = 75
const s_PLASMAUP = 76
const s_PLASMA1 = 77
const s_PLASMA2 = 78
const s_PLASMAFLASH1 = 79
const s_BFG = 81
const s_BFGDOWN = 82
const s_BFGUP = 83
const s_BFG1 = 84
const s_BFG2 = 85
const s_BFG3 = 86
const s_BFG4 = 87
const s_BFGFLASH1 = 88
const s_BFGFLASH2 = 89
const s_BLOOD1 = 90
const s_BLOOD2 = 91
const s_BLOOD3 = 92
const s_PUFF1 = 93
const s_PUFF2 = 94
const s_PUFF3 = 95
const s_PUFF4 = 96
const s_TBALL1 = 97
const s_TBALL2 = 98
const s_TBALLX1 = 99
const s_TBALLX2 = 100
const s_TBALLX3 = 101
const s_RBALL1 = 102
const s_RBALL2 = 103
const s_RBALLX1 = 104
const s_RBALLX2 = 105
const s_RBALLX3 = 106
const s_PLASBALL = 107
const s_PLASBALL2 = 108
const s_PLASEXP = 109
const s_PLASEXP2 = 110
const s_PLASEXP3 = 111
const s_PLASEXP4 = 112
const s_PLASEXP5 = 113
const s_ROCKET = 114
const s_BFGSHOT = 115
const s_BFGSHOT2 = 116
const s_BFGLAND = 117
const s_BFGLAND2 = 118
const s_BFGLAND3 = 119
const s_BFGLAND4 = 120
const s_BFGLAND5 = 121
const s_BFGLAND6 = 122
const s_BFGEXP = 123
const s_BFGEXP2 = 124
const s_BFGEXP3 = 125
const s_BFGEXP4 = 126
const s_EXPLODE1 = 127
const s_EXPLODE2 = 128
const s_EXPLODE3 = 129
const s_TFOG = 130
const s_TFOG01 = 131
const s_TFOG02 = 132
const s_TFOG2 = 133
const s_TFOG3 = 134
const s_TFOG4 = 135
const s_TFOG5 = 136
const s_TFOG6 = 137
const s_TFOG7 = 138
const s_TFOG8 = 139
const s_TFOG9 = 140
const s_TFOG10 = 141
const s_IFOG = 142
const s_IFOG01 = 143
const s_IFOG02 = 144
const s_IFOG2 = 145
const s_IFOG3 = 146
const s_IFOG4 = 147
const s_IFOG5 = 148
const s_PLAY = 149
const s_PLAY_RUN1 = 150
const s_PLAY_RUN2 = 151
const s_PLAY_RUN3 = 152
const s_PLAY_RUN4 = 153
const s_PLAY_ATK1 = 154
const s_PLAY_ATK2 = 155
const s_PLAY_PAIN = 156
const s_PLAY_PAIN2 = 157
const s_PLAY_DIE1 = 158
const s_PLAY_DIE2 = 159
const s_PLAY_DIE3 = 160
const s_PLAY_DIE4 = 161
const s_PLAY_DIE5 = 162
const s_PLAY_DIE6 = 163
const s_PLAY_DIE7 = 164
const s_PLAY_XDIE1 = 165
const s_PLAY_XDIE2 = 166
const s_PLAY_XDIE3 = 167
const s_PLAY_XDIE4 = 168
const s_PLAY_XDIE5 = 169
const s_PLAY_XDIE6 = 170
const s_PLAY_XDIE7 = 171
const s_PLAY_XDIE8 = 172
const s_PLAY_XDIE9 = 173
const s_POSS_STND = 174
const s_POSS_STND2 = 175
const s_POSS_RUN1 = 176
const s_POSS_RUN2 = 177
const s_POSS_RUN3 = 178
const s_POSS_RUN4 = 179
const s_POSS_RUN5 = 180
const s_POSS_RUN6 = 181
const s_POSS_RUN7 = 182
const s_POSS_RUN8 = 183
const s_POSS_ATK1 = 184
const s_POSS_ATK2 = 185
const s_POSS_ATK3 = 186
const s_POSS_PAIN = 187
const s_POSS_PAIN2 = 188
const s_POSS_DIE1 = 189
const s_POSS_DIE2 = 190
const s_POSS_DIE3 = 191
const s_POSS_DIE4 = 192
const s_POSS_DIE5 = 193
const s_POSS_XDIE1 = 194
const s_POSS_XDIE2 = 195
const s_POSS_XDIE3 = 196
const s_POSS_XDIE4 = 197
const s_POSS_XDIE5 = 198
const s_POSS_XDIE6 = 199
const s_POSS_XDIE7 = 200
const s_POSS_XDIE8 = 201
const s_POSS_XDIE9 = 202
const s_POSS_RAISE1 = 203
const s_POSS_RAISE2 = 204
const s_POSS_RAISE3 = 205
const s_POSS_RAISE4 = 206
const s_SPOS_STND = 207
const s_SPOS_STND2 = 208
const s_SPOS_RUN1 = 209
const s_SPOS_RUN2 = 210
const s_SPOS_RUN3 = 211
const s_SPOS_RUN4 = 212
const s_SPOS_RUN5 = 213
const s_SPOS_RUN6 = 214
const s_SPOS_RUN7 = 215
const s_SPOS_RUN8 = 216
const s_SPOS_ATK1 = 217
const s_SPOS_ATK2 = 218
const s_SPOS_ATK3 = 219
const s_SPOS_PAIN = 220
const s_SPOS_PAIN2 = 221
const s_SPOS_DIE1 = 222
const s_SPOS_DIE2 = 223
const s_SPOS_DIE3 = 224
const s_SPOS_DIE4 = 225
const s_SPOS_DIE5 = 226
const s_SPOS_XDIE1 = 227
const s_SPOS_XDIE2 = 228
const s_SPOS_XDIE3 = 229
const s_SPOS_XDIE4 = 230
const s_SPOS_XDIE5 = 231
const s_SPOS_XDIE6 = 232
const s_SPOS_XDIE7 = 233
const s_SPOS_XDIE8 = 234
const s_SPOS_XDIE9 = 235
const s_SPOS_RAISE1 = 236
const s_SPOS_RAISE2 = 237
const s_SPOS_RAISE3 = 238
const s_SPOS_RAISE4 = 239
const s_SPOS_RAISE5 = 240
const s_VILE_STND = 241
const s_VILE_STND2 = 242
const s_VILE_RUN1 = 243
const s_VILE_RUN2 = 244
const s_VILE_RUN3 = 245
const s_VILE_RUN4 = 246
const s_VILE_RUN5 = 247
const s_VILE_RUN6 = 248
const s_VILE_RUN7 = 249
const s_VILE_RUN8 = 250
const s_VILE_RUN9 = 251
const s_VILE_RUN10 = 252
const s_VILE_RUN11 = 253
const s_VILE_RUN12 = 254
const s_VILE_ATK1 = 255
const s_VILE_ATK2 = 256
const s_VILE_ATK3 = 257
const s_VILE_ATK4 = 258
const s_VILE_ATK5 = 259
const s_VILE_ATK6 = 260
const s_VILE_ATK7 = 261
const s_VILE_ATK8 = 262
const s_VILE_ATK9 = 263
const s_VILE_ATK10 = 264
const s_VILE_ATK11 = 265
const s_VILE_HEAL1 = 266
const s_VILE_HEAL2 = 267
const s_VILE_HEAL3 = 268
const s_VILE_PAIN = 269
const s_VILE_PAIN2 = 270
const s_VILE_DIE1 = 271
const s_VILE_DIE2 = 272
const s_VILE_DIE3 = 273
const s_VILE_DIE4 = 274
const s_VILE_DIE5 = 275
const s_VILE_DIE6 = 276
const s_VILE_DIE7 = 277
const s_VILE_DIE8 = 278
const s_VILE_DIE9 = 279
const s_VILE_DIE10 = 280
const s_FIRE1 = 281
const s_FIRE2 = 282
const s_FIRE3 = 283
const s_FIRE4 = 284
const s_FIRE5 = 285
const s_FIRE6 = 286
const s_FIRE7 = 287
const s_FIRE8 = 288
const s_FIRE9 = 289
const s_FIRE10 = 290
const s_FIRE11 = 291
const s_FIRE12 = 292
const s_FIRE13 = 293
const s_FIRE14 = 294
const s_FIRE15 = 295
const s_FIRE16 = 296
const s_FIRE17 = 297
const s_FIRE18 = 298
const s_FIRE19 = 299
const s_FIRE20 = 300
const s_FIRE21 = 301
const s_FIRE22 = 302
const s_FIRE23 = 303
const s_FIRE24 = 304
const s_FIRE25 = 305
const s_FIRE26 = 306
const s_FIRE27 = 307
const s_FIRE28 = 308
const s_FIRE29 = 309
const s_FIRE30 = 310
const s_SMOKE1 = 311
const s_SMOKE2 = 312
const s_SMOKE3 = 313
const s_SMOKE4 = 314
const s_SMOKE5 = 315
const s_TRACER = 316
const s_TRACER2 = 317
const s_TRACEEXP1 = 318
const s_TRACEEXP2 = 319
const s_TRACEEXP3 = 320
const s_SKEL_STND = 321
const s_SKEL_STND2 = 322
const s_SKEL_RUN1 = 323
const s_SKEL_RUN2 = 324
const s_SKEL_RUN3 = 325
const s_SKEL_RUN4 = 326
const s_SKEL_RUN5 = 327
const s_SKEL_RUN6 = 328
const s_SKEL_RUN7 = 329
const s_SKEL_RUN8 = 330
const s_SKEL_RUN9 = 331
const s_SKEL_RUN10 = 332
const s_SKEL_RUN11 = 333
const s_SKEL_RUN12 = 334
const s_SKEL_FIST1 = 335
const s_SKEL_FIST2 = 336
const s_SKEL_FIST3 = 337
const s_SKEL_FIST4 = 338
const s_SKEL_MISS1 = 339
const s_SKEL_MISS2 = 340
const s_SKEL_MISS3 = 341
const s_SKEL_MISS4 = 342
const s_SKEL_PAIN = 343
const s_SKEL_PAIN2 = 344
const s_SKEL_DIE1 = 345
const s_SKEL_DIE2 = 346
const s_SKEL_DIE3 = 347
const s_SKEL_DIE4 = 348
const s_SKEL_DIE5 = 349
const s_SKEL_DIE6 = 350
const s_SKEL_RAISE1 = 351
const s_SKEL_RAISE2 = 352
const s_SKEL_RAISE3 = 353
const s_SKEL_RAISE4 = 354
const s_SKEL_RAISE5 = 355
const s_SKEL_RAISE6 = 356
const s_FATSHOT1 = 357
const s_FATSHOT2 = 358
const s_FATSHOTX1 = 359
const s_FATSHOTX2 = 360
const s_FATSHOTX3 = 361
const s_FATT_STND = 362
const s_FATT_STND2 = 363
const s_FATT_RUN1 = 364
const s_FATT_RUN2 = 365
const s_FATT_RUN3 = 366
const s_FATT_RUN4 = 367
const s_FATT_RUN5 = 368
const s_FATT_RUN6 = 369
const s_FATT_RUN7 = 370
const s_FATT_RUN8 = 371
const s_FATT_RUN9 = 372
const s_FATT_RUN10 = 373
const s_FATT_RUN11 = 374
const s_FATT_RUN12 = 375
const s_FATT_ATK1 = 376
const s_FATT_ATK2 = 377
const s_FATT_ATK3 = 378
const s_FATT_ATK4 = 379
const s_FATT_ATK5 = 380
const s_FATT_ATK6 = 381
const s_FATT_ATK7 = 382
const s_FATT_ATK8 = 383
const s_FATT_ATK9 = 384
const s_FATT_ATK10 = 385
const s_FATT_PAIN = 386
const s_FATT_PAIN2 = 387
const s_FATT_DIE1 = 388
const s_FATT_DIE2 = 389
const s_FATT_DIE3 = 390
const s_FATT_DIE4 = 391
const s_FATT_DIE5 = 392
const s_FATT_DIE6 = 393
const s_FATT_DIE7 = 394
const s_FATT_DIE8 = 395
const s_FATT_DIE9 = 396
const s_FATT_DIE10 = 397
const s_FATT_RAISE1 = 398
const s_FATT_RAISE2 = 399
const s_FATT_RAISE3 = 400
const s_FATT_RAISE4 = 401
const s_FATT_RAISE5 = 402
const s_FATT_RAISE6 = 403
const s_FATT_RAISE7 = 404
const s_FATT_RAISE8 = 405
const s_CPOS_STND = 406
const s_CPOS_STND2 = 407
const s_CPOS_RUN1 = 408
const s_CPOS_RUN2 = 409
const s_CPOS_RUN3 = 410
const s_CPOS_RUN4 = 411
const s_CPOS_RUN5 = 412
const s_CPOS_RUN6 = 413
const s_CPOS_RUN7 = 414
const s_CPOS_RUN8 = 415
const s_CPOS_ATK1 = 416
const s_CPOS_ATK2 = 417
const s_CPOS_ATK3 = 418
const s_CPOS_ATK4 = 419
const s_CPOS_PAIN = 420
const s_CPOS_PAIN2 = 421
const s_CPOS_DIE1 = 422
const s_CPOS_DIE2 = 423
const s_CPOS_DIE3 = 424
const s_CPOS_DIE4 = 425
const s_CPOS_DIE5 = 426
const s_CPOS_DIE6 = 427
const s_CPOS_DIE7 = 428
const s_CPOS_XDIE1 = 429
const s_CPOS_XDIE2 = 430
const s_CPOS_XDIE3 = 431
const s_CPOS_XDIE4 = 432
const s_CPOS_XDIE5 = 433
const s_CPOS_XDIE6 = 434
const s_CPOS_RAISE1 = 435
const s_CPOS_RAISE2 = 436
const s_CPOS_RAISE3 = 437
const s_CPOS_RAISE4 = 438
const s_CPOS_RAISE5 = 439
const s_CPOS_RAISE6 = 440
const s_CPOS_RAISE7 = 441
const s_TROO_STND = 442
const s_TROO_STND2 = 443
const s_TROO_RUN1 = 444
const s_TROO_RUN2 = 445
const s_TROO_RUN3 = 446
const s_TROO_RUN4 = 447
const s_TROO_RUN5 = 448
const s_TROO_RUN6 = 449
const s_TROO_RUN7 = 450
const s_TROO_RUN8 = 451
const s_TROO_ATK1 = 452
const s_TROO_ATK2 = 453
const s_TROO_ATK3 = 454
const s_TROO_PAIN = 455
const s_TROO_PAIN2 = 456
const s_TROO_DIE1 = 457
const s_TROO_DIE2 = 458
const s_TROO_DIE3 = 459
const s_TROO_DIE4 = 460
const s_TROO_DIE5 = 461
const s_TROO_XDIE1 = 462
const s_TROO_XDIE2 = 463
const s_TROO_XDIE3 = 464
const s_TROO_XDIE4 = 465
const s_TROO_XDIE5 = 466
const s_TROO_XDIE6 = 467
const s_TROO_XDIE7 = 468
const s_TROO_XDIE8 = 469
const s_TROO_RAISE1 = 470
const s_TROO_RAISE2 = 471
const s_TROO_RAISE3 = 472
const s_TROO_RAISE4 = 473
const s_TROO_RAISE5 = 474
const s_SARG_STND = 475
const s_SARG_STND2 = 476
const s_SARG_RUN1 = 477
const s_SARG_RUN2 = 478
const s_SARG_RUN3 = 479
const s_SARG_RUN4 = 480
const s_SARG_RUN5 = 481
const s_SARG_RUN6 = 482
const s_SARG_RUN7 = 483
const s_SARG_RUN8 = 484
const s_SARG_ATK1 = 485
const s_SARG_ATK2 = 486
const s_SARG_ATK3 = 487
const s_SARG_PAIN = 488
const s_SARG_PAIN2 = 489
const s_SARG_DIE1 = 490
const s_SARG_DIE2 = 491
const s_SARG_DIE3 = 492
const s_SARG_DIE4 = 493
const s_SARG_DIE5 = 494
const s_SARG_DIE6 = 495
const s_SARG_RAISE1 = 496
const s_SARG_RAISE2 = 497
const s_SARG_RAISE3 = 498
const s_SARG_RAISE4 = 499
const s_SARG_RAISE5 = 500
const s_SARG_RAISE6 = 501
const s_HEAD_STND = 502
const s_HEAD_RUN1 = 503
const s_HEAD_ATK1 = 504
const s_HEAD_ATK2 = 505
const s_HEAD_ATK3 = 506
const s_HEAD_PAIN = 507
const s_HEAD_PAIN2 = 508
const s_HEAD_PAIN3 = 509
const s_HEAD_DIE1 = 510
const s_HEAD_DIE2 = 511
const s_HEAD_DIE3 = 512
const s_HEAD_DIE4 = 513
const s_HEAD_DIE5 = 514
const s_HEAD_DIE6 = 515
const s_HEAD_RAISE1 = 516
const s_HEAD_RAISE2 = 517
const s_HEAD_RAISE3 = 518
const s_HEAD_RAISE4 = 519
const s_HEAD_RAISE5 = 520
const s_HEAD_RAISE6 = 521
const s_BRBALL1 = 522
const s_BRBALL2 = 523
const s_BRBALLX1 = 524
const s_BRBALLX2 = 525
const s_BRBALLX3 = 526
const s_BOSS_STND = 527
const s_BOSS_STND2 = 528
const s_BOSS_RUN1 = 529
const s_BOSS_RUN2 = 530
const s_BOSS_RUN3 = 531
const s_BOSS_RUN4 = 532
const s_BOSS_RUN5 = 533
const s_BOSS_RUN6 = 534
const s_BOSS_RUN7 = 535
const s_BOSS_RUN8 = 536
const s_BOSS_ATK1 = 537
const s_BOSS_ATK2 = 538
const s_BOSS_ATK3 = 539
const s_BOSS_PAIN = 540
const s_BOSS_PAIN2 = 541
const s_BOSS_DIE1 = 542
const s_BOSS_DIE2 = 543
const s_BOSS_DIE3 = 544
const s_BOSS_DIE4 = 545
const s_BOSS_DIE5 = 546
const s_BOSS_DIE6 = 547
const s_BOSS_DIE7 = 548
const s_BOSS_RAISE1 = 549
const s_BOSS_RAISE2 = 550
const s_BOSS_RAISE3 = 551
const s_BOSS_RAISE4 = 552
const s_BOSS_RAISE5 = 553
const s_BOSS_RAISE6 = 554
const s_BOSS_RAISE7 = 555
const s_BOS2_STND = 556
const s_BOS2_STND2 = 557
const s_BOS2_RUN1 = 558
const s_BOS2_RUN2 = 559
const s_BOS2_RUN3 = 560
const s_BOS2_RUN4 = 561
const s_BOS2_RUN5 = 562
const s_BOS2_RUN6 = 563
const s_BOS2_RUN7 = 564
const s_BOS2_RUN8 = 565
const s_BOS2_ATK1 = 566
const s_BOS2_ATK2 = 567
const s_BOS2_ATK3 = 568
const s_BOS2_PAIN = 569
const s_BOS2_PAIN2 = 570
const s_BOS2_DIE1 = 571
const s_BOS2_DIE2 = 572
const s_BOS2_DIE3 = 573
const s_BOS2_DIE4 = 574
const s_BOS2_DIE5 = 575
const s_BOS2_DIE6 = 576
const s_BOS2_DIE7 = 577
const s_BOS2_RAISE1 = 578
const s_BOS2_RAISE2 = 579
const s_BOS2_RAISE3 = 580
const s_BOS2_RAISE4 = 581
const s_BOS2_RAISE5 = 582
const s_BOS2_RAISE6 = 583
const s_BOS2_RAISE7 = 584
const s_SKULL_STND = 585
const s_SKULL_STND2 = 586
const s_SKULL_RUN1 = 587
const s_SKULL_RUN2 = 588
const s_SKULL_ATK1 = 589
const s_SKULL_ATK2 = 590
const s_SKULL_ATK3 = 591
const s_SKULL_ATK4 = 592
const s_SKULL_PAIN = 593
const s_SKULL_PAIN2 = 594
const s_SKULL_DIE1 = 595
const s_SKULL_DIE2 = 596
const s_SKULL_DIE3 = 597
const s_SKULL_DIE4 = 598
const s_SKULL_DIE5 = 599
const s_SKULL_DIE6 = 600
const s_SPID_STND = 601
const s_SPID_STND2 = 602
const s_SPID_RUN1 = 603
const s_SPID_RUN2 = 604
const s_SPID_RUN3 = 605
const s_SPID_RUN4 = 606
const s_SPID_RUN5 = 607
const s_SPID_RUN6 = 608
const s_SPID_RUN7 = 609
const s_SPID_RUN8 = 610
const s_SPID_RUN9 = 611
const s_SPID_RUN10 = 612
const s_SPID_RUN11 = 613
const s_SPID_RUN12 = 614
const s_SPID_ATK1 = 615
const s_SPID_ATK2 = 616
const s_SPID_ATK3 = 617
const s_SPID_ATK4 = 618
const s_SPID_PAIN = 619
const s_SPID_PAIN2 = 620
const s_SPID_DIE1 = 621
const s_SPID_DIE2 = 622
const s_SPID_DIE3 = 623
const s_SPID_DIE4 = 624
const s_SPID_DIE5 = 625
const s_SPID_DIE6 = 626
const s_SPID_DIE7 = 627
const s_SPID_DIE8 = 628
const s_SPID_DIE9 = 629
const s_SPID_DIE10 = 630
const s_SPID_DIE11 = 631
const s_BSPI_STND = 632
const s_BSPI_STND2 = 633
const s_BSPI_SIGHT = 634
const s_BSPI_RUN1 = 635
const s_BSPI_RUN2 = 636
const s_BSPI_RUN3 = 637
const s_BSPI_RUN4 = 638
const s_BSPI_RUN5 = 639
const s_BSPI_RUN6 = 640
const s_BSPI_RUN7 = 641
const s_BSPI_RUN8 = 642
const s_BSPI_RUN9 = 643
const s_BSPI_RUN10 = 644
const s_BSPI_RUN11 = 645
const s_BSPI_RUN12 = 646
const s_BSPI_ATK1 = 647
const s_BSPI_ATK2 = 648
const s_BSPI_ATK3 = 649
const s_BSPI_ATK4 = 650
const s_BSPI_PAIN = 651
const s_BSPI_PAIN2 = 652
const s_BSPI_DIE1 = 653
const s_BSPI_DIE2 = 654
const s_BSPI_DIE3 = 655
const s_BSPI_DIE4 = 656
const s_BSPI_DIE5 = 657
const s_BSPI_DIE6 = 658
const s_BSPI_DIE7 = 659
const s_BSPI_RAISE1 = 660
const s_BSPI_RAISE2 = 661
const s_BSPI_RAISE3 = 662
const s_BSPI_RAISE4 = 663
const s_BSPI_RAISE5 = 664
const s_BSPI_RAISE6 = 665
const s_BSPI_RAISE7 = 666
const s_ARACH_PLAZ = 667
const s_ARACH_PLAZ2 = 668
const s_ARACH_PLEX = 669
const s_ARACH_PLEX2 = 670
const s_ARACH_PLEX3 = 671
const s_ARACH_PLEX4 = 672
const s_ARACH_PLEX5 = 673
const s_CYBER_STND = 674
const s_CYBER_STND2 = 675
const s_CYBER_RUN1 = 676
const s_CYBER_RUN2 = 677
const s_CYBER_RUN3 = 678
const s_CYBER_RUN4 = 679
const s_CYBER_RUN5 = 680
const s_CYBER_RUN6 = 681
const s_CYBER_RUN7 = 682
const s_CYBER_RUN8 = 683
const s_CYBER_ATK1 = 684
const s_CYBER_ATK2 = 685
const s_CYBER_ATK3 = 686
const s_CYBER_ATK4 = 687
const s_CYBER_ATK5 = 688
const s_CYBER_ATK6 = 689
const s_CYBER_PAIN = 690
const s_CYBER_DIE1 = 691
const s_CYBER_DIE2 = 692
const s_CYBER_DIE3 = 693
const s_CYBER_DIE4 = 694
const s_CYBER_DIE5 = 695
const s_CYBER_DIE6 = 696
const s_CYBER_DIE7 = 697
const s_CYBER_DIE8 = 698
const s_CYBER_DIE9 = 699
const s_CYBER_DIE10 = 700
const s_PAIN_STND = 701
const s_PAIN_RUN1 = 702
const s_PAIN_RUN2 = 703
const s_PAIN_RUN3 = 704
const s_PAIN_RUN4 = 705
const s_PAIN_RUN5 = 706
const s_PAIN_RUN6 = 707
const s_PAIN_ATK1 = 708
const s_PAIN_ATK2 = 709
const s_PAIN_ATK3 = 710
const s_PAIN_ATK4 = 711
const s_PAIN_PAIN = 712
const s_PAIN_PAIN2 = 713
const s_PAIN_DIE1 = 714
const s_PAIN_DIE2 = 715
const s_PAIN_DIE3 = 716
const s_PAIN_DIE4 = 717
const s_PAIN_DIE5 = 718
const s_PAIN_DIE6 = 719
const s_PAIN_RAISE1 = 720
const s_PAIN_RAISE2 = 721
const s_PAIN_RAISE3 = 722
const s_PAIN_RAISE4 = 723
const s_PAIN_RAISE5 = 724
const s_PAIN_RAISE6 = 725
const s_SSWV_STND = 726
const s_SSWV_STND2 = 727
const s_SSWV_RUN1 = 728
const s_SSWV_RUN2 = 729
const s_SSWV_RUN3 = 730
const s_SSWV_RUN4 = 731
const s_SSWV_RUN5 = 732
const s_SSWV_RUN6 = 733
const s_SSWV_RUN7 = 734
const s_SSWV_RUN8 = 735
const s_SSWV_ATK1 = 736
const s_SSWV_ATK2 = 737
const s_SSWV_ATK3 = 738
const s_SSWV_ATK4 = 739
const s_SSWV_ATK5 = 740
const s_SSWV_ATK6 = 741
const s_SSWV_PAIN = 742
const s_SSWV_PAIN2 = 743
const s_SSWV_DIE1 = 744
const s_SSWV_DIE2 = 745
const s_SSWV_DIE3 = 746
const s_SSWV_DIE4 = 747
const s_SSWV_DIE5 = 748
const s_SSWV_XDIE1 = 749
const s_SSWV_XDIE2 = 750
const s_SSWV_XDIE3 = 751
const s_SSWV_XDIE4 = 752
const s_SSWV_XDIE5 = 753
const s_SSWV_XDIE6 = 754
const s_SSWV_XDIE7 = 755
const s_SSWV_XDIE8 = 756
const s_SSWV_XDIE9 = 757
const s_SSWV_RAISE1 = 758
const s_SSWV_RAISE2 = 759
const s_SSWV_RAISE3 = 760
const s_SSWV_RAISE4 = 761
const s_SSWV_RAISE5 = 762
const s_KEENSTND = 763
const s_COMMKEEN = 764
const s_COMMKEEN2 = 765
const s_COMMKEEN3 = 766
const s_COMMKEEN4 = 767
const s_COMMKEEN5 = 768
const s_COMMKEEN6 = 769
const s_COMMKEEN7 = 770
const s_COMMKEEN8 = 771
const s_COMMKEEN9 = 772
const s_COMMKEEN10 = 773
const s_COMMKEEN11 = 774
const s_COMMKEEN12 = 775
const s_KEENPAIN = 776
const s_KEENPAIN2 = 777
const s_BRAIN = 778
const s_BRAIN_PAIN = 779
const s_BRAIN_DIE1 = 780
const s_BRAIN_DIE2 = 781
const s_BRAIN_DIE3 = 782
const s_BRAIN_DIE4 = 783
const s_BRAINEYE = 784
const s_BRAINEYESEE = 785
const s_BRAINEYE1 = 786
const s_SPAWN1 = 787
const s_SPAWN2 = 788
const s_SPAWN3 = 789
const s_SPAWN4 = 790
const s_SPAWNFIRE1 = 791
const s_SPAWNFIRE2 = 792
const s_SPAWNFIRE3 = 793
const s_SPAWNFIRE4 = 794
const s_SPAWNFIRE5 = 795
const s_SPAWNFIRE6 = 796
const s_SPAWNFIRE7 = 797
const s_SPAWNFIRE8 = 798
const s_BRAINEXPLODE1 = 799
const s_BRAINEXPLODE2 = 800
const s_BRAINEXPLODE3 = 801
const s_ARM1 = 802
const s_ARM1A = 803
const s_ARM2 = 804
const s_ARM2A = 805
const s_BAR1 = 806
const s_BAR2 = 807
const s_BEXP = 808
const s_BEXP2 = 809
const s_BEXP3 = 810
const s_BEXP4 = 811
const s_BEXP5 = 812
const s_BBAR1 = 813
const s_BBAR2 = 814
const s_BBAR3 = 815
const s_BON1 = 816
const s_BON1A = 817
const s_BON1B = 818
const s_BON1C = 819
const s_BON1D = 820
const s_BON1E = 821
const s_BON2 = 822
const s_BON2A = 823
const s_BON2B = 824
const s_BON2C = 825
const s_BON2D = 826
const s_BON2E = 827
const s_BKEY = 828
const s_BKEY2 = 829
const s_RKEY = 830
const s_RKEY2 = 831
const s_YKEY = 832
const s_YKEY2 = 833
const s_BSKULL = 834
const s_BSKULL2 = 835
const s_RSKULL = 836
const s_RSKULL2 = 837
const s_YSKULL = 838
const s_YSKULL2 = 839
const s_STIM = 840
const s_MEDI = 841
const s_SOUL = 842
const s_SOUL2 = 843
const s_SOUL3 = 844
const s_SOUL4 = 845
const s_SOUL5 = 846
const s_SOUL6 = 847
const s_PINV = 848
const s_PINV2 = 849
const s_PINV3 = 850
const s_PINV4 = 851
const s_PSTR = 852
const s_PINS = 853
const s_PINS2 = 854
const s_PINS3 = 855
const s_PINS4 = 856
const s_MEGA = 857
const s_MEGA2 = 858
const s_MEGA3 = 859
const s_MEGA4 = 860
const s_SUIT = 861
const s_PMAP = 862
const s_PMAP2 = 863
const s_PMAP3 = 864
const s_PMAP4 = 865
const s_PMAP5 = 866
const s_PMAP6 = 867
const s_PVIS = 868
const s_PVIS2 = 869
const s_CLIP = 870
const s_AMMO = 871
const s_ROCK = 872
const s_BROK = 873
const s_CELL = 874
const s_CELP = 875
const s_SHEL = 876
const s_SBOX = 877
const s_BPAK = 878
const s_BFUG = 879
const s_MGUN = 880
const s_CSAW = 881
const s_LAUN = 882
const s_PLAS = 883
const s_SHOT = 884
const s_SHOT2 = 885
const s_COLU = 886
const s_BLOODYTWITCH = 888
const s_BLOODYTWITCH2 = 889
const s_BLOODYTWITCH3 = 890
const s_BLOODYTWITCH4 = 891
const s_HEADSONSTICK = 894
const s_GIBS = 895
const s_HEADONASTICK = 896
const s_HEADCANDLES = 897
const s_HEADCANDLES2 = 898
const s_DEADSTICK = 899
const s_LIVESTICK = 900
const s_LIVESTICK2 = 901
const s_MEAT2 = 902
const s_MEAT3 = 903
const s_MEAT4 = 904
const s_MEAT5 = 905
const s_STALAGTITE = 906
const s_TALLGRNCOL = 907
const s_SHRTGRNCOL = 908
const s_TALLREDCOL = 909
const s_SHRTREDCOL = 910
const s_CANDLESTIK = 911
const s_CANDELABRA = 912
const s_SKULLCOL = 913
const s_TORCHTREE = 914
const s_BIGTREE = 915
const s_TECHPILLAR = 916
const s_EVILEYE = 917
const s_EVILEYE2 = 918
const s_EVILEYE3 = 919
const s_EVILEYE4 = 920
const s_FLOATSKULL = 921
const s_FLOATSKULL2 = 922
const s_FLOATSKULL3 = 923
const s_HEARTCOL = 924
const s_HEARTCOL2 = 925
const s_BLUETORCH = 926
const s_BLUETORCH2 = 927
const s_BLUETORCH3 = 928
const s_BLUETORCH4 = 929
const s_GREENTORCH = 930
const s_GREENTORCH2 = 931
const s_GREENTORCH3 = 932
const s_GREENTORCH4 = 933
const s_REDTORCH = 934
const s_REDTORCH2 = 935
const s_REDTORCH3 = 936
const s_REDTORCH4 = 937
const s_BTORCHSHRT = 938
const s_BTORCHSHRT2 = 939
const s_BTORCHSHRT3 = 940
const s_BTORCHSHRT4 = 941
const s_GTORCHSHRT = 942
const s_GTORCHSHRT2 = 943
const s_GTORCHSHRT3 = 944
const s_GTORCHSHRT4 = 945
const s_RTORCHSHRT = 946
const s_RTORCHSHRT2 = 947
const s_RTORCHSHRT3 = 948
const s_RTORCHSHRT4 = 949
const s_HANGNOGUTS = 950
const s_HANGBNOBRAIN = 951
const s_HANGTLOOKDN = 952
const s_HANGTSKULL = 953
const s_HANGTLOOKUP = 954
const s_HANGTNOBRAIN = 955
const s_COLONGIBS = 956
const s_SMALLPOOL = 957
const s_BRAINSTEM = 958
const s_TECHLAMP = 959
const s_TECHLAMP2 = 960
const s_TECHLAMP3 = 961
const s_TECHLAMP4 = 962
const s_TECH2LAMP = 963
const s_TECH2LAMP2 = 964
const s_TECH2LAMP3 = 965
const s_TECH2LAMP4 = 966

type state_t struct {
	Fsprite    spritenum_t
	Fframe     int32
	Ftics      int32
	Faction    func(*mobj_t, *pspdef_t)
	Fnextstate statenum_t
	Fmisc1     int32
	Fmisc2     int32
}

type mobjtype_t = int32

const mt_PLAYER = 0
const mt_POSSESSED = 1
const mt_SHOTGUY = 2
const mt_VILE = 3
const mt_FIRE = 4
const mt_UNDEAD = 5
const mt_TRACER = 6
const mt_SMOKE = 7
const mt_FATSO = 8
const mt_FATSHOT = 9
const mt_CHAINGUY = 10
const mt_TROOP = 11
const mt_SERGEANT = 12
const mt_SHADOWS = 13
const mt_HEAD = 14
const mt_BRUISER = 15
const mt_BRUISERSHOT = 16
const mt_KNIGHT = 17
const mt_SKULL = 18
const mt_SPIDER = 19
const mt_BABY = 20
const mt_CYBORG = 21
const mt_PAIN = 22
const mt_WOLFSS = 23
const mt_BOSSTARGET = 27
const mt_SPAWNSHOT = 28
const mt_SPAWNFIRE = 29
const mt_TROOPSHOT = 31
const mt_HEADSHOT = 32
const mt_ROCKET = 33
const mt_PLASMA = 34
const mt_BFG = 35
const mt_ARACHPLAZ = 36
const mt_PUFF = 37
const mt_BLOOD = 38
const mt_TFOG = 39
const mt_IFOG = 40
const mt_TELEPORTMAN = 41
const mt_EXTRABFG = 42
const mt_INV = 56
const mt_INS = 58
const mt_CLIP = 63
const mt_CHAINGUN = 73
const mt_SHOTGUN = 77
const NUMMOBJTYPES = 137

type mobjinfo_t struct {
	Fdoomednum    int32
	Fspawnstate   int32
	Fspawnhealth  int32
	Fseestate     int32
	Fseesound     int32
	Freactiontime int32
	Fattacksound  int32
	Fpainstate    int32
	Fpainchance   int32
	Fpainsound    int32
	Fmeleestate   int32
	Fmissilestate int32
	Fdeathstate   int32
	Fxdeathstate  int32
	Fdeathsound   int32
	Fspeed        int32
	Fradius       int32
	Fheight       int32
	Fmass         int32
	Fdamage       int32
	Factivesound  int32
	Fflags        int32
	Fraisestate   int32
}

const mf_SPECIAL = 1
const mf_SOLID = 2
const mf_SHOOTABLE = 4
const mf_NOSECTOR = 8
const mf_NOBLOCKMAP = 16
const mf_AMBUSH = 32
const mf_JUSTHIT = 64
const mf_JUSTATTACKED = 128
const mf_SPAWNCEILING = 256
const mf_NOGRAVITY = 512
const mf_DROPOFF = 1024
const mf_PICKUP = 2048
const mf_NOCLIP = 4096
const mf_FLOAT = 16384
const mf_TELEPORT = 32768
const mf_MISSILE = 65536
const mf_DROPPED = 131072
const mf_SHADOW = 262144
const mf_NOBLOOD = 524288
const mf_CORPSE = 1048576
const mf_INFLOAT = 2097152
const mf_COUNTKILL = 4194304
const mf_COUNTITEM = 8388608
const mf_SKULLFLY = 16777216
const mf_NOTDMATCH = 33554432
const mf_TRANSLATION = 201326592
const mf_TRANSSHIFT = 26

type mobj_t struct {
	degenmobj_t
	Fsnext        *mobj_t
	Fsprev        *mobj_t
	Fangle        angle_t
	Fsprite       spritenum_t
	Fframe        int32
	Fbnext        *mobj_t
	Fbprev        *mobj_t
	Fsubsector    *subsector_t
	Ffloorz       fixed_t
	Fceilingz     fixed_t
	Fradius       fixed_t
	Fheight       fixed_t
	Fmomx         fixed_t
	Fmomy         fixed_t
	Fmomz         fixed_t
	Fvalidcount   int32
	Ftype1        mobjtype_t
	Finfo         *mobjinfo_t
	Ftics         int32
	Fstate        *state_t
	Fflags        int32
	Fhealth       int32
	Fmovedir      int32
	Fmovecount    int32
	Ftarget       *mobj_t
	Freactiontime int32
	Fthreshold    int32
	Fplayer       *player_t
	Flastlook     int32
	Fspawnpoint   mapthing_t
	Ftracer       *mobj_t
}

type patch_t struct {
	Fwidth      int16
	Fheight     int16
	Fleftoffset int16
	Ftopoffset  int16
	// TODO: This is a bit of a lie, as this array is really of Fwidth in length.
	// but we don't have a way to express that in Go, as this data is loaded directly
	// from the lump data
	Fcolumnofs [320]int32
}

func (p *patch_t) GetColumn(i int32) *column_t {
	if i < 0 || i >= int32(p.Fwidth) {
		panic("GetColumn: index out of bounds")
	}
	return (*column_t)(unsafe.Pointer((uintptr)(unsafe.Pointer(p)) + uintptr(p.Fcolumnofs[i])))
}

type column_t struct {
	Ftopdelta uint8
	Flength   uint8
}

func (c *column_t) Next() *column_t {
	return (*column_t)(unsafe.Pointer(uintptr(unsafe.Pointer(c)) + uintptr(c.Flength+4)))
}

func (c *column_t) Data() []byte {
	source := (uintptr)(unsafe.Pointer(c)) + uintptr(3)
	return unsafe.Slice((*byte)(unsafe.Pointer(source)), c.Flength)
}

type vertex_t struct {
	Fx fixed_t
	Fy fixed_t
}

const st_HORIZONTAL = 0
const st_VERTICAL = 1
const st_POSITIVE = 2
const st_NEGATIVE = 3

type degenmobj_t struct {
	Fthinker thinker_t
	Fx       fixed_t
	Fy       fixed_t
	Fz       fixed_t
}

type sector_t struct {
	Ffloorheight    fixed_t
	Fceilingheight  fixed_t
	Ffloorpic       int16
	Fceilingpic     int16
	Flightlevel     int16
	Fspecial        int16
	Ftag            int16
	Fsoundtraversed int32
	Fsoundtarget    *mobj_t
	Fblockbox       [4]int32
	Fsoundorg       degenmobj_t
	Fvalidcount     int32
	Fthinglist      *mobj_t
	Fspecialdata    any
	Flinecount      int32
	Flines          []*line_t
}

type side_t struct {
	Ftextureoffset fixed_t
	Frowoffset     fixed_t
	Ftoptexture    int16
	Fbottomtexture int16
	Fmidtexture    int16
	Fsector        *sector_t
}

type slopetype_t int32

type box_t [4]fixed_t

type line_t struct {
	Fv1          *vertex_t
	Fv2          *vertex_t
	Fdx          fixed_t
	Fdy          fixed_t
	Fflags       int16
	Fspecial     int16
	Ftag         int16
	Fsidenum     [2]int16
	Fbbox        box_t
	Fslopetype   slopetype_t
	Ffrontsector *sector_t
	Fbacksector  *sector_t
	Fvalidcount  int32
	Fspecialdata any
}

type subsector_t struct {
	Fsector    *sector_t
	Fnumlines  int16
	Ffirstline int16
}

type seg_t struct {
	Fv1          *vertex_t
	Fv2          *vertex_t
	Foffset      fixed_t
	Fangle       angle_t
	Fsidedef     *side_t
	Flinedef     *line_t
	Ffrontsector *sector_t
	Fbacksector  *sector_t
}

type node_t struct {
	divline_t
	Fbbox     [2]box_t
	Fchildren [2]uint16
}

type lighttable_t = uint8

type drawseg_t struct {
	Fcurline          *seg_t
	Fx1               int32
	Fx2               int32
	Fscale1           fixed_t
	Fscale2           fixed_t
	Fscalestep        fixed_t
	Fsilhouette       int32
	Fbsilheight       fixed_t
	Ftsilheight       fixed_t
	Fsprtopclip       uintptr
	Fsprbottomclip    uintptr
	Fmaskedtexturecol uintptr
}

type vissprite_t struct {
	Fprev       *vissprite_t
	Fnext       *vissprite_t
	Fx1         int32
	Fx2         int32
	Fgx         fixed_t
	Fgy         fixed_t
	Fgz         fixed_t
	Fgzt        fixed_t
	Fstartfrac  fixed_t
	Fscale      fixed_t
	Fxiscale    fixed_t
	Ftexturemid fixed_t
	Fpatch      int32
	Fcolormap   []lighttable_t
	Fmobjflags  int32
}

type spriteframe_t struct {
	Frotate boolean
	Flump   [8]int16
	Fflip   [8]uint8
}

type spritedef_t struct {
	Fnumframes    int32
	Fspriteframes []spriteframe_t
}

type visplane_t struct {
	Fheight     fixed_t
	Fpicnum     int32
	Flightlevel int32
	Fminx       int32
	Fmaxx       int32
	Fpad1       uint8
	Ftop        [320]uint8
	Fpad2       uint8
	Fpad3       uint8
	Fbottom     [320]uint8
	Fpad4       uint8
}

type weaponinfo_t struct {
	Fammo       ammotype_t
	Fupstate    int32
	Fdownstate  int32
	Freadystate int32
	Fatkstate   int32
	Fflashstate int32
}

const ps_weapon = 0
const ps_flash = 1
const NUMPSPRITES = 2

type pspdef_t struct {
	Fstate *state_t
	Ftics  int32
	Fsx    fixed_t
	Fsy    fixed_t
}

type ticcmd_t struct {
	Fforwardmove int8
	Fsidemove    int8
	Fangleturn   int16
	Fchatchar    uint8
	Fbuttons     uint8
	Fconsistancy uint8
	Fbuttons2    uint8
	Finventory   int32
	Flookfly     uint8
	Farti        uint8
}

type net_connect_data_t struct {
	Fgamemode     gamemode_t
	Fgamemission  gamemission_t
	Flowres_turn  int32
	Fdrone        int32
	Fmax_players  int32
	Fis_freedoom  int32
	Fwad_sha1sum  sha1_digest_t
	Fdeh_sha1sum  sha1_digest_t
	Fplayer_class int32
}

type net_gamesettings_t struct {
	Fticdup           int32
	Fextratics        int32
	Fdeathmatch       int32
	Fepisode          int32
	Fnomonsters       int32
	Ffast_monsters    int32
	Frespawn_monsters int32
	Fmap1             int32
	Fskill            skill_t
	Fgameversion      gameversion_t
	Flowres_turn      int32
	Fnew_sync         int32
	Ftimelimit        int32
	Floadgame         int32
	Frandom           int32
	Fnum_players      int32
	Fconsoleplayer    int32
	Fplayer_classes   [8]int32
}

type playerstate_t = int32

const Pst_LIVE = 0
const Pst_DEAD = 1
const Pst_REBORN = 2

const CF_NOCLIP = 1
const CF_GODMODE = 2
const CF_NOMOMENTUM = 4

type player_t struct {
	Fmo              *mobj_t
	Fplayerstate     playerstate_t
	Fcmd             ticcmd_t
	Fviewz           fixed_t
	Fviewheight      fixed_t
	Fdeltaviewheight fixed_t
	Fbob             fixed_t
	Fhealth          int32
	Farmorpoints     int32
	Farmortype       int32
	Fpowers          [6]int32
	Fcards           [6]boolean
	Fbackpack        boolean
	Ffrags           [4]int32
	Freadyweapon     weapontype_t
	Fpendingweapon   weapontype_t
	Fweaponowned     [9]boolean
	Fammo            [4]int32
	Fmaxammo         [4]int32
	Fattackdown      int32
	Fusedown         int32
	Fcheats          int32
	Frefire          int32
	Fkillcount       int32
	Fitemcount       int32
	Fsecretcount     int32
	Fmessage         string
	Fdamagecount     int32
	Fbonuscount      int32
	Fattacker        *mobj_t
	Fextralight      int32
	Ffixedcolormap   int32
	Fcolormap        int32
	Fpsprites        [2]pspdef_t
	Fdidsecret       boolean
}

type wbplayerstruct_t struct {
	Fin      boolean
	Fskills  int32
	Fsitems  int32
	Fssecret int32
	Fstime   int32
	Ffrags   [4]int32
	Fscore   int32
}

type wbstartstruct_t struct {
	Fepsd      int32
	Fdidsecret boolean
	Flast      int32
	Fnext      int32
	Fmaxkills  int32
	Fmaxitems  int32
	Fmaxsecret int32
	Fmaxfrags  int32
	Fpartime   int32
	Fpnum      int32
	Fplyr      [4]wbplayerstruct_t
}

type divline_t struct {
	Fx  fixed_t
	Fy  fixed_t
	Fdx fixed_t
	Fdy fixed_t
}

type intercept_t struct {
	Ffrac    fixed_t
	Fisaline boolean
	Fd       struct {
		Fthing any
	}
}

type fireflicker_t struct {
	Fthinker  thinker_t
	Fsector   *sector_t
	Fcount    int32
	Fmaxlight int32
	Fminlight int32
}

type lightflash_t struct {
	Fthinker  thinker_t
	Fsector   *sector_t
	Fcount    int32
	Fmaxlight int32
	Fminlight int32
	Fmaxtime  int32
	Fmintime  int32
}

type strobe_t struct {
	Fthinker    thinker_t
	Fsector     *sector_t
	Fcount      int32
	Fminlight   int32
	Fmaxlight   int32
	Fdarktime   int32
	Fbrighttime int32
}

type glow_t struct {
	Fthinker   thinker_t
	Fsector    *sector_t
	Fminlight  int32
	Fmaxlight  int32
	Fdirection int32
}

type switchlist_t struct {
	Fname1   string
	Fname2   string
	Fepisode int16
}

type bwhere_e = int32

const top = 0
const middle = 1
const bottom = 2

type button_t struct {
	Fline     *line_t
	Fwhere    bwhere_e
	Fbtexture int32
	Fbtimer   int32
	Fsoundorg *degenmobj_t
}

type plat_e = int32

const up = 0
const down = 1
const waiting = 2
const in_stasis = 3

type plattype_e = int32

const perpetualRaise = 0
const downWaitUpStay = 1
const raiseAndChange = 2
const raiseToNearestAndChange = 3
const blazeDWUS = 4

type plat_t struct {
	Fthinker   thinker_t
	Fsector    *sector_t
	Fspeed     fixed_t
	Flow       fixed_t
	Fhigh      fixed_t
	Fwait      int32
	Fcount     int32
	Fstatus    plat_e
	Foldstatus plat_e
	Fcrush     boolean
	Ftag       int32
	Ftype1     plattype_e
}

type vldoor_e = int32

const vld_normal = 0
const vld_close30ThenOpen = 1
const vld_close = 2
const vld_open = 3
const vld_raiseIn5Mins = 4
const vld_blazeRaise = 5
const vld_blazeOpen = 6
const vld_blazeClose = 7

type vldoor_t struct {
	Fthinker      thinker_t
	Ftype1        vldoor_e
	Fsector       *sector_t
	Ftopheight    fixed_t
	Fspeed        fixed_t
	Fdirection    int32
	Ftopwait      int32
	Ftopcountdown int32
}

type ceiling_e = int32

const lowerToFloor = 0
const raiseToHighest = 1
const lowerAndCrush = 2
const crushAndRaise = 3
const fastCrushAndRaise = 4
const silentCrushAndRaise = 5

type ceiling_t struct {
	Fthinker      thinker_t
	Ftype1        ceiling_e
	Fsector       *sector_t
	Fbottomheight fixed_t
	Ftopheight    fixed_t
	Fspeed        fixed_t
	Fcrush        boolean
	Fdirection    int32
	Ftag          int32
	Folddirection int32
}

type floor_e = int32

const lowerFloor = 0
const lowerFloorToLowest = 1
const turboLower = 2
const raiseFloor = 3
const raiseFloorToNearest = 4
const raiseToTexture = 5
const lowerAndChange = 6
const raiseFloor24 = 7
const raiseFloor24AndChange = 8
const raiseFloorCrush = 9
const raiseFloorTurbo = 10
const donutRaise = 11
const raiseFloor512 = 12

type stair_e = int32

const build8 = 0
const turbo16 = 1

type floormove_t struct {
	Fthinker         thinker_t
	Ftype1           floor_e
	Fcrush           boolean
	Fsector          *sector_t
	Fdirection       int32
	Fnewspecial      int32
	Ftexture         int16
	Ffloordestheight fixed_t
	Fspeed           fixed_t
}

type result_e = int32

const ok = 0
const crushed = 1
const pastdest = 2

type lumpinfo_t struct {
	Fname     [8]byte
	Fwad_file *os.File
	Fposition int32
	Fsize     int32
	Fcache    []byte
	Fnext     *lumpinfo_t
}

func (l *lumpinfo_t) Name() string {
	for i := 0; i < len(l.Fname); i++ {
		if l.Fname[i] == 0 {
			return string(l.Fname[:i])
		}
	}
	return string(l.Fname[:])
}

type netgame_startup_callback_t = uintptr

type loop_interface_t struct {
	FProcessEvents func()
	FBuildTiccmd   func(*ticcmd_t, int32)
	FRunTic        func([]ticcmd_t, []boolean)
	FRunMenu       func()
}

// For use if I do walls with outsides/insides

// Automap colors

// drawing stuff

// scale on entry
// how much the automap moves window per tic in frame-buffer coordinates
// moves 140 pixels in 1 second
// how much zoom-in per tic
// goes to 2x in 1 second
// how much zoom-out per tic
// pulls out to 0.5x in 1 second

// translates between frame-buffer and map distances
// translates between frame-buffer and map coordinates

// the following is crap

type fpoint_t struct {
	Fx int32
	Fy int32
}

type fline_t struct {
	Fa fpoint_t
	Fb fpoint_t
}

type mpoint_t struct {
	Fx fixed_t
	Fy fixed_t
}

type mline_t struct {
	Fa mpoint_t
	Fb mpoint_t
}

func init() {
	player_arrow = [7]mline_t{
		0: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
			},
			Fb: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7,
			},
		},
		1: {
			Fa: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7,
			},
			Fb: mpoint_t{
				Fx: 8*16*(1<<FRACBITS)/7 - 8*16*(1<<FRACBITS)/7/2,
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 4,
			},
		},
		2: {
			Fa: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7,
			},
			Fb: mpoint_t{
				Fx: 8*16*(1<<FRACBITS)/7 - 8*16*(1<<FRACBITS)/7/2,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 4,
			},
		},
		3: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) - 8*16*(1<<FRACBITS)/7/8,
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 4,
			},
		},
		4: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) - 8*16*(1<<FRACBITS)/7/8,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 4,
			},
		},
		5: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 3*(8*16*(1<<FRACBITS)/7)/8,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 4,
			},
		},
		6: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 3*(8*16*(1<<FRACBITS)/7)/8,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 4,
			},
		},
	}

	cheat_player_arrow = [16]mline_t{
		0: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
			},
			Fb: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7,
			},
		},
		1: {
			Fa: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7,
			},
			Fb: mpoint_t{
				Fx: 8*16*(1<<FRACBITS)/7 - 8*16*(1<<FRACBITS)/7/2,
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 6,
			},
		},
		2: {
			Fa: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7,
			},
			Fb: mpoint_t{
				Fx: 8*16*(1<<FRACBITS)/7 - 8*16*(1<<FRACBITS)/7/2,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
		},
		3: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) - 8*16*(1<<FRACBITS)/7/8,
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 6,
			},
		},
		4: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) - 8*16*(1<<FRACBITS)/7/8,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
		},
		5: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 3*(8*16*(1<<FRACBITS)/7)/8,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 6,
			},
		},
		6: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 3*(8*16*(1<<FRACBITS)/7)/8,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) + 8*16*(1<<FRACBITS)/7/8,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
		},
		7: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) / 2,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) / 2,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
		},
		8: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) / 2,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
			Fb: mpoint_t{
				Fx: -(8*16*(1<<FRACBITS)/7)/2 + 8*16*(1<<FRACBITS)/7/6,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
		},
		9: {
			Fa: mpoint_t{
				Fx: -(8*16*(1<<FRACBITS)/7)/2 + 8*16*(1<<FRACBITS)/7/6,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
			Fb: mpoint_t{
				Fx: -(8*16*(1<<FRACBITS)/7)/2 + 8*16*(1<<FRACBITS)/7/6,
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 4,
			},
		},
		10: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
			Fb: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
		},
		11: {
			Fa: mpoint_t{
				Fx: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
			Fb: mpoint_t{
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
		},
		12: {
			Fa: mpoint_t{
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 6,
			},
			Fb: mpoint_t{
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 4,
			},
		},
		13: {
			Fa: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7 / 6,
				Fy: 8 * 16 * (1 << FRACBITS) / 7 / 4,
			},
			Fb: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7 / 6,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 7,
			},
		},
		14: {
			Fa: mpoint_t{
				Fx: 8 * 16 * (1 << FRACBITS) / 7 / 6,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 7,
			},
			Fb: mpoint_t{
				Fx: 8*16*(1<<FRACBITS)/7/6 + 8*16*(1<<FRACBITS)/7/32,
				Fy: -(8*16*(1<<FRACBITS)/7)/7 - 8*16*(1<<FRACBITS)/7/32,
			},
		},
		15: {
			Fa: mpoint_t{
				Fx: 8*16*(1<<FRACBITS)/7/6 + 8*16*(1<<FRACBITS)/7/32,
				Fy: -(8*16*(1<<FRACBITS)/7)/7 - 8*16*(1<<FRACBITS)/7/32,
			},
			Fb: mpoint_t{
				Fx: 8*16*(1<<FRACBITS)/7/6 + 8*16*(1<<FRACBITS)/7/10,
				Fy: -(8 * 16 * (1 << FRACBITS) / 7) / 7,
			},
		},
	}

	thintriangle_guy = [3]mline_t{
		0: {
			Fa: mpoint_t{
				Fx: float2fixed(-0.5),
				Fy: float2fixed(-0.7),
			},
			Fb: mpoint_t{
				Fx: 1 << FRACBITS,
			},
		},
		1: {
			Fa: mpoint_t{
				Fx: 1 << FRACBITS,
			},
			Fb: mpoint_t{
				Fx: float2fixed(-0.5),
				Fy: float2fixed(0.7),
			},
		},
		2: {
			Fa: mpoint_t{
				Fx: float2fixed(-0.5),
				Fy: float2fixed(0.7),
			},
			Fb: mpoint_t{
				Fx: float2fixed(-0.5),
				Fy: float2fixed(-0.7),
			},
		},
	}
}

var cheating int32 = 0
var grid bool

// C documentation
//
//	// location of window on screen
var f_x int32
var f_y int32

// C documentation
//
//	// size of window on screen
var f_w int32
var f_h int32

var lightlev int32 // used for funky strobing effect
var fb []byte      // pseudo-frame buffer

var m_paninc mpoint_t    // how far the window pans each tic (map coords)
var mtof_zoommul fixed_t // how far the window zooms in each tic (map coords)
var ftom_zoommul fixed_t // how far the window zooms in each tic (fb coords)

var m_x fixed_t
var m_y fixed_t // LL x,y where the window is on the map (map coords)
var m_x2 fixed_t
var m_y2 fixed_t // UR x,y where the window is on the map (map coords)

// C documentation
//
//	//
//	// width/height of window on map (map coords)
//	//
var m_w fixed_t
var m_h fixed_t

// C documentation
//
//	// based on level size
var min_x fixed_t
var min_y fixed_t
var max_x fixed_t
var max_y fixed_t

var max_w fixed_t // max_x-min_x,
var max_h fixed_t // max_y-min_y

// C documentation
//

var min_scale_mtof fixed_t // used to tell when to stop zooming out
var max_scale_mtof fixed_t // used to tell when to stop zooming in

// C documentation
//
//	// old stuff for recovery later
var old_m_w fixed_t
var old_m_h fixed_t
var old_m_x fixed_t
var old_m_y fixed_t

// C documentation
//
//	// old location used by the Follower routine
var f_oldloc mpoint_t

// C documentation
//
//	// used by MTOF to scale from map-to-frame-buffer coords
var scale_mtof = float2fixed(0.2)

// C documentation
//
//	// used by FTOM to scale from frame-buffer-to-map coords (=1/scale_mtof)
var scale_ftom fixed_t

var plr *player_t // the player represented by an arrow

var marknums [10]*patch_t   // numbers used for marking by the automap
var markpoints [10]mpoint_t // where the points are
var markpointnum int32 = 0  // next point to be assigned

var followplayer int32 = 1 // specifies whether to follow the player around

func init() {
	cheat_amap = cheatseq_t{
		Fsequence:      "iddt",
		Fsequence_len:  5 - 1,
		Fparameter_buf: [5]byte{},
	}
}

var stopped int32 = 1

// C documentation
//
//	//
//	//
//	//
func am_activateNewScale() {
	m_x += m_w / 2
	m_y += m_h / 2
	m_w = fixedMul(f_w<<16, scale_ftom)
	m_h = fixedMul(f_h<<16, scale_ftom)
	m_x -= m_w / 2
	m_y -= m_h / 2
	m_x2 = m_x + m_w
	m_y2 = m_y + m_h
}

// C documentation
//
//	//
//	//
//	//
func am_saveScaleAndLoc() {
	old_m_x = m_x
	old_m_y = m_y
	old_m_w = m_w
	old_m_h = m_h
}

// C documentation
//
//	//
//	//
//	//
func am_restoreScaleAndLoc() {
	m_w = old_m_w
	m_h = old_m_h
	if followplayer == 0 {
		m_x = old_m_x
		m_y = old_m_y
	} else {
		m_x = plr.Fmo.Fx - m_w/2
		m_y = plr.Fmo.Fy - m_h/2
	}
	m_x2 = m_x + m_w
	m_y2 = m_y + m_h
	// Change the scaling multipliers
	scale_mtof = fixedDiv(f_w<<FRACBITS, m_w)
	scale_ftom = fixedDiv(1<<FRACBITS, scale_mtof)
}

// C documentation
//
//	//
//	// adds a marker at the current location
//	//
func am_addMark() {
	markpoints[markpointnum].Fx = m_x + m_w/2
	markpoints[markpointnum].Fy = m_y + m_h/2
	markpointnum = (markpointnum + 1) % AM_NUMMARKPOINTS
}

// C documentation
//
//	//
//	// Determines bounding box of all vertices,
//	// sets global variables controlling zoom range.
//	//
func am_findMinMaxBoundaries() {
	var a, b, v1, v2 fixed_t
	var v4 int32
	v1 = INT_MAX1
	min_y = v1
	min_x = v1
	v2 = -INT_MAX1
	max_y = v2
	max_x = v2
	for i := int32(0); i < numvertexes; i++ {
		if vertexes[i].Fx < min_x {
			min_x = vertexes[i].Fx
		} else {
			if vertexes[i].Fx > max_x {
				max_x = vertexes[i].Fx
			}
		}
		if vertexes[i].Fy < min_y {
			min_y = vertexes[i].Fy
		} else {
			if vertexes[i].Fy > max_y {
				max_y = vertexes[i].Fy
			}
		}
	}
	max_w = max_x - min_x
	max_h = max_y - min_y
	a = fixedDiv(f_w<<FRACBITS, max_w)
	b = fixedDiv(f_h<<FRACBITS, max_h)
	if a < b {
		v4 = a
	} else {
		v4 = b
	}
	min_scale_mtof = v4
	max_scale_mtof = fixedDiv(f_h<<FRACBITS, 2*16*(1<<FRACBITS))
}

// C documentation
//
//	//
//	//
//	//
func am_changeWindowLoc() {
	if m_paninc.Fx != 0 || m_paninc.Fy != 0 {
		followplayer = 0
		f_oldloc.Fx = int32(INT_MAX1)
	}
	m_x += m_paninc.Fx
	m_y += m_paninc.Fy
	if m_x+m_w/2 > max_x {
		m_x = max_x - m_w/2
	} else {
		if m_x+m_w/2 < min_x {
			m_x = min_x - m_w/2
		}
	}
	if m_y+m_h/2 > max_y {
		m_y = max_y - m_h/2
	} else {
		if m_y+m_h/2 < min_y {
			m_y = min_y - m_h/2
		}
	}
	m_x2 = m_x + m_w
	m_y2 = m_y + m_h
}

// C documentation
//
//	//
//	//
//	//
func am_initVariables() {
	var v1 fixed_t
	automapactive = 1
	fb = I_VideoBuffer
	f_oldloc.Fx = int32(INT_MAX1)
	lightlev = 0
	v1 = 0
	m_paninc.Fy = v1
	m_paninc.Fx = v1
	ftom_zoommul = 1 << FRACBITS
	mtof_zoommul = 1 << FRACBITS
	m_w = fixedMul(f_w<<16, scale_ftom)
	m_h = fixedMul(f_h<<16, scale_ftom)
	// find player to center on initially
	if playeringame[consoleplayer] != 0 {
		plr = &players[consoleplayer]
	} else {
		plr = &players[0]
		for pnum := 0; pnum < MAXPLAYERS; pnum++ {
			if playeringame[pnum] != 0 {
				plr = &players[pnum]
				break
			}
		}
	}
	m_x = plr.Fmo.Fx - m_w/2
	m_y = plr.Fmo.Fy - m_h/2
	am_changeWindowLoc()
	// for saving & restoring
	old_m_x = m_x
	old_m_y = m_y
	old_m_w = m_w
	old_m_h = m_h
	// inform the status bar of the change
	st_Responder(&st_notify)
}

var st_notify = event_t{
	Ftype1: Ev_keyup,
	Fdata1: 'a'<<24 + 'm'<<16 | 'e'<<8,
}

// C documentation
//
//	//
//	//
//	//
func am_loadPics() {
	for i := range 10 {
		name := fmt.Sprintf("AMMNUM%d", i)
		marknums[i] = w_CacheLumpNameT(name)
	}
}

func am_unloadPics() {
	for i := range 10 {
		name := fmt.Sprintf("AMMNUM%d", i)
		w_ReleaseLumpName(name)
	}
}

func am_clearMarks() {
	for i := range AM_NUMMARKPOINTS {
		markpoints[i].Fx = -1
	}
	markpointnum = 0
}

// C documentation
//
//	//
//	// should be called at the start of every level
//	// right now, i figure it out myself
//	//
func am_LevelInit() {
	f_y = 0
	f_x = 0
	f_w = SCREENWIDTH
	f_h = SCREENHEIGHT - 32
	am_clearMarks()
	am_findMinMaxBoundaries()
	scale_mtof = fixedDiv(min_scale_mtof, float2fixed(0.7))
	if scale_mtof > max_scale_mtof {
		scale_mtof = min_scale_mtof
	}
	scale_ftom = fixedDiv(1<<FRACBITS, scale_mtof)
}

// C documentation
//
//	//
//	//
//	//
func am_Stop() {
	am_unloadPics()
	automapactive = 0
	st_Responder(&st_notify1)
	stopped = 1
}

var st_notify1 = event_t{
	Fdata1: int32(Ev_keyup),
	Fdata2: 'a'<<24 + 'm'<<16 | 'x'<<8,
}

// C documentation
//
//	//
//	//
//	//
func am_Start() {
	if stopped == 0 {
		am_Stop()
	}
	stopped = 0
	if lastlevel != gamemap || lastepisode != gameepisode {
		am_LevelInit()
		lastlevel = gamemap
		lastepisode = gameepisode
	}
	am_initVariables()
	am_loadPics()
}

var lastlevel int32 = -1

var lastepisode int32 = -1

// C documentation
//
//	//
//	// set the window scale to the maximum size
//	//
func am_minOutWindowScale() {
	scale_mtof = min_scale_mtof
	scale_ftom = fixedDiv(1<<FRACBITS, scale_mtof)
	am_activateNewScale()
}

// C documentation
//
//	//
//	// set the window scale to the minimum size
//	//
func am_maxOutWindowScale() {
	scale_mtof = max_scale_mtof
	scale_ftom = fixedDiv(1<<FRACBITS, scale_mtof)
	am_activateNewScale()
}

// C documentation
//
//	//
//	// Handle events (user inputs) in automap mode
//	//
func am_Responder(ev *event_t) boolean {
	var key, rc int32
	rc = 0
	if automapactive == 0 {
		if ev.Ftype1 == Ev_keydown && ev.Fdata1 == key_map_toggle {
			am_Start()
			viewactive = 0
			rc = 1
		}
	} else {
		if ev.Ftype1 == Ev_keydown {
			rc = 1
			key = ev.Fdata1
			if key == key_map_east { // pan right
				if followplayer == 0 {
					m_paninc.Fx = fixedMul(F_PANINC<<16, scale_ftom)
				} else {
					rc = 0
				}
			} else {
				if key == key_map_west { // pan left
					if followplayer == 0 {
						m_paninc.Fx = -fixedMul(F_PANINC<<16, scale_ftom)
					} else {
						rc = 0
					}
				} else {
					if key == key_map_north { // pan up
						if followplayer == 0 {
							m_paninc.Fy = fixedMul(F_PANINC<<16, scale_ftom)
						} else {
							rc = 0
						}
					} else {
						if key == key_map_south { // pan down
							if followplayer == 0 {
								m_paninc.Fy = -fixedMul(F_PANINC<<16, scale_ftom)
							} else {
								rc = 0
							}
						} else {
							if key == key_map_zoomout { // zoom out
								mtof_zoommul = float2fixedinv(1.02)
								ftom_zoommul = float2fixed(1.02)
							} else {
								if key == key_map_zoomin { // zoom in
									mtof_zoommul = float2fixed(1.02)
									ftom_zoommul = float2fixedinv(1.02)
								} else {
									if key == key_map_toggle {
										bigstate = 0
										viewactive = 1
										am_Stop()
									} else {
										if key == key_map_maxzoom {
											bigstate = boolint32(bigstate == 0)
											if bigstate != 0 {
												am_saveScaleAndLoc()
												am_minOutWindowScale()
											} else {
												am_restoreScaleAndLoc()
											}
										} else {
											if key == key_map_follow {
												followplayer = boolint32(followplayer == 0)
												f_oldloc.Fx = int32(INT_MAX1)
												if followplayer != 0 {
													plr.Fmessage = "Follow Mode ON"
												} else {
													plr.Fmessage = "Follow Mode OFF"
												}
											} else {
												if key == key_map_grid {
													grid = !grid
													if grid {
														plr.Fmessage = "Grid ON"
													} else {
														plr.Fmessage = "Grid OFF"
													}
												} else {
													if key == key_map_mark {
														plr.Fmessage = fmt.Sprintf("%s %d", "Marked Spot", markpointnum)
														am_addMark()
													} else {
														if key == key_map_clearmark {
															am_clearMarks()
															plr.Fmessage = "All Marks Cleared"
														} else {
															rc = 0
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
			if deathmatch == 0 && cht_CheckCheat(&cheat_amap, int8(ev.Fdata2)) != 0 {
				rc = 0
				cheating = (cheating + 1) % 3
			}
		} else {
			if ev.Ftype1 == Ev_keyup {
				rc = 0
				key = ev.Fdata1
				if key == key_map_east {
					if followplayer == 0 {
						m_paninc.Fx = 0
					}
				} else {
					if key == key_map_west {
						if followplayer == 0 {
							m_paninc.Fx = 0
						}
					} else {
						if key == key_map_north {
							if followplayer == 0 {
								m_paninc.Fy = 0
							}
						} else {
							if key == key_map_south {
								if followplayer == 0 {
									m_paninc.Fy = 0
								}
							} else {
								if key == key_map_zoomout || key == key_map_zoomin {
									mtof_zoommul = 1 << FRACBITS
									ftom_zoommul = 1 << FRACBITS
								}
							}
						}
					}
				}
			}
		}
	}
	return uint32(rc)
}

var bigstate int32

// C documentation
//
//	//
//	// Zooming
//	//
func am_changeWindowScale() {
	// Change the scaling multipliers
	scale_mtof = fixedMul(scale_mtof, mtof_zoommul)
	scale_ftom = fixedDiv(1<<FRACBITS, scale_mtof)
	if scale_mtof < min_scale_mtof {
		am_minOutWindowScale()
	} else {
		if scale_mtof > max_scale_mtof {
			am_maxOutWindowScale()
		} else {
			am_activateNewScale()
		}
	}
}

// C documentation
//
//	//
//	//
//	//
func am_doFollowPlayer() {
	if f_oldloc.Fx != plr.Fmo.Fx || f_oldloc.Fy != plr.Fmo.Fy {
		m_x = fixedMul(fixedMul(plr.Fmo.Fx, scale_mtof)>>int32(16)<<int32(16), scale_ftom) - m_w/2
		m_y = fixedMul(fixedMul(plr.Fmo.Fy, scale_mtof)>>int32(16)<<int32(16), scale_ftom) - m_h/2
		m_x2 = m_x + m_w
		m_y2 = m_y + m_h
		f_oldloc.Fx = plr.Fmo.Fx
		f_oldloc.Fy = plr.Fmo.Fy
		//  m_x = FTOM(MTOF(plr->mo->x - m_w/2));
		//  m_y = FTOM(MTOF(plr->mo->y - m_h/2));
		//  m_x = plr->mo->x - m_w/2;
		//  m_y = plr->mo->y - m_h/2;
	}
}

// C documentation
//
//	//
//	// Updates on Game Tick
//	//
func am_Ticker() {
	if automapactive == 0 {
		return
	}
	if followplayer != 0 {
		am_doFollowPlayer()
	}
	// Change the zoom if necessary
	if ftom_zoommul != 1<<FRACBITS {
		am_changeWindowScale()
	}
	// Change x,y location
	if m_paninc.Fx != 0 || m_paninc.Fy != 0 {
		am_changeWindowLoc()
	}
	// Update light level
	// AM_updateLightLev();
}

// C documentation
//
//	//
//	// Clear automap frame buffer.
//	//
func am_clearFB(color uint8) {
	for i := int32(0); i < f_w*f_h; i++ {
		fb[i] = color
	}
}

// C documentation
//
//	//
//	// Automap clipping of lines.
//	//
//	// Based on Cohen-Sutherland clipping algorithm but with a slightly
//	// faster reject and precalculated slopes.  If the speed is needed,
//	// use a hash algorithm to handle  the common cases.
//	//
func am_clipMline(ml *mline_t, fl *fline_t) boolean {
	var dx, dy, outcode1, outcode2, outside int32
	var tmp fpoint_t
	outcode1 = 0
	outcode2 = 0
	// do trivial rejects and outcodes
	if ml.Fa.Fy > m_y2 {
		outcode1 = 8
	} else {
		if ml.Fa.Fy < m_y {
			outcode1 = 4
		}
	}
	if ml.Fb.Fy > m_y2 {
		outcode2 = 8
	} else {
		if ml.Fb.Fy < m_y {
			outcode2 = 4
		}
	}
	if outcode1&outcode2 != 0 {
		return 0
	} // trivially outside
	if ml.Fa.Fx < m_x {
		outcode1 |= 1
	} else {
		if ml.Fa.Fx > m_x2 {
			outcode1 |= 2
		}
	}
	if ml.Fb.Fx < m_x {
		outcode2 |= 1
	} else {
		if ml.Fb.Fx > m_x2 {
			outcode2 |= 2
		}
	}
	if outcode1&outcode2 != 0 {
		return 0
	} // trivially outside
	// transform to frame-buffer coordinates.
	fl.Fa.Fx = f_x + fixedMul(ml.Fa.Fx-m_x, scale_mtof)>>16
	fl.Fa.Fy = f_y + (f_h - fixedMul(ml.Fa.Fy-m_y, scale_mtof)>>16)
	fl.Fb.Fx = f_x + fixedMul(ml.Fb.Fx-m_x, scale_mtof)>>16
	fl.Fb.Fy = f_y + (f_h - fixedMul(ml.Fb.Fy-m_y, scale_mtof)>>16)
	outcode1 = 0
	if fl.Fa.Fy < 0 {
		outcode1 |= 8
	} else {
		if fl.Fa.Fy >= f_h {
			outcode1 |= 4
		}
	}
	if fl.Fa.Fx < 0 {
		outcode1 |= 1
	} else {
		if fl.Fa.Fx >= f_w {
			outcode1 |= 2
		}
	}
	outcode2 = 0
	if fl.Fb.Fy < 0 {
		outcode2 |= 8
	} else {
		if fl.Fb.Fy >= f_h {
			outcode2 |= 4
		}
	}
	if fl.Fb.Fx < 0 {
		outcode2 |= 1
	} else {
		if fl.Fb.Fx >= f_w {
			outcode2 |= 2
		}
	}
	if outcode1&outcode2 != 0 {
		return 0
	}
	for outcode1|outcode2 != 0 {
		// may be partially inside box
		// find an outside point
		if outcode1 != 0 {
			outside = outcode1
		} else {
			outside = outcode2
		}
		// clip to each side
		if outside&8 != 0 {
			dy = fl.Fa.Fy - fl.Fb.Fy
			dx = fl.Fb.Fx - fl.Fa.Fx
			tmp.Fx = fl.Fa.Fx + dx*fl.Fa.Fy/dy
			tmp.Fy = 0
		} else {
			if outside&4 != 0 {
				dy = fl.Fa.Fy - fl.Fb.Fy
				dx = fl.Fb.Fx - fl.Fa.Fx
				tmp.Fx = fl.Fa.Fx + dx*(fl.Fa.Fy-f_h)/dy
				tmp.Fy = f_h - 1
			} else {
				if outside&2 != 0 {
					dy = fl.Fb.Fy - fl.Fa.Fy
					dx = fl.Fb.Fx - fl.Fa.Fx
					tmp.Fy = fl.Fa.Fy + dy*(f_w-1-fl.Fa.Fx)/dx
					tmp.Fx = f_w - 1
				} else {
					if outside&1 != 0 {
						dy = fl.Fb.Fy - fl.Fa.Fy
						dx = fl.Fb.Fx - fl.Fa.Fx
						tmp.Fy = fl.Fa.Fy + dy*-fl.Fa.Fx/dx
						tmp.Fx = 0
					} else {
						tmp.Fx = 0
						tmp.Fy = 0
					}
				}
			}
		}
		if outside == outcode1 {
			fl.Fa = tmp
			outcode1 = 0
			if fl.Fa.Fy < 0 {
				outcode1 |= 8
			} else {
				if fl.Fa.Fy >= f_h {
					outcode1 |= 4
				}
			}
			if fl.Fa.Fx < 0 {
				outcode1 |= 1
			} else {
				if fl.Fa.Fx >= f_w {
					outcode1 |= 2
				}
			}
		} else {
			fl.Fb = tmp
			outcode2 = 0
			if fl.Fb.Fy < 0 {
				outcode2 |= 8
			} else {
				if fl.Fb.Fy >= f_h {
					outcode2 |= 4
				}
			}
			if fl.Fb.Fx < 0 {
				outcode2 |= 1
			} else {
				if fl.Fb.Fx >= f_w {
					outcode2 |= 2
				}
			}
		}
		if outcode1&outcode2 != 0 {
			return 0
		} // trivially outside
	}
	return 1
}

// C documentation
//
//	//
//	// Classic Bresenham w/ whatever optimizations needed for speed
//	//
func am_drawFline(fl *fline_t, color int32) {
	var ax, ay, d, dx, dy, sx, sy, x, y, v1, v2, v3, v4, v5 int32
	// For debugging only
	if fl.Fa.Fx < 0 || fl.Fa.Fx >= f_w || fl.Fa.Fy < 0 || fl.Fa.Fy >= f_h || fl.Fb.Fx < 0 || fl.Fb.Fx >= f_w || fl.Fb.Fy < 0 || fl.Fb.Fy >= f_h {
		v1 = fuck
		fuck++

		fprintf_ccgo(os.Stderr, "fuck %d \r", v1)
		return
	}
	dx = fl.Fb.Fx - fl.Fa.Fx
	if dx < 0 {
		v2 = -dx
	} else {
		v2 = dx
	}
	ax = 2 * v2
	if dx < 0 {
		v3 = -1
	} else {
		v3 = 1
	}
	sx = v3
	dy = fl.Fb.Fy - fl.Fa.Fy
	if dy < 0 {
		v4 = -dy
	} else {
		v4 = dy
	}
	ay = 2 * v4
	if dy < 0 {
		v5 = -1
	} else {
		v5 = 1
	}
	sy = v5
	x = fl.Fa.Fx
	y = fl.Fa.Fy
	if ax > ay {
		d = ay - ax/2
		for 1 != 0 {
			fb[y*f_w+x] = uint8(color)
			if x == fl.Fb.Fx {
				return
			}
			if d >= 0 {
				y += sy
				d -= ax
			}
			x += sx
			d += ay
		}
	} else {
		d = ax - ay/2
		for 1 != 0 {
			fb[y*f_w+x] = uint8(color)
			if y == fl.Fb.Fy {
				return
			}
			if d >= 0 {
				x += sx
				d -= ay
			}
			y += sy
			d += ax
		}
	}
}

var fuck int32

// C documentation
//
//	//
//	// Clip lines, draw visible part sof lines.
//	//
func am_drawMline(ml *mline_t, color int32) {
	if am_clipMline(ml, &fl) != 0 {
		am_drawFline(&fl, color)
	} // draws it on frame buffer using fb coords
}

var fl fline_t

// C documentation
//
//	//
//	// Draws flat (floor/ceiling tile) aligned grid lines.
//	//
func am_drawGrid(color int32) {
	bp := &mline_t{}
	var end, start, x, y fixed_t
	// Figure out start of vertical gridlines
	start = m_x
	if (start-bmaporgx)%(MAPBLOCKUNITS<<FRACBITS) != 0 {
		start += MAPBLOCKUNITS<<FRACBITS - (start-bmaporgx)%(MAPBLOCKUNITS<<FRACBITS)
	}
	end = m_x + m_w
	// draw vertical gridlines
	bp.Fa.Fy = m_y
	bp.Fb.Fy = m_y + m_h
	for x = start; x < end; x += MAPBLOCKUNITS << FRACBITS {
		bp.Fa.Fx = x
		bp.Fb.Fx = x
		am_drawMline(bp, color)
	}
	// Figure out start of horizontal gridlines
	start = m_y
	if (start-bmaporgy)%(MAPBLOCKUNITS<<FRACBITS) != 0 {
		start += MAPBLOCKUNITS<<FRACBITS - (start-bmaporgy)%(MAPBLOCKUNITS<<FRACBITS)
	}
	end = m_y + m_h
	// draw horizontal gridlines
	bp.Fa.Fx = m_x
	bp.Fb.Fx = m_x + m_w
	for y = start; y < end; y += MAPBLOCKUNITS << FRACBITS {
		bp.Fa.Fy = y
		bp.Fb.Fy = y
		am_drawMline(bp, color)
	}
}

// C documentation
//
//	//
//	// Determines visible lines, draws them.
//	// This is LineDef based, not LineSeg based.
//	//
func am_drawWalls() {
	for i := int32(0); i < numlines; i++ {
		line := &lines[i]
		l.Fa.Fx = line.Fv1.Fx
		l.Fa.Fy = line.Fv1.Fy
		l.Fb.Fx = line.Fv2.Fx
		l.Fb.Fy = line.Fv2.Fy
		if cheating != 0 || int32(line.Fflags)&ml_MAPPED != 0 {
			if int32(line.Fflags)&ml_DONTDRAW != 0 && cheating == 0 {
				continue
			}
			if line.Fbacksector == nil {
				am_drawMline(&l, 256-5*16+lightlev)
			} else {
				if int32(line.Fspecial) == 39 {
					// teleporters
					am_drawMline(&l, 256-5*16+REDRANGE/2)
				} else {
					if int32(line.Fflags)&ml_SECRET != 0 { // secret door
						if cheating != 0 {
							am_drawMline(&l, 256-5*16+lightlev)
						} else {
							am_drawMline(&l, 256-5*16+lightlev)
						}
					} else {
						if line.Fbacksector.Ffloorheight != line.Ffrontsector.Ffloorheight {
							am_drawMline(&l, 4*16+lightlev) // floor level change
						} else {
							if line.Fbacksector.Fceilingheight != line.Ffrontsector.Fceilingheight {
								am_drawMline(&l, 256-32+7+lightlev) // ceiling level change
							} else {
								if cheating != 0 {
									am_drawMline(&l, 6*16+lightlev)
								}
							}
						}
					}
				}
			}
		} else {
			if plr.Fpowers[pw_allmap] != 0 {
				if int32(line.Fflags)&ml_DONTDRAW == 0 {
					am_drawMline(&l, 6*16+3)
				}
			}
		}
	}
}

var l mline_t

// C documentation
//
//	//
//	// Rotation in 2D.
//	// Used to rotate player arrow line character.
//	//
func am_rotate(x *fixed_t, y *fixed_t, a angle_t) {
	tmpx := fixedMul(*y, finecosine[a>>ANGLETOFINESHIFT]) - fixedMul(*y, finesine[a>>ANGLETOFINESHIFT])
	*y = fixedMul(*x, finesine[a>>ANGLETOFINESHIFT]) + fixedMul(*y, finecosine[a>>ANGLETOFINESHIFT])
	*x = tmpx
}

func am_drawLineCharacter(lineguy []mline_t, scale fixed_t, angle angle_t, color int32, x fixed_t, y fixed_t) {
	for i := range lineguy {
		var bp mline_t
		bp.Fa.Fx = lineguy[i].Fa.Fx
		bp.Fa.Fy = lineguy[i].Fa.Fy
		if scale != 0 {
			bp.Fa.Fx = fixedMul(scale, bp.Fa.Fx)
			bp.Fa.Fy = fixedMul(scale, bp.Fa.Fy)
		}
		if angle != 0 {
			am_rotate(&bp.Fa.Fx, &bp.Fb.Fy, angle)
		}
		bp.Fa.Fx += x
		bp.Fa.Fy += y
		bp.Fb.Fx = lineguy[i].Fb.Fx
		bp.Fb.Fy = lineguy[i].Fb.Fy
		if scale != 0 {
			bp.Fb.Fx = fixedMul(scale, bp.Fb.Fx)
			bp.Fb.Fy = fixedMul(scale, bp.Fb.Fy)
		}
		if angle != 0 {
			am_rotate(&bp.Fb.Fx, &bp.Fb.Fy, angle)
		}
		bp.Fb.Fx += x
		bp.Fb.Fy += y
		am_drawMline(&bp, color)
	}
}

func am_drawPlayers() {
	var color, their_color int32
	their_color = -1
	if netgame == 0 {
		if cheating != 0 {
			am_drawLineCharacter(cheat_player_arrow[:], 0, plr.Fmo.Fangle, 256-47, plr.Fmo.Fx, plr.Fmo.Fy)
		} else {
			am_drawLineCharacter(player_arrow[:], 0, plr.Fmo.Fangle, 256-47, plr.Fmo.Fx, plr.Fmo.Fy)
		}
		return
	}
	for i := 0; i < MAXPLAYERS; i++ {
		their_color++
		p := &players[i]
		if deathmatch != 0 && singledemo == 0 && p != plr {
			continue
		}
		if playeringame[i] == 0 {
			continue
		}
		if p.Fpowers[pw_invisibility] != 0 {
			color = 246
		} else {
			color = their_colors[their_color]
		}
		am_drawLineCharacter(player_arrow[:], 0, p.Fmo.Fangle, color, p.Fmo.Fx, p.Fmo.Fy)
	}
}

var their_colors = [4]int32{
	0: 7 * 16,
	1: 6 * 16,
	2: 4 * 16,
	3: 256 - 5*16,
}

func am_drawThings(colors int32, colorrange int32) {
	var t *mobj_t
	for i := int32(0); i < numsectors; i++ {
		t = sectors[i].Fthinglist
		for t != nil {
			am_drawLineCharacter(thintriangle_guy[:], 16<<FRACBITS, t.Fangle, colors+lightlev, t.Fx, t.Fy)
			t = t.Fsnext
		}
	}
}

func am_drawMarks() {
	var fx, fy, h, w int32
	for i := 0; i < AM_NUMMARKPOINTS; i++ {
		if markpoints[i].Fx != -1 {
			//      w = SHORT(marknums[i]->width);
			//      h = SHORT(marknums[i]->height);
			w = 5 // because something's wrong with the wad, i guess
			h = 6 // because something's wrong with the wad, i guess
			fx = f_x + fixedMul(markpoints[i].Fx-m_x, scale_mtof)>>16
			fy = f_y + (f_h - fixedMul(markpoints[i].Fy-m_y, scale_mtof)>>16)
			if fx >= f_x && fx <= f_w-w && fy >= f_y && fy <= f_h-h {
				v_DrawPatch(fx, fy, marknums[i])
			}
		}
	}
}

func am_drawCrosshair(color int32) {
	fb[f_w*(f_h+1)/2+f_w/2] = uint8(color) // single point for now
}

func am_Drawer() {
	if automapactive == 0 {
		return
	}
	am_clearFB(BLACK)
	if grid {
		am_drawGrid(6*16 + GRAYSRANGE/2)
	}
	am_drawWalls()
	am_drawPlayers()
	if cheating == 2 {
		am_drawThings(7*16, GREENRANGE)
	}
	am_drawCrosshair(6 * 16)
	am_drawMarks()
	v_MarkRect(f_x, f_y, f_w, f_h)
}

func init() {
	gamemode = indetermined
	gameversion = exe_final2
	doom1_endmsg = [8]string{
		0: "are you sure you want to\nquit this great game?",
		1: "please don't leave, there's more\ndemons to toast!",
		2: "let's beat it -- this is turning\ninto a bloodbath!",
		3: "i wouldn't leave if i were you.\ndos is much worse.",
		4: "you're trying to say you like dos\nbetter than me, right?",
		5: "don't leave yet -- there's a\ndemon around that corner!",
		6: "ya know, next time you come in here\ni'm gonna toast ya.",
		7: "go ahead and leave. see if i care.",
	}
	doom2_endmsg = [8]string{
		0: "are you sure you want to\nquit this great game?",
		1: "you want to quit?\nthen, thou hast lost an eighth!",
		2: "don't go now, there's a \ndimensional shambler waiting\nat the dos prompt!",
		3: "get outta here and go back\nto your boring programs.",
		4: "if i were your boss, i'd \n deathmatch ya in a minute!",
		5: "look, bud. you leave now\nand you forfeit your body count!",
		6: "just leave. when you come\nback, i'll be waiting with a bat.",
		7: "you're lucky i don't smack\nyou for thinking about leaving.",
	}
}

const MAXEVENTS = 64

var events [64]event_t
var eventhead int32
var eventtail int32

// C documentation
//
//	//
//	// D_PostEvent
//	// Called by the I/O functions when input is detected
//	//
func d_PostEvent(ev *event_t) {
	events[eventhead] = *ev
	eventhead = (eventhead + 1) % MAXEVENTS
}

// Read an event from the queue.

func d_PopEvent() *event_t {
	// No more events waiting.
	if eventtail == eventhead {
		return nil
	}
	result := &events[eventtail]
	// Advance to the next event in the queue.
	eventtail = (eventtail + 1) % MAXEVENTS
	return result
}

func init() {
	weaponinfo = [9]weaponinfo_t{
		0: {
			Fammo:       am_noammo,
			Fupstate:    s_PUNCHUP,
			Fdownstate:  s_PUNCHDOWN,
			Freadystate: s_PUNCH,
			Fatkstate:   s_PUNCH1,
		},
		1: {
			Fupstate:    s_PISTOLUP,
			Fdownstate:  s_PISTOLDOWN,
			Freadystate: s_PISTOL,
			Fatkstate:   s_PISTOL1,
			Fflashstate: s_PISTOLFLASH,
		},
		2: {
			Fammo:       am_shell,
			Fupstate:    s_SGUNUP,
			Fdownstate:  s_SGUNDOWN,
			Freadystate: s_SGUN,
			Fatkstate:   s_SGUN1,
			Fflashstate: s_SGUNFLASH1,
		},
		3: {
			Fupstate:    s_CHAINUP,
			Fdownstate:  s_CHAINDOWN,
			Freadystate: s_CHAIN,
			Fatkstate:   s_CHAIN1,
			Fflashstate: s_CHAINFLASH1,
		},
		4: {
			Fammo:       am_misl,
			Fupstate:    s_MISSILEUP,
			Fdownstate:  s_MISSILEDOWN,
			Freadystate: s_MISSILE,
			Fatkstate:   s_MISSILE1,
			Fflashstate: s_MISSILEFLASH1,
		},
		5: {
			Fammo:       am_cell,
			Fupstate:    s_PLASMAUP,
			Fdownstate:  s_PLASMADOWN,
			Freadystate: s_PLASMA,
			Fatkstate:   s_PLASMA1,
			Fflashstate: s_PLASMAFLASH1,
		},
		6: {
			Fammo:       am_cell,
			Fupstate:    s_BFGUP,
			Fdownstate:  s_BFGDOWN,
			Freadystate: s_BFG,
			Fatkstate:   s_BFG1,
			Fflashstate: s_BFGFLASH1,
		},
		7: {
			Fammo:       am_noammo,
			Fupstate:    s_SAWUP,
			Fdownstate:  s_SAWDOWN,
			Freadystate: s_SAW,
			Fatkstate:   s_SAW1,
		},
		8: {
			Fammo:       am_shell,
			Fupstate:    s_DSGUNUP,
			Fdownstate:  s_DSGUNDOWN,
			Freadystate: s_DSGUN,
			Fatkstate:   s_DSGUN1,
			Fflashstate: s_DSGUNFLASH1,
		},
	}
}

type iwad_t struct {
	Fname        string
	Fmission     gamemission_t
	Fmode        gamemode_t
	Fdescription string
}

//
// This is used to get the local FILE:LINE info from CPP
// prior to really call the function in question.
//

var iwads = [14]iwad_t{
	0: {
		Fname:        "doom2.wad",
		Fmission:     doom2,
		Fmode:        commercial,
		Fdescription: "Doom II",
	},
	1: {
		Fname:        "plutonia.wad",
		Fmission:     pack_plut,
		Fmode:        commercial,
		Fdescription: "Final Doom: Plutonia Experiment",
	},
	2: {
		Fname:        "tnt.wad",
		Fmission:     pack_tnt,
		Fmode:        commercial,
		Fdescription: "Final Doom: TNT: Evilution",
	},
	3: {
		Fname:        "doom.wad",
		Fmode:        retail,
		Fdescription: "Doom",
	},
	4: {
		Fname:        "doom1.wad",
		Fdescription: "Doom Shareware",
	},
	5: {
		Fname:        "chex.wad",
		Fmission:     pack_chex,
		Fdescription: "Chex Quest",
	},
	6: {
		Fname:        "hacx.wad",
		Fmission:     pack_hacx,
		Fmode:        commercial,
		Fdescription: "Hacx",
	},
	7: {
		Fname:        "freedm.wad",
		Fmission:     doom2,
		Fmode:        commercial,
		Fdescription: "FreeDM",
	},
	8: {
		Fname:        "freedoom2.wad",
		Fmission:     doom2,
		Fmode:        commercial,
		Fdescription: "Freedoom: Phase 2",
	},
	9: {
		Fname:        "freedoom1.wad",
		Fmode:        retail,
		Fdescription: "Freedoom: Phase 1",
	},
	10: {
		Fname:        "heretic.wad",
		Fmission:     heretic,
		Fmode:        retail,
		Fdescription: "Heretic",
	},
	11: {
		Fname:        "heretic1.wad",
		Fmission:     heretic,
		Fdescription: "Heretic Shareware",
	},
	12: {
		Fname:        "hexen.wad",
		Fmission:     hexen,
		Fmode:        commercial,
		Fdescription: "Hexen",
	},
	13: {
		Fname:        "strife1.wad",
		Fmission:     strife,
		Fmode:        commercial,
		Fdescription: "Strife",
	},
}

// Array of locations to search for IWAD files
var iwad_dirs []string

func addIWADDir(dir string) {
	iwad_dirs = append(iwad_dirs, dir)
}

// This is Windows-specific code that automatically finds the location
// of installed IWAD files.  The registry is inspected to find special
// keys installed by the Windows installers for various CD versions
// of Doom.  From these keys we can deduce where to find an IWAD.

// Returns true if the specified path is a path to a file
// of the specified name.

func dirIsFile(path string, filename string) boolean {
	if strings.HasPrefix(filename, path) && path[len(path)-1] == '/' {
		return 1
	}
	return 0
}

// Check if the specified directory contains the specified IWAD
// file, returning the full path to the IWAD if found, or NULL
// if not found.

func checkDirectoryHasIWAD(dir string, iwadname string) string {
	var filename string
	// As a special case, the "directory" may refer directly to an
	// IWAD file if the path comes from DOOMWADDIR or DOOMWADPATH.
	if dirIsFile(dir, iwadname) != 0 && m_FileExists(dir) != 0 {
		return dir
	}
	// Construct the full path to the IWAD if it is located in
	// this directory, and check if it exists.
	if dir == "." {
		filename = iwadname
	} else {
		filename = dir + "/" + iwadname
	}
	fprintf_ccgo(os.Stdout, "Trying IWAD file:%s\n", filename)
	if m_FileExists(filename) != 0 {
		return filename
	}
	return ""
}

// Search a directory to try to find an IWAD
// Returns the location of the IWAD if found, otherwise NULL.

func searchDirectoryForIWAD(dir string, mask int32, mission *gamemission_t) string {
	var filename string
	for i := 0; i < len(iwads); i++ {
		if 1<<iwads[i].Fmission&mask == 0 {
			continue
		}
		filename = checkDirectoryHasIWAD(dir, iwads[i].Fname)
		if filename != "" {
			*mission = iwads[i].Fmission
			return filename
		}
	}
	return ""
}

// When given an IWAD with the '-iwad' parameter,
// attempt to identify it by its name.

func identifyIWADByName(name string, mask int32) gamemission_t {
	var mission gamemission_t
	mission = none
	for i := 0; i < len(iwads); i++ {
		// Check if the filename is this IWAD name.
		// Only use supported missions:
		if 1<<iwads[i].Fmission&mask == 0 {
			continue
		}
		// Check if it ends in this IWAD name.
		if name == iwads[i].Fname {
			mission = iwads[i].Fmission
			break
		}
	}
	return mission
}

//
// Build a list of IWAD files
//

func buildIWADDirList() {
	addIWADDir(".")
}

//
// Searches WAD search paths for an WAD with a specific filename.
//

func d_FindWADByName(name string) string {
	// Absolute path?
	if m_FileExists(name) != 0 {
		return name
	}
	buildIWADDirList()
	// Search through all IWAD paths for a file with the given name.
	for i := 0; i < len(iwad_dirs); i++ {
		// As a special case, if this is in DOOMWADDIR or DOOMWADPATH,
		// the "directory" may actually refer directly to an IWAD
		// file.
		if dirIsFile(iwad_dirs[i], name) != 0 && m_FileExists(iwad_dirs[i]) != 0 {
			return iwad_dirs[i]
		}
		// Construct a string for the full path
		path := iwad_dirs[i] + "/"
		if m_FileExists(path) != 0 {
			return path
		}
	}
	// File not found
	return ""
}

//
// D_TryWADByName
//
// Searches for a WAD by its filename, or passes through the filename
// if not found.
//

func d_TryFindWADByName(filename string) string {
	result := d_FindWADByName(filename)
	if result != "" {
		return result
	} else {
		return filename
	}
}

//
// FindIWAD
// Checks availability of IWAD files by name,
// to determine whether registered/commercial features
// should be executed (notably loading PWADs).
//

func d_FindIWAD(mask int32, mission *gamemission_t) string {
	var iwadparm int32
	var result string
	var iwadfile string
	// Check for the -iwad parameter
	//!
	// Specify an IWAD file to use.
	//
	// @arg <file>
	//
	iwadparm = m_CheckParmWithArgs("-iwad", 1)
	if iwadparm != 0 {
		// Search through IWAD dirs for an IWAD with the given name.
		iwadfile = myargs[iwadparm+1]
		result = d_FindWADByName(iwadfile)
		if result == "" {
			i_Error("IWAD file '%s' not found!", iwadfile)
		}
		*mission = identifyIWADByName(result, mask)
	} else {
		// Search through the list and look for an IWAD
		fprintf_ccgo(os.Stdout, "-iwad not specified, trying a few iwad names\n")
		result = ""
		buildIWADDirList()
		for i := 0; i < len(iwad_dirs); i++ {
			result = searchDirectoryForIWAD(iwad_dirs[i], mask, mission)
			if result != "" {
				break
			}
		}
	}
	return result
}

//
// Get the IWAD name used for savegames.
//

func d_SaveGameIWADName(gamemission gamemission_t) string {
	// Determine the IWAD name to use for savegames.
	// This determines the directory the savegame files get put into.
	//
	// Note that we match on gamemission rather than on IWAD name.
	// This ensures that doom1.wad and doom.wad saves are stored
	// in the same place.
	for i := 0; i < len(iwads); i++ {
		if gamemission == iwads[i].Fmission {
			return iwads[i].Fname
		}
	}
	// Default fallback:
	return "unknown.wad"
}

func d_SuggestGameName(mission gamemission_t, mode gamemode_t) string {
	for i := 0; i < len(iwads); i++ {
		if iwads[i].Fmission == mission && (mode == indetermined || iwads[i].Fmode == mode) {
			return iwads[i].Fdescription
		}
	}
	return "Unknown game?"
}

// The complete set of data for a particular tic.

type ticcmd_set_t struct {
	Fcmds   [8]ticcmd_t
	Fingame [8]boolean
}

//
// gametic is the tic about to (or currently being) run
// maketic is the tic that hasn't had control made for it yet
// recvtic is the latest tic received from the server.
//
// a gametic cannot be run until ticcmds are received for it
// from all players.
//

var ticdata [128]ticcmd_set_t

// The index of the next tic to be made (with a call to BuildTiccmd).

var maketic int32

// The number of complete tics received from the server so far.

var recvtic int32

// Index of the local player.

var localplayer int32

// Used for original sync code.

var skiptics int32 = 0

// Use new client syncronisation code

var new_sync uint32 = 1

// Callback functions for loop code.

var loop_interface *loop_interface_t

// Current players in the multiplayer game.
// This is distinct from playeringame[] used by the game code, which may
// modify playeringame[] when playing back multiplayer demos.

var local_playeringame [8]boolean

// Requested player class "sent" to the server on connect.
// If we are only doing a single player game then this needs to be remembered
// and saved in the game settings.

var player_class int32

// 35 fps clock adjusted by offsetms milliseconds

func getAdjustedTime() int32 {
	var time_ms int32
	time_ms = I_GetTimeMS()
	if new_sync != 0 {
		// Use the adjustments from net_client.c only if we are
		// using the new sync mode.
		time_ms += offsetms / (1 << FRACBITS)
	}
	return time_ms * TICRATE / 1000
}

func buildNewTic() boolean {
	var cmd ticcmd_t
	var gameticdiv int32
	gameticdiv = gametic / ticdup
	i_StartTic()
	loop_interface.FProcessEvents()
	// Always run the menu
	loop_interface.FRunMenu()
	if drone != 0 {
		// In drone mode, do not generate any ticcmds.
		return 0
	}
	if new_sync != 0 {
		// If playing single player, do not allow tics to buffer
		// up very far
		if net_client_connected == 0 && maketic-gameticdiv > 2 {
			return 0
		}
		// Never go more than ~200ms ahead
		if maketic-gameticdiv > 8 {
			return 0
		}
	} else {
		if maketic-gameticdiv >= 5 {
			return 0
		}
	}
	//printf ("mk:%i ",maketic);
	loop_interface.FBuildTiccmd(&cmd, maketic)

	ticdata[maketic%BACKUPTICS].Fcmds[localplayer] = cmd
	ticdata[maketic%BACKUPTICS].Fingame[localplayer] = 1

	maketic++
	return 1
}

func netUpdate() {
	var newtics, nowtime int32
	// If we are running with singletics (timing a demo), this
	// is all done separately.
	if singletics != 0 {
		return
	}
	// check time
	nowtime = getAdjustedTime() / ticdup
	newtics = nowtime - lasttime
	lasttime = nowtime
	if skiptics <= newtics {
		newtics -= skiptics
		skiptics = 0
	} else {
		skiptics -= newtics
		newtics = 0
	}
	// build new ticcmds for console player
	for i := int32(0); i < newtics; i++ {
		if buildNewTic() == 0 {
			break
		}
	}
}

//
// Start game loop
//
// Called after the screen is set but before the game starts running.
//

func d_StartGameLoop() {
	lasttime = getAdjustedTime() / ticdup
}

func d_StartNetGame(settings *net_gamesettings_t, callback netgame_startup_callback_t) {
	settings.Fconsoleplayer = 0
	settings.Fnum_players = 1
	settings.Fplayer_classes[0] = player_class
	settings.Fnew_sync = 0
	settings.Fextratics = 1
	settings.Fticdup = 1
	ticdup = settings.Fticdup
	new_sync = uint32(settings.Fnew_sync)
}

func d_InitNetGame(connect_data *net_connect_data_t) boolean {
	// Call d_QuitNetGame on exit:
	i_AtExit(d_QuitNetGame, 1)
	player_class = connect_data.Fplayer_class
	return 0
}

// C documentation
//
//	//
//	// D_QuitNetGame
//	// Called before quitting to leave a net game
//	// without hanging the other players
//	//
func d_QuitNetGame() {
}

func getLowTic() int32 {
	var lowtic int32
	lowtic = maketic
	return lowtic
}

var frameon int32
var frameskip [4]int32
var oldnettics int32

func oldNetSync() {
	var keyplayer int32
	keyplayer = -1
	frameon++
	// ideally maketic should be 1 - 3 tics above lowtic
	// if we are consistantly slower, speed up time
	for i := 0; i < NET_MAXPLAYERS; i++ {
		if local_playeringame[i] != 0 {
			keyplayer = int32(i)
			break
		}
	}
	if keyplayer < 0 {
		// If there are no players, we can never advance anyway
		return
	}
	if localplayer == keyplayer {
		// the key player does not adapt
	} else {
		if maketic <= recvtic {
			lasttime--
			// printf ("-");
		}
		frameskip[frameon&3] = boolint32(oldnettics > recvtic)
		oldnettics = maketic
		if frameskip[0] != 0 && frameskip[1] != 0 && frameskip[2] != 0 && frameskip[3] != 0 {
			skiptics = 1
			// printf ("+");
		}
	}
}

// Returns true if there are players in the game:

func playersInGame() boolean {
	var result boolean
	result = 0
	// If we are connected to a server, check if there are any players
	// in the game.
	if net_client_connected != 0 {
		for i := 0; i < NET_MAXPLAYERS; i++ {
			result = booluint32(result != 0 || local_playeringame[i] != 0)
		}
	}
	// Whether single or multi-player, unless we are running as a drone,
	// we are in the game.
	if drone == 0 {
		result = 1
	}
	return result
}

// When using ticdup, certain values must be cleared out when running
// the duplicate ticcmds.

func ticdupSquash(set *ticcmd_set_t) {
	for i := 0; i < NET_MAXPLAYERS; i++ {
		cmd := &set.Fcmds[i]
		cmd.Fchatchar = 0
		if int32(cmd.Fbuttons)&bt_SPECIAL != 0 {
			cmd.Fbuttons = 0
		}
	}
}

// When running in single player mode, clear all the ingame[] array
// except the local player.

func singlePlayerClear(set *ticcmd_set_t) {
	for i := int32(0); i < NET_MAXPLAYERS; i++ {
		if i != localplayer {
			set.Fingame[i] = 0
		}
	}
}

//
// TryRunTics
//

func tryRunTics() {
	var availabletics, counts, entertic, lowtic, realtics, v1 int32
	var set *ticcmd_set_t
	// get real tics
	entertic = i_GetTime() / ticdup
	realtics = entertic - oldentertics
	oldentertics = entertic
	// in singletics mode, run a single tic every time this function
	// is called.
	if singletics != 0 {
		buildNewTic()
	} else {
		netUpdate()
	}
	lowtic = getLowTic()
	availabletics = lowtic - gametic/ticdup
	// decide how many tics to run
	if new_sync != 0 {
		counts = availabletics
	} else {
		// decide how many tics to run
		if realtics < availabletics-1 {
			counts = realtics + 1
		} else {
			if realtics < availabletics {
				counts = realtics
			} else {
				counts = availabletics
			}
		}
		if counts < 1 {
			counts = 1
		}
		if net_client_connected != 0 {
			oldNetSync()
		}
	}
	if counts < 1 {
		counts = 1
	}
	// wait for new tics if needed
	for playersInGame() == 0 || lowtic < gametic/ticdup+counts {
		netUpdate()
		lowtic = getLowTic()
		if lowtic < gametic/ticdup {
			i_Error("tryRunTics: lowtic < gametic")
		}
		// Don't stay in this loop forever.  The menu is still running,
		// so return to update the screen
		if i_GetTime()/ticdup-entertic > 0 {
			return
		}
		i_Sleep(1)
	}
	// run the count * ticdup dics
	for {
		v1 = counts
		counts--
		if v1 == 0 {
			break
		}
		if playersInGame() == 0 {
			return
		}
		set = &ticdata[gametic/ticdup%BACKUPTICS]
		if net_client_connected == 0 {
			singlePlayerClear(set)
		}
		for i := int32(0); i < ticdup; i++ {
			if gametic/ticdup > lowtic {
				i_Error("gametic>lowtic")
			}
			local_playeringame = set.Fingame
			loop_interface.FRunTic(set.Fcmds[:], set.Fingame[:])
			gametic++
			// modify command for duplicated tics
			ticdupSquash(set)
		}
		netUpdate() // check for new console commands
	}
}

var oldentertics int32

func d_RegisterLoopCallbacks(i *loop_interface_t) {
	loop_interface = i
}

const HU_BROADCAST = 5
const HU_MSGHEIGHT = 1
const HU_MSGX = 0
const HU_MSGY = 0
const NUM_VIRTUAL_BUTTONS = 10
const SAVESTRINGSIZE = 24

type sfxinfo_t struct {
	Ftagname     uintptr
	Fname        string
	Fpriority    int32
	Flink        *sfxinfo_t
	Fpitch       int32
	Fvolume      int32
	Fusefulness  int32
	Flumpnum     int32
	Fnumchannels int32
	Fdriver_data uintptr
}

type musicinfo_t struct {
	Fname    string
	Flumpnum int32
	Fdata    []byte
	Fhandle  uintptr
}

type snddevice_t = int32

const SNDDEVICE_ADLIB = 2
const SNDDEVICE_SB = 3
const SNDDEVICE_GUS = 5
const SNDDEVICE_GENMIDI = 8

type sound_module_t struct {
	Fsound_devices     []snddevice_t
	Fnum_sound_devices int32
	FInit              func(boolean) boolean
	FShutdown          func()
	FGetSfxLumpNum     func(sfxinfo *sfxinfo_t) int32
	FUpdate            func()
	FUpdateSoundParams func(channel int32, vol int32, sep int32)
	FStartSound        func(sfxinfo *sfxinfo_t, channel int32, vol int32, sep int32) int32
	FStopSound         func(channel int32)
	FSoundIsPlaying    func(channel int32) boolean
	FCacheSounds       func([]sfxinfo_t) boolean
}

type music_module_t struct {
	Fsound_devices     uintptr
	Fnum_sound_devices int32
	FInit              func()
	FShutdown          func()
	FSetMusicVolume    func(volume int32)
	FPauseMusic        func()
	FResumeMusic       func()
	FRegisterSong      func(data []byte) uintptr
	FUnRegisterSong    func(handle uintptr)
	FPlaySong          func(handle uintptr, looping boolean) (r boolean)
	FStopSong          func()
	FMusicIsPlaying    uintptr
	FPoll              func()
}

const mus_None = 0
const mus_e1m1 = 1
const mus_e1m5 = 5
const mus_e1m9 = 9
const mus_e2m4 = 13
const mus_e2m5 = 14
const mus_e2m6 = 15
const mus_e2m7 = 16
const mus_e3m2 = 20
const mus_e3m3 = 21
const mus_e3m4 = 22
const mus_inter = 28
const mus_intro = 29
const mus_bunny = 30
const mus_victor = 31
const mus_introa = 32
const mus_runnin = 33
const mus_evil = 63
const mus_read_m = 65
const mus_dm2ttl = 66
const mus_dm2int = 67
const NUMMUSIC = 68
const sfx_pistol = 1
const sfx_shotgn = 2
const sfx_sgcock = 3
const sfx_dshtgn = 4
const sfx_dbopn = 5
const sfx_dbcls = 6
const sfx_dbload = 7
const sfx_plasma = 8
const sfx_bfg = 9
const sfx_sawup = 10
const sfx_sawidl = 11
const sfx_sawful = 12
const sfx_sawhit = 13
const sfx_rlaunc = 14
const sfx_rxplod = 15
const sfx_firsht = 16
const sfx_firxpl = 17
const sfx_pstart = 18
const sfx_pstop = 19
const sfx_doropn = 20
const sfx_dorcls = 21
const sfx_stnmov = 22
const sfx_swtchn = 23
const sfx_swtchx = 24
const sfx_plpain = 25
const sfx_dmpain = 26
const sfx_popain = 27
const sfx_vipain = 28
const sfx_mnpain = 29
const sfx_pepain = 30
const sfx_slop = 31
const sfx_itemup = 32
const sfx_wpnup = 33
const sfx_oof = 34
const sfx_telept = 35
const sfx_posit1 = 36
const sfx_posit2 = 37
const sfx_posit3 = 38
const sfx_bgsit1 = 39
const sfx_bgsit2 = 40
const sfx_sgtsit = 41
const sfx_cacsit = 42
const sfx_brssit = 43
const sfx_cybsit = 44
const sfx_spisit = 45
const sfx_bspsit = 46
const sfx_kntsit = 47
const sfx_vilsit = 48
const sfx_mansit = 49
const sfx_pesit = 50
const sfx_sklatk = 51
const sfx_sgtatk = 52
const sfx_skepch = 53
const sfx_vilatk = 54
const sfx_claw = 55
const sfx_skeswg = 56
const sfx_pldeth = 57
const sfx_pdiehi = 58
const sfx_podth1 = 59
const sfx_podth2 = 60
const sfx_podth3 = 61
const sfx_bgdth1 = 62
const sfx_bgdth2 = 63
const sfx_sgtdth = 64
const sfx_cacdth = 65
const sfx_brsdth = 67
const sfx_cybdth = 68
const sfx_spidth = 69
const sfx_bspdth = 70
const sfx_vildth = 71
const sfx_kntdth = 72
const sfx_pedth = 73
const sfx_skedth = 74
const sfx_posact = 75
const sfx_bgact = 76
const sfx_dmact = 77
const sfx_bspact = 78
const sfx_bspwlk = 79
const sfx_vilact = 80
const sfx_noway = 81
const sfx_barexp = 82
const sfx_punch = 83
const sfx_hoof = 84
const sfx_metal = 85
const sfx_tink = 87
const sfx_bdopn = 88
const sfx_bdcls = 89
const sfx_itmbk = 90
const sfx_flame = 91
const sfx_flamst = 92
const sfx_getpow = 93
const sfx_bospit = 94
const sfx_boscub = 95
const sfx_bossit = 96
const sfx_bospn = 97
const sfx_bosdth = 98
const sfx_manatk = 99
const sfx_mandth = 100
const sfx_sssit = 101
const sfx_ssdth = 102
const sfx_keenpn = 103
const sfx_keendt = 104
const sfx_skeact = 105
const sfx_skesit = 106
const sfx_skeatk = 107
const sfx_radio = 108
const NUMSFX = 109
const wipe_Melt = 1

type stateenum_t = int32

const NoState = -1
const StatCount = 0
const ShowNextLoc = 1

func init() {
	show_endoom = 1
}

// C documentation
//
//	//
//	// D_ProcessEvents
//	// Send all the events of the given timestamp down the responder chain
//	//
func d_ProcessEvents() {
	// IF STORE DEMO, DO NOT ACCEPT INPUT
	if storedemo != 0 {
		return
	}
	for {
		ev := d_PopEvent()
		if ev == nil {
			break
		}
		if m_Responder(ev) != 0 {
			continue
		} // menu ate the event
		g_Responder(ev)
	}
}

func init() {
	wipegamestate = gs_DEMOSCREEN
}

func d_Display() {
	var done, redrawsbar, wipe boolean
	var nowtime, tics, wipestart, y int32
	var v1 gamestate_t
	if nodrawers != 0 {
		return
	} // for comparative timing / profiling
	redrawsbar = 0
	// change the view size if needed
	if setsizeneeded != 0 {
		r_ExecuteSetViewSize()
		oldgamestate1 = -1 // force background redraw
		borderdrawcount = 3
	}
	// save the current screen if about to wipe
	if gamestate != wipegamestate {
		wipe = 1
		wipe_StartScreen(0, 0, SCREENWIDTH, SCREENHEIGHT)
	} else {
		wipe = 0
	}
	if gamestate == gs_LEVEL && gametic != 0 {
		hu_Erase()
	}
	// do buffered drawing
	switch gamestate {
	case gs_LEVEL:
		if gametic == 0 {
			break
		}
		if automapactive != 0 {
			am_Drawer()
		}
		if wipe != 0 || viewheight != 200 && fullscreen != 0 {
			redrawsbar = 1
		}
		if inhelpscreensstate != 0 && inhelpscreens == 0 {
			redrawsbar = 1
		} // just put away the help screen
		st_Drawer(booluint32(viewheight == 200), redrawsbar)
		fullscreen = booluint32(viewheight == 200)
	case gs_INTERMISSION:
		wi_Drawer()
	case gs_FINALE:
		f_Drawer()
	case gs_DEMOSCREEN:
		d_PageDrawer()
		break
	}
	// draw buffered stuff to screen
	i_UpdateNoBlit()
	// draw the view directly
	if gamestate == gs_LEVEL && automapactive == 0 && gametic != 0 {
		r_RenderPlayerView(&players[displayplayer])
	}
	if gamestate == gs_LEVEL && gametic != 0 {
		hu_Drawer()
	}
	// clean up border stuff
	if gamestate != oldgamestate1 && gamestate != gs_LEVEL {
		i_SetPalette(w_CacheLumpNameBytes("PLAYPAL"))
	}
	// see if the border needs to be initially drawn
	if gamestate == gs_LEVEL && oldgamestate1 != gs_LEVEL {
		viewactivestate = 0 // view was not active
		r_FillBackScreen()  // draw the pattern into the back screen
	}
	// see if the border needs to be updated to the screen
	if gamestate == gs_LEVEL && automapactive == 0 && scaledviewwidth != 320 {
		if menuactive != 0 || menuactivestate != 0 || viewactivestate == 0 {
			borderdrawcount = 3
		}
		if borderdrawcount != 0 {
			r_DrawViewBorder() // erase old menu stuff
			borderdrawcount--
		}
	}
	if testcontrols != 0 {
		// Box showing current mouse speed
		v_DrawMouseSpeedBox(testcontrols_mousespeed)
	}
	menuactivestate = menuactive
	viewactivestate = viewactive
	inhelpscreensstate = inhelpscreens
	v1 = gamestate
	wipegamestate = v1
	oldgamestate1 = v1
	// draw pause pic
	if paused != 0 {
		if automapactive != 0 {
			y = 4
		} else {
			y = viewwindowy + 4
		}
		v_DrawPatchDirect(viewwindowx+(scaledviewwidth-int32(68))/2, y, w_CacheLumpNameT("M_PAUSE"))
	}
	// menus go directly to the screen
	m_Drawer()  // menu is drawn even on top of everything
	netUpdate() // send out any new accumulation
	// normal update
	if wipe == 0 {
		i_FinishUpdate() // page flip or blit buffer
		return
	}
	// wipe update
	wipe_EndScreen(0, 0, SCREENWIDTH, SCREENHEIGHT)
	wipestart = i_GetTime() - 1
	for cond := true; cond; cond = done == 0 {
		for cond := true; cond; cond = tics <= 0 {
			nowtime = i_GetTime()
			tics = nowtime - wipestart
			i_Sleep(1)
		}
		wipestart = nowtime
		done = uint32(wipe_ScreenWipe(int32(wipe_Melt), 0, 0, SCREENWIDTH, SCREENHEIGHT, tics))
		i_UpdateNoBlit()
		m_Drawer()       // menu is drawn even on top of wipes
		i_FinishUpdate() // page flip or blit buffer
	}
}

var viewactivestate boolean

var menuactivestate boolean

var inhelpscreensstate boolean

var fullscreen boolean

var oldgamestate1 gamestate_t = -1

var borderdrawcount int32

//
// Add configuration file variable bindings.
//

func d_BindVariables() {
	m_ApplyPlatformDefaults()
	i_BindVideoVariables()
	i_BindJoystickVariables()
	i_BindSoundVariables()
	m_BindBaseControls()
	m_BindWeaponControls()
	m_BindMapControls()
	m_BindMenuControls()
	m_BindChatControls(MAXPLAYERS)
	key_multi_msgplayer[0] = 'g'
	key_multi_msgplayer[1] = 'i'
	key_multi_msgplayer[2] = 'b'
	key_multi_msgplayer[3] = 'r'
	m_BindVariable("mouse_sensitivity", &mouseSensitivity)
	m_BindVariable("sfx_volume", &sfxVolume)
	m_BindVariable("music_volume", &musicVolume)
	m_BindVariable("show_messages", &showMessages)
	m_BindVariable("screenblocks", &screenblocks)
	m_BindVariable("detaillevel", &detailLevel)
	m_BindVariable("snd_channels", &snd_channels)
	m_BindVariable("vanilla_savegame_limit", &vanilla_savegame_limit)
	m_BindVariable("vanilla_demo_limit", &vanilla_demo_limit)
	m_BindVariable("show_endoom", &show_endoom)
	// Multiplayer chat macros
	for i := 0; i < len(chat_macros); i++ {
		name := fmt.Sprintf("chatmacro%d", i)
		m_BindVariable(name, &chat_macros[i])
	}
}

//
// D_GrabMouseCallback
//
// Called to determine whether to grab the mouse pointer
//

func d_GrabMouseCallback() boolean {
	// Drone players don't need mouse focus
	if drone != 0 {
		return 0
	}
	// when menu is active or game is paused, release the mouse
	if menuactive != 0 || paused != 0 {
		return 0
	}
	// only grab mouse when playing levels (but not demos)
	return booluint32(gamestate == gs_LEVEL && demoplayback == 0 && advancedemo == 0)
}

func doomgeneric_Tick() {
	// frame syncronous IO operations
	i_StartFrame()
	tryRunTics() // will run at least one tic
	var dmo *degenmobj_t
	if players[consoleplayer].Fmo != nil {
		dmo = &players[consoleplayer].Fmo.degenmobj_t // console player
	}
	s_UpdateSounds(dmo) // move positional sounds
	// Update display, next frame, with current state.
	d_Display()
}

// C documentation
//
//	//
//	//  D_DoomLoop
//	//
func d_DoomLoop() {
	if bfgedition != 0 && (demorecording != 0 || gameaction == ga_playdemo || netgame != 0) {
		fprintf_ccgo(os.Stdout, " WARNING: You are playing using one of the Doom Classic\n IWAD files shipped with the Doom 3: BFG Edition. These are\n known to be incompatible with the regular IWAD files and\n may cause demos and network games to get out of sync.\n")
	}
	if demorecording != 0 {
		g_BeginRecording()
	}
	main_loop_started = 1
	tryRunTics()
	i_SetWindowTitle(gamedescription)
	i_GraphicsCheckCommandLine()
	i_SetGrabMouseCallback(d_GrabMouseCallback)
	i_InitGraphics()
	i_EnableLoadingDisk()
	v_RestoreBuffer()
	r_ExecuteSetViewSize()
	d_StartGameLoop()
	if testcontrols != 0 {
		wipegamestate = gamestate
	}
	doomgeneric_Tick()
}

// C documentation
//
//	//
//	// D_PageTicker
//	// Handles timing for warped projection
//	//
func d_PageTicker() {
	var v1 int32
	pagetic--
	v1 = pagetic
	if v1 < 0 {
		d_AdvanceDemo()
	}
}

// C documentation
//
//	//
//	// D_PageDrawer
//	//
func d_PageDrawer() {
	v_DrawPatch(0, 0, w_CacheLumpNameT(pagename))
}

// C documentation
//
//	//
//	// D_AdvanceDemo
//	// Called after each demo or intro demosequence finishes
//	//
func d_AdvanceDemo() {
	advancedemo = 1
}

// C documentation
//
//	//
//	// This cycles through the demo sequences.
//	// FIXME - version dependend demo numbers?
//	//
func d_DoAdvanceDemo() {
	players[consoleplayer].Fplayerstate = Pst_LIVE // not reborn
	advancedemo = 0
	usergame = 0 // no save / end game here
	paused = 0
	gameaction = ga_nothing
	// The Ultimate Doom executable changed the demo sequence to add
	// a DEMO4 demo.  Final Doom was based on Ultimate, so also
	// includes this change; however, the Final Doom IWADs do not
	// include a DEMO4 lump, so the game bombs out with an error
	// when it reaches this point in the demo sequence.
	// However! There is an alternate version of Final Doom that
	// includes a fixed executable.
	if gameversion == exe_ultimate || gameversion == exe_final {
		demosequence = (demosequence + 1) % 7
	} else {
		demosequence = (demosequence + 1) % 6
	}
	switch demosequence {
	case 0:
		if gamemode == commercial {
			pagetic = TICRATE * 11
		} else {
			pagetic = 170
		}
		gamestate = gs_DEMOSCREEN
		pagename = "TITLEPIC"
		if gamemode == commercial {
			s_StartMusic(int32(mus_dm2ttl))
		} else {
			s_StartMusic(int32(mus_intro))
		}
	case 1:
		g_DeferedPlayDemo("demo1")
	case 2:
		pagetic = 200
		gamestate = gs_DEMOSCREEN
		pagename = "CREDIT"
	case 3:
		g_DeferedPlayDemo("demo2")
	case 4:
		gamestate = gs_DEMOSCREEN
		if gamemode == commercial {
			pagetic = TICRATE * 11
			pagename = "TITLEPIC"
			s_StartMusic(int32(mus_dm2ttl))
		} else {
			pagetic = 200
			if gamemode == retail {
				pagename = "CREDIT"
			} else {
				pagename = "HELP2"
			}
		}
	case 5:
		g_DeferedPlayDemo("demo3")
		break
		// THE DEFINITIVE DOOM Special Edition demo
		fallthrough
	case 6:
		g_DeferedPlayDemo("demo4")
		break
	}
	// The Doom 3: BFG Edition version of doom2.wad does not have a
	// TITLETPIC lump. Use INTERPIC instead as a workaround.
	if bfgedition != 0 && strings.EqualFold(pagename, "TITLEPIC") && w_CheckNumForName("titlepic") < 0 {
		pagename = "INTERPIC"
	}
}

// C documentation
//
//	//
//	// D_StartTitle
//	//
func d_StartTitle() {
	gameaction = ga_nothing
	demosequence = -1
	d_AdvanceDemo()
}

// Strings for dehacked replacements of the startup banner
//
// These are from the original source: some of them are perhaps
// not used in any dehacked patches

var banners = [7]string{
	0: "                         DOOM 2: Hell on Earth v%d.%d                           ",
	1: "                            DOOM Shareware Startup v%d.%d                           ",
	2: "                            DOOM Registered Startup v%d.%d                           ",
	3: "                          DOOM System Startup v%d.%d                          ",
	4: "                         The Ultimate DOOM Startup v%d.%d                        ",
	5: "                     DOOM 2: TNT - Evilution v%d.%d                           ",
	6: "                   DOOM 2: Plutonia Experiment v%d.%d                           ",
}

//
// Get game name: if the startup banner has been replaced, use that.
// Otherwise, use the name given
//

func getGameName(gamename string) string {
	var deh_sub string
	var version, v2, v3, v6, v7 int32
	for i := 0; i < len(banners); i++ {
		// Has the banner been replaced?
		deh_sub = banners[i]
		if deh_sub != banners[i] {
			// Has been replaced.
			// We need to expand via printf to include the Doom version number
			// We also need to cut off spaces to get the basic name
			version = g_VanillaVersionCode()
			gamename = fmt.Sprintf(deh_sub, version/int32(100), version%int32(100))
			for {
				if len(gamename) >= 1 {
					v2 = int32(gamename[0])
					v3 = boolint32(v2 == ' ' || uint32(v2)-'\t' < 5)
					goto _4
				_4:
				}
				if !(len(gamename) >= 1 && v3 != 0) {
					break
				}
				gamename = gamename[1:]
			}
			for {
				if len(gamename) >= 1 {
					v6 = int32(gamename[len(gamename)-1])
					v7 = boolint32(v6 == ' ' || uint32(v6)-'\t' < 5)
					goto _8
				_8:
				}
				if !(len(gamename) >= 1 && v7 != 0) {
					break
				}
				gamename = gamename[:len(gamename)-1]
			}
			return gamename
		}
	}
	return gamename
}

func setMissionForPackName(pack_name string) {
	for i := range packs {
		if strings.EqualFold(pack_name, packs[i].Fname) {
			gamemission = packs[i].Fmission
			return
		}
	}
	fprintf_ccgo(os.Stdout, "Valid mission packs are:\n")
	for i := range packs {
		fprintf_ccgo(os.Stdout, "\t%s\n", packs[i].Fname)
	}
	i_Error("Unknown mission pack name: %s", pack_name)
}

var packs = [3]struct {
	Fname    string
	Fmission gamemission_t
}{
	0: {
		Fname:    "doom2",
		Fmission: doom2,
	},
	1: {
		Fname:    "tnt",
		Fmission: pack_tnt,
	},
	2: {
		Fname:    "plutonia",
		Fmission: pack_plut,
	},
}

//
// Find out what version of Doom is playing.
//

func d_IdentifyVersion() {
	var p int32
	var v2, v3 gamemission_t
	// gamemission is set up by the d_FindIWAD function.  But if
	// we specify '-iwad', we have to identify using
	// identifyIWADByName.  However, if the iwad does not match
	// any known IWAD name, we may have a dilemma.  Try to
	// identify by its contents.
	if gamemission == none {
		for i := uint32(0); i < numlumps; i++ {
			if strings.EqualFold(lumpinfo[i].Name(), "MAP01") {
				gamemission = doom2
				break
			} else {
				if strings.EqualFold(lumpinfo[i].Name(), "E1M1") {
					gamemission = doom
					break
				}
			}
		}
		if gamemission == none {
			// Still no idea.  I don't think this is going to work.
			i_Error("Unknown or invalid IWAD file.")
		}
	}
	// Make sure gamemode is set up correctly
	if gamemission == pack_chex {
		v2 = doom
	} else {
		if gamemission == pack_hacx {
			v3 = doom2
		} else {
			v3 = gamemission
		}
		v2 = v3
	}
	if v2 == doom {
		// Doom 1.  But which version?
		if w_CheckNumForName("E4M1") > 0 {
			// Ultimate Doom
			gamemode = retail
		} else {
			if w_CheckNumForName("E3M1") > 0 {
				gamemode = registered
			} else {
				gamemode = shareware
			}
		}
	} else {
		// Doom 2 of some kind.
		gamemode = commercial
		// We can manually override the gamemission that we got from the
		// IWAD detection code. This allows us to eg. play Plutonia 2
		// with Freedoom and get the right level names.
		//!
		// @arg <pack>
		//
		// Explicitly specify a Doom II "mission pack" to run as, instead of
		// detecting it based on the filename. Valid values are: "doom2",
		// "tnt" and "plutonia".
		//
		p = m_CheckParmWithArgs("-pack", 1)
		if p > 0 {
			setMissionForPackName(myargs[p+1])
		}
	}
}

// Set the gamedescription string

func d_SetGameDescription() {
	var is_freedm, is_freedoom boolean
	var v7, v5, v3, v1 gamemission_t
	is_freedoom = booluint32(w_CheckNumForName("FREEDOOM") >= 0)
	is_freedm = booluint32(w_CheckNumForName("FREEDM") >= 0)
	gamedescription = "Unknown"
	if gamemission == pack_chex {
		v1 = doom
	} else {
		if gamemission == pack_hacx {
			v1 = doom2
		} else {
			v1 = gamemission
		}
	}
	if v1 == doom {
		// Doom 1.  But which version?
		if is_freedoom != 0 {
			gamedescription = getGameName("Freedoom: Phase 1")
		} else {
			if gamemode == retail {
				// Ultimate Doom
				gamedescription = getGameName("The Ultimate DOOM")
			} else {
				if gamemode == registered {
					gamedescription = getGameName("DOOM Registered")
				} else {
					if gamemode == shareware {
						gamedescription = getGameName("DOOM Shareware")
					}
				}
			}
		}
	} else {
		// Doom 2 of some kind.  But which mission?
		if is_freedoom != 0 {
			if is_freedm != 0 {
				gamedescription = getGameName("FreeDM")
			} else {
				gamedescription = getGameName("Freedoom: Phase 2")
			}
		} else {
			if gamemission == pack_chex {
				v3 = doom
			} else {
				if gamemission == pack_hacx {
					v3 = doom2
				} else {
					v3 = gamemission
				}
			}
			if v3 == doom2 {
				gamedescription = getGameName("DOOM 2: Hell on Earth")
			} else {
				if gamemission == pack_chex {
					v5 = doom
				} else {
					if gamemission == pack_hacx {
						v5 = doom2
					} else {
						v5 = gamemission
					}
				}
				if v5 == pack_plut {
					gamedescription = getGameName("DOOM 2: Plutonia Experiment")
				} else {
					if gamemission == pack_chex {
						v7 = doom
					} else {
						if gamemission == pack_hacx {
							v7 = doom2
						} else {
							v7 = gamemission
						}
					}
					if v7 == pack_tnt {
						gamedescription = getGameName("DOOM 2: TNT - Evilution")
					}
				}
			}
		}
	}
}

func d_AddFile(filename string) boolean {
	var handle *os.File
	fprintf_ccgo(os.Stdout, " adding %s\n", filename)
	handle = w_AddFile(filename)
	return booluint32(handle != nil)
}

// Copyright message banners
// Some dehacked mods replace these.  These are only displayed if they are
// replaced by dehacked.

var copyright_banners = [3]string{
	0: "===========================================================================\nATTENTION:  This version of DOOM has been modified.  If you would like to\nget a copy of the original game, call 1-800-IDGAMES or see the readme file.\n        You will not receive technical support for modified games.\n                      press enter to continue\n===========================================================================\n",
	1: "===========================================================================\n                 Commercial product - do not distribute!\n         Please report software piracy to the SPA: 1-800-388-PIR8\n===========================================================================\n",
	2: "===========================================================================\n                                Shareware!\n===========================================================================\n",
}

// Prints a message only if it has been modified by dehacked.

func printDehackedBanners() {
	var deh_s string
	for i := 0; i < len(copyright_banners); i++ {
		deh_s = copyright_banners[i]
		if deh_s != copyright_banners[i] {
			// Make sure the modified banner always ends in a newline character.
			// If it doesn't, add a newline.  This fixes av.wad.
			if deh_s[len(deh_s)-1] != '\n' {
				deh_s += "\n"
			}
			fprintf_ccgo(os.Stdout, "%s", deh_s)
		}
	}
}

var gameversions = [10]struct {
	Fdescription string
	Fcmdline     string
	Fversion     gameversion_t
}{
	0: {
		Fdescription: "Doom 1.666",
		Fcmdline:     "1.666",
		Fversion:     exe_doom_1_666,
	},
	1: {
		Fdescription: "Doom 1.7/1.7a",
		Fcmdline:     "1.7",
		Fversion:     exe_doom_1_7,
	},
	2: {
		Fdescription: "Doom 1.8",
		Fcmdline:     "1.8",
		Fversion:     exe_doom_1_8,
	},
	3: {
		Fdescription: "Doom 1.9",
		Fcmdline:     "1.9",
		Fversion:     exe_doom_1_9,
	},
	4: {
		Fdescription: "Hacx",
		Fcmdline:     "hacx",
		Fversion:     exe_hacx,
	},
	5: {
		Fdescription: "Ultimate Doom",
		Fcmdline:     "ultimate",
		Fversion:     exe_ultimate,
	},
	6: {
		Fdescription: "Final Doom",
		Fcmdline:     "final",
		Fversion:     exe_final,
	},
	7: {
		Fdescription: "Final Doom (alt)",
		Fcmdline:     "final2",
		Fversion:     exe_final2,
	},
	8: {
		Fdescription: "Chex Quest",
		Fcmdline:     "chex",
		Fversion:     exe_chex,
	},
	9: {},
}

// Initialize the game version

func initGameVersion() {
	var i, p int32
	//!
	// @arg <version>
	// @category compat
	//
	// Emulate a specific version of Doom.  Valid values are "1.9",
	// "ultimate", "final", "final2", "hacx" and "chex".
	//
	p = m_CheckParmWithArgs("-gameversion", 1)
	if p != 0 {
		i = 0
		for {
			if gameversions[i].Fdescription == "" {
				break
			}
			if strings.EqualFold(myargs[p+1], gameversions[i].Fcmdline) {
				gameversion = gameversions[i].Fversion
				break
			}
			goto _1
		_1:
			;
			i++
		}
		if gameversions[i].Fdescription == "" {
			fprintf_ccgo(os.Stdout, "Supported game versions:\n")
			i = 0
			for {
				if gameversions[i].Fdescription == "" {
					break
				}
				fprintf_ccgo(os.Stdout, "\t%s (%s)\n", gameversions[i].Fcmdline, gameversions[i].Fdescription)
				goto _2
			_2:
				;
				i++
			}
			i_Error("Unknown game version '%s'", myargs[p+1])
		}
	} else {
		// Determine automatically
		if gamemission == pack_chex {
			// chex.exe - identified by iwad filename
			gameversion = exe_chex
		} else {
			if gamemission == pack_hacx {
				// hacx.exe: identified by iwad filename
				gameversion = exe_hacx
			} else {
				if gamemode == shareware || gamemode == registered {
					// original
					gameversion = exe_doom_1_9
					// TODO: Detect IWADs earlier than Doom v1.9.
				} else {
					if gamemode == retail {
						gameversion = exe_ultimate
					} else {
						if gamemode == commercial {
							if gamemission == doom2 {
								gameversion = exe_doom_1_9
							} else {
								// Final Doom: tnt or plutonia
								// Defaults to emulating the first Final Doom executable,
								// which has the crash in the demo loop; however, having
								// this as the default should mean that it plays back
								// most demos correctly.
								gameversion = exe_final
							}
						}
					}
				}
			}
		}
	}
	// The original exe does not support retail - 4th episode not supported
	if gameversion < exe_ultimate && gamemode == retail {
		gamemode = registered
	}
	// EXEs prior to the Final Doom exes do not support Final Doom.
	if gameversion < exe_final && gamemode == commercial && (gamemission == pack_tnt || gamemission == pack_plut) {
		gamemission = doom2
	}
}

func printGameVersion() {
	var i int32
	i = 0
	for {
		if gameversions[i].Fdescription == "" {
			break
		}
		if gameversions[i].Fversion == gameversion {
			fprintf_ccgo(os.Stdout, "Emulating the behavior of the '%s' executable.\n", gameversions[i].Fdescription)
			break
		}
		goto _1
	_1:
		;
		i++
	}
}

// Function called at exit to display the ENDOOM screen

func d_Endoom() {
	var endoom uintptr
	// Don't show ENDOOM if we have it disabled, or we're running
	// in screensaver or control test mode. Only show it once the
	// game has actually started.
	if show_endoom == 0 || main_loop_started == 0 || screensaver_mode != 0 || m_CheckParm("-testcontrols") > 0 {
		return
	}
	endoom = w_CacheLumpName("ENDOOM")
	i_Endoom(endoom)
	dg_exiting = true
}

// C documentation
//
//	//
//	// D_DoomMain
//	//
func d_DoomMain() {
	var argDemoName string
	var p, v1 int32
	i_AtExit(d_Endoom, 0)
	// print banner
	i_PrintBanner("Doom Generic 0.1")
	fprintf_ccgo(os.Stdout, "z_Init: Init zone memory allocation daemon. \n")
	z_Init()
	//!
	// @vanilla
	//
	// Disable monsters.
	//
	nomonsters = uint32(m_CheckParm("-nomonsters"))
	//!
	// @vanilla
	//
	// Monsters respawn after being killed.
	//
	respawnparm = uint32(m_CheckParm("-respawn"))
	//!
	// @vanilla
	//
	// Monsters move faster.
	//
	fastparm = uint32(m_CheckParm("-fast"))
	//!
	// @vanilla
	//
	// Developer mode.  F1 saves a screenshot in the current working
	// directory.
	//
	devparm = uint32(m_CheckParm("-devparm"))
	i_DisplayFPSDots(devparm)
	//!
	// @category net
	// @vanilla
	//
	// Start a deathmatch game.
	//
	if m_CheckParm("-deathmatch") != 0 {
		deathmatch = 1
	}
	//!
	// @category net
	// @vanilla
	//
	// Start a deathmatch 2.0 game.  Weapons do not stay in place and
	// all items respawn after 30 seconds.
	//
	if m_CheckParm("-altdeath") != 0 {
		deathmatch = 2
	}
	if devparm != 0 {
		fprintf_ccgo(os.Stdout, "Development mode ON.\n")
	}
	// find which dir to use for config files
	// Auto-detect the configuration dir.
	m_SetConfigDir("")
	//!
	// @arg <x>
	// @vanilla
	//
	// Turbo mode.  The player's speed is multiplied by x%.  If unspecified,
	// x defaults to 200.  Values are rounded up to 10 and down to 400.
	//
	v1 = m_CheckParm("-turbo")
	p = v1
	if v1 != 0 {
		scale := 200
		if p < int32(len(myargs)-1) {
			scale, _ = strconv.Atoi(myargs[p+1])
		}
		if scale < 10 {
			scale = 10
		}
		if scale > 400 {
			scale = 400
		}
		fprintf_ccgo(os.Stdout, "turbo scale: %d%%\n", scale)
		forwardmove[0] = forwardmove[0] * int32(scale) / 100
		forwardmove[1] = forwardmove[1] * int32(scale) / 100
		sidemove[0] = sidemove[0] * int32(scale) / 100
		sidemove[1] = sidemove[1] * int32(scale) / 100
	}
	// init subsystems
	fprintf_ccgo(os.Stdout, "v_Init: allocate screens.\n")
	v_Init()
	// Load configuration files before initialising other subsystems.
	fprintf_ccgo(os.Stdout, "m_LoadDefaults: Load system defaults.\n")
	m_SetConfigFilenames("default.cfg", "doomgenericdoom.cfg")
	d_BindVariables()
	m_LoadDefaults()
	// Save configuration at exit.
	i_AtExit(m_SaveDefaults, 0)
	// Find main IWAD file and load it.
	iwadfile = d_FindIWAD(1<<int32(doom)|1<<int32(doom2)|1<<int32(pack_tnt)|1<<int32(pack_plut)|1<<int32(pack_chex)|1<<int32(pack_hacx), &gamemission)
	// None found?
	if iwadfile == "" {
		i_Error("Game mode indeterminate.  No IWAD file was found.  Try\nspecifying one with the '-iwad' command line parameter.\n")
	}
	modifiedgame = 0
	fprintf_ccgo(os.Stdout, "W_Init: Init WADfiles.\n")
	d_AddFile(iwadfile)
	w_CheckCorrectIWAD(doom)
	// Now that we've loaded the IWAD, we can figure out what gamemission
	// we're playing and which version of Vanilla Doom we need to emulate.
	d_IdentifyVersion()
	initGameVersion()
	// Doom 3: BFG Edition includes modified versions of the classic
	// IWADs which can be identified by an additional DMENUPIC lump.
	// Furthermore, the M_GDHIGH lumps have been modified in a way that
	// makes them incompatible to Vanilla Doom and the modified version
	// of doom2.wad is missing the TITLEPIC lump.
	// We specifically check for DMENUPIC here, before PWADs have been
	// loaded which could probably include a lump of that name.
	if w_CheckNumForName("dmenupic") >= 0 {
		fprintf_ccgo(os.Stdout, "BFG Edition: Using workarounds as needed.\n")
		bfgedition = 1
		// BFG Edition changes the names of the secret levels to
		// censor the Wolfenstein references. It also has an extra
		// secret level (MAP33). In Vanilla Doom (meaning the DOS
		// version), MAP33 overflows into the Plutonia level names
		// array, so HUSTR_33 is actually PHUSTR_1.
		// The BFG edition doesn't have the "low detail" menu option (fair
		// enough). But bizarrely, it reuses the M_GDHIGH patch as a label
		// for the options menu (says "Fullscreen:"). Why the perpetrators
		// couldn't just add a new graphic lump and had to reuse this one,
		// I don't know.
		//
		// The end result is that M_GDHIGH is too wide and causes the game
		// to crash. As a workaround to get a minimum level of support for
		// the BFG edition IWADs, use the "ON"/"OFF" graphics instead.
	}
	// Load PWAD files.
	modifiedgame = w_ParseCommandLine()
	// Debug:
	//    W_PrintDirectory();
	//!
	// @arg <demo>
	// @category demo
	// @vanilla
	//
	// Play back the demo named demo.lmp.
	//
	p = m_CheckParmWithArgs("-playdemo", 1)
	if p == 0 {
		//!
		// @arg <demo>
		// @category demo
		// @vanilla
		//
		// Play back the demo named demo.lmp, determining the framerate
		// of the screen.
		//
		p = m_CheckParmWithArgs("-timedemo", 1)
	}
	if p != 0 {
		// With Vanilla you have to specify the file without extension,
		// but make that optional.
		var name string
		if strings.HasSuffix(myargs[p+1], ".lmp") {
			name = myargs[p+1]
		} else {
			name = fmt.Sprintf("%s.lmp", myargs[p+1])
		}
		if d_AddFile(name) != 0 {
			argDemoName = lumpinfo[numlumps-1].Name()
		} else {
			// If file failed to load, still continue trying to play
			// the demo in the same way as Vanilla Doom.  This makes
			// tricks like "-playdemo demo1" possible.
			argDemoName = myargs[p+1]
		}
		fprintf_ccgo(os.Stdout, "Playing demo %s.\n", name)
	}
	i_AtExit(g_CheckDemoStatus, 1)
	// Generate the WAD hash table.  Speed things up a bit.
	w_GenerateHashTable()
	// Load DEHACKED lumps from WAD files - but only if we give the right
	// command line parameter.
	// Set the gamedescription string. This is only possible now that
	// we've finished loading Dehacked patches.
	d_SetGameDescription()
	savegamedir = m_GetSaveGameDir(d_SaveGameIWADName(gamemission))
	// Check for -file in shareware
	if modifiedgame != 0 {
		// These are the lumps that will be checked in IWAD,
		// if any one is not present, execution will be aborted.
		levelLumps := [23]string{
			0:  "e2m1",
			1:  "e2m2",
			2:  "e2m3",
			3:  "e2m4",
			4:  "e2m5",
			5:  "e2m6",
			6:  "e2m7",
			7:  "e2m8",
			8:  "e2m9",
			9:  "e3m1",
			10: "e3m2",
			11: "e3m3",
			12: "e3m4",
			13: "e3m5",
			14: "e3m6",
			15: "e3m7",
			16: "e3m8",
			17: "e3m9",
			18: "dphoof",
			19: "bfgga0",
			20: "heada1",
			21: "cyrba1",
			22: "spida1d1",
		}
		if gamemode == shareware {
			i_Error("\nYou cannot -file with the shareware version. Register!")
		}
		// Check for fake IWAD with right name,
		// but w/o all the lumps of the registered version.
		if gamemode == registered {
			for i := 0; i < len(levelLumps); i++ {
				if w_CheckNumForName(levelLumps[i]) < 0 {
					i_Error("\nThis is not the registered version.")
				}
			}
		}
	}
	if w_CheckNumForName("SS_START") >= 0 || w_CheckNumForName("FF_END") >= 0 {
		i_PrintDivider()
		fprintf_ccgo(os.Stdout, " WARNING: The loaded WAD file contains modified sprites or\n floor textures.  You may want to use the '-merge' command\n line option instead of '-file'.\n")
	}
	i_PrintStartupBanner(gamedescription)
	printDehackedBanners()
	// Freedoom's IWADs are Boom-compatible, which means they usually
	// don't work in Vanilla (though FreeDM is okay). Show a warning
	// message and give a link to the website.
	if w_CheckNumForName("FREEDOOM") >= 0 && w_CheckNumForName("FREEDM") < 0 {
		fprintf_ccgo(os.Stdout, " WARNING: You are playing using one of the Freedoom IWAD\n files, which might not work in this port. See this page\n for more information on how to play using Freedoom:\n   http://www.chocolate-doom.org/wiki/index.php/Freedoom\n")
		i_PrintDivider()
	}
	fprintf_ccgo(os.Stdout, "I_Init: Setting up machine state.\n")
	i_CheckIsScreensaver()
	i_InitSound(1)
	i_InitMusic()
	// Initial netgame startup. Connect to server etc.
	d_ConnectNetGame()
	// get skill / episode / map from parms
	startskill = sk_medium
	startepisode = 1
	startmap = 1
	autostart = 0
	//!
	// @arg <skill>
	// @vanilla
	//
	// Set the game skill, 1-5 (1: easiest, 5: hardest).  A skill of
	// 0 disables all monsters.
	//
	p = m_CheckParmWithArgs("-skill", 1)
	if p != 0 {
		startskill = skill_t(myargs[p+1][0] - '1')
		autostart = 1
	}
	//!
	// @arg <n>
	// @vanilla
	//
	// Start playing on episode n (1-4)
	//
	p = m_CheckParmWithArgs("-episode", 1)
	if p != 0 {
		startepisode = int32(myargs[p+1][0] - '0')
		startmap = 1
		autostart = 1
	}
	timelimit = 0
	//!
	// @arg <n>
	// @category net
	// @vanilla
	//
	// For multiplayer games: exit each level after n minutes.
	//
	p = m_CheckParmWithArgs("-timer", 1)
	if p != 0 {
		v, _ := strconv.Atoi(myargs[p+1])
		timelimit = int32(v)
	}
	//!
	// @category net
	// @vanilla
	//
	// Austin Virtual Gaming: end levels after 20 minutes.
	//
	p = m_CheckParm("-avg")
	if p != 0 {
		timelimit = 20
	}
	//!
	// @arg [<x> <y> | <xy>]
	// @vanilla
	//
	// Start a game immediately, warping to ExMy (Doom 1) or MAPxy
	// (Doom 2)
	//
	p = m_CheckParmWithArgs("-warp", 1)
	if p != 0 {
		if gamemode == commercial {
			v, _ := strconv.Atoi(myargs[p+1])
			startmap = int32(v)
		} else {
			startepisode = int32(myargs[p+1][0] - '0')
			if p+2 < int32(len(myargs)) {
				startmap = int32(myargs[p+2][0] - '0')
			} else {
				startmap = 1
			}
		}
		autostart = 1
	}
	// Undocumented:
	// Invoked by setup to test the controls.
	p = m_CheckParm("-testcontrols")
	if p > 0 {
		startepisode = 1
		startmap = 1
		autostart = 1
		testcontrols = 1
	}
	// Check for load game parameter
	// We do this here and save the slot number, so that the network code
	// can override it or send the load slot to other players.
	//!
	// @arg <s>
	// @vanilla
	//
	// Load the game in slot s.
	//
	p = m_CheckParmWithArgs("-loadgame", 1)
	if p != 0 {
		v, _ := strconv.Atoi(myargs[p+1])
		startloadgame = int32(v)
	} else {
		// Not loading a game
		startloadgame = -1
	}
	fprintf_ccgo(os.Stdout, "m_Init: Init miscellaneous info.\n")
	m_Init()
	fprintf_ccgo(os.Stdout, "r_Init: Init DOOM refresh daemon - ")
	r_Init()
	fprintf_ccgo(os.Stdout, "\nP_Init: Init Playloop state.\n")
	p_Init()
	fprintf_ccgo(os.Stdout, "s_Init: Setting up sound.\n")
	s_Init(sfxVolume*8, musicVolume*8)
	fprintf_ccgo(os.Stdout, "d_CheckNetGame: Checking network game status.\n")
	d_CheckNetGame()
	printGameVersion()
	fprintf_ccgo(os.Stdout, "hu_Init: Setting up heads up display.\n")
	hu_Init()
	fprintf_ccgo(os.Stdout, "st_Init: Init status bar.\n")
	st_Init()
	// If Doom II without a MAP01 lump, this is a store demo.
	// Moved this here so that MAP01 isn't constantly looked up
	// in the main loop.
	if gamemode == commercial && w_CheckNumForName("map01") < 0 {
		storedemo = 1
	}
	if m_CheckParmWithArgs("-statdump", 1) != 0 {
		i_AtExit(statDump, 1)
		fprintf_ccgo(os.Stdout, "External statistics registered.\n")
	}
	//!
	// @arg <x>
	// @category demo
	// @vanilla
	//
	// Record a demo named x.lmp.
	//
	p = m_CheckParmWithArgs("-record", 1)
	if p != 0 {
		g_RecordDemo(myargs[p+1])
		autostart = 1
	}
	p = m_CheckParmWithArgs("-playdemo", 1)
	if p != 0 {
		singledemo = 1 // quit after one demo
		g_DeferedPlayDemo(argDemoName)
		d_DoomLoop()
		return
	}
	p = m_CheckParmWithArgs("-timedemo", 1)
	if p != 0 {
		g_TimeDemo(argDemoName)
		d_DoomLoop()
		return
	}
	if startloadgame >= 0 {
		g_LoadGame(p_SaveGameFile(startloadgame))
	}
	if gameaction != ga_loadgame {
		if autostart != 0 || netgame != 0 {
			g_InitNew(startskill, startepisode, startmap)
		} else {
			d_StartTitle()
		} // start up intro loop
	}
	d_DoomLoop()
}

func d_GameMissionString(mission gamemission_t) string {
	switch mission {
	case none:
		fallthrough
	default:
		return "none"
	case doom:
		return "doom"
	case doom2:
		return "doom2"
	case pack_tnt:
		return "tnt"
	case pack_plut:
		return "plutonia"
	case pack_hacx:
		return "hacx"
	case pack_chex:
		return "chex"
	case heretic:
		return "heretic"
	case hexen:
		return "hexen"
	case strife:
		return "strife"
	}
	return ""
}

const ANG2701 = 3221225472
const ANG901 = 1073741824

// Called when a player leaves the game

func playerQuitGame(player *player_t) {
	player_num := playerIndex(player)
	// Do this the same way as Vanilla Doom does, to allow dehacked
	// replacements of this message
	exitmsg = fmt.Sprintf("Player %d left the game", player_num+1)
	playeringame[player_num] = 0
	players[consoleplayer].Fmessage = exitmsg
	// TODO: check if it is sensible to do this:
	if demorecording != 0 {
		g_CheckDemoStatus()
	}
}

var exitmsg string

func runTic(cmds []ticcmd_t, ingame []boolean) {
	// Check for player quits.
	for i := 0; i < MAXPLAYERS; i++ {
		if demoplayback == 0 && playeringame[i] != 0 && ingame[i] == 0 {
			playerQuitGame(&players[i])
		}
	}
	netcmds = cmds
	// check that there are players in the game.  if not, we cannot
	// run a tic.
	if advancedemo != 0 {
		d_DoAdvanceDemo()
	}
	g_Ticker()
}

var doom_loop_interface = loop_interface_t{
	d_ProcessEvents,
	g_BuildTiccmd,
	runTic,
	m_Ticker,
}

// Load game settings from the specified structure and
// set global variables.

func loadGameSettings(settings *net_gamesettings_t) {
	deathmatch = settings.Fdeathmatch
	startepisode = settings.Fepisode
	startmap = settings.Fmap1
	startskill = settings.Fskill
	startloadgame = settings.Floadgame
	lowres_turn = uint32(settings.Flowres_turn)
	nomonsters = uint32(settings.Fnomonsters)
	fastparm = uint32(settings.Ffast_monsters)
	respawnparm = uint32(settings.Frespawn_monsters)
	timelimit = settings.Ftimelimit
	consoleplayer = settings.Fconsoleplayer
	if lowres_turn != 0 {
		fprintf_ccgo(os.Stdout, "NOTE: Turning resolution is reduced; this is probably because there is a client recording a Vanilla demo.\n")
	}
	for i := 0; i < MAXPLAYERS; i++ {
		playeringame[i] = booluint32(i < int(settings.Fnum_players))
	}
}

// Save the game settings from global variables to the specified
// game settings structure.

func saveGameSettings(settings *net_gamesettings_t) {
	// Fill in game settings structure with appropriate parameters
	// for the new game
	settings.Fdeathmatch = deathmatch
	settings.Fepisode = startepisode
	settings.Fmap1 = startmap
	settings.Fskill = startskill
	settings.Floadgame = startloadgame
	settings.Fgameversion = gameversion
	settings.Fnomonsters = int32(nomonsters)
	settings.Ffast_monsters = int32(fastparm)
	settings.Frespawn_monsters = int32(respawnparm)
	settings.Ftimelimit = timelimit
	settings.Flowres_turn = boolint32(m_CheckParm("-record") > 0 && m_CheckParm("-longtics") == 0)
}

func initConnectData(connect_data *net_connect_data_t) {
	connect_data.Fmax_players = MAXPLAYERS
	connect_data.Fdrone = 0
	//!
	// @category net
	//
	// Run as the left screen in three screen mode.
	//
	if m_CheckParm("-left") > 0 {
		viewangleoffset = ANG901
		connect_data.Fdrone = 1
	}
	//!
	// @category net
	//
	// Run as the right screen in three screen mode.
	//
	if m_CheckParm("-right") > 0 {
		viewangleoffset = ANG2701
		connect_data.Fdrone = 1
	}
	//
	// Connect data
	//
	// Game type fields:
	connect_data.Fgamemode = gamemode
	connect_data.Fgamemission = gamemission
	// Are we recording a demo? Possibly set lowres turn mode
	connect_data.Flowres_turn = boolint32(m_CheckParm("-record") > 0 && m_CheckParm("-longtics") == 0)
	// Read checksums of our WAD directory and dehacked information
	w_Checksum(&connect_data.Fwad_sha1sum)
	// Are we playing with the Freedoom IWAD?
	connect_data.Fis_freedoom = boolint32(w_CheckNumForName("FREEDOOM") >= 0)
}

func d_ConnectNetGame() {
	connect_data := &net_connect_data_t{}
	initConnectData(connect_data)
	netgame = d_InitNetGame(connect_data)
	//!
	// @category net
	//
	// Start the game playing as though in a netgame with a single
	// player.  This can also be used to play back single player netgame
	// demos.
	//
	if m_CheckParm("-solo-net") > 0 {
		netgame = 1
	}
}

// C documentation
//
//	//
//	// D_CheckNetGame
//	// Works out player numbers among the net participants
//	//
func d_CheckNetGame() {
	settings := &net_gamesettings_t{}
	if netgame != 0 {
		autostart = 1
	}
	d_RegisterLoopCallbacks(&doom_loop_interface)
	saveGameSettings(settings)
	d_StartNetGame(settings, 0)
	loadGameSettings(settings)
	fprintf_ccgo(os.Stdout, "startskill %d  deathmatch: %d  startmap: %d  startepisode: %d\n", startskill, deathmatch, startmap, startepisode)
	fprintf_ccgo(os.Stdout, "player %d of %d (%d nodes)\n", consoleplayer+1, settings.Fnum_players, settings.Fnum_players)
	// Show players here; the server might have specified a time limit
	if timelimit > 0 && deathmatch != 0 {
		// Gross hack to work like Vanilla:
		if timelimit == 20 && m_CheckParm("-avg") != 0 {
			fprintf_ccgo(os.Stdout, "Austin Virtual Gaming: Levels will end after 20 minutes\n")
		} else {
			fprintf_ccgo(os.Stdout, "Levels will end after %d minute", timelimit)
			if timelimit > 1 {
				fprintf_ccgo(os.Stdout, "s")
			}
			fprintf_ccgo(os.Stdout, ".\n")
		}
	}
}

const FF_FRAMEMASK1 = 32767
const TEXTSPEED = 3
const TEXTWAIT = 250

type finalestage_t = int32

const F_STAGE_TEXT = 0
const F_STAGE_ARTSCREEN = 1
const F_STAGE_CAST = 2

type textscreen_t struct {
	Fmission    gamemission_t
	Fepisode    int32
	Flevel      int32
	Fbackground string
	Ftext       string
}

var textscreens = [22]textscreen_t{
	0: {
		Fepisode:    1,
		Flevel:      8,
		Fbackground: "FLOOR4_8",
		Ftext:       "Once you beat the big badasses and\nclean out the moon base you're supposed\nto win, aren't you? Aren't you? Where's\nyour fat reward and ticket home? What\nthe hell is this? It's not supposed to\nend this way!\n\nIt stinks like rotten meat, but looks\nlike the lost Deimos base.  Looks like\nyou're stuck on The Shores of Hell.\nThe only way out is through.\n\nTo continue the DOOM experience, play\nThe Shores of Hell and its amazing\nsequel, Inferno!\n",
	},
	1: {
		Fepisode:    2,
		Flevel:      8,
		Fbackground: "SFLR6_1",
		Ftext:       "You've done it! The hideous cyber-\ndemon lord that ruled the lost Deimos\nmoon base has been slain and you\nare triumphant! But ... where are\nyou? You clamber to the edge of the\nmoon and look down to see the awful\ntruth.\n\nDeimos floats above Hell itself!\nYou've never heard of anyone escaping\nfrom Hell, but you'll make the bastards\nsorry they ever heard of you! Quickly,\nyou rappel down to  the surface of\nHell.\n\nNow, it's on to the final chapter of\nDOOM! -- Inferno.",
	},
	2: {
		Fepisode:    3,
		Flevel:      8,
		Fbackground: "MFLR8_4",
		Ftext:       "The loathsome spiderdemon that\nmasterminded the invasion of the moon\nbases and caused so much death has had\nits ass kicked for all time.\n\nA hidden doorway opens and you enter.\nYou've proven too tough for Hell to\ncontain, and now Hell at last plays\nfair -- for you emerge from the door\nto see the green fields of Earth!\nHome at last.\n\nYou wonder what's been happening on\nEarth while you were battling evil\nunleashed. It's good that no Hell-\nspawn could have come through that\ndoor with you ...",
	},
	3: {
		Fepisode:    4,
		Flevel:      8,
		Fbackground: "MFLR8_3",
		Ftext:       "the spider mastermind must have sent forth\nits legions of hellspawn before your\nfinal confrontation with that terrible\nbeast from hell.  but you stepped forward\nand brought forth eternal damnation and\nsuffering upon the horde as a true hero\nwould in the face of something so evil.\n\nbesides, someone was gonna pay for what\nhappened to daisy, your pet rabbit.\n\nbut now, you see spread before you more\npotential pain and gibbitude as a nation\nof demons run amok among our cities.\n\nnext stop, hell on earth!",
	},
	4: {
		Fmission:    doom2,
		Fepisode:    1,
		Flevel:      6,
		Fbackground: "SLIME16",
		Ftext:       "YOU HAVE ENTERED DEEPLY INTO THE INFESTED\nSTARPORT. BUT SOMETHING IS WRONG. THE\nMONSTERS HAVE BROUGHT THEIR OWN REALITY\nWITH THEM, AND THE STARPORT'S TECHNOLOGY\nIS BEING SUBVERTED BY THEIR PRESENCE.\n\nAHEAD, YOU SEE AN OUTPOST OF HELL, A\nFORTIFIED ZONE. IF YOU CAN GET PAST IT,\nYOU CAN PENETRATE INTO THE HAUNTED HEART\nOF THE STARBASE AND FIND THE CONTROLLING\nSWITCH WHICH HOLDS EARTH'S POPULATION\nHOSTAGE.",
	},
	5: {
		Fmission:    doom2,
		Fepisode:    1,
		Flevel:      11,
		Fbackground: "RROCK14",
		Ftext:       "YOU HAVE WON! YOUR VICTORY HAS ENABLED\nHUMANKIND TO EVACUATE EARTH AND ESCAPE\nTHE NIGHTMARE.  NOW YOU ARE THE ONLY\nHUMAN LEFT ON THE FACE OF THE PLANET.\nCANNIBAL MUTATIONS, CARNIVOROUS ALIENS,\nAND EVIL SPIRITS ARE YOUR ONLY NEIGHBORS.\nYOU SIT BACK AND WAIT FOR DEATH, CONTENT\nTHAT YOU HAVE SAVED YOUR SPECIES.\n\nBUT THEN, EARTH CONTROL BEAMS DOWN A\nMESSAGE FROM SPACE: \"SENSORS HAVE LOCATED\nTHE SOURCE OF THE ALIEN INVASION. IF YOU\nGO THERE, YOU MAY BE ABLE TO BLOCK THEIR\nENTRY.  THE ALIEN BASE IS IN THE HEART OF\nYOUR OWN HOME CITY, NOT FAR FROM THE\nSTARPORT.\" SLOWLY AND PAINFULLY YOU GET\nUP AND RETURN TO THE FRAY.",
	},
	6: {
		Fmission:    doom2,
		Fepisode:    1,
		Flevel:      20,
		Fbackground: "RROCK07",
		Ftext:       "YOU ARE AT THE CORRUPT HEART OF THE CITY,\nSURROUNDED BY THE CORPSES OF YOUR ENEMIES.\nYOU SEE NO WAY TO DESTROY THE CREATURES'\nENTRYWAY ON THIS SIDE, SO YOU CLENCH YOUR\nTEETH AND PLUNGE THROUGH IT.\n\nTHERE MUST BE A WAY TO CLOSE IT ON THE\nOTHER SIDE. WHAT DO YOU CARE IF YOU'VE\nGOT TO GO THROUGH HELL TO GET TO IT?",
	},
	7: {
		Fmission:    doom2,
		Fepisode:    1,
		Flevel:      30,
		Fbackground: "RROCK17",
		Ftext:       "THE HORRENDOUS VISAGE OF THE BIGGEST\nDEMON YOU'VE EVER SEEN CRUMBLES BEFORE\nYOU, AFTER YOU PUMP YOUR ROCKETS INTO\nHIS EXPOSED BRAIN. THE MONSTER SHRIVELS\nUP AND DIES, ITS THRASHING LIMBS\nDEVASTATING UNTOLD MILES OF HELL'S\nSURFACE.\n\nYOU'VE DONE IT. THE INVASION IS OVER.\nEARTH IS SAVED. HELL IS A WRECK. YOU\nWONDER WHERE BAD FOLKS WILL GO WHEN THEY\nDIE, NOW. WIPING THE SWEAT FROM YOUR\nFOREHEAD YOU BEGIN THE LONG TREK BACK\nHOME. REBUILDING EARTH OUGHT TO BE A\nLOT MORE FUN THAN RUINING IT WAS.\n",
	},
	8: {
		Fmission:    doom2,
		Fepisode:    1,
		Flevel:      15,
		Fbackground: "RROCK13",
		Ftext:       "CONGRATULATIONS, YOU'VE FOUND THE SECRET\nLEVEL! LOOKS LIKE IT'S BEEN BUILT BY\nHUMANS, RATHER THAN DEMONS. YOU WONDER\nWHO THE INMATES OF THIS CORNER OF HELL\nWILL BE.",
	},
	9: {
		Fmission:    doom2,
		Fepisode:    1,
		Flevel:      31,
		Fbackground: "RROCK19",
		Ftext:       "CONGRATULATIONS, YOU'VE FOUND THE\nSUPER SECRET LEVEL!  YOU'D BETTER\nBLAZE THROUGH THIS ONE!\n",
	},
	10: {
		Fmission:    pack_tnt,
		Fepisode:    1,
		Flevel:      6,
		Fbackground: "SLIME16",
		Ftext:       "You've fought your way out of the infested\nexperimental labs.   It seems that UAC has\nonce again gulped it down.  With their\nhigh turnover, it must be hard for poor\nold UAC to buy corporate health insurance\nnowadays..\n\nAhead lies the military complex, now\nswarming with diseased horrors hot to get\ntheir teeth into you. With luck, the\ncomplex still has some warlike ordnance\nlaying around.",
	},
	11: {
		Fmission:    pack_tnt,
		Fepisode:    1,
		Flevel:      11,
		Fbackground: "RROCK14",
		Ftext:       "You hear the grinding of heavy machinery\nahead.  You sure hope they're not stamping\nout new hellspawn, but you're ready to\nream out a whole herd if you have to.\nThey might be planning a blood feast, but\nyou feel about as mean as two thousand\nmaniacs packed into one mad killer.\n\nYou don't plan to go down easy.",
	},
	12: {
		Fmission:    pack_tnt,
		Fepisode:    1,
		Flevel:      20,
		Fbackground: "RROCK07",
		Ftext:       "The vista opening ahead looks real damn\nfamiliar. Smells familiar, too -- like\nfried excrement. You didn't like this\nplace before, and you sure as hell ain't\nplanning to like it now. The more you\nbrood on it, the madder you get.\nHefting your gun, an evil grin trickles\nonto your face. Time to take some names.",
	},
	13: {
		Fmission:    pack_tnt,
		Fepisode:    1,
		Flevel:      30,
		Fbackground: "RROCK17",
		Ftext:       "Suddenly, all is silent, from one horizon\nto the other. The agonizing echo of Hell\nfades away, the nightmare sky turns to\nblue, the heaps of monster corpses start \nto evaporate along with the evil stench \nthat filled the air. Jeeze, maybe you've\ndone it. Have you really won?\n\nSomething rumbles in the distance.\nA blue light begins to glow inside the\nruined skull of the demon-spitter.",
	},
	14: {
		Fmission:    pack_tnt,
		Fepisode:    1,
		Flevel:      15,
		Fbackground: "RROCK13",
		Ftext:       "What now? Looks totally different. Kind\nof like King Tut's condo. Well,\nwhatever's here can't be any worse\nthan usual. Can it?  Or maybe it's best\nto let sleeping gods lie..",
	},
	15: {
		Fmission:    pack_tnt,
		Fepisode:    1,
		Flevel:      31,
		Fbackground: "RROCK19",
		Ftext:       "Time for a vacation. You've burst the\nbowels of hell and by golly you're ready\nfor a break. You mutter to yourself,\nMaybe someone else can kick Hell's ass\nnext time around. Ahead lies a quiet town,\nwith peaceful flowing water, quaint\nbuildings, and presumably no Hellspawn.\n\nAs you step off the transport, you hear\nthe stomp of a cyberdemon's iron shoe.",
	},
	16: {
		Fmission:    pack_plut,
		Fepisode:    1,
		Flevel:      6,
		Fbackground: "SLIME16",
		Ftext:       "You gloat over the steaming carcass of the\nGuardian.  With its death, you've wrested\nthe Accelerator from the stinking claws\nof Hell.  You relax and glance around the\nroom.  Damn!  There was supposed to be at\nleast one working prototype, but you can't\nsee it. The demons must have taken it.\n\nYou must find the prototype, or all your\nstruggles will have been wasted. Keep\nmoving, keep fighting, keep killing.\nOh yes, keep living, too.",
	},
	17: {
		Fmission:    pack_plut,
		Fepisode:    1,
		Flevel:      11,
		Fbackground: "RROCK14",
		Ftext:       "Even the deadly Arch-Vile labyrinth could\nnot stop you, and you've gotten to the\nprototype Accelerator which is soon\nefficiently and permanently deactivated.\n\nYou're good at that kind of thing.",
	},
	18: {
		Fmission:    pack_plut,
		Fepisode:    1,
		Flevel:      20,
		Fbackground: "RROCK07",
		Ftext:       "You've bashed and battered your way into\nthe heart of the devil-hive.  Time for a\nSearch-and-Destroy mission, aimed at the\nGatekeeper, whose foul offspring is\ncascading to Earth.  Yeah, he's bad. But\nyou know who's worse!\n\nGrinning evilly, you check your gear, and\nget ready to give the bastard a little Hell\nof your own making!",
	},
	19: {
		Fmission:    pack_plut,
		Fepisode:    1,
		Flevel:      30,
		Fbackground: "RROCK17",
		Ftext:       "The Gatekeeper's evil face is splattered\nall over the place.  As its tattered corpse\ncollapses, an inverted Gate forms and\nsucks down the shards of the last\nprototype Accelerator, not to mention the\nfew remaining demons.  You're done. Hell\nhas gone back to pounding bad dead folks \ninstead of good live ones.  Remember to\ntell your grandkids to put a rocket\nlauncher in your coffin. If you go to Hell\nwhen you die, you'll need it for some\nfinal cleaning-up ...",
	},
	20: {
		Fmission:    pack_plut,
		Fepisode:    1,
		Flevel:      15,
		Fbackground: "RROCK13",
		Ftext:       "You've found the second-hardest level we\ngot. Hope you have a saved game a level or\ntwo previous.  If not, be prepared to die\naplenty. For master marines only.",
	},
	21: {
		Fmission:    pack_plut,
		Fepisode:    1,
		Flevel:      31,
		Fbackground: "RROCK19",
		Ftext:       "Betcha wondered just what WAS the hardest\nlevel we had ready for ya?  Now you know.\nNo one gets out alive.",
	},
}

// C documentation
//
//	//
//	// F_StartFinale
//	//
func f_StartFinale() {
	var screen *textscreen_t
	var v1, v4, v6 gamemission_t
	var v8 bool
	gameaction = ga_nothing
	gamestate = gs_FINALE
	viewactive = 0
	automapactive = 0
	if gamemission == pack_chex {
		v1 = doom
	} else {
		if gamemission == pack_hacx {
			v1 = doom2
		} else {
			v1 = gamemission
		}
	}
	if v1 == doom {
		s_ChangeMusic(int32(mus_victor), 1)
	} else {
		s_ChangeMusic(int32(mus_read_m), 1)
	}
	// Find the right screen and set the text and background
	for i := 0; i < len(textscreens); i++ {
		screen = &textscreens[i]
		// Hack for Chex Quest
		if gameversion == exe_chex && screen.Fmission == doom {
			screen.Flevel = 5
		}
		if gamemission == pack_chex {
			v4 = doom
		} else {
			if gamemission == pack_hacx {
				v4 = doom2
			} else {
				v4 = gamemission
			}
		}
		if v8 = v4 == screen.Fmission; v8 {
			if gamemission == pack_chex {
				v6 = doom
			} else {
				if gamemission == pack_hacx {
					v6 = doom2
				} else {
					v6 = gamemission
				}
			}
		}
		if v8 && (v6 != doom || gameepisode == screen.Fepisode) && gamemap == screen.Flevel {
			finaletext = screen.Ftext
			finaleflat = screen.Fbackground
		}
	}
	// Do dehacked substitutions of strings
	finalestage = F_STAGE_TEXT
	finalecount = 0
}

func f_Responder(event *event_t) boolean {
	if finalestage == F_STAGE_CAST {
		return f_CastResponder(event)
	}
	return 0
}

// C documentation
//
//	//
//	// F_Ticker
//	//
func f_Ticker() {
	var i uint64
	// check for skipping
	if gamemode == commercial && finalecount > 50 {
		// go on to the next level
		for i = 0; i < MAXPLAYERS; i++ {
			if players[i].Fcmd.Fbuttons != 0 {
				break
			}
		}
		if i < uint64(MAXPLAYERS) {
			if gamemap == 30 {
				f_StartCast()
			} else {
				gameaction = ga_worlddone
			}
		}
	}
	// advance animation
	finalecount++
	if finalestage == F_STAGE_CAST {
		f_CastTicker()
		return
	}
	if gamemode == commercial {
		return
	}
	if finalestage == F_STAGE_TEXT && uint64(finalecount) > uint64(len(finaletext)*TEXTSPEED+TEXTWAIT) {
		finalecount = 0
		finalestage = F_STAGE_ARTSCREEN
		wipegamestate = -1 // force a wipe
		if gameepisode == 3 {
			s_StartMusic(int32(mus_bunny))
		}
	}
}

func f_TextWrite() {
	var c, count, cx, cy, w int32
	var pos int
	// erase the entire screen to a tiled background
	src := w_CacheLumpNameBytes(finaleflat)
	destPos := 0
	for y := 0; y < SCREENHEIGHT; y++ {
		for x := 0; x < SCREENWIDTH/64; x++ {
			copy(I_VideoBuffer[destPos:], src[uintptr(y&63<<6):uintptr(y&63<<6)+64])
			destPos += 64
		}
		if SCREENWIDTH&63 != 0 {
			length := SCREENWIDTH & 63
			copy(I_VideoBuffer[destPos:destPos+length], src[y&63<<6:])
			destPos += length
		}
	}
	v_MarkRect(0, 0, SCREENWIDTH, SCREENHEIGHT)
	// draw some of the text onto the screen
	cx = 10
	cy = 10
	pos = 0
	count = (int32(finalecount) - 10) / TEXTSPEED
	if count < 0 {
		count = 0
	}
	for {
		if count == 0 {
			break
		}
		c = int32(finaletext[pos])
		pos++
		if c == 0 {
			break
		}
		if c == '\n' {
			cx = 10
			cy += 11
			goto _3
		}
		c = xtoupper(c) - '!'
		if c < 0 || c > '_'-'!'+1 {
			cx += 4
			goto _3
		}
		w = int32(hu_font[c].Fwidth)
		if cx+w > SCREENWIDTH {
			break
		}
		v_DrawPatch(cx, cy, hu_font[c])
		cx += w
		goto _3
	_3:
		;
		count--
	}
}

// C documentation
//
//	//
//	// Final DOOM 2 animation
//	// Casting by id Software.
//	//   in order of appearance
//	//
type castinfo_t struct {
	Fname  string
	Ftype1 mobjtype_t
}

func init() {
	castorder = [18]castinfo_t{
		0: {
			Fname:  "ZOMBIEMAN",
			Ftype1: mt_POSSESSED,
		},
		1: {
			Fname:  "SHOTGUN GUY",
			Ftype1: mt_SHOTGUY,
		},
		2: {
			Fname:  "HEAVY WEAPON DUDE",
			Ftype1: mt_CHAINGUY,
		},
		3: {
			Fname:  "IMP",
			Ftype1: mt_TROOP,
		},
		4: {
			Fname:  "DEMON",
			Ftype1: mt_SERGEANT,
		},
		5: {
			Fname:  "LOST SOUL",
			Ftype1: mt_SKULL,
		},
		6: {
			Fname:  "CACODEMON",
			Ftype1: mt_HEAD,
		},
		7: {
			Fname:  "HELL KNIGHT",
			Ftype1: mt_KNIGHT,
		},
		8: {
			Fname:  "BARON OF HELL",
			Ftype1: mt_BRUISER,
		},
		9: {
			Fname:  "ARACHNOTRON",
			Ftype1: mt_BABY,
		},
		10: {
			Fname:  "PAIN ELEMENTAL",
			Ftype1: mt_PAIN,
		},
		11: {
			Fname:  "REVENANT",
			Ftype1: mt_UNDEAD,
		},
		12: {
			Fname:  "MANCUBUS",
			Ftype1: mt_FATSO,
		},
		13: {
			Fname:  "ARCH-VILE",
			Ftype1: mt_VILE,
		},
		14: {
			Fname:  "THE SPIDER MASTERMIND",
			Ftype1: mt_SPIDER,
		},
		15: {
			Fname:  "THE CYBERDEMON",
			Ftype1: mt_CYBORG,
		},
		16: {
			Fname: "OUR HERO",
		},
		17: {},
	}
}

// C documentation
//
//	//
//	// F_StartCast
//	//
func f_StartCast() {
	wipegamestate = -1 // force a screen wipe
	castnum = 0
	caststate = &states[mobjinfo[castorder[castnum].Ftype1].Fseestate]
	casttics = caststate.Ftics
	castdeath = 0
	finalestage = F_STAGE_CAST
	castframes = 0
	castonmelee = 0
	castattacking = 0
	s_ChangeMusic(int32(mus_evil), 1)
}

// C documentation
//
//	//
//	// F_CastTicker
//	//
func f_CastTicker() {
	var sfx, st, v1 int32
	casttics--
	v1 = casttics
	if v1 > 0 {
		return
	} // not time to change state yet
	if caststate.Ftics == -1 || caststate.Fnextstate == s_NULL {
		// switch from deathstate to next monster
		castnum++
		castdeath = 0
		if castorder[castnum].Fname == "" {
			castnum = 0
		}
		if mobjinfo[castorder[castnum].Ftype1].Fseesound != 0 {
			s_StartSound(nil, mobjinfo[castorder[castnum].Ftype1].Fseesound)
		}
		caststate = &states[mobjinfo[castorder[castnum].Ftype1].Fseestate]
		castframes = 0
	} else {
		// just advance to next state in animation
		if caststate == &states[s_PLAY_ATK1] {
			goto stopattack
		} // Oh, gross hack!
		st = caststate.Fnextstate
		caststate = &states[st]
		castframes++
		// sound hacks....
		switch st {
		case s_PLAY_ATK1:
			sfx = int32(sfx_dshtgn)
		case s_POSS_ATK2:
			sfx = int32(sfx_pistol)
		case s_SPOS_ATK2:
			sfx = int32(sfx_shotgn)
		case s_VILE_ATK2:
			sfx = int32(sfx_vilatk)
		case s_SKEL_FIST2:
			sfx = int32(sfx_skeswg)
		case s_SKEL_FIST4:
			sfx = int32(sfx_skepch)
		case s_SKEL_MISS2:
			sfx = int32(sfx_skeatk)
		case s_FATT_ATK8:
			fallthrough
		case s_FATT_ATK5:
			fallthrough
		case s_FATT_ATK2:
			sfx = int32(sfx_firsht)
		case s_CPOS_ATK2:
			fallthrough
		case s_CPOS_ATK3:
			fallthrough
		case s_CPOS_ATK4:
			sfx = int32(sfx_shotgn)
		case s_TROO_ATK3:
			sfx = int32(sfx_claw)
		case s_SARG_ATK2:
			sfx = int32(sfx_sgtatk)
		case s_BOSS_ATK2:
			fallthrough
		case s_BOS2_ATK2:
			fallthrough
		case s_HEAD_ATK2:
			sfx = int32(sfx_firsht)
		case s_SKULL_ATK2:
			sfx = int32(sfx_sklatk)
		case s_SPID_ATK2:
			fallthrough
		case s_SPID_ATK3:
			sfx = int32(sfx_shotgn)
		case s_BSPI_ATK2:
			sfx = int32(sfx_plasma)
		case s_CYBER_ATK2:
			fallthrough
		case s_CYBER_ATK4:
			fallthrough
		case s_CYBER_ATK6:
			sfx = int32(sfx_rlaunc)
		case s_PAIN_ATK3:
			sfx = int32(sfx_sklatk)
		default:
			sfx = 0
			break
		}
		if sfx != 0 {
			s_StartSound(nil, sfx)
		}
	}
	if castframes == 12 {
		// go into attack frame
		castattacking = 1
		if castonmelee != 0 {
			caststate = &states[mobjinfo[castorder[castnum].Ftype1].Fmeleestate]
		} else {
			caststate = &states[mobjinfo[castorder[castnum].Ftype1].Fmissilestate]
		}
		castonmelee ^= 1
		if caststate == &states[0] {
			if castonmelee != 0 {
				caststate = &states[mobjinfo[castorder[castnum].Ftype1].Fmeleestate]
			} else {
				caststate = &states[mobjinfo[castorder[castnum].Ftype1].Fmissilestate]
			}
		}
	}
	if castattacking == 0 {
		goto _2
	}
	if !(castframes == 24 || caststate == &states[mobjinfo[castorder[castnum].Ftype1].Fseestate]) {
		goto _3
	}
	goto stopattack
stopattack:
	;
	castattacking = 0
	castframes = 0
	caststate = &states[mobjinfo[castorder[castnum].Ftype1].Fseestate]
_3:
	;
_2:
	;
	casttics = caststate.Ftics
	if casttics == -1 {
		casttics = 15
	}
}

//
// F_CastResponder
//

func f_CastResponder(ev *event_t) boolean {
	if ev.Ftype1 != Ev_keydown {
		return 0
	}
	if castdeath != 0 {
		return 1
	} // already in dying frames
	// go into death frame
	castdeath = 1
	caststate = &states[mobjinfo[castorder[castnum].Ftype1].Fdeathstate]
	casttics = caststate.Ftics
	castframes = 0
	castattacking = 0
	if mobjinfo[castorder[castnum].Ftype1].Fdeathsound != 0 {
		s_StartSound(nil, mobjinfo[castorder[castnum].Ftype1].Fdeathsound)
	}
	return 1
}

func f_CastPrint(text string) {
	var c, cx, w, width int32
	// find width
	width = 0
	for i := 0; i < len(text); i++ {
		c = xtoupper(int32(text[i])) - '!'
		if c < 0 || c > '_'-'!'+1 {
			width += 4
			continue
		}
		w = int32(hu_font[c].Fwidth)
		width += w
	}
	// draw it
	cx = 160 - width/2
	for i := 0; i < len(text); i++ {
		c = xtoupper(int32(text[i])) - '!'
		if c < 0 || c > '_'-'!'+1 {
			cx += 4
			continue
		}
		w = int32(hu_font[c].Fwidth)
		v_DrawPatch(cx, 180, hu_font[c])
		cx += w
	}
}

//
// F_CastDrawer
//

func f_CastDrawer() {
	var flip boolean
	var lump int32
	var patch *patch_t
	var sprdef *spritedef_t
	var sprframe *spriteframe_t
	// erase the entire screen to a background
	v_DrawPatch(0, 0, w_CacheLumpNameT("BOSSBACK"))
	f_CastPrint(castorder[castnum].Fname)
	// draw the current frame in the middle of the screen
	sprdef = &sprites[caststate.Fsprite]
	sprframe = &sprdef.Fspriteframes[caststate.Fframe&int32(FF_FRAMEMASK1)]
	lump = int32(sprframe.Flump[0])
	flip = uint32(sprframe.Fflip[0])
	patch = w_CacheLumpNumT(lump + firstspritelump)
	if flip != 0 {
		v_DrawPatchFlipped(160, 170, patch)
	} else {
		v_DrawPatch(160, 170, patch)
	}
}

// C documentation
//
//	//
//	// F_DrawPatchCol
//	//
func f_DrawPatchCol(x int32, patch *patch_t, col int32) {
	var dest int32
	var count int32
	column := patch.GetColumn(col)
	// step through the posts in a column
	for int32(column.Ftopdelta) != 0xff {
		source := column.Data()
		dest = x + int32(column.Ftopdelta)*SCREENWIDTH
		count = int32(column.Flength)
		for i := int32(0); i < count; i++ {
			I_VideoBuffer[dest] = source[i]
			dest += SCREENWIDTH
		}
		column = column.Next()
	}
}

// C documentation
//
//	//
//	// F_BunnyScroll
//	//
func f_BunnyScroll() {
	var scrolled, stage int32
	p1 := w_CacheLumpNameT("PFUB2")
	p2 := w_CacheLumpNameT("PFUB1")
	v_MarkRect(0, 0, SCREENWIDTH, SCREENHEIGHT)
	scrolled = 320 - (int32(finalecount)-int32(230))/2
	if scrolled > 320 {
		scrolled = 320
	}
	if scrolled < 0 {
		scrolled = 0
	}
	for x := int32(0); x < SCREENWIDTH; x++ {
		if x+scrolled < 320 {
			f_DrawPatchCol(x, p1, x+scrolled)
		} else {
			f_DrawPatchCol(x, p2, x+scrolled-int32(320))
		}
	}
	if finalecount < 1130 {
		return
	}
	if finalecount < 1180 {
		v_DrawPatch((SCREENWIDTH-13*8)/2, (SCREENHEIGHT-8*8)/2, w_CacheLumpNameT("END0"))
		laststage = 0
		return
	}
	stage = int32((finalecount - 1180) / 5)
	if stage > 6 {
		stage = 6
	}
	if stage > laststage {
		s_StartSound(nil, int32(sfx_pistol))
		laststage = stage
	}
	name := fmt.Sprintf("END%d", stage)
	v_DrawPatch((SCREENWIDTH-13*8)/2, (SCREENHEIGHT-8*8)/2, w_CacheLumpNameT(name))
}

var laststage int32

func f_ArtScreenDrawer() {
	var lumpname string
	if gameepisode == 3 {
		f_BunnyScroll()
	} else {
		switch gameepisode {
		case 1:
			if gamemode == retail {
				lumpname = "CREDIT"
			} else {
				lumpname = "HELP2"
			}
		case 2:
			lumpname = "VICTORY2"
		case 4:
			lumpname = "ENDPIC"
		default:
			return
		}
		v_DrawPatch(0, 0, w_CacheLumpNameT(lumpname))
	}
}

// C documentation
//
//	//
//	// F_Drawer
//	//
func f_Drawer() {
	switch finalestage {
	case F_STAGE_CAST:
		f_CastDrawer()
	case F_STAGE_TEXT:
		f_TextWrite()
	case F_STAGE_ARTSCREEN:
		f_ArtScreenDrawer()
		break
	}
}

//
//                       SCREEN WIPE PACKAGE
//

// C documentation
//
//	// when zero, stop the wipe
var wipe_running = 0

var wipe_scr_start []byte
var wipe_scr_end []byte
var wipe_scr []byte

// TODO: Stop doing width*2
func wipe_shittyColMajorXform(array []byte, width int32, height int32) {
	dest := make([]byte, width*2*height)
	for y := int32(0); y < height; y++ {
		for x := int32(0); x < width; x++ {
			dest[(x*height+y)*2] = array[(y*width+x)*2]
			dest[(x*height+y)*2+1] = array[(y*width+x)*2+1]
		}
	}
	copy(array, dest)
}

func wipe_initColorXForm(width int32, height int32, ticks int32) int32 {
	copy(wipe_scr, wipe_scr_start)
	return 0
}

func wipe_doColorXForm(width int32, height int32, ticks int32) int32 {
	var changed boolean
	var e, w int32
	var newval int32
	changed = 0
	w = 0
	e = 0
	for w != width*height {
		if wipe_scr[w] != wipe_scr_end[e] {
			if wipe_scr[w] > wipe_scr_end[e] {
				newval = int32(wipe_scr[w]) - ticks
				if newval < int32(wipe_scr_end[e]) {
					wipe_scr[w] = wipe_scr_end[e]
				} else {
					wipe_scr[w] = uint8(newval)
				}
				changed = 1
			} else {
				if int32(wipe_scr[w]) < int32(wipe_scr_end[e]) {
					newval = int32(wipe_scr[w]) + ticks
					if newval > int32(wipe_scr_end[e]) {
						wipe_scr[w] = wipe_scr_end[e]
					} else {
						wipe_scr[w] = uint8(newval)
					}
					changed = 1
				}
			}
		}
		w++
		e++
	}
	return boolint32(changed == 0)
}

func wipe_exitColorXForm(width int32, height int32, ticks int32) int32 {
	return 0
}

var y_screen []int32

func wipe_initMelt(width int32, height int32, ticks int32) (r1 int32) {
	var r int32
	// copy start screen to main screen
	copy(wipe_scr, wipe_scr_start)
	// makes this wipe faster (in theory)
	// to have stuff in column-major format
	wipe_shittyColMajorXform(wipe_scr_start, width/2, height)
	wipe_shittyColMajorXform(wipe_scr_end, width/2, height)
	// setup initial column positions
	// (y<0 => not ready to scroll yet)
	y_screen = make([]int32, width)
	y_screen[0] = -(m_Random() % 16)
	for i := int32(1); i < width; i++ {
		r = m_Random()%3 - 1
		y_screen[i] = y_screen[i-1] + r
		if y_screen[i] > 0 {
			y_screen[i] = 0
		} else {
			if y_screen[i] == -16 {
				y_screen[i] = -15
			}
		}
	}
	return 0
}

func wipe_doMelt(width int32, height int32, ticks int32) int32 {
	var d, s int32
	var done boolean
	var dy, idx, v1, v3 int32
	done = 1
	width /= 2
	for {
		v1 = ticks
		ticks--
		if v1 == 0 {
			break
		}
		for i := int32(0); i < width; i++ {
			if y_screen[i] < 0 {
				y_screen[i]++
				done = 0
			} else {
				if y_screen[i] < height {
					if y_screen[i] < 16 {
						v3 = y_screen[i] + 1
					} else {
						v3 = 8
					}
					dy = v3
					if y_screen[i]+dy >= height {
						dy = height - y_screen[i]
					}
					s = (i*height + y_screen[i]) * 2
					d = (y_screen[i]*width + i) * 2
					idx = 0
					for j := dy; j > 0; j-- {
						wipe_scr[d+idx*2] = wipe_scr_end[s]
						wipe_scr[d+idx*2+1] = wipe_scr_end[s+1]
						s += 2
						idx += width
					}
					y_screen[i] += dy
					s = (i * height) * 2
					d = (y_screen[i]*width + i) * 2
					idx = 0
					for j := height - y_screen[i]; j > 0; j-- {
						wipe_scr[d+idx*2] = wipe_scr_start[s]
						wipe_scr[d+idx*2+1] = wipe_scr_start[s+1]
						s += 2
						idx += width
					}
					done = 0
				}
			}
		}
	}
	return int32(done)
}

func wipe_exitMelt(width int32, height int32, ticks int32) int32 {
	y_screen = nil
	wipe_scr_start = nil
	wipe_scr_end = nil
	return 0
}

func wipe_StartScreen(x int32, y int32, width int32, height int32) int32 {
	wipe_scr_start = make([]byte, SCREENWIDTH*SCREENHEIGHT)
	i_ReadScreen(wipe_scr_start)
	return 0
}

func wipe_EndScreen(x int32, y int32, width int32, height int32) int32 {
	wipe_scr_end = make([]byte, SCREENWIDTH*SCREENHEIGHT)
	i_ReadScreen(wipe_scr_end)
	v_DrawBlock(x, y, width, height, wipe_scr_start) // restore start scr.
	return 0
}

func wipe_ScreenWipe(wipeno int32, x int32, y int32, width int32, height int32, ticks int32) int32 {
	var rc int32
	// initial stuff
	if wipe_running == 0 {
		wipe_running = 1
		wipe_scr = I_VideoBuffer
		wipes[wipeno*3](width, height, ticks)
	}
	// do a piece of wipe-in
	v_MarkRect(0, 0, width, height)
	rc = wipes[wipeno*3+1](width, height, ticks)
	//  v_DrawBlock(x, y, 0, width, height, wipe_scr); // DEBUG
	// final stuff
	if rc != 0 {
		wipe_running = 0
		wipes[wipeno*3+2](width, height, ticks)
	}
	return boolint32(wipe_running == 0)
}

var wipes = [6]func(int32, int32, int32) int32{
	wipe_initColorXForm,
	wipe_doColorXForm,
	wipe_exitColorXForm,
	wipe_initMelt,
	wipe_doMelt,
	wipe_exitMelt,
}

const ANG451 = 536870912
const ANGLETOSKYSHIFT = 22
const BODYQUESIZE = 32
const DEH_DEFAULT_BFG_CELLS_PER_SHOT = 40
const DEH_DEFAULT_BLUE_ARMOR_CLASS = 2
const DEH_DEFAULT_GOD_MODE_HEALTH = 100
const DEH_DEFAULT_GREEN_ARMOR_CLASS = 1
const DEH_DEFAULT_IDFA_ARMOR = 200
const DEH_DEFAULT_IDFA_ARMOR_CLASS = 2
const DEH_DEFAULT_IDKFA_ARMOR = 200
const DEH_DEFAULT_IDKFA_ARMOR_CLASS = 2
const DEH_DEFAULT_INITIAL_BULLETS = 50
const DEH_DEFAULT_INITIAL_HEALTH = 100
const DEH_DEFAULT_MAX_ARMOR = 200
const DEH_DEFAULT_MAX_HEALTH = 200
const DEH_DEFAULT_MAX_SOULSPHERE = 200
const DEH_DEFAULT_MEGASPHERE_HEALTH = 200
const DEH_DEFAULT_SOULSPHERE_HEALTH = 100
const DEH_DEFAULT_SPECIES_INFIGHTING = 0
const DEMOMARKER = 128
const MAX_JOY_BUTTONS = 20
const NUMKEYS = 256
const SAVEGAMESIZE = 180224
const SLOWTURNTICS = 6
const TURBOTHRESHOLD = 50
const VERSIONSIZE = 16

func init() {
	precache = 1
}

func init() {
	forwardmove = [2]fixed_t{
		0: 0x19,
		1: 0x32,
	}
}

func init() {
	sidemove = [2]fixed_t{
		0: 0x18,
		1: 0x28,
	}
}

func init() {
	angleturn = [3]fixed_t{
		0: 640,
		1: 1280,
		2: 320,
	}
}

// + slow turn

var weapon_keys = [8]*int32{
	0: &key_weapon1,
	1: &key_weapon2,
	2: &key_weapon3,
	3: &key_weapon4,
	4: &key_weapon5,
	5: &key_weapon6,
	6: &key_weapon7,
	7: &key_weapon8,
}

// Set to -1 or +1 to switch to the previous or next weapon.

var next_weapon weapontype_t = 0

// Used for prev/next weapon keys.

var weapon_order_table = [9]struct {
	Fweapon     weapontype_t
	Fweapon_num weapontype_t
}{
	0: {},
	1: {
		Fweapon: wp_chainsaw,
	},
	2: {
		Fweapon:     wp_pistol,
		Fweapon_num: wp_pistol,
	},
	3: {
		Fweapon:     wp_shotgun,
		Fweapon_num: wp_shotgun,
	},
	4: {
		Fweapon:     wp_supershotgun,
		Fweapon_num: wp_shotgun,
	},
	5: {
		Fweapon:     wp_chaingun,
		Fweapon_num: wp_chaingun,
	},
	6: {
		Fweapon:     wp_missile,
		Fweapon_num: wp_missile,
	},
	7: {
		Fweapon:     wp_plasma,
		Fweapon_num: wp_plasma,
	},
	8: {
		Fweapon:     wp_bfg,
		Fweapon_num: wp_bfg,
	},
}

var gamekeydown [256]bool
var turnheld int32 // for accelerative turning
var mousearray [9]bool

func mouseButton(button int32) bool {
	return mousearray[button+1] // allow for [-1]
}

func setMouseButton(button int32, state bool) {
	mousearray[button+1] = state // allow for [-1]
}

var dclicktime int32
var dclickstate bool
var dclicks int32
var dclicktime2 int32
var dclickstate2 boolean
var dclicks2 int32

// C documentation
//
//	// joystick values are repeated
var joyxmove int32
var joyymove int32
var joystrafemove int32
var joyarray [21]bool

func joyButton(button int32) bool {
	return joyarray[button+1] // allow for [-1]
}
func setJoyButton(button int32, state bool) {
	joyarray[button+1] = state // allow for [-1]
}

var savegameslot int32
var savedescription string

func init() {
	vanilla_savegame_limit = 1
}

func init() {
	vanilla_demo_limit = 1
}

func weaponSelectable(weapon weapontype_t) boolean {
	var v1 gamemission_t
	var v3 bool
	// Can't select the super shotgun in Doom 1.
	if v3 = weapon == wp_supershotgun; v3 {
		if gamemission == pack_chex {
			v1 = doom
		} else {
			if gamemission == pack_hacx {
				v1 = doom2
			} else {
				v1 = gamemission
			}
		}
	}
	if v3 && v1 == doom {
		return 0
	}
	// These weapons aren't available in shareware.
	if (weapon == wp_plasma || weapon == wp_bfg) && gamemission == doom && gamemode == shareware {
		return 0
	}
	// Can't select a weapon if we don't own it.
	if players[consoleplayer].Fweaponowned[weapon] == 0 {
		return 0
	}
	// Can't select the fist if we have the chainsaw, unless
	// we also have the berserk pack.
	if weapon == wp_fist && players[consoleplayer].Fweaponowned[wp_chainsaw] != 0 && players[consoleplayer].Fpowers[pw_strength] == 0 {
		return 0
	}
	return 1
}

func g_NextWeapon(direction int32) weapontype_t {
	var i, start_i int32
	var weapon weapontype_t
	// Find index in the table.
	if players[consoleplayer].Fpendingweapon == wp_nochange {
		weapon = players[consoleplayer].Freadyweapon
	} else {
		weapon = players[consoleplayer].Fpendingweapon
	}
	for i = 0; i < int32(len(weapon_order_table)); i++ {
		if weapon_order_table[i].Fweapon == weapon {
			break
		}
	}
	// Switch weapon. Don't loop forever.
	start_i = i
	for cond := true; cond; cond = i != start_i && weaponSelectable(weapon_order_table[i].Fweapon) == 0 {
		i += direction
		i = int32((int(i) + len(weapon_order_table)) % len(weapon_order_table))
	}
	return weapon_order_table[i].Fweapon_num
}

// C documentation
//
//	//
//	// G_BuildTiccmd
//	// Builds a ticcmd from all of the available inputs
//	// or reads it from the demo buffer.
//	// If recording a demo, write it out
//	//
func g_BuildTiccmd(cmd *ticcmd_t, maketic int32) {
	var bstrafe, strafe boolean
	var desired_angleturn int16
	var forward, side, speed, tspeed, v1, v16 int32
	*cmd = ticcmd_t{}
	cmd.Fconsistancy = consistancy[consoleplayer][maketic%BACKUPTICS]
	strafe = booluint32(gamekeydown[key_strafe] || mouseButton(mousebstrafe) || joyButton(joybstrafe))
	// fraggle: support the old "joyb_speed = 31" hack which
	// allowed an autorun effect
	speed = boolint32(key_speed >= NUMKEYS || joybspeed >= MAX_JOY_BUTTONS || gamekeydown[key_speed] || joyButton(joybspeed))
	v1 = 0
	side = v1
	forward = v1
	// use two stage accelerative turning
	// on the keyboard and joystick
	if joyxmove < 0 || joyxmove > 0 || gamekeydown[key_right] || gamekeydown[key_left] {
		turnheld += ticdup
	} else {
		turnheld = 0
	}
	if turnheld < SLOWTURNTICS {
		tspeed = 2
	} else {
		tspeed = speed
	}
	// let movement keys cancel each other out
	if strafe != 0 {
		if gamekeydown[key_right] {
			// fprintf(stderr, "strafe right\n");
			side += sidemove[speed]
		}
		if gamekeydown[key_left] {
			//	fprintf(stderr, "strafe left\n");
			side -= sidemove[speed]
		}
		if joyxmove > 0 {
			side += sidemove[speed]
		}
		if joyxmove < 0 {
			side -= sidemove[speed]
		}
	} else {
		if gamekeydown[key_right] {
			cmd.Fangleturn -= int16(angleturn[tspeed])
		}
		if gamekeydown[key_left] {
			cmd.Fangleturn += int16(angleturn[tspeed])
		}
		if joyxmove > 0 {
			cmd.Fangleturn -= int16(angleturn[tspeed])
		}
		if joyxmove < 0 {
			cmd.Fangleturn += int16(angleturn[tspeed])
		}
	}
	if gamekeydown[key_up] {
		// fprintf(stderr, "up\n");
		forward += forwardmove[speed]
	}
	if gamekeydown[key_down] {
		// fprintf(stderr, "down\n");
		forward -= forwardmove[speed]
	}
	if joyymove < 0 {
		forward += forwardmove[speed]
	}
	if joyymove > 0 {
		forward -= forwardmove[speed]
	}
	if gamekeydown[key_strafeleft] || joyButton(joybstrafeleft) || mouseButton(mousebstrafeleft) || joystrafemove < 0 {
		side -= sidemove[speed]
	}
	if gamekeydown[key_straferight] || joyButton(joybstraferight) || mouseButton(mousebstraferight) || joystrafemove > 0 {
		side += sidemove[speed]
	}
	// buttons
	cmd.Fchatchar = uint8(hu_dequeueChatChar())
	if gamekeydown[key_fire] || mouseButton(mousebfire) || joyButton(joybfire) {
		cmd.Fbuttons |= bt_ATTACK
	}
	if gamekeydown[key_use] || joyButton(joybuse) || mouseButton(mousebuse) {
		cmd.Fbuttons |= bt_USE
		// clear double clicks if hit use button
		dclicks = 0
	}
	// If the previous or next weapon button is pressed, the
	// next_weapon variable is set to change weapons when
	// we generate a ticcmd.  Choose a new weapon.
	if gamestate == gs_LEVEL && next_weapon != 0 {
		i := g_NextWeapon(int32(next_weapon))
		cmd.Fbuttons |= bt_CHANGE
		cmd.Fbuttons |= uint8(i << bt_WEAPONSHIFT)
	} else {
		// Check weapon keys.
		for i := 0; i < len(weapon_keys); i++ {
			if gamekeydown[*weapon_keys[i]] {
				cmd.Fbuttons |= bt_CHANGE
				cmd.Fbuttons |= uint8(i << bt_WEAPONSHIFT)
				break
			}
		}
	}
	next_weapon = 0
	// mouse
	if mouseButton(mousebforward) {
		forward += forwardmove[speed]
	}
	if mouseButton(mousebbackward) {
		forward -= forwardmove[speed]
	}
	if dclick_use != 0 {
		// forward double click
		if mouseButton(mousebforward) != dclickstate && dclicktime > 1 {
			dclickstate = mouseButton(mousebforward)
			if dclickstate {
				dclicks++
			}
			if dclicks == 2 {
				cmd.Fbuttons |= bt_USE
				dclicks = 0
			} else {
				dclicktime = 0
			}
		} else {
			dclicktime += ticdup
			if dclicktime > 20 {
				dclicks = 0
				dclickstate = false
			}
		}
		// strafe double click
		bstrafe = booluint32(mouseButton(mousebstrafe) || joyButton(joybstrafe))
		if bstrafe != dclickstate2 && dclicktime2 > 1 {
			dclickstate2 = bstrafe
			if dclickstate2 != 0 {
				dclicks2++
			}
			if dclicks2 == 2 {
				cmd.Fbuttons |= bt_USE
				dclicks2 = 0
			} else {
				dclicktime2 = 0
			}
		} else {
			dclicktime2 += ticdup
			if dclicktime2 > 20 {
				dclicks2 = 0
				dclickstate2 = 0
			}
		}
	}
	forward += mousey
	if strafe != 0 {
		side += mousex * 2
	} else {
		cmd.Fangleturn -= int16(mousex * 8)
	}
	if mousex == 0 {
		// No movement in the previous frame
		testcontrols_mousespeed = 0
	}
	v16 = 0
	mousey = v16
	mousex = v16
	if forward > forwardmove[1] {
		forward = forwardmove[1]
	} else {
		if forward < -forwardmove[1] {
			forward = -forwardmove[1]
		}
	}
	if side > forwardmove[1] {
		side = forwardmove[1]
	} else {
		if side < -forwardmove[1] {
			side = -forwardmove[1]
		}
	}
	cmd.Fforwardmove += int8(forward)
	cmd.Fsidemove += int8(side)
	// special buttons
	if sendpause != 0 {
		sendpause = 0
		cmd.Fbuttons = uint8(bt_SPECIAL | bts_PAUSE)
	}
	if sendsave != 0 {
		sendsave = 0
		cmd.Fbuttons = uint8(bt_SPECIAL | bts_SAVEGAME | savegameslot<<bts_SAVESHIFT)
	}
	// low-res turning
	if lowres_turn != 0 {
		desired_angleturn = int16(int32(cmd.Fangleturn) + int32(carry))
		// round angleturn to the nearest 256 unit boundary
		// for recording demos with single byte values for turn
		cmd.Fangleturn = int16((int32(desired_angleturn) + 128) & 0xff00)
		// Carry forward the error from the reduced resolution to the
		// next tic, so that successive small movements can accumulate.
		carry = int16(int32(desired_angleturn) - int32(cmd.Fangleturn))
	}
}

var carry int16

// C documentation
//
//	//
//	// G_DoLoadLevel
//	//
func g_DoLoadLevel() {
	var v2, v3, v4 int32
	var skytexturename string
	var v5, v6 boolean
	// Set the sky map.
	// First thing, we have a dummy sky texture name,
	//  a flat. The data is in the WAD only because
	//  we look for an actual index, instead of simply
	//  setting one.
	skyflatnum = r_FlatNumForName("F_SKY1")
	// The "Sky never changes in Doom II" bug was fixed in
	// the id Anthology version of doom2.exe for Final Doom.
	if gamemode == commercial && (gameversion == exe_final2 || gameversion == exe_chex) {
		if gamemap < 12 {
			skytexturename = "SKY1"
		} else {
			if gamemap < 21 {
				skytexturename = "SKY2"
			} else {
				skytexturename = "SKY3"
			}
		}
		skytexturename = skytexturename
		skytexture = r_TextureNumForName(skytexturename)
	}
	if wipegamestate == gs_LEVEL {
		wipegamestate = -1
	} // force a wipe
	gamestate = gs_LEVEL
	for i := 0; i < MAXPLAYERS; i++ {
		turbodetected[i] = 0
		if playeringame[i] != 0 && players[i].Fplayerstate == Pst_DEAD {
			players[i].Fplayerstate = Pst_REBORN
		}
		clear(players[i].Ffrags[:])
	}
	p_SetupLevel(gameepisode, gamemap, 0, gameskill)
	displayplayer = consoleplayer // view the guy you are playing
	gameaction = ga_nothing
	// clear cmd building stuff
	clear(gamekeydown[:])
	v3 = 0
	joystrafemove = v3
	v2 = v3
	joyymove = v2
	joyxmove = v2
	v4 = 0
	mousey = v4
	mousex = v4
	v6 = 0
	paused = v6
	v5 = v6
	sendsave = v5
	sendpause = v5
	clear(mousearray[:])
	clear(joyarray[:])
	if testcontrols != 0 {
		players[consoleplayer].Fmessage = "Press escape to quit."
	}
}

func setJoyButtons(buttons_mask uint32) {
	var button_on int32
	for i := int32(0); i < MAX_JOY_BUTTONS; i++ {
		button_on = boolint32(buttons_mask&uint32(1<<i) != 0)
		// Detect button press:
		if joyButton(i) && button_on != 0 {
			// Weapon cycling:
			if i == joybprevweapon {
				next_weapon = -1
			} else {
				if i == joybnextweapon {
					next_weapon = 1
				}
			}
		}
		setJoyButton(i, button_on != 0)
	}
}

func setMouseButtons(buttons_mask uint32) {
	var button_on uint32
	for i := int32(0); i < MAX_MOUSE_BUTTONS; i++ {
		button_on = booluint32(buttons_mask&uint32(1<<i) != 0)
		// Detect button press:
		if !mouseButton(i) && button_on != 0 {
			if i == mousebprevweapon {
				next_weapon = -1
			} else {
				if i == mousebnextweapon {
					next_weapon = 1
				}
			}
		}
		setMouseButton(i, button_on != 0)
	}
}

// C documentation
//
//	//
//	// G_Responder
//	// Get info needed to make ticcmd_ts for the players.
//	//
func g_Responder(ev *event_t) boolean {
	// allow spy mode changes even during the demo
	if gamestate == gs_LEVEL && ev.Ftype1 == Ev_keydown && ev.Fdata1 == key_spy && (singledemo != 0 || deathmatch == 0) {
		// spy mode
		for cond := true; cond; cond = playeringame[displayplayer] == 0 && displayplayer != consoleplayer {
			displayplayer++
			if displayplayer == MAXPLAYERS {
				displayplayer = 0
			}
		}
		return 1
	}
	// any other key pops up menu if in demos
	if gameaction == ga_nothing && singledemo == 0 && (demoplayback != 0 || gamestate == gs_DEMOSCREEN) {
		if ev.Ftype1 == Ev_keydown || ev.Ftype1 == Ev_mouse && ev.Fdata1 != 0 || ev.Ftype1 == Ev_joystick && ev.Fdata1 != 0 {
			m_StartControlPanel()
			return 1
		}
		return 0
	}
	if gamestate == gs_LEVEL {
		if hu_Responder(ev) != 0 {
			return 1
		} // chat ate the event
		if st_Responder(ev) != 0 {
			return 1
		} // status window ate it
		if am_Responder(ev) != 0 {
			return 1
		} // automap ate it
	}
	if gamestate == gs_FINALE {
		if f_Responder(ev) != 0 {
			return 1
		} // finale ate the event
	}
	if testcontrols != 0 && ev.Ftype1 == Ev_mouse {
		// If we are invoked by setup to test the controls, save the
		// mouse speed so that we can display it on-screen.
		// Perform a low pass filter on this so that the thermometer
		// appears to move smoothly.
		testcontrols_mousespeed = xabs(ev.Fdata2)
	}
	// If the next/previous weapon keys are pressed, set the next_weapon
	// variable to change weapons when the next ticcmd is generated.
	if ev.Ftype1 == Ev_keydown && ev.Fdata1 == key_prevweapon {
		next_weapon = -1
	} else {
		if ev.Ftype1 == Ev_keydown && ev.Fdata1 == key_nextweapon {
			next_weapon = 1
		}
	}
	switch ev.Ftype1 {
	case Ev_keydown:
		if ev.Fdata1 == key_pause {
			sendpause = 1
		} else {
			if ev.Fdata1 < NUMKEYS {
				gamekeydown[ev.Fdata1] = true
			}
		}
		return 1 // eat key down events
	case Ev_keyup:
		if ev.Fdata1 < NUMKEYS {
			gamekeydown[ev.Fdata1] = false
		}
		return 0 // always let key up events filter down
	case Ev_mouse:
		setMouseButtons(uint32(ev.Fdata1))
		mousex = ev.Fdata2 * (mouseSensitivity + 5) / 10
		mousey = ev.Fdata3 * (mouseSensitivity + 5) / 10
		return 1 // eat events
	case Ev_joystick:
		setJoyButtons(uint32(ev.Fdata1))
		joyxmove = ev.Fdata2
		joyymove = ev.Fdata3
		joystrafemove = ev.Fdata4
		return 1 // eat events
	default:
		break
	}
	return 0
}

// C documentation
//
//	//
//	// G_Ticker
//	// Make ticcmd_ts for the players.
//	//
func g_Ticker() {
	var buf int32
	var cmd *ticcmd_t
	// do player reborns if needed
	for i := int32(0); i < MAXPLAYERS; i++ {
		if playeringame[i] != 0 && players[i].Fplayerstate == Pst_REBORN {
			g_DoReborn(i)
		}
	}
	// do things to change the game state
	for gameaction != ga_nothing {
		switch gameaction {
		case ga_loadlevel:
			g_DoLoadLevel()
		case ga_newgame:
			g_DoNewGame()
		case ga_loadgame:
			g_DoLoadGame()
		case ga_savegame:
			g_DoSaveGame()
		case ga_playdemo:
			g_DoPlayDemo()
		case ga_completed:
			g_DoCompleted()
		case ga_victory:
			f_StartFinale()
		case ga_worlddone:
			g_DoWorldDone()
		case ga_nothing:
			break
		}
	}
	// get commands, check consistancy,
	// and build new consistancy check
	buf = gametic / ticdup % BACKUPTICS
	for i := int32(0); i < MAXPLAYERS; i++ {
		if playeringame[i] != 0 {
			cmd = &players[i].Fcmd
			*cmd = netcmds[i]
			if demoplayback != 0 {
				g_ReadDemoTiccmd(cmd)
			}
			if demorecording != 0 {
				g_WriteDemoTiccmd(cmd)
			}
			// check for turbo cheats
			// check ~ 4 seconds whether to display the turbo message.
			// store if the turbo threshold was exceeded in any tics
			// over the past 4 seconds.  offset the checking period
			// for each player so messages are not displayed at the
			// same time.
			if int32(cmd.Fforwardmove) > TURBOTHRESHOLD {
				turbodetected[i] = 1
			}
			if gametic&int32(31) == 0 && gametic>>5%MAXPLAYERS == i && turbodetected[i] != 0 {
				players[consoleplayer].Fmessage = fmt.Sprintf("%s is turbo!", player_names[i])
				turbodetected[i] = 0
			}
			if netgame != 0 && netdemo == 0 && gametic%ticdup == 0 {
				if gametic > BACKUPTICS && int32(consistancy[i][buf]) != int32(cmd.Fconsistancy) {
					i_Error("consistency failure (%d should be %d)", int32(cmd.Fconsistancy), int32(consistancy[i][buf]))
				}
				if players[i].Fmo != nil {
					consistancy[i][buf] = uint8(players[i].Fmo.Fx)
				} else {
					consistancy[i][buf] = uint8(rndindex)
				}
			}
		}
	}
	// check for special buttons
	for i := int32(0); i < MAXPLAYERS; i++ {
		if playeringame[i] != 0 {
			if int32(players[i].Fcmd.Fbuttons)&bt_SPECIAL != 0 {
				switch int32(players[i].Fcmd.Fbuttons) & bt_SPECIALMASK {
				case bts_PAUSE:
					paused = boolean(paused ^ 1)
					if paused != 0 {
						s_PauseSound()
					} else {
						s_ResumeSound()
					}
				case bts_SAVEGAME:
					if len(savedescription) == 0 {
						savedescription = "NET GAME"
					}
					savegameslot = int32(players[i].Fcmd.Fbuttons) & bts_SAVEMASK >> bts_SAVESHIFT
					gameaction = ga_savegame
					break
				}
			}
		}
	}
	// Have we just finished displaying an intermission screen?
	if oldgamestate == gs_INTERMISSION && gamestate != gs_INTERMISSION {
		wi_End()
	}
	oldgamestate = gamestate
	// do main actions
	switch gamestate {
	case gs_LEVEL:
		p_Ticker()
		st_Ticker()
		am_Ticker()
		hu_Ticker()
	case gs_INTERMISSION:
		wi_Ticker()
	case gs_FINALE:
		f_Ticker()
	case gs_DEMOSCREEN:
		d_PageTicker()
		break
	}
}

// C documentation
//
//	//
//	// G_PlayerFinishLevel
//	// Can when a player completes a level.
//	//
func g_PlayerFinishLevel(player int32) {
	p := &players[player]
	clear(p.Fpowers[:])
	clear(p.Fcards[:])
	p.Fmo.Fflags &= ^mf_SHADOW // cancel invisibility
	p.Fextralight = 0          // cancel gun flashes
	p.Ffixedcolormap = 0       // cancel ir gogles
	p.Fdamagecount = 0         // no palette changes
	p.Fbonuscount = 0
}

// C documentation
//
//	//
//	// G_PlayerReborn
//	// Called after a player dies
//	// almost everything is cleared and initialized
//	//
func g_PlayerReborn(player int32) {
	var itemcount, killcount, secretcount int32
	var frags [4]int32

	p := &players[player]
	frags = p.Ffrags
	killcount = p.Fkillcount
	itemcount = p.Fitemcount
	secretcount = p.Fsecretcount
	*p = player_t{Ffrags: frags} // clear the player structure
	p.Fkillcount = killcount
	p.Fitemcount = itemcount
	p.Fsecretcount = secretcount
	p.Fattackdown = 1
	p.Fusedown = 1 // don't do anything immediately
	p.Fplayerstate = Pst_LIVE
	p.Fhealth = DEH_DEFAULT_INITIAL_HEALTH // Use dehacked value
	p.Fpendingweapon = wp_pistol
	p.Freadyweapon = wp_pistol

	p.Fweaponowned[wp_fist] = 1
	p.Fweaponowned[wp_pistol] = 1
	p.Fammo[am_clip] = DEH_DEFAULT_INITIAL_BULLETS
	for i := 0; i < NUMAMMO; i++ {
		p.Fmaxammo[i] = maxammo[i]
	}
}

func g_CheckSpot(playernum int32, mthing *mapthing_t) boolean {
	var an int32
	var mo *mobj_t
	var ss *subsector_t
	var x, xa, y, ya, v2 fixed_t
	if players[playernum].Fmo == nil {
		// first spawn of level, before corpses
		for i := int32(0); i < playernum; i++ {
			if players[i].Fmo.Fx == int32(mthing.Fx)<<FRACBITS && players[i].Fmo.Fy == int32(mthing.Fy)<<FRACBITS {
				return 0
			}
		}
		return 1
	}
	x = int32(mthing.Fx) << FRACBITS
	y = int32(mthing.Fy) << FRACBITS
	if p_CheckPosition(players[playernum].Fmo, x, y) == 0 {
		return 0
	}
	// flush an old corpse if needed
	if bodyqueslot >= BODYQUESIZE {
		p_RemoveMobj(bodyque[bodyqueslot%BODYQUESIZE])
	}
	bodyque[bodyqueslot%BODYQUESIZE] = players[playernum].Fmo
	bodyqueslot++
	// spawn a teleport fog
	ss = r_PointInSubsector(x, y)
	// The code in the released source looks like this:
	//
	//    an = ( ANG45 * (((unsigned int) mthing->angle)/45) )
	//         >> ANGLETOFINESHIFT;
	//    mo = p_SpawnMobj (x+20*finecosine[an], y+20*finesine[an]
	//                     , ss->sector->floorheight
	//                     , mt_TFOG);
	//
	// But 'an' can be a signed value in the DOS version. This means that
	// we get a negative index and the lookups into finecosine/finesine
	// end up dereferencing values in finetangent[].
	// A player spawning on a deathmatch start facing directly west spawns
	// "silently" with no spawn fog. Emulate this.
	//
	// This code is imported from PrBoom+.
	// This calculation overflows in Vanilla Doom, but here we deliberately
	// avoid integer overflow as it is undefined behavior, so the value of
	// 'an' will always be positive.
	an = ANG451 >> ANGLETOFINESHIFT * (int32(mthing.Fangle) / 45)
	switch an {
	case 4096: // -4096:
		xa = finetangent[int32(2048)] // finecosine[-4096]
		ya = finetangent[0]           // finesine[-4096]
	case 5120: // -3072:
		xa = finetangent[int32(3072)] // finecosine[-3072]
		ya = finetangent[int32(1024)] // finesine[-3072]
	case 6144: // -2048:
		xa = finesine[0]              // finecosine[-2048]
		ya = finetangent[int32(2048)] // finesine[-2048]
	case 7168: // -1024:
		xa = finesine[int32(1024)]    // finecosine[-1024]
		ya = finetangent[int32(3072)] // finesine[-1024]
	case 0:
		fallthrough
	case 1024:
		fallthrough
	case 2048:
		fallthrough
	case 3072:
		xa = finecosine[an]
		ya = finesine[an]
	default:
		i_Error("g_CheckSpot: unexpected angle %d\n", an)
		v2 = 0
		ya = v2
		xa = v2
		break
	}
	mo = p_SpawnMobj(x+int32(20)*xa, y+int32(20)*ya, ss.Fsector.Ffloorheight, mt_TFOG)
	if players[consoleplayer].Fviewz != 1 {
		s_StartSound(&mo.degenmobj_t, int32(sfx_telept))
	} // don't start sound on first frame
	return 1
}

// C documentation
//
//	//
//	// G_DeathMatchSpawnPlayer
//	// Spawns a player at one of the random death match spots
//	// called at level load and each death
//	//
func g_DeathMatchSpawnPlayer(playernum int32) {
	var i, selections int32
	selections = int32(deathmatch_pos)
	if selections < 4 {
		i_Error("Only %d deathmatch spots, 4 required", selections)
	}
	for range 20 {
		i = p_Random() % selections
		if g_CheckSpot(playernum, &deathmatchstarts[i]) != 0 {
			deathmatchstarts[i].Ftype1 = int16(playernum + 1)
			p_SpawnPlayer(&deathmatchstarts[i])
			return
		}
	}
	// no good spot, so the player will probably get stuck
	p_SpawnPlayer(&playerstarts[playernum])
}

// C documentation
//
//	//
//	// G_DoReborn
//	//
func g_DoReborn(playernum int32) {
	if netgame == 0 {
		// reload the level from scratch
		gameaction = ga_loadlevel
	} else {
		// respawn at the start
		// first dissasociate the corpse
		players[playernum].Fmo.Fplayer = nil
		// spawn at random spot if in death match
		if deathmatch != 0 {
			g_DeathMatchSpawnPlayer(playernum)
			return
		}
		if g_CheckSpot(playernum, &playerstarts[playernum]) != 0 {
			p_SpawnPlayer(&playerstarts[playernum])
			return
		}
		// try to spawn at one of the other players spots
		for i := range MAXPLAYERS {
			if g_CheckSpot(playernum, &playerstarts[i]) != 0 {
				playerstarts[i].Ftype1 = int16(playernum + 1) // fake as other player
				p_SpawnPlayer(&playerstarts[i])
				playerstarts[i].Ftype1 = int16(i + 1) // restore
				return
			}
			// he's going to be inside something.  Too bad.
		}
		p_SpawnPlayer(&playerstarts[playernum])
	}
}

func init() {
	pars = [4][10]int32{
		0: {},
		1: {
			1: 30,
			2: 75,
			3: 120,
			4: 90,
			5: 165,
			6: 180,
			7: 180,
			8: 30,
			9: 165,
		},
		2: {
			1: 90,
			2: 90,
			3: 90,
			4: 120,
			5: 90,
			6: 360,
			7: 240,
			8: 30,
			9: 170,
		},
		3: {
			1: 90,
			2: 45,
			3: 90,
			4: 150,
			5: 90,
			6: 90,
			7: 165,
			8: 30,
			9: 135,
		},
	}
}

func init() {
	cpars = [32]int32{
		0:  30,
		1:  90,
		2:  120,
		3:  120,
		4:  90,
		5:  150,
		6:  120,
		7:  120,
		8:  270,
		9:  90,
		10: 210,
		11: 150,
		12: 150,
		13: 150,
		14: 210,
		15: 150,
		16: 420,
		17: 150,
		18: 210,
		19: 150,
		20: 240,
		21: 150,
		22: 180,
		23: 150,
		24: 150,
		25: 300,
		26: 330,
		27: 420,
		28: 300,
		29: 180,
		30: 120,
		31: 30,
	}
}

func g_ExitLevel() {
	secretexit = 0
	gameaction = ga_completed
}

// C documentation
//
//	// Here's for the german edition.
func g_SecretExitLevel() {
	// IF NO WOLF3D LEVELS, NO SECRET EXIT!
	if gamemode == commercial && w_CheckNumForName("map31") < 0 {
		secretexit = 0
	} else {
		secretexit = 1
	}
	gameaction = ga_completed
}

func g_DoCompleted() {
	gameaction = ga_nothing
	for i := range int32(MAXPLAYERS) {
		if playeringame[i] != 0 {
			g_PlayerFinishLevel(i)
		}
	} // take away cards and stuff
	if automapactive != 0 {
		am_Stop()
	}
	if gamemode != commercial {
		// Chex Quest ends after 5 levels, rather than 8.
		if gameversion == exe_chex {
			if gamemap == 5 {
				gameaction = ga_victory
				return
			}
		} else {
			switch gamemap {
			case 8:
				gameaction = ga_victory
				return
			case 9:
				for i := range MAXPLAYERS {
					players[i].Fdidsecret = 1
				}
			}
		}
	}
	//#if 0  Hmmm - why?
	if gamemap == 8 && gamemode != commercial {
		// victory
		gameaction = ga_victory
		return
	}
	if gamemap == 9 && gamemode != commercial {
		// exit secret level
		for i := range MAXPLAYERS {
			players[i].Fdidsecret = 1
		}
	}
	//#endif
	wminfo.Fdidsecret = players[consoleplayer].Fdidsecret
	wminfo.Fepsd = gameepisode - 1
	wminfo.Flast = gamemap - 1
	// wminfo.next is 0 biased, unlike gamemap
	if gamemode == commercial {
		if secretexit != 0 {
			switch gamemap {
			case 15:
				wminfo.Fnext = 30
			case 31:
				wminfo.Fnext = 31
				break
			}
		} else {
			switch gamemap {
			case 31:
				fallthrough
			case 32:
				wminfo.Fnext = 15
			default:
				wminfo.Fnext = gamemap
			}
		}
	} else {
		if secretexit != 0 {
			wminfo.Fnext = 8
		} else {
			if gamemap == 9 {
				// returning from secret level
				switch gameepisode {
				case 1:
					wminfo.Fnext = 3
				case 2:
					wminfo.Fnext = 5
				case 3:
					wminfo.Fnext = 6
				case 4:
					wminfo.Fnext = 2
					break
				}
			} else {
				wminfo.Fnext = gamemap
			}
		} // go to next level
	}
	wminfo.Fmaxkills = totalkills
	wminfo.Fmaxitems = totalitems
	wminfo.Fmaxsecret = totalsecret
	wminfo.Fmaxfrags = 0
	// Set par time. Doom episode 4 doesn't have a par time, so this
	// overflows into the cpars array. It's necessary to emulate this
	// for statcheck regression testing.
	if gamemode == commercial {
		wminfo.Fpartime = TICRATE * cpars[gamemap-1]
	} else {
		if gameepisode < 4 {
			wminfo.Fpartime = TICRATE * pars[gameepisode][gamemap]
		} else {
			wminfo.Fpartime = TICRATE * cpars[gamemap]
		}
	}
	wminfo.Fpnum = consoleplayer
	for i := range MAXPLAYERS {
		wminfo.Fplyr[i].Fin = playeringame[i]
		wminfo.Fplyr[i].Fskills = players[i].Fkillcount
		wminfo.Fplyr[i].Fsitems = players[i].Fitemcount
		wminfo.Fplyr[i].Fssecret = players[i].Fsecretcount
		wminfo.Fplyr[i].Fstime = leveltime
		wminfo.Fplyr[i].Ffrags = players[i].Ffrags
	}
	gamestate = gs_INTERMISSION
	viewactive = 0
	automapactive = 0
	statCopy(&wminfo)
	wi_Start(&wminfo)
}

// C documentation
//
//	//
//	// G_WorldDone
//	//
func g_WorldDone() {
	gameaction = ga_worlddone
	if secretexit != 0 {
		players[consoleplayer].Fdidsecret = 1
	}
	if gamemode == commercial {
		switch gamemap {
		case 15:
			fallthrough
		case 31:
			if secretexit == 0 {
				break
			}
			fallthrough
		case 6:
			fallthrough
		case 11:
			fallthrough
		case 20:
			fallthrough
		case 30:
			f_StartFinale()
			break
		}
	}
}

func g_DoWorldDone() {
	gamestate = gs_LEVEL
	gamemap = wminfo.Fnext + 1
	g_DoLoadLevel()
	gameaction = ga_nothing
	viewactive = 1
}

func g_LoadGame(name string) {
	savename = name
	gameaction = ga_loadgame
}

func g_DoLoadGame() {
	var savedleveltime int32
	var err error
	gameaction = ga_nothing
	save_stream, err = os.Open(savename)
	if err != nil {
		log.Printf("g_DoLoadGame: error opening savegame file %s: %v\n", savename, err)
		return
	}
	defer save_stream.Close()
	savegame_error = 0
	if p_ReadSaveGameHeader() == 0 {
		save_stream.Close()
		return
	}
	savedleveltime = leveltime
	// load a base level
	g_InitNew(gameskill, gameepisode, gamemap)
	leveltime = savedleveltime
	// dearchive all the modifications
	p_UnArchivePlayers()
	p_UnArchiveWorld()
	p_UnArchiveThinkers()
	p_UnArchiveSpecials()
	if p_ReadSaveGameEOF() == 0 {
		i_Error("Bad savegame")
	}
	if setsizeneeded != 0 {
		r_ExecuteSetViewSize()
	}
	// draw the pattern into the back screen
	r_FillBackScreen()
}

// C documentation
//
//	//
//	// G_SaveGame
//	// Called by the menu task.
//	// Description is a 24 byte text string
//	//
func g_SaveGame(slot int32, description string) {
	savegameslot = slot
	savedescription = description
	sendsave = 1
}

func g_DoSaveGame() {
	var recovery_savegame_file, savegame_file, temp_savegame_file string
	recovery_savegame_file = ""
	temp_savegame_file = p_TempSaveGameFile()
	savegame_file = p_SaveGameFile(savegameslot)
	// Open the savegame file for writing.  We write to a temporary file
	// and then rename it at the end if it was successfully written.
	// This prevents an existing savegame from being overwritten by
	// a corrupted one, or if a savegame buffer overrun occurs.
	var err error
	save_stream, err = os.OpenFile(temp_savegame_file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("g_DoSaveGame: error opening savegame file %s: %v\n", temp_savegame_file, err)
		// Failed to save the game, so we're going to have to abort. But
		// to be nice, save to somewhere else before we call i_Error().
		recovery_savegame_file = m_TempFile("recovery.dsg")
		save_stream, err = os.OpenFile(recovery_savegame_file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Printf("g_DoSaveGame: error opening recovery savegame file %s: %v\n", recovery_savegame_file, err)
			i_Error("Failed to open either '%s' or '%s' to write savegame.", temp_savegame_file, recovery_savegame_file)
		}
	}
	savegame_error = 0
	p_WriteSaveGameHeader(savedescription)
	p_ArchivePlayers()
	p_ArchiveWorld()
	p_ArchiveThinkers()
	p_ArchiveSpecials()
	p_WriteSaveGameEOF()
	// Enforce the same savegame size limit as in Vanilla Doom,
	// except if the vanilla_savegame_limit setting is turned off.
	pos, err := save_stream.Seek(0, io.SeekCurrent)
	if vanilla_savegame_limit != 0 && pos > SAVEGAMESIZE {
		i_Error("Savegame buffer overrun")
	}
	// Finish up, close the savegame file.
	save_stream.Close()
	if recovery_savegame_file != "" {
		// We failed to save to the normal location, but we wrote a
		// recovery file to the temp directory. Now we can bomb out
		// with an error.
		i_Error("Failed to open savegame file '%s' for writing.\nBut your game has been saved to '%s' for recovery.", temp_savegame_file, recovery_savegame_file)
	}
	// Now rename the temporary savegame file to the actual savegame
	// file, overwriting the old savegame if there was one there.
	os.Remove(savegame_file) // remove the old savegame file
	os.Rename(temp_savegame_file, savegame_file)
	gameaction = ga_nothing
	savedescription = ""
	players[consoleplayer].Fmessage = "game saved."
	// draw the pattern into the back screen
	r_FillBackScreen()
}

func g_DeferedInitNew(skill skill_t, episode int32, map1 int32) {
	d_skill = skill
	d_episode = episode
	d_map = map1
	gameaction = ga_newgame
}

func g_DoNewGame() {
	var v1, v2 boolean
	demoplayback = 0
	netdemo = 0
	netgame = 0
	deathmatch = 0
	v2 = 0
	playeringame[3] = v2
	v1 = v2
	playeringame[2] = v1
	playeringame[1] = v1
	respawnparm = 0
	fastparm = 0
	nomonsters = 0
	consoleplayer = 0
	g_InitNew(d_skill, d_episode, d_map)
	gameaction = ga_nothing
}

func g_InitNew(skill skill_t, episode int32, map1 int32) {
	var i int32
	var skytexturename string
	if paused != 0 {
		paused = 0
		s_ResumeSound()
	}
	/*
		    // Note: This commented-out block of code was added at some point
		    // between the DOS version(s) and the Doom source release. It isn't
		    // found in disassemblies of the DOS version and causes IDCLEV and
		    // the -warp command line parameter to behave differently.
		    // This is left here for posterity.
		    // This was quite messy with SPECIAL and commented parts.
		    // Supposedly hacks to make the latest edition work.
		    // It might not work properly.
		    if (episode < 1)
		      episode = 1;
		    if ( gamemode == retail )
		    {
		      if (episode > 4)
			episode = 4;
		    }
		    else if ( gamemode == shareware )
		    {
		      if (episode > 1)
			   episode = 1;	// only start episode 1 on shareware
		    }
		    else
		    {
		      if (episode > 3)
			episode = 3;
		    }
	*/
	if skill > sk_nightmare {
		skill = sk_nightmare
	}
	if gameversion >= exe_ultimate {
		if episode == 0 {
			episode = 4
		}
	} else {
		if episode < 1 {
			episode = 1
		}
		if episode > 3 {
			episode = 3
		}
	}
	if episode > 1 && gamemode == shareware {
		episode = 1
	}
	if map1 < 1 {
		map1 = 1
	}
	if map1 > 9 && gamemode != commercial {
		map1 = 9
	}
	m_ClearRandom()
	if skill == sk_nightmare || respawnparm != 0 {
		respawnmonsters = 1
	} else {
		respawnmonsters = 0
	}
	if fastparm != 0 || skill == sk_nightmare && gameskill != sk_nightmare {
		i = s_SARG_RUN1
		for {
			if !(i <= s_SARG_PAIN2) {
				break
			}
			states[i].Ftics >>= 1
			goto _1
		_1:
			;
			i++
		}
		mobjinfo[mt_BRUISERSHOT].Fspeed = 20 * (1 << FRACBITS)
		mobjinfo[mt_HEADSHOT].Fspeed = 20 * (1 << FRACBITS)
		mobjinfo[mt_TROOPSHOT].Fspeed = 20 * (1 << FRACBITS)
	} else {
		if skill != sk_nightmare && gameskill == sk_nightmare {
			i = s_SARG_RUN1
			for {
				if !(i <= s_SARG_PAIN2) {
					break
				}
				states[i].Ftics <<= 1
				goto _2
			_2:
				;
				i++
			}
			mobjinfo[mt_BRUISERSHOT].Fspeed = 15 * (1 << FRACBITS)
			mobjinfo[mt_HEADSHOT].Fspeed = 10 * (1 << FRACBITS)
			mobjinfo[mt_TROOPSHOT].Fspeed = 10 * (1 << FRACBITS)
		}
	}
	// force players to be initialized upon first level load
	for i := range MAXPLAYERS {
		players[i].Fplayerstate = Pst_REBORN
	}
	usergame = 1 // will be set false if a demo
	paused = 0
	demoplayback = 0
	automapactive = 0
	viewactive = 1
	gameepisode = episode
	gamemap = map1
	gameskill = skill
	viewactive = 1
	// Set the sky to use.
	//
	// Note: This IS broken, but it is how Vanilla Doom behaves.
	// See http://doomwiki.org/wiki/Sky_never_changes_in_Doom_II.
	//
	// Because we set the sky here at the start of a game, not at the
	// start of a level, the sky texture never changes unless we
	// restore from a saved game.  This was fixed before the Doom
	// source release, but this IS the way Vanilla DOS Doom behaves.
	if gamemode == commercial {
		if gamemap < 12 {
			skytexturename = "SKY1"
		} else {
			if gamemap < 21 {
				skytexturename = "SKY2"
			} else {
				skytexturename = "SKY3"
			}
		}
	} else {
		switch gameepisode {
		default:
			fallthrough
		case 1:
			skytexturename = "SKY1"
		case 2:
			skytexturename = "SKY2"
		case 3:
			skytexturename = "SKY3"
		case 4: // Special Edition sky
			skytexturename = "SKY4"
			break
		}
	}
	skytexture = r_TextureNumForName(skytexturename)
	g_DoLoadLevel()
}

//
// DEMO RECORDING
//

func g_ReadDemoTiccmd(cmd *ticcmd_t) {
	if demobuffer[demo_pos] == DEMOMARKER {
		// end of demo data stream
		g_CheckDemoStatus()
		return
	}
	cmd.Fforwardmove = int8(demobuffer[demo_pos])
	demo_pos++
	cmd.Fsidemove = int8(demobuffer[demo_pos])
	demo_pos++
	// If this is a longtics demo, read back in higher resolution
	if longtics != 0 {
		cmd.Fangleturn = int16(demobuffer[demo_pos]) | int16(demobuffer[demo_pos+1])<<8
		demo_pos += 2
	} else {
		cmd.Fangleturn = int16(demobuffer[demo_pos]) << 8
		demo_pos++
	}
	cmd.Fbuttons = demobuffer[demo_pos]
	demo_pos++
}

func g_WriteDemoTiccmd(cmd *ticcmd_t) {
	if gamekeydown[key_demo_quit] { // press q to end demo recording
		g_CheckDemoStatus()
	}
	demo_start := demo_pos
	if len(demobuffer)-demo_pos < 6 {
		// not enough space left in the demo buffer
		demobuffer = append(demobuffer, make([]byte, 64)...)
	}
	demobuffer[demo_pos] = uint8(cmd.Fforwardmove)
	demo_pos++
	demobuffer[demo_pos] = uint8(cmd.Fsidemove)
	demo_pos++
	// If this is a longtics demo, record in higher resolution
	if longtics != 0 {
		demobuffer[demo_pos] = uint8(int32(cmd.Fangleturn) & 0xff)
		demo_pos++
		demobuffer[demo_pos] = uint8(int32(cmd.Fangleturn) >> 8 & 0xff)
		demo_pos++
	} else {
		demobuffer[demo_pos] = uint8(int32(cmd.Fangleturn) >> 8)
		demo_pos++
	}
	demobuffer[demo_pos] = cmd.Fbuttons
	demo_pos++
	// reset demo pointer back
	demo_pos = demo_start
	g_ReadDemoTiccmd(cmd) // make SURE it is exactly the same
}

// C documentation
//
//	//
//	// G_RecordDemo
//	//
func g_RecordDemo(name string) {
	var i, maxsize int32
	usergame = 0
	demoname = fmt.Sprintf("%s.lmp", name)
	maxsize = 0x20000
	//!
	// @arg <size>
	// @category demo
	// @vanilla
	//
	// Specify the demo buffer size (KiB)
	//
	i = m_CheckParmWithArgs("-maxdemo", 1)
	if i != 0 {
		v, _ := strconv.Atoi(myargs[i+1])
		maxsize = int32(v) * 1024
	}
	demobuffer = make([]byte, maxsize)
	demorecording = 1
}

// C documentation
//
//	// Get the demo version code appropriate for the version set in gameversion.
func g_VanillaVersionCode() int32 {
	switch gameversion {
	case exe_doom_1_2:
		i_Error("Doom 1.2 does not have a version code!")
		fallthrough
	case exe_doom_1_666:
		return 106
	case exe_doom_1_7:
		return 107
	case exe_doom_1_8:
		return 108
	case exe_doom_1_9:
		fallthrough
	default: // All other versions are variants on v1.9:
		return 109
	}
	return 0
}

func g_BeginRecording() {
	//!
	// @category demo
	//
	// Record a high resolution "Doom 1.91" demo.
	//
	longtics = booluint32(m_CheckParm("-longtics") != 0)
	// If not recording a longtics demo, record in low res
	lowres_turn = booluint32(longtics == 0)
	demo_pos = 0
	if len(demobuffer) < 64 {
		demobuffer = append(demobuffer, make([]byte, 64)...)
	}
	// Save the right version code for this demo
	if longtics != 0 {
		demobuffer[demo_pos] = uint8(DOOM_191_VERSION)
		demo_pos++
	} else {
		demobuffer[demo_pos] = uint8(g_VanillaVersionCode())
		demo_pos++
	}
	demobuffer[demo_pos] = uint8(gameskill)
	demo_pos++
	demobuffer[demo_pos] = uint8(gameepisode)
	demo_pos++
	demobuffer[demo_pos] = uint8(gamemap)
	demo_pos++
	demobuffer[demo_pos] = uint8(deathmatch)
	demo_pos++
	demobuffer[demo_pos] = uint8(respawnparm)
	demo_pos++
	demobuffer[demo_pos] = uint8(fastparm)
	demo_pos++
	demobuffer[demo_pos] = uint8(nomonsters)
	demo_pos++
	demobuffer[demo_pos] = uint8(consoleplayer)
	demo_pos++
	for i := range MAXPLAYERS {
		demobuffer[demo_pos] = uint8(playeringame[i])
	}
}

func g_DeferedPlayDemo(name string) {
	if dont_run_demo {
		return
	}
	defdemoname = name
	gameaction = ga_playdemo
}

// Generate a string describing a demo version

func demoVersionDescription(version int32) string {
	switch version {
	case 104:
		return "v1.4"
	case 105:
		return "v1.5"
	case 106:
		return "v1.6/v1.666"
	case 107:
		return "v1.7/v1.7a"
	case 108:
		return "v1.8"
	case 109:
		return "v1.9"
	default:
		break
	}
	// Unknown version.  Perhaps this is a pre-v1.4 IWAD?  If the version
	// byte is in the range 0-4 then it can be a v1.0-v1.2 demo.
	if version >= 0 && version <= 4 {
		return "v1.0/v1.1/v1.2"
	} else {
		return fmt.Sprintf("%d.%d (unknown)", version/100, version%100)
	}
}

func g_DoPlayDemo() {
	var demoversion, episode, map1 int32
	var skill skill_t
	gameaction = ga_nothing
	num := w_GetNumForName(defdemoname)
	length := w_LumpLength(uint32(num))
	demodata := w_CacheLumpNum(num)
	demobuffer = make([]byte, length)
	copy(demobuffer, unsafe.Slice((*uint8)(unsafe.Pointer(demodata)), length))
	demo_pos = 0
	demoversion = int32(demobuffer[demo_pos])
	demo_pos++
	if demoversion == g_VanillaVersionCode() {
		longtics = 0
	} else {
		if demoversion == int32(DOOM_191_VERSION) {
			// demo recorded with cph's modified "v1.91" doom exe
			longtics = 1
		} else {
			//i_Error(message, demoversion, g_VanillaVersionCode(),
			fprintf_ccgo(os.Stdout, "Demo is from a different game version!\n(read %d, should be %d)\n\n*** You may need to upgrade your version of Doom to v1.9. ***\n    See: https://www.doomworld.com/classicdoom/info/patches.php\n    This appears to be %s.", demoversion, g_VanillaVersionCode(), demoVersionDescription(demoversion))
		}
	}
	skill = skill_t(demobuffer[demo_pos])
	demo_pos++
	episode = int32(demobuffer[demo_pos])
	demo_pos++
	map1 = int32(demobuffer[demo_pos])
	demo_pos++
	deathmatch = int32(demobuffer[demo_pos])
	demo_pos++
	respawnparm = uint32(demobuffer[demo_pos])
	demo_pos++
	fastparm = uint32(demobuffer[demo_pos])
	demo_pos++
	nomonsters = uint32(demobuffer[demo_pos])
	demo_pos++
	consoleplayer = int32(demobuffer[demo_pos])
	demo_pos++
	for i := range MAXPLAYERS {
		playeringame[i] = uint32(demobuffer[demo_pos])
		demo_pos++
	}
	if playeringame[1] != 0 || m_CheckParm("-solo-net") > 0 || m_CheckParm("-netdemo") > 0 {
		netgame = 1
		netdemo = 1
	}
	// don't spend a lot of time in loadlevel
	precache = 0
	g_InitNew(skill, episode, map1)
	precache = 1
	starttime = i_GetTime()
	usergame = 0
	demoplayback = 1
}

// C documentation
//
//	//
//	// G_TimeDemo
//	//
func g_TimeDemo(name string) {
	//!
	// @vanilla
	//
	// Disable rendering the screen entirely.
	//
	nodrawers = uint32(m_CheckParm("-nodraw"))
	timingdemo = 1
	singletics = 1
	defdemoname = name
	gameaction = ga_playdemo
}

/*
===================
=
= G_CheckDemoStatus
=
= Called after a death or level completion to allow demos to be cleaned up
= Returns true if a new demo loop action will take place
===================
*/
func g_CheckDemoStatus() {
	var endtime, realtics int32
	var fps float32
	var v1, v2 boolean
	if timingdemo != 0 {
		endtime = i_GetTime()
		realtics = endtime - starttime
		fps = float32(gametic) * float32(TICRATE) / float32(realtics)
		// Prevent recursive calls
		timingdemo = 0
		demoplayback = 0
		i_Error("timed %d gametics in %d realtics (%f fps)", gametic, realtics, float64(fps))
	}
	if demoplayback != 0 {
		w_ReleaseLumpName(defdemoname)
		demoplayback = 0
		netdemo = 0
		netgame = 0
		deathmatch = 0
		v2 = 0
		playeringame[3] = v2
		v1 = v2
		playeringame[2] = v1
		playeringame[1] = v1
		respawnparm = 0
		fastparm = 0
		nomonsters = 0
		consoleplayer = 0
		if singledemo != 0 {
			i_Quit()
		} else {
			d_AdvanceDemo()
		}
		return
	}
	if demorecording != 0 {
		demobuffer[demo_pos] = uint8(DEMOMARKER)
		demo_pos++
		m_WriteFile(demoname, demobuffer[:demo_pos])
		demobuffer = nil
		demorecording = 0
		i_Error("Demo %s recorded", demoname)
	}
	return
}

const HU_MAXLINELENGTH = 80
const KEY_BACKSPACE1 = 127

type hu_textline_t struct {
	Fx           int32
	Fy           int32
	Ff           []*patch_t
	Fsc          int32
	Fl           [81]byte
	Flen1        int32
	Fneedsupdate int32
}

type hu_stext_t struct {
	Fl      [4]hu_textline_t
	Fh      int32
	Fcl     int32
	Fon     *boolean
	Flaston boolean
}

type hu_itext_t struct {
	Fl      hu_textline_t
	Flm     int32
	Fon     *boolean
	Flaston boolean
}

func hulib_clearTextLine(t *hu_textline_t) {
	t.Flen1 = 0
	t.Fl[0] = 0
	t.Fneedsupdate = 1
}

func hulib_initTextLine(t *hu_textline_t, x int32, y int32, f []*patch_t, sc int32) {
	t.Fx = x
	t.Fy = y
	t.Ff = f
	t.Fsc = sc
	hulib_clearTextLine(t)
}

func hulib_addCharToTextLine(t *hu_textline_t, ch byte) boolean {
	if t.Flen1 == HU_MAXLINELENGTH {
		return 0
	} else {
		t.Fl[t.Flen1] = ch
		t.Flen1++
		t.Fl[t.Flen1] = 0
		t.Fneedsupdate = 4
		return 1
	}
}

func hulib_delCharFromTextLine(t *hu_textline_t) boolean {
	if t.Flen1 == 0 {
		return 0
	} else {
		t.Flen1--
		t.Fl[t.Flen1] = 0
		t.Fneedsupdate = 4
		return 1
	}
}

func hulib_drawTextLine(l *hu_textline_t, drawcursor boolean) {
	var c uint8
	var w, x int32
	// draw the new stuff
	x = l.Fx
	for i := range l.Flen1 {
		c = uint8(xtoupper(int32(l.Fl[i])))
		if int32(c) != ' ' && int32(c) >= l.Fsc && int32(c) <= '_' {
			w = int32(l.Ff[int32(c)-l.Fsc].Fwidth)
			if x+w > SCREENWIDTH {
				break
			}
			v_DrawPatchDirect(x, l.Fy, l.Ff[int32(c)-l.Fsc])
			x += w
		} else {
			x += 4
			if x >= SCREENWIDTH {
				break
			}
		}
	}
	// draw the cursor if requested
	if drawcursor != 0 && x+int32(l.Ff['_'-l.Fsc].Fwidth) <= SCREENWIDTH {
		v_DrawPatchDirect(x, l.Fy, l.Ff['_'-l.Fsc])
	}
}

// C documentation
//
//	// sorta called by hu_Erase and just better darn get things straight
func hulib_eraseTextLine(l *hu_textline_t) {
	var lh, y, yoffset int32
	// Only erases when NOT in automap and the screen is reduced,
	// and the text must either need updating or refreshing
	// (because of a recent change back from the automap)
	if automapactive == 0 && viewwindowx != 0 && l.Fneedsupdate != 0 {
		lh = int32(l.Ff[0].Fheight) + 1
		y = l.Fy
		yoffset = y * SCREENWIDTH
		for {
			if y >= l.Fy+lh {
				break
			}
			if y < viewwindowy || y >= viewwindowy+viewheight {
				r_VideoErase(uint32(yoffset), SCREENWIDTH)
			} else {
				r_VideoErase(uint32(yoffset), viewwindowx) // erase left border
				r_VideoErase(uint32(yoffset+viewwindowx+viewwidth), viewwindowx)
				// erase right border
			}
			goto _1
		_1:
			;
			y++
			yoffset += SCREENWIDTH
		}
	}
	if l.Fneedsupdate != 0 {
		l.Fneedsupdate--
	}
}

func hulib_initSText(s *hu_stext_t, x int32, y int32, h int32, font []*patch_t, startchar int32, on *boolean) {
	s.Fh = h
	s.Fon = on
	s.Flaston = 1
	s.Fcl = 0
	for i := int32(0); i < h; i++ {
		hulib_initTextLine(&s.Fl[i], x, y-i*(int32(font[0].Fheight)+1), font, startchar)
	}
}

func hulib_addLineToSText(s *hu_stext_t) {
	// add a clear line
	s.Fcl++
	if s.Fcl == s.Fh {
		s.Fcl = 0
	}
	hulib_clearTextLine(&s.Fl[s.Fcl])
	// everything needs updating
	for i := int32(0); i < s.Fh; i++ {
		s.Fl[i].Fneedsupdate = 4 // needs updating
	}
}

func hulib_addMessageToSText(s *hu_stext_t, prefix string, msg string) {
	hulib_addLineToSText(s)
	for _, i := range prefix {
		hulib_addCharToTextLine(&s.Fl[s.Fcl], byte(i))
	}
	for _, i := range msg {
		hulib_addCharToTextLine(&s.Fl[s.Fcl], byte(i))
	}
}

func hulib_drawSText(s *hu_stext_t) {
	var idx int32
	if *s.Fon == 0 {
		return
	} // if not on, don't draw
	// draw everything
	for i := int32(0); i < s.Fh; i++ {
		idx = s.Fcl - i
		if idx < 0 {
			idx += s.Fh
		} // handle queue of lines
		l := &s.Fl[idx]
		// need a decision made here on whether to skip the draw
		hulib_drawTextLine(l, 0) // no cursor, please
	}
}

func hulib_eraseSText(s *hu_stext_t) {
	for i := int32(0); i < s.Fh; i++ {
		if s.Flaston != 0 && *s.Fon == 0 {
			s.Fl[i].Fneedsupdate = 4
		}
		hulib_eraseTextLine(&s.Fl[i])
	}
	s.Flaston = *s.Fon
}

func hulib_initIText(it *hu_itext_t, x int32, y int32, font []*patch_t, startchar int32, on *boolean) {
	it.Flm = 0 // default left margin is start of text
	it.Fon = on
	it.Flaston = 1
	hulib_initTextLine(&it.Fl, x, y, font, startchar)
}

// C documentation
//
//	// The following deletion routines adhere to the left margin restriction
func hulib_delCharFromIText(it *hu_itext_t) {
	if it.Fl.Flen1 != it.Flm {
		hulib_delCharFromTextLine(&it.Fl)
	}
}

// C documentation
//
//	// Resets left margin as well
func hulib_resetIText(it *hu_itext_t) {
	it.Flm = 0
	hulib_clearTextLine(&it.Fl)
}

// C documentation
//
//	// wrapper function for handling general keyed input.
//	// returns true if it ate the key
func hulib_keyInIText(it *hu_itext_t, ch uint8) boolean {
	ch = uint8(xtoupper(int32(ch)))
	if int32(ch) >= ' ' && int32(ch) <= '_' {
		hulib_addCharToTextLine(&it.Fl, ch)
	} else {
		if int32(ch) == int32(KEY_BACKSPACE1) {
			hulib_delCharFromIText(it)
		} else {
			if int32(ch) != KEY_ENTER {
				return 0
			}
		}
	} // did not eat key
	return 1 // ate the key
}

func hulib_drawIText(it *hu_itext_t) {
	if *it.Fon == 0 {
		return
	}
	hulib_drawTextLine(&it.Fl, 1) // draw the line w/ cursor
}

func hulib_eraseIText(it *hu_itext_t) {
	if it.Flaston != 0 && *it.Fon == 0 {
		it.Fl.Fneedsupdate = 4
	}
	hulib_eraseTextLine(&it.Fl)
	it.Flaston = *it.Fon
}

const HU_TITLEX = 0
const QUEUESIZE = 128

func init() {
	chat_macros = [10]string{
		0: "No",
		1: "I'm ready to kick butt!",
		2: "I'm OK.",
		3: "I'm not looking too good!",
		4: "Help!",
		5: "You suck!",
		6: "Next time, scumbag...",
		7: "Come here!",
		8: "I'll take care of it.",
		9: "Yes",
	}
}

func init() {
	player_names = [4]string{
		0: "Green: ",
		1: "Indigo: ",
		2: "Brown: ",
		3: "Red: ",
	}
}

var plr1 *player_t
var w_title hu_textline_t
var w_chat hu_itext_t
var always_off boolean = 0
var chat_dest [4]int8
var w_inputbuffer [4]hu_itext_t

var message_on boolean
var message_nottobefuckedwith boolean

var w_message hu_stext_t
var message_counter int32

var headsupactive = 0

func init() {
	mapnames = [45]string{
		0:  "E1M1: Hangar",
		1:  "E1M2: Nuclear Plant",
		2:  "E1M3: Toxin Refinery",
		3:  "E1M4: Command Control",
		4:  "E1M5: Phobos Lab",
		5:  "E1M6: Central Processing",
		6:  "E1M7: Computer Station",
		7:  "E1M8: Phobos Anomaly",
		8:  "E1M9: Military Base",
		9:  "E2M1: Deimos Anomaly",
		10: "E2M2: Containment Area",
		11: "E2M3: Refinery",
		12: "E2M4: Deimos Lab",
		13: "E2M5: Command Center",
		14: "E2M6: Halls of the Damned",
		15: "E2M7: Spawning Vats",
		16: "E2M8: Tower of Babel",
		17: "E2M9: Fortress of Mystery",
		18: "E3M1: Hell Keep",
		19: "E3M2: Slough of Despair",
		20: "E3M3: Pandemonium",
		21: "E3M4: House of Pain",
		22: "E3M5: Unholy Cathedral",
		23: "E3M6: Mt. Erebus",
		24: "E3M7: Limbo",
		25: "E3M8: Dis",
		26: "E3M9: Warrens",
		27: "E4M1: Hell Beneath",
		28: "E4M2: Perfect Hatred",
		29: "E4M3: Sever The Wicked",
		30: "E4M4: Unruly Evil",
		31: "E4M5: They Will Repent",
		32: "E4M6: Against Thee Wickedly",
		33: "E4M7: And Hell Followed",
		34: "E4M8: Unto The Cruel",
		35: "E4M9: Fear",
		36: "NEWLEVEL",
		37: "NEWLEVEL",
		38: "NEWLEVEL",
		39: "NEWLEVEL",
		40: "NEWLEVEL",
		41: "NEWLEVEL",
		42: "NEWLEVEL",
		43: "NEWLEVEL",
		44: "NEWLEVEL",
	}
}

func init() {
	mapnames_commercial = [96]string{
		0:  "level 1: entryway",
		1:  "level 2: underhalls",
		2:  "level 3: the gantlet",
		3:  "level 4: the focus",
		4:  "level 5: the waste tunnels",
		5:  "level 6: the crusher",
		6:  "level 7: dead simple",
		7:  "level 8: tricks and traps",
		8:  "level 9: the pit",
		9:  "level 10: refueling base",
		10: "level 11: 'o' of destruction!",
		11: "level 12: the factory",
		12: "level 13: downtown",
		13: "level 14: the inmost dens",
		14: "level 15: industrial zone",
		15: "level 16: suburbs",
		16: "level 17: tenements",
		17: "level 18: the courtyard",
		18: "level 19: the citadel",
		19: "level 20: gotcha!",
		20: "level 21: nirvana",
		21: "level 22: the catacombs",
		22: "level 23: barrels o' fun",
		23: "level 24: the chasm",
		24: "level 25: bloodfalls",
		25: "level 26: the abandoned mines",
		26: "level 27: monster condo",
		27: "level 28: the spirit world",
		28: "level 29: the living end",
		29: "level 30: icon of sin",
		30: "level 31: wolfenstein",
		31: "level 32: grosse",
		32: "level 1: congo",
		33: "level 2: well of souls",
		34: "level 3: aztec",
		35: "level 4: caged",
		36: "level 5: ghost town",
		37: "level 6: baron's lair",
		38: "level 7: caughtyard",
		39: "level 8: realm",
		40: "level 9: abattoire",
		41: "level 10: onslaught",
		42: "level 11: hunted",
		43: "level 12: speed",
		44: "level 13: the crypt",
		45: "level 14: genesis",
		46: "level 15: the twilight",
		47: "level 16: the omen",
		48: "level 17: compound",
		49: "level 18: neurosphere",
		50: "level 19: nme",
		51: "level 20: the death domain",
		52: "level 21: slayer",
		53: "level 22: impossible mission",
		54: "level 23: tombstone",
		55: "level 24: the final frontier",
		56: "level 25: the temple of darkness",
		57: "level 26: bunker",
		58: "level 27: anti-christ",
		59: "level 28: the sewers",
		60: "level 29: odyssey of noises",
		61: "level 30: the gateway of hell",
		62: "level 31: cyberden",
		63: "level 32: go 2 it",
		64: "level 1: system control",
		65: "level 2: human bbq",
		66: "level 3: power control",
		67: "level 4: wormhole",
		68: "level 5: hanger",
		69: "level 6: open season",
		70: "level 7: prison",
		71: "level 8: metal",
		72: "level 9: stronghold",
		73: "level 10: redemption",
		74: "level 11: storage facility",
		75: "level 12: crater",
		76: "level 13: nukage processing",
		77: "level 14: steel works",
		78: "level 15: dead zone",
		79: "level 16: deepest reaches",
		80: "level 17: processing area",
		81: "level 18: mill",
		82: "level 19: shipping/respawning",
		83: "level 20: central processing",
		84: "level 21: administration center",
		85: "level 22: habitat",
		86: "level 23: lunar mining project",
		87: "level 24: quarry",
		88: "level 25: baron's den",
		89: "level 26: ballistyx",
		90: "level 27: mount pain",
		91: "level 28: heck",
		92: "level 29: river styx",
		93: "level 30: last call",
		94: "level 31: pharaoh",
		95: "level 32: caribbean",
	}
}

func hu_Init() {
	var i, j, v2 int32
	// load the heads-up font
	j = '!'
	i = 0
	for {
		if i >= '_'-'!'+1 {
			break
		}
		v2 = j
		j++
		name := fmt.Sprintf("STCFN%.3d", v2)
		hu_font[i] = w_CacheLumpNameT(name)
		goto _1
	_1:
		;
		i++
	}
}

func hu_Stop() {
	headsupactive = 0
}

func hu_Start() {
	var v1 gamemission_t
	var s string
	if headsupactive != 0 {
		hu_Stop()
	}
	plr1 = &players[consoleplayer]
	message_on = 0
	message_dontfuckwithme = 0
	message_nottobefuckedwith = 0
	chat_on = 0
	// create the message widget
	hulib_initSText(&w_message, HU_MSGX, HU_MSGY, HU_MSGHEIGHT, hu_font[:], '!', &message_on)
	// create the map title widget
	hulib_initTextLine(&w_title, HU_TITLEX, 167-int32(hu_font[0].Fheight), hu_font[:], '!')
	if gamemission == pack_chex {
		v1 = doom
	} else {
		if gamemission == pack_hacx {
			v1 = doom2
		} else {
			v1 = gamemission
		}
	}
	switch v1 {
	case doom:
		s = mapnames[(gameepisode-1)*9+gamemap-1]
	case doom2:
		s = mapnames_commercial[gamemap-1]
	case pack_plut:
		s = mapnames_commercial[gamemap-1+int32(32)]
	case pack_tnt:
		s = mapnames_commercial[gamemap-1+int32(64)]
	default:
		s = "Unknown level"
		break
	}
	// Chex.exe always uses the episode 1 level title
	// eg. E2M1 gives the title for E1M1
	if gameversion == exe_chex {
		s = mapnames[gamemap-1]
	}
	// dehacked substitution to get modified level name
	for _, i := range s {
		hulib_addCharToTextLine(&w_title, byte(i))
	}
	// create the chat widget
	hulib_initIText(&w_chat, HU_MSGX, HU_MSGY+HU_MSGHEIGHT*(int32(hu_font[0].Fheight)+1), hu_font[:], '!', &chat_on)
	// create the inputbuffer widgets
	for i := range MAXPLAYERS {
		hulib_initIText(&w_inputbuffer[i], 0, 0, nil, 0, &always_off)
	}
	headsupactive = 1
}

func hu_Drawer() {
	hulib_drawSText(&w_message)
	hulib_drawIText(&w_chat)
	if automapactive != 0 {
		hulib_drawTextLine(&w_title, 0)
	}
}

func hu_Erase() {
	hulib_eraseSText(&w_message)
	hulib_eraseIText(&w_chat)
	hulib_eraseTextLine(&w_title)
}

func hu_Ticker() {
	var c, v4 int8
	var rc, v1 int32
	var v2, v5 bool
	// tick down message counter if message is up
	if v2 = message_counter != 0; v2 {
		message_counter--
		v1 = message_counter
	}
	if v2 && v1 == 0 {
		message_on = 0
		message_nottobefuckedwith = 0
	}
	if showMessages != 0 || message_dontfuckwithme != 0 {
		// display message if necessary
		if plr1.Fmessage != "" && message_nottobefuckedwith == 0 || plr1.Fmessage != "" && message_dontfuckwithme != 0 {
			hulib_addMessageToSText(&w_message, "", plr1.Fmessage)
			plr1.Fmessage = ""
			message_on = 1
			message_counter = 4 * TICRATE
			message_nottobefuckedwith = message_dontfuckwithme
			message_dontfuckwithme = 0
		}
	} // else message_on = false;
	// check for incoming chat characters
	if netgame != 0 {
		for i := range int32(MAXPLAYERS) {
			if playeringame[i] == 0 {
				continue
			}
			if v5 = i != consoleplayer; v5 {
				v4 = int8(players[i].Fcmd.Fchatchar)
				c = v4
			}
			if v5 && v4 != 0 {
				if int32(c) <= HU_BROADCAST {
					chat_dest[i] = c
				} else {
					rc = int32(hulib_keyInIText(&w_inputbuffer[i], uint8(c)))
					if rc != 0 && int32(c) == KEY_ENTER {
						if w_inputbuffer[i].Fl.Flen1 != 0 && (int32(chat_dest[i]) == consoleplayer+1 || int32(chat_dest[i]) == HU_BROADCAST) {
							hulib_addMessageToSText(&w_message, player_names[i], gostring_bytes(w_inputbuffer[i].Fl.Fl[:]))
							message_nottobefuckedwith = 1
							message_on = 1
							message_counter = 4 * TICRATE
							if gamemode == commercial {
								s_StartSound(nil, int32(sfx_radio))
							} else {
								s_StartSound(nil, int32(sfx_tink))
							}
						}
						hulib_resetIText(&w_inputbuffer[i])
					}
				}
				players[i].Fcmd.Fchatchar = 0
			}
		}
	}
}

var chatchars [128]int8
var head = 0
var tail = 0

func hu_queueChatChar(c int8) {
	if (head+1)&(QUEUESIZE-1) == tail {
		plr1.Fmessage = "[Message unsent]"
	} else {
		chatchars[head] = c
		head = (head + 1) & (QUEUESIZE - 1)
	}
}

func hu_dequeueChatChar() int8 {
	var c int8
	if head != tail {
		c = chatchars[tail]
		tail = (tail + 1) & (QUEUESIZE - 1)
	} else {
		c = 0
	}
	return c
}

func hu_Responder(ev *event_t) boolean {
	var c uint8
	var eatkey, v2, v4 boolean
	var numplayers int32
	var macromessage string
	eatkey = 0
	numplayers = 0
	for i := range MAXPLAYERS {
		numplayers = int32(uint32(numplayers) + playeringame[i])
	}
	if ev.Fdata1 == 0x80+0x36 {
		return 0
	} else {
		if ev.Fdata1 == 0x80+0x38 {
			altdown = booluint32(ev.Ftype1 == Ev_keydown)
			return 0
		}
	}
	if ev.Ftype1 != Ev_keydown {
		return 0
	}
	if chat_on == 0 {
		if ev.Fdata1 == key_message_refresh {
			message_on = 1
			message_counter = 4 * TICRATE
			eatkey = 1
		} else {
			if netgame != 0 && ev.Fdata2 == key_multi_msg {
				v2 = 1
				chat_on = v2
				eatkey = v2
				hulib_resetIText(&w_chat)
				hu_queueChatChar(int8(HU_BROADCAST))
			} else {
				if netgame != 0 && numplayers > 2 {
					for i := range int32(MAXPLAYERS) {
						if ev.Fdata2 == key_multi_msgplayer[i] {
							if playeringame[i] != 0 && i != consoleplayer {
								v4 = 1
								chat_on = v4
								eatkey = v4
								hulib_resetIText(&w_chat)
								hu_queueChatChar(int8(i + 1))
								break
							} else {
								if i == consoleplayer {
									num_nobrainers++
									if num_nobrainers < 3 {
										plr1.Fmessage = "You mumble to yourself"
									} else {
										if num_nobrainers < 6 {
											plr1.Fmessage = "Who's there?"
										} else {
											if num_nobrainers < 9 {
												plr1.Fmessage = "You scare yourself"
											} else {
												if num_nobrainers < 32 {
													plr1.Fmessage = "You start to rave"
												} else {
													plr1.Fmessage = "You've lost it..."
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	} else {
		// send a macro
		if altdown != 0 {
			c = uint8(ev.Fdata1 - '0')
			if int32(c) > 9 {
				return 0
			}
			// fprintf(stderr, "got here\n");
			macromessage = chat_macros[c]
			// kill last message with a '\n'
			hu_queueChatChar(int8(KEY_ENTER)) // DEBUG!!!
			// send the macro message
			for i := 0; macromessage[i] != 0; i++ {
				hu_queueChatChar(int8(macromessage[i]))
			}
			hu_queueChatChar(int8(KEY_ENTER))
			// leave chat mode and notify that it was sent
			chat_on = 0
			lastmessage = chat_macros[c]
			plr1.Fmessage = lastmessage
			eatkey = 1
		} else {
			c = uint8(ev.Fdata2)
			eatkey = hulib_keyInIText(&w_chat, c)
			if eatkey != 0 {
				// static unsigned char buf[20]; // DEBUG
				hu_queueChatChar(int8(c))
				// M_snprintf(buf, sizeof(buf), "KEY: %d => %d", ev->data1, c);
				//        plr->message = buf;
			}
			if int32(c) == KEY_ENTER {
				chat_on = 0
				if w_chat.Fl.Flen1 != 0 {
					lastmessage = gostring_bytes(w_chat.Fl.Fl[:])
					plr1.Fmessage = lastmessage
				}
			} else {
				if int32(c) == KEY_ESCAPE {
					chat_on = 0
				}
			}
		}
	}
	return eatkey
}

var lastmessage string

var altdown boolean

var num_nobrainers int32

func init() {
	sprnames = []string{
		0:   "TROO",
		1:   "SHTG",
		2:   "PUNG",
		3:   "PISG",
		4:   "PISF",
		5:   "SHTF",
		6:   "SHT2",
		7:   "CHGG",
		8:   "CHGF",
		9:   "MISG",
		10:  "MISF",
		11:  "SAWG",
		12:  "PLSG",
		13:  "PLSF",
		14:  "BFGG",
		15:  "BFGF",
		16:  "BLUD",
		17:  "PUFF",
		18:  "BAL1",
		19:  "BAL2",
		20:  "PLSS",
		21:  "PLSE",
		22:  "MISL",
		23:  "BFS1",
		24:  "BFE1",
		25:  "BFE2",
		26:  "TFOG",
		27:  "IFOG",
		28:  "PLAY",
		29:  "POSS",
		30:  "SPOS",
		31:  "VILE",
		32:  "FIRE",
		33:  "FATB",
		34:  "FBXP",
		35:  "SKEL",
		36:  "MANF",
		37:  "FATT",
		38:  "CPOS",
		39:  "SARG",
		40:  "HEAD",
		41:  "BAL7",
		42:  "BOSS",
		43:  "BOS2",
		44:  "SKUL",
		45:  "SPID",
		46:  "BSPI",
		47:  "APLS",
		48:  "APBX",
		49:  "CYBR",
		50:  "PAIN",
		51:  "SSWV",
		52:  "KEEN",
		53:  "BBRN",
		54:  "BOSF",
		55:  "ARM1",
		56:  "ARM2",
		57:  "BAR1",
		58:  "BEXP",
		59:  "FCAN",
		60:  "BON1",
		61:  "BON2",
		62:  "BKEY",
		63:  "RKEY",
		64:  "YKEY",
		65:  "BSKU",
		66:  "RSKU",
		67:  "YSKU",
		68:  "STIM",
		69:  "MEDI",
		70:  "SOUL",
		71:  "PINV",
		72:  "PSTR",
		73:  "PINS",
		74:  "MEGA",
		75:  "SUIT",
		76:  "PMAP",
		77:  "PVIS",
		78:  "CLIP",
		79:  "AMMO",
		80:  "ROCK",
		81:  "BROK",
		82:  "CELL",
		83:  "CELP",
		84:  "SHEL",
		85:  "SBOX",
		86:  "BPAK",
		87:  "BFUG",
		88:  "MGUN",
		89:  "CSAW",
		90:  "LAUN",
		91:  "PLAS",
		92:  "SHOT",
		93:  "SGN2",
		94:  "COLU",
		95:  "SMT2",
		96:  "GOR1",
		97:  "POL2",
		98:  "POL5",
		99:  "POL4",
		100: "POL3",
		101: "POL1",
		102: "POL6",
		103: "GOR2",
		104: "GOR3",
		105: "GOR4",
		106: "GOR5",
		107: "SMIT",
		108: "COL1",
		109: "COL2",
		110: "COL3",
		111: "COL4",
		112: "CAND",
		113: "CBRA",
		114: "COL6",
		115: "TRE1",
		116: "TRE2",
		117: "ELEC",
		118: "CEYE",
		119: "FSKU",
		120: "COL5",
		121: "TBLU",
		122: "TGRN",
		123: "TRED",
		124: "SMBT",
		125: "SMGT",
		126: "SMRT",
		127: "HDB1",
		128: "HDB2",
		129: "HDB3",
		130: "HDB4",
		131: "HDB5",
		132: "HDB6",
		133: "POB1",
		134: "POB2",
		135: "BRS1",
		136: "TLMP",
		137: "TLP2",
	}
}

// Original doom had function pointers relying on the base address of *mobj_t & *player_t being the same.
// Go doesn't like that without a lot of unsafe casting, so we use a wrapper function to convert
// the player function to a mobj function.
// We now relay on mo.Fplayer being set properly, but that is done in the game code
func playerFuncToAction(f func(p *player_t, psp *pspdef_t)) func(*mobj_t, *pspdef_t) {
	return func(mo *mobj_t, psp *pspdef_t) {
		f(mo.Fplayer, psp)
	}
}

func mobjFuncToAction(f func(mo *mobj_t)) func(*mobj_t, *pspdef_t) {
	return func(mo *mobj_t, psp *pspdef_t) {
		f(mo)
	}
}

func init() {
	states = [967]state_t{
		0: {
			Ftics: -1,
		},
		1: {
			Fsprite: spr_SHTG,
			Fframe:  4,
			Faction: playerFuncToAction(a_Light0),
		},
		2: {
			Fsprite:    spr_PUNG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_PUNCH,
		},
		3: {
			Fsprite:    spr_PUNG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_PUNCHDOWN,
		},
		4: {
			Fsprite:    spr_PUNG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_PUNCHUP,
		},
		5: {
			Fsprite:    spr_PUNG,
			Fframe:     1,
			Ftics:      4,
			Fnextstate: s_PUNCH2,
		},
		6: {
			Fsprite:    spr_PUNG,
			Fframe:     2,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Punch),
			Fnextstate: s_PUNCH3,
		},
		7: {
			Fsprite:    spr_PUNG,
			Fframe:     3,
			Ftics:      5,
			Fnextstate: s_PUNCH4,
		},
		8: {
			Fsprite:    spr_PUNG,
			Fframe:     2,
			Ftics:      4,
			Fnextstate: s_PUNCH5,
		},
		9: {
			Fsprite:    spr_PUNG,
			Fframe:     1,
			Ftics:      5,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_PUNCH,
		},
		10: {
			Fsprite:    spr_PISG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_PISTOL,
		},
		11: {
			Fsprite:    spr_PISG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_PISTOLDOWN,
		},
		12: {
			Fsprite:    spr_PISG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_PISTOLUP,
		},
		13: {
			Fsprite:    spr_PISG,
			Ftics:      4,
			Fnextstate: s_PISTOL2,
		},
		14: {
			Fsprite:    spr_PISG,
			Fframe:     1,
			Ftics:      6,
			Faction:    playerFuncToAction(a_FirePistol),
			Fnextstate: s_PISTOL3,
		},
		15: {
			Fsprite:    spr_PISG,
			Fframe:     2,
			Ftics:      4,
			Fnextstate: s_PISTOL4,
		},
		16: {
			Fsprite:    spr_PISG,
			Fframe:     1,
			Ftics:      5,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_PISTOL,
		},
		17: {
			Fsprite:    spr_PISF,
			Fframe:     32768,
			Ftics:      7,
			Faction:    playerFuncToAction(a_Light1),
			Fnextstate: s_LIGHTDONE,
		},
		18: {
			Fsprite:    spr_SHTG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_SGUN,
		},
		19: {
			Fsprite:    spr_SHTG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_SGUNDOWN,
		},
		20: {
			Fsprite:    spr_SHTG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_SGUNUP,
		},
		21: {
			Fsprite:    spr_SHTG,
			Ftics:      3,
			Fnextstate: s_SGUN2,
		},
		22: {
			Fsprite:    spr_SHTG,
			Ftics:      7,
			Faction:    playerFuncToAction(a_FireShotgun),
			Fnextstate: s_SGUN3,
		},
		23: {
			Fsprite:    spr_SHTG,
			Fframe:     1,
			Ftics:      5,
			Fnextstate: s_SGUN4,
		},
		24: {
			Fsprite:    spr_SHTG,
			Fframe:     2,
			Ftics:      5,
			Fnextstate: s_SGUN5,
		},
		25: {
			Fsprite:    spr_SHTG,
			Fframe:     3,
			Ftics:      4,
			Fnextstate: s_SGUN6,
		},
		26: {
			Fsprite:    spr_SHTG,
			Fframe:     2,
			Ftics:      5,
			Fnextstate: s_SGUN7,
		},
		27: {
			Fsprite:    spr_SHTG,
			Fframe:     1,
			Ftics:      5,
			Fnextstate: s_SGUN8,
		},
		28: {
			Fsprite:    spr_SHTG,
			Ftics:      3,
			Fnextstate: s_SGUN9,
		},
		29: {
			Fsprite:    spr_SHTG,
			Ftics:      7,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_SGUN,
		},
		30: {
			Fsprite:    spr_SHTF,
			Fframe:     32768,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Light1),
			Fnextstate: s_SGUNFLASH2,
		},
		31: {
			Fsprite:    spr_SHTF,
			Fframe:     32769,
			Ftics:      3,
			Faction:    playerFuncToAction(a_Light2),
			Fnextstate: s_LIGHTDONE,
		},
		32: {
			Fsprite:    spr_SHT2,
			Ftics:      1,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_DSGUN,
		},
		33: {
			Fsprite:    spr_SHT2,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_DSGUNDOWN,
		},
		34: {
			Fsprite:    spr_SHT2,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_DSGUNUP,
		},
		35: {
			Fsprite:    spr_SHT2,
			Ftics:      3,
			Fnextstate: s_DSGUN2,
		},
		36: {
			Fsprite:    spr_SHT2,
			Ftics:      7,
			Faction:    playerFuncToAction(a_FireShotgun2),
			Fnextstate: s_DSGUN3,
		},
		37: {
			Fsprite:    spr_SHT2,
			Fframe:     1,
			Ftics:      7,
			Fnextstate: s_DSGUN4,
		},
		38: {
			Fsprite:    spr_SHT2,
			Fframe:     2,
			Ftics:      7,
			Faction:    playerFuncToAction(a_CheckReload),
			Fnextstate: s_DSGUN5,
		},
		39: {
			Fsprite:    spr_SHT2,
			Fframe:     3,
			Ftics:      7,
			Faction:    playerFuncToAction(a_OpenShotgun2),
			Fnextstate: s_DSGUN6,
		},
		40: {
			Fsprite:    spr_SHT2,
			Fframe:     4,
			Ftics:      7,
			Fnextstate: s_DSGUN7,
		},
		41: {
			Fsprite:    spr_SHT2,
			Fframe:     5,
			Ftics:      7,
			Faction:    playerFuncToAction(a_LoadShotgun2),
			Fnextstate: s_DSGUN8,
		},
		42: {
			Fsprite:    spr_SHT2,
			Fframe:     6,
			Ftics:      6,
			Fnextstate: s_DSGUN9,
		},
		43: {
			Fsprite:    spr_SHT2,
			Fframe:     7,
			Ftics:      6,
			Faction:    playerFuncToAction(a_CloseShotgun2),
			Fnextstate: s_DSGUN10,
		},
		44: {
			Fsprite:    spr_SHT2,
			Ftics:      5,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_DSGUN,
		},
		45: {
			Fsprite:    spr_SHT2,
			Fframe:     1,
			Ftics:      7,
			Fnextstate: s_DSNR2,
		},
		46: {
			Fsprite:    spr_SHT2,
			Ftics:      3,
			Fnextstate: s_DSGUNDOWN,
		},
		47: {
			Fsprite:    spr_SHT2,
			Fframe:     32776,
			Ftics:      5,
			Faction:    playerFuncToAction(a_Light1),
			Fnextstate: s_DSGUNFLASH2,
		},
		48: {
			Fsprite:    spr_SHT2,
			Fframe:     32777,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Light2),
			Fnextstate: s_LIGHTDONE,
		},
		49: {
			Fsprite:    spr_CHGG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_CHAIN,
		},
		50: {
			Fsprite:    spr_CHGG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_CHAINDOWN,
		},
		51: {
			Fsprite:    spr_CHGG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_CHAINUP,
		},
		52: {
			Fsprite:    spr_CHGG,
			Ftics:      4,
			Faction:    playerFuncToAction(a_FireCGun),
			Fnextstate: s_CHAIN2,
		},
		53: {
			Fsprite:    spr_CHGG,
			Fframe:     1,
			Ftics:      4,
			Faction:    playerFuncToAction(a_FireCGun),
			Fnextstate: s_CHAIN3,
		},
		54: {
			Fsprite:    spr_CHGG,
			Fframe:     1,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_CHAIN,
		},
		55: {
			Fsprite:    spr_CHGF,
			Fframe:     32768,
			Ftics:      5,
			Faction:    playerFuncToAction(a_Light1),
			Fnextstate: s_LIGHTDONE,
		},
		56: {
			Fsprite:    spr_CHGF,
			Fframe:     32769,
			Ftics:      5,
			Faction:    playerFuncToAction(a_Light2),
			Fnextstate: s_LIGHTDONE,
		},
		57: {
			Fsprite:    spr_MISG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_MISSILE,
		},
		58: {
			Fsprite:    spr_MISG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_MISSILEDOWN,
		},
		59: {
			Fsprite:    spr_MISG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_MISSILEUP,
		},
		60: {
			Fsprite:    spr_MISG,
			Fframe:     1,
			Ftics:      8,
			Faction:    playerFuncToAction(a_GunFlash),
			Fnextstate: s_MISSILE2,
		},
		61: {
			Fsprite:    spr_MISG,
			Fframe:     1,
			Ftics:      12,
			Faction:    playerFuncToAction(a_FireMissile),
			Fnextstate: s_MISSILE3,
		},
		62: {
			Fsprite:    spr_MISG,
			Fframe:     1,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_MISSILE,
		},
		63: {
			Fsprite:    spr_MISF,
			Fframe:     32768,
			Ftics:      3,
			Faction:    playerFuncToAction(a_Light1),
			Fnextstate: s_MISSILEFLASH2,
		},
		64: {
			Fsprite:    spr_MISF,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_MISSILEFLASH3,
		},
		65: {
			Fsprite:    spr_MISF,
			Fframe:     32770,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Light2),
			Fnextstate: s_MISSILEFLASH4,
		},
		66: {
			Fsprite:    spr_MISF,
			Fframe:     32771,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Light2),
			Fnextstate: s_LIGHTDONE,
		},
		67: {
			Fsprite:    spr_SAWG,
			Fframe:     2,
			Ftics:      4,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_SAWB,
		},
		68: {
			Fsprite:    spr_SAWG,
			Fframe:     3,
			Ftics:      4,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_SAW,
		},
		69: {
			Fsprite:    spr_SAWG,
			Fframe:     2,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_SAWDOWN,
		},
		70: {
			Fsprite:    spr_SAWG,
			Fframe:     2,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_SAWUP,
		},
		71: {
			Fsprite:    spr_SAWG,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Saw),
			Fnextstate: s_SAW2,
		},
		72: {
			Fsprite:    spr_SAWG,
			Fframe:     1,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Saw),
			Fnextstate: s_SAW3,
		},
		73: {
			Fsprite:    spr_SAWG,
			Fframe:     1,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_SAW,
		},
		74: {
			Fsprite:    spr_PLSG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_PLASMA,
		},
		75: {
			Fsprite:    spr_PLSG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_PLASMADOWN,
		},
		76: {
			Fsprite:    spr_PLSG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_PLASMAUP,
		},
		77: {
			Fsprite:    spr_PLSG,
			Ftics:      3,
			Faction:    playerFuncToAction(a_FirePlasma),
			Fnextstate: s_PLASMA2,
		},
		78: {
			Fsprite:    spr_PLSG,
			Fframe:     1,
			Ftics:      20,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_PLASMA,
		},
		79: {
			Fsprite:    spr_PLSF,
			Fframe:     32768,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Light1),
			Fnextstate: s_LIGHTDONE,
		},
		80: {
			Fsprite:    spr_PLSF,
			Fframe:     32769,
			Ftics:      4,
			Faction:    playerFuncToAction(a_Light1),
			Fnextstate: s_LIGHTDONE,
		},
		81: {
			Fsprite:    spr_BFGG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_WeaponReady),
			Fnextstate: s_BFG,
		},
		82: {
			Fsprite:    spr_BFGG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Lower),
			Fnextstate: s_BFGDOWN,
		},
		83: {
			Fsprite:    spr_BFGG,
			Ftics:      1,
			Faction:    playerFuncToAction(a_Raise),
			Fnextstate: s_BFGUP,
		},
		84: {
			Fsprite:    spr_BFGG,
			Ftics:      20,
			Faction:    playerFuncToAction(a_BFGsound),
			Fnextstate: s_BFG2,
		},
		85: {
			Fsprite:    spr_BFGG,
			Fframe:     1,
			Ftics:      10,
			Faction:    playerFuncToAction(a_GunFlash),
			Fnextstate: s_BFG3,
		},
		86: {
			Fsprite:    spr_BFGG,
			Fframe:     1,
			Ftics:      10,
			Faction:    playerFuncToAction(a_FireBFG),
			Fnextstate: s_BFG4,
		},
		87: {
			Fsprite:    spr_BFGG,
			Fframe:     1,
			Ftics:      20,
			Faction:    playerFuncToAction(a_ReFire),
			Fnextstate: s_BFG,
		},
		88: {
			Fsprite:    spr_BFGF,
			Fframe:     32768,
			Ftics:      11,
			Faction:    playerFuncToAction(a_Light1),
			Fnextstate: s_BFGFLASH2,
		},
		89: {
			Fsprite:    spr_BFGF,
			Fframe:     32769,
			Ftics:      6,
			Faction:    playerFuncToAction(a_Light2),
			Fnextstate: s_LIGHTDONE,
		},
		90: {
			Fsprite:    spr_BLUD,
			Fframe:     2,
			Ftics:      8,
			Fnextstate: s_BLOOD2,
		},
		91: {
			Fsprite:    spr_BLUD,
			Fframe:     1,
			Ftics:      8,
			Fnextstate: s_BLOOD3,
		},
		92: {
			Fsprite: spr_BLUD,
			Ftics:   8,
		},
		93: {
			Fsprite:    spr_PUFF,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_PUFF2,
		},
		94: {
			Fsprite:    spr_PUFF,
			Fframe:     1,
			Ftics:      4,
			Fnextstate: s_PUFF3,
		},
		95: {
			Fsprite:    spr_PUFF,
			Fframe:     2,
			Ftics:      4,
			Fnextstate: s_PUFF4,
		},
		96: {
			Fsprite: spr_PUFF,
			Fframe:  3,
			Ftics:   4,
		},
		97: {
			Fsprite:    spr_BAL1,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_TBALL2,
		},
		98: {
			Fsprite:    spr_BAL1,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_TBALL1,
		},
		99: {
			Fsprite:    spr_BAL1,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_TBALLX2,
		},
		100: {
			Fsprite:    spr_BAL1,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_TBALLX3,
		},
		101: {
			Fsprite: spr_BAL1,
			Fframe:  32772,
			Ftics:   6,
		},
		102: {
			Fsprite:    spr_BAL2,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_RBALL2,
		},
		103: {
			Fsprite:    spr_BAL2,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_RBALL1,
		},
		104: {
			Fsprite:    spr_BAL2,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_RBALLX2,
		},
		105: {
			Fsprite:    spr_BAL2,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_RBALLX3,
		},
		106: {
			Fsprite: spr_BAL2,
			Fframe:  32772,
			Ftics:   6,
		},
		107: {
			Fsprite:    spr_PLSS,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_PLASBALL2,
		},
		108: {
			Fsprite:    spr_PLSS,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_PLASBALL,
		},
		109: {
			Fsprite:    spr_PLSE,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_PLASEXP2,
		},
		110: {
			Fsprite:    spr_PLSE,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_PLASEXP3,
		},
		111: {
			Fsprite:    spr_PLSE,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_PLASEXP4,
		},
		112: {
			Fsprite:    spr_PLSE,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_PLASEXP5,
		},
		113: {
			Fsprite: spr_PLSE,
			Fframe:  32772,
			Ftics:   4,
		},
		114: {
			Fsprite:    spr_MISL,
			Fframe:     32768,
			Ftics:      1,
			Fnextstate: s_ROCKET,
		},
		115: {
			Fsprite:    spr_BFS1,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_BFGSHOT2,
		},
		116: {
			Fsprite:    spr_BFS1,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_BFGSHOT,
		},
		117: {
			Fsprite:    spr_BFE1,
			Fframe:     32768,
			Ftics:      8,
			Fnextstate: s_BFGLAND2,
		},
		118: {
			Fsprite:    spr_BFE1,
			Fframe:     32769,
			Ftics:      8,
			Fnextstate: s_BFGLAND3,
		},
		119: {
			Fsprite:    spr_BFE1,
			Fframe:     32770,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_BFGSpray),
			Fnextstate: s_BFGLAND4,
		},
		120: {
			Fsprite:    spr_BFE1,
			Fframe:     32771,
			Ftics:      8,
			Fnextstate: s_BFGLAND5,
		},
		121: {
			Fsprite:    spr_BFE1,
			Fframe:     32772,
			Ftics:      8,
			Fnextstate: s_BFGLAND6,
		},
		122: {
			Fsprite: spr_BFE1,
			Fframe:  32773,
			Ftics:   8,
		},
		123: {
			Fsprite:    spr_BFE2,
			Fframe:     32768,
			Ftics:      8,
			Fnextstate: s_BFGEXP2,
		},
		124: {
			Fsprite:    spr_BFE2,
			Fframe:     32769,
			Ftics:      8,
			Fnextstate: s_BFGEXP3,
		},
		125: {
			Fsprite:    spr_BFE2,
			Fframe:     32770,
			Ftics:      8,
			Fnextstate: s_BFGEXP4,
		},
		126: {
			Fsprite: spr_BFE2,
			Fframe:  32771,
			Ftics:   8,
		},
		127: {
			Fsprite:    spr_MISL,
			Fframe:     32769,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Explode),
			Fnextstate: s_EXPLODE2,
		},
		128: {
			Fsprite:    spr_MISL,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_EXPLODE3,
		},
		129: {
			Fsprite: spr_MISL,
			Fframe:  32771,
			Ftics:   4,
		},
		130: {
			Fsprite:    spr_TFOG,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_TFOG01,
		},
		131: {
			Fsprite:    spr_TFOG,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_TFOG02,
		},
		132: {
			Fsprite:    spr_TFOG,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_TFOG2,
		},
		133: {
			Fsprite:    spr_TFOG,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_TFOG3,
		},
		134: {
			Fsprite:    spr_TFOG,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_TFOG4,
		},
		135: {
			Fsprite:    spr_TFOG,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_TFOG5,
		},
		136: {
			Fsprite:    spr_TFOG,
			Fframe:     32772,
			Ftics:      6,
			Fnextstate: s_TFOG6,
		},
		137: {
			Fsprite:    spr_TFOG,
			Fframe:     32773,
			Ftics:      6,
			Fnextstate: s_TFOG7,
		},
		138: {
			Fsprite:    spr_TFOG,
			Fframe:     32774,
			Ftics:      6,
			Fnextstate: s_TFOG8,
		},
		139: {
			Fsprite:    spr_TFOG,
			Fframe:     32775,
			Ftics:      6,
			Fnextstate: s_TFOG9,
		},
		140: {
			Fsprite:    spr_TFOG,
			Fframe:     32776,
			Ftics:      6,
			Fnextstate: s_TFOG10,
		},
		141: {
			Fsprite: spr_TFOG,
			Fframe:  32777,
			Ftics:   6,
		},
		142: {
			Fsprite:    spr_IFOG,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_IFOG01,
		},
		143: {
			Fsprite:    spr_IFOG,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_IFOG02,
		},
		144: {
			Fsprite:    spr_IFOG,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_IFOG2,
		},
		145: {
			Fsprite:    spr_IFOG,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_IFOG3,
		},
		146: {
			Fsprite:    spr_IFOG,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_IFOG4,
		},
		147: {
			Fsprite:    spr_IFOG,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_IFOG5,
		},
		148: {
			Fsprite: spr_IFOG,
			Fframe:  32772,
			Ftics:   6,
		},
		149: {
			Fsprite: spr_PLAY,
			Ftics:   -1,
		},
		150: {
			Fsprite:    spr_PLAY,
			Ftics:      4,
			Fnextstate: s_PLAY_RUN2,
		},
		151: {
			Fsprite:    spr_PLAY,
			Fframe:     1,
			Ftics:      4,
			Fnextstate: s_PLAY_RUN3,
		},
		152: {
			Fsprite:    spr_PLAY,
			Fframe:     2,
			Ftics:      4,
			Fnextstate: s_PLAY_RUN4,
		},
		153: {
			Fsprite:    spr_PLAY,
			Fframe:     3,
			Ftics:      4,
			Fnextstate: s_PLAY_RUN1,
		},
		154: {
			Fsprite:    spr_PLAY,
			Fframe:     4,
			Ftics:      12,
			Fnextstate: s_PLAY,
		},
		155: {
			Fsprite:    spr_PLAY,
			Fframe:     32773,
			Ftics:      6,
			Fnextstate: s_PLAY_ATK1,
		},
		156: {
			Fsprite:    spr_PLAY,
			Fframe:     6,
			Ftics:      4,
			Fnextstate: s_PLAY_PAIN2,
		},
		157: {
			Fsprite:    spr_PLAY,
			Fframe:     6,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_PLAY,
		},
		158: {
			Fsprite:    spr_PLAY,
			Fframe:     7,
			Ftics:      10,
			Fnextstate: s_PLAY_DIE2,
		},
		159: {
			Fsprite:    spr_PLAY,
			Fframe:     8,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_PlayerScream),
			Fnextstate: s_PLAY_DIE3,
		},
		160: {
			Fsprite:    spr_PLAY,
			Fframe:     9,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_PLAY_DIE4,
		},
		161: {
			Fsprite:    spr_PLAY,
			Fframe:     10,
			Ftics:      10,
			Fnextstate: s_PLAY_DIE5,
		},
		162: {
			Fsprite:    spr_PLAY,
			Fframe:     11,
			Ftics:      10,
			Fnextstate: s_PLAY_DIE6,
		},
		163: {
			Fsprite:    spr_PLAY,
			Fframe:     12,
			Ftics:      10,
			Fnextstate: s_PLAY_DIE7,
		},
		164: {
			Fsprite: spr_PLAY,
			Fframe:  13,
			Ftics:   -1,
		},
		165: {
			Fsprite:    spr_PLAY,
			Fframe:     14,
			Ftics:      5,
			Fnextstate: s_PLAY_XDIE2,
		},
		166: {
			Fsprite:    spr_PLAY,
			Fframe:     15,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_XScream),
			Fnextstate: s_PLAY_XDIE3,
		},
		167: {
			Fsprite:    spr_PLAY,
			Fframe:     16,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_PLAY_XDIE4,
		},
		168: {
			Fsprite:    spr_PLAY,
			Fframe:     17,
			Ftics:      5,
			Fnextstate: s_PLAY_XDIE5,
		},
		169: {
			Fsprite:    spr_PLAY,
			Fframe:     18,
			Ftics:      5,
			Fnextstate: s_PLAY_XDIE6,
		},
		170: {
			Fsprite:    spr_PLAY,
			Fframe:     19,
			Ftics:      5,
			Fnextstate: s_PLAY_XDIE7,
		},
		171: {
			Fsprite:    spr_PLAY,
			Fframe:     20,
			Ftics:      5,
			Fnextstate: s_PLAY_XDIE8,
		},
		172: {
			Fsprite:    spr_PLAY,
			Fframe:     21,
			Ftics:      5,
			Fnextstate: s_PLAY_XDIE9,
		},
		173: {
			Fsprite: spr_PLAY,
			Fframe:  22,
			Ftics:   -1,
		},
		174: {
			Fsprite:    spr_POSS,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_POSS_STND2,
		},
		175: {
			Fsprite:    spr_POSS,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_POSS_STND,
		},
		176: {
			Fsprite:    spr_POSS,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_POSS_RUN2,
		},
		177: {
			Fsprite:    spr_POSS,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_POSS_RUN3,
		},
		178: {
			Fsprite:    spr_POSS,
			Fframe:     1,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_POSS_RUN4,
		},
		179: {
			Fsprite:    spr_POSS,
			Fframe:     1,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_POSS_RUN5,
		},
		180: {
			Fsprite:    spr_POSS,
			Fframe:     2,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_POSS_RUN6,
		},
		181: {
			Fsprite:    spr_POSS,
			Fframe:     2,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_POSS_RUN7,
		},
		182: {
			Fsprite:    spr_POSS,
			Fframe:     3,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_POSS_RUN8,
		},
		183: {
			Fsprite:    spr_POSS,
			Fframe:     3,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_POSS_RUN1,
		},
		184: {
			Fsprite:    spr_POSS,
			Fframe:     4,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_POSS_ATK2,
		},
		185: {
			Fsprite:    spr_POSS,
			Fframe:     5,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_PosAttack),
			Fnextstate: s_POSS_ATK3,
		},
		186: {
			Fsprite:    spr_POSS,
			Fframe:     4,
			Ftics:      8,
			Fnextstate: s_POSS_RUN1,
		},
		187: {
			Fsprite:    spr_POSS,
			Fframe:     6,
			Ftics:      3,
			Fnextstate: s_POSS_PAIN2,
		},
		188: {
			Fsprite:    spr_POSS,
			Fframe:     6,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_POSS_RUN1,
		},
		189: {
			Fsprite:    spr_POSS,
			Fframe:     7,
			Ftics:      5,
			Fnextstate: s_POSS_DIE2,
		},
		190: {
			Fsprite:    spr_POSS,
			Fframe:     8,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_POSS_DIE3,
		},
		191: {
			Fsprite:    spr_POSS,
			Fframe:     9,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_POSS_DIE4,
		},
		192: {
			Fsprite:    spr_POSS,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_POSS_DIE5,
		},
		193: {
			Fsprite: spr_POSS,
			Fframe:  11,
			Ftics:   -1,
		},
		194: {
			Fsprite:    spr_POSS,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_POSS_XDIE2,
		},
		195: {
			Fsprite:    spr_POSS,
			Fframe:     13,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_XScream),
			Fnextstate: s_POSS_XDIE3,
		},
		196: {
			Fsprite:    spr_POSS,
			Fframe:     14,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_POSS_XDIE4,
		},
		197: {
			Fsprite:    spr_POSS,
			Fframe:     15,
			Ftics:      5,
			Fnextstate: s_POSS_XDIE5,
		},
		198: {
			Fsprite:    spr_POSS,
			Fframe:     16,
			Ftics:      5,
			Fnextstate: s_POSS_XDIE6,
		},
		199: {
			Fsprite:    spr_POSS,
			Fframe:     17,
			Ftics:      5,
			Fnextstate: s_POSS_XDIE7,
		},
		200: {
			Fsprite:    spr_POSS,
			Fframe:     18,
			Ftics:      5,
			Fnextstate: s_POSS_XDIE8,
		},
		201: {
			Fsprite:    spr_POSS,
			Fframe:     19,
			Ftics:      5,
			Fnextstate: s_POSS_XDIE9,
		},
		202: {
			Fsprite: spr_POSS,
			Fframe:  20,
			Ftics:   -1,
		},
		203: {
			Fsprite:    spr_POSS,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_POSS_RAISE2,
		},
		204: {
			Fsprite:    spr_POSS,
			Fframe:     9,
			Ftics:      5,
			Fnextstate: s_POSS_RAISE3,
		},
		205: {
			Fsprite:    spr_POSS,
			Fframe:     8,
			Ftics:      5,
			Fnextstate: s_POSS_RAISE4,
		},
		206: {
			Fsprite:    spr_POSS,
			Fframe:     7,
			Ftics:      5,
			Fnextstate: s_POSS_RUN1,
		},
		207: {
			Fsprite:    spr_SPOS,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SPOS_STND2,
		},
		208: {
			Fsprite:    spr_SPOS,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SPOS_STND,
		},
		209: {
			Fsprite:    spr_SPOS,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPOS_RUN2,
		},
		210: {
			Fsprite:    spr_SPOS,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPOS_RUN3,
		},
		211: {
			Fsprite:    spr_SPOS,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPOS_RUN4,
		},
		212: {
			Fsprite:    spr_SPOS,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPOS_RUN5,
		},
		213: {
			Fsprite:    spr_SPOS,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPOS_RUN6,
		},
		214: {
			Fsprite:    spr_SPOS,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPOS_RUN7,
		},
		215: {
			Fsprite:    spr_SPOS,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPOS_RUN8,
		},
		216: {
			Fsprite:    spr_SPOS,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPOS_RUN1,
		},
		217: {
			Fsprite:    spr_SPOS,
			Fframe:     4,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SPOS_ATK2,
		},
		218: {
			Fsprite:    spr_SPOS,
			Fframe:     32773,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_SPosAttack),
			Fnextstate: s_SPOS_ATK3,
		},
		219: {
			Fsprite:    spr_SPOS,
			Fframe:     4,
			Ftics:      10,
			Fnextstate: s_SPOS_RUN1,
		},
		220: {
			Fsprite:    spr_SPOS,
			Fframe:     6,
			Ftics:      3,
			Fnextstate: s_SPOS_PAIN2,
		},
		221: {
			Fsprite:    spr_SPOS,
			Fframe:     6,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_SPOS_RUN1,
		},
		222: {
			Fsprite:    spr_SPOS,
			Fframe:     7,
			Ftics:      5,
			Fnextstate: s_SPOS_DIE2,
		},
		223: {
			Fsprite:    spr_SPOS,
			Fframe:     8,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_SPOS_DIE3,
		},
		224: {
			Fsprite:    spr_SPOS,
			Fframe:     9,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_SPOS_DIE4,
		},
		225: {
			Fsprite:    spr_SPOS,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_SPOS_DIE5,
		},
		226: {
			Fsprite: spr_SPOS,
			Fframe:  11,
			Ftics:   -1,
		},
		227: {
			Fsprite:    spr_SPOS,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_SPOS_XDIE2,
		},
		228: {
			Fsprite:    spr_SPOS,
			Fframe:     13,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_XScream),
			Fnextstate: s_SPOS_XDIE3,
		},
		229: {
			Fsprite:    spr_SPOS,
			Fframe:     14,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_SPOS_XDIE4,
		},
		230: {
			Fsprite:    spr_SPOS,
			Fframe:     15,
			Ftics:      5,
			Fnextstate: s_SPOS_XDIE5,
		},
		231: {
			Fsprite:    spr_SPOS,
			Fframe:     16,
			Ftics:      5,
			Fnextstate: s_SPOS_XDIE6,
		},
		232: {
			Fsprite:    spr_SPOS,
			Fframe:     17,
			Ftics:      5,
			Fnextstate: s_SPOS_XDIE7,
		},
		233: {
			Fsprite:    spr_SPOS,
			Fframe:     18,
			Ftics:      5,
			Fnextstate: s_SPOS_XDIE8,
		},
		234: {
			Fsprite:    spr_SPOS,
			Fframe:     19,
			Ftics:      5,
			Fnextstate: s_SPOS_XDIE9,
		},
		235: {
			Fsprite: spr_SPOS,
			Fframe:  20,
			Ftics:   -1,
		},
		236: {
			Fsprite:    spr_SPOS,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_SPOS_RAISE2,
		},
		237: {
			Fsprite:    spr_SPOS,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_SPOS_RAISE3,
		},
		238: {
			Fsprite:    spr_SPOS,
			Fframe:     9,
			Ftics:      5,
			Fnextstate: s_SPOS_RAISE4,
		},
		239: {
			Fsprite:    spr_SPOS,
			Fframe:     8,
			Ftics:      5,
			Fnextstate: s_SPOS_RAISE5,
		},
		240: {
			Fsprite:    spr_SPOS,
			Fframe:     7,
			Ftics:      5,
			Fnextstate: s_SPOS_RUN1,
		},
		241: {
			Fsprite:    spr_VILE,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_VILE_STND2,
		},
		242: {
			Fsprite:    spr_VILE,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_VILE_STND,
		},
		243: {
			Fsprite:    spr_VILE,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN2,
		},
		244: {
			Fsprite:    spr_VILE,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN3,
		},
		245: {
			Fsprite:    spr_VILE,
			Fframe:     1,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN4,
		},
		246: {
			Fsprite:    spr_VILE,
			Fframe:     1,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN5,
		},
		247: {
			Fsprite:    spr_VILE,
			Fframe:     2,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN6,
		},
		248: {
			Fsprite:    spr_VILE,
			Fframe:     2,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN7,
		},
		249: {
			Fsprite:    spr_VILE,
			Fframe:     3,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN8,
		},
		250: {
			Fsprite:    spr_VILE,
			Fframe:     3,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN9,
		},
		251: {
			Fsprite:    spr_VILE,
			Fframe:     4,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN10,
		},
		252: {
			Fsprite:    spr_VILE,
			Fframe:     4,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN11,
		},
		253: {
			Fsprite:    spr_VILE,
			Fframe:     5,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN12,
		},
		254: {
			Fsprite:    spr_VILE,
			Fframe:     5,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_VileChase),
			Fnextstate: s_VILE_RUN1,
		},
		255: {
			Fsprite:    spr_VILE,
			Fframe:     32774,
			Faction:    mobjFuncToAction(a_VileStart),
			Fnextstate: s_VILE_ATK2,
		},
		256: {
			Fsprite:    spr_VILE,
			Fframe:     32774,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_VILE_ATK3,
		},
		257: {
			Fsprite:    spr_VILE,
			Fframe:     32775,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_VileTarget),
			Fnextstate: s_VILE_ATK4,
		},
		258: {
			Fsprite:    spr_VILE,
			Fframe:     32776,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_VILE_ATK5,
		},
		259: {
			Fsprite:    spr_VILE,
			Fframe:     32777,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_VILE_ATK6,
		},
		260: {
			Fsprite:    spr_VILE,
			Fframe:     32778,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_VILE_ATK7,
		},
		261: {
			Fsprite:    spr_VILE,
			Fframe:     32779,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_VILE_ATK8,
		},
		262: {
			Fsprite:    spr_VILE,
			Fframe:     32780,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_VILE_ATK9,
		},
		263: {
			Fsprite:    spr_VILE,
			Fframe:     32781,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_VILE_ATK10,
		},
		264: {
			Fsprite:    spr_VILE,
			Fframe:     32782,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_VileAttack),
			Fnextstate: s_VILE_ATK11,
		},
		265: {
			Fsprite:    spr_VILE,
			Fframe:     32783,
			Ftics:      20,
			Fnextstate: s_VILE_RUN1,
		},
		266: {
			Fsprite:    spr_VILE,
			Fframe:     32794,
			Ftics:      10,
			Fnextstate: s_VILE_HEAL2,
		},
		267: {
			Fsprite:    spr_VILE,
			Fframe:     32795,
			Ftics:      10,
			Fnextstate: s_VILE_HEAL3,
		},
		268: {
			Fsprite:    spr_VILE,
			Fframe:     32796,
			Ftics:      10,
			Fnextstate: s_VILE_RUN1,
		},
		269: {
			Fsprite:    spr_VILE,
			Fframe:     16,
			Ftics:      5,
			Fnextstate: s_VILE_PAIN2,
		},
		270: {
			Fsprite:    spr_VILE,
			Fframe:     16,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_VILE_RUN1,
		},
		271: {
			Fsprite:    spr_VILE,
			Fframe:     16,
			Ftics:      7,
			Fnextstate: s_VILE_DIE2,
		},
		272: {
			Fsprite:    spr_VILE,
			Fframe:     17,
			Ftics:      7,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_VILE_DIE3,
		},
		273: {
			Fsprite:    spr_VILE,
			Fframe:     18,
			Ftics:      7,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_VILE_DIE4,
		},
		274: {
			Fsprite:    spr_VILE,
			Fframe:     19,
			Ftics:      7,
			Fnextstate: s_VILE_DIE5,
		},
		275: {
			Fsprite:    spr_VILE,
			Fframe:     20,
			Ftics:      7,
			Fnextstate: s_VILE_DIE6,
		},
		276: {
			Fsprite:    spr_VILE,
			Fframe:     21,
			Ftics:      7,
			Fnextstate: s_VILE_DIE7,
		},
		277: {
			Fsprite:    spr_VILE,
			Fframe:     22,
			Ftics:      7,
			Fnextstate: s_VILE_DIE8,
		},
		278: {
			Fsprite:    spr_VILE,
			Fframe:     23,
			Ftics:      5,
			Fnextstate: s_VILE_DIE9,
		},
		279: {
			Fsprite:    spr_VILE,
			Fframe:     24,
			Ftics:      5,
			Fnextstate: s_VILE_DIE10,
		},
		280: {
			Fsprite: spr_VILE,
			Fframe:  25,
			Ftics:   -1,
		},
		281: {
			Fsprite:    spr_FIRE,
			Fframe:     32768,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_StartFire),
			Fnextstate: s_FIRE2,
		},
		282: {
			Fsprite:    spr_FIRE,
			Fframe:     32769,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE3,
		},
		283: {
			Fsprite:    spr_FIRE,
			Fframe:     32768,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE4,
		},
		284: {
			Fsprite:    spr_FIRE,
			Fframe:     32769,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE5,
		},
		285: {
			Fsprite:    spr_FIRE,
			Fframe:     32770,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_FireCrackle),
			Fnextstate: s_FIRE6,
		},
		286: {
			Fsprite:    spr_FIRE,
			Fframe:     32769,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE7,
		},
		287: {
			Fsprite:    spr_FIRE,
			Fframe:     32770,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE8,
		},
		288: {
			Fsprite:    spr_FIRE,
			Fframe:     32769,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE9,
		},
		289: {
			Fsprite:    spr_FIRE,
			Fframe:     32770,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE10,
		},
		290: {
			Fsprite:    spr_FIRE,
			Fframe:     32771,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE11,
		},
		291: {
			Fsprite:    spr_FIRE,
			Fframe:     32770,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE12,
		},
		292: {
			Fsprite:    spr_FIRE,
			Fframe:     32771,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE13,
		},
		293: {
			Fsprite:    spr_FIRE,
			Fframe:     32770,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE14,
		},
		294: {
			Fsprite:    spr_FIRE,
			Fframe:     32771,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE15,
		},
		295: {
			Fsprite:    spr_FIRE,
			Fframe:     32772,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE16,
		},
		296: {
			Fsprite:    spr_FIRE,
			Fframe:     32771,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE17,
		},
		297: {
			Fsprite:    spr_FIRE,
			Fframe:     32772,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE18,
		},
		298: {
			Fsprite:    spr_FIRE,
			Fframe:     32771,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE19,
		},
		299: {
			Fsprite:    spr_FIRE,
			Fframe:     32772,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_FireCrackle),
			Fnextstate: s_FIRE20,
		},
		300: {
			Fsprite:    spr_FIRE,
			Fframe:     32773,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE21,
		},
		301: {
			Fsprite:    spr_FIRE,
			Fframe:     32772,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE22,
		},
		302: {
			Fsprite:    spr_FIRE,
			Fframe:     32773,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE23,
		},
		303: {
			Fsprite:    spr_FIRE,
			Fframe:     32772,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE24,
		},
		304: {
			Fsprite:    spr_FIRE,
			Fframe:     32773,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE25,
		},
		305: {
			Fsprite:    spr_FIRE,
			Fframe:     32774,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE26,
		},
		306: {
			Fsprite:    spr_FIRE,
			Fframe:     32775,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE27,
		},
		307: {
			Fsprite:    spr_FIRE,
			Fframe:     32774,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE28,
		},
		308: {
			Fsprite:    spr_FIRE,
			Fframe:     32775,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE29,
		},
		309: {
			Fsprite:    spr_FIRE,
			Fframe:     32774,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_FIRE30,
		},
		310: {
			Fsprite: spr_FIRE,
			Fframe:  32775,
			Ftics:   2,
			Faction: mobjFuncToAction(a_Fire),
		},
		311: {
			Fsprite:    spr_PUFF,
			Fframe:     1,
			Ftics:      4,
			Fnextstate: s_SMOKE2,
		},
		312: {
			Fsprite:    spr_PUFF,
			Fframe:     2,
			Ftics:      4,
			Fnextstate: s_SMOKE3,
		},
		313: {
			Fsprite:    spr_PUFF,
			Fframe:     1,
			Ftics:      4,
			Fnextstate: s_SMOKE4,
		},
		314: {
			Fsprite:    spr_PUFF,
			Fframe:     2,
			Ftics:      4,
			Fnextstate: s_SMOKE5,
		},
		315: {
			Fsprite: spr_PUFF,
			Fframe:  3,
			Ftics:   4,
		},
		316: {
			Fsprite:    spr_FATB,
			Fframe:     32768,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Tracer),
			Fnextstate: s_TRACER2,
		},
		317: {
			Fsprite:    spr_FATB,
			Fframe:     32769,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Tracer),
			Fnextstate: s_TRACER,
		},
		318: {
			Fsprite:    spr_FBXP,
			Fframe:     32768,
			Ftics:      8,
			Fnextstate: s_TRACEEXP2,
		},
		319: {
			Fsprite:    spr_FBXP,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_TRACEEXP3,
		},
		320: {
			Fsprite: spr_FBXP,
			Fframe:  32770,
			Ftics:   4,
		},
		321: {
			Fsprite:    spr_SKEL,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SKEL_STND2,
		},
		322: {
			Fsprite:    spr_SKEL,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SKEL_STND,
		},
		323: {
			Fsprite:    spr_SKEL,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN2,
		},
		324: {
			Fsprite:    spr_SKEL,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN3,
		},
		325: {
			Fsprite:    spr_SKEL,
			Fframe:     1,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN4,
		},
		326: {
			Fsprite:    spr_SKEL,
			Fframe:     1,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN5,
		},
		327: {
			Fsprite:    spr_SKEL,
			Fframe:     2,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN6,
		},
		328: {
			Fsprite:    spr_SKEL,
			Fframe:     2,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN7,
		},
		329: {
			Fsprite:    spr_SKEL,
			Fframe:     3,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN8,
		},
		330: {
			Fsprite:    spr_SKEL,
			Fframe:     3,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN9,
		},
		331: {
			Fsprite:    spr_SKEL,
			Fframe:     4,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN10,
		},
		332: {
			Fsprite:    spr_SKEL,
			Fframe:     4,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN11,
		},
		333: {
			Fsprite:    spr_SKEL,
			Fframe:     5,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN12,
		},
		334: {
			Fsprite:    spr_SKEL,
			Fframe:     5,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKEL_RUN1,
		},
		335: {
			Fsprite:    spr_SKEL,
			Fframe:     6,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SKEL_FIST2,
		},
		336: {
			Fsprite:    spr_SKEL,
			Fframe:     6,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_SkelWhoosh),
			Fnextstate: s_SKEL_FIST3,
		},
		337: {
			Fsprite:    spr_SKEL,
			Fframe:     7,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SKEL_FIST4,
		},
		338: {
			Fsprite:    spr_SKEL,
			Fframe:     8,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_SkelFist),
			Fnextstate: s_SKEL_RUN1,
		},
		339: {
			Fsprite:    spr_SKEL,
			Fframe:     32777,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SKEL_MISS2,
		},
		340: {
			Fsprite:    spr_SKEL,
			Fframe:     32777,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SKEL_MISS3,
		},
		341: {
			Fsprite:    spr_SKEL,
			Fframe:     10,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_SkelMissile),
			Fnextstate: s_SKEL_MISS4,
		},
		342: {
			Fsprite:    spr_SKEL,
			Fframe:     10,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SKEL_RUN1,
		},
		343: {
			Fsprite:    spr_SKEL,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_SKEL_PAIN2,
		},
		344: {
			Fsprite:    spr_SKEL,
			Fframe:     11,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_SKEL_RUN1,
		},
		345: {
			Fsprite:    spr_SKEL,
			Fframe:     11,
			Ftics:      7,
			Fnextstate: s_SKEL_DIE2,
		},
		346: {
			Fsprite:    spr_SKEL,
			Fframe:     12,
			Ftics:      7,
			Fnextstate: s_SKEL_DIE3,
		},
		347: {
			Fsprite:    spr_SKEL,
			Fframe:     13,
			Ftics:      7,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_SKEL_DIE4,
		},
		348: {
			Fsprite:    spr_SKEL,
			Fframe:     14,
			Ftics:      7,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_SKEL_DIE5,
		},
		349: {
			Fsprite:    spr_SKEL,
			Fframe:     15,
			Ftics:      7,
			Fnextstate: s_SKEL_DIE6,
		},
		350: {
			Fsprite: spr_SKEL,
			Fframe:  16,
			Ftics:   -1,
		},
		351: {
			Fsprite:    spr_SKEL,
			Fframe:     16,
			Ftics:      5,
			Fnextstate: s_SKEL_RAISE2,
		},
		352: {
			Fsprite:    spr_SKEL,
			Fframe:     15,
			Ftics:      5,
			Fnextstate: s_SKEL_RAISE3,
		},
		353: {
			Fsprite:    spr_SKEL,
			Fframe:     14,
			Ftics:      5,
			Fnextstate: s_SKEL_RAISE4,
		},
		354: {
			Fsprite:    spr_SKEL,
			Fframe:     13,
			Ftics:      5,
			Fnextstate: s_SKEL_RAISE5,
		},
		355: {
			Fsprite:    spr_SKEL,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_SKEL_RAISE6,
		},
		356: {
			Fsprite:    spr_SKEL,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_SKEL_RUN1,
		},
		357: {
			Fsprite:    spr_MANF,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_FATSHOT2,
		},
		358: {
			Fsprite:    spr_MANF,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_FATSHOT1,
		},
		359: {
			Fsprite:    spr_MISL,
			Fframe:     32769,
			Ftics:      8,
			Fnextstate: s_FATSHOTX2,
		},
		360: {
			Fsprite:    spr_MISL,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_FATSHOTX3,
		},
		361: {
			Fsprite: spr_MISL,
			Fframe:  32771,
			Ftics:   4,
		},
		362: {
			Fsprite:    spr_FATT,
			Ftics:      15,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_FATT_STND2,
		},
		363: {
			Fsprite:    spr_FATT,
			Fframe:     1,
			Ftics:      15,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_FATT_STND,
		},
		364: {
			Fsprite:    spr_FATT,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN2,
		},
		365: {
			Fsprite:    spr_FATT,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN3,
		},
		366: {
			Fsprite:    spr_FATT,
			Fframe:     1,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN4,
		},
		367: {
			Fsprite:    spr_FATT,
			Fframe:     1,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN5,
		},
		368: {
			Fsprite:    spr_FATT,
			Fframe:     2,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN6,
		},
		369: {
			Fsprite:    spr_FATT,
			Fframe:     2,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN7,
		},
		370: {
			Fsprite:    spr_FATT,
			Fframe:     3,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN8,
		},
		371: {
			Fsprite:    spr_FATT,
			Fframe:     3,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN9,
		},
		372: {
			Fsprite:    spr_FATT,
			Fframe:     4,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN10,
		},
		373: {
			Fsprite:    spr_FATT,
			Fframe:     4,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN11,
		},
		374: {
			Fsprite:    spr_FATT,
			Fframe:     5,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN12,
		},
		375: {
			Fsprite:    spr_FATT,
			Fframe:     5,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_FATT_RUN1,
		},
		376: {
			Fsprite:    spr_FATT,
			Fframe:     6,
			Ftics:      20,
			Faction:    mobjFuncToAction(a_FatRaise),
			Fnextstate: s_FATT_ATK2,
		},
		377: {
			Fsprite:    spr_FATT,
			Fframe:     32775,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FatAttack1),
			Fnextstate: s_FATT_ATK3,
		},
		378: {
			Fsprite:    spr_FATT,
			Fframe:     8,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_FATT_ATK4,
		},
		379: {
			Fsprite:    spr_FATT,
			Fframe:     6,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_FATT_ATK5,
		},
		380: {
			Fsprite:    spr_FATT,
			Fframe:     32775,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FatAttack2),
			Fnextstate: s_FATT_ATK6,
		},
		381: {
			Fsprite:    spr_FATT,
			Fframe:     8,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_FATT_ATK7,
		},
		382: {
			Fsprite:    spr_FATT,
			Fframe:     6,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_FATT_ATK8,
		},
		383: {
			Fsprite:    spr_FATT,
			Fframe:     32775,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FatAttack3),
			Fnextstate: s_FATT_ATK9,
		},
		384: {
			Fsprite:    spr_FATT,
			Fframe:     8,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_FATT_ATK10,
		},
		385: {
			Fsprite:    spr_FATT,
			Fframe:     6,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_FATT_RUN1,
		},
		386: {
			Fsprite:    spr_FATT,
			Fframe:     9,
			Ftics:      3,
			Fnextstate: s_FATT_PAIN2,
		},
		387: {
			Fsprite:    spr_FATT,
			Fframe:     9,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_FATT_RUN1,
		},
		388: {
			Fsprite:    spr_FATT,
			Fframe:     10,
			Ftics:      6,
			Fnextstate: s_FATT_DIE2,
		},
		389: {
			Fsprite:    spr_FATT,
			Fframe:     11,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_FATT_DIE3,
		},
		390: {
			Fsprite:    spr_FATT,
			Fframe:     12,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_FATT_DIE4,
		},
		391: {
			Fsprite:    spr_FATT,
			Fframe:     13,
			Ftics:      6,
			Fnextstate: s_FATT_DIE5,
		},
		392: {
			Fsprite:    spr_FATT,
			Fframe:     14,
			Ftics:      6,
			Fnextstate: s_FATT_DIE6,
		},
		393: {
			Fsprite:    spr_FATT,
			Fframe:     15,
			Ftics:      6,
			Fnextstate: s_FATT_DIE7,
		},
		394: {
			Fsprite:    spr_FATT,
			Fframe:     16,
			Ftics:      6,
			Fnextstate: s_FATT_DIE8,
		},
		395: {
			Fsprite:    spr_FATT,
			Fframe:     17,
			Ftics:      6,
			Fnextstate: s_FATT_DIE9,
		},
		396: {
			Fsprite:    spr_FATT,
			Fframe:     18,
			Ftics:      6,
			Fnextstate: s_FATT_DIE10,
		},
		397: {
			Fsprite: spr_FATT,
			Fframe:  19,
			Ftics:   -1,
			Faction: mobjFuncToAction(a_BossDeath),
		},
		398: {
			Fsprite:    spr_FATT,
			Fframe:     17,
			Ftics:      5,
			Fnextstate: s_FATT_RAISE2,
		},
		399: {
			Fsprite:    spr_FATT,
			Fframe:     16,
			Ftics:      5,
			Fnextstate: s_FATT_RAISE3,
		},
		400: {
			Fsprite:    spr_FATT,
			Fframe:     15,
			Ftics:      5,
			Fnextstate: s_FATT_RAISE4,
		},
		401: {
			Fsprite:    spr_FATT,
			Fframe:     14,
			Ftics:      5,
			Fnextstate: s_FATT_RAISE5,
		},
		402: {
			Fsprite:    spr_FATT,
			Fframe:     13,
			Ftics:      5,
			Fnextstate: s_FATT_RAISE6,
		},
		403: {
			Fsprite:    spr_FATT,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_FATT_RAISE7,
		},
		404: {
			Fsprite:    spr_FATT,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_FATT_RAISE8,
		},
		405: {
			Fsprite:    spr_FATT,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_FATT_RUN1,
		},
		406: {
			Fsprite:    spr_CPOS,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_CPOS_STND2,
		},
		407: {
			Fsprite:    spr_CPOS,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_CPOS_STND,
		},
		408: {
			Fsprite:    spr_CPOS,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CPOS_RUN2,
		},
		409: {
			Fsprite:    spr_CPOS,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CPOS_RUN3,
		},
		410: {
			Fsprite:    spr_CPOS,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CPOS_RUN4,
		},
		411: {
			Fsprite:    spr_CPOS,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CPOS_RUN5,
		},
		412: {
			Fsprite:    spr_CPOS,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CPOS_RUN6,
		},
		413: {
			Fsprite:    spr_CPOS,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CPOS_RUN7,
		},
		414: {
			Fsprite:    spr_CPOS,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CPOS_RUN8,
		},
		415: {
			Fsprite:    spr_CPOS,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CPOS_RUN1,
		},
		416: {
			Fsprite:    spr_CPOS,
			Fframe:     4,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_CPOS_ATK2,
		},
		417: {
			Fsprite:    spr_CPOS,
			Fframe:     32773,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_CPosAttack),
			Fnextstate: s_CPOS_ATK3,
		},
		418: {
			Fsprite:    spr_CPOS,
			Fframe:     32772,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_CPosAttack),
			Fnextstate: s_CPOS_ATK4,
		},
		419: {
			Fsprite:    spr_CPOS,
			Fframe:     5,
			Ftics:      1,
			Faction:    mobjFuncToAction(a_CPosRefire),
			Fnextstate: s_CPOS_ATK2,
		},
		420: {
			Fsprite:    spr_CPOS,
			Fframe:     6,
			Ftics:      3,
			Fnextstate: s_CPOS_PAIN2,
		},
		421: {
			Fsprite:    spr_CPOS,
			Fframe:     6,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_CPOS_RUN1,
		},
		422: {
			Fsprite:    spr_CPOS,
			Fframe:     7,
			Ftics:      5,
			Fnextstate: s_CPOS_DIE2,
		},
		423: {
			Fsprite:    spr_CPOS,
			Fframe:     8,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_CPOS_DIE3,
		},
		424: {
			Fsprite:    spr_CPOS,
			Fframe:     9,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_CPOS_DIE4,
		},
		425: {
			Fsprite:    spr_CPOS,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_CPOS_DIE5,
		},
		426: {
			Fsprite:    spr_CPOS,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_CPOS_DIE6,
		},
		427: {
			Fsprite:    spr_CPOS,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_CPOS_DIE7,
		},
		428: {
			Fsprite: spr_CPOS,
			Fframe:  13,
			Ftics:   -1,
		},
		429: {
			Fsprite:    spr_CPOS,
			Fframe:     14,
			Ftics:      5,
			Fnextstate: s_CPOS_XDIE2,
		},
		430: {
			Fsprite:    spr_CPOS,
			Fframe:     15,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_XScream),
			Fnextstate: s_CPOS_XDIE3,
		},
		431: {
			Fsprite:    spr_CPOS,
			Fframe:     16,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_CPOS_XDIE4,
		},
		432: {
			Fsprite:    spr_CPOS,
			Fframe:     17,
			Ftics:      5,
			Fnextstate: s_CPOS_XDIE5,
		},
		433: {
			Fsprite:    spr_CPOS,
			Fframe:     18,
			Ftics:      5,
			Fnextstate: s_CPOS_XDIE6,
		},
		434: {
			Fsprite: spr_CPOS,
			Fframe:  19,
			Ftics:   -1,
		},
		435: {
			Fsprite:    spr_CPOS,
			Fframe:     13,
			Ftics:      5,
			Fnextstate: s_CPOS_RAISE2,
		},
		436: {
			Fsprite:    spr_CPOS,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_CPOS_RAISE3,
		},
		437: {
			Fsprite:    spr_CPOS,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_CPOS_RAISE4,
		},
		438: {
			Fsprite:    spr_CPOS,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_CPOS_RAISE5,
		},
		439: {
			Fsprite:    spr_CPOS,
			Fframe:     9,
			Ftics:      5,
			Fnextstate: s_CPOS_RAISE6,
		},
		440: {
			Fsprite:    spr_CPOS,
			Fframe:     8,
			Ftics:      5,
			Fnextstate: s_CPOS_RAISE7,
		},
		441: {
			Fsprite:    spr_CPOS,
			Fframe:     7,
			Ftics:      5,
			Fnextstate: s_CPOS_RUN1,
		},
		442: {
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_TROO_STND2,
		},
		443: {
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_TROO_STND,
		},
		444: {
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_TROO_RUN2,
		},
		445: {
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_TROO_RUN3,
		},
		446: {
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_TROO_RUN4,
		},
		447: {
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_TROO_RUN5,
		},
		448: {
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_TROO_RUN6,
		},
		449: {
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_TROO_RUN7,
		},
		450: {
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_TROO_RUN8,
		},
		451: {
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_TROO_RUN1,
		},
		452: {
			Fframe:     4,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_TROO_ATK2,
		},
		453: {
			Fframe:     5,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_TROO_ATK3,
		},
		454: {
			Fframe:     6,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_TroopAttack),
			Fnextstate: s_TROO_RUN1,
		},
		455: {
			Fframe:     7,
			Ftics:      2,
			Fnextstate: s_TROO_PAIN2,
		},
		456: {
			Fframe:     7,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_TROO_RUN1,
		},
		457: {
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_TROO_DIE2,
		},
		458: {
			Fframe:     9,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_TROO_DIE3,
		},
		459: {
			Fframe:     10,
			Ftics:      6,
			Fnextstate: s_TROO_DIE4,
		},
		460: {
			Fframe:     11,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_TROO_DIE5,
		},
		461: {
			Fframe: 12,
			Ftics:  -1,
		},
		462: {
			Fframe:     13,
			Ftics:      5,
			Fnextstate: s_TROO_XDIE2,
		},
		463: {
			Fframe:     14,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_XScream),
			Fnextstate: s_TROO_XDIE3,
		},
		464: {
			Fframe:     15,
			Ftics:      5,
			Fnextstate: s_TROO_XDIE4,
		},
		465: {
			Fframe:     16,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_TROO_XDIE5,
		},
		466: {
			Fframe:     17,
			Ftics:      5,
			Fnextstate: s_TROO_XDIE6,
		},
		467: {
			Fframe:     18,
			Ftics:      5,
			Fnextstate: s_TROO_XDIE7,
		},
		468: {
			Fframe:     19,
			Ftics:      5,
			Fnextstate: s_TROO_XDIE8,
		},
		469: {
			Fframe: 20,
			Ftics:  -1,
		},
		470: {
			Fframe:     12,
			Ftics:      8,
			Fnextstate: s_TROO_RAISE2,
		},
		471: {
			Fframe:     11,
			Ftics:      8,
			Fnextstate: s_TROO_RAISE3,
		},
		472: {
			Fframe:     10,
			Ftics:      6,
			Fnextstate: s_TROO_RAISE4,
		},
		473: {
			Fframe:     9,
			Ftics:      6,
			Fnextstate: s_TROO_RAISE5,
		},
		474: {
			Fframe:     8,
			Ftics:      6,
			Fnextstate: s_TROO_RUN1,
		},
		475: {
			Fsprite:    spr_SARG,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SARG_STND2,
		},
		476: {
			Fsprite:    spr_SARG,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SARG_STND,
		},
		477: {
			Fsprite:    spr_SARG,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SARG_RUN2,
		},
		478: {
			Fsprite:    spr_SARG,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SARG_RUN3,
		},
		479: {
			Fsprite:    spr_SARG,
			Fframe:     1,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SARG_RUN4,
		},
		480: {
			Fsprite:    spr_SARG,
			Fframe:     1,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SARG_RUN5,
		},
		481: {
			Fsprite:    spr_SARG,
			Fframe:     2,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SARG_RUN6,
		},
		482: {
			Fsprite:    spr_SARG,
			Fframe:     2,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SARG_RUN7,
		},
		483: {
			Fsprite:    spr_SARG,
			Fframe:     3,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SARG_RUN8,
		},
		484: {
			Fsprite:    spr_SARG,
			Fframe:     3,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SARG_RUN1,
		},
		485: {
			Fsprite:    spr_SARG,
			Fframe:     4,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SARG_ATK2,
		},
		486: {
			Fsprite:    spr_SARG,
			Fframe:     5,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SARG_ATK3,
		},
		487: {
			Fsprite:    spr_SARG,
			Fframe:     6,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_SargAttack),
			Fnextstate: s_SARG_RUN1,
		},
		488: {
			Fsprite:    spr_SARG,
			Fframe:     7,
			Ftics:      2,
			Fnextstate: s_SARG_PAIN2,
		},
		489: {
			Fsprite:    spr_SARG,
			Fframe:     7,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_SARG_RUN1,
		},
		490: {
			Fsprite:    spr_SARG,
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_SARG_DIE2,
		},
		491: {
			Fsprite:    spr_SARG,
			Fframe:     9,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_SARG_DIE3,
		},
		492: {
			Fsprite:    spr_SARG,
			Fframe:     10,
			Ftics:      4,
			Fnextstate: s_SARG_DIE4,
		},
		493: {
			Fsprite:    spr_SARG,
			Fframe:     11,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_SARG_DIE5,
		},
		494: {
			Fsprite:    spr_SARG,
			Fframe:     12,
			Ftics:      4,
			Fnextstate: s_SARG_DIE6,
		},
		495: {
			Fsprite: spr_SARG,
			Fframe:  13,
			Ftics:   -1,
		},
		496: {
			Fsprite:    spr_SARG,
			Fframe:     13,
			Ftics:      5,
			Fnextstate: s_SARG_RAISE2,
		},
		497: {
			Fsprite:    spr_SARG,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_SARG_RAISE3,
		},
		498: {
			Fsprite:    spr_SARG,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_SARG_RAISE4,
		},
		499: {
			Fsprite:    spr_SARG,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_SARG_RAISE5,
		},
		500: {
			Fsprite:    spr_SARG,
			Fframe:     9,
			Ftics:      5,
			Fnextstate: s_SARG_RAISE6,
		},
		501: {
			Fsprite:    spr_SARG,
			Fframe:     8,
			Ftics:      5,
			Fnextstate: s_SARG_RUN1,
		},
		502: {
			Fsprite:    spr_HEAD,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_HEAD_STND,
		},
		503: {
			Fsprite:    spr_HEAD,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_HEAD_RUN1,
		},
		504: {
			Fsprite:    spr_HEAD,
			Fframe:     1,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_HEAD_ATK2,
		},
		505: {
			Fsprite:    spr_HEAD,
			Fframe:     2,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_HEAD_ATK3,
		},
		506: {
			Fsprite:    spr_HEAD,
			Fframe:     32771,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_HeadAttack),
			Fnextstate: s_HEAD_RUN1,
		},
		507: {
			Fsprite:    spr_HEAD,
			Fframe:     4,
			Ftics:      3,
			Fnextstate: s_HEAD_PAIN2,
		},
		508: {
			Fsprite:    spr_HEAD,
			Fframe:     4,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_HEAD_PAIN3,
		},
		509: {
			Fsprite:    spr_HEAD,
			Fframe:     5,
			Ftics:      6,
			Fnextstate: s_HEAD_RUN1,
		},
		510: {
			Fsprite:    spr_HEAD,
			Fframe:     6,
			Ftics:      8,
			Fnextstate: s_HEAD_DIE2,
		},
		511: {
			Fsprite:    spr_HEAD,
			Fframe:     7,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_HEAD_DIE3,
		},
		512: {
			Fsprite:    spr_HEAD,
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_HEAD_DIE4,
		},
		513: {
			Fsprite:    spr_HEAD,
			Fframe:     9,
			Ftics:      8,
			Fnextstate: s_HEAD_DIE5,
		},
		514: {
			Fsprite:    spr_HEAD,
			Fframe:     10,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_HEAD_DIE6,
		},
		515: {
			Fsprite: spr_HEAD,
			Fframe:  11,
			Ftics:   -1,
		},
		516: {
			Fsprite:    spr_HEAD,
			Fframe:     11,
			Ftics:      8,
			Fnextstate: s_HEAD_RAISE2,
		},
		517: {
			Fsprite:    spr_HEAD,
			Fframe:     10,
			Ftics:      8,
			Fnextstate: s_HEAD_RAISE3,
		},
		518: {
			Fsprite:    spr_HEAD,
			Fframe:     9,
			Ftics:      8,
			Fnextstate: s_HEAD_RAISE4,
		},
		519: {
			Fsprite:    spr_HEAD,
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_HEAD_RAISE5,
		},
		520: {
			Fsprite:    spr_HEAD,
			Fframe:     7,
			Ftics:      8,
			Fnextstate: s_HEAD_RAISE6,
		},
		521: {
			Fsprite:    spr_HEAD,
			Fframe:     6,
			Ftics:      8,
			Fnextstate: s_HEAD_RUN1,
		},
		522: {
			Fsprite:    spr_BAL7,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_BRBALL2,
		},
		523: {
			Fsprite:    spr_BAL7,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_BRBALL1,
		},
		524: {
			Fsprite:    spr_BAL7,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_BRBALLX2,
		},
		525: {
			Fsprite:    spr_BAL7,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_BRBALLX3,
		},
		526: {
			Fsprite: spr_BAL7,
			Fframe:  32772,
			Ftics:   6,
		},
		527: {
			Fsprite:    spr_BOSS,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_BOSS_STND2,
		},
		528: {
			Fsprite:    spr_BOSS,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_BOSS_STND,
		},
		529: {
			Fsprite:    spr_BOSS,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOSS_RUN2,
		},
		530: {
			Fsprite:    spr_BOSS,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOSS_RUN3,
		},
		531: {
			Fsprite:    spr_BOSS,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOSS_RUN4,
		},
		532: {
			Fsprite:    spr_BOSS,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOSS_RUN5,
		},
		533: {
			Fsprite:    spr_BOSS,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOSS_RUN6,
		},
		534: {
			Fsprite:    spr_BOSS,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOSS_RUN7,
		},
		535: {
			Fsprite:    spr_BOSS,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOSS_RUN8,
		},
		536: {
			Fsprite:    spr_BOSS,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOSS_RUN1,
		},
		537: {
			Fsprite:    spr_BOSS,
			Fframe:     4,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_BOSS_ATK2,
		},
		538: {
			Fsprite:    spr_BOSS,
			Fframe:     5,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_BOSS_ATK3,
		},
		539: {
			Fsprite:    spr_BOSS,
			Fframe:     6,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_BruisAttack),
			Fnextstate: s_BOSS_RUN1,
		},
		540: {
			Fsprite:    spr_BOSS,
			Fframe:     7,
			Ftics:      2,
			Fnextstate: s_BOSS_PAIN2,
		},
		541: {
			Fsprite:    spr_BOSS,
			Fframe:     7,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_BOSS_RUN1,
		},
		542: {
			Fsprite:    spr_BOSS,
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_BOSS_DIE2,
		},
		543: {
			Fsprite:    spr_BOSS,
			Fframe:     9,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_BOSS_DIE3,
		},
		544: {
			Fsprite:    spr_BOSS,
			Fframe:     10,
			Ftics:      8,
			Fnextstate: s_BOSS_DIE4,
		},
		545: {
			Fsprite:    spr_BOSS,
			Fframe:     11,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_BOSS_DIE5,
		},
		546: {
			Fsprite:    spr_BOSS,
			Fframe:     12,
			Ftics:      8,
			Fnextstate: s_BOSS_DIE6,
		},
		547: {
			Fsprite:    spr_BOSS,
			Fframe:     13,
			Ftics:      8,
			Fnextstate: s_BOSS_DIE7,
		},
		548: {
			Fsprite: spr_BOSS,
			Fframe:  14,
			Ftics:   -1,
			Faction: mobjFuncToAction(a_BossDeath),
		},
		549: {
			Fsprite:    spr_BOSS,
			Fframe:     14,
			Ftics:      8,
			Fnextstate: s_BOSS_RAISE2,
		},
		550: {
			Fsprite:    spr_BOSS,
			Fframe:     13,
			Ftics:      8,
			Fnextstate: s_BOSS_RAISE3,
		},
		551: {
			Fsprite:    spr_BOSS,
			Fframe:     12,
			Ftics:      8,
			Fnextstate: s_BOSS_RAISE4,
		},
		552: {
			Fsprite:    spr_BOSS,
			Fframe:     11,
			Ftics:      8,
			Fnextstate: s_BOSS_RAISE5,
		},
		553: {
			Fsprite:    spr_BOSS,
			Fframe:     10,
			Ftics:      8,
			Fnextstate: s_BOSS_RAISE6,
		},
		554: {
			Fsprite:    spr_BOSS,
			Fframe:     9,
			Ftics:      8,
			Fnextstate: s_BOSS_RAISE7,
		},
		555: {
			Fsprite:    spr_BOSS,
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_BOSS_RUN1,
		},
		556: {
			Fsprite:    spr_BOS2,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_BOS2_STND2,
		},
		557: {
			Fsprite:    spr_BOS2,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_BOS2_STND,
		},
		558: {
			Fsprite:    spr_BOS2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOS2_RUN2,
		},
		559: {
			Fsprite:    spr_BOS2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOS2_RUN3,
		},
		560: {
			Fsprite:    spr_BOS2,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOS2_RUN4,
		},
		561: {
			Fsprite:    spr_BOS2,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOS2_RUN5,
		},
		562: {
			Fsprite:    spr_BOS2,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOS2_RUN6,
		},
		563: {
			Fsprite:    spr_BOS2,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOS2_RUN7,
		},
		564: {
			Fsprite:    spr_BOS2,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOS2_RUN8,
		},
		565: {
			Fsprite:    spr_BOS2,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BOS2_RUN1,
		},
		566: {
			Fsprite:    spr_BOS2,
			Fframe:     4,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_BOS2_ATK2,
		},
		567: {
			Fsprite:    spr_BOS2,
			Fframe:     5,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_BOS2_ATK3,
		},
		568: {
			Fsprite:    spr_BOS2,
			Fframe:     6,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_BruisAttack),
			Fnextstate: s_BOS2_RUN1,
		},
		569: {
			Fsprite:    spr_BOS2,
			Fframe:     7,
			Ftics:      2,
			Fnextstate: s_BOS2_PAIN2,
		},
		570: {
			Fsprite:    spr_BOS2,
			Fframe:     7,
			Ftics:      2,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_BOS2_RUN1,
		},
		571: {
			Fsprite:    spr_BOS2,
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_BOS2_DIE2,
		},
		572: {
			Fsprite:    spr_BOS2,
			Fframe:     9,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_BOS2_DIE3,
		},
		573: {
			Fsprite:    spr_BOS2,
			Fframe:     10,
			Ftics:      8,
			Fnextstate: s_BOS2_DIE4,
		},
		574: {
			Fsprite:    spr_BOS2,
			Fframe:     11,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_BOS2_DIE5,
		},
		575: {
			Fsprite:    spr_BOS2,
			Fframe:     12,
			Ftics:      8,
			Fnextstate: s_BOS2_DIE6,
		},
		576: {
			Fsprite:    spr_BOS2,
			Fframe:     13,
			Ftics:      8,
			Fnextstate: s_BOS2_DIE7,
		},
		577: {
			Fsprite: spr_BOS2,
			Fframe:  14,
			Ftics:   -1,
		},
		578: {
			Fsprite:    spr_BOS2,
			Fframe:     14,
			Ftics:      8,
			Fnextstate: s_BOS2_RAISE2,
		},
		579: {
			Fsprite:    spr_BOS2,
			Fframe:     13,
			Ftics:      8,
			Fnextstate: s_BOS2_RAISE3,
		},
		580: {
			Fsprite:    spr_BOS2,
			Fframe:     12,
			Ftics:      8,
			Fnextstate: s_BOS2_RAISE4,
		},
		581: {
			Fsprite:    spr_BOS2,
			Fframe:     11,
			Ftics:      8,
			Fnextstate: s_BOS2_RAISE5,
		},
		582: {
			Fsprite:    spr_BOS2,
			Fframe:     10,
			Ftics:      8,
			Fnextstate: s_BOS2_RAISE6,
		},
		583: {
			Fsprite:    spr_BOS2,
			Fframe:     9,
			Ftics:      8,
			Fnextstate: s_BOS2_RAISE7,
		},
		584: {
			Fsprite:    spr_BOS2,
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_BOS2_RUN1,
		},
		585: {
			Fsprite:    spr_SKUL,
			Fframe:     32768,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SKULL_STND2,
		},
		586: {
			Fsprite:    spr_SKUL,
			Fframe:     32769,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SKULL_STND,
		},
		587: {
			Fsprite:    spr_SKUL,
			Fframe:     32768,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKULL_RUN2,
		},
		588: {
			Fsprite:    spr_SKUL,
			Fframe:     32769,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SKULL_RUN1,
		},
		589: {
			Fsprite:    spr_SKUL,
			Fframe:     32770,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SKULL_ATK2,
		},
		590: {
			Fsprite:    spr_SKUL,
			Fframe:     32771,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_SkullAttack),
			Fnextstate: s_SKULL_ATK3,
		},
		591: {
			Fsprite:    spr_SKUL,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_SKULL_ATK4,
		},
		592: {
			Fsprite:    spr_SKUL,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_SKULL_ATK3,
		},
		593: {
			Fsprite:    spr_SKUL,
			Fframe:     32772,
			Ftics:      3,
			Fnextstate: s_SKULL_PAIN2,
		},
		594: {
			Fsprite:    spr_SKUL,
			Fframe:     32772,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_SKULL_RUN1,
		},
		595: {
			Fsprite:    spr_SKUL,
			Fframe:     32773,
			Ftics:      6,
			Fnextstate: s_SKULL_DIE2,
		},
		596: {
			Fsprite:    spr_SKUL,
			Fframe:     32774,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_SKULL_DIE3,
		},
		597: {
			Fsprite:    spr_SKUL,
			Fframe:     32775,
			Ftics:      6,
			Fnextstate: s_SKULL_DIE4,
		},
		598: {
			Fsprite:    spr_SKUL,
			Fframe:     32776,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_SKULL_DIE5,
		},
		599: {
			Fsprite:    spr_SKUL,
			Fframe:     9,
			Ftics:      6,
			Fnextstate: s_SKULL_DIE6,
		},
		600: {
			Fsprite: spr_SKUL,
			Fframe:  10,
			Ftics:   6,
		},
		601: {
			Fsprite:    spr_SPID,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SPID_STND2,
		},
		602: {
			Fsprite:    spr_SPID,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SPID_STND,
		},
		603: {
			Fsprite:    spr_SPID,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Metal),
			Fnextstate: s_SPID_RUN2,
		},
		604: {
			Fsprite:    spr_SPID,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN3,
		},
		605: {
			Fsprite:    spr_SPID,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN4,
		},
		606: {
			Fsprite:    spr_SPID,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN5,
		},
		607: {
			Fsprite:    spr_SPID,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Metal),
			Fnextstate: s_SPID_RUN6,
		},
		608: {
			Fsprite:    spr_SPID,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN7,
		},
		609: {
			Fsprite:    spr_SPID,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN8,
		},
		610: {
			Fsprite:    spr_SPID,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN9,
		},
		611: {
			Fsprite:    spr_SPID,
			Fframe:     4,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Metal),
			Fnextstate: s_SPID_RUN10,
		},
		612: {
			Fsprite:    spr_SPID,
			Fframe:     4,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN11,
		},
		613: {
			Fsprite:    spr_SPID,
			Fframe:     5,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN12,
		},
		614: {
			Fsprite:    spr_SPID,
			Fframe:     5,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SPID_RUN1,
		},
		615: {
			Fsprite:    spr_SPID,
			Fframe:     32768,
			Ftics:      20,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SPID_ATK2,
		},
		616: {
			Fsprite:    spr_SPID,
			Fframe:     32774,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_SPosAttack),
			Fnextstate: s_SPID_ATK3,
		},
		617: {
			Fsprite:    spr_SPID,
			Fframe:     32775,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_SPosAttack),
			Fnextstate: s_SPID_ATK4,
		},
		618: {
			Fsprite:    spr_SPID,
			Fframe:     32775,
			Ftics:      1,
			Faction:    mobjFuncToAction(a_SpidRefire),
			Fnextstate: s_SPID_ATK2,
		},
		619: {
			Fsprite:    spr_SPID,
			Fframe:     8,
			Ftics:      3,
			Fnextstate: s_SPID_PAIN2,
		},
		620: {
			Fsprite:    spr_SPID,
			Fframe:     8,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_SPID_RUN1,
		},
		621: {
			Fsprite:    spr_SPID,
			Fframe:     9,
			Ftics:      20,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_SPID_DIE2,
		},
		622: {
			Fsprite:    spr_SPID,
			Fframe:     10,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_SPID_DIE3,
		},
		623: {
			Fsprite:    spr_SPID,
			Fframe:     11,
			Ftics:      10,
			Fnextstate: s_SPID_DIE4,
		},
		624: {
			Fsprite:    spr_SPID,
			Fframe:     12,
			Ftics:      10,
			Fnextstate: s_SPID_DIE5,
		},
		625: {
			Fsprite:    spr_SPID,
			Fframe:     13,
			Ftics:      10,
			Fnextstate: s_SPID_DIE6,
		},
		626: {
			Fsprite:    spr_SPID,
			Fframe:     14,
			Ftics:      10,
			Fnextstate: s_SPID_DIE7,
		},
		627: {
			Fsprite:    spr_SPID,
			Fframe:     15,
			Ftics:      10,
			Fnextstate: s_SPID_DIE8,
		},
		628: {
			Fsprite:    spr_SPID,
			Fframe:     16,
			Ftics:      10,
			Fnextstate: s_SPID_DIE9,
		},
		629: {
			Fsprite:    spr_SPID,
			Fframe:     17,
			Ftics:      10,
			Fnextstate: s_SPID_DIE10,
		},
		630: {
			Fsprite:    spr_SPID,
			Fframe:     18,
			Ftics:      30,
			Fnextstate: s_SPID_DIE11,
		},
		631: {
			Fsprite: spr_SPID,
			Fframe:  18,
			Ftics:   -1,
			Faction: mobjFuncToAction(a_BossDeath),
		},
		632: {
			Fsprite:    spr_BSPI,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_BSPI_STND2,
		},
		633: {
			Fsprite:    spr_BSPI,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_BSPI_STND,
		},
		634: {
			Fsprite:    spr_BSPI,
			Ftics:      20,
			Fnextstate: s_BSPI_RUN1,
		},
		635: {
			Fsprite:    spr_BSPI,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_BabyMetal),
			Fnextstate: s_BSPI_RUN2,
		},
		636: {
			Fsprite:    spr_BSPI,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN3,
		},
		637: {
			Fsprite:    spr_BSPI,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN4,
		},
		638: {
			Fsprite:    spr_BSPI,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN5,
		},
		639: {
			Fsprite:    spr_BSPI,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN6,
		},
		640: {
			Fsprite:    spr_BSPI,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN7,
		},
		641: {
			Fsprite:    spr_BSPI,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_BabyMetal),
			Fnextstate: s_BSPI_RUN8,
		},
		642: {
			Fsprite:    spr_BSPI,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN9,
		},
		643: {
			Fsprite:    spr_BSPI,
			Fframe:     4,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN10,
		},
		644: {
			Fsprite:    spr_BSPI,
			Fframe:     4,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN11,
		},
		645: {
			Fsprite:    spr_BSPI,
			Fframe:     5,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN12,
		},
		646: {
			Fsprite:    spr_BSPI,
			Fframe:     5,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_BSPI_RUN1,
		},
		647: {
			Fsprite:    spr_BSPI,
			Fframe:     32768,
			Ftics:      20,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_BSPI_ATK2,
		},
		648: {
			Fsprite:    spr_BSPI,
			Fframe:     32774,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_BspiAttack),
			Fnextstate: s_BSPI_ATK3,
		},
		649: {
			Fsprite:    spr_BSPI,
			Fframe:     32775,
			Ftics:      4,
			Fnextstate: s_BSPI_ATK4,
		},
		650: {
			Fsprite:    spr_BSPI,
			Fframe:     32775,
			Ftics:      1,
			Faction:    mobjFuncToAction(a_SpidRefire),
			Fnextstate: s_BSPI_ATK2,
		},
		651: {
			Fsprite:    spr_BSPI,
			Fframe:     8,
			Ftics:      3,
			Fnextstate: s_BSPI_PAIN2,
		},
		652: {
			Fsprite:    spr_BSPI,
			Fframe:     8,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_BSPI_RUN1,
		},
		653: {
			Fsprite:    spr_BSPI,
			Fframe:     9,
			Ftics:      20,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_BSPI_DIE2,
		},
		654: {
			Fsprite:    spr_BSPI,
			Fframe:     10,
			Ftics:      7,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_BSPI_DIE3,
		},
		655: {
			Fsprite:    spr_BSPI,
			Fframe:     11,
			Ftics:      7,
			Fnextstate: s_BSPI_DIE4,
		},
		656: {
			Fsprite:    spr_BSPI,
			Fframe:     12,
			Ftics:      7,
			Fnextstate: s_BSPI_DIE5,
		},
		657: {
			Fsprite:    spr_BSPI,
			Fframe:     13,
			Ftics:      7,
			Fnextstate: s_BSPI_DIE6,
		},
		658: {
			Fsprite:    spr_BSPI,
			Fframe:     14,
			Ftics:      7,
			Fnextstate: s_BSPI_DIE7,
		},
		659: {
			Fsprite: spr_BSPI,
			Fframe:  15,
			Ftics:   -1,
			Faction: mobjFuncToAction(a_BossDeath),
		},
		660: {
			Fsprite:    spr_BSPI,
			Fframe:     15,
			Ftics:      5,
			Fnextstate: s_BSPI_RAISE2,
		},
		661: {
			Fsprite:    spr_BSPI,
			Fframe:     14,
			Ftics:      5,
			Fnextstate: s_BSPI_RAISE3,
		},
		662: {
			Fsprite:    spr_BSPI,
			Fframe:     13,
			Ftics:      5,
			Fnextstate: s_BSPI_RAISE4,
		},
		663: {
			Fsprite:    spr_BSPI,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_BSPI_RAISE5,
		},
		664: {
			Fsprite:    spr_BSPI,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_BSPI_RAISE6,
		},
		665: {
			Fsprite:    spr_BSPI,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_BSPI_RAISE7,
		},
		666: {
			Fsprite:    spr_BSPI,
			Fframe:     9,
			Ftics:      5,
			Fnextstate: s_BSPI_RUN1,
		},
		667: {
			Fsprite:    spr_APLS,
			Fframe:     32768,
			Ftics:      5,
			Fnextstate: s_ARACH_PLAZ2,
		},
		668: {
			Fsprite:    spr_APLS,
			Fframe:     32769,
			Ftics:      5,
			Fnextstate: s_ARACH_PLAZ,
		},
		669: {
			Fsprite:    spr_APBX,
			Fframe:     32768,
			Ftics:      5,
			Fnextstate: s_ARACH_PLEX2,
		},
		670: {
			Fsprite:    spr_APBX,
			Fframe:     32769,
			Ftics:      5,
			Fnextstate: s_ARACH_PLEX3,
		},
		671: {
			Fsprite:    spr_APBX,
			Fframe:     32770,
			Ftics:      5,
			Fnextstate: s_ARACH_PLEX4,
		},
		672: {
			Fsprite:    spr_APBX,
			Fframe:     32771,
			Ftics:      5,
			Fnextstate: s_ARACH_PLEX5,
		},
		673: {
			Fsprite: spr_APBX,
			Fframe:  32772,
			Ftics:   5,
		},
		674: {
			Fsprite:    spr_CYBR,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_CYBER_STND2,
		},
		675: {
			Fsprite:    spr_CYBR,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_CYBER_STND,
		},
		676: {
			Fsprite:    spr_CYBR,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Hoof),
			Fnextstate: s_CYBER_RUN2,
		},
		677: {
			Fsprite:    spr_CYBR,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CYBER_RUN3,
		},
		678: {
			Fsprite:    spr_CYBR,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CYBER_RUN4,
		},
		679: {
			Fsprite:    spr_CYBR,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CYBER_RUN5,
		},
		680: {
			Fsprite:    spr_CYBR,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CYBER_RUN6,
		},
		681: {
			Fsprite:    spr_CYBR,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CYBER_RUN7,
		},
		682: {
			Fsprite:    spr_CYBR,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Metal),
			Fnextstate: s_CYBER_RUN8,
		},
		683: {
			Fsprite:    spr_CYBR,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_CYBER_RUN1,
		},
		684: {
			Fsprite:    spr_CYBR,
			Fframe:     4,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_CYBER_ATK2,
		},
		685: {
			Fsprite:    spr_CYBR,
			Fframe:     5,
			Ftics:      12,
			Faction:    mobjFuncToAction(a_CyberAttack),
			Fnextstate: s_CYBER_ATK3,
		},
		686: {
			Fsprite:    spr_CYBR,
			Fframe:     4,
			Ftics:      12,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_CYBER_ATK4,
		},
		687: {
			Fsprite:    spr_CYBR,
			Fframe:     5,
			Ftics:      12,
			Faction:    mobjFuncToAction(a_CyberAttack),
			Fnextstate: s_CYBER_ATK5,
		},
		688: {
			Fsprite:    spr_CYBR,
			Fframe:     4,
			Ftics:      12,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_CYBER_ATK6,
		},
		689: {
			Fsprite:    spr_CYBR,
			Fframe:     5,
			Ftics:      12,
			Faction:    mobjFuncToAction(a_CyberAttack),
			Fnextstate: s_CYBER_RUN1,
		},
		690: {
			Fsprite:    spr_CYBR,
			Fframe:     6,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_CYBER_RUN1,
		},
		691: {
			Fsprite:    spr_CYBR,
			Fframe:     7,
			Ftics:      10,
			Fnextstate: s_CYBER_DIE2,
		},
		692: {
			Fsprite:    spr_CYBR,
			Fframe:     8,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_CYBER_DIE3,
		},
		693: {
			Fsprite:    spr_CYBR,
			Fframe:     9,
			Ftics:      10,
			Fnextstate: s_CYBER_DIE4,
		},
		694: {
			Fsprite:    spr_CYBR,
			Fframe:     10,
			Ftics:      10,
			Fnextstate: s_CYBER_DIE5,
		},
		695: {
			Fsprite:    spr_CYBR,
			Fframe:     11,
			Ftics:      10,
			Fnextstate: s_CYBER_DIE6,
		},
		696: {
			Fsprite:    spr_CYBR,
			Fframe:     12,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_CYBER_DIE7,
		},
		697: {
			Fsprite:    spr_CYBR,
			Fframe:     13,
			Ftics:      10,
			Fnextstate: s_CYBER_DIE8,
		},
		698: {
			Fsprite:    spr_CYBR,
			Fframe:     14,
			Ftics:      10,
			Fnextstate: s_CYBER_DIE9,
		},
		699: {
			Fsprite:    spr_CYBR,
			Fframe:     15,
			Ftics:      30,
			Fnextstate: s_CYBER_DIE10,
		},
		700: {
			Fsprite: spr_CYBR,
			Fframe:  15,
			Ftics:   -1,
			Faction: mobjFuncToAction(a_BossDeath),
		},
		701: {
			Fsprite:    spr_PAIN,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_PAIN_STND,
		},
		702: {
			Fsprite:    spr_PAIN,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_PAIN_RUN2,
		},
		703: {
			Fsprite:    spr_PAIN,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_PAIN_RUN3,
		},
		704: {
			Fsprite:    spr_PAIN,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_PAIN_RUN4,
		},
		705: {
			Fsprite:    spr_PAIN,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_PAIN_RUN5,
		},
		706: {
			Fsprite:    spr_PAIN,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_PAIN_RUN6,
		},
		707: {
			Fsprite:    spr_PAIN,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_PAIN_RUN1,
		},
		708: {
			Fsprite:    spr_PAIN,
			Fframe:     3,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_PAIN_ATK2,
		},
		709: {
			Fsprite:    spr_PAIN,
			Fframe:     4,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_PAIN_ATK3,
		},
		710: {
			Fsprite:    spr_PAIN,
			Fframe:     32773,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_PAIN_ATK4,
		},
		711: {
			Fsprite:    spr_PAIN,
			Fframe:     32773,
			Faction:    mobjFuncToAction(a_PainAttack),
			Fnextstate: s_PAIN_RUN1,
		},
		712: {
			Fsprite:    spr_PAIN,
			Fframe:     6,
			Ftics:      6,
			Fnextstate: s_PAIN_PAIN2,
		},
		713: {
			Fsprite:    spr_PAIN,
			Fframe:     6,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_PAIN_RUN1,
		},
		714: {
			Fsprite:    spr_PAIN,
			Fframe:     32775,
			Ftics:      8,
			Fnextstate: s_PAIN_DIE2,
		},
		715: {
			Fsprite:    spr_PAIN,
			Fframe:     32776,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_PAIN_DIE3,
		},
		716: {
			Fsprite:    spr_PAIN,
			Fframe:     32777,
			Ftics:      8,
			Fnextstate: s_PAIN_DIE4,
		},
		717: {
			Fsprite:    spr_PAIN,
			Fframe:     32778,
			Ftics:      8,
			Fnextstate: s_PAIN_DIE5,
		},
		718: {
			Fsprite:    spr_PAIN,
			Fframe:     32779,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_PainDie),
			Fnextstate: s_PAIN_DIE6,
		},
		719: {
			Fsprite: spr_PAIN,
			Fframe:  32780,
			Ftics:   8,
		},
		720: {
			Fsprite:    spr_PAIN,
			Fframe:     12,
			Ftics:      8,
			Fnextstate: s_PAIN_RAISE2,
		},
		721: {
			Fsprite:    spr_PAIN,
			Fframe:     11,
			Ftics:      8,
			Fnextstate: s_PAIN_RAISE3,
		},
		722: {
			Fsprite:    spr_PAIN,
			Fframe:     10,
			Ftics:      8,
			Fnextstate: s_PAIN_RAISE4,
		},
		723: {
			Fsprite:    spr_PAIN,
			Fframe:     9,
			Ftics:      8,
			Fnextstate: s_PAIN_RAISE5,
		},
		724: {
			Fsprite:    spr_PAIN,
			Fframe:     8,
			Ftics:      8,
			Fnextstate: s_PAIN_RAISE6,
		},
		725: {
			Fsprite:    spr_PAIN,
			Fframe:     7,
			Ftics:      8,
			Fnextstate: s_PAIN_RUN1,
		},
		726: {
			Fsprite:    spr_SSWV,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SSWV_STND2,
		},
		727: {
			Fsprite:    spr_SSWV,
			Fframe:     1,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_SSWV_STND,
		},
		728: {
			Fsprite:    spr_SSWV,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SSWV_RUN2,
		},
		729: {
			Fsprite:    spr_SSWV,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SSWV_RUN3,
		},
		730: {
			Fsprite:    spr_SSWV,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SSWV_RUN4,
		},
		731: {
			Fsprite:    spr_SSWV,
			Fframe:     1,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SSWV_RUN5,
		},
		732: {
			Fsprite:    spr_SSWV,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SSWV_RUN6,
		},
		733: {
			Fsprite:    spr_SSWV,
			Fframe:     2,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SSWV_RUN7,
		},
		734: {
			Fsprite:    spr_SSWV,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SSWV_RUN8,
		},
		735: {
			Fsprite:    spr_SSWV,
			Fframe:     3,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Chase),
			Fnextstate: s_SSWV_RUN1,
		},
		736: {
			Fsprite:    spr_SSWV,
			Fframe:     4,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SSWV_ATK2,
		},
		737: {
			Fsprite:    spr_SSWV,
			Fframe:     5,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SSWV_ATK3,
		},
		738: {
			Fsprite:    spr_SSWV,
			Fframe:     32774,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_CPosAttack),
			Fnextstate: s_SSWV_ATK4,
		},
		739: {
			Fsprite:    spr_SSWV,
			Fframe:     5,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_FaceTarget),
			Fnextstate: s_SSWV_ATK5,
		},
		740: {
			Fsprite:    spr_SSWV,
			Fframe:     32774,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_CPosAttack),
			Fnextstate: s_SSWV_ATK6,
		},
		741: {
			Fsprite:    spr_SSWV,
			Fframe:     5,
			Ftics:      1,
			Faction:    mobjFuncToAction(a_CPosRefire),
			Fnextstate: s_SSWV_ATK2,
		},
		742: {
			Fsprite:    spr_SSWV,
			Fframe:     7,
			Ftics:      3,
			Fnextstate: s_SSWV_PAIN2,
		},
		743: {
			Fsprite:    spr_SSWV,
			Fframe:     7,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_SSWV_RUN1,
		},
		744: {
			Fsprite:    spr_SSWV,
			Fframe:     8,
			Ftics:      5,
			Fnextstate: s_SSWV_DIE2,
		},
		745: {
			Fsprite:    spr_SSWV,
			Fframe:     9,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_SSWV_DIE3,
		},
		746: {
			Fsprite:    spr_SSWV,
			Fframe:     10,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_SSWV_DIE4,
		},
		747: {
			Fsprite:    spr_SSWV,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_SSWV_DIE5,
		},
		748: {
			Fsprite: spr_SSWV,
			Fframe:  12,
			Ftics:   -1,
		},
		749: {
			Fsprite:    spr_SSWV,
			Fframe:     13,
			Ftics:      5,
			Fnextstate: s_SSWV_XDIE2,
		},
		750: {
			Fsprite:    spr_SSWV,
			Fframe:     14,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_XScream),
			Fnextstate: s_SSWV_XDIE3,
		},
		751: {
			Fsprite:    spr_SSWV,
			Fframe:     15,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Fall),
			Fnextstate: s_SSWV_XDIE4,
		},
		752: {
			Fsprite:    spr_SSWV,
			Fframe:     16,
			Ftics:      5,
			Fnextstate: s_SSWV_XDIE5,
		},
		753: {
			Fsprite:    spr_SSWV,
			Fframe:     17,
			Ftics:      5,
			Fnextstate: s_SSWV_XDIE6,
		},
		754: {
			Fsprite:    spr_SSWV,
			Fframe:     18,
			Ftics:      5,
			Fnextstate: s_SSWV_XDIE7,
		},
		755: {
			Fsprite:    spr_SSWV,
			Fframe:     19,
			Ftics:      5,
			Fnextstate: s_SSWV_XDIE8,
		},
		756: {
			Fsprite:    spr_SSWV,
			Fframe:     20,
			Ftics:      5,
			Fnextstate: s_SSWV_XDIE9,
		},
		757: {
			Fsprite: spr_SSWV,
			Fframe:  21,
			Ftics:   -1,
		},
		758: {
			Fsprite:    spr_SSWV,
			Fframe:     12,
			Ftics:      5,
			Fnextstate: s_SSWV_RAISE2,
		},
		759: {
			Fsprite:    spr_SSWV,
			Fframe:     11,
			Ftics:      5,
			Fnextstate: s_SSWV_RAISE3,
		},
		760: {
			Fsprite:    spr_SSWV,
			Fframe:     10,
			Ftics:      5,
			Fnextstate: s_SSWV_RAISE4,
		},
		761: {
			Fsprite:    spr_SSWV,
			Fframe:     9,
			Ftics:      5,
			Fnextstate: s_SSWV_RAISE5,
		},
		762: {
			Fsprite:    spr_SSWV,
			Fframe:     8,
			Ftics:      5,
			Fnextstate: s_SSWV_RUN1,
		},
		763: {
			Fsprite:    spr_KEEN,
			Ftics:      -1,
			Fnextstate: s_KEENSTND,
		},
		764: {
			Fsprite:    spr_KEEN,
			Ftics:      6,
			Fnextstate: s_COMMKEEN2,
		},
		765: {
			Fsprite:    spr_KEEN,
			Fframe:     1,
			Ftics:      6,
			Fnextstate: s_COMMKEEN3,
		},
		766: {
			Fsprite:    spr_KEEN,
			Fframe:     2,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_COMMKEEN4,
		},
		767: {
			Fsprite:    spr_KEEN,
			Fframe:     3,
			Ftics:      6,
			Fnextstate: s_COMMKEEN5,
		},
		768: {
			Fsprite:    spr_KEEN,
			Fframe:     4,
			Ftics:      6,
			Fnextstate: s_COMMKEEN6,
		},
		769: {
			Fsprite:    spr_KEEN,
			Fframe:     5,
			Ftics:      6,
			Fnextstate: s_COMMKEEN7,
		},
		770: {
			Fsprite:    spr_KEEN,
			Fframe:     6,
			Ftics:      6,
			Fnextstate: s_COMMKEEN8,
		},
		771: {
			Fsprite:    spr_KEEN,
			Fframe:     7,
			Ftics:      6,
			Fnextstate: s_COMMKEEN9,
		},
		772: {
			Fsprite:    spr_KEEN,
			Fframe:     8,
			Ftics:      6,
			Fnextstate: s_COMMKEEN10,
		},
		773: {
			Fsprite:    spr_KEEN,
			Fframe:     9,
			Ftics:      6,
			Fnextstate: s_COMMKEEN11,
		},
		774: {
			Fsprite:    spr_KEEN,
			Fframe:     10,
			Ftics:      6,
			Faction:    mobjFuncToAction(a_KeenDie),
			Fnextstate: s_COMMKEEN12,
		},
		775: {
			Fsprite: spr_KEEN,
			Fframe:  11,
			Ftics:   -1,
		},
		776: {
			Fsprite:    spr_KEEN,
			Fframe:     12,
			Ftics:      4,
			Fnextstate: s_KEENPAIN2,
		},
		777: {
			Fsprite:    spr_KEEN,
			Fframe:     12,
			Ftics:      8,
			Faction:    mobjFuncToAction(a_Pain),
			Fnextstate: s_KEENSTND,
		},
		778: {
			Fsprite: spr_BBRN,
			Ftics:   -1,
		},
		779: {
			Fsprite:    spr_BBRN,
			Fframe:     1,
			Ftics:      36,
			Faction:    mobjFuncToAction(a_BrainPain),
			Fnextstate: s_BRAIN,
		},
		780: {
			Fsprite:    spr_BBRN,
			Ftics:      100,
			Faction:    mobjFuncToAction(a_BrainScream),
			Fnextstate: s_BRAIN_DIE2,
		},
		781: {
			Fsprite:    spr_BBRN,
			Ftics:      10,
			Fnextstate: s_BRAIN_DIE3,
		},
		782: {
			Fsprite:    spr_BBRN,
			Ftics:      10,
			Fnextstate: s_BRAIN_DIE4,
		},
		783: {
			Fsprite: spr_BBRN,
			Ftics:   -1,
			Faction: mobjFuncToAction(a_BrainDie),
		},
		784: {
			Fsprite:    spr_SSWV,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Look),
			Fnextstate: s_BRAINEYE,
		},
		785: {
			Fsprite:    spr_SSWV,
			Ftics:      181,
			Faction:    mobjFuncToAction(a_BrainAwake),
			Fnextstate: s_BRAINEYE1,
		},
		786: {
			Fsprite:    spr_SSWV,
			Ftics:      150,
			Faction:    mobjFuncToAction(a_BrainSpit),
			Fnextstate: s_BRAINEYE1,
		},
		787: {
			Fsprite:    spr_BOSF,
			Fframe:     32768,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_SpawnSound),
			Fnextstate: s_SPAWN2,
		},
		788: {
			Fsprite:    spr_BOSF,
			Fframe:     32769,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_SpawnFly),
			Fnextstate: s_SPAWN3,
		},
		789: {
			Fsprite:    spr_BOSF,
			Fframe:     32770,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_SpawnFly),
			Fnextstate: s_SPAWN4,
		},
		790: {
			Fsprite:    spr_BOSF,
			Fframe:     32771,
			Ftics:      3,
			Faction:    mobjFuncToAction(a_SpawnFly),
			Fnextstate: s_SPAWN1,
		},
		791: {
			Fsprite:    spr_FIRE,
			Fframe:     32768,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_SPAWNFIRE2,
		},
		792: {
			Fsprite:    spr_FIRE,
			Fframe:     32769,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_SPAWNFIRE3,
		},
		793: {
			Fsprite:    spr_FIRE,
			Fframe:     32770,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_SPAWNFIRE4,
		},
		794: {
			Fsprite:    spr_FIRE,
			Fframe:     32771,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_SPAWNFIRE5,
		},
		795: {
			Fsprite:    spr_FIRE,
			Fframe:     32772,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_SPAWNFIRE6,
		},
		796: {
			Fsprite:    spr_FIRE,
			Fframe:     32773,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_SPAWNFIRE7,
		},
		797: {
			Fsprite:    spr_FIRE,
			Fframe:     32774,
			Ftics:      4,
			Faction:    mobjFuncToAction(a_Fire),
			Fnextstate: s_SPAWNFIRE8,
		},
		798: {
			Fsprite: spr_FIRE,
			Fframe:  32775,
			Ftics:   4,
			Faction: mobjFuncToAction(a_Fire),
		},
		799: {
			Fsprite:    spr_MISL,
			Fframe:     32769,
			Ftics:      10,
			Fnextstate: s_BRAINEXPLODE2,
		},
		800: {
			Fsprite:    spr_MISL,
			Fframe:     32770,
			Ftics:      10,
			Fnextstate: s_BRAINEXPLODE3,
		},
		801: {
			Fsprite: spr_MISL,
			Fframe:  32771,
			Ftics:   10,
			Faction: mobjFuncToAction(a_BrainExplode),
		},
		802: {
			Fsprite:    spr_ARM1,
			Ftics:      6,
			Fnextstate: s_ARM1A,
		},
		803: {
			Fsprite:    spr_ARM1,
			Fframe:     32769,
			Ftics:      7,
			Fnextstate: s_ARM1,
		},
		804: {
			Fsprite:    spr_ARM2,
			Ftics:      6,
			Fnextstate: s_ARM2A,
		},
		805: {
			Fsprite:    spr_ARM2,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_ARM2,
		},
		806: {
			Fsprite:    spr_BAR1,
			Ftics:      6,
			Fnextstate: s_BAR2,
		},
		807: {
			Fsprite:    spr_BAR1,
			Fframe:     1,
			Ftics:      6,
			Fnextstate: s_BAR1,
		},
		808: {
			Fsprite:    spr_BEXP,
			Fframe:     32768,
			Ftics:      5,
			Fnextstate: s_BEXP2,
		},
		809: {
			Fsprite:    spr_BEXP,
			Fframe:     32769,
			Ftics:      5,
			Faction:    mobjFuncToAction(a_Scream),
			Fnextstate: s_BEXP3,
		},
		810: {
			Fsprite:    spr_BEXP,
			Fframe:     32770,
			Ftics:      5,
			Fnextstate: s_BEXP4,
		},
		811: {
			Fsprite:    spr_BEXP,
			Fframe:     32771,
			Ftics:      10,
			Faction:    mobjFuncToAction(a_Explode),
			Fnextstate: s_BEXP5,
		},
		812: {
			Fsprite: spr_BEXP,
			Fframe:  32772,
			Ftics:   10,
		},
		813: {
			Fsprite:    spr_FCAN,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_BBAR2,
		},
		814: {
			Fsprite:    spr_FCAN,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_BBAR3,
		},
		815: {
			Fsprite:    spr_FCAN,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_BBAR1,
		},
		816: {
			Fsprite:    spr_BON1,
			Ftics:      6,
			Fnextstate: s_BON1A,
		},
		817: {
			Fsprite:    spr_BON1,
			Fframe:     1,
			Ftics:      6,
			Fnextstate: s_BON1B,
		},
		818: {
			Fsprite:    spr_BON1,
			Fframe:     2,
			Ftics:      6,
			Fnextstate: s_BON1C,
		},
		819: {
			Fsprite:    spr_BON1,
			Fframe:     3,
			Ftics:      6,
			Fnextstate: s_BON1D,
		},
		820: {
			Fsprite:    spr_BON1,
			Fframe:     2,
			Ftics:      6,
			Fnextstate: s_BON1E,
		},
		821: {
			Fsprite:    spr_BON1,
			Fframe:     1,
			Ftics:      6,
			Fnextstate: s_BON1,
		},
		822: {
			Fsprite:    spr_BON2,
			Ftics:      6,
			Fnextstate: s_BON2A,
		},
		823: {
			Fsprite:    spr_BON2,
			Fframe:     1,
			Ftics:      6,
			Fnextstate: s_BON2B,
		},
		824: {
			Fsprite:    spr_BON2,
			Fframe:     2,
			Ftics:      6,
			Fnextstate: s_BON2C,
		},
		825: {
			Fsprite:    spr_BON2,
			Fframe:     3,
			Ftics:      6,
			Fnextstate: s_BON2D,
		},
		826: {
			Fsprite:    spr_BON2,
			Fframe:     2,
			Ftics:      6,
			Fnextstate: s_BON2E,
		},
		827: {
			Fsprite:    spr_BON2,
			Fframe:     1,
			Ftics:      6,
			Fnextstate: s_BON2,
		},
		828: {
			Fsprite:    spr_BKEY,
			Ftics:      10,
			Fnextstate: s_BKEY2,
		},
		829: {
			Fsprite:    spr_BKEY,
			Fframe:     32769,
			Ftics:      10,
			Fnextstate: s_BKEY,
		},
		830: {
			Fsprite:    spr_RKEY,
			Ftics:      10,
			Fnextstate: s_RKEY2,
		},
		831: {
			Fsprite:    spr_RKEY,
			Fframe:     32769,
			Ftics:      10,
			Fnextstate: s_RKEY,
		},
		832: {
			Fsprite:    spr_YKEY,
			Ftics:      10,
			Fnextstate: s_YKEY2,
		},
		833: {
			Fsprite:    spr_YKEY,
			Fframe:     32769,
			Ftics:      10,
			Fnextstate: s_YKEY,
		},
		834: {
			Fsprite:    spr_BSKU,
			Ftics:      10,
			Fnextstate: s_BSKULL2,
		},
		835: {
			Fsprite:    spr_BSKU,
			Fframe:     32769,
			Ftics:      10,
			Fnextstate: s_BSKULL,
		},
		836: {
			Fsprite:    spr_RSKU,
			Ftics:      10,
			Fnextstate: s_RSKULL2,
		},
		837: {
			Fsprite:    spr_RSKU,
			Fframe:     32769,
			Ftics:      10,
			Fnextstate: s_RSKULL,
		},
		838: {
			Fsprite:    spr_YSKU,
			Ftics:      10,
			Fnextstate: s_YSKULL2,
		},
		839: {
			Fsprite:    spr_YSKU,
			Fframe:     32769,
			Ftics:      10,
			Fnextstate: s_YSKULL,
		},
		840: {
			Fsprite: spr_STIM,
			Ftics:   -1,
		},
		841: {
			Fsprite: spr_MEDI,
			Ftics:   -1,
		},
		842: {
			Fsprite:    spr_SOUL,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_SOUL2,
		},
		843: {
			Fsprite:    spr_SOUL,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_SOUL3,
		},
		844: {
			Fsprite:    spr_SOUL,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_SOUL4,
		},
		845: {
			Fsprite:    spr_SOUL,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_SOUL5,
		},
		846: {
			Fsprite:    spr_SOUL,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_SOUL6,
		},
		847: {
			Fsprite:    spr_SOUL,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_SOUL,
		},
		848: {
			Fsprite:    spr_PINV,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_PINV2,
		},
		849: {
			Fsprite:    spr_PINV,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_PINV3,
		},
		850: {
			Fsprite:    spr_PINV,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_PINV4,
		},
		851: {
			Fsprite:    spr_PINV,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_PINV,
		},
		852: {
			Fsprite: spr_PSTR,
			Fframe:  32768,
			Ftics:   -1,
		},
		853: {
			Fsprite:    spr_PINS,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_PINS2,
		},
		854: {
			Fsprite:    spr_PINS,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_PINS3,
		},
		855: {
			Fsprite:    spr_PINS,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_PINS4,
		},
		856: {
			Fsprite:    spr_PINS,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_PINS,
		},
		857: {
			Fsprite:    spr_MEGA,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_MEGA2,
		},
		858: {
			Fsprite:    spr_MEGA,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_MEGA3,
		},
		859: {
			Fsprite:    spr_MEGA,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_MEGA4,
		},
		860: {
			Fsprite:    spr_MEGA,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_MEGA,
		},
		861: {
			Fsprite: spr_SUIT,
			Fframe:  32768,
			Ftics:   -1,
		},
		862: {
			Fsprite:    spr_PMAP,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_PMAP2,
		},
		863: {
			Fsprite:    spr_PMAP,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_PMAP3,
		},
		864: {
			Fsprite:    spr_PMAP,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_PMAP4,
		},
		865: {
			Fsprite:    spr_PMAP,
			Fframe:     32771,
			Ftics:      6,
			Fnextstate: s_PMAP5,
		},
		866: {
			Fsprite:    spr_PMAP,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_PMAP6,
		},
		867: {
			Fsprite:    spr_PMAP,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_PMAP,
		},
		868: {
			Fsprite:    spr_PVIS,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_PVIS2,
		},
		869: {
			Fsprite:    spr_PVIS,
			Fframe:     1,
			Ftics:      6,
			Fnextstate: s_PVIS,
		},
		870: {
			Fsprite: spr_CLIP,
			Ftics:   -1,
		},
		871: {
			Fsprite: spr_AMMO,
			Ftics:   -1,
		},
		872: {
			Fsprite: spr_ROCK,
			Ftics:   -1,
		},
		873: {
			Fsprite: spr_BROK,
			Ftics:   -1,
		},
		874: {
			Fsprite: spr_CELL,
			Ftics:   -1,
		},
		875: {
			Fsprite: spr_CELP,
			Ftics:   -1,
		},
		876: {
			Fsprite: spr_SHEL,
			Ftics:   -1,
		},
		877: {
			Fsprite: spr_SBOX,
			Ftics:   -1,
		},
		878: {
			Fsprite: spr_BPAK,
			Ftics:   -1,
		},
		879: {
			Fsprite: spr_BFUG,
			Ftics:   -1,
		},
		880: {
			Fsprite: spr_MGUN,
			Ftics:   -1,
		},
		881: {
			Fsprite: spr_CSAW,
			Ftics:   -1,
		},
		882: {
			Fsprite: spr_LAUN,
			Ftics:   -1,
		},
		883: {
			Fsprite: spr_PLAS,
			Ftics:   -1,
		},
		884: {
			Fsprite: spr_SHOT,
			Ftics:   -1,
		},
		885: {
			Fsprite: spr_SGN2,
			Ftics:   -1,
		},
		886: {
			Fsprite: spr_COLU,
			Fframe:  32768,
			Ftics:   -1,
		},
		887: {
			Fsprite: spr_SMT2,
			Ftics:   -1,
		},
		888: {
			Fsprite:    spr_GOR1,
			Ftics:      10,
			Fnextstate: s_BLOODYTWITCH2,
		},
		889: {
			Fsprite:    spr_GOR1,
			Fframe:     1,
			Ftics:      15,
			Fnextstate: s_BLOODYTWITCH3,
		},
		890: {
			Fsprite:    spr_GOR1,
			Fframe:     2,
			Ftics:      8,
			Fnextstate: s_BLOODYTWITCH4,
		},
		891: {
			Fsprite:    spr_GOR1,
			Fframe:     1,
			Ftics:      6,
			Fnextstate: s_BLOODYTWITCH,
		},
		892: {
			Fsprite: spr_PLAY,
			Fframe:  13,
			Ftics:   -1,
		},
		893: {
			Fsprite: spr_PLAY,
			Fframe:  18,
			Ftics:   -1,
		},
		894: {
			Fsprite: spr_POL2,
			Ftics:   -1,
		},
		895: {
			Fsprite: spr_POL5,
			Ftics:   -1,
		},
		896: {
			Fsprite: spr_POL4,
			Ftics:   -1,
		},
		897: {
			Fsprite:    spr_POL3,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_HEADCANDLES2,
		},
		898: {
			Fsprite:    spr_POL3,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_HEADCANDLES,
		},
		899: {
			Fsprite: spr_POL1,
			Ftics:   -1,
		},
		900: {
			Fsprite:    spr_POL6,
			Ftics:      6,
			Fnextstate: s_LIVESTICK2,
		},
		901: {
			Fsprite:    spr_POL6,
			Fframe:     1,
			Ftics:      8,
			Fnextstate: s_LIVESTICK,
		},
		902: {
			Fsprite: spr_GOR2,
			Ftics:   -1,
		},
		903: {
			Fsprite: spr_GOR3,
			Ftics:   -1,
		},
		904: {
			Fsprite: spr_GOR4,
			Ftics:   -1,
		},
		905: {
			Fsprite: spr_GOR5,
			Ftics:   -1,
		},
		906: {
			Fsprite: spr_SMIT,
			Ftics:   -1,
		},
		907: {
			Fsprite: spr_COL1,
			Ftics:   -1,
		},
		908: {
			Fsprite: spr_COL2,
			Ftics:   -1,
		},
		909: {
			Fsprite: spr_COL3,
			Ftics:   -1,
		},
		910: {
			Fsprite: spr_COL4,
			Ftics:   -1,
		},
		911: {
			Fsprite: spr_CAND,
			Fframe:  32768,
			Ftics:   -1,
		},
		912: {
			Fsprite: spr_CBRA,
			Fframe:  32768,
			Ftics:   -1,
		},
		913: {
			Fsprite: spr_COL6,
			Ftics:   -1,
		},
		914: {
			Fsprite: spr_TRE1,
			Ftics:   -1,
		},
		915: {
			Fsprite: spr_TRE2,
			Ftics:   -1,
		},
		916: {
			Fsprite: spr_ELEC,
			Ftics:   -1,
		},
		917: {
			Fsprite:    spr_CEYE,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_EVILEYE2,
		},
		918: {
			Fsprite:    spr_CEYE,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_EVILEYE3,
		},
		919: {
			Fsprite:    spr_CEYE,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_EVILEYE4,
		},
		920: {
			Fsprite:    spr_CEYE,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_EVILEYE,
		},
		921: {
			Fsprite:    spr_FSKU,
			Fframe:     32768,
			Ftics:      6,
			Fnextstate: s_FLOATSKULL2,
		},
		922: {
			Fsprite:    spr_FSKU,
			Fframe:     32769,
			Ftics:      6,
			Fnextstate: s_FLOATSKULL3,
		},
		923: {
			Fsprite:    spr_FSKU,
			Fframe:     32770,
			Ftics:      6,
			Fnextstate: s_FLOATSKULL,
		},
		924: {
			Fsprite:    spr_COL5,
			Ftics:      14,
			Fnextstate: s_HEARTCOL2,
		},
		925: {
			Fsprite:    spr_COL5,
			Fframe:     1,
			Ftics:      14,
			Fnextstate: s_HEARTCOL,
		},
		926: {
			Fsprite:    spr_TBLU,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_BLUETORCH2,
		},
		927: {
			Fsprite:    spr_TBLU,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_BLUETORCH3,
		},
		928: {
			Fsprite:    spr_TBLU,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_BLUETORCH4,
		},
		929: {
			Fsprite:    spr_TBLU,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_BLUETORCH,
		},
		930: {
			Fsprite:    spr_TGRN,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_GREENTORCH2,
		},
		931: {
			Fsprite:    spr_TGRN,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_GREENTORCH3,
		},
		932: {
			Fsprite:    spr_TGRN,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_GREENTORCH4,
		},
		933: {
			Fsprite:    spr_TGRN,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_GREENTORCH,
		},
		934: {
			Fsprite:    spr_TRED,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_REDTORCH2,
		},
		935: {
			Fsprite:    spr_TRED,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_REDTORCH3,
		},
		936: {
			Fsprite:    spr_TRED,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_REDTORCH4,
		},
		937: {
			Fsprite:    spr_TRED,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_REDTORCH,
		},
		938: {
			Fsprite:    spr_SMBT,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_BTORCHSHRT2,
		},
		939: {
			Fsprite:    spr_SMBT,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_BTORCHSHRT3,
		},
		940: {
			Fsprite:    spr_SMBT,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_BTORCHSHRT4,
		},
		941: {
			Fsprite:    spr_SMBT,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_BTORCHSHRT,
		},
		942: {
			Fsprite:    spr_SMGT,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_GTORCHSHRT2,
		},
		943: {
			Fsprite:    spr_SMGT,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_GTORCHSHRT3,
		},
		944: {
			Fsprite:    spr_SMGT,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_GTORCHSHRT4,
		},
		945: {
			Fsprite:    spr_SMGT,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_GTORCHSHRT,
		},
		946: {
			Fsprite:    spr_SMRT,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_RTORCHSHRT2,
		},
		947: {
			Fsprite:    spr_SMRT,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_RTORCHSHRT3,
		},
		948: {
			Fsprite:    spr_SMRT,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_RTORCHSHRT4,
		},
		949: {
			Fsprite:    spr_SMRT,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_RTORCHSHRT,
		},
		950: {
			Fsprite: spr_HDB1,
			Ftics:   -1,
		},
		951: {
			Fsprite: spr_HDB2,
			Ftics:   -1,
		},
		952: {
			Fsprite: spr_HDB3,
			Ftics:   -1,
		},
		953: {
			Fsprite: spr_HDB4,
			Ftics:   -1,
		},
		954: {
			Fsprite: spr_HDB5,
			Ftics:   -1,
		},
		955: {
			Fsprite: spr_HDB6,
			Ftics:   -1,
		},
		956: {
			Fsprite: spr_POB1,
			Ftics:   -1,
		},
		957: {
			Fsprite: spr_POB2,
			Ftics:   -1,
		},
		958: {
			Fsprite: spr_BRS1,
			Ftics:   -1,
		},
		959: {
			Fsprite:    spr_TLMP,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_TECHLAMP2,
		},
		960: {
			Fsprite:    spr_TLMP,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_TECHLAMP3,
		},
		961: {
			Fsprite:    spr_TLMP,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_TECHLAMP4,
		},
		962: {
			Fsprite:    spr_TLMP,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_TECHLAMP,
		},
		963: {
			Fsprite:    spr_TLP2,
			Fframe:     32768,
			Ftics:      4,
			Fnextstate: s_TECH2LAMP2,
		},
		964: {
			Fsprite:    spr_TLP2,
			Fframe:     32769,
			Ftics:      4,
			Fnextstate: s_TECH2LAMP3,
		},
		965: {
			Fsprite:    spr_TLP2,
			Fframe:     32770,
			Ftics:      4,
			Fnextstate: s_TECH2LAMP4,
		},
		966: {
			Fsprite:    spr_TLP2,
			Fframe:     32771,
			Ftics:      4,
			Fnextstate: s_TECH2LAMP,
		},
	}
}

func init() {
	mobjinfo = [137]mobjinfo_t{
		0: {
			Fdoomednum:    -1,
			Fspawnstate:   s_PLAY,
			Fspawnhealth:  100,
			Fseestate:     s_PLAY_RUN1,
			Fpainstate:    s_PLAY_PAIN,
			Fpainchance:   255,
			Fpainsound:    int32(sfx_plpain),
			Fmissilestate: s_PLAY_ATK1,
			Fdeathstate:   s_PLAY_DIE1,
			Fxdeathstate:  s_PLAY_XDIE1,
			Fdeathsound:   int32(sfx_pldeth),
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_DROPOFF | mf_PICKUP | mf_NOTDMATCH,
		},
		1: {
			Fdoomednum:    3004,
			Fspawnstate:   s_POSS_STND,
			Fspawnhealth:  20,
			Fseestate:     s_POSS_RUN1,
			Fseesound:     int32(sfx_posit1),
			Freactiontime: 8,
			Fattacksound:  int32(sfx_pistol),
			Fpainstate:    s_POSS_PAIN,
			Fpainchance:   200,
			Fpainsound:    int32(sfx_popain),
			Fmissilestate: s_POSS_ATK1,
			Fdeathstate:   s_POSS_DIE1,
			Fxdeathstate:  s_POSS_XDIE1,
			Fdeathsound:   int32(sfx_podth1),
			Fspeed:        8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         100,
			Factivesound:  int32(sfx_posact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_POSS_RAISE1,
		},
		2: {
			Fdoomednum:    9,
			Fspawnstate:   s_SPOS_STND,
			Fspawnhealth:  30,
			Fseestate:     s_SPOS_RUN1,
			Fseesound:     int32(sfx_posit2),
			Freactiontime: 8,
			Fpainstate:    s_SPOS_PAIN,
			Fpainchance:   170,
			Fpainsound:    int32(sfx_popain),
			Fmissilestate: s_SPOS_ATK1,
			Fdeathstate:   s_SPOS_DIE1,
			Fxdeathstate:  s_SPOS_XDIE1,
			Fdeathsound:   int32(sfx_podth2),
			Fspeed:        8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         100,
			Factivesound:  int32(sfx_posact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_SPOS_RAISE1,
		},
		3: {
			Fdoomednum:    64,
			Fspawnstate:   s_VILE_STND,
			Fspawnhealth:  700,
			Fseestate:     s_VILE_RUN1,
			Fseesound:     int32(sfx_vilsit),
			Freactiontime: 8,
			Fpainstate:    s_VILE_PAIN,
			Fpainchance:   10,
			Fpainsound:    int32(sfx_vipain),
			Fmissilestate: s_VILE_ATK1,
			Fdeathstate:   s_VILE_DIE1,
			Fdeathsound:   int32(sfx_vildth),
			Fspeed:        15,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         500,
			Factivesound:  int32(sfx_vilact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
		},
		4: {
			Fdoomednum:    -1,
			Fspawnstate:   s_FIRE1,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOGRAVITY,
		},
		5: {
			Fdoomednum:    66,
			Fspawnstate:   s_SKEL_STND,
			Fspawnhealth:  300,
			Fseestate:     s_SKEL_RUN1,
			Fseesound:     int32(sfx_skesit),
			Freactiontime: 8,
			Fpainstate:    s_SKEL_PAIN,
			Fpainchance:   100,
			Fpainsound:    int32(sfx_popain),
			Fmeleestate:   s_SKEL_FIST1,
			Fmissilestate: s_SKEL_MISS1,
			Fdeathstate:   s_SKEL_DIE1,
			Fdeathsound:   int32(sfx_skedth),
			Fspeed:        10,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         500,
			Factivesound:  int32(sfx_skeact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_SKEL_RAISE1,
		},
		6: {
			Fdoomednum:    -1,
			Fspawnstate:   s_TRACER,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_skeatk),
			Freactiontime: 8,
			Fdeathstate:   s_TRACEEXP1,
			Fdeathsound:   int32(sfx_barexp),
			Fspeed:        10 * (1 << FRACBITS),
			Fradius:       11 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       10,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		7: {
			Fdoomednum:    -1,
			Fspawnstate:   s_SMOKE1,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOGRAVITY,
		},
		8: {
			Fdoomednum:    67,
			Fspawnstate:   s_FATT_STND,
			Fspawnhealth:  600,
			Fseestate:     s_FATT_RUN1,
			Fseesound:     int32(sfx_mansit),
			Freactiontime: 8,
			Fpainstate:    s_FATT_PAIN,
			Fpainchance:   80,
			Fpainsound:    int32(sfx_mnpain),
			Fmissilestate: s_FATT_ATK1,
			Fdeathstate:   s_FATT_DIE1,
			Fdeathsound:   int32(sfx_mandth),
			Fspeed:        8,
			Fradius:       48 * (1 << FRACBITS),
			Fheight:       64 * (1 << FRACBITS),
			Fmass:         1000,
			Factivesound:  int32(sfx_posact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_FATT_RAISE1,
		},
		9: {
			Fdoomednum:    -1,
			Fspawnstate:   s_FATSHOT1,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_firsht),
			Freactiontime: 8,
			Fdeathstate:   s_FATSHOTX1,
			Fdeathsound:   int32(sfx_firxpl),
			Fspeed:        20 * (1 << FRACBITS),
			Fradius:       6 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       8,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		10: {
			Fdoomednum:    65,
			Fspawnstate:   s_CPOS_STND,
			Fspawnhealth:  70,
			Fseestate:     s_CPOS_RUN1,
			Fseesound:     int32(sfx_posit2),
			Freactiontime: 8,
			Fpainstate:    s_CPOS_PAIN,
			Fpainchance:   170,
			Fpainsound:    int32(sfx_popain),
			Fmissilestate: s_CPOS_ATK1,
			Fdeathstate:   s_CPOS_DIE1,
			Fxdeathstate:  s_CPOS_XDIE1,
			Fdeathsound:   int32(sfx_podth2),
			Fspeed:        8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         100,
			Factivesound:  int32(sfx_posact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_CPOS_RAISE1,
		},
		11: {
			Fdoomednum:    3001,
			Fspawnstate:   s_TROO_STND,
			Fspawnhealth:  60,
			Fseestate:     s_TROO_RUN1,
			Fseesound:     int32(sfx_bgsit1),
			Freactiontime: 8,
			Fpainstate:    s_TROO_PAIN,
			Fpainchance:   200,
			Fpainsound:    int32(sfx_popain),
			Fmeleestate:   s_TROO_ATK1,
			Fmissilestate: s_TROO_ATK1,
			Fdeathstate:   s_TROO_DIE1,
			Fxdeathstate:  s_TROO_XDIE1,
			Fdeathsound:   int32(sfx_bgdth1),
			Fspeed:        8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         100,
			Factivesound:  int32(sfx_bgact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_TROO_RAISE1,
		},
		12: {
			Fdoomednum:    3002,
			Fspawnstate:   s_SARG_STND,
			Fspawnhealth:  150,
			Fseestate:     s_SARG_RUN1,
			Fseesound:     int32(sfx_sgtsit),
			Freactiontime: 8,
			Fattacksound:  int32(sfx_sgtatk),
			Fpainstate:    s_SARG_PAIN,
			Fpainchance:   180,
			Fpainsound:    int32(sfx_dmpain),
			Fmeleestate:   s_SARG_ATK1,
			Fdeathstate:   s_SARG_DIE1,
			Fdeathsound:   int32(sfx_sgtdth),
			Fspeed:        10,
			Fradius:       30 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         400,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_SARG_RAISE1,
		},
		13: {
			Fdoomednum:    58,
			Fspawnstate:   s_SARG_STND,
			Fspawnhealth:  150,
			Fseestate:     s_SARG_RUN1,
			Fseesound:     int32(sfx_sgtsit),
			Freactiontime: 8,
			Fattacksound:  int32(sfx_sgtatk),
			Fpainstate:    s_SARG_PAIN,
			Fpainchance:   180,
			Fpainsound:    int32(sfx_dmpain),
			Fmeleestate:   s_SARG_ATK1,
			Fdeathstate:   s_SARG_DIE1,
			Fdeathsound:   int32(sfx_sgtdth),
			Fspeed:        10,
			Fradius:       30 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         400,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_SHADOW | mf_COUNTKILL,
			Fraisestate:   s_SARG_RAISE1,
		},
		14: {
			Fdoomednum:    3005,
			Fspawnstate:   s_HEAD_STND,
			Fspawnhealth:  400,
			Fseestate:     s_HEAD_RUN1,
			Fseesound:     int32(sfx_cacsit),
			Freactiontime: 8,
			Fpainstate:    s_HEAD_PAIN,
			Fpainchance:   128,
			Fpainsound:    int32(sfx_dmpain),
			Fmissilestate: s_HEAD_ATK1,
			Fdeathstate:   s_HEAD_DIE1,
			Fdeathsound:   int32(sfx_cacdth),
			Fspeed:        8,
			Fradius:       31 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         400,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_FLOAT | mf_NOGRAVITY | mf_COUNTKILL,
			Fraisestate:   s_HEAD_RAISE1,
		},
		15: {
			Fdoomednum:    3003,
			Fspawnstate:   s_BOSS_STND,
			Fspawnhealth:  1000,
			Fseestate:     s_BOSS_RUN1,
			Fseesound:     int32(sfx_brssit),
			Freactiontime: 8,
			Fpainstate:    s_BOSS_PAIN,
			Fpainchance:   50,
			Fpainsound:    int32(sfx_dmpain),
			Fmeleestate:   s_BOSS_ATK1,
			Fmissilestate: s_BOSS_ATK1,
			Fdeathstate:   s_BOSS_DIE1,
			Fdeathsound:   int32(sfx_brsdth),
			Fspeed:        8,
			Fradius:       24 * (1 << FRACBITS),
			Fheight:       64 * (1 << FRACBITS),
			Fmass:         1000,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_BOSS_RAISE1,
		},
		16: {
			Fdoomednum:    -1,
			Fspawnstate:   s_BRBALL1,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_firsht),
			Freactiontime: 8,
			Fdeathstate:   s_BRBALLX1,
			Fdeathsound:   int32(sfx_firxpl),
			Fspeed:        15 * (1 << FRACBITS),
			Fradius:       6 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       8,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		17: {
			Fdoomednum:    69,
			Fspawnstate:   s_BOS2_STND,
			Fspawnhealth:  500,
			Fseestate:     s_BOS2_RUN1,
			Fseesound:     int32(sfx_kntsit),
			Freactiontime: 8,
			Fpainstate:    s_BOS2_PAIN,
			Fpainchance:   50,
			Fpainsound:    int32(sfx_dmpain),
			Fmeleestate:   s_BOS2_ATK1,
			Fmissilestate: s_BOS2_ATK1,
			Fdeathstate:   s_BOS2_DIE1,
			Fdeathsound:   int32(sfx_kntdth),
			Fspeed:        8,
			Fradius:       24 * (1 << FRACBITS),
			Fheight:       64 * (1 << FRACBITS),
			Fmass:         1000,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_BOS2_RAISE1,
		},
		18: {
			Fdoomednum:    3006,
			Fspawnstate:   s_SKULL_STND,
			Fspawnhealth:  100,
			Fseestate:     s_SKULL_RUN1,
			Freactiontime: 8,
			Fattacksound:  int32(sfx_sklatk),
			Fpainstate:    s_SKULL_PAIN,
			Fpainchance:   256,
			Fpainsound:    int32(sfx_dmpain),
			Fmissilestate: s_SKULL_ATK1,
			Fdeathstate:   s_SKULL_DIE1,
			Fdeathsound:   int32(sfx_firxpl),
			Fspeed:        8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         50,
			Fdamage:       3,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_FLOAT | mf_NOGRAVITY,
		},
		19: {
			Fdoomednum:    7,
			Fspawnstate:   s_SPID_STND,
			Fspawnhealth:  3000,
			Fseestate:     s_SPID_RUN1,
			Fseesound:     int32(sfx_spisit),
			Freactiontime: 8,
			Fattacksound:  int32(sfx_shotgn),
			Fpainstate:    s_SPID_PAIN,
			Fpainchance:   40,
			Fpainsound:    int32(sfx_dmpain),
			Fmissilestate: s_SPID_ATK1,
			Fdeathstate:   s_SPID_DIE1,
			Fdeathsound:   int32(sfx_spidth),
			Fspeed:        12,
			Fradius:       128 * (1 << FRACBITS),
			Fheight:       100 * (1 << FRACBITS),
			Fmass:         1000,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
		},
		20: {
			Fdoomednum:    68,
			Fspawnstate:   s_BSPI_STND,
			Fspawnhealth:  500,
			Fseestate:     s_BSPI_SIGHT,
			Fseesound:     int32(sfx_bspsit),
			Freactiontime: 8,
			Fpainstate:    s_BSPI_PAIN,
			Fpainchance:   128,
			Fpainsound:    int32(sfx_dmpain),
			Fmissilestate: s_BSPI_ATK1,
			Fdeathstate:   s_BSPI_DIE1,
			Fdeathsound:   int32(sfx_bspdth),
			Fspeed:        12,
			Fradius:       64 * (1 << FRACBITS),
			Fheight:       64 * (1 << FRACBITS),
			Fmass:         600,
			Factivesound:  int32(sfx_bspact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_BSPI_RAISE1,
		},
		21: {
			Fdoomednum:    16,
			Fspawnstate:   s_CYBER_STND,
			Fspawnhealth:  4000,
			Fseestate:     s_CYBER_RUN1,
			Fseesound:     int32(sfx_cybsit),
			Freactiontime: 8,
			Fpainstate:    s_CYBER_PAIN,
			Fpainchance:   20,
			Fpainsound:    int32(sfx_dmpain),
			Fmissilestate: s_CYBER_ATK1,
			Fdeathstate:   s_CYBER_DIE1,
			Fdeathsound:   int32(sfx_cybdth),
			Fspeed:        16,
			Fradius:       40 * (1 << FRACBITS),
			Fheight:       110 * (1 << FRACBITS),
			Fmass:         1000,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
		},
		22: {
			Fdoomednum:    71,
			Fspawnstate:   s_PAIN_STND,
			Fspawnhealth:  400,
			Fseestate:     s_PAIN_RUN1,
			Fseesound:     int32(sfx_pesit),
			Freactiontime: 8,
			Fpainstate:    s_PAIN_PAIN,
			Fpainchance:   128,
			Fpainsound:    int32(sfx_pepain),
			Fmissilestate: s_PAIN_ATK1,
			Fdeathstate:   s_PAIN_DIE1,
			Fdeathsound:   int32(sfx_pedth),
			Fspeed:        8,
			Fradius:       31 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         400,
			Factivesound:  int32(sfx_dmact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_FLOAT | mf_NOGRAVITY | mf_COUNTKILL,
			Fraisestate:   s_PAIN_RAISE1,
		},
		23: {
			Fdoomednum:    84,
			Fspawnstate:   s_SSWV_STND,
			Fspawnhealth:  50,
			Fseestate:     s_SSWV_RUN1,
			Fseesound:     int32(sfx_sssit),
			Freactiontime: 8,
			Fpainstate:    s_SSWV_PAIN,
			Fpainchance:   170,
			Fpainsound:    int32(sfx_popain),
			Fmissilestate: s_SSWV_ATK1,
			Fdeathstate:   s_SSWV_DIE1,
			Fxdeathstate:  s_SSWV_XDIE1,
			Fdeathsound:   int32(sfx_ssdth),
			Fspeed:        8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       56 * (1 << FRACBITS),
			Fmass:         100,
			Factivesound:  int32(sfx_posact),
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_COUNTKILL,
			Fraisestate:   s_SSWV_RAISE1,
		},
		24: {
			Fdoomednum:    72,
			Fspawnstate:   s_KEENSTND,
			Fspawnhealth:  100,
			Freactiontime: 8,
			Fpainstate:    s_KEENPAIN,
			Fpainchance:   256,
			Fpainsound:    int32(sfx_keenpn),
			Fdeathstate:   s_COMMKEEN,
			Fdeathsound:   int32(sfx_keendt),
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       72 * (1 << FRACBITS),
			Fmass:         10000000,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY | mf_SHOOTABLE | mf_COUNTKILL,
		},
		25: {
			Fdoomednum:    88,
			Fspawnstate:   s_BRAIN,
			Fspawnhealth:  250,
			Freactiontime: 8,
			Fpainstate:    s_BRAIN_PAIN,
			Fpainchance:   255,
			Fpainsound:    int32(sfx_bospn),
			Fdeathstate:   s_BRAIN_DIE1,
			Fdeathsound:   int32(sfx_bosdth),
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         10000000,
			Fflags:        mf_SOLID | mf_SHOOTABLE,
		},
		26: {
			Fdoomednum:    89,
			Fspawnstate:   s_BRAINEYE,
			Fspawnhealth:  1000,
			Fseestate:     s_BRAINEYESEE,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       32 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOSECTOR,
		},
		27: {
			Fdoomednum:    87,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       32 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOSECTOR,
		},
		28: {
			Fdoomednum:    -1,
			Fspawnstate:   s_SPAWN1,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_bospit),
			Freactiontime: 8,
			Fdeathsound:   int32(sfx_firxpl),
			Fspeed:        10 * (1 << FRACBITS),
			Fradius:       6 * (1 << FRACBITS),
			Fheight:       32 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       3,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY | mf_NOCLIP,
		},
		29: {
			Fdoomednum:    -1,
			Fspawnstate:   s_SPAWNFIRE1,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOGRAVITY,
		},
		30: {
			Fdoomednum:    2035,
			Fspawnstate:   s_BAR1,
			Fspawnhealth:  20,
			Freactiontime: 8,
			Fdeathstate:   s_BEXP,
			Fdeathsound:   int32(sfx_barexp),
			Fradius:       10 * (1 << FRACBITS),
			Fheight:       42 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SHOOTABLE | mf_NOBLOOD,
		},
		31: {
			Fdoomednum:    -1,
			Fspawnstate:   s_TBALL1,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_firsht),
			Freactiontime: 8,
			Fdeathstate:   s_TBALLX1,
			Fdeathsound:   int32(sfx_firxpl),
			Fspeed:        10 * (1 << FRACBITS),
			Fradius:       6 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       3,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		32: {
			Fdoomednum:    -1,
			Fspawnstate:   s_RBALL1,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_firsht),
			Freactiontime: 8,
			Fdeathstate:   s_RBALLX1,
			Fdeathsound:   int32(sfx_firxpl),
			Fspeed:        10 * (1 << FRACBITS),
			Fradius:       6 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       5,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		33: {
			Fdoomednum:    -1,
			Fspawnstate:   s_ROCKET,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_rlaunc),
			Freactiontime: 8,
			Fdeathstate:   s_EXPLODE1,
			Fdeathsound:   int32(sfx_barexp),
			Fspeed:        20 * (1 << FRACBITS),
			Fradius:       11 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       20,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		34: {
			Fdoomednum:    -1,
			Fspawnstate:   s_PLASBALL,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_plasma),
			Freactiontime: 8,
			Fdeathstate:   s_PLASEXP,
			Fdeathsound:   int32(sfx_firxpl),
			Fspeed:        25 * (1 << FRACBITS),
			Fradius:       13 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       5,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		35: {
			Fdoomednum:    -1,
			Fspawnstate:   s_BFGSHOT,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fdeathstate:   s_BFGLAND,
			Fdeathsound:   int32(sfx_rxplod),
			Fspeed:        25 * (1 << FRACBITS),
			Fradius:       13 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       100,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		36: {
			Fdoomednum:    -1,
			Fspawnstate:   s_ARACH_PLAZ,
			Fspawnhealth:  1000,
			Fseesound:     int32(sfx_plasma),
			Freactiontime: 8,
			Fdeathstate:   s_ARACH_PLEX,
			Fdeathsound:   int32(sfx_firxpl),
			Fspeed:        25 * (1 << FRACBITS),
			Fradius:       13 * (1 << FRACBITS),
			Fheight:       8 * (1 << FRACBITS),
			Fmass:         100,
			Fdamage:       5,
			Fflags:        mf_NOBLOCKMAP | mf_MISSILE | mf_DROPOFF | mf_NOGRAVITY,
		},
		37: {
			Fdoomednum:    -1,
			Fspawnstate:   s_PUFF1,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOGRAVITY,
		},
		38: {
			Fdoomednum:    -1,
			Fspawnstate:   s_BLOOD1,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP,
		},
		39: {
			Fdoomednum:    -1,
			Fspawnstate:   s_TFOG,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOGRAVITY,
		},
		40: {
			Fdoomednum:    -1,
			Fspawnstate:   s_IFOG,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOGRAVITY,
		},
		41: {
			Fdoomednum:    14,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOSECTOR,
		},
		42: {
			Fdoomednum:    -1,
			Fspawnstate:   s_BFGEXP,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP | mf_NOGRAVITY,
		},
		43: {
			Fdoomednum:    2018,
			Fspawnstate:   s_ARM1,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		44: {
			Fdoomednum:    2019,
			Fspawnstate:   s_ARM2,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		45: {
			Fdoomednum:    2014,
			Fspawnstate:   s_BON1,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		46: {
			Fdoomednum:    2015,
			Fspawnstate:   s_BON2,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		47: {
			Fdoomednum:    5,
			Fspawnstate:   s_BKEY,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_NOTDMATCH,
		},
		48: {
			Fdoomednum:    13,
			Fspawnstate:   s_RKEY,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_NOTDMATCH,
		},
		49: {
			Fdoomednum:    6,
			Fspawnstate:   s_YKEY,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_NOTDMATCH,
		},
		50: {
			Fdoomednum:    39,
			Fspawnstate:   s_YSKULL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_NOTDMATCH,
		},
		51: {
			Fdoomednum:    38,
			Fspawnstate:   s_RSKULL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_NOTDMATCH,
		},
		52: {
			Fdoomednum:    40,
			Fspawnstate:   s_BSKULL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_NOTDMATCH,
		},
		53: {
			Fdoomednum:    2011,
			Fspawnstate:   s_STIM,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		54: {
			Fdoomednum:    2012,
			Fspawnstate:   s_MEDI,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		55: {
			Fdoomednum:    2013,
			Fspawnstate:   s_SOUL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		56: {
			Fdoomednum:    2022,
			Fspawnstate:   s_PINV,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		57: {
			Fdoomednum:    2023,
			Fspawnstate:   s_PSTR,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		58: {
			Fdoomednum:    2024,
			Fspawnstate:   s_PINS,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		59: {
			Fdoomednum:    2025,
			Fspawnstate:   s_SUIT,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		60: {
			Fdoomednum:    2026,
			Fspawnstate:   s_PMAP,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		61: {
			Fdoomednum:    2045,
			Fspawnstate:   s_PVIS,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		62: {
			Fdoomednum:    83,
			Fspawnstate:   s_MEGA,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL | mf_COUNTITEM,
		},
		63: {
			Fdoomednum:    2007,
			Fspawnstate:   s_CLIP,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		64: {
			Fdoomednum:    2048,
			Fspawnstate:   s_AMMO,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		65: {
			Fdoomednum:    2010,
			Fspawnstate:   s_ROCK,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		66: {
			Fdoomednum:    2046,
			Fspawnstate:   s_BROK,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		67: {
			Fdoomednum:    2047,
			Fspawnstate:   s_CELL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		68: {
			Fdoomednum:    17,
			Fspawnstate:   s_CELP,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		69: {
			Fdoomednum:    2008,
			Fspawnstate:   s_SHEL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		70: {
			Fdoomednum:    2049,
			Fspawnstate:   s_SBOX,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		71: {
			Fdoomednum:    8,
			Fspawnstate:   s_BPAK,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		72: {
			Fdoomednum:    2006,
			Fspawnstate:   s_BFUG,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		73: {
			Fdoomednum:    2002,
			Fspawnstate:   s_MGUN,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		74: {
			Fdoomednum:    2005,
			Fspawnstate:   s_CSAW,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		75: {
			Fdoomednum:    2003,
			Fspawnstate:   s_LAUN,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		76: {
			Fdoomednum:    2004,
			Fspawnstate:   s_PLAS,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		77: {
			Fdoomednum:    2001,
			Fspawnstate:   s_SHOT,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		78: {
			Fdoomednum:    82,
			Fspawnstate:   s_SHOT2,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPECIAL,
		},
		79: {
			Fdoomednum:    85,
			Fspawnstate:   s_TECHLAMP,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		80: {
			Fdoomednum:    86,
			Fspawnstate:   s_TECH2LAMP,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		81: {
			Fdoomednum:    2028,
			Fspawnstate:   s_COLU,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		82: {
			Fdoomednum:    30,
			Fspawnstate:   s_TALLGRNCOL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		83: {
			Fdoomednum:    31,
			Fspawnstate:   s_SHRTGRNCOL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		84: {
			Fdoomednum:    32,
			Fspawnstate:   s_TALLREDCOL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		85: {
			Fdoomednum:    33,
			Fspawnstate:   s_SHRTREDCOL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		86: {
			Fdoomednum:    37,
			Fspawnstate:   s_SKULLCOL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		87: {
			Fdoomednum:    36,
			Fspawnstate:   s_HEARTCOL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		88: {
			Fdoomednum:    41,
			Fspawnstate:   s_EVILEYE,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		89: {
			Fdoomednum:    42,
			Fspawnstate:   s_FLOATSKULL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		90: {
			Fdoomednum:    43,
			Fspawnstate:   s_TORCHTREE,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		91: {
			Fdoomednum:    44,
			Fspawnstate:   s_BLUETORCH,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		92: {
			Fdoomednum:    45,
			Fspawnstate:   s_GREENTORCH,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		93: {
			Fdoomednum:    46,
			Fspawnstate:   s_REDTORCH,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		94: {
			Fdoomednum:    55,
			Fspawnstate:   s_BTORCHSHRT,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		95: {
			Fdoomednum:    56,
			Fspawnstate:   s_GTORCHSHRT,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		96: {
			Fdoomednum:    57,
			Fspawnstate:   s_RTORCHSHRT,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		97: {
			Fdoomednum:    47,
			Fspawnstate:   s_STALAGTITE,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		98: {
			Fdoomednum:    48,
			Fspawnstate:   s_TECHPILLAR,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		99: {
			Fdoomednum:    34,
			Fspawnstate:   s_CANDLESTIK,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		100: {
			Fdoomednum:    35,
			Fspawnstate:   s_CANDELABRA,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		101: {
			Fdoomednum:    49,
			Fspawnstate:   s_BLOODYTWITCH,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       68 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		102: {
			Fdoomednum:    50,
			Fspawnstate:   s_MEAT2,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       84 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		103: {
			Fdoomednum:    51,
			Fspawnstate:   s_MEAT3,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       84 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		104: {
			Fdoomednum:    52,
			Fspawnstate:   s_MEAT4,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       68 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		105: {
			Fdoomednum:    53,
			Fspawnstate:   s_MEAT5,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       52 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		106: {
			Fdoomednum:    59,
			Fspawnstate:   s_MEAT2,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       84 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		107: {
			Fdoomednum:    60,
			Fspawnstate:   s_MEAT4,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       68 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		108: {
			Fdoomednum:    61,
			Fspawnstate:   s_MEAT3,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       52 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		109: {
			Fdoomednum:    62,
			Fspawnstate:   s_MEAT5,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       52 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		110: {
			Fdoomednum:    63,
			Fspawnstate:   s_BLOODYTWITCH,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       68 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		111: {
			Fdoomednum:    22,
			Fspawnstate:   s_HEAD_DIE6,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		112: {
			Fdoomednum:    15,
			Fspawnstate:   s_PLAY_DIE7,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		113: {
			Fdoomednum:    18,
			Fspawnstate:   s_POSS_DIE5,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		114: {
			Fdoomednum:    21,
			Fspawnstate:   s_SARG_DIE6,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		115: {
			Fdoomednum:    23,
			Fspawnstate:   s_SKULL_DIE6,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		116: {
			Fdoomednum:    20,
			Fspawnstate:   s_TROO_DIE5,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		117: {
			Fdoomednum:    19,
			Fspawnstate:   s_SPOS_DIE5,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		118: {
			Fdoomednum:    10,
			Fspawnstate:   s_PLAY_XDIE9,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		119: {
			Fdoomednum:    12,
			Fspawnstate:   s_PLAY_XDIE9,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		120: {
			Fdoomednum:    28,
			Fspawnstate:   s_HEADSONSTICK,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		121: {
			Fdoomednum:    24,
			Fspawnstate:   s_GIBS,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
		},
		122: {
			Fdoomednum:    27,
			Fspawnstate:   s_HEADONASTICK,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		123: {
			Fdoomednum:    29,
			Fspawnstate:   s_HEADCANDLES,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		124: {
			Fdoomednum:    25,
			Fspawnstate:   s_DEADSTICK,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		125: {
			Fdoomednum:    26,
			Fspawnstate:   s_LIVESTICK,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		126: {
			Fdoomednum:    54,
			Fspawnstate:   s_BIGTREE,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       32 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		127: {
			Fdoomednum:    70,
			Fspawnstate:   s_BBAR1,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID,
		},
		128: {
			Fdoomednum:    73,
			Fspawnstate:   s_HANGNOGUTS,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       88 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		129: {
			Fdoomednum:    74,
			Fspawnstate:   s_HANGBNOBRAIN,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       88 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		130: {
			Fdoomednum:    75,
			Fspawnstate:   s_HANGTLOOKDN,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       64 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		131: {
			Fdoomednum:    76,
			Fspawnstate:   s_HANGTSKULL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       64 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		132: {
			Fdoomednum:    77,
			Fspawnstate:   s_HANGTLOOKUP,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       64 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		133: {
			Fdoomednum:    78,
			Fspawnstate:   s_HANGTNOBRAIN,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       16 * (1 << FRACBITS),
			Fheight:       64 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_SOLID | mf_SPAWNCEILING | mf_NOGRAVITY,
		},
		134: {
			Fdoomednum:    79,
			Fspawnstate:   s_COLONGIBS,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP,
		},
		135: {
			Fdoomednum:    80,
			Fspawnstate:   s_SMALLPOOL,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP,
		},
		136: {
			Fdoomednum:    81,
			Fspawnstate:   s_BRAINSTEM,
			Fspawnhealth:  1000,
			Freactiontime: 8,
			Fradius:       20 * (1 << FRACBITS),
			Fheight:       16 * (1 << FRACBITS),
			Fmass:         100,
			Fflags:        mf_NOBLOCKMAP,
		},
	}
}

//
// Displays the text mode ending screen after the game quits
//

func i_Endoom(endoom_data uintptr) {
}

// When an axis is within the dead zone, it is set to zero.
// This is 5% of the full range:

// Configuration variables:

// Standard default.cfg Joystick enable/disable

var usejoystick = 0

// Joystick to use, as an SDL joystick index:

var joystick_index = -1

// Which joystick axis to use for horizontal movement, and whether to
// invert the direction:

var joystick_x_axis = 0
var joystick_x_invert = 0

// Which joystick axis to use for vertical movement, and whether to
// invert the direction:

var joystick_y_axis int32 = 1
var joystick_y_invert int32 = 0

// Which joystick axis to use for strafing?

var joystick_strafe_axis = -1
var joystick_strafe_invert = 0

// C documentation
//
//	// Virtual to physical button joystick button mapping. By default this
//	// is a straight mapping.
var joystick_physical_buttons = [10]int32{
	1: 1,
	2: 2,
	3: 3,
	4: 4,
	5: 5,
	6: 6,
	7: 7,
	8: 8,
	9: 9,
}

func i_BindJoystickVariables() {
	m_BindVariable("use_joystick", &usejoystick)
	m_BindVariable("joystick_index", &joystick_index)
	m_BindVariable("joystick_x_axis", &joystick_x_axis)
	m_BindVariable("joystick_y_axis", &joystick_y_axis)
	m_BindVariable("joystick_strafe_axis", &joystick_strafe_axis)
	m_BindVariable("joystick_x_invert", &joystick_x_invert)
	m_BindVariable("joystick_y_invert", &joystick_y_invert)
	m_BindVariable("joystick_strafe_invert", &joystick_strafe_invert)
	for i := range NUM_VIRTUAL_BUTTONS {
		name := fmt.Sprintf("joystick_physical_button%d", i)
		m_BindVariable(name, &joystick_physical_buttons[i])
	}
}

// 1x scale doesn't really do any scaling: it just copies the buffer
// a line at a time for when pitch != SCREENWIDTH (!native_surface)

func init() {
	snd_samplerate = 44100
}

func init() {
	snd_cachesize = 64 * 1024 * 1024
}

func init() {
	snd_maxslicetime_ms = 28
}

func init() {
	snd_musiccmd = ""
}

// Low-level sound and music modules we are using

var sound_module *sound_module_t
var music_module *music_module_t

func init() {
	snd_musicdevice = SNDDEVICE_SB
}

func init() {
	snd_sfxdevice = SNDDEVICE_SB
}

// DOS-specific options: These are unused but should be maintained
// so that the config file can be shared between chocolate
// doom and doom.exe

var snd_sbport = 0
var snd_sbirq = 0
var snd_sbdma = 0
var snd_mport = 0

// Compiled-in sound modules:

var sound_modules = []sound_module_t{}

// Check if a sound device is in the given list of devices

func sndDeviceInList(device snddevice_t, list []snddevice_t, len1 int32) boolean {
	for i := range len1 {
		if device == list[i] {
			return 1
		}
	}
	return 0
}

// Find and initialize a sound_module_t appropriate for the setting
// in snd_sfxdevice.

func initSfxModule(use_sfx_prefix boolean) {
	for i := range sound_modules {
		s := &sound_modules[i]
		// Is the sfx device in the list of devices supported by
		// this module?
		if sndDeviceInList(snd_sfxdevice, s.Fsound_devices, s.Fnum_sound_devices) != 0 {
			// Initialize the module
			if s.FInit(use_sfx_prefix) != 0 {
				sound_module = s
				return
			}
		}
	}
}

// Initialize music according to snd_musicdevice.

func initMusicModule() {
}

//
// Initializes sound stuff, including volume
// Sets channels, SFX and music volume,
//  allocates channel buffer, sets S_sfx lookup.
//

func i_InitSound(use_sfx_prefix boolean) {
	var nomusic, nosfx, nosound boolean
	//!
	// @vanilla
	//
	// Disable all sound output.
	//
	nosound = booluint32(m_CheckParm("-nosound") > 0)
	//!
	// @vanilla
	//
	// Disable sound effects.
	//
	nosfx = booluint32(m_CheckParm("-nosfx") > 0)
	//!
	// @vanilla
	//
	// Disable music.
	//
	nomusic = booluint32(m_CheckParm("-nomusic") > 0)
	// Initialize the sound and music subsystems.
	if nosound == 0 && screensaver_mode == 0 {
		// This is kind of a hack. If native MIDI is enabled, set up
		// the TIMIDITY_CFG environment variable here before SDL_mixer
		// is opened.
		if nomusic == 0 && (snd_musicdevice == SNDDEVICE_GENMIDI || snd_musicdevice == SNDDEVICE_GUS) {
			//I_InitTimidityConfig();
		}
		if nosfx == 0 {
			initSfxModule(use_sfx_prefix)
		}
		if nomusic == 0 {
			initMusicModule()
		}
	}
}

func i_ShutdownSound() {
	if sound_module != nil {
		sound_module.FShutdown()
	}
	if music_module != nil {
		music_module.FShutdown()
	}
}

func i_GetSfxLumpNum(sfxinfo *sfxinfo_t) int32 {
	if sound_module != nil {
		return sound_module.FGetSfxLumpNum(sfxinfo)
	} else {
		return 0
	}
}

func i_UpdateSound() {
	if sound_module != nil {
		sound_module.FUpdate()
	}
	if music_module != nil && music_module.FPoll != nil {
		music_module.FPoll()
	}
}

func checkVolumeSeparation(vol *int32, sep *int32) {
	if *sep < 0 {
		*sep = 0
	} else {
		if *sep > 254 {
			*sep = 254
		}
	}
	if *vol < 0 {
		*vol = 0
	} else {
		if *vol > 127 {
			*vol = 127
		}
	}
}

func i_UpdateSoundParams(channel int32, vol int32, sep int32) {
	if sound_module != nil {
		checkVolumeSeparation(&vol, &sep)
		sound_module.FUpdateSoundParams(channel, vol, sep)
	}
}

func i_StartSound(sfxinfo *sfxinfo_t, channel int32, vol int32, sep int32) int32 {
	if sound_module != nil {
		checkVolumeSeparation(&vol, &sep)
		return sound_module.FStartSound(sfxinfo, channel, vol, sep)
	}
	return 0
}

func i_StopSound(channel int32) {
	if sound_module != nil {
		sound_module.FStopSound(channel)
	}
}

func i_SoundIsPlaying(channel int32) boolean {
	if sound_module != nil {
		return sound_module.FSoundIsPlaying(channel)
	}
	return 0
}

func i_PrecacheSounds(sounds []sfxinfo_t) {
	if sound_module != nil && sound_module.FCacheSounds != nil {
		sound_module.FCacheSounds(sounds)
	}
}

func i_InitMusic() {
	if music_module != nil {
		music_module.FInit()
	}
}

func i_ShutdownMusic() {
}

func i_SetMusicVolume(volume int32) {
	if music_module != nil {
		music_module.FSetMusicVolume(volume)
	}
}

func i_PauseSong() {
	if music_module != nil {
		music_module.FPauseMusic()
	}
}

func i_ResumeSong() {
	if music_module != nil {
		music_module.FResumeMusic()
	}
}

func i_RegisterSong(data []byte) uintptr {
	if music_module != nil {
		music_module.FRegisterSong(data)
	}
	return 0
}

func i_UnRegisterSong(handle uintptr) {
	if music_module != nil {
		music_module.FUnRegisterSong(handle)
	}
}

func i_PlaySong(handle uintptr, looping boolean) {
	if music_module != nil {
		music_module.FPlaySong(handle, looping)
	}
}

func i_StopSong() {
	if music_module != nil {
		music_module.FStopSong()
	}
}

func i_BindSoundVariables() {
	m_BindVariable("snd_musicdevice", &snd_musicdevice)
	m_BindVariable("snd_sfxdevice", &snd_sfxdevice)
	m_BindVariable("snd_sbport", &snd_sbport)
	m_BindVariable("snd_sbirq", &snd_sbirq)
	m_BindVariable("snd_sbdma", &snd_sbdma)
	m_BindVariable("snd_mport", &snd_mport)
	m_BindVariable("snd_maxslicetime_ms", &snd_maxslicetime_ms)
	m_BindVariable("snd_musiccmd", &snd_musiccmd)
	m_BindVariable("snd_samplerate", &snd_samplerate)
	m_BindVariable("snd_cachesize", &snd_cachesize)
	// Before SDL_mixer version 1.2.11, MIDI music caused the game
	// to crash when it looped.  If this is an old SDL_mixer version,
	// disable MIDI.
}

const DEFAULT_RAM = 16
const DOS_MEM_DUMP_SIZE = 10
const MIN_RAM = 16

//
// This is used to get the local FILE:LINE info from CPP
// prior to really call the function in question.
//

type atexit_listentry_t struct {
	Ffunc         func()
	Frun_on_error boolean
}

var exit_funcs []atexit_listentry_t

func i_AtExit(func1 func(), run_on_error boolean) {
	exit_funcs = append(exit_funcs, atexit_listentry_t{
		Ffunc:         func1,
		Frun_on_error: run_on_error,
	})
}

// Tactile feedback function, probably used for the Logitech Cyberman

func i_Tactile(on int32, off int32, total int32) {
}

func i_PrintBanner(msg string) {
	spaces := 35 - len(msg)/2
	for i := 0; i < spaces; i++ {
		fmt.Print(" ")
	}
	fmt.Printf("%s\n", msg)
}

func i_PrintDivider() {
	for range 75 {
		fmt.Print("=")
	}
	fmt.Print("\n")
}

func i_PrintStartupBanner(gamedescription string) {
	i_PrintDivider()
	i_PrintBanner(gamedescription)
	i_PrintDivider()
	fprintf_ccgo(os.Stdout, " Doom Generic is free software, covered by the GNU General Public\n License.  There is NO warranty; not even for MERCHANTABILITY or FITNESS\n FOR A PARTICULAR PURPOSE. You are welcome to change and distribute\n copies under certain conditions. See the source for more information.\n")
	i_PrintDivider()
}

//
// I_ConsoleStdout
//
// Returns true if stdout is a real console, false if it is a file
//

func i_ConsoleStdout() boolean {
	return 0
}

//
// I_Quit
//

func i_Quit() {
	// Run through all exit functions, from last to first
	for i := len(exit_funcs) - 1; i >= 0; i-- {
		// Call the exit function.
		exit_funcs[i].Ffunc()
	}
}

//
// I_Error
//

var already_quitting = 0

func i_Error(errStr string, args ...any) {
	var exit_gui_popup boolean
	if already_quitting != 0 {
		fprintf_ccgo(os.Stderr, "Warning: recursive call to i_Error detected.\n")
	} else {
		already_quitting = 1
	}
	fmt.Fprintf(os.Stderr, errStr, args...)
	fprintf_ccgo(os.Stderr, "\n\n")

	debug.PrintStack()

	// Shutdown. Here might be other errors.
	for i := len(exit_funcs) - 1; i >= 0; i-- {
		// Call the exit function.

		if exit_funcs[i].Frun_on_error != 0 {
			exit_funcs[i].Ffunc()
		}
	}
	exit_gui_popup = booluint32(m_ParmExists("-nogui") == 0)
	// Pop up a GUI dialog box to show the error message, if the
	// game was not run from the console (and the user will
	// therefore be unable to otherwise see the message).
	if exit_gui_popup != 0 && i_ConsoleStdout() == 0 {
		// TODO: Expose error message somehow?
	}
	// abort();
	for 1 != 0 {
	}
}

//
// Read Access Violation emulation.
//
// From PrBoom+, by entryway.
//

// C:\>debug
// -d 0:0
//
// DOS 6.22:
// 0000:0000  (57 92 19 00) F4 06 70 00-(16 00)
// DOS 7.1:
// 0000:0000  (9E 0F C9 00) 65 04 70 00-(16 00)
// Win98:
// 0000:0000  (9E 0F C9 00) 65 04 70 00-(16 00)
// DOSBox under XP:
// 0000:0000  (00 00 00 F1) ?? ?? ?? 00-(07 00)

var mem_dump_dos622 = [DOS_MEM_DUMP_SIZE]uint8{0x57, 0x92, 0x19, 0x00, 0xF4, 0x06, 0x70, 0x00, 0x16, 0x00}
var mem_dump_win98 = [DOS_MEM_DUMP_SIZE]uint8{0x9e, 0x0f, 0xc9, 0x00, 0x65, 0x04, 0x70, 0x00, 0x16, 0x00}
var mem_dump_dosbox = [DOS_MEM_DUMP_SIZE]uint8{0x00, 0x00, 0x00, 0xf1, 0x00, 0x00, 0x00, 0x00, 0x07, 0x00}
var mem_dump_custom [DOS_MEM_DUMP_SIZE]uint8

var dos_mem_dump []byte = mem_dump_dos622[:]

func i_GetMemoryValue(offset uint32, value uintptr, size int32) boolean {
	var p int32
	if firsttime != 0 {
		firsttime = 0
		//!
		// @category compat
		// @arg <version>
		//
		// Specify DOS version to emulate for NULL pointer dereference
		// emulation.  Supported versions are: dos622, dos71, dosbox.
		// The default is to emulate DOS 7.1 (Windows 98).
		//
		p = m_CheckParmWithArgs("-setmem", 1)
		if p > 0 {
			if strings.EqualFold(myargs[p+1], "dos622") {
				dos_mem_dump = mem_dump_dos622[:]
			}
			if strings.EqualFold(myargs[p+1], "dos71") {
				dos_mem_dump = mem_dump_win98[:]
			} else {
				if strings.EqualFold(myargs[p+1], "dosbox") {
					dos_mem_dump = mem_dump_dosbox[:]
				} else {
					for i := range DOS_MEM_DUMP_SIZE {
						p++
						if p >= int32(len(myargs)) || myargs[p][0] == '-' {
							break
						}
						f, _ := strconv.Atoi(myargs[p])
						mem_dump_custom[i] = uint8(f)
					}
					dos_mem_dump = mem_dump_custom[:]
				}
			}
		}
	}
	switch size {
	case 1:
		*(*uint8)(unsafe.Pointer(value)) = dos_mem_dump[offset]
		return 1
	case 2:
		*(*uint16)(unsafe.Pointer(value)) = uint16(int32(dos_mem_dump[offset]) | int32(dos_mem_dump[offset+1])<<8)
		return 1
	case 4:
		*(*uint32)(unsafe.Pointer(value)) = uint32(int32(dos_mem_dump[offset]) | int32(dos_mem_dump[offset+1])<<8 | int32(dos_mem_dump[offset+2])<<16 | int32(dos_mem_dump[offset+3])<<24)
		return 1
	}
	return 0
}

var firsttime = 1

var basetime uint32 = 0

var last_tick int32 = 0

func i_GetTicks() int32 {
	if dg_run_full_speed {
		// Just increment by 1 each frame
		return int32(dg_fake_tics)
	}
	return int32(float64(time.Since(start_time).Milliseconds()))
}

func i_GetTime() int32 {
	var ticks uint32
	ticks = uint32(i_GetTicks())
	if basetime == 0 {
		basetime = ticks
	}
	ticks -= basetime
	return int32(ticks * TICRATE / 1000)
}

//
// Same as i_GetTime, but returns time in milliseconds
//

func I_GetTimeMS() int32 {
	var ticks uint32
	ticks = uint32(i_GetTicks())
	if basetime == 0 {
		basetime = ticks
	}
	return int32(ticks - basetime)
}

// Sleep for a specified number of ms

func i_Sleep(ms uint32) {
	if dg_run_full_speed {
		dg_fake_tics++
	} else {
		time.Sleep(time.Duration(float64(ms)) * time.Millisecond)
	}
}

//
// M_CheckParm
// Checks for the given parameter
// in the program's command line arguments.
// Returns the argument number (1 to argc-1)
// or 0 if not present
//

func m_CheckParmWithArgs(check string, num_args int32) int32 {
	for i := int32(1); i < int32(len(myargs))-num_args; i++ {
		if strings.EqualFold(myargs[i], check) {
			return i
		}
	}
	return 0
}

//
// M_ParmExists
//
// Returns true if the given parameter exists in the program's command
// line arguments, false if not.
//

func m_ParmExists(check string) boolean {
	return booluint32(m_CheckParm(check) != 0)
}

func m_CheckParm(check string) int32 {
	return m_CheckParmWithArgs(check, 0)
}

func loadResponseFile(argv_index int32) {
}

//
// Find a Response File
//

func m_FindResponseFile() {
	for i := int32(1); i < int32(len(myargs)); i++ {
		if myargs[i][0] == '@' {
			loadResponseFile(i)
		}
	}
}

const INT_MAX5 = 2147483647

const BOXTOP = 0
const BOXBOTTOM = 1
const BOXLEFT = 2
const BOXRIGHT = 3

func m_ClearBox(box *box_t) {
	box[BOXRIGHT] = -1 - 0x7fffffff
	box[BOXRIGHT] = -1 - 0x7fffffff
	box[BOXLEFT] = INT_MAX5
	box[BOXBOTTOM] = INT_MAX5
}

func m_AddToBox(box *box_t, x fixed_t, y fixed_t) {
	if x < box[BOXLEFT] {
		box[BOXLEFT] = x
	} else {
		if x > box[BOXRIGHT] {
			box[BOXRIGHT] = x
		}
	}
	if y < box[BOXBOTTOM] {
		box[BOXBOTTOM] = y
	} else {
		if y > box[BOXTOP] {
			box[BOXTOP] = y
		}
	}
}

//
// CHEAT SEQUENCE PACKAGE
//

// C documentation
//
//	//
//	// Called in st_stuff module, which handles the input.
//	// Returns a 1 if the cheat was successful, 0 if failed.
//	//
func cht_CheckCheat(cht *cheatseq_t, key int8) int32 {
	var v1 int32
	// if we make a short sequence on a cheat with parameters, this
	// will not work in vanilla doom.  behave the same.
	if cht.Fparameter_chars > 0 && uint64(len(cht.Fsequence)) < cht.Fsequence_len {
		return 0
	}
	if cht.Fchars_read < uint64(len(cht.Fsequence)) {
		// still reading characters from the cheat code
		// and verifying.  reset back to the beginning
		// if a key is wrong
		if int32(key) == int32(cht.Fsequence[cht.Fchars_read]) {
			cht.Fchars_read++
		} else {
			cht.Fchars_read = 0
		}
		cht.Fparam_chars_read = 0
	} else {
		if cht.Fparam_chars_read < cht.Fparameter_chars {
			// we have passed the end of the cheat sequence and are
			// entering parameters now
			cht.Fparameter_buf[cht.Fparam_chars_read] = byte(key)
			cht.Fparam_chars_read++
		}
	}
	if cht.Fchars_read >= uint64(len(cht.Fsequence)) && cht.Fparam_chars_read >= cht.Fparameter_chars {
		v1 = 0
		cht.Fparam_chars_read = v1
		cht.Fchars_read = uint64(v1)
		return 1
	}
	// cheat not matched yet
	return 0
}

func cht_GetParam(cht *cheatseq_t, buffer []byte) {
	copy(buffer, cht.Fparameter_buf[:cht.Fparameter_chars])
}

const EISDIR = 21
const KEY_BACKSPACE3 = 127
const KEY_DOWNARROW1 = 175
const KEY_LEFTARROW1 = 172
const KEY_MINUS1 = 45
const KEY_PAUSE1 = 255
const KEY_RIGHTARROW1 = 174
const KEY_UPARROW1 = 173

// Default filenames for configuration files.

var default_main_config string
var default_extra_config string

type default_type_t = int32

const DEFAULT_INT_HEX = 1
const DEFAULT_STRING = 2
const DEFAULT_FLOAT = 3
const DEFAULT_KEY = 4

type default_t struct {
	Fname                string
	Flocation            any
	Ftype1               default_type_t
	Funtranslated        int32
	Foriginal_translated int32
	Fbound               boolean
}

type default_collection_t struct {
	Fdefaults    []default_t
	Fnumdefaults int32
	Ffilename    string
}

//! @begin_config_file default

var doom_defaults_list = [76]default_t{
	0: {
		Fname: "mouse_sensitivity",
	},
	1: {
		Fname: "sfx_volume",
	},
	2: {
		Fname: "music_volume",
	},
	3: {
		Fname: "show_talk",
	},
	4: {
		Fname: "voice_volume",
	},
	5: {
		Fname: "show_messages",
	},
	6: {
		Fname:  "key_right",
		Ftype1: DEFAULT_KEY,
	},
	7: {
		Fname:  "key_left",
		Ftype1: DEFAULT_KEY,
	},
	8: {
		Fname:  "key_up",
		Ftype1: DEFAULT_KEY,
	},
	9: {
		Fname:  "key_down",
		Ftype1: DEFAULT_KEY,
	},
	10: {
		Fname:  "key_strafeleft",
		Ftype1: DEFAULT_KEY,
	},
	11: {
		Fname:  "key_straferight",
		Ftype1: DEFAULT_KEY,
	},
	12: {
		Fname:  "key_useHealth",
		Ftype1: DEFAULT_KEY,
	},
	13: {
		Fname:  "key_jump",
		Ftype1: DEFAULT_KEY,
	},
	14: {
		Fname:  "key_flyup",
		Ftype1: DEFAULT_KEY,
	},
	15: {
		Fname:  "key_flydown",
		Ftype1: DEFAULT_KEY,
	},
	16: {
		Fname:  "key_flycenter",
		Ftype1: DEFAULT_KEY,
	},
	17: {
		Fname:  "key_lookup",
		Ftype1: DEFAULT_KEY,
	},
	18: {
		Fname:  "key_lookdown",
		Ftype1: DEFAULT_KEY,
	},
	19: {
		Fname:  "key_lookcenter",
		Ftype1: DEFAULT_KEY,
	},
	20: {
		Fname:  "key_invquery",
		Ftype1: DEFAULT_KEY,
	},
	21: {
		Fname:  "key_mission",
		Ftype1: DEFAULT_KEY,
	},
	22: {
		Fname:  "key_invPop",
		Ftype1: DEFAULT_KEY,
	},
	23: {
		Fname:  "key_invKey",
		Ftype1: DEFAULT_KEY,
	},
	24: {
		Fname:  "key_invHome",
		Ftype1: DEFAULT_KEY,
	},
	25: {
		Fname:  "key_invEnd",
		Ftype1: DEFAULT_KEY,
	},
	26: {
		Fname:  "key_invleft",
		Ftype1: DEFAULT_KEY,
	},
	27: {
		Fname:  "key_invright",
		Ftype1: DEFAULT_KEY,
	},
	28: {
		Fname:  "key_invLeft",
		Ftype1: DEFAULT_KEY,
	},
	29: {
		Fname:  "key_invRight",
		Ftype1: DEFAULT_KEY,
	},
	30: {
		Fname:  "key_useartifact",
		Ftype1: DEFAULT_KEY,
	},
	31: {
		Fname:  "key_invUse",
		Ftype1: DEFAULT_KEY,
	},
	32: {
		Fname:  "key_invDrop",
		Ftype1: DEFAULT_KEY,
	},
	33: {
		Fname:  "key_lookUp",
		Ftype1: DEFAULT_KEY,
	},
	34: {
		Fname:  "key_lookDown",
		Ftype1: DEFAULT_KEY,
	},
	35: {
		Fname:  "key_fire",
		Ftype1: DEFAULT_KEY,
	},
	36: {
		Fname:  "key_use",
		Ftype1: DEFAULT_KEY,
	},
	37: {
		Fname:  "key_strafe",
		Ftype1: DEFAULT_KEY,
	},
	38: {
		Fname:  "key_speed",
		Ftype1: DEFAULT_KEY,
	},
	39: {
		Fname: "use_mouse",
	},
	40: {
		Fname: "mouseb_fire",
	},
	41: {
		Fname: "mouseb_strafe",
	},
	42: {
		Fname: "mouseb_forward",
	},
	43: {
		Fname: "mouseb_jump",
	},
	44: {
		Fname: "use_joystick",
	},
	45: {
		Fname: "joyb_fire",
	},
	46: {
		Fname: "joyb_strafe",
	},
	47: {
		Fname: "joyb_use",
	},
	48: {
		Fname: "joyb_speed",
	},
	49: {
		Fname: "joyb_jump",
	},
	50: {
		Fname: "screenblocks",
	},
	51: {
		Fname: "screensize",
	},
	52: {
		Fname: "detaillevel",
	},
	53: {
		Fname: "snd_channels",
	},
	54: {
		Fname: "snd_musicdevice",
	},
	55: {
		Fname: "snd_sfxdevice",
	},
	56: {
		Fname: "snd_sbport",
	},
	57: {
		Fname: "snd_sbirq",
	},
	58: {
		Fname: "snd_sbdma",
	},
	59: {
		Fname: "snd_mport",
	},
	60: {
		Fname: "usegamma",
	},
	61: {
		Fname:  "savedir",
		Ftype1: DEFAULT_STRING,
	},
	62: {
		Fname: "messageson",
	},
	63: {
		Fname:  "back_flat",
		Ftype1: DEFAULT_STRING,
	},
	64: {
		Fname:  "nickname",
		Ftype1: DEFAULT_STRING,
	},
	65: {
		Fname:  "chatmacro0",
		Ftype1: DEFAULT_STRING,
	},
	66: {
		Fname:  "chatmacro1",
		Ftype1: DEFAULT_STRING,
	},
	67: {
		Fname:  "chatmacro2",
		Ftype1: DEFAULT_STRING,
	},
	68: {
		Fname:  "chatmacro3",
		Ftype1: DEFAULT_STRING,
	},
	69: {
		Fname:  "chatmacro4",
		Ftype1: DEFAULT_STRING,
	},
	70: {
		Fname:  "chatmacro5",
		Ftype1: DEFAULT_STRING,
	},
	71: {
		Fname:  "chatmacro6",
		Ftype1: DEFAULT_STRING,
	},
	72: {
		Fname:  "chatmacro7",
		Ftype1: DEFAULT_STRING,
	},
	73: {
		Fname:  "chatmacro8",
		Ftype1: DEFAULT_STRING,
	},
	74: {
		Fname:  "chatmacro9",
		Ftype1: DEFAULT_STRING,
	},
	75: {
		Fname: "comport",
	},
}

var doom_defaults = default_collection_t{
	Fdefaults:    doom_defaults_list[:],
	Fnumdefaults: int32(len(doom_defaults_list)),
}

//! @begin_config_file extended

var extra_defaults_list = [119]default_t{
	0: {
		Fname: "graphical_startup",
	},
	1: {
		Fname: "autoadjust_video_settings",
	},
	2: {
		Fname: "fullscreen",
	},
	3: {
		Fname: "aspect_ratio_correct",
	},
	4: {
		Fname: "startup_delay",
	},
	5: {
		Fname: "screen_width",
	},
	6: {
		Fname: "screen_height",
	},
	7: {
		Fname: "screen_bpp",
	},
	8: {
		Fname: "grabmouse",
	},
	9: {
		Fname: "novert",
	},
	10: {
		Fname:  "mouse_acceleration",
		Ftype1: DEFAULT_FLOAT,
	},
	11: {
		Fname: "mouse_threshold",
	},
	12: {
		Fname: "snd_samplerate",
	},
	13: {
		Fname: "snd_cachesize",
	},
	14: {
		Fname: "snd_maxslicetime_ms",
	},
	15: {
		Fname:  "snd_musiccmd",
		Ftype1: DEFAULT_STRING,
	},
	16: {
		Fname:  "opl_io_port",
		Ftype1: DEFAULT_INT_HEX,
	},
	17: {
		Fname: "show_endoom",
	},
	18: {
		Fname: "png_screenshots",
	},
	19: {
		Fname: "vanilla_savegame_limit",
	},
	20: {
		Fname: "vanilla_demo_limit",
	},
	21: {
		Fname: "vanilla_keyboard_mapping",
	},
	22: {
		Fname:  "video_driver",
		Ftype1: DEFAULT_STRING,
	},
	23: {
		Fname:  "window_position",
		Ftype1: DEFAULT_STRING,
	},
	24: {
		Fname: "joystick_index",
	},
	25: {
		Fname: "joystick_x_axis",
	},
	26: {
		Fname: "joystick_x_invert",
	},
	27: {
		Fname: "joystick_y_axis",
	},
	28: {
		Fname: "joystick_y_invert",
	},
	29: {
		Fname: "joystick_strafe_axis",
	},
	30: {
		Fname: "joystick_strafe_invert",
	},
	31: {
		Fname: "joystick_physical_button0",
	},
	32: {
		Fname: "joystick_physical_button1",
	},
	33: {
		Fname: "joystick_physical_button2",
	},
	34: {
		Fname: "joystick_physical_button3",
	},
	35: {
		Fname: "joystick_physical_button4",
	},
	36: {
		Fname: "joystick_physical_button5",
	},
	37: {
		Fname: "joystick_physical_button6",
	},
	38: {
		Fname: "joystick_physical_button7",
	},
	39: {
		Fname: "joystick_physical_button8",
	},
	40: {
		Fname: "joystick_physical_button9",
	},
	41: {
		Fname: "joyb_strafeleft",
	},
	42: {
		Fname: "joyb_straferight",
	},
	43: {
		Fname: "joyb_menu_activate",
	},
	44: {
		Fname: "joyb_prevweapon",
	},
	45: {
		Fname: "joyb_nextweapon",
	},
	46: {
		Fname: "mouseb_strafeleft",
	},
	47: {
		Fname: "mouseb_straferight",
	},
	48: {
		Fname: "mouseb_use",
	},
	49: {
		Fname: "mouseb_backward",
	},
	50: {
		Fname: "mouseb_prevweapon",
	},
	51: {
		Fname: "mouseb_nextweapon",
	},
	52: {
		Fname: "dclick_use",
	},
	53: {
		Fname:  "key_pause",
		Ftype1: DEFAULT_KEY,
	},
	54: {
		Fname:  "key_menu_activate",
		Ftype1: DEFAULT_KEY,
	},
	55: {
		Fname:  "key_menu_up",
		Ftype1: DEFAULT_KEY,
	},
	56: {
		Fname:  "key_menu_down",
		Ftype1: DEFAULT_KEY,
	},
	57: {
		Fname:  "key_menu_left",
		Ftype1: DEFAULT_KEY,
	},
	58: {
		Fname:  "key_menu_right",
		Ftype1: DEFAULT_KEY,
	},
	59: {
		Fname:  "key_menu_back",
		Ftype1: DEFAULT_KEY,
	},
	60: {
		Fname:  "key_menu_forward",
		Ftype1: DEFAULT_KEY,
	},
	61: {
		Fname:  "key_menu_confirm",
		Ftype1: DEFAULT_KEY,
	},
	62: {
		Fname:  "key_menu_abort",
		Ftype1: DEFAULT_KEY,
	},
	63: {
		Fname:  "key_menu_help",
		Ftype1: DEFAULT_KEY,
	},
	64: {
		Fname:  "key_menu_save",
		Ftype1: DEFAULT_KEY,
	},
	65: {
		Fname:  "key_menu_load",
		Ftype1: DEFAULT_KEY,
	},
	66: {
		Fname:  "key_menu_volume",
		Ftype1: DEFAULT_KEY,
	},
	67: {
		Fname:  "key_menu_detail",
		Ftype1: DEFAULT_KEY,
	},
	68: {
		Fname:  "key_menu_qsave",
		Ftype1: DEFAULT_KEY,
	},
	69: {
		Fname:  "key_menu_endgame",
		Ftype1: DEFAULT_KEY,
	},
	70: {
		Fname:  "key_menu_messages",
		Ftype1: DEFAULT_KEY,
	},
	71: {
		Fname:  "key_menu_qload",
		Ftype1: DEFAULT_KEY,
	},
	72: {
		Fname:  "key_menu_quit",
		Ftype1: DEFAULT_KEY,
	},
	73: {
		Fname:  "key_menu_gamma",
		Ftype1: DEFAULT_KEY,
	},
	74: {
		Fname:  "key_spy",
		Ftype1: DEFAULT_KEY,
	},
	75: {
		Fname:  "key_menu_incscreen",
		Ftype1: DEFAULT_KEY,
	},
	76: {
		Fname:  "key_menu_decscreen",
		Ftype1: DEFAULT_KEY,
	},
	77: {
		Fname:  "key_menu_screenshot",
		Ftype1: DEFAULT_KEY,
	},
	78: {
		Fname:  "key_map_toggle",
		Ftype1: DEFAULT_KEY,
	},
	79: {
		Fname:  "key_map_north",
		Ftype1: DEFAULT_KEY,
	},
	80: {
		Fname:  "key_map_south",
		Ftype1: DEFAULT_KEY,
	},
	81: {
		Fname:  "key_map_east",
		Ftype1: DEFAULT_KEY,
	},
	82: {
		Fname:  "key_map_west",
		Ftype1: DEFAULT_KEY,
	},
	83: {
		Fname:  "key_map_zoomin",
		Ftype1: DEFAULT_KEY,
	},
	84: {
		Fname:  "key_map_zoomout",
		Ftype1: DEFAULT_KEY,
	},
	85: {
		Fname:  "key_map_maxzoom",
		Ftype1: DEFAULT_KEY,
	},
	86: {
		Fname:  "key_map_follow",
		Ftype1: DEFAULT_KEY,
	},
	87: {
		Fname:  "key_map_grid",
		Ftype1: DEFAULT_KEY,
	},
	88: {
		Fname:  "key_map_mark",
		Ftype1: DEFAULT_KEY,
	},
	89: {
		Fname:  "key_map_clearmark",
		Ftype1: DEFAULT_KEY,
	},
	90: {
		Fname:  "key_weapon1",
		Ftype1: DEFAULT_KEY,
	},
	91: {
		Fname:  "key_weapon2",
		Ftype1: DEFAULT_KEY,
	},
	92: {
		Fname:  "key_weapon3",
		Ftype1: DEFAULT_KEY,
	},
	93: {
		Fname:  "key_weapon4",
		Ftype1: DEFAULT_KEY,
	},
	94: {
		Fname:  "key_weapon5",
		Ftype1: DEFAULT_KEY,
	},
	95: {
		Fname:  "key_weapon6",
		Ftype1: DEFAULT_KEY,
	},
	96: {
		Fname:  "key_weapon7",
		Ftype1: DEFAULT_KEY,
	},
	97: {
		Fname:  "key_weapon8",
		Ftype1: DEFAULT_KEY,
	},
	98: {
		Fname:  "key_prevweapon",
		Ftype1: DEFAULT_KEY,
	},
	99: {
		Fname:  "key_nextweapon",
		Ftype1: DEFAULT_KEY,
	},
	100: {
		Fname:  "key_arti_all",
		Ftype1: DEFAULT_KEY,
	},
	101: {
		Fname:  "key_arti_health",
		Ftype1: DEFAULT_KEY,
	},
	102: {
		Fname:  "key_arti_poisonbag",
		Ftype1: DEFAULT_KEY,
	},
	103: {
		Fname:  "key_arti_blastradius",
		Ftype1: DEFAULT_KEY,
	},
	104: {
		Fname:  "key_arti_teleport",
		Ftype1: DEFAULT_KEY,
	},
	105: {
		Fname:  "key_arti_teleportother",
		Ftype1: DEFAULT_KEY,
	},
	106: {
		Fname:  "key_arti_egg",
		Ftype1: DEFAULT_KEY,
	},
	107: {
		Fname:  "key_arti_invulnerability",
		Ftype1: DEFAULT_KEY,
	},
	108: {
		Fname:  "key_message_refresh",
		Ftype1: DEFAULT_KEY,
	},
	109: {
		Fname:  "key_demo_quit",
		Ftype1: DEFAULT_KEY,
	},
	110: {
		Fname:  "key_multi_msg",
		Ftype1: DEFAULT_KEY,
	},
	111: {
		Fname:  "key_multi_msgplayer1",
		Ftype1: DEFAULT_KEY,
	},
	112: {
		Fname:  "key_multi_msgplayer2",
		Ftype1: DEFAULT_KEY,
	},
	113: {
		Fname:  "key_multi_msgplayer3",
		Ftype1: DEFAULT_KEY,
	},
	114: {
		Fname:  "key_multi_msgplayer4",
		Ftype1: DEFAULT_KEY,
	},
	115: {
		Fname:  "key_multi_msgplayer5",
		Ftype1: DEFAULT_KEY,
	},
	116: {
		Fname:  "key_multi_msgplayer6",
		Ftype1: DEFAULT_KEY,
	},
	117: {
		Fname:  "key_multi_msgplayer7",
		Ftype1: DEFAULT_KEY,
	},
	118: {
		Fname:  "key_multi_msgplayer8",
		Ftype1: DEFAULT_KEY,
	},
}

var extra_defaults = default_collection_t{
	Fdefaults:    extra_defaults_list[:],
	Fnumdefaults: int32(len(extra_defaults_list)),
}

// Search a collection for a variable

func searchCollection(collection *default_collection_t, name string) *default_t {
	for i := range collection.Fnumdefaults {
		if strings.EqualFold(name, collection.Fdefaults[i].Fname) {
			return &collection.Fdefaults[i]
		}
	}
	return nil
}

func saveDefaultCollection(collection *default_collection_t) {
}

func loadDefaultCollection(collection *default_collection_t) {
}

// Set the default filenames to use for configuration files.

func m_SetConfigFilenames(main_config string, extra_config string) {
	default_main_config = main_config
	default_extra_config = extra_config
}

//
// M_SaveDefaults
//

func m_SaveDefaults() {
	saveDefaultCollection(&doom_defaults)
	saveDefaultCollection(&extra_defaults)
}

//
// M_LoadDefaults
//

func m_LoadDefaults() {
	var i int32
	// check for a custom default file
	//!
	// @arg <file>
	// @vanilla
	//
	// Load main configuration from the specified file, instead of the
	// default.
	//
	i = m_CheckParmWithArgs("-config", 1)
	if i != 0 {
		doom_defaults.Ffilename = myargs[i+1]
		fprintf_ccgo(os.Stdout, "\tdefault file: %s\n", doom_defaults.Ffilename)
	} else {
		doom_defaults.Ffilename = configdir + default_main_config
	}
	fprintf_ccgo(os.Stdout, "saving config in %s\n", doom_defaults.Ffilename)
	//!
	// @arg <file>
	//
	// Load additional configuration from the specified file, instead of
	// the default.
	//
	i = m_CheckParmWithArgs("-extraconfig", 1)
	if i != 0 {
		extra_defaults.Ffilename = myargs[i+1]
		fprintf_ccgo(os.Stdout, "        extra configuration file: %s\n", extra_defaults.Ffilename)
	} else {
		extra_defaults.Ffilename = configdir + default_extra_config
	}
	loadDefaultCollection(&doom_defaults)
	loadDefaultCollection(&extra_defaults)
}

// Get a configuration file variable by its name

func getDefaultForName(name string) *default_t {
	var result *default_t
	// Try the main list and the extras
	result = searchCollection(&doom_defaults, name)
	if result == nil {
		result = searchCollection(&extra_defaults, name)
	}
	// Not found? Internal error.
	if result == nil {
		i_Error("Unknown configuration variable: '%s'", name)
	}
	return result
}

//
// Bind a variable to a given configuration file variable, by name.
//

func m_BindVariable(name string, location any) {
	var variable *default_t
	variable = getDefaultForName(name)
	variable.Flocation = location
	variable.Fbound = 1
}

// Get the path to the default configuration dir to use, if NULL
// is passed to m_SetConfigDir.

func getDefaultConfigDir() string {
	return "."
}

//
// SetConfigDir:
//
// Sets the location of the configuration directory, where configuration
// files are stored - default.cfg, chocolate-doom.cfg, savegames, etc.
//

func m_SetConfigDir(dir string) {
	// Use the directory that was passed, or find the default.
	if dir != "" {
		configdir = dir
	} else {
		configdir = getDefaultConfigDir()

		if configdir == "" {
			fprintf_ccgo(os.Stdout, "Using %s for configuration and saves\n", configdir)
		}
	}
	// Make the directory if it doesn't already exist:
	m_MakeDirectory(configdir)
}

//
// Calculate the path to the directory to use to store save games.
// Creates the directory as necessary.
//

func m_GetSaveGameDir(iwadname string) string {
	var savegamedir string
	// If not "doing" a configuration directory (Windows), don't "do"
	// a savegame directory, either.
	if configdir == "" {
		savegamedir = ""
	} else {
		savegamedir = configdir + "/" + ".savegame/"
		m_MakeDirectory(savegamedir)
		fprintf_ccgo(os.Stdout, "Using %s for savegames\n", savegamedir)
	}
	return savegamedir
}

const KEY_EQUALS1 = 61
const KEY_FIRE1 = 163
const KEY_STRAFE_L1 = 160
const KEY_STRAFE_R1 = 161
const KEY_USE1 = 162

func init() {
	key_right = int32(KEY_RIGHTARROW1)
	key_left = int32(KEY_LEFTARROW1)
	key_up = int32(KEY_UPARROW1)
	key_down = int32(KEY_DOWNARROW1)
	key_strafeleft = int32(KEY_STRAFE_L1)
	key_straferight = int32(KEY_STRAFE_R1)
	key_fire = int32(KEY_FIRE1)
	key_use = int32(KEY_USE1)
	key_strafe = 0x80 + 0x38
	key_speed = 0x80 + 0x36
	key_flyup = 0x80 + 0x49
	key_flydown = 0x80 + 0x52
	key_flycenter = 0x80 + 0x47
	key_lookup = 0x80 + 0x51
	key_lookdown = 0x80 + 0x53
	key_lookcenter = 0x80 + 0x4f
	key_invleft = '['
	key_invright = ']'
	key_useartifact = KEY_ENTER
	key_jump = '/'
	key_arti_all = int32(KEY_BACKSPACE3)
	key_arti_health = '\\'
	key_arti_poisonbag = '0'
	key_arti_blastradius = '9'
	key_arti_teleport = '8'
	key_arti_teleportother = '7'
	key_arti_egg = '6'
	key_arti_invulnerability = '5'
	key_usehealth = 'h'
	key_invquery = 'q'
	key_mission = 'w'
	key_invpop = 'z'
	key_invkey = 'k'
	key_invhome = 0x80 + 0x47
	key_invend = 0x80 + 0x4f
	key_invuse = KEY_ENTER
	key_invdrop = int32(KEY_BACKSPACE3)
	mousebstrafe = 1
	mousebforward = 2
	mousebjump = -1
	mousebstrafeleft = -1
	mousebstraferight = -1
	mousebbackward = -1
	mousebuse = -1
	mousebprevweapon = -1
	mousebnextweapon = -1
	key_message_refresh = KEY_ENTER
	key_pause = int32(KEY_PAUSE1)
	key_demo_quit = 'q'
	key_spy = 0x80 + 0x58
	key_multi_msg = 't'
	key_weapon1 = '1'
	key_weapon2 = '2'
	key_weapon3 = '3'
	key_weapon4 = '4'
	key_weapon5 = '5'
	key_weapon6 = '6'
	key_weapon7 = '7'
	key_weapon8 = '8'
	key_map_north = int32(KEY_UPARROW1)
	key_map_south = int32(KEY_DOWNARROW1)
	key_map_east = int32(KEY_RIGHTARROW1)
	key_map_west = int32(KEY_LEFTARROW1)
	key_map_zoomin = '='
	key_map_zoomout = '-'
	key_map_toggle = KEY_TAB
	key_map_maxzoom = '0'
	key_map_follow = 'f'
	key_map_grid = 'g'
	key_map_mark = 'm'
	key_map_clearmark = 'c'
	key_menu_activate = KEY_ESCAPE
	key_menu_up = int32(KEY_UPARROW1)
	key_menu_down = int32(KEY_DOWNARROW1)
	key_menu_left = int32(KEY_LEFTARROW1)
	key_menu_right = int32(KEY_RIGHTARROW1)
	key_menu_back = int32(KEY_BACKSPACE3)
	key_menu_forward = KEY_ENTER
	key_menu_confirm = 'y'
	key_menu_abort = 'n'
	key_menu_help = 0x80 + 0x3b
	key_menu_save = 0x80 + 0x3c
	key_menu_load = 0x80 + 0x3d
	key_menu_volume = 0x80 + 0x3e
	key_menu_detail = 0x80 + 0x3f
	key_menu_qsave = 0x80 + 0x40
	key_menu_endgame = 0x80 + 0x41
	key_menu_messages = 0x80 + 0x42
	key_menu_qload = 0x80 + 0x43
	key_menu_quit = 0x80 + 0x44
	key_menu_gamma = 0x80 + 0x57
	key_menu_incscreen = int32(KEY_EQUALS1)
	key_menu_decscreen = int32(KEY_MINUS1)
	joybstrafe = 1
	joybuse = 3
	joybspeed = 2
	joybstrafeleft = -1
	joybstraferight = -1
	joybjump = -1
	joybprevweapon = -1
	joybnextweapon = -1
	joybmenu = -1
	dclick_use = 1
}

//
// Bind all of the common controls used by Doom and all other games.
//

func m_BindBaseControls() {
	m_BindVariable("key_right", &key_right)
	m_BindVariable("key_left", &key_left)
	m_BindVariable("key_up", &key_up)
	m_BindVariable("key_down", &key_down)
	m_BindVariable("key_strafeleft", &key_strafeleft)
	m_BindVariable("key_straferight", &key_straferight)
	m_BindVariable("key_fire", &key_fire)
	m_BindVariable("key_use", &key_use)
	m_BindVariable("key_strafe", &key_strafe)
	m_BindVariable("key_speed", &key_speed)
	m_BindVariable("mouseb_fire", &mousebfire)
	m_BindVariable("mouseb_strafe", &mousebstrafe)
	m_BindVariable("mouseb_forward", &mousebforward)
	m_BindVariable("joyb_fire", &joybfire)
	m_BindVariable("joyb_strafe", &joybstrafe)
	m_BindVariable("joyb_use", &joybuse)
	m_BindVariable("joyb_speed", &joybspeed)
	m_BindVariable("joyb_menu_activate", &joybmenu)
	// Extra controls that are not in the Vanilla versions:
	m_BindVariable("joyb_strafeleft", &joybstrafeleft)
	m_BindVariable("joyb_straferight", &joybstraferight)
	m_BindVariable("mouseb_strafeleft", &mousebstrafeleft)
	m_BindVariable("mouseb_straferight", &mousebstraferight)
	m_BindVariable("mouseb_use", &mousebuse)
	m_BindVariable("mouseb_backward", &mousebbackward)
	m_BindVariable("dclick_use", &dclick_use)
	m_BindVariable("key_pause", &key_pause)
	m_BindVariable("key_message_refresh", &key_message_refresh)
}

func m_BindWeaponControls() {
	m_BindVariable("key_weapon1", &key_weapon1)
	m_BindVariable("key_weapon2", &key_weapon2)
	m_BindVariable("key_weapon3", &key_weapon3)
	m_BindVariable("key_weapon4", &key_weapon4)
	m_BindVariable("key_weapon5", &key_weapon5)
	m_BindVariable("key_weapon6", &key_weapon6)
	m_BindVariable("key_weapon7", &key_weapon7)
	m_BindVariable("key_weapon8", &key_weapon8)
	m_BindVariable("key_prevweapon", &key_prevweapon)
	m_BindVariable("key_nextweapon", &key_nextweapon)
	m_BindVariable("joyb_prevweapon", &joybprevweapon)
	m_BindVariable("joyb_nextweapon", &joybnextweapon)
	m_BindVariable("mouseb_prevweapon", &mousebprevweapon)
	m_BindVariable("mouseb_nextweapon", &mousebnextweapon)
}

func m_BindMapControls() {
	m_BindVariable("key_map_north", &key_map_north)
	m_BindVariable("key_map_south", &key_map_south)
	m_BindVariable("key_map_east", &key_map_east)
	m_BindVariable("key_map_west", &key_map_west)
	m_BindVariable("key_map_zoomin", &key_map_zoomin)
	m_BindVariable("key_map_zoomout", &key_map_zoomout)
	m_BindVariable("key_map_toggle", &key_map_toggle)
	m_BindVariable("key_map_maxzoom", &key_map_maxzoom)
	m_BindVariable("key_map_follow", &key_map_follow)
	m_BindVariable("key_map_grid", &key_map_grid)
	m_BindVariable("key_map_mark", &key_map_mark)
	m_BindVariable("key_map_clearmark", &key_map_clearmark)
}

func m_BindMenuControls() {
	m_BindVariable("key_menu_activate", &key_menu_activate)
	m_BindVariable("key_menu_up", &key_menu_up)
	m_BindVariable("key_menu_down", &key_menu_down)
	m_BindVariable("key_menu_left", &key_menu_left)
	m_BindVariable("key_menu_right", &key_menu_right)
	m_BindVariable("key_menu_back", &key_menu_back)
	m_BindVariable("key_menu_forward", &key_menu_forward)
	m_BindVariable("key_menu_confirm", &key_menu_confirm)
	m_BindVariable("key_menu_abort", &key_menu_abort)
	m_BindVariable("key_menu_help", &key_menu_help)
	m_BindVariable("key_menu_save", &key_menu_save)
	m_BindVariable("key_menu_load", &key_menu_load)
	m_BindVariable("key_menu_volume", &key_menu_volume)
	m_BindVariable("key_menu_detail", &key_menu_detail)
	m_BindVariable("key_menu_qsave", &key_menu_qsave)
	m_BindVariable("key_menu_endgame", &key_menu_endgame)
	m_BindVariable("key_menu_messages", &key_menu_messages)
	m_BindVariable("key_menu_qload", &key_menu_qload)
	m_BindVariable("key_menu_quit", &key_menu_quit)
	m_BindVariable("key_menu_gamma", &key_menu_gamma)
	m_BindVariable("key_menu_incscreen", &key_menu_incscreen)
	m_BindVariable("key_menu_decscreen", &key_menu_decscreen)
	m_BindVariable("key_menu_screenshot", &key_menu_screenshot)
	m_BindVariable("key_demo_quit", &key_demo_quit)
	m_BindVariable("key_spy", &key_spy)
}

func m_BindChatControls(num_players uint32) {
	m_BindVariable("key_multi_msg", &key_multi_msg)
	for i := range num_players {
		name := fmt.Sprintf("key_multi_msgplayer%d", i+1)
		m_BindVariable(name, &key_multi_msgplayer[i])
	}
}

//
// Apply custom patches to the default values depending on the
// platform we are running on.
//

func m_ApplyPlatformDefaults() {
	// no-op. Add your platform-specific patches here.
}

const INT_MAX7 = 2147483647

// Fixme. __USE_C_FIXED__ or something.

func fixedMul(a fixed_t, b fixed_t) fixed_t {
	return int32(int64(a) * int64(b) >> FRACBITS)
}

//
// fixedDiv, C version.
//

func fixedDiv(a fixed_t, b fixed_t) fixed_t {
	var result int64
	var v1 int32
	if xabs(a)>>int32(14) >= xabs(b) {
		if a^b < 0 {
			v1 = -1 - 0x7fffffff
		} else {
			v1 = int32(INT_MAX7)
		}
		return v1
	} else {
		result = int64(a) << 16 / int64(b)
		return int32(result)
	}
}

const LINEHEIGHT = 16

func init() {
	mouseSensitivity = 5
	showMessages = 1
	screenblocks = 10
	gammamsg = [5]string{
		"Gamma correction OFF",
		"Gamma correction level 1",
		"Gamma correction level 2",
		"Gamma correction level 3",
		"Gamma correction level 4",
	}
}

//static boolean opldev;

// C documentation
//
//	//
//	// MENU TYPEDEFS
//	//
type menuitem_t struct {
	Fstatus   int16
	Fname     string
	Froutine  func(choice int32)
	FalphaKey int8
}

type menu_t struct {
	Fnumitems  int16
	FprevMenu  *menu_t
	Fmenuitems []menuitem_t
	Froutine   func()
	Fx         int16
	Fy         int16
	FlastOn    int16
}

func init() {
	skullName = [2]string{
		0: "M_SKULL1",
		1: "M_SKULL2",
	}
}

const readthis = 4
const quitdoom = 5
const main_end = 6

func init() {
	MainMenu = [6]menuitem_t{
		0: {
			Fstatus:   1,
			Fname:     "M_NGAME",
			FalphaKey: 'n',
			Froutine:  m_NewGame,
		},
		1: {
			Fstatus:   1,
			Fname:     "M_OPTION",
			FalphaKey: 'o',
			Froutine:  m_Options,
		},
		2: {
			Fstatus:   1,
			Fname:     "M_LOADG",
			FalphaKey: 'l',
			Froutine:  m_LoadGame,
		},
		3: {
			Fstatus:   1,
			Fname:     "M_SAVEG",
			FalphaKey: 's',
			Froutine:  m_SaveGame,
		},
		4: {
			Fstatus:   1,
			Fname:     "M_RDTHIS",
			FalphaKey: 'r',
			Froutine:  m_ReadThis,
		},
		5: {
			Fstatus:   1,
			Fname:     "M_QUITG",
			FalphaKey: 'q',
			Froutine:  m_QuitDOOM,
		},
	}
}

func init() {
	MainDef = menu_t{
		Fnumitems:  int16(main_end),
		Fmenuitems: MainMenu[:],
		Froutine:   m_DrawMainMenu,
		Fx:         97,
		Fy:         64,
	}
}

const ep_end = 4

func init() {
	EpisodeMenu = [4]menuitem_t{
		0: {
			Fstatus:   1,
			Fname:     "M_EPI1",
			FalphaKey: 'k',
			Froutine:  m_Episode,
		},
		1: {
			Fstatus:   1,
			Fname:     "M_EPI2",
			FalphaKey: 't',
			Froutine:  m_Episode,
		},
		2: {
			Fstatus:   1,
			Fname:     "M_EPI3",
			FalphaKey: 'i',
			Froutine:  m_Episode,
		},
		3: {
			Fstatus:   1,
			Fname:     "M_EPI4",
			FalphaKey: 't',
			Froutine:  m_Episode,
		},
	}
}

func init() {
	EpiDef = menu_t{
		Fnumitems:  int16(ep_end),
		FprevMenu:  &MainDef,
		Fmenuitems: EpisodeMenu[:],
		Froutine:   m_DrawEpisode,
		Fx:         48,
		Fy:         63,
	}
}

const hurtme = 2

// const nightmare = 4
const newg_end = 5

func init() {
	NewGameMenu = [5]menuitem_t{
		0: {
			Fstatus:   1,
			Fname:     "M_JKILL",
			FalphaKey: 'i',
			Froutine:  m_ChooseSkill,
		},
		1: {
			Fstatus:   1,
			Fname:     "M_ROUGH",
			FalphaKey: 'h',
			Froutine:  m_ChooseSkill,
		},
		2: {
			Fstatus:   1,
			Fname:     "M_HURT",
			FalphaKey: 'h',
			Froutine:  m_ChooseSkill,
		},
		3: {
			Fstatus:   1,
			Fname:     "M_ULTRA",
			FalphaKey: 'u',
			Froutine:  m_ChooseSkill,
		},
		4: {
			Fstatus:   1,
			Fname:     "M_NMARE",
			FalphaKey: 'n',
			Froutine:  m_ChooseSkill,
		},
	}
}

func init() {
	NewDef = menu_t{
		Fnumitems:  int16(newg_end),
		FprevMenu:  &EpiDef,
		Fmenuitems: NewGameMenu[:],
		Froutine:   m_DrawNewGame,
		Fx:         48,
		Fy:         63,
		FlastOn:    int16(hurtme),
	}
}

const messages = 1
const detail = 2
const scrnsize = 3
const mousesens = 5
const opt_end = 8

func init() {
	OptionsMenu = [8]menuitem_t{
		0: {
			Fstatus:   1,
			Fname:     "M_ENDGAM",
			FalphaKey: 'e',
			Froutine:  m_EndGame,
		},
		1: {
			Fstatus:   1,
			Fname:     "M_MESSG",
			FalphaKey: 'm',
			Froutine:  m_ChangeMessages,
		},
		2: {
			Fstatus:   1,
			Fname:     "M_DETAIL",
			FalphaKey: 'g',
			Froutine:  m_ChangeDetail,
		},
		3: {
			Fstatus:   2,
			Fname:     "M_SCRNSZ",
			FalphaKey: 's',
			Froutine:  m_SizeDisplay,
		},
		4: {
			Fstatus: int16(-1),
			Fname:   "",
		},
		5: {
			Fstatus:   2,
			Fname:     "M_MSENS",
			FalphaKey: 'm',
			Froutine:  m_ChangeSensitivity,
		},
		6: {
			Fstatus: int16(-1),
			Fname:   "",
		},
		7: {
			Fstatus:   1,
			Fname:     "M_SVOL",
			FalphaKey: 's',
			Froutine:  m_Sound,
		},
	}
}

func init() {
	OptionsDef = menu_t{
		Fnumitems:  int16(opt_end),
		FprevMenu:  &MainDef,
		Fmenuitems: OptionsMenu[:],
		Froutine:   m_DrawOptions,
		Fx:         60,
		Fy:         37,
	}
}

const read1_end = 1

func init() {
	ReadMenu1 = [1]menuitem_t{
		0: {
			Fstatus:  1,
			Froutine: m_ReadThis2,
		},
	}
}

func init() {
	ReadDef1 = menu_t{
		Fnumitems:  int16(read1_end),
		FprevMenu:  &MainDef,
		Fmenuitems: ReadMenu1[:],
		Froutine:   m_DrawReadThis1,
		Fx:         280,
		Fy:         185,
	}
}

const read2_end = 1

func init() {
	ReadMenu2 = [1]menuitem_t{
		0: {
			Fstatus:  1,
			Froutine: m_FinishReadThis,
		},
	}
}

func init() {
	ReadDef2 = menu_t{
		Fnumitems:  int16(read2_end),
		FprevMenu:  &ReadDef1,
		Fmenuitems: ReadMenu2[:],
		Froutine:   m_DrawReadThis2,
		Fx:         330,
		Fy:         175,
	}
}

const sfx_vol = 0
const music_vol = 2
const sound_end = 4

func init() {
	SoundMenu = [4]menuitem_t{
		0: {
			Fstatus:   2,
			Fname:     "M_SFXVOL",
			FalphaKey: 's',
			Froutine:  m_SfxVol,
		},
		1: {
			Fstatus: int16(-1),
		},
		2: {
			Fstatus:   2,
			Fname:     "M_MUSVOL",
			FalphaKey: 'm',
			Froutine:  m_MusicVol,
		},
		3: {
			Fstatus: int16(-1),
		},
	}
}

func init() {
	SoundDef = menu_t{
		Fnumitems:  int16(sound_end),
		FprevMenu:  &OptionsDef,
		Fmenuitems: SoundMenu[:],
		Froutine:   m_DrawSound,
		Fx:         80,
		Fy:         64,
	}
}

const load_end = 6

func init() {
	LoadMenu = [6]menuitem_t{
		0: {
			Fstatus:   1,
			FalphaKey: '1',
			Froutine:  m_LoadSelect,
		},
		1: {
			Fstatus:   1,
			FalphaKey: '2',
			Froutine:  m_LoadSelect,
		},
		2: {
			Fstatus:   1,
			FalphaKey: '3',
			Froutine:  m_LoadSelect,
		},
		3: {
			Fstatus:   1,
			FalphaKey: '4',
			Froutine:  m_LoadSelect,
		},
		4: {
			Fstatus:   1,
			FalphaKey: '5',
			Froutine:  m_LoadSelect,
		},
		5: {
			Fstatus:   1,
			FalphaKey: '6',
			Froutine:  m_LoadSelect,
		},
	}
}

func init() {
	LoadDef = menu_t{
		Fnumitems:  int16(load_end),
		FprevMenu:  &MainDef,
		Fmenuitems: LoadMenu[:],
		Froutine:   m_DrawLoad,
		Fx:         80,
		Fy:         54,
	}
}

func init() {
	SaveMenu = [6]menuitem_t{
		0: {
			Fstatus:   1,
			FalphaKey: '1',
			Froutine:  m_SaveSelect,
		},
		1: {
			Fstatus:   1,
			FalphaKey: '2',
			Froutine:  m_SaveSelect,
		},
		2: {
			Fstatus:   1,
			FalphaKey: '3',
			Froutine:  m_SaveSelect,
		},
		3: {
			Fstatus:   1,
			FalphaKey: '4',
			Froutine:  m_SaveSelect,
		},
		4: {
			Fstatus:   1,
			FalphaKey: '5',
			Froutine:  m_SaveSelect,
		},
		5: {
			Fstatus:   1,
			FalphaKey: '6',
			Froutine:  m_SaveSelect,
		},
	}
}

func init() {
	SaveDef = menu_t{
		Fnumitems:  int16(load_end),
		FprevMenu:  &MainDef,
		Fmenuitems: SaveMenu[:],
		Froutine:   m_DrawSave,
		Fx:         80,
		Fy:         54,
	}
}

// C documentation
//
//	//
//	// M_ReadSaveStrings
//	//  read the strings from the savegame files
//	//
func m_ReadSaveStrings() {
	for i := range int32(load_end) {
		var thisString [SAVESTRINGSIZE]byte
		var err error
		handle, err := os.Open(p_SaveGameFile(i))
		if err != nil {
			savegamestrings[i] = "empty slot"
			LoadMenu[i].Fstatus = 0
			continue
		}

		handle.Read(thisString[:])
		savegamestrings[i] = gostring_bytes(thisString[:])
		handle.Close()
		LoadMenu[i].Fstatus = 1
	}
}

// C documentation
//
//	//
//	// m_LoadGame & Cie.
//	//
func m_DrawLoad() {
	v_DrawPatchDirect(72, 28, w_CacheLumpNameT("M_LOADG"))
	for i := range int32(load_end) {
		m_DrawSaveLoadBorder(int32(LoadDef.Fx), int32(LoadDef.Fy)+LINEHEIGHT*i)
		m_WriteText(int32(LoadDef.Fx), int32(LoadDef.Fy)+LINEHEIGHT*i, savegamestrings[i])
	}
}

// C documentation
//
//	//
//	// Draw border for the savegame description
//	//
func m_DrawSaveLoadBorder(x int32, y int32) {
	var i int32
	v_DrawPatchDirect(x-8, y+7, w_CacheLumpNameT("M_LSLEFT"))
	i = 0
	for {
		if i >= 24 {
			break
		}
		v_DrawPatchDirect(x, y+7, w_CacheLumpNameT("M_LSCNTR"))
		x += 8
		goto _1
	_1:
		;
		i++
	}
	v_DrawPatchDirect(x, y+7, w_CacheLumpNameT("M_LSRGHT"))
}

// C documentation
//
//	//
//	// User wants to load this game
//	//
func m_LoadSelect(choice int32) {
	g_LoadGame(p_SaveGameFile(choice))
	m_ClearMenus()
}

// C documentation
//
//	//
//	// Selected from DOOM menu
//	//
func m_LoadGame(choice int32) {
	if netgame != 0 {
		m_StartMessage("you can't do load while in a net game!\n\npress a key.", nil, 0)
		return
	}
	m_SetupNextMenu(&LoadDef)
	m_ReadSaveStrings()
}

// C documentation
//
//	//
//	//  m_SaveGame & Cie.
//	//
func m_DrawSave() {
	var i int32
	v_DrawPatchDirect(72, 28, w_CacheLumpNameT("M_SAVEG"))
	i = 0
	for {
		if i >= int32(load_end) {
			break
		}
		m_DrawSaveLoadBorder(int32(LoadDef.Fx), int32(LoadDef.Fy)+LINEHEIGHT*i)
		m_WriteText(int32(LoadDef.Fx), int32(LoadDef.Fy)+LINEHEIGHT*i, savegamestrings[i])
		goto _1
	_1:
		;
		i++
	}
	if saveStringEnter != 0 {
		i = m_StringWidth(savegamestrings[saveSlot])
		m_WriteText(int32(LoadDef.Fx)+i, int32(LoadDef.Fy)+LINEHEIGHT*saveSlot, "_")
	}
}

// C documentation
//
//	//
//	// m_Responder calls this when user is finished
//	//
func m_DoSave(slot int32) {
	g_SaveGame(slot, savegamestrings[slot])
	m_ClearMenus()
	// PICK QUICKSAVE SLOT YET?
	if quickSaveSlot == -2 {
		quickSaveSlot = slot
	}
}

// C documentation
//
//	//
//	// User wants to save. Start string input for M_Responder
//	//
func m_SaveSelect(choice int32) {
	// we are going to be intercepting all chars
	saveStringEnter = 1
	saveSlot = choice
	saveOldString = savegamestrings[choice]
	if strings.EqualFold(savegamestrings[choice], "empty slot") {
		savegamestrings[choice] = ""
	}
	saveCharIndex = len(savegamestrings[choice])
}

// C documentation
//
//	//
//	// Selected from DOOM menu
//	//
func m_SaveGame(choice int32) {
	if usergame == 0 {
		m_StartMessage("you can't save if you aren't playing!\n\npress a key.", nil, 0)
		return
	}
	if gamestate != gs_LEVEL {
		return
	}
	m_SetupNextMenu(&SaveDef)
	m_ReadSaveStrings()
}

func m_QuickSaveResponse(key int32) {
	if key == key_menu_confirm {
		m_DoSave(quickSaveSlot)
		s_StartSound(nil, int32(sfx_swtchx))
	}
}

func m_QuickSave() {
	if usergame == 0 {
		s_StartSound(nil, int32(sfx_oof))
		return
	}
	if gamestate != gs_LEVEL {
		return
	}
	if quickSaveSlot < 0 {
		m_StartControlPanel()
		m_ReadSaveStrings()
		m_SetupNextMenu(&SaveDef)
		quickSaveSlot = -2 // means to pick a slot now
		return
	}
	tempstring := fmt.Sprintf("quicksave over your game named\n\n'%s'?\n\npress y or n.", savegamestrings[quickSaveSlot])
	m_StartMessage(tempstring, m_QuickSaveResponse, 1)
}

// C documentation
//
//	//
//	// M_QuickLoad
//	//
func m_QuickLoadResponse(key int32) {
	if key == key_menu_confirm {
		m_LoadSelect(quickSaveSlot)
		s_StartSound(nil, int32(sfx_swtchx))
	}
}

func m_QuickLoad() {
	if netgame != 0 {
		m_StartMessage("you can't quickload during a netgame!\n\npress a key.", nil, 0)
		return
	}
	if quickSaveSlot < 0 {
		m_StartMessage("you haven't picked a quicksave slot yet!\n\npress a key.", nil, 0)
		return
	}
	tempstring := fmt.Sprintf("do you want to quickload the game named\n\n'%s'?\n\npress y or n.", savegamestrings[quickSaveSlot])
	m_StartMessage(tempstring, m_QuickLoadResponse, 1)
}

// C documentation
//
//	//
//	// Read This Menus
//	// Had a "quick hack to fix romero bug"
//	//
func m_DrawReadThis1() {
	var lumpname string
	var skullx, skully int32
	lumpname = "CREDIT"
	skullx = 330
	skully = 175
	inhelpscreens = 1
	// Different versions of Doom 1.9 work differently
	switch gameversion {
	case exe_doom_1_666:
		fallthrough
	case exe_doom_1_7:
		fallthrough
	case exe_doom_1_8:
		fallthrough
	case exe_doom_1_9:
		fallthrough
	case exe_hacx:
		if gamemode == commercial {
			// Doom 2
			lumpname = "HELP"
			skullx = 330
			skully = 165
		} else {
			// Doom 1
			// HELP2 is the first screen shown in Doom 1
			lumpname = "HELP2"
			skullx = 280
			skully = 185
		}
	case exe_ultimate:
		fallthrough
	case exe_chex:
		// Ultimate Doom always displays "HELP1".
		// Chex Quest version also uses "HELP1", even though it is based
		// on Final Doom.
		lumpname = "HELP1"
	case exe_final:
		fallthrough
	case exe_final2:
		// Final Doom always displays "HELP".
		lumpname = "HELP"
	default:
		i_Error("Unhandled game version")
		break
	}
	v_DrawPatchDirect(0, 0, w_CacheLumpNameT(lumpname))
	ReadDef1.Fx = int16(skullx)
	ReadDef1.Fy = int16(skully)
}

// C documentation
//
//	//
//	// Read This Menus - optional second page.
//	//
func m_DrawReadThis2() {
	inhelpscreens = 1
	// We only ever draw the second page if this is
	// gameversion == exe_doom_1_9 and gamemode == registered
	v_DrawPatchDirect(0, 0, w_CacheLumpNameT("HELP1"))
}

// C documentation
//
//	//
//	// Change Sfx & Music volumes
//	//
func m_DrawSound() {
	v_DrawPatchDirect(60, 38, w_CacheLumpNameT("M_SVOL"))
	m_DrawThermo(int32(SoundDef.Fx), int32(SoundDef.Fy)+LINEHEIGHT*(int32(sfx_vol)+1), 16, sfxVolume)
	m_DrawThermo(int32(SoundDef.Fx), int32(SoundDef.Fy)+LINEHEIGHT*(int32(music_vol)+1), 16, musicVolume)
}

func m_Sound(choice int32) {
	m_SetupNextMenu(&SoundDef)
}

func m_SfxVol(choice int32) {
	switch choice {
	case 0:
		if sfxVolume != 0 {
			sfxVolume--
		}
	case 1:
		if sfxVolume < 15 {
			sfxVolume++
		}
		break
	}
	s_SetSfxVolume(sfxVolume * 8)
}

func m_MusicVol(choice int32) {
	switch choice {
	case 0:
		if musicVolume != 0 {
			musicVolume--
		}
	case 1:
		if musicVolume < 15 {
			musicVolume++
		}
		break
	}
	s_SetMusicVolume(musicVolume * 8)
}

// C documentation
//
//	//
//	// M_DrawMainMenu
//	//
func m_DrawMainMenu() {
	v_DrawPatchDirect(94, 2, w_CacheLumpNameT("M_DOOM"))
}

// C documentation
//
//	//
//	// M_NewGame
//	//
func m_DrawNewGame() {
	v_DrawPatchDirect(96, 14, w_CacheLumpNameT("M_NEWG"))
	v_DrawPatchDirect(54, 38, w_CacheLumpNameT("M_SKILL"))
}

func m_NewGame(choice int32) {
	if netgame != 0 && demoplayback == 0 {
		m_StartMessage("you can't start a new game\nwhile in a network game.\n\npress a key.", nil, 0)
		return
	}
	// Chex Quest disabled the episode select screen, as did Doom II.
	if gamemode == commercial || gameversion == exe_chex {
		m_SetupNextMenu(&NewDef)
	} else {
		m_SetupNextMenu(&EpiDef)
	}
}

func m_DrawEpisode() {
	v_DrawPatchDirect(54, 38, w_CacheLumpNameT("M_EPISOD"))
}

func m_VerifyNightmare(key int32) {
	if key != key_menu_confirm {
		return
	}
	g_DeferedInitNew(sk_nightmare, epi+1, 1)
	m_ClearMenus()
}

func m_ChooseSkill(choice int32) {
	if skill_t(choice) == sk_nightmare {
		m_StartMessage("are you sure? this skill level\nisn't even remotely fair.\n\npress y or n.", m_VerifyNightmare, 1)
		return
	}
	g_DeferedInitNew(skill_t(choice), epi+1, 1)
	m_ClearMenus()
}

func m_Episode(choice int32) {
	if gamemode == shareware && choice != 0 {
		m_StartMessage("this is the shareware version of doom.\n\nyou need to order the entire trilogy.\n\npress a key.", nil, 0)
		m_SetupNextMenu(&ReadDef1)
		return
	}
	// Yet another hack...
	if gamemode == registered && choice > 2 {
		fprintf_ccgo(os.Stderr, "m_Episode: 4th episode requires UltimateDOOM\n")
		choice = 0
	}
	epi = choice
	m_SetupNextMenu(&NewDef)
}

// C documentation
//
//	//
//	// M_Options
//	//
var detailNames = [2]string{
	0: "M_GDHIGH",
	1: "M_GDLOW",
}
var msgNames = [2]string{
	0: "M_MSGOFF",
	1: "M_MSGON",
}

func m_DrawOptions() {
	v_DrawPatchDirect(108, 15, w_CacheLumpNameT("M_OPTTTL"))
	v_DrawPatchDirect(int32(OptionsDef.Fx)+int32(175), int32(OptionsDef.Fy)+LINEHEIGHT*int32(detail), w_CacheLumpNameT(detailNames[detailLevel]))
	v_DrawPatchDirect(int32(OptionsDef.Fx)+int32(120), int32(OptionsDef.Fy)+LINEHEIGHT*int32(messages), w_CacheLumpNameT(msgNames[showMessages]))
	m_DrawThermo(int32(OptionsDef.Fx), int32(OptionsDef.Fy)+LINEHEIGHT*(int32(mousesens)+1), 10, mouseSensitivity)
	m_DrawThermo(int32(OptionsDef.Fx), int32(OptionsDef.Fy)+LINEHEIGHT*(int32(scrnsize)+1), 9, screenSize)
}

func m_Options(choice int32) {
	m_SetupNextMenu(&OptionsDef)
}

// C documentation
//
//	//
//	//      Toggle messages on/off
//	//
func m_ChangeMessages(choice int32) {
	showMessages = 1 - showMessages
	if showMessages == 0 {
		players[consoleplayer].Fmessage = "Messages OFF"
	} else {
		players[consoleplayer].Fmessage = "Messages ON"
	}
	message_dontfuckwithme = 1
}

// C documentation
//
//	//
//	// M_EndGame
//	//
func m_EndGameResponse(key int32) {
	if key != key_menu_confirm {
		return
	}
	currentMenu.FlastOn = itemOn
	m_ClearMenus()
	d_StartTitle()
}

func m_EndGame(choice int32) {
	if usergame == 0 {
		s_StartSound(nil, int32(sfx_oof))
		return
	}
	if netgame != 0 {
		m_StartMessage("you can't end a netgame!\n\npress a key.", nil, 0)
		return
	}
	m_StartMessage("are you sure you want to end the game?\n\npress y or n.", m_EndGameResponse, 1)
}

// C documentation
//
//	//
//	// M_ReadThis
//	//
func m_ReadThis(choice int32) {
	m_SetupNextMenu(&ReadDef1)
}

func m_ReadThis2(choice int32) {
	// Doom 1.9 had two menus when playing Doom 1
	// All others had only one
	if gameversion <= exe_doom_1_9 && gamemode != commercial {
		m_SetupNextMenu(&ReadDef2)
	} else {
		// Close the menu
		m_FinishReadThis(0)
	}
}

func m_FinishReadThis(choice int32) {
	m_SetupNextMenu(&MainDef)
}

func init() {
	quitsounds = [8]int32{
		0: int32(sfx_pldeth),
		1: int32(sfx_dmpain),
		2: int32(sfx_popain),
		3: int32(sfx_slop),
		4: int32(sfx_telept),
		5: int32(sfx_posit1),
		6: int32(sfx_posit3),
		7: int32(sfx_sgtatk),
	}
}

func init() {
	quitsounds2 = [8]int32{
		0: int32(sfx_vilact),
		1: int32(sfx_getpow),
		2: int32(sfx_boscub),
		3: int32(sfx_slop),
		4: int32(sfx_skeswg),
		5: int32(sfx_kntdth),
		6: int32(sfx_bspact),
		7: int32(sfx_sgtatk),
	}
}

var M_QuitResponse = func(key int32) {
	if key != key_menu_confirm {
		return
	}
	if netgame == 0 {
		if gamemode == commercial {
			s_StartSound(nil, quitsounds2[gametic>>2&7])
		} else {
			s_StartSound(nil, quitsounds[gametic>>2&7])
		}
	}
	i_Quit()
}

func m_SelectEndMessage() string {
	var endmsg []string
	var v1 gamemission_t
	if gamemission == pack_chex {
		v1 = doom
	} else {
		if gamemission == pack_hacx {
			v1 = doom2
		} else {
			v1 = gamemission
		}
	}
	if v1 == doom {
		// Doom 1
		endmsg = doom1_endmsg[:]
	} else {
		// Doom 2
		endmsg = doom2_endmsg[:]
	}
	return endmsg[gametic%NUM_QUITMESSAGES]
}

func m_QuitDOOM(choice int32) {
	endstring = fmt.Sprintf("%s\n\n(press y to quit to dos.)", m_SelectEndMessage())
	m_StartMessage(endstring, M_QuitResponse, 1)
}

func m_ChangeSensitivity(choice int32) {
	switch choice {
	case 0:
		if mouseSensitivity != 0 {
			mouseSensitivity--
		}
	case 1:
		if mouseSensitivity < 9 {
			mouseSensitivity++
		}
		break
	}
}

func m_ChangeDetail(choice int32) {
	choice = 0
	detailLevel = 1 - detailLevel
	r_SetViewSize(screenblocks, detailLevel)
	if detailLevel == 0 {
		players[consoleplayer].Fmessage = "High detail"
	} else {
		players[consoleplayer].Fmessage = "Low detail"
	}
}

func m_SizeDisplay(choice int32) {
	switch choice {
	case 0:
		if screenSize > 0 {
			screenblocks--
			screenSize--
		}
	case 1:
		if screenSize < 8 {
			screenblocks++
			screenSize++
		}
		break
	}
	r_SetViewSize(screenblocks, detailLevel)
}

// C documentation
//
//	//
//	//      Menu Functions
//	//
func m_DrawThermo(x int32, y int32, thermWidth int32, thermDot int32) {
	var i, xx int32
	xx = x
	v_DrawPatchDirect(xx, y, w_CacheLumpNameT("M_THERML"))
	xx += 8
	i = 0
	for {
		if i >= thermWidth {
			break
		}
		v_DrawPatchDirect(xx, y, w_CacheLumpNameT("M_THERMM"))
		xx += 8
		goto _1
	_1:
		;
		i++
	}
	v_DrawPatchDirect(xx, y, w_CacheLumpNameT("M_THERMR"))
	v_DrawPatchDirect(x+8+thermDot*8, y, w_CacheLumpNameT("M_THERMO"))
}

func m_StartMessage(string1 string, routine func(int32), input boolean) {
	messageLastMenuActive = int32(menuactive)
	messageToPrint = 1
	messageString = string1
	if routine == nil {
		messageRoutine = nil
	} else {
		messageRoutine = &routine
	}
	messageNeedsInput = input
	menuactive = 1
}

// C documentation
//
//	//
//	// Find string width from hu_font chars
//	//
func m_StringWidth(string1 string) int32 {
	var c, w int32
	w = 0
	for i := 0; i < len(string1); i++ {
		c = xtoupper(int32(string1[i])) - '!'
		if c < 0 || c >= '_'-'!'+1 {
			w += 4
		} else {
			w += int32(hu_font[c].Fwidth)
		}
	}
	return w
}

// C documentation
//
//	//
//	//      Find string height from hu_font chars
//	//
func m_StringHeight(string1 string) int32 {
	var h, height int32
	height = int32(hu_font[0].Fheight)
	h = height
	for i := 0; i < len(string1); i++ {
		if string1[i] == '\n' {
			h += height
		}
	}
	return h
}

// C documentation
//
//	//
//	//      Write a string using the hu_font
//	//
func m_WriteText(x int32, y int32, string1 string) {
	var c, cx, cy, w int32
	cx = x
	cy = y
	for i := 0; i < len(string1); i++ {
		c = int32(string1[i])
		if c == 0 {
			break
		}
		if c == '\n' {
			cx = x
			cy += 12
			continue
		}
		c = xtoupper(c) - '!'
		if c < 0 || c >= '_'-'!'+1 {
			cx += 4
			continue
		}
		w = int32(hu_font[c].Fwidth)
		if cx+w > SCREENWIDTH {
			break
		}
		v_DrawPatchDirect(cx, cy, hu_font[c])
		cx += w
	}
}

// These keys evaluate to a "null" key in Vanilla Doom that allows weird
// jumping in the menus. Preserve this behavior for accuracy.

func isNullKey(key int32) boolean {
	return booluint32(key == int32(KEY_PAUSE1) || key == 0x80+0x3a || key == 0x80+0x46 || key == 0x80+0x45)
}

//
// CONTROL PANEL
//

// C documentation
//
//	//
//	// M_Responder
//	//
func m_Responder(ev *event_t) boolean {
	var ch, i, key int32
	// In testcontrols mode, none of the function keys should do anything
	// - the only key is escape to quit.
	if testcontrols != 0 {
		if ev.Ftype1 == Ev_quit || ev.Ftype1 == Ev_keydown && (ev.Fdata1 == key_menu_activate || ev.Fdata1 == key_menu_quit) {
			i_Quit()
			return 1
		}
		return 0
	}
	// "close" button pressed on window?
	if ev.Ftype1 == Ev_quit {
		// First click on close button = bring up quit confirm message.
		// Second click on close button = confirm quit

		if menuactive != 0 && messageToPrint != 0 && messageRoutine == &M_QuitResponse {
			M_QuitResponse(key_menu_confirm)
		} else {
			s_StartSound(nil, int32(sfx_swtchn))
			m_QuitDOOM(0)
		}
		return 1
	}
	// key is the key pressed, ch is the actual character typed
	ch = 0
	key = -1
	if ev.Ftype1 == Ev_joystick && joywait < i_GetTime() {
		if ev.Fdata3 < 0 {
			key = key_menu_up
			joywait = i_GetTime() + 5
		} else {
			if ev.Fdata3 > 0 {
				key = key_menu_down
				joywait = i_GetTime() + 5
			}
		}
		if ev.Fdata2 < 0 {
			key = key_menu_left
			joywait = i_GetTime() + 2
		} else {
			if ev.Fdata2 > 0 {
				key = key_menu_right
				joywait = i_GetTime() + 2
			}
		}
		if ev.Fdata1&1 != 0 {
			key = key_menu_forward
			joywait = i_GetTime() + 5
		}
		if ev.Fdata1&2 != 0 {
			key = key_menu_back
			joywait = i_GetTime() + 5
		}
		if joybmenu >= 0 && ev.Fdata1&(1<<joybmenu) != 0 {
			key = key_menu_activate
			joywait = i_GetTime() + 5
		}
	} else {
		if ev.Ftype1 == Ev_mouse && mousewait < i_GetTime() {
			mousey1 += ev.Fdata3
			if mousey1 < lasty-int32(30) {
				key = key_menu_down
				mousewait = i_GetTime() + 5
				lasty -= 30
				mousey1 = lasty
			} else {
				if mousey1 > lasty+int32(30) {
					key = key_menu_up
					mousewait = i_GetTime() + 5
					lasty += 30
					mousey1 = lasty
				}
			}
			mousex1 += ev.Fdata2
			if mousex1 < lastx-int32(30) {
				key = key_menu_left
				mousewait = i_GetTime() + 5
				lastx -= 30
				mousex1 = lastx
			} else {
				if mousex1 > lastx+int32(30) {
					key = key_menu_right
					mousewait = i_GetTime() + 5
					lastx += 30
					mousex1 = lastx
				}
			}
			if ev.Fdata1&1 != 0 {
				key = key_menu_forward
				mousewait = i_GetTime() + 15
			}
			if ev.Fdata1&2 != 0 {
				key = key_menu_back
				mousewait = i_GetTime() + 15
			}
		} else {
			if ev.Ftype1 == Ev_keydown {
				key = ev.Fdata1
				ch = ev.Fdata2
			}
		}
	}
	if key == -1 {
		return 0
	}
	// Save Game string input
	if saveStringEnter != 0 {
		switch key {
		case int32(KEY_BACKSPACE3):
			if saveCharIndex > 0 {
				saveCharIndex--
				savegamestrings[saveSlot] = savegamestrings[saveSlot][:saveCharIndex]
			}
		case KEY_ESCAPE:
			saveStringEnter = 0
			savegamestrings[saveSlot] = saveOldString
		case KEY_ENTER:
			saveStringEnter = 0
			if len(savegamestrings[saveSlot]) > 0 {
				m_DoSave(saveSlot)
			}
		default:
			// This is complicated.
			// Vanilla has a bug where the shift key is ignored when entering
			// a savegame name. If vanilla_keyboard_mapping is on, we want
			// to emulate this bug by using 'data1'. But if it's turned off,
			// it implies the user doesn't care about Vanilla emulation: just
			// use the correct 'data2'.
			if vanilla_keyboard_mapping != 0 {
				ch = key
			}
			ch = xtoupper(ch)
			if ch != ' ' && (ch-'!' < 0 || ch-'!' >= '_'-'!'+1) {
				break
			}
			if ch >= 32 && ch <= 127 && saveCharIndex < SAVESTRINGSIZE-1 && m_StringWidth(savegamestrings[saveSlot]) < (SAVESTRINGSIZE-2)*8 {
				savegamestrings[saveSlot] += string(ch)
				saveCharIndex++
			}
			break
		}
		return 1
	}
	// Take care of any messages that need input
	if messageToPrint != 0 {
		if messageNeedsInput != 0 {
			if key != ' ' && key != KEY_ESCAPE && key != key_menu_confirm && key != key_menu_abort {
				return 0
			}
		}
		menuactive = uint32(messageLastMenuActive)
		messageToPrint = 0
		if messageRoutine != nil {
			(*messageRoutine)(key)
		}
		menuactive = 0
		s_StartSound(nil, int32(sfx_swtchx))
		return 1
	}
	// F-Keys
	if menuactive == 0 {
		if key == key_menu_decscreen { // Screen size down
			if automapactive != 0 || chat_on != 0 {
				return 0
			}
			m_SizeDisplay(0)
			s_StartSound(nil, int32(sfx_stnmov))
			return 1
		} else {
			if key == key_menu_incscreen { // Screen size up
				if automapactive != 0 || chat_on != 0 {
					return 0
				}
				m_SizeDisplay(1)
				s_StartSound(nil, int32(sfx_stnmov))
				return 1
			} else {
				if key == key_menu_help { // Help key
					m_StartControlPanel()
					if gamemode == retail {
						currentMenu = &ReadDef2
					} else {
						currentMenu = &ReadDef1
					}
					itemOn = 0
					s_StartSound(nil, int32(sfx_swtchn))
					return 1
				} else {
					if key == key_menu_save { // Save
						m_StartControlPanel()
						s_StartSound(nil, int32(sfx_swtchn))
						m_SaveGame(0)
						return 1
					} else {
						if key == key_menu_load { // Load
							m_StartControlPanel()
							s_StartSound(nil, int32(sfx_swtchn))
							m_LoadGame(0)
							return 1
						} else {
							if key == key_menu_volume { // Sound Volume
								m_StartControlPanel()
								currentMenu = &SoundDef
								itemOn = int16(sfx_vol)
								s_StartSound(nil, int32(sfx_swtchn))
								return 1
							} else {
								if key == key_menu_detail { // Detail toggle
									m_ChangeDetail(0)
									s_StartSound(nil, int32(sfx_swtchn))
									return 1
								} else {
									if key == key_menu_qsave { // Quicksave
										s_StartSound(nil, int32(sfx_swtchn))
										m_QuickSave()
										return 1
									} else {
										if key == key_menu_endgame { // End game
											s_StartSound(nil, int32(sfx_swtchn))
											m_EndGame(0)
											return 1
										} else {
											if key == key_menu_messages { // Toggle messages
												m_ChangeMessages(0)
												s_StartSound(nil, int32(sfx_swtchn))
												return 1
											} else {
												if key == key_menu_qload { // Quickload
													s_StartSound(nil, int32(sfx_swtchn))
													m_QuickLoad()
													return 1
												} else {
													if key == key_menu_quit { // Quit DOOM
														s_StartSound(nil, int32(sfx_swtchn))
														m_QuitDOOM(0)
														return 1
													} else {
														if key == key_menu_gamma { // gamma toggle
															usegamma++
															if usegamma > 4 {
																usegamma = 0
															}
															players[consoleplayer].Fmessage = gammamsg[usegamma]
															i_SetPalette(w_CacheLumpNameBytes("PLAYPAL"))
															return 1
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	// Pop-up menu?
	if menuactive == 0 {
		if key == key_menu_activate {
			m_StartControlPanel()
			s_StartSound(nil, int32(sfx_swtchn))
			return 1
		}
		return 0
	}
	// Keys usable within menu
	if key == key_menu_down {
		// Move down to next item
		for cond := true; cond; cond = int32(currentMenu.Fmenuitems[itemOn].Fstatus) == -1 {
			if int32(itemOn)+1 > int32(currentMenu.Fnumitems)-1 {
				itemOn = 0
			} else {
				itemOn++
			}
			s_StartSound(nil, int32(sfx_pstop))
		}
		return 1
	} else {
		if key == key_menu_up {
			// Move back up to previous item
			for cond := true; cond; cond = int32(currentMenu.Fmenuitems[itemOn].Fstatus) == -1 {
				if itemOn == 0 {
					itemOn = int16(int32(currentMenu.Fnumitems) - 1)
				} else {
					itemOn--
				}
				s_StartSound(nil, int32(sfx_pstop))
			}
			return 1
		} else {
			if key == key_menu_left {
				// Slide slider left
				if currentMenu.Fmenuitems[itemOn].Froutine != nil && int32(currentMenu.Fmenuitems[itemOn].Fstatus) == 2 {
					s_StartSound(nil, int32(sfx_stnmov))
					currentMenu.Fmenuitems[itemOn].Froutine(0)
				}
				return 1
			} else {
				if key == key_menu_right {
					// Slide slider right
					if currentMenu.Fmenuitems[itemOn].Froutine != nil && int32(currentMenu.Fmenuitems[itemOn].Fstatus) == 2 {
						s_StartSound(nil, int32(sfx_stnmov))
						currentMenu.Fmenuitems[itemOn].Froutine(1)
					}
					return 1
				} else {
					if key == key_menu_forward {
						// Activate menu item
						if currentMenu.Fmenuitems[itemOn].Froutine != nil && currentMenu.Fmenuitems[itemOn].Fstatus != 0 {
							currentMenu.FlastOn = itemOn
							if int32(currentMenu.Fmenuitems[itemOn].Fstatus) == 2 {
								currentMenu.Fmenuitems[itemOn].Froutine(1) // right arrow
								s_StartSound(nil, int32(sfx_stnmov))
							} else {
								currentMenu.Fmenuitems[itemOn].Froutine(int32(itemOn))
								s_StartSound(nil, int32(sfx_pistol))
							}
						}
						return 1
					} else {
						if key == key_menu_activate {
							// Deactivate menu
							currentMenu.FlastOn = itemOn
							m_ClearMenus()
							s_StartSound(nil, int32(sfx_swtchx))
							return 1
						} else {
							if key == key_menu_back {
								// Go back to previous menu
								currentMenu.FlastOn = itemOn
								if currentMenu.FprevMenu != nil {
									currentMenu = currentMenu.FprevMenu
									itemOn = currentMenu.FlastOn
									s_StartSound(nil, int32(sfx_swtchn))
								}
								return 1
							} else {
								if ch != 0 || isNullKey(key) != 0 {
									i = int32(itemOn) + 1
									for {
										if i >= int32(currentMenu.Fnumitems) {
											break
										}
										if int32(currentMenu.Fmenuitems[i].FalphaKey) == ch {
											itemOn = int16(i)
											s_StartSound(nil, int32(sfx_pstop))
											return 1
										}
										goto _2
									_2:
										;
										i++
									}
									i = 0
									for {
										if !(i <= int32(itemOn)) {
											break
										}
										if int32(currentMenu.Fmenuitems[i].FalphaKey) == ch {
											itemOn = int16(i)
											s_StartSound(nil, int32(sfx_pstop))
											return 1
										}
										goto _3
									_3:
										;
										i++
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return 0
}

var joywait int32

var mousewait int32

var mousey1 int32

var lasty int32

var mousex1 int32

var lastx int32

// C documentation
//
//	//
//	// M_StartControlPanel
//	//
func m_StartControlPanel() {
	// intro might call this repeatedly
	if menuactive != 0 {
		return
	}
	menuactive = 1
	currentMenu = &MainDef       // JDC
	itemOn = currentMenu.FlastOn // JDC
}

// Display OPL debug messages - hack for GENMIDI development.

// C documentation
//
//	//
//	// M_Drawer
//	// Called after the view has been rendered,
//	// but before it has been blitted.
//	//
func m_Drawer() {
	var bp string
	var foundnewline, start int32
	var max uint32
	inhelpscreens = 0
	// Horiz. & Vertically center string and print it.
	if messageToPrint != 0 {
		start = 0
		y2 = int16(SCREENHEIGHT/2 - m_StringHeight(messageString)/2)
		for start < int32(len(messageString)) {
			foundnewline = 0
			for i := uint32(0); i < uint32(len(messageString[start:])); i++ {
				if messageString[start+int32(i)] == '\n' {
					bp = messageString[start : start+int32(i)]
					foundnewline = 1
					start = int32(uint32(start) + (i + 1))
					break
				}
			}
			if foundnewline == 0 {
				bp = messageString[start:]
				start = int32(uint64(start) + uint64(len(bp)))
			}
			x = int16(SCREENWIDTH/2 - m_StringWidth(bp)/2)
			m_WriteText(int32(x), int32(y2), bp)
			y2 = int16(int32(y2) + int32(hu_font[0].Fheight))
		}
		return
	}
	//if (opldev)
	//{
	//    M_DrawOPLDev();
	//}
	if menuactive == 0 {
		return
	}
	if currentMenu.Froutine != nil {
		currentMenu.Froutine()
	} // call Draw routine
	// DRAW MENU
	x = currentMenu.Fx
	y2 = currentMenu.Fy
	max = uint32(currentMenu.Fnumitems)
	for i := uint32(0); i < max; i++ {
		name := currentMenu.Fmenuitems[i].Fname[:]
		if name != "" {
			v_DrawPatchDirect(int32(x), int32(y2), w_CacheLumpNameT(name))
		}
		y2 = int16(int32(y2) + LINEHEIGHT)
	}
	// DRAW SKULL
	v_DrawPatchDirect(int32(x)+-int32(32), int32(currentMenu.Fy)-5+int32(itemOn)*LINEHEIGHT, w_CacheLumpNameT(skullName[whichSkull]))
}

var x int16

var y2 int16

// C documentation
//
//	//
//	// M_ClearMenus
//	//
func m_ClearMenus() {
	menuactive = 0
	// if (!netgame && usergame && paused)
	//       sendpause = true;
}

// C documentation
//
//	//
//	// M_SetupNextMenu
//	//
func m_SetupNextMenu(menudef *menu_t) {
	currentMenu = menudef
	itemOn = currentMenu.FlastOn
}

// C documentation
//
//	//
//	// M_Ticker
//	//
func m_Ticker() {
	var v1 int16
	skullAnimCounter--
	v1 = skullAnimCounter
	if int32(v1) <= 0 {
		whichSkull = int16(int32(whichSkull) ^ 1)
		skullAnimCounter = 8
	}
}

// C documentation
//
//	//
//	// M_Init
//	//
func m_Init() {
	currentMenu = &MainDef
	menuactive = 0
	itemOn = currentMenu.FlastOn
	whichSkull = 0
	skullAnimCounter = 10
	screenSize = screenblocks - 3
	messageToPrint = 0
	messageString = ""
	messageLastMenuActive = int32(menuactive)
	quickSaveSlot = -1
	// Here we could catch other version dependencies,
	//  like HELP1/2, and four episodes.
	switch gamemode {
	case commercial:
		// Commercial has no "read this" entry.
		MainMenu[int32(readthis)] = MainMenu[int32(quitdoom)]
		MainDef.Fnumitems--
		MainDef.Fmenuitems = append([]menuitem_t{}, MainMenu[:int32(readthis)]...)
		MainDef.Fmenuitems = append(MainDef.Fmenuitems, MainMenu[int32(readthis)+1:]...)
		NewDef.FprevMenu = &MainDef
	case shareware:
		// Episode 2 and 3 are handled,
		//  branching to an ad screen.
		fallthrough
	case registered:
	case retail:
		// We are fine.
		fallthrough
	default:
		break
	}
	// Versions of doom.exe before the Ultimate Doom release only had
	// three episodes; if we're emulating one of those then don't try
	// to show episode four. If we are, then do show episode four
	// (should crash if missing).
	if gameversion < exe_ultimate {
		EpiDef.Fnumitems--
	}
	//opldev = m_CheckParm("-opldev") > 0;
}

//
// This is used to get the local FILE:LINE info from CPP
// prior to really call the function in question.
//

//
// Create a directory
//

func m_MakeDirectory(path string) {
	os.MkdirAll(path, 0755)
}

// Check if a file exists

func m_FileExists(filename string) boolean {
	if _, err := os.Stat(filename); err == nil {
		return 1
	}
	return 0
}

//
// M_WriteFile
//

func m_WriteFile(name string, source []byte) boolean {
	if err := os.WriteFile(name, source, 0644); err != nil {
		return 0
	}
	return 1
}

// Returns the path to a temporary file of the given name, stored
// inside the system temporary directory.
//
// The returned value must be freed with z_Free after use.

func m_TempFile(s string) string {
	return "/tmp" + "/" + s
}

func m_ExtractFileBase(path string, dest []byte) {
	src := filepath.Base(path)
	// Copy up to eight characters
	// Note: Vanilla Doom exits with an error if a filename is specified
	// with a base of more than eight characters.  To remove the 8.3
	// filename limit, instead we simply truncate the name.
	copy(dest, src)
}

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Random number LUT.
//

//
// M_Random
// Returns a 0-255 number
//

var rndtable = [256]uint8{
	1:   8,
	2:   109,
	3:   220,
	4:   222,
	5:   241,
	6:   149,
	7:   107,
	8:   75,
	9:   248,
	10:  254,
	11:  140,
	12:  16,
	13:  66,
	14:  74,
	15:  21,
	16:  211,
	17:  47,
	18:  80,
	19:  242,
	20:  154,
	21:  27,
	22:  205,
	23:  128,
	24:  161,
	25:  89,
	26:  77,
	27:  36,
	28:  95,
	29:  110,
	30:  85,
	31:  48,
	32:  212,
	33:  140,
	34:  211,
	35:  249,
	36:  22,
	37:  79,
	38:  200,
	39:  50,
	40:  28,
	41:  188,
	42:  52,
	43:  140,
	44:  202,
	45:  120,
	46:  68,
	47:  145,
	48:  62,
	49:  70,
	50:  184,
	51:  190,
	52:  91,
	53:  197,
	54:  152,
	55:  224,
	56:  149,
	57:  104,
	58:  25,
	59:  178,
	60:  252,
	61:  182,
	62:  202,
	63:  182,
	64:  141,
	65:  197,
	66:  4,
	67:  81,
	68:  181,
	69:  242,
	70:  145,
	71:  42,
	72:  39,
	73:  227,
	74:  156,
	75:  198,
	76:  225,
	77:  193,
	78:  219,
	79:  93,
	80:  122,
	81:  175,
	82:  249,
	84:  175,
	85:  143,
	86:  70,
	87:  239,
	88:  46,
	89:  246,
	90:  163,
	91:  53,
	92:  163,
	93:  109,
	94:  168,
	95:  135,
	96:  2,
	97:  235,
	98:  25,
	99:  92,
	100: 20,
	101: 145,
	102: 138,
	103: 77,
	104: 69,
	105: 166,
	106: 78,
	107: 176,
	108: 173,
	109: 212,
	110: 166,
	111: 113,
	112: 94,
	113: 161,
	114: 41,
	115: 50,
	116: 239,
	117: 49,
	118: 111,
	119: 164,
	120: 70,
	121: 60,
	122: 2,
	123: 37,
	124: 171,
	125: 75,
	126: 136,
	127: 156,
	128: 11,
	129: 56,
	130: 42,
	131: 146,
	132: 138,
	133: 229,
	134: 73,
	135: 146,
	136: 77,
	137: 61,
	138: 98,
	139: 196,
	140: 135,
	141: 106,
	142: 63,
	143: 197,
	144: 195,
	145: 86,
	146: 96,
	147: 203,
	148: 113,
	149: 101,
	150: 170,
	151: 247,
	152: 181,
	153: 113,
	154: 80,
	155: 250,
	156: 108,
	157: 7,
	158: 255,
	159: 237,
	160: 129,
	161: 226,
	162: 79,
	163: 107,
	164: 112,
	165: 166,
	166: 103,
	167: 241,
	168: 24,
	169: 223,
	170: 239,
	171: 120,
	172: 198,
	173: 58,
	174: 60,
	175: 82,
	176: 128,
	177: 3,
	178: 184,
	179: 66,
	180: 143,
	181: 224,
	182: 145,
	183: 224,
	184: 81,
	185: 206,
	186: 163,
	187: 45,
	188: 63,
	189: 90,
	190: 168,
	191: 114,
	192: 59,
	193: 33,
	194: 159,
	195: 95,
	196: 28,
	197: 139,
	198: 123,
	199: 98,
	200: 125,
	201: 196,
	202: 15,
	203: 70,
	204: 194,
	205: 253,
	206: 54,
	207: 14,
	208: 109,
	209: 226,
	210: 71,
	211: 17,
	212: 161,
	213: 93,
	214: 186,
	215: 87,
	216: 244,
	217: 138,
	218: 20,
	219: 52,
	220: 123,
	221: 251,
	222: 26,
	223: 36,
	224: 17,
	225: 46,
	226: 52,
	227: 231,
	228: 232,
	229: 76,
	230: 31,
	231: 221,
	232: 84,
	233: 37,
	234: 216,
	235: 165,
	236: 212,
	237: 106,
	238: 197,
	239: 242,
	240: 98,
	241: 43,
	242: 39,
	243: 175,
	244: 254,
	245: 145,
	246: 190,
	247: 84,
	248: 118,
	249: 222,
	250: 187,
	251: 136,
	252: 120,
	253: 163,
	254: 236,
	255: 249,
}

// C documentation
//
//	// Which one is deterministic?
func p_Random() int32 {
	prndindex = (prndindex + 1) & 0xff
	return int32(rndtable[prndindex])
}

func m_Random() int32 {
	rndindex = (rndindex + 1) & 0xff
	return int32(rndtable[rndindex])
}

func m_ClearRandom() {
	var v1 int32
	v1 = 0
	prndindex = v1
	rndindex = v1
}

//
// T_MoveCeiling
//

func (c *ceiling_t) ThinkerFunc() {
	t_MoveCeiling(c)
}

func t_MoveCeiling(ceiling *ceiling_t) {
	var res result_e
	switch ceiling.Fdirection {
	case 0:
		// IN STASIS
	case 1:
		// UP
		res = t_MovePlane(ceiling.Fsector, ceiling.Fspeed, ceiling.Ftopheight, 0, 1, ceiling.Fdirection)
		if leveltime&7 == 0 {
			switch ceiling.Ftype1 {
			case int32(silentCrushAndRaise):
			default:
				s_StartSound(&ceiling.Fsector.Fsoundorg, int32(sfx_stnmov))
				// ?
				break
			}
		}
		if res == int32(pastdest) {
			switch ceiling.Ftype1 {
			case int32(raiseToHighest):
				p_RemoveActiveCeiling(ceiling)
			case int32(silentCrushAndRaise):
				s_StartSound(&ceiling.Fsector.Fsoundorg, int32(sfx_pstop))
				fallthrough
			case int32(fastCrushAndRaise):
				fallthrough
			case int32(crushAndRaise):
				ceiling.Fdirection = -1
			default:
				break
			}
		}
	case -1:
		// DOWN
		res = t_MovePlane(ceiling.Fsector, ceiling.Fspeed, ceiling.Fbottomheight, ceiling.Fcrush, 1, ceiling.Fdirection)
		if leveltime&7 == 0 {
			switch ceiling.Ftype1 {
			case int32(silentCrushAndRaise):
			default:
				s_StartSound(&ceiling.Fsector.Fsoundorg, int32(sfx_stnmov))
			}
		}
		if res == int32(pastdest) {
			switch ceiling.Ftype1 {
			case int32(silentCrushAndRaise):
				s_StartSound(&ceiling.Fsector.Fsoundorg, int32(sfx_pstop))
				fallthrough
			case int32(crushAndRaise):
				ceiling.Fspeed = 1 << FRACBITS
				fallthrough
			case int32(fastCrushAndRaise):
				ceiling.Fdirection = 1
			case int32(lowerAndCrush):
				fallthrough
			case int32(lowerToFloor):
				p_RemoveActiveCeiling(ceiling)
			default:
				break
			}
		} else { // ( res != pastdest )
			if res == int32(crushed) {
				switch ceiling.Ftype1 {
				case int32(silentCrushAndRaise):
					fallthrough
				case int32(crushAndRaise):
					fallthrough
				case int32(lowerAndCrush):
					ceiling.Fspeed = 1 << FRACBITS / 8
				default:
					break
				}
			}
		}
		break
	}
}

// C documentation
//
//	//
//	// EV_DoCeiling
//	// Move a ceiling up/down and all around!
//	//
func ev_DoCeiling(line *line_t, type1 ceiling_e) int32 {
	var rtn, secnum, v1 int32
	secnum = -1
	rtn = 0
	//	Reactivate in-stasis ceilings...for certain types.
	switch type1 {
	case int32(fastCrushAndRaise):
		fallthrough
	case int32(silentCrushAndRaise):
		fallthrough
	case int32(crushAndRaise):
		p_ActivateInStasisCeiling(line)
		fallthrough
	default:
		break
	}
	for {
		v1 = p_FindSectorFromLineTag(line, secnum)
		secnum = v1
		if !(v1 >= 0) {
			break
		}
		sec := &sectors[secnum]
		if sec.Fspecialdata != nil {
			continue
		}
		// new door thinker
		rtn = 1
		ceiling := &ceiling_t{}
		p_AddThinker(&ceiling.Fthinker)
		sec.Fspecialdata = ceiling
		ceiling.Fthinker.Ffunction = ceiling
		ceiling.Fsector = sec
		ceiling.Fcrush = 0
		switch type1 {
		case int32(fastCrushAndRaise):
			ceiling.Fcrush = 1
			ceiling.Ftopheight = sec.Fceilingheight
			ceiling.Fbottomheight = sec.Ffloorheight + 8*(1<<FRACBITS)
			ceiling.Fdirection = -1
			ceiling.Fspeed = 1 << FRACBITS * 2
		case int32(silentCrushAndRaise):
			fallthrough
		case int32(crushAndRaise):
			ceiling.Fcrush = 1
			ceiling.Ftopheight = sec.Fceilingheight
			fallthrough
		case int32(lowerAndCrush):
			fallthrough
		case int32(lowerToFloor):
			ceiling.Fbottomheight = sec.Ffloorheight
			if type1 != int32(lowerToFloor) {
				ceiling.Fbottomheight += 8 * (1 << FRACBITS)
			}
			ceiling.Fdirection = -1
			ceiling.Fspeed = 1 << FRACBITS
		case int32(raiseToHighest):
			ceiling.Ftopheight = p_FindHighestCeilingSurrounding(sec)
			ceiling.Fdirection = 1
			ceiling.Fspeed = 1 << FRACBITS
			break
		}
		ceiling.Ftag = int32(sec.Ftag)
		ceiling.Ftype1 = type1
		p_AddActiveCeiling(ceiling)
	}
	return rtn
}

// C documentation
//
//	//
//	// Add an active ceiling
//	//
func p_AddActiveCeiling(c *ceiling_t) {
	for i := int32(0); i < MAXCEILINGS; i++ {
		if activeceilings[i] == nil {
			activeceilings[i] = c
			return
		}
	}
}

// C documentation
//
//	//
//	// Remove a ceiling's thinker
//	//
func p_RemoveActiveCeiling(c *ceiling_t) {
	for i := 0; i < MAXCEILINGS; i++ {
		if activeceilings[i] == c {
			activeceilings[i].Fsector.Fspecialdata = nil
			p_RemoveThinker(&activeceilings[i].Fthinker)
			activeceilings[i] = nil
			break
		}
	}
}

// C documentation
//
//	//
//	// Restart a ceiling that's in-stasis
//	//
func p_ActivateInStasisCeiling(line *line_t) {
	for i := int32(0); i < MAXCEILINGS; i++ {
		if activeceilings[i] != nil && activeceilings[i].Ftag == int32(line.Ftag) && activeceilings[i].Fdirection == 0 {
			activeceilings[i].Fdirection = activeceilings[i].Folddirection
			activeceilings[i].Fthinker.Ffunction = activeceilings[i]
		}
	}
}

// C documentation
//
//	//
//	// EV_CeilingCrushStop
//	// Stop a ceiling from crushing!
//	//
func ev_CeilingCrushStop(line *line_t) int32 {
	var rtn int32
	rtn = 0
	for i := int32(0); i < MAXCEILINGS; i++ {
		if activeceilings[i] != nil && activeceilings[i].Ftag == int32(line.Ftag) && activeceilings[i].Fdirection != 0 {
			activeceilings[i].Folddirection = activeceilings[i].Fdirection
			activeceilings[i].Fthinker.Ffunction = nil
			activeceilings[i].Fdirection = 0 // in-stasis
			rtn = 1
		}
	}
	return rtn
}

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

//
// VERTICAL DOORS
//

// C documentation
//
//	//
//	// T_VerticalDoor
//	//
func (door *vldoor_t) ThinkerFunc() {
	t_VerticalDoor(door)
}
func t_VerticalDoor(door *vldoor_t) {
	var res result_e
	switch door.Fdirection {
	case 0:
		// WAITING
		door.Ftopcountdown--
		if door.Ftopcountdown == 0 {
			switch door.Ftype1 {
			case int32(vld_blazeRaise):
				door.Fdirection = -1 // time to go back down
				s_StartSound(&door.Fsector.Fsoundorg, int32(sfx_bdcls))
			case int32(vld_normal):
				door.Fdirection = -1 // time to go back down
				s_StartSound(&door.Fsector.Fsoundorg, int32(sfx_dorcls))
			case int32(vld_close30ThenOpen):
				door.Fdirection = 1
				s_StartSound(&door.Fsector.Fsoundorg, int32(sfx_doropn))
			default:
				break
			}
		}
	case 2:
		//  INITIAL WAIT
		door.Ftopcountdown--
		if door.Ftopcountdown == 0 {
			switch door.Ftype1 {
			case int32(vld_raiseIn5Mins):
				door.Fdirection = 1
				door.Ftype1 = int32(vld_normal)
				s_StartSound(&door.Fsector.Fsoundorg, int32(sfx_doropn))
			default:
				break
			}
		}
	case -1:
		// DOWN
		res = t_MovePlane(door.Fsector, door.Fspeed, door.Fsector.Ffloorheight, 0, 1, door.Fdirection)
		if res == int32(pastdest) {
			switch door.Ftype1 {
			case int32(vld_blazeRaise):
				fallthrough
			case int32(vld_blazeClose):
				door.Fsector.Fspecialdata = nil
				p_RemoveThinker(&door.Fthinker) // unlink and free
				s_StartSound(&door.Fsector.Fsoundorg, int32(sfx_bdcls))
			case int32(vld_normal):
				fallthrough
			case int32(vld_close):
				door.Fsector.Fspecialdata = nil
				p_RemoveThinker(&door.Fthinker) // unlink and free
			case int32(vld_close30ThenOpen):
				door.Fdirection = 0
				door.Ftopcountdown = TICRATE * 30
			default:
				break
			}
		} else {
			if res == int32(crushed) {
				switch door.Ftype1 {
				case int32(vld_blazeClose):
					fallthrough
				case int32(vld_close): // DO NOT GO BACK UP!
				default:
					door.Fdirection = 1
					s_StartSound(&door.Fsector.Fsoundorg, int32(sfx_doropn))
					break
				}
			}
		}
	case 1:
		// UP
		res = t_MovePlane(door.Fsector, door.Fspeed, door.Ftopheight, 0, 1, door.Fdirection)
		if res == int32(pastdest) {
			switch door.Ftype1 {
			case int32(vld_blazeRaise):
				fallthrough
			case int32(vld_normal):
				door.Fdirection = 0 // wait at top
				door.Ftopcountdown = door.Ftopwait
			case int32(vld_close30ThenOpen):
				fallthrough
			case int32(vld_blazeOpen):
				fallthrough
			case int32(vld_open):
				door.Fsector.Fspecialdata = nil
				p_RemoveThinker(&door.Fthinker) // unlink and free
			default:
				break
			}
		}
		break
	}
}

//
// EV_DoLockedDoor
// Move a locked door up/down
//

func ev_DoLockedDoor(line *line_t, type1 vldoor_e, thing *mobj_t) int32 {
	var p *player_t
	p = thing.Fplayer
	if p == nil {
		return 0
	}
	switch int32(line.Fspecial) {
	case 99: // Blue Lock
		fallthrough
	case 133:
		if p == nil {
			return 0
		}
		if p.Fcards[it_bluecard] == 0 && p.Fcards[it_blueskull] == 0 {
			p.Fmessage = "You need a blue key to activate this object"
			s_StartSound(nil, int32(sfx_oof))
			return 0
		}
	case 134: // Red Lock
		fallthrough
	case 135:
		if p == nil {
			return 0
		}
		if p.Fcards[it_redcard] == 0 && p.Fcards[it_redskull] == 0 {
			p.Fmessage = "You need a red key to activate this object"
			s_StartSound(nil, int32(sfx_oof))
			return 0
		}
	case 136: // Yellow Lock
		fallthrough
	case 137:
		if p == nil {
			return 0
		}
		if p.Fcards[it_yellowcard] == 0 && p.Fcards[it_yellowskull] == 0 {
			p.Fmessage = "You need a yellow key to activate this object"
			s_StartSound(nil, int32(sfx_oof))
			return 0
		}
		break
	}
	return ev_DoDoor(line, type1)
}

func ev_DoDoor(line *line_t, type1 vldoor_e) int32 {
	var rtn, secnum, v1 int32
	secnum = -1
	rtn = 0
	for {
		v1 = p_FindSectorFromLineTag(line, secnum)
		secnum = v1
		if !(v1 >= 0) {
			break
		}
		sec := &sectors[secnum]
		if sec.Fspecialdata != nil {
			continue
		}
		// new door thinker
		rtn = 1
		doorP := &vldoor_t{}
		p_AddThinker(&doorP.Fthinker)
		sec.Fspecialdata = doorP
		doorP.Fthinker.Ffunction = doorP
		doorP.Fsector = sec
		doorP.Ftype1 = type1
		doorP.Ftopwait = VDOORWAIT
		doorP.Fspeed = 1 << FRACBITS * 2
		switch type1 {
		case int32(vld_blazeClose):
			doorP.Ftopheight = p_FindLowestCeilingSurrounding(sec)
			doorP.Ftopheight -= 4 * (1 << FRACBITS)
			doorP.Fdirection = -1
			doorP.Fspeed = 1 << FRACBITS * 2 * 4
			s_StartSound(&doorP.Fsector.Fsoundorg, int32(sfx_bdcls))
		case int32(vld_close):
			doorP.Ftopheight = p_FindLowestCeilingSurrounding(sec)
			doorP.Ftopheight -= 4 * (1 << FRACBITS)
			doorP.Fdirection = -1
			s_StartSound(&doorP.Fsector.Fsoundorg, int32(sfx_dorcls))
		case int32(vld_close30ThenOpen):
			doorP.Ftopheight = sec.Fceilingheight
			doorP.Fdirection = -1
			s_StartSound(&doorP.Fsector.Fsoundorg, int32(sfx_dorcls))
		case int32(vld_blazeRaise):
			fallthrough
		case int32(vld_blazeOpen):
			doorP.Fdirection = 1
			doorP.Ftopheight = p_FindLowestCeilingSurrounding(sec)
			doorP.Ftopheight -= 4 * (1 << FRACBITS)
			doorP.Fspeed = 1 << FRACBITS * 2 * 4
			if doorP.Ftopheight != sec.Fceilingheight {
				s_StartSound(&doorP.Fsector.Fsoundorg, int32(sfx_bdopn))
			}
		case int32(vld_normal):
			fallthrough
		case int32(vld_open):
			doorP.Fdirection = 1
			doorP.Ftopheight = p_FindLowestCeilingSurrounding(sec)
			doorP.Ftopheight -= 4 * (1 << FRACBITS)
			if doorP.Ftopheight != sec.Fceilingheight {
				s_StartSound(&doorP.Fsector.Fsoundorg, int32(sfx_doropn))
			}
		default:
			break
		}
	}
	return rtn
}

// C documentation
//
//	//
//	// ev_VerticalDoor : open a door manually, no tag value
//	//
func ev_VerticalDoor(line *line_t, thing *mobj_t) {
	var player *player_t
	var sec *sector_t
	var side int32
	side = 0 // only front sides can be used
	//	Check for locks
	player = thing.Fplayer
	switch int32(line.Fspecial) {
	case 26: // Blue Lock
		fallthrough
	case 32:
		if player == nil {
			return
		}
		if player.Fcards[it_bluecard] == 0 && player.Fcards[it_blueskull] == 0 {
			player.Fmessage = "You need a blue key to open this door"
			s_StartSound(nil, int32(sfx_oof))
			return
		}
	case 27: // Yellow Lock
		fallthrough
	case 34:
		if player == nil {
			return
		}
		if player.Fcards[it_yellowcard] == 0 && player.Fcards[it_yellowskull] == 0 {
			player.Fmessage = "You need a yellow key to open this door"
			s_StartSound(nil, int32(sfx_oof))
			return
		}
	case 28: // Red Lock
		fallthrough
	case 33:
		if player == nil {
			return
		}
		if player.Fcards[it_redcard] == 0 && player.Fcards[it_redskull] == 0 {
			player.Fmessage = "You need a red key to open this door"
			s_StartSound(nil, int32(sfx_oof))
			return
		}
		break
	}
	// if the sector has an active thinker, use it
	sec = sides[line.Fsidenum[side^1]].Fsector
	if sec.Fspecialdata != nil {
		special := sec.Fspecialdata
		switch int32(line.Fspecial) {
		case 1: // ONLY FOR "RAISE" DOORS, NOT "OPEN"s
			fallthrough
		case 26:
			fallthrough
		case 27:
			fallthrough
		case 28:
			fallthrough
		case 117:
			if doorP, ok := special.(*vldoor_t); ok {
				if doorP.Fdirection == -1 {
					doorP.Fdirection = 1
				} else {
					if thing.Fplayer == nil {
						return
					} // JDC: bad guys never close doors
					// When is a door not a door?
					// In Vanilla, door->direction is set, even though
					// "specialdata" might not actually point at a door.
					doorP.Fdirection = -1 // start going down immediately
				}
			} else if platP, ok := special.(*plat_t); ok {
				platP.Fwait = -1
			} else {
				// This isn't a door OR a plat.  Now we're in trouble.
				fprintf_ccgo(os.Stderr, "ev_VerticalDoor: Tried to close something that wasn't a door.\n")
			}
			return
		}
	}
	// for proper sound
	switch int32(line.Fspecial) {
	case 117: // BLAZING DOOR RAISE
		fallthrough
	case 118: // BLAZING DOOR OPN
		s_StartSound(&sec.Fsoundorg, int32(sfx_bdopn))
	case 1: // NORMAL DOOR SOUND
		fallthrough
	case 31:
		s_StartSound(&sec.Fsoundorg, int32(sfx_doropn))
	default: // LOCKED DOOR SOUND
		s_StartSound(&sec.Fsoundorg, int32(sfx_doropn))
		break
	}
	// new door thinker
	doorP := &vldoor_t{}
	p_AddThinker(&doorP.Fthinker)
	sec.Fspecialdata = doorP
	doorP.Fthinker.Ffunction = doorP
	doorP.Fsector = sec
	doorP.Fdirection = 1
	doorP.Fspeed = 1 << FRACBITS * 2
	doorP.Ftopwait = VDOORWAIT
	switch int32(line.Fspecial) {
	case 1:
		fallthrough
	case 26:
		fallthrough
	case 27:
		fallthrough
	case 28:
		doorP.Ftype1 = int32(vld_normal)
	case 31:
		fallthrough
	case 32:
		fallthrough
	case 33:
		fallthrough
	case 34:
		doorP.Ftype1 = int32(vld_open)
		line.Fspecial = 0
	case 117: // blazing door raise
		doorP.Ftype1 = int32(vld_blazeRaise)
		doorP.Fspeed = 1 << FRACBITS * 2 * 4
	case 118: // blazing door open
		doorP.Ftype1 = int32(vld_blazeOpen)
		line.Fspecial = 0
		doorP.Fspeed = 1 << FRACBITS * 2 * 4
		break
	}
	// find the top and bottom of the movement range
	doorP.Ftopheight = p_FindLowestCeilingSurrounding(sec)
	doorP.Ftopheight -= 4 * (1 << FRACBITS)
}

// C documentation
//
//	//
//	// Spawn a door that closes after 30 seconds
//	//
func p_SpawnDoorCloseIn30(sec *sector_t) {
	doorP := &vldoor_t{}
	p_AddThinker(&doorP.Fthinker)
	sec.Fspecialdata = doorP
	sec.Fspecial = 0
	doorP.Fthinker.Ffunction = doorP
	doorP.Fsector = sec
	doorP.Fdirection = 0
	doorP.Ftype1 = int32(vld_normal)
	doorP.Fspeed = 1 << FRACBITS * 2
	doorP.Ftopcountdown = 30 * TICRATE
}

// C documentation
//
//	//
//	// Spawn a door that opens after 5 minutes
//	//
func p_SpawnDoorRaiseIn5Mins(sec *sector_t, secnum int32) {
	doorP := &vldoor_t{}
	p_AddThinker(&doorP.Fthinker)
	sec.Fspecialdata = doorP
	sec.Fspecial = 0
	doorP.Fthinker.Ffunction = doorP
	doorP.Fsector = sec
	doorP.Fdirection = 2
	doorP.Ftype1 = int32(vld_raiseIn5Mins)
	doorP.Fspeed = 1 << FRACBITS * 2
	doorP.Ftopheight = p_FindLowestCeilingSurrounding(sec)
	doorP.Ftopheight -= 4 * (1 << FRACBITS)
	doorP.Ftopwait = VDOORWAIT
	doorP.Ftopcountdown = 5 * 60 * TICRATE
}

const ANG1801 = 2147483648
const ANG2703 = 3221225472
const ANG903 = 1073741824

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// Data.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

type dirtype_t = int32

const DI_EAST = 0
const DI_NORTHEAST = 1
const DI_NORTH = 2
const DI_NORTHWEST = 3
const DI_WEST = 4
const DI_SOUTHWEST = 5
const DI_SOUTH = 6
const DI_SOUTHEAST = 7
const DI_NODIR = 8

func init() {
	opposite = [9]dirtype_t{
		0: DI_WEST,
		1: DI_SOUTHWEST,
		2: DI_SOUTH,
		3: DI_SOUTHEAST,
		5: DI_NORTHEAST,
		6: DI_NORTH,
		7: DI_NORTHWEST,
		8: DI_NODIR,
	}
}

func init() {
	diags = [4]dirtype_t{
		0: DI_NORTHWEST,
		1: DI_NORTHEAST,
		2: DI_SOUTHWEST,
		3: DI_SOUTHEAST,
	}
}

func p_RecursiveSound(sec *sector_t, soundblocks int32) {
	var check *line_t
	var other *sector_t
	// wake up all monsters in this sector
	if sec.Fvalidcount == validcount && sec.Fsoundtraversed <= soundblocks+1 {
		return // already flooded
	}
	sec.Fvalidcount = validcount
	sec.Fsoundtraversed = soundblocks + 1
	sec.Fsoundtarget = soundtarget
	for i := int32(0); i < sec.Flinecount; i++ {
		check = sec.Flines[i]
		if int32(check.Fflags)&ml_TWOSIDED == 0 {
			continue
		}
		p_LineOpening(check)
		if openrange <= 0 {
			continue
		} // closed door
		if sides[check.Fsidenum[0]].Fsector == sec {
			other = sides[check.Fsidenum[1]].Fsector
		} else {
			other = sides[check.Fsidenum[0]].Fsector
		}
		if int32(check.Fflags)&ml_SOUNDBLOCK != 0 {
			if soundblocks == 0 {
				p_RecursiveSound(other, 1)
			}
		} else {
			p_RecursiveSound(other, soundblocks)
		}
	}
}

// C documentation
//
//	//
//	// P_NoiseAlert
//	// If a monster yells at a player,
//	// it will alert other monsters to the player.
//	//
func p_NoiseAlert(target *mobj_t, emmiter *mobj_t) {
	soundtarget = target
	validcount++
	p_RecursiveSound(emmiter.Fsubsector.Fsector, 0)
}

// C documentation
//
//	//
//	// P_CheckMeleeRange
//	//
func p_CheckMeleeRange(actor *mobj_t) boolean {
	var dist fixed_t
	var pl *mobj_t
	if actor.Ftarget == nil {
		return 0
	}
	pl = actor.Ftarget
	dist = p_AproxDistance(pl.Fx-actor.Fx, pl.Fy-actor.Fy)
	if dist >= 64*(1<<FRACBITS)-20*(1<<FRACBITS)+pl.Finfo.Fradius {
		return 0
	}
	if p_CheckSight(actor, actor.Ftarget) == 0 {
		return 0
	}
	return 1
}

// C documentation
//
//	//
//	// P_CheckMissileRange
//	//
func p_CheckMissileRange(actor *mobj_t) boolean {
	var dist fixed_t
	if p_CheckSight(actor, actor.Ftarget) == 0 {
		return 0
	}
	if actor.Fflags&mf_JUSTHIT != 0 {
		// the target just hit the enemy,
		// so fight back!
		actor.Fflags &^= mf_JUSTHIT
		return 1
	}
	if actor.Freactiontime != 0 {
		return 0
	} // do not attack yet
	// OPTIMIZE: get this from a global checksight
	dist = p_AproxDistance(actor.Fx-actor.Ftarget.Fx, actor.Fy-actor.Ftarget.Fy) - 64*(1<<FRACBITS)
	if actor.Finfo.Fmeleestate == 0 {
		dist -= 128 * (1 << FRACBITS)
	} // no melee attack, so fire more
	dist >>= 16
	if actor.Ftype1 == mt_VILE {
		if dist > 14*64 {
			return 0
		} // too far away
	}
	if actor.Ftype1 == mt_UNDEAD {
		if dist < 196 {
			return 0
		} // close for fist attack
		dist >>= 1
	}
	if actor.Ftype1 == mt_CYBORG || actor.Ftype1 == mt_SPIDER || actor.Ftype1 == mt_SKULL {
		dist >>= 1
	}
	if dist > 200 {
		dist = 200
	}
	if actor.Ftype1 == mt_CYBORG && dist > 160 {
		dist = 160
	}
	if p_Random() < dist {
		return 0
	}
	return 1
}

func init() {
	xspeed = [8]fixed_t{
		0: 1 << FRACBITS,
		1: 47000,
		3: -int32(47000),
		4: -(1 << FRACBITS),
		5: -int32(47000),
		7: 47000,
	}
}

func init() {
	yspeed = [8]fixed_t{
		1: 47000,
		2: 1 << FRACBITS,
		3: 47000,
		5: -int32(47000),
		6: -(1 << FRACBITS),
		7: -int32(47000),
	}
}

func p_Move(actor *mobj_t) boolean {
	var good, try_ok boolean
	var tryx, tryy fixed_t
	var v1 int32
	if actor.Fmovedir == DI_NODIR {
		return 0
	}
	if uint32(actor.Fmovedir) >= 8 {
		i_Error("Weird actor->movedir!")
	}
	tryx = actor.Fx + actor.Finfo.Fspeed*xspeed[actor.Fmovedir]
	tryy = actor.Fy + actor.Finfo.Fspeed*yspeed[actor.Fmovedir]
	try_ok = p_TryMove(actor, tryx, tryy)
	if try_ok == 0 {
		// open any specials
		if actor.Fflags&mf_FLOAT != 0 && floatok != 0 {
			// must adjust height
			if actor.Fz < tmfloorz {
				actor.Fz += 1 << FRACBITS * 4
			} else {
				actor.Fz -= 1 << FRACBITS * 4
			}
			actor.Fflags |= mf_INFLOAT
			return 1
		}
		if numspechit == 0 {
			return 0
		}
		actor.Fmovedir = DI_NODIR
		good = 0
		for {
			v1 = numspechit
			numspechit--
			if v1 == 0 {
				break
			}
			ld := spechit[numspechit]
			// if the special is not a door
			// that can be opened,
			// return false
			if p_UseSpecialLine(actor, ld, 0) != 0 {
				good = 1
			}
		}
		return good
	} else {
		actor.Fflags &^= mf_INFLOAT
	}
	if actor.Fflags&mf_FLOAT == 0 {
		actor.Fz = actor.Ffloorz
	}
	return 1
}

// C documentation
//
//	//
//	// TryWalk
//	// Attempts to move actor on
//	// in its current (ob->moveangle) direction.
//	// If blocked by either a wall or an actor
//	// returns FALSE
//	// If move is either clear or blocked only by a door,
//	// returns TRUE and sets...
//	// If a door is in the way,
//	// an OpenDoor call is made to start it opening.
//	//
func p_TryWalk(actor *mobj_t) boolean {
	if p_Move(actor) == 0 {
		return 0
	}
	actor.Fmovecount = p_Random() & 15
	return 1
}

func p_NewChaseDir(actor *mobj_t) {
	var d [3]dirtype_t
	var deltax, deltay fixed_t
	var olddir, turnaround dirtype_t
	var tdir int32
	if actor.Ftarget == nil {
		i_Error("p_NewChaseDir: called with no target")
	}
	olddir = actor.Fmovedir
	turnaround = opposite[olddir]
	deltax = actor.Ftarget.Fx - actor.Fx
	deltay = actor.Ftarget.Fy - actor.Fy
	if deltax > 10*(1<<FRACBITS) {
		d[1] = DI_EAST
	} else {
		if deltax < -10*(1<<FRACBITS) {
			d[1] = DI_WEST
		} else {
			d[1] = DI_NODIR
		}
	}
	if deltay < -10*(1<<FRACBITS) {
		d[2] = DI_SOUTH
	} else {
		if deltay > 10*(1<<FRACBITS) {
			d[2] = DI_NORTH
		} else {
			d[2] = DI_NODIR
		}
	}
	// try direct route
	if d[1] != DI_NODIR && d[2] != DI_NODIR {
		actor.Fmovedir = diags[boolint32(deltay < 0)<<1+boolint32(deltax > 0)]
		if actor.Fmovedir != turnaround && p_TryWalk(actor) != 0 {
			return
		}
	}
	// try other directions
	if p_Random() > 200 || xabs(deltay) > xabs(deltax) {
		tdir = d[1]
		d[1] = d[2]
		d[2] = tdir
	}
	if d[1] == turnaround {
		d[1] = DI_NODIR
	}
	if d[2] == turnaround {
		d[2] = DI_NODIR
	}
	if d[1] != DI_NODIR {
		actor.Fmovedir = d[1]
		if p_TryWalk(actor) != 0 {
			// either moved forward or attacked
			return
		}
	}
	if d[2] != DI_NODIR {
		actor.Fmovedir = d[2]
		if p_TryWalk(actor) != 0 {
			return
		}
	}
	// there is no direct path to the player,
	// so pick another direction.
	if olddir != DI_NODIR {
		actor.Fmovedir = olddir
		if p_TryWalk(actor) != 0 {
			return
		}
	}
	// randomly determine direction of search
	if p_Random()&1 != 0 {
		tdir = DI_EAST
		for {
			if !(tdir <= DI_SOUTHEAST) {
				break
			}
			if tdir != turnaround {
				actor.Fmovedir = tdir
				if p_TryWalk(actor) != 0 {
					return
				}
			}
			goto _1
		_1:
			;
			tdir++
		}
	} else {
		tdir = DI_SOUTHEAST
		for {
			if tdir == DI_EAST-1 {
				break
			}
			if tdir != turnaround {
				actor.Fmovedir = tdir
				if p_TryWalk(actor) != 0 {
					return
				}
			}
			goto _2
		_2:
			;
			tdir--
		}
	}
	if turnaround != DI_NODIR {
		actor.Fmovedir = turnaround
		if p_TryWalk(actor) != 0 {
			return
		}
	}
	actor.Fmovedir = DI_NODIR // can not move
}

// C documentation
//
//	//
//	// P_LookForPlayers
//	// If allaround is false, only look 180 degrees in front.
//	// Returns true if a player is targeted.
//	//
func p_LookForPlayers(actor *mobj_t, allaround boolean) boolean {
	var an angle_t
	var c, stop, v2 int32
	var dist fixed_t
	var player *player_t
	c = 0
	stop = (actor.Flastlook - 1) & 3
	for {
		if playeringame[actor.Flastlook] == 0 {
			goto _1
		}
		v2 = c
		c++
		if v2 == 2 || actor.Flastlook == stop {
			// done looking
			return 0
		}
		player = &players[actor.Flastlook]
		if player.Fhealth <= 0 {
			goto _1
		} // dead
		if p_CheckSight(actor, player.Fmo) == 0 {
			goto _1
		} // out of sight
		if allaround == 0 {
			an = r_PointToAngle2(actor.Fx, actor.Fy, player.Fmo.Fx, player.Fmo.Fy) - actor.Fangle
			if an > ANG903 && an < ANG2703 {
				dist = p_AproxDistance(player.Fmo.Fx-actor.Fx, player.Fmo.Fy-actor.Fy)
				// if real close, react anyway
				if dist > 64*(1<<FRACBITS) {
					goto _1
				} // behind back
			}
		}
		actor.Ftarget = player.Fmo
		return 1
		goto _1
	_1:
		;
		actor.Flastlook = (actor.Flastlook + 1) & 3
	}
	return 0
}

// C documentation
//
//	//
//	// A_KeenDie
//	// DOOM II special, map 32.
//	// Uses special tag 666.
//	//
func a_KeenDie(mo *mobj_t) {
	var th *thinker_t
	a_Fall(mo)
	// scan the remaining thinkers
	// to see if all Keens are dead
	th = thinkercap.Fnext
	for {
		if th == &thinkercap {
			break
		}
		mo2, ok := th.Ffunction.(*mobj_t)
		if !ok {
			goto _1
		}
		if mo2 != mo && mo2.Ftype1 == mo.Ftype1 && mo2.Fhealth > 0 {
			// other Keen not dead
			return
		}
		goto _1
	_1:
		;
		th = th.Fnext
	}
	line := &line_t{Ftag: 666}
	ev_DoDoor(line, int32(vld_open))
}

//
// ACTION ROUTINES
//

// C documentation
//
//	//
//	// A_Look
//	// Stay in state until a player is sighted.
//	//
func a_Look(actor *mobj_t) {
	var sound int32
	var targ *mobj_t
	actor.Fthreshold = 0 // any shot will wake up
	targ = actor.Fsubsector.Fsector.Fsoundtarget
	if targ != nil && targ.Fflags&mf_SHOOTABLE != 0 {
		actor.Ftarget = targ
		if actor.Fflags&mf_AMBUSH != 0 {
			if p_CheckSight(actor, actor.Ftarget) != 0 {
				goto seeyou
			}
		} else {
			goto seeyou
		}
	}
	if p_LookForPlayers(actor, 0) == 0 {
		return
	}
	// go into chase state
	goto seeyou
seeyou:
	;
	if actor.Finfo.Fseesound != 0 {
		switch actor.Finfo.Fseesound {
		case int32(sfx_posit1):
			fallthrough
		case int32(sfx_posit2):
			fallthrough
		case int32(sfx_posit3):
			sound = int32(sfx_posit1) + p_Random()%3
		case int32(sfx_bgsit1):
			fallthrough
		case int32(sfx_bgsit2):
			sound = int32(sfx_bgsit1) + p_Random()%2
		default:
			sound = actor.Finfo.Fseesound
			break
		}
		if actor.Ftype1 == mt_SPIDER || actor.Ftype1 == mt_CYBORG {
			// full volume
			s_StartSound(nil, sound)
		} else {
			s_StartSound(&actor.degenmobj_t, sound)
		}
	}
	p_SetMobjState(actor, actor.Finfo.Fseestate)
}

// C documentation
//
//	//
//	// A_Chase
//	// Actor has a melee attack,
//	// so it tries to close as fast as possible
//	//
func a_Chase(actor *mobj_t) {
	var delta int32
	if actor.Freactiontime != 0 {
		actor.Freactiontime--
	}
	// modify target threshold
	if actor.Fthreshold != 0 {
		if actor.Ftarget == nil || actor.Ftarget.Fhealth <= 0 {
			actor.Fthreshold = 0
		} else {
			actor.Fthreshold--
		}
	}
	// turn towards movement direction if not there yet
	if actor.Fmovedir < 8 {
		actor.Fangle &= 7 << 29
		delta = int32(actor.Fangle - uint32(actor.Fmovedir<<29))
		if delta > 0 {
			actor.Fangle -= uint32(ANG903 / 2)
		} else {
			if delta < 0 {
				actor.Fangle += uint32(ANG903 / 2)
			}
		}
	}
	if actor.Ftarget == nil || actor.Ftarget.Fflags&mf_SHOOTABLE == 0 {
		// look for a new target
		if p_LookForPlayers(actor, 1) != 0 {
			return
		} // got a new target
		p_SetMobjState(actor, actor.Finfo.Fspawnstate)
		return
	}
	// do not attack twice in a row
	if actor.Fflags&mf_JUSTATTACKED != 0 {
		actor.Fflags &= ^mf_JUSTATTACKED
		if gameskill != sk_nightmare && fastparm == 0 {
			p_NewChaseDir(actor)
		}
		return
	}
	// check for melee attack
	if actor.Finfo.Fmeleestate != 0 && p_CheckMeleeRange(actor) != 0 {
		if actor.Finfo.Fattacksound != 0 {
			s_StartSound(&actor.degenmobj_t, actor.Finfo.Fattacksound)
		}
		p_SetMobjState(actor, actor.Finfo.Fmeleestate)
		return
	}
	// check for missile attack
	if actor.Finfo.Fmissilestate != 0 {
		if gameskill < sk_nightmare && fastparm == 0 && actor.Fmovecount != 0 {
			goto nomissile
		}
		if p_CheckMissileRange(actor) == 0 {
			goto nomissile
		}
		p_SetMobjState(actor, actor.Finfo.Fmissilestate)
		actor.Fflags |= mf_JUSTATTACKED
		return
	}
	// ?
	goto nomissile
nomissile:
	;
	// possibly choose another target
	if netgame != 0 && actor.Fthreshold == 0 && p_CheckSight(actor, actor.Ftarget) == 0 {
		if p_LookForPlayers(actor, 1) != 0 {
			return
		} // got a new target
	}
	// chase towards player
	actor.Fmovecount--
	if actor.Fmovecount < 0 || p_Move(actor) == 0 {
		p_NewChaseDir(actor)
	}
	// make active sound
	if actor.Finfo.Factivesound != 0 && p_Random() < 3 {
		s_StartSound(&actor.degenmobj_t, actor.Finfo.Factivesound)
	}
}

// C documentation
//
//	//
//	// A_FaceTarget
//	//
func a_FaceTarget(actor *mobj_t) {
	if actor.Ftarget == nil {
		return
	}
	actor.Fflags |= mf_AMBUSH
	actor.Fangle = r_PointToAngle2(actor.Fx, actor.Fy, actor.Ftarget.Fx, actor.Ftarget.Fy)
	if actor.Ftarget.Fflags&mf_SHADOW != 0 {
		actor.Fangle += uint32((p_Random() - p_Random()) << 21)
	}
}

// C documentation
//
//	//
//	// A_PosAttack
//	//
func a_PosAttack(actor *mobj_t) {
	var angle, damage, slope int32
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	angle = int32(actor.Fangle)
	slope = p_AimLineAttack(actor, uint32(angle), 32*64*(1<<FRACBITS))
	s_StartSound(&actor.degenmobj_t, int32(sfx_pistol))
	angle += (p_Random() - p_Random()) << 20
	damage = (p_Random()%5 + 1) * 3
	p_LineAttack(actor, uint32(angle), 32*64*(1<<FRACBITS), slope, damage)
}

func a_SPosAttack(actor *mobj_t) {
	var angle, bangle, damage, slope int32
	if actor.Ftarget == nil {
		return
	}
	s_StartSound(&actor.degenmobj_t, int32(sfx_shotgn))
	a_FaceTarget(actor)
	bangle = int32(actor.Fangle)
	slope = p_AimLineAttack(actor, uint32(bangle), 32*64*(1<<FRACBITS))
	for i := 0; i < 3; i++ {
		angle = bangle + (p_Random()-p_Random())<<int32(20)
		damage = (p_Random()%5 + 1) * 3
		p_LineAttack(actor, uint32(angle), 32*64*(1<<FRACBITS), slope, damage)
	}
}

func a_CPosAttack(actor *mobj_t) {
	var angle, bangle, damage, slope int32
	if actor.Ftarget == nil {
		return
	}
	s_StartSound(&actor.degenmobj_t, int32(sfx_shotgn))
	a_FaceTarget(actor)
	bangle = int32(actor.Fangle)
	slope = p_AimLineAttack(actor, uint32(bangle), 32*64*(1<<FRACBITS))
	angle = bangle + (p_Random()-p_Random())<<int32(20)
	damage = (p_Random()%5 + 1) * 3
	p_LineAttack(actor, uint32(angle), 32*64*(1<<FRACBITS), slope, damage)
}

func a_CPosRefire(actor *mobj_t) {
	// keep firing unless target got out of sight
	a_FaceTarget(actor)
	if p_Random() < 40 {
		return
	}
	if actor.Ftarget == nil || actor.Ftarget.Fhealth <= 0 || p_CheckSight(actor, actor.Ftarget) == 0 {
		p_SetMobjState(actor, actor.Finfo.Fseestate)
	}
}

func a_SpidRefire(actor *mobj_t) {
	// keep firing unless target got out of sight
	a_FaceTarget(actor)
	if p_Random() < 10 {
		return
	}
	if actor.Ftarget == nil || actor.Ftarget.Fhealth <= 0 || p_CheckSight(actor, actor.Ftarget) == 0 {
		p_SetMobjState(actor, actor.Finfo.Fseestate)
	}
}

func a_BspiAttack(actor *mobj_t) {
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	// launch a missile
	p_SpawnMissile(actor, actor.Ftarget, mt_ARACHPLAZ)
}

// C documentation
//
//	//
//	// A_TroopAttack
//	//
func a_TroopAttack(actor *mobj_t) {
	var damage int32
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	if p_CheckMeleeRange(actor) != 0 {
		s_StartSound(&actor.degenmobj_t, int32(sfx_claw))
		damage = (p_Random()%8 + 1) * 3
		p_DamageMobj(actor.Ftarget, actor, actor, damage)
		return
	}
	// launch a missile
	p_SpawnMissile(actor, actor.Ftarget, mt_TROOPSHOT)
}

func a_SargAttack(actor *mobj_t) {
	var damage int32
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	if p_CheckMeleeRange(actor) != 0 {
		damage = (p_Random()%int32(10) + 1) * 4
		p_DamageMobj(actor.Ftarget, actor, actor, damage)
	}
}

func a_HeadAttack(actor *mobj_t) {
	var damage int32
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	if p_CheckMeleeRange(actor) != 0 {
		damage = (p_Random()%6 + 1) * 10
		p_DamageMobj(actor.Ftarget, actor, actor, damage)
		return
	}
	// launch a missile
	p_SpawnMissile(actor, actor.Ftarget, mt_HEADSHOT)
}

func a_CyberAttack(actor *mobj_t) {
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	p_SpawnMissile(actor, actor.Ftarget, mt_ROCKET)
}

func a_BruisAttack(actor *mobj_t) {
	var damage int32
	if actor.Ftarget == nil {
		return
	}
	if p_CheckMeleeRange(actor) != 0 {
		s_StartSound(&actor.degenmobj_t, int32(sfx_claw))
		damage = (p_Random()%8 + 1) * 10
		p_DamageMobj(actor.Ftarget, actor, actor, damage)
		return
	}
	// launch a missile
	p_SpawnMissile(actor, actor.Ftarget, mt_BRUISERSHOT)
}

// C documentation
//
//	//
//	// A_SkelMissile
//	//
func a_SkelMissile(actor *mobj_t) {
	var mo *mobj_t
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	actor.Fz += 16 * (1 << FRACBITS) // so missile spawns higher
	mo = p_SpawnMissile(actor, actor.Ftarget, mt_TRACER)
	actor.Fz -= 16 * (1 << FRACBITS) // back to normal
	actor.Fx += mo.Fmomx
	actor.Fy += mo.Fmomy
	mo.Ftracer = actor.Ftarget
}

func init() {
	TRACEANGLE = 0xc000000
}

func a_Tracer(actor *mobj_t) {
	var dest *mobj_t
	var th *mobj_t
	var dist, slope fixed_t
	var exact angle_t
	if gametic&3 != 0 {
		return
	}
	// spawn a puff of smoke behind the rocket
	p_SpawnPuff(actor.Fx, actor.Fy, actor.Fz)
	th = p_SpawnMobj(actor.Fx-actor.Fmomx, actor.Fy-actor.Fmomy, actor.Fz, mt_SMOKE)
	th.Fmomz = 1 << FRACBITS
	th.Ftics -= p_Random() & 3
	if th.Ftics < 1 {
		th.Ftics = 1
	}
	// adjust direction
	dest = actor.Ftracer
	if dest == nil || dest.Fhealth <= 0 {
		return
	}
	// change angle
	exact = r_PointToAngle2(actor.Fx, actor.Fy, dest.Fx, dest.Fy)
	if exact != actor.Fangle {
		if exact-actor.Fangle > 0x80000000 {
			actor.Fangle -= TRACEANGLE
			if exact-actor.Fangle < 0x80000000 {
				actor.Fangle = exact
			}
		} else {
			actor.Fangle += TRACEANGLE
			if exact-actor.Fangle > 0x80000000 {
				actor.Fangle = exact
			}
		}
	}
	exact = actor.Fangle >> ANGLETOFINESHIFT
	actor.Fmomx = fixedMul(actor.Finfo.Fspeed, finecosine[exact])
	actor.Fmomy = fixedMul(actor.Finfo.Fspeed, finesine[exact])
	// change slope
	dist = p_AproxDistance(dest.Fx-actor.Fx, dest.Fy-actor.Fy)
	dist = dist / actor.Finfo.Fspeed
	if dist < 1 {
		dist = 1
	}
	slope = (dest.Fz + 40*(1<<FRACBITS) - actor.Fz) / dist
	if slope < actor.Fmomz {
		actor.Fmomz -= 1 << FRACBITS / 8
	} else {
		actor.Fmomz += 1 << FRACBITS / 8
	}
}

func a_SkelWhoosh(actor *mobj_t) {
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	s_StartSound(&actor.degenmobj_t, int32(sfx_skeswg))
}

func a_SkelFist(actor *mobj_t) {
	var damage int32
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	if p_CheckMeleeRange(actor) != 0 {
		damage = (p_Random()%int32(10) + 1) * 6
		s_StartSound(&actor.degenmobj_t, int32(sfx_skepch))
		p_DamageMobj(actor.Ftarget, actor, actor, damage)
	}
}

func pit_VileCheck(thing *mobj_t) boolean {
	var check boolean
	var maxdist int32
	var v1 fixed_t
	if thing.Fflags&mf_CORPSE == 0 {
		return 1
	} // not a monster
	if thing.Ftics != -1 {
		return 1
	} // not lying still yet
	if thing.Finfo.Fraisestate == s_NULL {
		return 1
	} // monster doesn't have a raise state
	maxdist = thing.Finfo.Fradius + mobjinfo[mt_VILE].Fradius
	if xabs(thing.Fx-viletryx) > maxdist || xabs(thing.Fy-viletryy) > maxdist {
		return 1
	} // not actually touching
	corpsehit = thing
	v1 = 0
	corpsehit.Fmomy = v1
	corpsehit.Fmomx = v1
	corpsehit.Fheight <<= 2
	check = p_CheckPosition(corpsehit, corpsehit.Fx, corpsehit.Fy)
	corpsehit.Fheight >>= 2
	if check == 0 {
		return 1
	} // doesn't fit here
	return 0 // got one, so stop checking
}

// C documentation
//
//	//
//	// A_VileChase
//	// Check for ressurecting a body
//	//
func a_VileChase(actor *mobj_t) {
	var bx, by, xh, xl, yh, yl int32
	var temp *mobj_t
	var info *mobjinfo_t
	if actor.Fmovedir != DI_NODIR {
		// check for corpses to raise
		viletryx = actor.Fx + actor.Finfo.Fspeed*xspeed[actor.Fmovedir]
		viletryy = actor.Fy + actor.Finfo.Fspeed*yspeed[actor.Fmovedir]
		xl = (viletryx - bmaporgx - 32*(1<<FRACBITS)*2) >> (FRACBITS + 7)
		xh = (viletryx - bmaporgx + 32*(1<<FRACBITS)*2) >> (FRACBITS + 7)
		yl = (viletryy - bmaporgy - 32*(1<<FRACBITS)*2) >> (FRACBITS + 7)
		yh = (viletryy - bmaporgy + 32*(1<<FRACBITS)*2) >> (FRACBITS + 7)
		bx = xl
		for {
			if !(bx <= xh) {
				break
			}
			by = yl
			for {
				if !(by <= yh) {
					break
				}
				// Call pit_VileCheck to check
				// whether object is a corpse
				// that canbe raised.
				if p_BlockThingsIterator(bx, by, pit_VileCheck) == 0 {
					// got one!
					temp = actor.Ftarget
					actor.Ftarget = corpsehit
					a_FaceTarget(actor)
					actor.Ftarget = temp
					p_SetMobjState(actor, s_VILE_HEAL1)
					s_StartSound(&corpsehit.degenmobj_t, int32(sfx_slop))
					info = corpsehit.Finfo
					p_SetMobjState(corpsehit, info.Fraisestate)
					corpsehit.Fheight <<= 2
					corpsehit.Fflags = info.Fflags
					corpsehit.Fhealth = info.Fspawnhealth
					corpsehit.Ftarget = nil
					return
				}
				goto _2
			_2:
				;
				by++
			}
			goto _1
		_1:
			;
			bx++
		}
	}
	// Return to normal attack.
	a_Chase(actor)
}

// C documentation
//
//	//
//	// A_VileStart
//	//
func a_VileStart(actor *mobj_t) {
	s_StartSound(&actor.degenmobj_t, int32(sfx_vilatk))
}

func a_StartFire(actor *mobj_t) {
	s_StartSound(&actor.degenmobj_t, int32(sfx_flamst))
	a_Fire(actor)
}

func a_FireCrackle(actor *mobj_t) {
	s_StartSound(&actor.degenmobj_t, int32(sfx_flame))
	a_Fire(actor)
}

func a_Fire(actor *mobj_t) {
	var an uint32
	var dest, target *mobj_t
	dest = actor.Ftracer
	if dest == nil {
		return
	}
	target = p_SubstNullMobj(actor.Ftarget)
	// don't move it if the vile lost sight
	if p_CheckSight(target, dest) == 0 {
		return
	}
	an = dest.Fangle >> ANGLETOFINESHIFT
	p_UnsetThingPosition(actor)
	actor.Fx = dest.Fx + fixedMul(24*(1<<FRACBITS), finecosine[an])
	actor.Fy = dest.Fy + fixedMul(24*(1<<FRACBITS), finesine[an])
	actor.Fz = dest.Fz
	p_SetThingPosition(actor)
}

// C documentation
//
//	//
//	// A_VileTarget
//	// Spawn the hellfire
//	//
func a_VileTarget(actor *mobj_t) {
	var fog *mobj_t
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	fog = p_SpawnMobj(actor.Ftarget.Fx, actor.Ftarget.Fx, actor.Ftarget.Fz, mt_FIRE)
	actor.Ftracer = fog
	fog.Ftarget = actor
	fog.Ftracer = actor.Ftarget
	a_Fire(fog)
}

// C documentation
//
//	//
//	// A_VileAttack
//	//
func a_VileAttack(actor *mobj_t) {
	var an int32
	var fire *mobj_t
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	if p_CheckSight(actor, actor.Ftarget) == 0 {
		return
	}
	s_StartSound(&actor.degenmobj_t, int32(sfx_barexp))
	p_DamageMobj(actor.Ftarget, actor, actor, 20)
	actor.Ftarget.Fmomz = 1000 * (1 << FRACBITS) / actor.Ftarget.Finfo.Fmass
	an = int32(actor.Fangle >> ANGLETOFINESHIFT)
	fire = actor.Ftracer
	if fire == nil {
		return
	}
	// move the fire between the vile and the player
	fire.Fx = actor.Ftarget.Fx - fixedMul(24*(1<<FRACBITS), finecosine[an])
	fire.Fy = actor.Ftarget.Fy - fixedMul(24*(1<<FRACBITS), finesine[an])
	p_RadiusAttack(fire, actor, 70)
}

//
// Mancubus attack,
// firing three missiles (bruisers)
// in three different directions?
// Doesn't look like it.
//

func a_FatRaise(actor *mobj_t) {
	a_FaceTarget(actor)
	s_StartSound(&actor.degenmobj_t, int32(sfx_manatk))
}

func a_FatAttack1(actor *mobj_t) {
	var an int32
	var mo, target *mobj_t
	a_FaceTarget(actor)
	// Change direction  to ...
	actor.Fangle += uint32(ANG903 / 8)
	target = p_SubstNullMobj(actor.Ftarget)
	p_SpawnMissile(actor, target, mt_FATSHOT)
	mo = p_SpawnMissile(actor, target, mt_FATSHOT)
	mo.Fangle += uint32(ANG903 / 8 * 2)
	an = int32(mo.Fangle >> ANGLETOFINESHIFT)
	mo.Fmomx = fixedMul(mo.Finfo.Fspeed, finecosine[an])
	mo.Fmomy = fixedMul(mo.Finfo.Fspeed, finesine[an])
}

func a_FatAttack2(actor *mobj_t) {
	var an int32
	var mo, target *mobj_t
	a_FaceTarget(actor)
	// Now here choose opposite deviation.
	actor.Fangle -= uint32(ANG903 / 8)
	target = p_SubstNullMobj(actor.Ftarget)
	p_SpawnMissile(actor, target, mt_FATSHOT)
	mo = p_SpawnMissile(actor, target, mt_FATSHOT)
	mo.Fangle -= uint32(ANG903 / 8 * 2)
	an = int32(mo.Fangle >> ANGLETOFINESHIFT)
	mo.Fmomx = fixedMul(mo.Finfo.Fspeed, finecosine[an])
	mo.Fmomy = fixedMul(mo.Finfo.Fspeed, finesine[an])
}

func a_FatAttack3(actor *mobj_t) {
	var an int32
	var mo, target *mobj_t
	a_FaceTarget(actor)
	target = p_SubstNullMobj(actor.Ftarget)
	mo = p_SpawnMissile(actor, target, mt_FATSHOT)
	mo.Fangle -= uint32(ANG903 / 8 / 2)
	an = int32(mo.Fangle >> ANGLETOFINESHIFT)
	mo.Fmomx = fixedMul(mo.Finfo.Fspeed, finecosine[an])
	mo.Fmomy = fixedMul(mo.Finfo.Fspeed, finesine[an])
	mo = p_SpawnMissile(actor, target, mt_FATSHOT)
	mo.Fangle += uint32(ANG903 / 8 / 2)
	an = int32(mo.Fangle >> ANGLETOFINESHIFT)
	mo.Fmomx = fixedMul(mo.Finfo.Fspeed, finecosine[an])
	mo.Fmomy = fixedMul(mo.Finfo.Fspeed, finesine[an])
}

//
// SkullAttack
// Fly at the player like a missile.
//

func a_SkullAttack(actor *mobj_t) {
	var an angle_t
	var dest *mobj_t
	var dist int32
	if actor.Ftarget == nil {
		return
	}
	dest = actor.Ftarget
	actor.Fflags |= mf_SKULLFLY
	s_StartSound(&actor.degenmobj_t, actor.Finfo.Fattacksound)
	a_FaceTarget(actor)
	an = actor.Fangle >> ANGLETOFINESHIFT
	actor.Fmomx = fixedMul(20*(1<<FRACBITS), finecosine[an])
	actor.Fmomy = fixedMul(20*(1<<FRACBITS), finesine[an])
	dist = p_AproxDistance(dest.Fx-actor.Fx, dest.Fy-actor.Fy)
	dist = dist / (20 * (1 << FRACBITS))
	if dist < 1 {
		dist = 1
	}
	actor.Fmomz = (dest.Fz + dest.Fheight>>1 - actor.Fz) / dist
}

// C documentation
//
//	//
//	// A_PainShootSkull
//	// Spawn a lost soul and launch it at the target
//	//
func a_PainShootSkull(actor *mobj_t, angle angle_t) {
	var an angle_t
	var count, prestep int32
	var currentthinker *thinker_t
	var newmobj *mobj_t
	var x, y, z fixed_t
	// count total number of skull currently on the level
	count = 0
	currentthinker = thinkercap.Fnext
	for currentthinker != &thinkercap {
		if mobj, ok := currentthinker.Ffunction.(*mobj_t); ok {
			if mobj.Ftype1 == mt_SKULL {
				count++
			}
		}
		currentthinker = currentthinker.Fnext
	}
	// if there are allready 20 skulls on the level,
	// don't spit another one
	if count > 20 {
		return
	}
	// okay, there's playe for another one
	an = angle >> ANGLETOFINESHIFT
	prestep = 4*(1<<FRACBITS) + 3*(actor.Finfo.Fradius+mobjinfo[mt_SKULL].Fradius)/2
	x = actor.Fx + fixedMul(prestep, finecosine[an])
	y = actor.Fy + fixedMul(prestep, finesine[an])
	z = actor.Fz + 8*(1<<FRACBITS)
	newmobj = p_SpawnMobj(x, y, z, mt_SKULL)
	// Check for movements.
	if p_TryMove(newmobj, newmobj.Fx, newmobj.Fy) == 0 {
		// kill it immediately
		p_DamageMobj(newmobj, actor, actor, 10000)
		return
	}
	newmobj.Ftarget = actor.Ftarget
	a_SkullAttack(newmobj)
}

// C documentation
//
//	//
//	// A_PainAttack
//	// Spawn a lost soul and launch it at the target
//	//
func a_PainAttack(actor *mobj_t) {
	if actor.Ftarget == nil {
		return
	}
	a_FaceTarget(actor)
	a_PainShootSkull(actor, actor.Fangle)
}

func a_PainDie(actor *mobj_t) {
	a_Fall(actor)
	a_PainShootSkull(actor, actor.Fangle+uint32(ANG903))
	a_PainShootSkull(actor, actor.Fangle+uint32(ANG1801))
	a_PainShootSkull(actor, actor.Fangle+uint32(ANG2703))
}

func a_Scream(actor *mobj_t) {
	var sound int32
	switch actor.Finfo.Fdeathsound {
	case 0:
		return
	case int32(sfx_podth1):
		fallthrough
	case int32(sfx_podth2):
		fallthrough
	case int32(sfx_podth3):
		sound = int32(sfx_podth1) + p_Random()%3
	case int32(sfx_bgdth1):
		fallthrough
	case int32(sfx_bgdth2):
		sound = int32(sfx_bgdth1) + p_Random()%2
	default:
		sound = actor.Finfo.Fdeathsound
		break
	}
	// Check for bosses.
	if actor.Ftype1 == mt_SPIDER || actor.Ftype1 == mt_CYBORG {
		// full volume
		s_StartSound(nil, sound)
	} else {
		s_StartSound(&actor.degenmobj_t, sound)
	}
}

func a_XScream(actor *mobj_t) {
	s_StartSound(&actor.degenmobj_t, int32(sfx_slop))
}

func a_Pain(actor *mobj_t) {
	if actor.Finfo.Fpainsound != 0 {
		s_StartSound(&actor.degenmobj_t, actor.Finfo.Fpainsound)
	}
}

func a_Fall(actor *mobj_t) {
	// actor is on ground, it can be walked over
	actor.Fflags &^= mf_SOLID
	// So change this if corpse objects
	// are meant to be obstacles.
}

// C documentation
//
//	//
//	// A_Explode
//	//
func a_Explode(thingy *mobj_t) {
	p_RadiusAttack(thingy, thingy.Ftarget, 128)
}

// Check whether the death of the specified monster type is allowed
// to trigger the end of episode special action.
//
// This behavior changed in v1.9, the most notable effect of which
// was to break uac_dead.wad

func checkBossEnd(motype mobjtype_t) boolean {
	if gameversion < exe_ultimate {
		if gamemap != 8 {
			return 0
		}
		// Baron death on later episodes is nothing special.
		if motype == mt_BRUISER && gameepisode != 1 {
			return 0
		}
		return 1
	} else {
		// New logic that appeared in Ultimate Doom.
		// Looks like the logic was overhauled while adding in the
		// episode 4 support.  Now bosses only trigger on their
		// specific episode.
		switch gameepisode {
		case 1:
			return booluint32(gamemap == 8 && motype == mt_BRUISER)
		case 2:
			return booluint32(gamemap == 8 && motype == mt_CYBORG)
		case 3:
			return booluint32(gamemap == 8 && motype == mt_SPIDER)
		case 4:
			return booluint32(gamemap == 6 && motype == mt_CYBORG || gamemap == 8 && motype == mt_SPIDER)
		default:
			return booluint32(gamemap == 8)
		}
	}
}

// C documentation
//
//	//
//	// A_BossDeath
//	// Possibly trigger special effects
//	// if on first boss level
//	//
func a_BossDeath(mo *mobj_t) {
	var i int32
	var th *thinker_t
	if gamemode == commercial {
		if gamemap != 7 {
			return
		}
		if mo.Ftype1 != mt_FATSO && mo.Ftype1 != mt_BABY {
			return
		}
	} else {
		if checkBossEnd(mo.Ftype1) == 0 {
			return
		}
	}
	// make sure there is a player alive for victory
	for i = int32(0); i < MAXPLAYERS; i++ {
		if playeringame[i] != 0 && players[i].Fhealth > 0 {
			break
		}
	}
	if i == MAXPLAYERS {
		return
	} // no one left alive, so do not end game
	// scan the remaining thinkers to see
	// if all bosses are dead
	th = thinkercap.Fnext
	for {
		if th == &thinkercap {
			break
		}
		mo2, ok := th.Ffunction.(*mobj_t)
		if !ok {
			goto _2
		}
		if mo2 != mo && mo2.Ftype1 == mo.Ftype1 && mo2.Fhealth > 0 {
			// other boss not dead
			return
		}
		goto _2
	_2:
		;
		th = th.Fnext
	}
	// victory!
	if gamemode == commercial {
		if gamemap == 7 {
			if mo.Ftype1 == mt_FATSO {
				ev_DoFloor(&line_t{Ftag: 666}, int32(lowerFloorToLowest))
				return
			}
			if mo.Ftype1 == mt_BABY {
				ev_DoFloor(&line_t{Ftag: 667}, int32(raiseToTexture))
				return
			}
		}
	} else {
		switch gameepisode {
		case 1:
			ev_DoFloor(&line_t{Ftag: 666}, int32(lowerFloorToLowest))
			return
		case 4:
			switch gamemap {
			case 6:
				ev_DoDoor(&line_t{Ftag: 666}, int32(vld_blazeOpen))
				return
			case 8:
				ev_DoFloor(&line_t{Ftag: 666}, int32(lowerFloorToLowest))
				return
				break
			}
		}
	}
	g_ExitLevel()
}

func a_Hoof(mo *mobj_t) {
	s_StartSound(&mo.degenmobj_t, int32(sfx_hoof))
	a_Chase(mo)
}

func a_Metal(mo *mobj_t) {
	s_StartSound(&mo.degenmobj_t, int32(sfx_metal))
	a_Chase(mo)
}

func a_BabyMetal(mo *mobj_t) {
	s_StartSound(&mo.degenmobj_t, int32(sfx_bspwlk))
	a_Chase(mo)
}

func a_OpenShotgun2(player *player_t, psp *pspdef_t) {
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_dbopn))
}

func a_LoadShotgun2(player *player_t, psp *pspdef_t) {
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_dbload))
}

func a_CloseShotgun2(player *player_t, psp *pspdef_t) {
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_dbcls))
	a_ReFire(player, psp)
}

func a_BrainAwake(mo *mobj_t) {
	var thinker *thinker_t
	// find all the target spots
	numbraintargets = 0
	braintargeton = 0
	thinker = thinkercap.Fnext
	thinker = thinkercap.Fnext
	for {
		if thinker == &thinkercap {
			break
		}
		m, ok := thinker.Ffunction.(*mobj_t)
		if !ok {
			goto _1
		} // not a mobj
		if m.Ftype1 == mt_BOSSTARGET {
			braintargets[numbraintargets] = m
			numbraintargets++
		}
		goto _1
	_1:
		;
		thinker = thinker.Fnext
	}
	s_StartSound(nil, int32(sfx_bossit))
}

func a_BrainPain(mo *mobj_t) {
	s_StartSound(nil, int32(sfx_bospn))
}

func a_BrainScream(mo *mobj_t) {
	var th *mobj_t
	var x, y, z int32
	x = mo.Fx - 196*(1<<FRACBITS)
	for {
		if x >= mo.Fx+320*(1<<FRACBITS) {
			break
		}
		y = mo.Fy - 320*(1<<FRACBITS)
		z = 128 + p_Random()*2*(1<<FRACBITS)
		th = p_SpawnMobj(x, y, z, mt_ROCKET)
		th.Fmomz = p_Random() * 512
		p_SetMobjState(th, s_BRAINEXPLODE1)
		th.Ftics -= p_Random() & 7
		if th.Ftics < 1 {
			th.Ftics = 1
		}
		goto _1
	_1:
		;
		x += 1 << FRACBITS * 8
	}
	s_StartSound(nil, int32(sfx_bosdth))
}

func a_BrainExplode(mo *mobj_t) {
	var th *mobj_t
	var x, y, z int32
	x = mo.Fx + (p_Random()-p_Random())*int32(2048)
	y = mo.Fy
	z = 128 + p_Random()*2*(1<<FRACBITS)
	th = p_SpawnMobj(x, y, z, mt_ROCKET)
	th.Fmomz = p_Random() * 512
	p_SetMobjState(th, s_BRAINEXPLODE1)
	th.Ftics -= p_Random() & 7
	if th.Ftics < 1 {
		th.Ftics = 1
	}
}

func a_BrainDie(mo *mobj_t) {
	g_ExitLevel()
}

func a_BrainSpit(mo *mobj_t) {
	var newmobj, targ *mobj_t
	easy ^= 1
	if gameskill <= sk_easy && easy == 0 {
		return
	}
	// shoot a cube at current target
	targ = braintargets[braintargeton]
	braintargeton = (braintargeton + 1) % numbraintargets
	// spawn brain missile
	newmobj = p_SpawnMissile(mo, targ, mt_SPAWNSHOT)
	newmobj.Ftarget = targ
	newmobj.Freactiontime = (targ.Fy - mo.Fy) / newmobj.Fmomy / newmobj.Fstate.Ftics
	s_StartSound(nil, int32(sfx_bospit))
}

var easy int32

// C documentation
//
//	// travelling cube sound
func a_SpawnSound(mo *mobj_t) {
	s_StartSound(&mo.degenmobj_t, int32(sfx_boscub))
	a_SpawnFly(mo)
}

func a_SpawnFly(mo *mobj_t) {
	var fog, newmobj, targ *mobj_t
	var r int32
	var type1 mobjtype_t
	mo.Freactiontime--
	if mo.Freactiontime != 0 {
		return
	} // still flying
	targ = p_SubstNullMobj(mo.Ftarget)
	// First spawn teleport fog.
	fog = p_SpawnMobj(targ.Fx, targ.Fy, targ.Fz, mt_SPAWNFIRE)
	s_StartSound(&fog.degenmobj_t, int32(sfx_telept))
	// Randomly select monster to spawn.
	r = p_Random()
	// Probability distribution (kind of :),
	// decreasing likelihood.
	if r < 50 {
		type1 = mt_TROOP
	} else {
		if r < 90 {
			type1 = mt_SERGEANT
		} else {
			if r < 120 {
				type1 = mt_SHADOWS
			} else {
				if r < 130 {
					type1 = mt_PAIN
				} else {
					if r < 160 {
						type1 = mt_HEAD
					} else {
						if r < 162 {
							type1 = mt_VILE
						} else {
							if r < 172 {
								type1 = mt_UNDEAD
							} else {
								if r < 192 {
									type1 = mt_BABY
								} else {
									if r < 222 {
										type1 = mt_FATSO
									} else {
										if r < 246 {
											type1 = mt_KNIGHT
										} else {
											type1 = mt_BRUISER
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	newmobj = p_SpawnMobj(targ.Fx, targ.Fy, targ.Fz, type1)
	if p_LookForPlayers(newmobj, 1) != 0 {
		p_SetMobjState(newmobj, newmobj.Finfo.Fseestate)
	}
	// telefrag anything in this spot
	p_TeleportMove(newmobj, newmobj.Fx, newmobj.Fy)
	// remove self (i.e., cube).
	p_RemoveMobj(mo)
}

func a_PlayerScream(mo *mobj_t) {
	var sound int32
	// Default death sound.
	sound = int32(sfx_pldeth)
	if gamemode == commercial && mo.Fhealth < -int32(50) {
		// IF THE PLAYER DIES
		// LESS THAN -50% WITHOUT GIBBING
		sound = int32(sfx_pdiehi)
	}
	s_StartSound(&mo.degenmobj_t, sound)
}

const INT_MAX9 = 2147483647

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// Data.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

//
// FLOORS
//

// C documentation
//
//	//
//	// Move a plane (floor or ceiling) and check for crushing
//	//
func t_MovePlane(sector *sector_t, speed fixed_t, dest fixed_t, crush boolean, floorOrCeiling int32, direction int32) result_e {
	var flag boolean
	var lastpos fixed_t
	switch floorOrCeiling {
	case 0:
		// FLOOR
		switch direction {
		case -1:
			// DOWN
			if sector.Ffloorheight-speed < dest {
				lastpos = sector.Ffloorheight
				sector.Ffloorheight = dest
				flag = p_ChangeSector(sector, crush)
				if flag == 1 {
					sector.Ffloorheight = lastpos
					p_ChangeSector(sector, crush)
					//return crushed;
				}
				return int32(pastdest)
			} else {
				lastpos = sector.Ffloorheight
				sector.Ffloorheight -= speed
				flag = p_ChangeSector(sector, crush)
				if flag == 1 {
					sector.Ffloorheight = lastpos
					p_ChangeSector(sector, crush)
					return int32(crushed)
				}
			}
		case 1:
			// UP
			if sector.Ffloorheight+speed > dest {
				lastpos = sector.Ffloorheight
				sector.Ffloorheight = dest
				flag = p_ChangeSector(sector, crush)
				if flag == 1 {
					sector.Ffloorheight = lastpos
					p_ChangeSector(sector, crush)
					//return crushed;
				}
				return int32(pastdest)
			} else {
				// COULD GET CRUSHED
				lastpos = sector.Ffloorheight
				sector.Ffloorheight += speed
				flag = p_ChangeSector(sector, crush)
				if flag == 1 {
					if crush == 1 {
						return int32(crushed)
					}
					sector.Ffloorheight = lastpos
					p_ChangeSector(sector, crush)
					return int32(crushed)
				}
			}
			break
		}
	case 1:
		// CEILING
		switch direction {
		case -1:
			// DOWN
			if sector.Fceilingheight-speed < dest {
				lastpos = sector.Fceilingheight
				sector.Fceilingheight = dest
				flag = p_ChangeSector(sector, crush)
				if flag == 1 {
					sector.Fceilingheight = lastpos
					p_ChangeSector(sector, crush)
					//return crushed;
				}
				return int32(pastdest)
			} else {
				// COULD GET CRUSHED
				lastpos = sector.Fceilingheight
				sector.Fceilingheight -= speed
				flag = p_ChangeSector(sector, crush)
				if flag == 1 {
					if crush == 1 {
						return int32(crushed)
					}
					sector.Fceilingheight = lastpos
					p_ChangeSector(sector, crush)
					return int32(crushed)
				}
			}
		case 1:
			// UP
			if sector.Fceilingheight+speed > dest {
				lastpos = sector.Fceilingheight
				sector.Fceilingheight = dest
				flag = p_ChangeSector(sector, crush)
				if flag == 1 {
					sector.Fceilingheight = lastpos
					p_ChangeSector(sector, crush)
					//return crushed;
				}
				return int32(pastdest)
			} else {
				lastpos = sector.Fceilingheight
				sector.Fceilingheight += speed
				flag = p_ChangeSector(sector, crush)
				// UNUSED
			}
			break
		}
		break
	}
	return int32(ok)
}

// C documentation
//
//	//
//	// MOVE A FLOOR TO IT'S DESTINATION (UP OR DOWN)
//	//
func (floor *floormove_t) ThinkerFunc() {
	t_MoveFloor(floor)
}
func t_MoveFloor(floor *floormove_t) {
	var res result_e
	res = t_MovePlane(floor.Fsector, floor.Fspeed, floor.Ffloordestheight, floor.Fcrush, 0, floor.Fdirection)
	if leveltime&7 == 0 {
		s_StartSound(&floor.Fsector.Fsoundorg, int32(sfx_stnmov))
	}
	if res == int32(pastdest) {
		floor.Fsector.Fspecialdata = nil
		if floor.Fdirection == 1 {
			switch floor.Ftype1 {
			case int32(donutRaise):
				floor.Fsector.Fspecial = int16(floor.Fnewspecial)
				floor.Fsector.Ffloorpic = floor.Ftexture
				fallthrough
			default:
				break
			}
		} else {
			if floor.Fdirection == -1 {
				switch floor.Ftype1 {
				case int32(lowerAndChange):
					floor.Fsector.Fspecial = int16(floor.Fnewspecial)
					floor.Fsector.Ffloorpic = floor.Ftexture
					fallthrough
				default:
					break
				}
			}
		}
		p_RemoveThinker(&floor.Fthinker)
		s_StartSound(&floor.Fsector.Fsoundorg, int32(sfx_pstop))
	}
}

// C documentation
//
//	//
//	// HANDLE FLOOR TYPES
//	//
func ev_DoFloor(line *line_t, floortype floor_e) int32 {
	var side *side_t
	var sec *sector_t
	var minsize, rtn, secnum, v1 int32
	secnum = -1
	rtn = 0
	for {
		v1 = p_FindSectorFromLineTag(line, secnum)
		secnum = v1
		if !(v1 >= 0) {
			break
		}
		sec = &sectors[secnum]
		// ALREADY MOVING?  IF SO, KEEP GOING...
		if sec.Fspecialdata != nil {
			continue
		}
		// new floor thinker
		rtn = 1
		floorP := &floormove_t{}
		p_AddThinker(&floorP.Fthinker)
		sec.Fspecialdata = floorP
		floorP.Fthinker.Ffunction = floorP
		floorP.Ftype1 = floortype
		floorP.Fcrush = 0
		switch floortype {
		case int32(lowerFloor):
			floorP.Fdirection = -1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			floorP.Ffloordestheight = p_FindHighestFloorSurrounding(sec)
		case int32(lowerFloorToLowest):
			floorP.Fdirection = -1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			floorP.Ffloordestheight = p_FindLowestFloorSurrounding(sec)
		case int32(turboLower):
			floorP.Fdirection = -1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS * 4
			floorP.Ffloordestheight = p_FindHighestFloorSurrounding(sec)
			if floorP.Ffloordestheight != sec.Ffloorheight {
				floorP.Ffloordestheight += 8 * (1 << FRACBITS)
			}
		case int32(raiseFloorCrush):
			floorP.Fcrush = 1
			fallthrough
		case int32(raiseFloor):
			floorP.Fdirection = 1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			floorP.Ffloordestheight = p_FindLowestCeilingSurrounding(sec)
			if floorP.Ffloordestheight > sec.Fceilingheight {
				floorP.Ffloordestheight = sec.Fceilingheight
			}
			floorP.Ffloordestheight -= 8 * (1 << FRACBITS) * boolint32(floortype == int32(raiseFloorCrush))
		case int32(raiseFloorTurbo):
			floorP.Fdirection = 1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS * 4
			floorP.Ffloordestheight = p_FindNextHighestFloor(sec, sec.Ffloorheight)
		case int32(raiseFloorToNearest):
			floorP.Fdirection = 1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			floorP.Ffloordestheight = p_FindNextHighestFloor(sec, sec.Ffloorheight)
		case int32(raiseFloor24):
			floorP.Fdirection = 1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			floorP.Ffloordestheight = floorP.Fsector.Ffloorheight + 24*(1<<FRACBITS)
		case int32(raiseFloor512):
			floorP.Fdirection = 1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			floorP.Ffloordestheight = floorP.Fsector.Ffloorheight + 512*(1<<FRACBITS)
		case int32(raiseFloor24AndChange):
			floorP.Fdirection = 1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			floorP.Ffloordestheight = floorP.Fsector.Ffloorheight + 24*(1<<FRACBITS)
			sec.Ffloorpic = line.Ffrontsector.Ffloorpic
			sec.Fspecial = line.Ffrontsector.Fspecial
		case int32(raiseToTexture):
			minsize = int32(INT_MAX9)
			floorP.Fdirection = 1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			for i := int32(0); i < sec.Flinecount; i++ {
				if twoSided(secnum, i) != 0 {
					side = getSide(secnum, i, 0)
					if int32(side.Fbottomtexture) >= 0 {
						if textureheight[side.Fbottomtexture] < minsize {
							minsize = textureheight[side.Fbottomtexture]
						}
					}
					side = getSide(secnum, i, 1)
					if int32(side.Fbottomtexture) >= 0 {
						if textureheight[side.Fbottomtexture] < minsize {
							minsize = textureheight[side.Fbottomtexture]
						}
					}
				}
			}
			floorP.Ffloordestheight = floorP.Fsector.Ffloorheight + minsize
		case int32(lowerAndChange):
			floorP.Fdirection = -1
			floorP.Fsector = sec
			floorP.Fspeed = 1 << FRACBITS
			floorP.Ffloordestheight = p_FindLowestFloorSurrounding(sec)
			floorP.Ftexture = sec.Ffloorpic
			for i := int32(0); i < sec.Flinecount; i++ {
				if twoSided(secnum, i) != 0 {
					if sectorIndex(getSide(secnum, i, 0).Fsector) == secnum {
						sec = getSector(secnum, i, 1)
						if sec.Ffloorheight == floorP.Ffloordestheight {
							floorP.Ftexture = sec.Ffloorpic
							floorP.Fnewspecial = int32(sec.Fspecial)
							break
						}
					} else {
						sec = getSector(secnum, i, 0)
						if sec.Ffloorheight == floorP.Ffloordestheight {
							floorP.Ftexture = sec.Ffloorpic
							floorP.Fnewspecial = int32(sec.Fspecial)
							break
						}
					}
				}
			}
			fallthrough
		default:
			break
		}
	}
	return rtn
}

// C documentation
//
//	//
//	// BUILD A STAIRCASE!
//	//
func ev_BuildStairs(line *line_t, type1 stair_e) int32 {
	var sec, tsec *sector_t
	var height, newsecnum, ok, rtn, secnum, texture, v1 int32
	var speed, stairsize fixed_t
	stairsize = 0
	speed = 0
	secnum = -1
	rtn = 0
	for {
		v1 = p_FindSectorFromLineTag(line, secnum)
		secnum = v1
		if !(v1 >= 0) {
			break
		}
		sec = &sectors[secnum]
		// ALREADY MOVING?  IF SO, KEEP GOING...
		if sec.Fspecialdata != nil {
			continue
		}
		// new floor thinker
		rtn = 1
		floorP := &floormove_t{}
		p_AddThinker(&floorP.Fthinker)
		sec.Fspecialdata = floorP
		floorP.Fthinker.Ffunction = floorP
		floorP.Fdirection = 1
		floorP.Fsector = sec
		switch type1 {
		case int32(build8):
			speed = 1 << FRACBITS / 4
			stairsize = 8 * (1 << FRACBITS)
		case int32(turbo16):
			speed = 1 << FRACBITS * 4
			stairsize = 16 * (1 << FRACBITS)
			break
		}
		floorP.Fspeed = speed
		height = sec.Ffloorheight + stairsize
		floorP.Ffloordestheight = height
		texture = int32(sec.Ffloorpic)
		// Find next sector to raise
		// 1.	Find 2-sided line with same sector side[0]
		// 2.	Other side is the next sector to raise
		for cond := true; cond; cond = ok != 0 {
			ok = 0
			for i := int32(0); i < sec.Flinecount; i++ {
				if int32(sec.Flines[i].Fflags)&ml_TWOSIDED == 0 {
					continue
				}
				tsec = sec.Flines[i].Ffrontsector
				newsecnum = sectorIndex(tsec)
				if secnum != newsecnum {
					continue
				}
				tsec = sec.Flines[i].Fbacksector
				newsecnum = sectorIndex(tsec)
				if int32(tsec.Ffloorpic) != texture {
					continue
				}
				height += stairsize
				if tsec.Fspecialdata != 0 {
					continue
				}
				sec = tsec
				secnum = newsecnum
				floorP = &floormove_t{}
				p_AddThinker(&floorP.Fthinker)
				sec.Fspecialdata = floorP
				floorP.Fthinker.Ffunction = floorP
				floorP.Fdirection = 1
				floorP.Fsector = sec
				floorP.Fspeed = speed
				floorP.Ffloordestheight = height
				ok = 1
				break
			}
		}
	}
	return rtn
}

const ANG1803 = 2147483648
const BONUSADD = 6

func init() {
	maxammo = [4]int32{
		0: 200,
		1: 50,
		2: 300,
		3: 50,
	}
}

func init() {
	clipammo = [4]int32{
		0: 10,
		1: 4,
		2: 20,
		3: 1,
	}
}

//
// GET STUFF
//

//
// P_GiveAmmo
// Num is the number of clip loads,
// not the individual count (0= 1/2 clip).
// Returns false if the ammo can't be picked up at all
//

func p_GiveAmmo(player *player_t, ammo ammotype_t, num int32) boolean {
	var oldammo int32
	if ammo == am_noammo {
		return 0
	}
	if ammo > NUMAMMO {
		i_Error("p_GiveAmmo: bad type %d", ammo)
	}
	if player.Fammo[ammo] == player.Fmaxammo[ammo] {
		return 0
	}
	if num != 0 {
		num *= clipammo[ammo]
	} else {
		num = clipammo[ammo] / 2
	}
	if gameskill == sk_baby || gameskill == sk_nightmare {
		// give double ammo in trainer mode,
		// you'll need in nightmare
		num <<= 1
	}
	oldammo = player.Fammo[ammo]
	player.Fammo[ammo] += num
	if player.Fammo[ammo] > player.Fmaxammo[ammo] {
		player.Fammo[ammo] = player.Fmaxammo[ammo]
	}
	// If non zero ammo,
	// don't change up weapons,
	// player was lower on purpose.
	if oldammo != 0 {
		return 1
	}
	// We were down to zero,
	// so select a new weapon.
	// Preferences are not user selectable.
	switch ammo {
	case am_clip:
		if player.Freadyweapon == wp_fist {
			if player.Fweaponowned[wp_chainsaw] != 0 {
				player.Fpendingweapon = wp_chaingun
			} else {
				player.Fpendingweapon = wp_pistol
			}
		}
	case am_shell:
		if player.Freadyweapon == wp_fist || player.Freadyweapon == wp_pistol {
			if player.Fweaponowned[wp_shotgun] != 0 {
				player.Fpendingweapon = wp_shotgun
			}
		}
	case am_cell:
		if player.Freadyweapon == wp_fist || player.Freadyweapon == wp_pistol {
			if player.Fweaponowned[wp_plasma] != 0 {
				player.Fpendingweapon = wp_plasma
			}
		}
	case am_misl:
		if player.Freadyweapon == wp_fist {
			if player.Fweaponowned[wp_missile] != 0 {
				player.Fpendingweapon = wp_missile
			}
		}
		fallthrough
	default:
		break
	}
	return 1
}

// C documentation
//
//	//
//	// P_GiveWeapon
//	// The weapon name may have a mf_DROPPED flag ored in.
//	//
func p_GiveWeapon(player *player_t, weapon weapontype_t, dropped boolean) boolean {
	var gaveammo, gaveweapon boolean
	if netgame != 0 && deathmatch != 2 && dropped == 0 {
		// leave placed weapons forever on net games
		if player.Fweaponowned[weapon] != 0 {
			return 0
		}
		player.Fbonuscount += BONUSADD
		player.Fweaponowned[weapon] = 1
		if deathmatch != 0 {
			p_GiveAmmo(player, weaponinfo[weapon].Fammo, 5)
		} else {
			p_GiveAmmo(player, weaponinfo[weapon].Fammo, 2)
		}
		player.Fpendingweapon = weapon
		if player == &players[consoleplayer] {
			s_StartSound(nil, int32(sfx_wpnup))
		}
		return 0
	}
	if weaponinfo[weapon].Fammo != am_noammo {
		// give one clip with a dropped weapon,
		// two clips with a found weapon
		if dropped != 0 {
			gaveammo = p_GiveAmmo(player, weaponinfo[weapon].Fammo, 1)
		} else {
			gaveammo = p_GiveAmmo(player, weaponinfo[weapon].Fammo, 2)
		}
	} else {
		gaveammo = 0
	}
	if player.Fweaponowned[weapon] != 0 {
		gaveweapon = 0
	} else {
		gaveweapon = 1
		player.Fweaponowned[weapon] = 1
		player.Fpendingweapon = weapon
	}
	return booluint32(gaveweapon != 0 || gaveammo != 0)
}

// C documentation
//
//	//
//	// P_GiveBody
//	// Returns false if the body isn't needed at all
//	//
func p_GiveBody(player *player_t, num int32) boolean {
	if player.Fhealth >= MAXHEALTH {
		return 0
	}
	player.Fhealth += num
	if player.Fhealth > MAXHEALTH {
		player.Fhealth = MAXHEALTH
	}
	player.Fmo.Fhealth = player.Fhealth
	return 1
}

// C documentation
//
//	//
//	// P_GiveArmor
//	// Returns false if the armor is worse
//	// than the current armor.
//	//
func p_GiveArmor(player *player_t, armortype int32) boolean {
	var hits int32
	hits = armortype * 100
	if player.Farmorpoints >= hits {
		return 0
	} // don't pick up
	player.Farmortype = armortype
	player.Farmorpoints = hits
	return 1
}

// C documentation
//
//	//
//	// P_GiveCard
//	//
func p_GiveCard(player *player_t, card card_t) {
	if player.Fcards[card] != 0 {
		return
	}
	player.Fbonuscount = BONUSADD
	player.Fcards[card] = 1
}

// C documentation
//
//	//
//	// P_GivePower
//	//
func p_GivePower(player *player_t, power int32) boolean {
	if power == int32(pw_invulnerability) {
		player.Fpowers[power] = INVULNTICS
		return 1
	}
	if power == int32(pw_invisibility) {
		player.Fpowers[power] = INVISTICS
		player.Fmo.Fflags |= mf_SHADOW
		return 1
	}
	if power == int32(pw_infrared) {
		player.Fpowers[power] = INFRATICS
		return 1
	}
	if power == int32(pw_ironfeet) {
		player.Fpowers[power] = IRONTICS
		return 1
	}
	if power == int32(pw_strength) {
		p_GiveBody(player, 100)
		player.Fpowers[power] = 1
		return 1
	}
	if player.Fpowers[power] != 0 {
		return 0
	} // already got it
	player.Fpowers[power] = 1
	return 1
}

// C documentation
//
//	//
//	// P_TouchSpecialThing
//	//
func p_TouchSpecialThing(special *mobj_t, toucher *mobj_t) {
	var delta fixed_t
	var sound int32
	var player *player_t
	delta = special.Fz - toucher.Fz
	if delta > toucher.Fheight || delta < -8*(1<<FRACBITS) {
		// out of reach
		return
	}
	sound = int32(sfx_itemup)
	player = toucher.Fplayer
	// Dead thing touching.
	// Can happen with a sliding player corpse.
	if toucher.Fhealth <= 0 {
		return
	}
	// Identify by sprite.
	switch special.Fsprite {
	// armor
	case spr_ARM1:
		if p_GiveArmor(player, DEH_DEFAULT_GREEN_ARMOR_CLASS) == 0 {
			return
		}
		player.Fmessage = "Picked up the armor."
	case spr_ARM2:
		if p_GiveArmor(player, DEH_DEFAULT_BLUE_ARMOR_CLASS) == 0 {
			return
		}
		player.Fmessage = "Picked up the MegaArmor!"
		break
		// bonus items
		fallthrough
	case spr_BON1:
		player.Fhealth++ // can go over 100%
		if player.Fhealth > DEH_DEFAULT_MAX_HEALTH {
			player.Fhealth = DEH_DEFAULT_MAX_HEALTH
		}
		player.Fmo.Fhealth = player.Fhealth
		player.Fmessage = "Picked up a health bonus."
	case spr_BON2:
		player.Farmorpoints++ // can go over 100%
		if player.Farmorpoints > DEH_DEFAULT_MAX_ARMOR {
			player.Farmorpoints = DEH_DEFAULT_MAX_ARMOR
		}
		// deh_green_armor_class only applies to the green armor shirt;
		// for the armor helmets, armortype 1 is always used.
		if player.Farmortype == 0 {
			player.Farmortype = 1
		}
		player.Fmessage = "Picked up an armor bonus."
	case spr_SOUL:
		player.Fhealth += DEH_DEFAULT_SOULSPHERE_HEALTH
		if player.Fhealth > DEH_DEFAULT_MAX_SOULSPHERE {
			player.Fhealth = DEH_DEFAULT_MAX_SOULSPHERE
		}
		player.Fmo.Fhealth = player.Fhealth
		player.Fmessage = "Supercharge!"
		sound = int32(sfx_getpow)
	case spr_MEGA:
		if gamemode != commercial {
			return
		}
		player.Fhealth = DEH_DEFAULT_MEGASPHERE_HEALTH
		player.Fmo.Fhealth = player.Fhealth
		// We always give armor type 2 for the megasphere; dehacked only
		// affects the MegaArmor.
		p_GiveArmor(player, 2)
		player.Fmessage = "MegaSphere!"
		sound = int32(sfx_getpow)
		break
		// cards
		// leave cards for everyone
		fallthrough
	case spr_BKEY:
		if player.Fcards[it_bluecard] == 0 {
			player.Fmessage = "Picked up a blue keycard."
		}
		p_GiveCard(player, it_bluecard)
		if netgame == 0 {
			break
		}
		return
	case spr_YKEY:
		if player.Fcards[it_yellowcard] == 0 {
			player.Fmessage = "Picked up a yellow keycard."
		}
		p_GiveCard(player, it_yellowcard)
		if netgame == 0 {
			break
		}
		return
	case spr_RKEY:
		if player.Fcards[it_redcard] == 0 {
			player.Fmessage = "Picked up a red keycard."
		}
		p_GiveCard(player, it_redcard)
		if netgame == 0 {
			break
		}
		return
	case spr_BSKU:
		if player.Fcards[it_blueskull] == 0 {
			player.Fmessage = "Picked up a blue skull key."
		}
		p_GiveCard(player, it_blueskull)
		if netgame == 0 {
			break
		}
		return
	case spr_YSKU:
		if player.Fcards[it_yellowskull] == 0 {
			player.Fmessage = "Picked up a yellow skull key."
		}
		p_GiveCard(player, it_yellowskull)
		if netgame == 0 {
			break
		}
		return
	case spr_RSKU:
		if player.Fcards[it_redskull] == 0 {
			player.Fmessage = "Picked up a red skull key."
		}
		p_GiveCard(player, it_redskull)
		if netgame == 0 {
			break
		}
		return
		// medikits, heals
		fallthrough
	case spr_STIM:
		if p_GiveBody(player, 10) == 0 {
			return
		}
		player.Fmessage = "Picked up a stimpack."
	case spr_MEDI:
		if p_GiveBody(player, 25) == 0 {
			return
		}
		if player.Fhealth < 25 {
			player.Fmessage = "Picked up a medikit that you REALLY need!"
		} else {
			player.Fmessage = "Picked up a medikit."
		}
		break
		// power ups
		fallthrough
	case spr_PINV:
		if p_GivePower(player, int32(pw_invulnerability)) == 0 {
			return
		}
		player.Fmessage = "Invulnerability!"
		sound = int32(sfx_getpow)
	case spr_PSTR:
		if p_GivePower(player, int32(pw_strength)) == 0 {
			return
		}
		player.Fmessage = "Berserk!"
		if player.Freadyweapon != wp_fist {
			player.Fpendingweapon = wp_fist
		}
		sound = int32(sfx_getpow)
	case spr_PINS:
		if p_GivePower(player, int32(pw_invisibility)) == 0 {
			return
		}
		player.Fmessage = "Partial Invisibility"
		sound = int32(sfx_getpow)
	case spr_SUIT:
		if p_GivePower(player, int32(pw_ironfeet)) == 0 {
			return
		}
		player.Fmessage = "Radiation Shielding Suit"
		sound = int32(sfx_getpow)
	case spr_PMAP:
		if p_GivePower(player, int32(pw_allmap)) == 0 {
			return
		}
		player.Fmessage = "Computer Area Map"
		sound = int32(sfx_getpow)
	case spr_PVIS:
		if p_GivePower(player, int32(pw_infrared)) == 0 {
			return
		}
		player.Fmessage = "Light Amplification Visor"
		sound = int32(sfx_getpow)
		break
		// ammo
		fallthrough
	case spr_CLIP:
		if special.Fflags&mf_DROPPED != 0 {
			if p_GiveAmmo(player, am_clip, 0) == 0 {
				return
			}
		} else {
			if p_GiveAmmo(player, am_clip, 1) == 0 {
				return
			}
		}
		player.Fmessage = "Picked up a clip."
	case spr_AMMO:
		if p_GiveAmmo(player, am_clip, 5) == 0 {
			return
		}
		player.Fmessage = "Picked up a box of bullets."
	case spr_ROCK:
		if p_GiveAmmo(player, am_misl, 1) == 0 {
			return
		}
		player.Fmessage = "Picked up a rocket."
	case spr_BROK:
		if p_GiveAmmo(player, am_misl, 5) == 0 {
			return
		}
		player.Fmessage = "Picked up a box of rockets."
	case spr_CELL:
		if p_GiveAmmo(player, am_cell, 1) == 0 {
			return
		}
		player.Fmessage = "Picked up an energy cell."
	case spr_CELP:
		if p_GiveAmmo(player, am_cell, 5) == 0 {
			return
		}
		player.Fmessage = "Picked up an energy cell pack."
	case spr_SHEL:
		if p_GiveAmmo(player, am_shell, 1) == 0 {
			return
		}
		player.Fmessage = "Picked up 4 shotgun shells."
	case spr_SBOX:
		if p_GiveAmmo(player, am_shell, 5) == 0 {
			return
		}
		player.Fmessage = "Picked up a box of shotgun shells."
	case spr_BPAK:
		if player.Fbackpack == 0 {
			for i := int32(0); i < NUMAMMO; i++ {
				player.Fmaxammo[i] *= 2
			}
			player.Fbackpack = 1
		}
		for i := int32(0); i < NUMAMMO; i++ {
			p_GiveAmmo(player, ammotype_t(i), 1)
		}
		player.Fmessage = "Picked up a backpack full of ammo!"
		break
		// weapons
		fallthrough
	case spr_BFUG:
		if p_GiveWeapon(player, wp_bfg, 0) == 0 {
			return
		}
		player.Fmessage = "You got the BFG9000!  Oh, yes."
		sound = int32(sfx_wpnup)
	case spr_MGUN:
		if p_GiveWeapon(player, wp_chaingun, booluint32(special.Fflags&mf_DROPPED != 0)) == 0 {
			return
		}
		player.Fmessage = "You got the chaingun!"
		sound = int32(sfx_wpnup)
	case spr_CSAW:
		if p_GiveWeapon(player, wp_chainsaw, 0) == 0 {
			return
		}
		player.Fmessage = "A chainsaw!  Find some meat!"
		sound = int32(sfx_wpnup)
	case spr_LAUN:
		if p_GiveWeapon(player, wp_missile, 0) == 0 {
			return
		}
		player.Fmessage = "You got the rocket launcher!"
		sound = int32(sfx_wpnup)
	case spr_PLAS:
		if p_GiveWeapon(player, wp_plasma, 0) == 0 {
			return
		}
		player.Fmessage = "You got the plasma gun!"
		sound = int32(sfx_wpnup)
	case spr_SHOT:
		if p_GiveWeapon(player, wp_shotgun, booluint32(special.Fflags&mf_DROPPED != 0)) == 0 {
			return
		}
		player.Fmessage = "You got the shotgun!"
		sound = int32(sfx_wpnup)
	case spr_SGN2:
		if p_GiveWeapon(player, wp_supershotgun, booluint32(special.Fflags&mf_DROPPED != 0)) == 0 {
			return
		}
		player.Fmessage = "You got the super shotgun!"
		sound = int32(sfx_wpnup)
	default:
		i_Error("P_SpecialThing: Unknown gettable thing")
	}
	if special.Fflags&mf_COUNTITEM != 0 {
		player.Fitemcount++
	}
	p_RemoveMobj(special)
	player.Fbonuscount += BONUSADD
	if player == &players[consoleplayer] {
		s_StartSound(nil, sound)
	}
}

// C documentation
//
//	//
//	// KillMobj
//	//
func p_KillMobj(source *mobj_t, target *mobj_t) {
	var item mobjtype_t
	var mo *mobj_t
	target.Fflags &= ^(mf_SHOOTABLE | mf_FLOAT | mf_SKULLFLY)
	if target.Ftype1 != mt_SKULL {
		target.Fflags &= ^mf_NOGRAVITY
	}
	target.Fflags |= mf_CORPSE | mf_DROPOFF
	target.Fheight >>= 2
	if source != nil && source.Fplayer != nil {
		// count for intermission
		if target.Fflags&mf_COUNTKILL != 0 {
			source.Fplayer.Fkillcount++
		}
		if target.Fplayer != nil {
			idx := playerIndex(target.Fplayer)
			source.Fplayer.Ffrags[idx]++
		}
	} else {
		if netgame == 0 && target.Fflags&mf_COUNTKILL != 0 {
			// count all monster deaths,
			// even those caused by other monsters
			players[0].Fkillcount++
		}
	}
	if target.Fplayer != nil {
		// count environment kills against you
		if source == nil {
			idx := playerIndex(target.Fplayer)
			target.Fplayer.Ffrags[idx]++
		}
		target.Fflags &= ^mf_SOLID
		target.Fplayer.Fplayerstate = Pst_DEAD
		p_DropWeapon(target.Fplayer)
		if target.Fplayer == &players[consoleplayer] && automapactive != 0 {
			// don't die in auto map,
			// switch view prior to dying
			am_Stop()
		}
	}
	if target.Fhealth < -target.Finfo.Fspawnhealth && target.Finfo.Fxdeathstate != 0 {
		p_SetMobjState(target, target.Finfo.Fxdeathstate)
	} else {
		p_SetMobjState(target, target.Finfo.Fdeathstate)
	}
	target.Ftics -= p_Random() & 3
	if target.Ftics < 1 {
		target.Ftics = 1
	}
	//	i_StartSound (&actor->r, actor->info->deathsound);
	// In Chex Quest, monsters don't drop items.
	if gameversion == exe_chex {
		return
	}
	// Drop stuff.
	// This determines the kind of object spawned
	// during the death frame of a thing.
	switch target.Ftype1 {
	case mt_WOLFSS:
		fallthrough
	case mt_POSSESSED:
		item = mt_CLIP
	case mt_SHOTGUY:
		item = mt_SHOTGUN
	case mt_CHAINGUY:
		item = mt_CHAINGUN
	default:
		return
	}
	mo = p_SpawnMobj(target.Fx, target.Fy, -1-0x7fffffff, item)
	mo.Fflags |= mf_DROPPED // special versions of items
}

// C documentation
//
//	//
//	// P_DamageMobj
//	// Damages both enemies and players
//	// "inflictor" is the thing that caused the damage
//	//  creature or missile, can be NULL (slime, etc)
//	// "source" is the thing to target after taking damage
//	//  creature or NULL
//	// Source and inflictor are the same for melee attacks.
//	// Source can be NULL for slime, barrel explosions
//	// and other environmental stuff.
//	//
func p_DamageMobj(target *mobj_t, inflictor *mobj_t, source *mobj_t, damage int32) {
	var ang uint32
	var player *player_t
	var saved, temp, v3 int32
	var thrust, v1, v2 fixed_t
	if target.Fflags&mf_SHOOTABLE == 0 {
		return
	} // shouldn't happen...
	if target.Fhealth <= 0 {
		return
	}
	if target.Fflags&mf_SKULLFLY != 0 {
		v2 = 0
		target.Fmomz = v2
		v1 = v2
		target.Fmomy = v1
		target.Fmomx = v1
	}
	player = target.Fplayer
	if player != nil && gameskill == sk_baby {
		damage >>= 1
	} // take half damage in trainer mode
	// Some close combat weapons should not
	// inflict thrust and push the victim out of reach,
	// thus kick away unless using the chainsaw.
	if inflictor != nil && target.Fflags&mf_NOCLIP == 0 && (source == nil || source.Fplayer == nil || source.Fplayer.Freadyweapon != wp_chainsaw) {
		ang = r_PointToAngle2(inflictor.Fx, inflictor.Fy, target.Fx, target.Fy)
		thrust = damage * (1 << FRACBITS >> 3) * 100 / target.Finfo.Fmass
		// make fall forwards sometimes
		if damage < 40 && damage > target.Fhealth && target.Fz-inflictor.Fz > 64*(1<<FRACBITS) && p_Random()&1 != 0 {
			ang += uint32(ANG1803)
			thrust *= 4
		}
		ang >>= ANGLETOFINESHIFT
		target.Fmomx += fixedMul(thrust, finecosine[ang])
		target.Fmomy += fixedMul(thrust, finesine[ang])
	}
	// player specific
	if player != nil {
		// end of game hell hack
		if int32(target.Fsubsector.Fsector.Fspecial) == 11 && damage >= target.Fhealth {
			damage = target.Fhealth - 1
		}
		// Below certain threshold,
		// ignore damage in GOD mode, or with INVUL power.
		if damage < 1000 && (player.Fcheats&CF_GODMODE != 0 || player.Fpowers[pw_invulnerability] != 0) {
			return
		}
		if player.Farmortype != 0 {
			if player.Farmortype == 1 {
				saved = damage / 3
			} else {
				saved = damage / 2
			}
			if player.Farmorpoints <= saved {
				// armor is used up
				saved = player.Farmorpoints
				player.Farmortype = 0
			}
			player.Farmorpoints -= saved
			damage -= saved
		}
		player.Fhealth -= damage // mirror mobj health here for Dave
		if player.Fhealth < 0 {
			player.Fhealth = 0
		}
		player.Fattacker = source
		player.Fdamagecount += damage // add damage before armor / invuln
		if player.Fdamagecount > 100 {
			player.Fdamagecount = 100
		} // teleport stomp does 10k points...
		if damage < 100 {
			v3 = damage
		} else {
			v3 = 100
		}
		temp = v3
		if player == &players[consoleplayer] {
			i_Tactile(40, 10, 40+temp*2)
		}
	}
	// do the damage
	target.Fhealth -= damage
	if target.Fhealth <= 0 {
		p_KillMobj(source, target)
		return
	}
	if p_Random() < target.Finfo.Fpainchance && target.Fflags&mf_SKULLFLY == 0 {
		target.Fflags |= mf_JUSTHIT // fight back!
		p_SetMobjState(target, target.Finfo.Fpainstate)
	}
	target.Freactiontime = 0 // we're awake now...
	if (target.Fthreshold == 0 || target.Ftype1 == mt_VILE) && source != nil && source != target && source.Ftype1 != mt_VILE {
		// if not intent on another player,
		// chase after this one
		target.Ftarget = source
		target.Fthreshold = BASETHRESHOLD
		if target.Fstate == &states[target.Finfo.Fspawnstate] && target.Finfo.Fseestate != s_NULL {
			p_SetMobjState(target, target.Finfo.Fseestate)
		}
	}
}

// State.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

//
// FIRELIGHT FLICKER
//

// C documentation
//
//	//
//	// T_FireFlicker
//	//
func (flick *fireflicker_t) ThinkerFunc() {
	t_FireFlicker(flick)
}
func t_FireFlicker(flick *fireflicker_t) {
	var amount int32
	flick.Fcount--
	if flick.Fcount != 0 {
		return
	}
	amount = p_Random() & 3 * 16
	if int32(flick.Fsector.Flightlevel)-amount < flick.Fminlight {
		flick.Fsector.Flightlevel = int16(flick.Fminlight)
	} else {
		flick.Fsector.Flightlevel = int16(flick.Fmaxlight - amount)
	}
	flick.Fcount = 4
}

// C documentation
//
//	//
//	// P_SpawnFireFlicker
//	//
func p_SpawnFireFlicker(sector *sector_t) {
	// Note that we are resetting sector attributes.
	// Nothing special about it during gameplay.
	sector.Fspecial = 0
	flick := &fireflicker_t{}
	p_AddThinker(&flick.Fthinker)
	flick.Fthinker.Ffunction = flick
	flick.Fsector = sector
	flick.Fmaxlight = int32(sector.Flightlevel)
	flick.Fminlight = p_FindMinSurroundingLight(sector, int32(sector.Flightlevel)) + 16
	flick.Fcount = 4
}

//
// BROKEN LIGHT FLASHING
//

// C documentation
//
//	//
//	// T_LightFlash
//	// Do flashing lights.
//	//
func (flash *lightflash_t) ThinkerFunc() {
	t_LightFlash(flash)
}
func t_LightFlash(flash *lightflash_t) {
	flash.Fcount--
	if flash.Fcount != 0 {
		return
	}
	if int32(flash.Fsector.Flightlevel) == flash.Fmaxlight {
		flash.Fsector.Flightlevel = int16(flash.Fminlight)
		flash.Fcount = p_Random()&flash.Fmintime + 1
	} else {
		flash.Fsector.Flightlevel = int16(flash.Fmaxlight)
		flash.Fcount = p_Random()&flash.Fmaxtime + 1
	}
}

// C documentation
//
//	//
//	// P_SpawnLightFlash
//	// After the map has been loaded, scan each sector
//	// for specials that spawn thinkers
//	//
func p_SpawnLightFlash(sector *sector_t) {
	// nothing special about it during gameplay
	sector.Fspecial = 0
	flashP := &lightflash_t{}
	p_AddThinker(&flashP.Fthinker)
	flashP.Fthinker.Ffunction = flashP
	flashP.Fsector = sector
	flashP.Fmaxlight = int32(sector.Flightlevel)
	flashP.Fminlight = p_FindMinSurroundingLight(sector, int32(sector.Flightlevel))
	flashP.Fmaxtime = 64
	flashP.Fmintime = 7
	flashP.Fcount = p_Random()&flashP.Fmaxtime + 1
}

//
// STROBE LIGHT FLASHING
//

// C documentation
//
//	//
//	// T_StrobeFlash
//	//
func (flash *strobe_t) ThinkerFunc() {
	t_StrobeFlash(flash)
}
func t_StrobeFlash(flash *strobe_t) {
	flash.Fcount--
	if flash.Fcount != 0 {
		return
	}
	if int32(flash.Fsector.Flightlevel) == flash.Fminlight {
		flash.Fsector.Flightlevel = int16(flash.Fmaxlight)
		flash.Fcount = flash.Fbrighttime
	} else {
		flash.Fsector.Flightlevel = int16(flash.Fminlight)
		flash.Fcount = flash.Fdarktime
	}
}

// C documentation
//
//	//
//	// P_SpawnStrobeFlash
//	// After the map has been loaded, scan each sector
//	// for specials that spawn thinkers
//	//
func p_SpawnStrobeFlash(sector *sector_t, fastOrSlow int32, inSync int32) {
	flashP := &strobe_t{}
	p_AddThinker(&flashP.Fthinker)
	flashP.Fsector = sector
	flashP.Fdarktime = fastOrSlow
	flashP.Fbrighttime = STROBEBRIGHT
	flashP.Fthinker.Ffunction = flashP
	flashP.Fmaxlight = int32(sector.Flightlevel)
	flashP.Fminlight = p_FindMinSurroundingLight(sector, int32(sector.Flightlevel))
	if flashP.Fminlight == flashP.Fmaxlight {
		flashP.Fminlight = 0
	}
	// nothing special about it during gameplay
	sector.Fspecial = 0
	if inSync == 0 {
		flashP.Fcount = p_Random()&7 + 1
	} else {
		flashP.Fcount = 1
	}
}

// C documentation
//
//	//
//	// Start strobing lights (usually from a trigger)
//	//
func ev_StartLightStrobing(line *line_t) {
	var sec *sector_t
	var secnum, v1 int32
	secnum = -1
	for {
		v1 = p_FindSectorFromLineTag(line, secnum)
		secnum = v1
		if !(v1 >= 0) {
			break
		}
		sec = &sectors[secnum]
		if sec.Fspecialdata != nil {
			continue
		}
		p_SpawnStrobeFlash(sec, SLOWDARK, 0)
	}
}

// C documentation
//
//	//
//	// TURN LINE'S TAG LIGHTS OFF
//	//
func ev_TurnTagLightsOff(line *line_t) {
	var min int32
	var templine *line_t
	var tsec *sector_t
	for j := int32(0); j < numsectors; j++ {
		sector := &sectors[j]
		if int32(sector.Ftag) == int32(line.Ftag) {
			min = int32(sector.Flightlevel)
			for i := int32(0); i < sector.Flinecount; i++ {
				templine = sector.Flines[i]
				tsec = getNextSector(templine, sector)
				if tsec == nil {
					continue
				}
				if int32(tsec.Flightlevel) < min {
					min = int32(tsec.Flightlevel)
				}
			}
			sector.Flightlevel = int16(min)
		}
	}
}

// C documentation
//
//	//
//	// TURN LINE'S TAG LIGHTS ON
//	//
func ev_LightTurnOn(line *line_t, bright int32) {
	var temp *sector_t
	var templine *line_t
	for i := int32(0); i < numsectors; i++ {
		sector := &sectors[i]
		if int32(sector.Ftag) == int32(line.Ftag) {
			// bright = 0 means to search
			// for highest light level
			// surrounding sector
			if bright == 0 {
				for j := int32(0); j < sector.Flinecount; j++ {
					templine = sector.Flines[j]
					temp = getNextSector(templine, sector)
					if temp == nil {
						continue
					}
					if int32(temp.Flightlevel) > bright {
						bright = int32(temp.Flightlevel)
					}
				}
			}
			sector.Flightlevel = int16(bright)
		}
	}
}

//
// Spawn glowing light
//

func (g *glow_t) ThinkerFunc() {
	t_Glow(g)
}
func t_Glow(g *glow_t) {
	switch g.Fdirection {
	case -1:
		// DOWN
		g.Fsector.Flightlevel -= GLOWSPEED
		if int32(g.Fsector.Flightlevel) <= g.Fminlight {
			g.Fsector.Flightlevel += GLOWSPEED
			g.Fdirection = 1
		}
	case 1:
		// UP
		g.Fsector.Flightlevel += GLOWSPEED
		if int32(g.Fsector.Flightlevel) >= g.Fmaxlight {
			g.Fsector.Flightlevel -= GLOWSPEED
			g.Fdirection = -1
		}
		break
	}
}

func p_SpawnGlowingLight(sector *sector_t) {
	gP := &glow_t{}
	p_AddThinker(&gP.Fthinker)
	gP.Fsector = sector
	gP.Fminlight = p_FindMinSurroundingLight(sector, int32(sector.Flightlevel))
	gP.Fmaxlight = int32(sector.Flightlevel)
	gP.Fthinker.Ffunction = gP
	gP.Fdirection = -1
	sector.Fspecial = 0
}

const ANG1805 = 2147483648
const DEFAULT_SPECHIT_MAGIC = 29400216

//
// TELEPORT MOVE
//

// C documentation
//
//	//
//	// PIT_StompThing
//	//
func pit_StompThing(thing *mobj_t) boolean {
	var blockdist fixed_t
	if thing.Fflags&mf_SHOOTABLE == 0 {
		return 1
	}
	blockdist = thing.Fradius + tmthing.Fradius
	if xabs(thing.Fx-tmx) >= blockdist || xabs(thing.Fy-tmy) >= blockdist {
		// didn't hit it
		return 1
	}
	// don't clip against self
	if thing == tmthing {
		return 1
	}
	// monsters don't stomp things except on boss level
	if tmthing.Fplayer == nil && gamemap != 30 {
		return 0
	}
	p_DamageMobj(thing, tmthing, tmthing, 10000)
	return 1
}

// C documentation
//
//	//
//	// P_TeleportMove
//	//
func p_TeleportMove(thing *mobj_t, x fixed_t, y fixed_t) boolean {
	var bx, by, xh, xl, yh, yl int32
	var newsubsec *subsector_t
	var v1 fixed_t
	// kill anything occupying the position
	tmthing = thing
	tmflags = thing.Fflags
	tmx = x
	tmy = y
	tmbbox[BOXTOP] = y + tmthing.Fradius
	tmbbox[BOXBOTTOM] = y - tmthing.Fradius
	tmbbox[BOXRIGHT] = x + tmthing.Fradius
	tmbbox[BOXLEFT] = x - tmthing.Fradius
	newsubsec = r_PointInSubsector(x, y)
	ceilingline = nil
	// The base floor/ceiling is from the subsector
	// that contains the point.
	// Any contacted lines the step closer together
	// will adjust them.
	v1 = newsubsec.Fsector.Ffloorheight
	tmdropoffz = v1
	tmfloorz = v1
	tmceilingz = newsubsec.Fsector.Fceilingheight
	validcount++
	numspechit = 0
	// stomp on any things contacted
	xl = (tmbbox[BOXLEFT] - bmaporgx - 32*(1<<FRACBITS)) >> (FRACBITS + 7)
	xh = (tmbbox[BOXRIGHT] - bmaporgx + 32*(1<<FRACBITS)) >> (FRACBITS + 7)
	yl = (tmbbox[BOXBOTTOM] - bmaporgy - 32*(1<<FRACBITS)) >> (FRACBITS + 7)
	yh = (tmbbox[BOXTOP] - bmaporgy + 32*(1<<FRACBITS)) >> (FRACBITS + 7)
	bx = xl
	for {
		if !(bx <= xh) {
			break
		}
		by = yl
		for {
			if !(by <= yh) {
				break
			}
			if p_BlockThingsIterator(bx, by, pit_StompThing) == 0 {
				return 0
			}
			goto _3
		_3:
			;
			by++
		}
		goto _2
	_2:
		;
		bx++
	}
	// the move is ok,
	// so link the thing into its new position
	p_UnsetThingPosition(thing)
	thing.Ffloorz = tmfloorz
	thing.Fceilingz = tmceilingz
	thing.Fx = x
	thing.Fy = y
	p_SetThingPosition(thing)
	return 1
}

// C documentation
//
//	//
//	// PIT_CheckLine
//	// Adjusts tmfloorz and tmceilingz as lines are contacted
//	//
func pit_CheckLine(ld *line_t) boolean {
	if tmbbox[BOXRIGHT] <= ld.Fbbox[BOXLEFT] || tmbbox[BOXLEFT] >= ld.Fbbox[BOXRIGHT] || tmbbox[BOXTOP] <= ld.Fbbox[BOXBOTTOM] || tmbbox[BOXBOTTOM] >= ld.Fbbox[BOXTOP] {
		return 1
	}
	if p_BoxOnLineSide(&tmbbox, ld) != -1 {
		return 1
	}
	// A line has been hit
	// The moving thing's destination position will cross
	// the given line.
	// If this should not be allowed, return false.
	// If the line is special, keep track of it
	// to process later if the move is proven ok.
	// NOTE: specials are NOT sorted by order,
	// so two special lines that are only 8 pixels apart
	// could be crossed in either order.
	if ld.Fbacksector == nil {
		return 0
	} // one sided line
	if tmthing.Fflags&mf_MISSILE == 0 {
		if int32(ld.Fflags)&ml_BLOCKING != 0 {
			return 0
		} // explicitly blocking everything
		if tmthing.Fplayer == nil && int32(ld.Fflags)&ml_BLOCKMONSTERS != 0 {
			return 0
		} // block monsters only
	}
	// set openrange, opentop, openbottom
	p_LineOpening(ld)
	// adjust floor / ceiling heights
	if opentop < tmceilingz {
		tmceilingz = opentop
		ceilingline = ld
	}
	if openbottom > tmfloorz {
		tmfloorz = openbottom
	}
	if lowfloor < tmdropoffz {
		tmdropoffz = lowfloor
	}
	// if contacted a special line, add it to the list
	if ld.Fspecial != 0 {
		spechit[numspechit] = ld
		numspechit++
		// fraggle: spechits overrun emulation code from prboom-plus
		if numspechit > MAXSPECIALCROSS_ORIGINAL {
			spechitOverrun(ld)
		}
	}
	return 1
}

// C documentation
//
//	//
//	// PIT_CheckThing
//	//
func pit_CheckThing(thing *mobj_t) boolean {
	var blockdist, v1, v2 fixed_t
	var damage int32
	var solid boolean
	if thing.Fflags&(mf_SOLID|mf_SPECIAL|mf_SHOOTABLE) == 0 {
		return 1
	}
	blockdist = thing.Fradius + tmthing.Fradius
	if xabs(thing.Fx-tmx) >= blockdist || xabs(thing.Fy-tmy) >= blockdist {
		// didn't hit it
		return 1
	}
	// don't clip against self
	if thing == tmthing {
		return 1
	}
	// check for skulls slamming into things
	if tmthing.Fflags&mf_SKULLFLY != 0 {
		damage = (p_Random()%8 + 1) * tmthing.Finfo.Fdamage
		p_DamageMobj(thing, tmthing, tmthing, damage)
		tmthing.Fflags &^= mf_SKULLFLY
		v2 = 0
		tmthing.Fmomz = v2
		v1 = v2
		tmthing.Fmomy = v1
		tmthing.Fmomx = v1
		p_SetMobjState(tmthing, tmthing.Finfo.Fspawnstate)
		return 0 // stop moving
	}
	// missiles can hit other things
	if tmthing.Fflags&mf_MISSILE != 0 {
		// see if it went over / under
		if tmthing.Fz > thing.Fz+thing.Fheight {
			return 1
		} // overhead
		if tmthing.Fz+tmthing.Fheight < thing.Fz {
			return 1
		} // underneath
		if tmthing.Ftarget != nil && (tmthing.Ftarget.Ftype1 == thing.Ftype1 || tmthing.Ftarget.Ftype1 == mt_KNIGHT && thing.Ftype1 == mt_BRUISER || tmthing.Ftarget.Ftype1 == mt_BRUISER && thing.Ftype1 == mt_KNIGHT) {
			// Don't hit same species as originator.
			if thing == tmthing.Ftarget {
				return 1
			}
			// sdh: Add deh_species_infighting here.  We can override the
			// "monsters of the same species cant hurt each other" behavior
			// through dehacked patches
			if thing.Ftype1 != mt_PLAYER && DEH_DEFAULT_SPECIES_INFIGHTING == 0 {
				// Explode, but do no damage.
				// Let players missile other players.
				return 0
			}
		}
		if thing.Fflags&mf_SHOOTABLE == 0 {
			// didn't do any damage
			return booluint32(thing.Fflags&mf_SOLID == 0)
		}
		// damage / explode
		damage = (p_Random()%8 + 1) * tmthing.Finfo.Fdamage
		p_DamageMobj(thing, tmthing, tmthing.Ftarget, damage)
		// don't traverse any more
		return 0
	}
	// check for special pickup
	if thing.Fflags&mf_SPECIAL != 0 {
		solid = uint32(thing.Fflags & mf_SOLID)
		if tmflags&mf_PICKUP != 0 {
			// can remove thing
			p_TouchSpecialThing(thing, tmthing)
		}
		return booluint32(solid == 0)
	}
	return booluint32(thing.Fflags&mf_SOLID == 0)
}

//
// MOVEMENT CLIPPING
//

// C documentation
//
//	//
//	// P_CheckPosition
//	// This is purely informative, nothing is modified
//	// (except things picked up).
//	//
//	// in:
//	//  a mobj_t (can be valid or invalid)
//	//  a position to be checked
//	//   (doesn't need to be related to the mobj_t->x,y)
//	//
//	// during:
//	//  special things are touched if mf_PICKUP
//	//  early out on solid lines?
//	//
//	// out:
//	//  newsubsec
//	//  floorz
//	//  ceilingz
//	//  tmdropoffz
//	//   the lowest point contacted
//	//   (monsters won't move to a dropoff)
//	//  speciallines[]
//	//  numspeciallines
//	//
func p_CheckPosition(thing *mobj_t, x fixed_t, y fixed_t) boolean {
	var bx, by, xh, xl, yh, yl int32
	var newsubsec *subsector_t
	var v1 fixed_t
	tmthing = thing
	tmflags = thing.Fflags
	tmx = x
	tmy = y
	tmbbox[BOXTOP] = y + tmthing.Fradius
	tmbbox[BOXBOTTOM] = y - tmthing.Fradius
	tmbbox[BOXRIGHT] = x + tmthing.Fradius
	tmbbox[BOXLEFT] = x - tmthing.Fradius
	newsubsec = r_PointInSubsector(x, y)
	ceilingline = nil
	// The base floor / ceiling is from the subsector
	// that contains the point.
	// Any contacted lines the step closer together
	// will adjust them.
	v1 = newsubsec.Fsector.Ffloorheight
	tmdropoffz = v1
	tmfloorz = v1
	tmceilingz = newsubsec.Fsector.Fceilingheight
	validcount++
	numspechit = 0
	if tmflags&mf_NOCLIP != 0 {
		return 1
	}
	// Check things first, possibly picking things up.
	// The bounding box is extended by MAXRADIUS
	// because mobj_ts are grouped into mapblocks
	// based on their origin point, and can overlap
	// into adjacent blocks by up to MAXRADIUS units.
	xl = (tmbbox[BOXLEFT] - bmaporgx - 32*(1<<FRACBITS)) >> (FRACBITS + 7)
	xh = (tmbbox[BOXRIGHT] - bmaporgx + 32*(1<<FRACBITS)) >> (FRACBITS + 7)
	yl = (tmbbox[BOXBOTTOM] - bmaporgy - 32*(1<<FRACBITS)) >> (FRACBITS + 7)
	yh = (tmbbox[BOXTOP] - bmaporgy + 32*(1<<FRACBITS)) >> (FRACBITS + 7)
	bx = xl
	for {
		if !(bx <= xh) {
			break
		}
		by = yl
		for {
			if !(by <= yh) {
				break
			}
			if p_BlockThingsIterator(bx, by, pit_CheckThing) == 0 {
				return 0
			}
			goto _3
		_3:
			;
			by++
		}
		goto _2
	_2:
		;
		bx++
	}
	// check lines
	xl = (tmbbox[BOXLEFT] - bmaporgx) >> (FRACBITS + 7)
	xh = (tmbbox[BOXRIGHT] - bmaporgx) >> (FRACBITS + 7)
	yl = (tmbbox[BOXBOTTOM] - bmaporgy) >> (FRACBITS + 7)
	yh = (tmbbox[BOXTOP] - bmaporgy) >> (FRACBITS + 7)
	bx = xl
	for {
		if !(bx <= xh) {
			break
		}
		by = yl
		for {
			if !(by <= yh) {
				break
			}
			if p_BlockLinesIterator(bx, by, pit_CheckLine) == 0 {
				return 0
			}
			goto _5
		_5:
			;
			by++
		}
		goto _4
	_4:
		;
		bx++
	}
	return 1
}

// C documentation
//
//	//
//	// P_TryMove
//	// Attempt to move to a new position,
//	// crossing special lines unless mf_TELEPORT is set.
//	//
func p_TryMove(thing *mobj_t, x fixed_t, y fixed_t) boolean {
	var oldside, side, v1 int32
	var oldx, oldy fixed_t
	mthing := thing
	floatok = 0
	if p_CheckPosition(thing, x, y) == 0 {
		return 0
	} // solid wall or thing
	if mthing.Fflags&mf_NOCLIP == 0 {
		if tmceilingz-tmfloorz < mthing.Fheight {
			return 0
		} // doesn't fit
		floatok = 1
		if mthing.Fflags&mf_TELEPORT == 0 && tmceilingz-mthing.Fz < mthing.Fheight {
			return 0
		} // mobj must lower itself to fit
		if mthing.Fflags&mf_TELEPORT == 0 && tmfloorz-mthing.Fz > 24*(1<<FRACBITS) {
			return 0
		} // too big a step up
		if mthing.Fflags&(mf_DROPOFF|mf_FLOAT) == 0 && tmfloorz-tmdropoffz > 24*(1<<FRACBITS) {
			return 0
		} // don't stand over a dropoff
	}
	// the move is ok,
	// so link the thing into its new position
	p_UnsetThingPosition(thing)
	oldx = mthing.Fx
	oldy = mthing.Fy
	mthing.Ffloorz = tmfloorz
	mthing.Fceilingz = tmceilingz
	mthing.Fx = x
	mthing.Fy = y
	p_SetThingPosition(thing)
	// if any special lines were hit, do the effect
	if mthing.Fflags&(mf_TELEPORT|mf_NOCLIP) == 0 {
		for {
			v1 = numspechit
			numspechit--
			if v1 == 0 {
				break
			}
			// see if the line was crossed
			ld := spechit[numspechit]
			side = p_PointOnLineSide(mthing.Fx, mthing.Fy, ld)
			oldside = p_PointOnLineSide(oldx, oldy, ld)
			if side != oldside {
				if ld.Fspecial != 0 {
					p_CrossSpecialLine(lineIndex(ld), oldside, thing)
				}
			}
		}
	}
	return 1
}

// C documentation
//
//	//
//	// P_ThingHeightClip
//	// Takes a valid thing and adjusts the thing->floorz,
//	// thing->ceilingz, and possibly thing->z.
//	// This is called for all nearby monsters
//	// whenever a sector changes height.
//	// If the thing doesn't fit,
//	// the z will be set to the lowest value
//	// and false will be returned.
//	//
func p_ThingHeightClip(thing *mobj_t) boolean {
	var onfloor boolean
	onfloor = booluint32(thing.Fz == thing.Ffloorz)
	p_CheckPosition(thing, thing.Fx, thing.Fy)
	// what about stranding a monster partially off an edge?
	thing.Ffloorz = tmfloorz
	thing.Fceilingz = tmceilingz
	if onfloor != 0 {
		// walking monsters rise and fall with the floor
		thing.Fz = thing.Ffloorz
	} else {
		// don't adjust a floating monster unless forced to
		if thing.Fz+thing.Fheight > thing.Fceilingz {
			thing.Fz = thing.Fceilingz - thing.Fheight
		}
	}
	if thing.Fceilingz-thing.Ffloorz < thing.Fheight {
		return 0
	}
	return 1
}

// C documentation
//
//	//
//	// P_HitSlideLine
//	// Adjusts the xmove / ymove
//	// so that the next move will slide along the wall.
//	//
func p_HitSlideLine(ld *line_t) {
	var deltaangle, lineangle, moveangle angle_t
	var movelen, newlen fixed_t
	var side int32
	if ld.Fslopetype == st_HORIZONTAL {
		tmymove = 0
		return
	}
	if ld.Fslopetype == st_VERTICAL {
		tmxmove = 0
		return
	}
	side = p_PointOnLineSide(slidemo.Fx, slidemo.Fy, ld)
	lineangle = r_PointToAngle2(0, 0, ld.Fdx, ld.Fdy)
	if side == 1 {
		lineangle += uint32(ANG1805)
	}
	moveangle = r_PointToAngle2(0, 0, tmxmove, tmymove)
	deltaangle = moveangle - lineangle
	if deltaangle > uint32(ANG1805) {
		deltaangle += uint32(ANG1805)
	}
	//	i_Error ("SlideLine: ang>ANG180");
	lineangle >>= ANGLETOFINESHIFT
	deltaangle >>= ANGLETOFINESHIFT
	movelen = p_AproxDistance(tmxmove, tmymove)
	newlen = fixedMul(movelen, finecosine[deltaangle])
	tmxmove = fixedMul(newlen, finecosine[lineangle])
	tmymove = fixedMul(newlen, finesine[lineangle])
}

// C documentation
//
//	//
//	// PTR_SlideTraverse
//	//
func ptr_SlideTraverse(in *intercept_t) boolean {
	var li *line_t
	if in.Fisaline == 0 {
		i_Error("ptr_SlideTraverse: not a line?")
	}
	li = in.Fd.Fthing.(*line_t)
	if int32(li.Fflags)&ml_TWOSIDED == 0 {
		if p_PointOnLineSide(slidemo.Fx, slidemo.Fy, li) != 0 {
			// don't hit the back side
			return 1
		}
		goto isblocking
	}
	// set openrange, opentop, openbottom
	p_LineOpening(li)
	if openrange < slidemo.Fheight {
		goto isblocking
	} // doesn't fit
	if opentop-slidemo.Fz < slidemo.Fheight {
		goto isblocking
	} // mobj is too high
	if openbottom-slidemo.Fz > 24*(1<<FRACBITS) {
		goto isblocking
	} // too big a step up
	// this line doesn't block movement
	return 1
	// the line does block movement,
	// see if it is closer than best so far
	goto isblocking
isblocking:
	;
	if in.Ffrac < bestslidefrac {
		bestslidefrac = in.Ffrac
		bestslideline = li
	}
	return 0 // stop
}

// C documentation
//
//	//
//	// P_SlideMove
//	// The momx / momy move is bad, so try to slide
//	// along a wall.
//	// Find the first line hit, move flush to it,
//	// and slide along it
//	//
//	// This is a kludgy mess.
//	//
func p_SlideMove(mo *mobj_t) {
	var hitcount, v1 int32
	var leadx, leady, newx, newy, trailx, traily fixed_t
	slidemo = mo
	hitcount = 0
	goto retry
retry:
	;
	hitcount++
	v1 = hitcount
	if v1 == 3 {
		goto stairstep
	} // don't loop forever
	// trace along the three leading corners
	if mo.Fmomx > 0 {
		leadx = mo.Fx + mo.Fradius
		trailx = mo.Fx - mo.Fradius
	} else {
		leadx = mo.Fx - mo.Fradius
		trailx = mo.Fx + mo.Fradius
	}
	if mo.Fmomy > 0 {
		leady = mo.Fy + mo.Fradius
		traily = mo.Fy - mo.Fradius
	} else {
		leady = mo.Fy - mo.Fradius
		traily = mo.Fy + mo.Fradius
	}
	bestslidefrac = 1<<FRACBITS + 1
	p_PathTraverse(leadx, leady, leadx+mo.Fmomx, leady+mo.Fmomy, PT_ADDLINES, ptr_SlideTraverse)
	p_PathTraverse(trailx, leady, trailx+mo.Fmomx, leady+mo.Fmomy, PT_ADDLINES, ptr_SlideTraverse)
	p_PathTraverse(leadx, traily, leadx+mo.Fmomx, traily+mo.Fmomy, PT_ADDLINES, ptr_SlideTraverse)
	// move up to the wall
	if bestslidefrac != 1<<FRACBITS+1 {
		goto _2
	}
	// the move most have hit the middle, so stairstep
	goto stairstep
stairstep:
	;
	if p_TryMove(mo, mo.Fx, mo.Fy+mo.Fmomy) == 0 {
		p_TryMove(mo, mo.Fx+mo.Fmomx, mo.Fy)
	}
	return
_2:
	;
	// fudge a bit to make sure it doesn't hit
	bestslidefrac -= 0x800
	if bestslidefrac > 0 {
		newx = fixedMul(mo.Fmomx, bestslidefrac)
		newy = fixedMul(mo.Fmomy, bestslidefrac)
		if p_TryMove(mo, mo.Fx+newx, mo.Fy+newy) == 0 {
			goto stairstep
		}
	}
	// Now continue along the wall.
	// First calculate remainder.
	bestslidefrac = 1<<FRACBITS - (bestslidefrac + 0x800)
	if bestslidefrac > 1<<FRACBITS {
		bestslidefrac = 1 << FRACBITS
	}
	if bestslidefrac <= 0 {
		return
	}
	tmxmove = fixedMul(mo.Fmomx, bestslidefrac)
	tmymove = fixedMul(mo.Fmomy, bestslidefrac)
	p_HitSlideLine(bestslideline) // clip the moves
	mo.Fmomx = tmxmove
	mo.Fmomy = tmymove
	if p_TryMove(mo, mo.Fx+tmxmove, mo.Fy+tmymove) == 0 {
		goto retry
	}
}

// C documentation
//
//	//
//	// PTR_AimTraverse
//	// Sets linetaget and aimslope when a target is aimed at.
//	//
func ptr_AimTraverse(in *intercept_t) boolean {
	var dist, slope, thingbottomslope, thingtopslope fixed_t
	var th *mobj_t
	var li *line_t
	if in.Fisaline != 0 {
		li = in.Fd.Fthing.(*line_t)
		if int32(li.Fflags)&ml_TWOSIDED == 0 {
			return 0
		} // stop
		// Crosses a two sided line.
		// A two sided line will restrict
		// the possible target ranges.
		p_LineOpening(li)
		if openbottom >= opentop {
			return 0
		} // stop
		dist = fixedMul(attackrange, in.Ffrac)
		if li.Fbacksector == nil || li.Ffrontsector.Ffloorheight != li.Fbacksector.Ffloorheight {
			slope = fixedDiv(openbottom-shootz, dist)
			if slope > bottomslope {
				bottomslope = slope
			}
		}
		if li.Fbacksector == nil || li.Ffrontsector.Fceilingheight != li.Fbacksector.Fceilingheight {
			slope = fixedDiv(opentop-shootz, dist)
			if slope < topslope {
				topslope = slope
			}
		}
		if topslope <= bottomslope {
			return 0
		} // stop
		return 1 // shot continues
	}
	// shoot a thing
	th = in.Fd.Fthing.(*mobj_t)
	if th == shootthing {
		return 1
	} // can't shoot self
	if th.Fflags&mf_SHOOTABLE == 0 {
		return 1
	} // corpse or something
	// check angles to see if the thing can be aimed at
	dist = fixedMul(attackrange, in.Ffrac)
	thingtopslope = fixedDiv(th.Fz+th.Fheight-shootz, dist)
	if thingtopslope < bottomslope {
		return 1
	} // shot over the thing
	thingbottomslope = fixedDiv(th.Fz-shootz, dist)
	if thingbottomslope > topslope {
		return 1
	} // shot under the thing
	// this thing can be hit!
	if thingtopslope > topslope {
		thingtopslope = topslope
	}
	if thingbottomslope < bottomslope {
		thingbottomslope = bottomslope
	}
	aimslope = (thingtopslope + thingbottomslope) / 2
	linetarget = th
	return 0 // don't go any farther
}

// C documentation
//
//	//
//	// PTR_ShootTraverse
//	//
func ptr_ShootTraverse(in *intercept_t) boolean {
	var dist, frac, slope, thingbottomslope, thingtopslope, x, y, z fixed_t
	var th *mobj_t
	var li *line_t
	if in.Fisaline != 0 {
		li = in.Fd.Fthing.(*line_t)
		if li.Fspecial != 0 {
			p_ShootSpecialLine(shootthing, li)
		}
		if int32(li.Fflags)&ml_TWOSIDED == 0 {
			goto hitline
		}
		// crosses a two sided line
		p_LineOpening(li)
		dist = fixedMul(attackrange, in.Ffrac)
		// e6y: emulation of missed back side on two-sided lines.
		// backsector can be NULL when emulating missing back side.
		if li.Fbacksector == nil {
			slope = fixedDiv(openbottom-shootz, dist)
			if slope > aimslope {
				goto hitline
			}
			slope = fixedDiv(opentop-shootz, dist)
			if slope < aimslope {
				goto hitline
			}
		} else {
			if li.Ffrontsector.Ffloorheight != li.Fbacksector.Ffloorheight {
				slope = fixedDiv(openbottom-shootz, dist)
				if slope > aimslope {
					goto hitline
				}
			}
			if li.Ffrontsector.Fceilingheight != li.Fbacksector.Fceilingheight {
				slope = fixedDiv(opentop-shootz, dist)
				if slope < aimslope {
					goto hitline
				}
			}
		}
		// shot continues
		return 1
		// hit line
		goto hitline
	hitline:
		;
		// position a bit closer
		frac = in.Ffrac - fixedDiv(4*(1<<FRACBITS), attackrange)
		x = trace.Fx + fixedMul(trace.Fdx, frac)
		y = trace.Fy + fixedMul(trace.Fdy, frac)
		z = shootz + fixedMul(aimslope, fixedMul(frac, attackrange))
		if int32(li.Ffrontsector.Fceilingpic) == skyflatnum {
			// don't shoot the sky!
			if z > li.Ffrontsector.Fceilingheight {
				return 0
			}
			// it's a sky hack wall
			if li.Fbacksector != nil && int32(li.Fbacksector.Fceilingpic) == skyflatnum {
				return 0
			}
		}
		// Spawn bullet puffs.
		p_SpawnPuff(x, y, z)
		// don't go any farther
		return 0
	}
	// shoot a thing
	th = in.Fd.Fthing.(*mobj_t)
	if th == shootthing {
		return 1
	} // can't shoot self
	if th.Fflags&mf_SHOOTABLE == 0 {
		return 1
	} // corpse or something
	// check angles to see if the thing can be aimed at
	dist = fixedMul(attackrange, in.Ffrac)
	thingtopslope = fixedDiv(th.Fz+th.Fheight-shootz, dist)
	if thingtopslope < aimslope {
		return 1
	} // shot over the thing
	thingbottomslope = fixedDiv(th.Fz-shootz, dist)
	if thingbottomslope > aimslope {
		return 1
	} // shot under the thing
	// hit thing
	// position a bit closer
	frac = in.Ffrac - fixedDiv(10*(1<<FRACBITS), attackrange)
	x = trace.Fx + fixedMul(trace.Fdx, frac)
	y = trace.Fy + fixedMul(trace.Fdy, frac)
	z = shootz + fixedMul(aimslope, fixedMul(frac, attackrange))
	// Spawn bullet puffs or blod spots,
	// depending on target type.
	if in.Fd.Fthing.(*mobj_t).Fflags&mf_NOBLOOD != 0 {
		p_SpawnPuff(x, y, z)
	} else {
		p_SpawnBlood(x, y, z, la_damage)
	}
	if la_damage != 0 {
		p_DamageMobj(th, shootthing, shootthing, la_damage)
	}
	// don't go any farther
	return 0
}

// C documentation
//
//	//
//	// P_AimLineAttack
//	//
func p_AimLineAttack(t1 *mobj_t, angle angle_t, distance fixed_t) fixed_t {
	var x2, y2 fixed_t
	t1 = p_SubstNullMobj(t1)
	angle >>= ANGLETOFINESHIFT
	shootthing = t1
	x2 = t1.Fx + distance>>FRACBITS*finecosine[angle]
	y2 = t1.Fy + distance>>FRACBITS*finesine[angle]
	shootz = t1.Fz + t1.Fheight>>1 + 8*(1<<FRACBITS)
	// can't shoot outside view angles
	topslope = 100 * (1 << FRACBITS) / 160
	bottomslope = -100 * (1 << FRACBITS) / 160
	attackrange = distance
	linetarget = nil
	p_PathTraverse(t1.Fx, t1.Fy, x2, y2, PT_ADDLINES|PT_ADDTHINGS, ptr_AimTraverse)
	if linetarget != nil {
		return aimslope
	}
	return 0
}

// C documentation
//
//	//
//	// P_LineAttack
//	// If damage == 0, it is just a test trace
//	// that will leave linetarget set.
//	//
func p_LineAttack(t1 *mobj_t, angle angle_t, distance fixed_t, slope fixed_t, damage int32) {
	var x2, y2 fixed_t
	angle >>= ANGLETOFINESHIFT
	shootthing = t1
	la_damage = damage
	x2 = t1.Fx + distance>>FRACBITS*finecosine[angle]
	y2 = t1.Fy + distance>>FRACBITS*finesine[angle]
	shootz = t1.Fz + t1.Fheight>>1 + 8*(1<<FRACBITS)
	attackrange = distance
	aimslope = slope
	p_PathTraverse(t1.Fx, t1.Fy, x2, y2, PT_ADDLINES|PT_ADDTHINGS, ptr_ShootTraverse)
}

func ptr_UseTraverse(in *intercept_t) boolean {
	var side int32
	line := in.Fd.Fthing.(*line_t)
	if line.Fspecial == 0 {
		p_LineOpening(line)
		if openrange <= 0 {
			s_StartSound(&usething.degenmobj_t, int32(sfx_noway))
			// can't use through a wall
			return 0
		}
		// not a special line, but keep checking
		return 1
	}
	side = 0
	if p_PointOnLineSide(usething.Fx, usething.Fy, line) == 1 {
		side = 1
	}
	//	return false;		// don't use back side
	p_UseSpecialLine(usething, line, side)
	// can't use for than one special line in a row
	return 0
}

// C documentation
//
//	//
//	// P_UseLines
//	// Looks for special lines in front of the player to activate.
//	//
func p_UseLines(player *player_t) {
	var angle int32
	var x1, x2, y1, y2 fixed_t
	usething = player.Fmo
	angle = int32(player.Fmo.Fangle >> ANGLETOFINESHIFT)
	x1 = player.Fmo.Fx
	y1 = player.Fmo.Fy
	x2 = x1 + 64*(1<<FRACBITS)>>FRACBITS*finecosine[angle]
	y2 = y1 + 64*(1<<FRACBITS)>>FRACBITS*finesine[angle]
	p_PathTraverse(x1, y1, x2, y2, PT_ADDLINES, ptr_UseTraverse)
}

// C documentation
//
//	//
//	// PIT_RadiusAttack
//	// "bombsource" is the creature
//	// that caused the explosion at "bombspot".
//	//
func pit_RadiusAttack(thing *mobj_t) boolean {
	var dist, dx, dy fixed_t
	var v1 int32
	if thing.Fflags&mf_SHOOTABLE == 0 {
		return 1
	}
	// Boss spider and cyborg
	// take no damage from concussion.
	if thing.Ftype1 == mt_CYBORG || thing.Ftype1 == mt_SPIDER {
		return 1
	}
	dx = xabs(thing.Fx - bombspot.Fx)
	dy = xabs(thing.Fy - bombspot.Fy)
	if dx > dy {
		v1 = dx
	} else {
		v1 = dy
	}
	dist = v1
	dist = (dist - thing.Fradius) >> FRACBITS
	if dist < 0 {
		dist = 0
	}
	if dist >= bombdamage {
		return 1
	} // out of range
	if p_CheckSight(thing, bombspot) != 0 {
		// must be in direct path
		p_DamageMobj(thing, bombspot, bombsource, bombdamage-dist)
	}
	return 1
}

// C documentation
//
//	//
//	// P_RadiusAttack
//	// Source is the creature that caused the explosion at spot.
//	//
func p_RadiusAttack(spot *mobj_t, source *mobj_t, damage int32) {
	var dist fixed_t
	var x, xh, xl, y, yh, yl int32
	dist = (damage + 32*(1<<FRACBITS)) << FRACBITS
	yh = (spot.Fy + dist - bmaporgy) >> (FRACBITS + 7)
	yl = (spot.Fy - dist - bmaporgy) >> (FRACBITS + 7)
	xh = (spot.Fx + dist - bmaporgx) >> (FRACBITS + 7)
	xl = (spot.Fx - dist - bmaporgx) >> (FRACBITS + 7)
	bombspot = spot
	bombsource = source
	bombdamage = damage
	y = yl
	for {
		if !(y <= yh) {
			break
		}
		x = xl
		for {
			if !(x <= xh) {
				break
			}
			p_BlockThingsIterator(x, y, pit_RadiusAttack)
			goto _2
		_2:
			;
			x++
		}
		goto _1
	_1:
		;
		y++
	}
}

// C documentation
//
//	//
//	// PIT_ChangeSector
//	//
func pit_ChangeSector(thing *mobj_t) boolean {
	var mo *mobj_t
	if p_ThingHeightClip(thing) != 0 {
		// keep checking
		return 1
	}
	// crunch bodies to giblets
	if thing.Fhealth <= 0 {
		p_SetMobjState(thing, s_GIBS)
		thing.Fflags &= ^mf_SOLID
		thing.Fheight = 0
		thing.Fradius = 0
		// keep checking
		return 1
	}
	// crunch dropped items
	if thing.Fflags&mf_DROPPED != 0 {
		p_RemoveMobj(thing)
		// keep checking
		return 1
	}
	if thing.Fflags&mf_SHOOTABLE == 0 {
		// assume it is bloody gibs or something
		return 1
	}
	nofit = 1
	if crushchange != 0 && leveltime&3 == 0 {
		p_DamageMobj(thing, nil, nil, 10)
		// spray blood in a random direction
		mo = p_SpawnMobj(thing.Fx, thing.Fy, thing.Fz+thing.Fheight/2, mt_BLOOD)
		mo.Fmomx = (p_Random() - p_Random()) << 12
		mo.Fmomy = (p_Random() - p_Random()) << 12
	}
	// keep checking (crush other things)
	return 1
}

// C documentation
//
//	//
//	// P_ChangeSector
//	//
func p_ChangeSector(sector *sector_t, crunch boolean) boolean {
	var x, y int32
	nofit = 0
	crushchange = crunch
	// re-check heights for all things near the moving sector
	x = sector.Fblockbox[BOXLEFT]
	for {
		if !(x <= sector.Fblockbox[BOXRIGHT]) {
			break
		}
		y = sector.Fblockbox[BOXBOTTOM]
		for {
			if !(y <= sector.Fblockbox[BOXTOP]) {
				break
			}
			p_BlockThingsIterator(x, y, pit_ChangeSector)
			goto _2
		_2:
			;
			y++
		}
		goto _1
	_1:
		;
		x++
	}
	return nofit
}

// Code to emulate the behavior of Vanilla Doom when encountering an overrun
// of the spechit array.  This is by Andrey Budko (e6y) and comes from his
// PrBoom plus port.  A big thanks to Andrey for this.

func spechitOverrun(ld *line_t) {
	var addr int32
	var p int32
	if baseaddr == 0 {
		// This is the first time we have had an overrun.  Work out
		// what base address we are going to use.
		// Allow a spechit value to be specified on the command line.
		//!
		// @category compat
		// @arg <n>
		//
		// Use the specified magic value when emulating spechit overruns.
		//
		p = m_CheckParmWithArgs("-spechit", 1)
		if p > 0 {
			v, _ := strconv.Atoi(myargs[p+1])
			baseaddr = int32(v)
		} else {
			baseaddr = DEFAULT_SPECHIT_MAGIC
		}
	}
	// Calculate address used in doom2.exe
	addr = baseaddr + int32(lineIndex(ld))*0x3E
	switch numspechit {
	case 9:
		fallthrough
	case 10:
		fallthrough
	case 11:
		fallthrough
	case 12:
		tmbbox[numspechit-9] = addr
	case 13:
		crushchange = boolean(addr)
	case 14:
		nofit = boolean(addr)
	default:
		fprintf_ccgo(os.Stderr, "spechitOverrun: Warning: unable to emulatean overrun where numspechit=%d\n", numspechit)
		break
	}
}

var baseaddr int32

const INT_MAX11 = 2147483647

// State.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

//
// P_AproxDistance
// Gives an estimation of distance (not exact)
//

func p_AproxDistance(dx fixed_t, dy fixed_t) fixed_t {
	dx = xabs(dx)
	dy = xabs(dy)
	if dx < dy {
		return dx + dy - dx>>1
	}
	return dx + dy - dy>>1
}

// C documentation
//
//	//
//	// P_PointOnLineSide
//	// Returns 0 or 1
//	//
func p_PointOnLineSide(x fixed_t, y fixed_t, line *line_t) int32 {
	var dx, dy, left, right fixed_t
	if line.Fdx == 0 {
		if x <= line.Fv1.Fx {
			return boolint32(line.Fdy > 0)
		}
		return boolint32(line.Fdy < 0)
	}
	if line.Fdy == 0 {
		if y <= line.Fv1.Fy {
			return boolint32(line.Fdx < 0)
		}
		return boolint32(line.Fdx > 0)
	}
	dx = x - line.Fv1.Fx
	dy = y - line.Fv1.Fy
	left = fixedMul(line.Fdy>>FRACBITS, dx)
	right = fixedMul(dy, line.Fdx>>FRACBITS)
	if right < left {
		return 0
	} // front side
	return 1 // back side
}

// C documentation
//
//	//
//	// P_BoxOnLineSide
//	// Considers the line to be infinite
//	// Returns side 0 or 1, -1 if box crosses the line.
//	//
func p_BoxOnLineSide(tmbox *box_t, ld *line_t) int32 {
	var p1, p2 int32
	p1 = 0
	p2 = 0
	switch ld.Fslopetype {
	case st_HORIZONTAL:
		p1 = boolint32(tmbox[BOXTOP] > ld.Fv1.Fy)
		p2 = boolint32(tmbox[BOXBOTTOM] > ld.Fv1.Fy)
		if ld.Fdx < 0 {
			p1 ^= 1
			p2 ^= 1
		}
	case st_VERTICAL:
		p1 = boolint32(tmbox[BOXRIGHT] < ld.Fv1.Fx)
		p2 = boolint32(tmbox[BOXLEFT] < ld.Fv1.Fx)
		if ld.Fdy < 0 {
			p1 ^= 1
			p2 ^= 1
		}
	case st_POSITIVE:
		p1 = p_PointOnLineSide(tmbox[BOXLEFT], tmbox[BOXTOP], ld)
		p2 = p_PointOnLineSide(tmbox[BOXRIGHT], tmbox[BOXBOTTOM], ld)
	case st_NEGATIVE:
		p1 = p_PointOnLineSide(tmbox[BOXRIGHT], tmbox[BOXTOP], ld)
		p2 = p_PointOnLineSide(tmbox[BOXLEFT], tmbox[BOXBOTTOM], ld)
		break
	}
	if p1 == p2 {
		return p1
	}
	return -1
}

// C documentation
//
//	//
//	// P_PointOnDivlineSide
//	// Returns 0 or 1.
//	//
func p_PointOnDivlineSide(x fixed_t, y fixed_t, line *divline_t) int32 {
	var dx, dy, left, right fixed_t
	if line.Fdx == 0 {
		if x <= line.Fx {
			return boolint32(line.Fdy > 0)
		}
		return boolint32(line.Fdy < 0)
	}
	if line.Fdy == 0 {
		if y <= line.Fy {
			return boolint32(line.Fdx < 0)
		}
		return boolint32(line.Fdx > 0)
	}
	dx = x - line.Fx
	dy = y - line.Fy
	// try to quickly decide by looking at sign bits
	if uint32(line.Fdy^line.Fdx^dx^dy)&0x80000000 != 0 {
		if uint32(line.Fdy^dx)&0x80000000 != 0 {
			return 1
		} // (left is negative)
		return 0
	}
	left = fixedMul(line.Fdy>>8, dx>>8)
	right = fixedMul(dy>>8, line.Fdx>>8)
	if right < left {
		return 0
	} // front side
	return 1 // back side
}

// C documentation
//
//	//
//	// P_MakeDivline
//	//
func p_MakeDivline(li *line_t, dl *divline_t) {
	dl.Fx = li.Fv1.Fx
	dl.Fy = li.Fv1.Fy
	dl.Fdx = li.Fdx
	dl.Fdy = li.Fdy
}

// C documentation
//
//	//
//	// P_InterceptVector
//	// Returns the fractional intercept point
//	// along the first divline.
//	// This is only called by the addthings
//	// and addlines traversers.
//	//
func p_InterceptVector(v2 *divline_t, v1 *divline_t) fixed_t {
	var den, frac, num fixed_t
	den = fixedMul(v1.Fdy>>8, v2.Fdx) - fixedMul(v1.Fdx>>8, v2.Fdy)
	if den == 0 {
		return 0
	}
	//	i_Error ("p_InterceptVector: parallel");
	num = fixedMul((v1.Fx-v2.Fx)>>8, v1.Fdy) + fixedMul((v2.Fy-v1.Fy)>>8, v1.Fdx)
	frac = fixedDiv(num, den)
	return frac
}

func p_LineOpening(linedef *line_t) {
	var back, front *sector_t
	if linedef.Fsidenum[1] == -1 {
		// single sided line
		openrange = 0
		return
	}
	front = linedef.Ffrontsector
	back = linedef.Fbacksector
	if front.Fceilingheight < back.Fceilingheight {
		opentop = front.Fceilingheight
	} else {
		opentop = back.Fceilingheight
	}
	if front.Ffloorheight > back.Ffloorheight {
		openbottom = front.Ffloorheight
		lowfloor = back.Ffloorheight
	} else {
		openbottom = back.Ffloorheight
		lowfloor = front.Ffloorheight
	}
	openrange = opentop - openbottom
}

//
// THING POSITION SETTING
//

// C documentation
//
//	//
//	// P_UnsetThingPosition
//	// Unlinks a thing from block map and sectors.
//	// On each position change, BLOCKMAP and other
//	// lookups maintaining lists ot things inside
//	// these structures need to be updated.
//	//
func p_UnsetThingPosition(thing *mobj_t) {
	var blockx, blocky int32
	if thing.Fflags&mf_NOSECTOR == 0 {
		// inert things don't need to be in blockmap?
		// unlink from subsector
		if thing.Fsnext != nil {
			thing.Fsnext.Fsprev = thing.Fsprev
		}
		if thing.Fsprev != nil {
			thing.Fsprev.Fsnext = thing.Fsnext
		} else {
			thing.Fsubsector.Fsector.Fthinglist = thing.Fsnext
		}
	}
	if thing.Fflags&mf_NOBLOCKMAP == 0 {
		// inert things don't need to be in blockmap
		// unlink from block map
		if thing.Fbnext != nil {
			thing.Fbnext.Fbprev = thing.Fbprev
		}
		if thing.Fbprev != nil {
			thing.Fbprev.Fbnext = thing.Fbnext
		} else {
			blockx = (thing.Fx - bmaporgx) >> (FRACBITS + 7)
			blocky = (thing.Fy - bmaporgy) >> (FRACBITS + 7)
			if blockx >= 0 && blockx < bmapwidth && blocky >= 0 && blocky < bmapheight {
				blocklinks[blocky*bmapwidth+blockx] = thing.Fbnext
			}
		}
	}
}

// C documentation
//
//	//
//	// P_SetThingPosition
//	// Links a thing into both a block and a subsector
//	// based on it's x y.
//	// Sets thing->subsector properly
//	//
func p_SetThingPosition(thing *mobj_t) {
	var blockx, blocky int32
	var ss *subsector_t
	var sec *sector_t
	// link into subsector
	ss = r_PointInSubsector(thing.Fx, thing.Fy)
	thing.Fsubsector = ss
	if thing.Fflags&mf_NOSECTOR == 0 {
		// invisible things don't go into the sector links
		sec = ss.Fsector
		thing.Fsprev = nil
		thing.Fsnext = sec.Fthinglist
		if sec.Fthinglist != nil {
			sec.Fthinglist.Fsprev = thing
		}
		sec.Fthinglist = thing
	}
	// link into blockmap
	if thing.Fflags&mf_NOBLOCKMAP == 0 {
		// inert things don't need to be in blockmap
		blockx = (thing.Fx - bmaporgx) >> (FRACBITS + 7)
		blocky = (thing.Fy - bmaporgy) >> (FRACBITS + 7)
		if blockx >= 0 && blockx < bmapwidth && blocky >= 0 && blocky < bmapheight {
			link := &blocklinks[blocky*bmapwidth+blockx]
			thing.Fbprev = nil
			thing.Fbnext = *link
			if *link != nil {
				(*link).Fbprev = thing
			}
			*link = thing
		} else {
			// thing is off the map
			thing.Fbprev = nil
			thing.Fbnext = nil
		}
	}
}

//
// BLOCK MAP ITERATORS
// For each line/thing in the given mapblock,
// call the passed PIT_* function.
// If the function returns false,
// exit with false without checking anything else.
//

// C documentation
//
//	//
//	// P_BlockLinesIterator
//	// The validcount flags are used to avoid checking lines
//	// that are marked in multiple mapblocks,
//	// so increment validcount before the first call
//	// to p_BlockLinesIterator, then make one or more calls
//	// to it.
//	//
func p_BlockLinesIterator(x int32, y int32, func1 func(*line_t) boolean) boolean {
	var offset int32
	if x < 0 || y < 0 || x >= bmapwidth || y >= bmapheight {
		return 1
	}
	offset = y*bmapwidth + x
	offset = int32(blockmap[offset])
	for listpos := offset; blockmaplump[listpos] != -1; listpos++ {
		ld := &lines[blockmaplump[listpos]]
		if ld.Fvalidcount == validcount {
			continue
		} // line has already been checked
		ld.Fvalidcount = validcount
		if func1(ld) == 0 {
			return 0
		}
	}
	return 1 // everything was checked
}

// C documentation
//
//	//
//	// P_BlockThingsIterator
//	//
func p_BlockThingsIterator(x int32, y int32, func1 func(*mobj_t) boolean) boolean {
	if x < 0 || y < 0 || x >= bmapwidth || y >= bmapheight {
		return 1
	}
	for mobj := blocklinks[y*bmapwidth+x]; mobj != nil; mobj = mobj.Fbnext {
		if func1(mobj) == 0 {
			return 0
		}
	}
	return 1
}

// C documentation
//
//	//
//	// pit_AddLineIntercepts.
//	// Looks for lines in the given block
//	// that intercept the given trace
//	// to add to the intercepts list.
//	//
//	// A line is crossed if its endpoints
//	// are on opposite sides of the trace.
//	// Returns true if earlyout and a solid line hit.
//	//
func pit_AddLineIntercepts(ld *line_t) boolean {
	var frac fixed_t
	var s1, s2 int32
	// avoid precision problems with two routines
	if trace.Fdx > 1<<FRACBITS*16 || trace.Fdy > 1<<FRACBITS*16 || trace.Fdx < -(1<<FRACBITS)*16 || trace.Fdy < -(1<<FRACBITS)*16 {
		s1 = p_PointOnDivlineSide(ld.Fv1.Fx, ld.Fv1.Fy, &trace)
		s2 = p_PointOnDivlineSide(ld.Fv2.Fx, ld.Fv2.Fy, &trace)
	} else {
		s1 = p_PointOnLineSide(trace.Fx, trace.Fy, ld)
		s2 = p_PointOnLineSide(trace.Fx+trace.Fdx, trace.Fy+trace.Fdy, ld)
	}
	if s1 == s2 {
		return 1
	} // line isn't crossed
	// hit the line
	var divline divline_t
	p_MakeDivline(ld, &divline)
	frac = p_InterceptVector(&trace, &divline)
	if frac < 0 {
		return 1
	} // behind source
	// try to early out the check
	if earlyout != 0 && frac < 1<<FRACBITS && ld.Fbacksector == nil {
		return 0 // stop checking
	}
	intercepts[intercept_pos].Ffrac = frac
	intercepts[intercept_pos].Fisaline = 1
	intercepts[intercept_pos].Fd.Fthing = ld
	interceptsOverrun(intercept_pos, &intercepts[intercept_pos])
	intercept_pos++
	return 1 // continue
}

// C documentation
//
//	//
//	// PIT_AddThingIntercepts
//	//
func pit_AddThingIntercepts(thing *mobj_t) boolean {
	var divline divline_t
	var frac, x1, x2, y1, y2 fixed_t
	var s1, s2 int32
	var tracepositive boolean
	tracepositive = booluint32(trace.Fdx^trace.Fdy > 0)
	// check a corner to corner crossection for hit
	if tracepositive != 0 {
		x1 = thing.Fx - thing.Fradius
		y1 = thing.Fy + thing.Fradius
		x2 = thing.Fx + thing.Fradius
		y2 = thing.Fy - thing.Fradius
	} else {
		x1 = thing.Fx - thing.Fradius
		y1 = thing.Fy - thing.Fradius
		x2 = thing.Fx + thing.Fradius
		y2 = thing.Fy + thing.Fradius
	}
	s1 = p_PointOnDivlineSide(x1, y1, &trace)
	s2 = p_PointOnDivlineSide(x2, y2, &trace)
	if s1 == s2 {
		return 1
	} // line isn't crossed
	divline.Fx = x1
	divline.Fy = y1
	divline.Fdx = x2 - x1
	divline.Fdy = y2 - y1
	frac = p_InterceptVector(&trace, &divline)
	if frac < 0 {
		return 1
	} // behind source
	intercepts[intercept_pos].Ffrac = frac
	intercepts[intercept_pos].Fisaline = 0
	intercepts[intercept_pos].Fd.Fthing = thing
	interceptsOverrun(intercept_pos, &intercepts[intercept_pos])
	intercept_pos++
	return 1 // keep going
}

// C documentation
//
//	//
//	// P_TraverseIntercepts
//	// Returns true if the traverser function returns true
//	// for all lines.
//	//
func p_TraverseIntercepts(func1 func(*intercept_t) boolean, maxfrac fixed_t) boolean {
	var count, v1 int32
	var dist fixed_t
	var in *intercept_t
	count = intercept_pos
	for {
		v1 = count
		count--
		if v1 == 0 {
			break
		}
		dist = int32(INT_MAX11)
		for scan := int32(0); scan < intercept_pos; scan++ {
			if intercepts[scan].Ffrac < dist {
				dist = intercepts[scan].Ffrac
				in = &intercepts[scan]
			}
		}
		if dist > maxfrac {
			return 1
		} // checked everything in range
		if func1(in) == 0 {
			return 0
		} // don't bother going farther
		in.Ffrac = int32(INT_MAX11)
	}
	return 1 // everything was traversed
}

// Intercepts Overrun emulation, from PrBoom-plus.
// Thanks to Andrey Budko (entryway) for researching this and his
// implementation of Intercepts Overrun emulation in PrBoom-plus
// which this is based on.

type intercepts_overrun_t struct {
	Flen1        int32
	Faddr        uintptr
	Fint16_array boolean
}

// Intercepts memory table.  This is where various variables are located
// in memory in Vanilla Doom.  When the intercepts table overflows, we
// need to write to them.
//
// Almost all of the values to overwrite are 32-bit integers, except for
// playerstarts, which is effectively an array of 16-bit integers and
// must be treated differently.

var intercepts_overrun = [23]intercepts_overrun_t{
	0: {
		Flen1: 4,
	},
	1: {
		Flen1: 4,
	},
	2: {
		Flen1: 4,
	},
	3: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&lowfloor)),
	},
	4: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&openbottom)),
	},
	5: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&opentop)),
	},
	6: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&openrange)),
	},
	7: {
		Flen1: 4,
	},
	8: {
		Flen1: 120,
	},
	9: {
		Flen1: 8,
	},
	10: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&bulletslope)),
	},
	11: {
		Flen1: 4,
	},
	12: {
		Flen1: 4,
	},
	13: {
		Flen1: 4,
	},
	14: {
		Flen1:        40,
		Faddr:        uintptr(unsafe.Pointer(&playerstarts)),
		Fint16_array: 1,
	},
	15: {
		Flen1: 4,
	},
	16: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&bmapwidth)),
	},
	17: {
		Flen1: 4,
	},
	18: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&bmaporgx)),
	},
	19: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&bmaporgy)),
	},
	20: {
		Flen1: 4,
	},
	21: {
		Flen1: 4,
		Faddr: uintptr(unsafe.Pointer(&bmapheight)),
	},
	22: {},
}

// Overwrite a specific memory location with a value.

func interceptsMemoryOverrun(location int32, value int32) {
	var addr uintptr
	var i, index, offset int32
	i = 0
	offset = 0
	// Search down the array until we find the right entry
	for intercepts_overrun[i].Flen1 != 0 {
		if offset+intercepts_overrun[i].Flen1 > location {
			addr = intercepts_overrun[i].Faddr
			// Write the value to the memory location.
			// 16-bit and 32-bit values are written differently.
			if addr != 0 {
				if intercepts_overrun[i].Fint16_array != 0 {
					index = (location - offset) / 2
					*(*int16)(unsafe.Pointer(addr + uintptr(index)*2)) = int16(value & 0xffff)
					*(*int16)(unsafe.Pointer(addr + uintptr(index+1)*2)) = int16(value >> 16 & 0xffff)
				} else {
					index = (location - offset) / 4
					*(*int32)(unsafe.Pointer(addr + uintptr(index)*4)) = value
				}
			}
			break
		}
		offset += intercepts_overrun[i].Flen1
		i++
	}
}

// Emulate overruns of the intercepts[] array.

func interceptsOverrun(num_intercepts int32, intercept *intercept_t) {
	var location int32
	if num_intercepts <= MAXINTERCEPTS_ORIGINAL {
		// No overrun
		return
	}
	location = (num_intercepts - MAXINTERCEPTS_ORIGINAL - 1) * 12
	// Overwrite memory that is overwritten in Vanilla Doom, using
	// the values from the intercept structure.
	//
	// Note: the ->d.{thing,line} member should really have its
	// address translated into the correct address value for
	// Vanilla Doom.
	interceptsMemoryOverrun(location, intercept.Ffrac)
	interceptsMemoryOverrun(location+4, int32(intercept.Fisaline))
	interceptsMemoryOverrun(location+8, int32(*(*uintptr)(unsafe.Pointer(&intercept.Fd))))
}

// C documentation
//
//	//
//	// P_PathTraverse
//	// Traces a line from x1,y1 to x2,y2,
//	// calling the traverser function for each.
//	// Returns true if the traverser function returns true
//	// for all lines.
//	//
func p_PathTraverse(x1 fixed_t, y1 fixed_t, x2 fixed_t, y2 fixed_t, flags int32, trav func(*intercept_t) boolean) boolean {
	var mapx, mapxstep, mapy, mapystep int32
	var partial, xintercept, xstep, xt1, xt2, yintercept, ystep, yt1, yt2 fixed_t
	earlyout = uint32(flags & PT_EARLYOUT)
	validcount++
	intercept_pos = 0
	if (x1-bmaporgx)&(MAPBLOCKUNITS*(1<<FRACBITS)-1) == 0 {
		x1 += 1 << FRACBITS
	} // don't side exactly on a line
	if (y1-bmaporgy)&(MAPBLOCKUNITS*(1<<FRACBITS)-1) == 0 {
		y1 += 1 << FRACBITS
	} // don't side exactly on a line
	trace.Fx = x1
	trace.Fy = y1
	trace.Fdx = x2 - x1
	trace.Fdy = y2 - y1
	x1 -= bmaporgx
	y1 -= bmaporgy
	xt1 = x1 >> (FRACBITS + 7)
	yt1 = y1 >> (FRACBITS + 7)
	x2 -= bmaporgx
	y2 -= bmaporgy
	xt2 = x2 >> (FRACBITS + 7)
	yt2 = y2 >> (FRACBITS + 7)
	if xt2 > xt1 {
		mapxstep = 1
		partial = 1<<FRACBITS - x1>>(FRACBITS+7-FRACBITS)&(1<<FRACBITS-1)
		ystep = fixedDiv(y2-y1, xabs(x2-x1))
	} else {
		if xt2 < xt1 {
			mapxstep = -1
			partial = x1 >> (FRACBITS + 7 - FRACBITS) & (1<<FRACBITS - 1)
			ystep = fixedDiv(y2-y1, xabs(x2-x1))
		} else {
			mapxstep = 0
			partial = 1 << FRACBITS
			ystep = 256 * (1 << FRACBITS)
		}
	}
	yintercept = y1>>(FRACBITS+7-FRACBITS) + fixedMul(partial, ystep)
	if yt2 > yt1 {
		mapystep = 1
		partial = 1<<FRACBITS - y1>>(FRACBITS+7-FRACBITS)&(1<<FRACBITS-1)
		xstep = fixedDiv(x2-x1, xabs(y2-y1))
	} else {
		if yt2 < yt1 {
			mapystep = -1
			partial = y1 >> (FRACBITS + 7 - FRACBITS) & (1<<FRACBITS - 1)
			xstep = fixedDiv(x2-x1, xabs(y2-y1))
		} else {
			mapystep = 0
			partial = 1 << FRACBITS
			xstep = 256 * (1 << FRACBITS)
		}
	}
	xintercept = x1>>(FRACBITS+7-FRACBITS) + fixedMul(partial, xstep)
	// Step through map blocks.
	// Count is present to prevent a round off error
	// from skipping the break.
	mapx = xt1
	mapy = yt1
	for count := 0; count < 64; count++ {
		if flags&PT_ADDLINES != 0 {
			if p_BlockLinesIterator(mapx, mapy, pit_AddLineIntercepts) == 0 {
				return 0
			} // early out
		}
		if flags&PT_ADDTHINGS != 0 {
			if p_BlockThingsIterator(mapx, mapy, pit_AddThingIntercepts) == 0 {
				return 0
			} // early out
		}
		if mapx == xt2 && mapy == yt2 {
			break
		}
		if yintercept>>FRACBITS == mapy {
			yintercept += ystep
			mapx += mapxstep
		} else {
			if xintercept>>FRACBITS == mapx {
				xintercept += xstep
				mapy += mapystep
			}
		}
	}
	// go through the sorted list
	return p_TraverseIntercepts(trav, 1<<FRACBITS)
}

const ANG453 = 536870912
const FRICTION = 59392
const STOPSPEED = 4096

func p_SetMobjState(mobj *mobj_t, state statenum_t) boolean {
	for cond := true; cond; cond = mobj.Ftics == 0 {
		if state == s_NULL {
			mobj.Fstate = nil
			p_RemoveMobj(mobj)
			return 0
		}
		st := &states[state]
		mobj.Fstate = st
		mobj.Ftics = st.Ftics
		mobj.Fsprite = st.Fsprite
		mobj.Fframe = st.Fframe
		// Modified handling.
		// Call action functions when the state is set
		if st.Faction != nil {
			st.Faction(mobj, nil)
		}
		state = st.Fnextstate
	}
	return 1
}

// C documentation
//
//	//
//	// P_ExplodeMissile
//	//
func p_ExplodeMissile(mo *mobj_t) {
	var v1, v2 fixed_t
	v2 = 0
	mo.Fmomz = v2
	v1 = v2
	mo.Fmomy = v1
	mo.Fmomx = v1
	p_SetMobjState(mo, mobjinfo[mo.Ftype1].Fdeathstate)
	mo.Ftics -= p_Random() & 3
	if mo.Ftics < 1 {
		mo.Ftics = 1
	}
	mo.Fflags &= ^mf_MISSILE
	if mo.Finfo.Fdeathsound != 0 {
		s_StartSound(&mo.degenmobj_t, mo.Finfo.Fdeathsound)
	}
}

//
// P_XYMovement
//

func p_XYMovement(mo *mobj_t) {
	var player *player_t
	var ptryx, ptryy, xmove, ymove, v1, v2, v3, v4, v5 fixed_t
	if mo.Fmomx == 0 && mo.Fmomy == 0 {
		if mo.Fflags&mf_SKULLFLY != 0 {
			// the skull slammed into something
			mo.Fflags &= ^mf_SKULLFLY
			v2 = 0
			mo.Fmomz = v2
			v1 = v2
			mo.Fmomy = v1
			mo.Fmomx = v1
			p_SetMobjState(mo, mo.Finfo.Fspawnstate)
		}
		return
	}
	player = mo.Fplayer
	if mo.Fmomx > 30*(1<<FRACBITS) {
		mo.Fmomx = 30 * (1 << FRACBITS)
	} else {
		if mo.Fmomx < -(30 * (1 << FRACBITS)) {
			mo.Fmomx = -(30 * (1 << FRACBITS))
		}
	}
	if mo.Fmomy > 30*(1<<FRACBITS) {
		mo.Fmomy = 30 * (1 << FRACBITS)
	} else {
		if mo.Fmomy < -(30 * (1 << FRACBITS)) {
			mo.Fmomy = -(30 * (1 << FRACBITS))
		}
	}
	xmove = mo.Fmomx
	ymove = mo.Fmomy
	for cond := true; cond; cond = xmove != 0 || ymove != 0 {
		if xmove > 30*(1<<FRACBITS)/2 || ymove > 30*(1<<FRACBITS)/2 {
			ptryx = mo.Fx + xmove/2
			ptryy = mo.Fy + ymove/2
			xmove >>= 1
			ymove >>= 1
		} else {
			ptryx = mo.Fx + xmove
			ptryy = mo.Fy + ymove
			v3 = 0
			ymove = v3
			xmove = v3
		}
		if p_TryMove(mo, ptryx, ptryy) == 0 {
			// blocked move
			if mo.Fplayer != nil {
				// try to slide along it
				p_SlideMove(mo)
			} else {
				if mo.Fflags&mf_MISSILE != 0 {
					// explode a missile
					if ceilingline != nil && ceilingline.Fbacksector != nil && int32(ceilingline.Fbacksector.Fceilingpic) == skyflatnum {
						// Hack to prevent missiles exploding
						// against the sky.
						// Does not handle sky floors.
						p_RemoveMobj(mo)
						return
					}
					p_ExplodeMissile(mo)
				} else {
					v4 = 0
					mo.Fmomy = v4
					mo.Fmomx = v4
				}
			}
		}
	}
	// slow down
	if player != nil && player.Fcheats&CF_NOMOMENTUM != 0 {
		// debug option for no sliding at all
		v5 = 0
		mo.Fmomy = v5
		mo.Fmomx = v5
		return
	}
	if mo.Fflags&(mf_MISSILE|mf_SKULLFLY) != 0 {
		return
	} // no friction for missiles ever
	if mo.Fz > mo.Ffloorz {
		return
	} // no friction when airborne
	if mo.Fflags&mf_CORPSE != 0 {
		// do not stop sliding
		//  if halfway off a step with some momentum
		if mo.Fmomx > 1<<FRACBITS/4 || mo.Fmomx < -(1<<FRACBITS)/4 || mo.Fmomy > 1<<FRACBITS/4 || mo.Fmomy < -(1<<FRACBITS)/4 {
			if mo.Ffloorz != mo.Fsubsector.Fsector.Ffloorheight {
				return
			}
		}
	}
	if mo.Fmomx > -STOPSPEED && mo.Fmomx < STOPSPEED && mo.Fmomy > -STOPSPEED && mo.Fmomy < STOPSPEED && (player == nil || int32(player.Fcmd.Fforwardmove) == 0 && int32(player.Fcmd.Fsidemove) == 0) {
		// if in a walking frame, stop moving
		if player != nil && stateIndex(player.Fmo.Fstate)-s_PLAY_RUN1 < 4 {
			p_SetMobjState(player.Fmo, s_PLAY)
		}
		mo.Fmomx = 0
		mo.Fmomy = 0
	} else {
		mo.Fmomx = fixedMul(mo.Fmomx, FRICTION)
		mo.Fmomy = fixedMul(mo.Fmomy, FRICTION)
	}
}

// C documentation
//
//	//
//	// P_ZMovement
//	//
func p_ZMovement(mo *mobj_t) {
	var correct_lost_soul_bounce int32
	var delta, dist fixed_t
	// check for smooth step up
	if mo.Fplayer != nil && mo.Fz < mo.Ffloorz {
		mo.Fplayer.Fviewheight -= mo.Ffloorz - mo.Fz
		mo.Fplayer.Fdeltaviewheight = (41*(1<<FRACBITS) - mo.Fplayer.Fviewheight) >> 3
	}
	// adjust height
	mo.Fz += mo.Fmomz
	if mo.Fflags&mf_FLOAT != 0 && mo.Ftarget != nil {
		// float down towards target if too close
		if mo.Fflags&mf_SKULLFLY == 0 && mo.Fflags&mf_INFLOAT == 0 {
			dist = p_AproxDistance(mo.Fx-mo.Ftarget.Fx, mo.Fy-mo.Ftarget.Fy)
			delta = mo.Ftarget.Fz + mo.Fheight>>1 - mo.Fz
			if delta < 0 && dist < -(delta*3) {
				mo.Fz -= 1 << FRACBITS * 4
			} else {
				if delta > 0 && dist < delta*3 {
					mo.Fz += 1 << FRACBITS * 4
				}
			}
		}
	}
	// clip movement
	if mo.Fz <= mo.Ffloorz {
		// hit the floor
		// Note (id):
		//  somebody left this after the setting momz to 0,
		//  kinda useless there.
		//
		// cph - This was the a bug in the linuxdoom-1.10 source which
		//  caused it not to sync Doom 2 v1.9 demos. Someone
		//  added the above comment and moved up the following code. So
		//  demos would desync in close lost soul fights.
		// Note that this only applies to original Doom 1 or Doom2 demos - not
		//  Final Doom and Ultimate Doom.  So we test demo_compatibility *and*
		//  gamemission. (Note we assume that Doom1 is always Ult Doom, which
		//  seems to hold for most published demos.)
		//
		//  fraggle - cph got the logic here slightly wrong.  There are three
		//  versions of Doom 1.9:
		//
		//  * The version used in registered doom 1.9 + doom2 - no bounce
		//  * The version used in ultimate doom - has bounce
		//  * The version used in final doom - has bounce
		//
		// So we need to check that this is either retail or commercial
		// (but not doom2)
		correct_lost_soul_bounce = boolint32(gameversion >= exe_ultimate)
		if correct_lost_soul_bounce != 0 && mo.Fflags&mf_SKULLFLY != 0 {
			// the skull slammed into something
			mo.Fmomz = -mo.Fmomz
		}
		if mo.Fmomz < 0 {
			if mo.Fplayer != nil && mo.Fmomz < -(1<<FRACBITS)*8 {
				// Squat down.
				// Decrease viewheight for a moment
				// after hitting the ground (hard),
				// and utter appropriate sound.
				mo.Fplayer.Fdeltaviewheight = mo.Fmomz >> 3
				s_StartSound(&mo.degenmobj_t, int32(sfx_oof))
			}
			mo.Fmomz = 0
		}
		mo.Fz = mo.Ffloorz
		// cph 2001/05/26 -
		// See lost soul bouncing comment above. We need this here for bug
		// compatibility with original Doom2 v1.9 - if a soul is charging and
		// hit by a raising floor this incorrectly reverses its Y momentum.
		//
		if correct_lost_soul_bounce == 0 && mo.Fflags&mf_SKULLFLY != 0 {
			mo.Fmomz = -mo.Fmomz
		}
		if mo.Fflags&mf_MISSILE != 0 && mo.Fflags&mf_NOCLIP == 0 {
			p_ExplodeMissile(mo)
			return
		}
	} else {
		if mo.Fflags&mf_NOGRAVITY == 0 {
			if mo.Fmomz == 0 {
				mo.Fmomz = -(1 << FRACBITS) * 2
			} else {
				mo.Fmomz -= 1 << FRACBITS
			}
		}
	}
	if mo.Fz+mo.Fheight > mo.Fceilingz {
		// hit the ceiling
		if mo.Fmomz > 0 {
			mo.Fmomz = 0
		}
		mo.Fz = mo.Fceilingz - mo.Fheight
		if mo.Fflags&mf_SKULLFLY != 0 {
			// the skull slammed into something
			mo.Fmomz = -mo.Fmomz
		}
		if mo.Fflags&mf_MISSILE != 0 && mo.Fflags&mf_NOCLIP == 0 {
			p_ExplodeMissile(mo)
			return
		}
	}
}

// C documentation
//
//	//
//	// P_NightmareRespawn
//	//
func p_NightmareRespawn(mobj *mobj_t) {
	var mo *mobj_t
	var mthing *mapthing_t
	var ss *subsector_t
	var x, y, z fixed_t
	x = int32(mobj.Fspawnpoint.Fx) << FRACBITS
	y = int32(mobj.Fspawnpoint.Fy) << FRACBITS
	// somthing is occupying it's position?
	if p_CheckPosition(mobj, x, y) == 0 {
		return
	} // no respwan
	// spawn a teleport fog at old spot
	// because of removal of the body?
	mo = p_SpawnMobj(mobj.Fx, mobj.Fy, mobj.Fsubsector.Fsector.Ffloorheight, mt_TFOG)
	// initiate teleport sound
	s_StartSound(&mo.degenmobj_t, int32(sfx_telept))
	// spawn a teleport fog at the new spot
	ss = r_PointInSubsector(x, y)
	mo = p_SpawnMobj(x, y, ss.Fsector.Ffloorheight, mt_TFOG)
	s_StartSound(&mo.degenmobj_t, int32(sfx_telept))
	// spawn the new monster
	mthing = &mobj.Fspawnpoint
	// spawn it
	if mobj.Finfo.Fflags&mf_SPAWNCEILING != 0 {
		z = int32(INT_MAX11)
	} else {
		z = -1 - 0x7fffffff
	}
	// inherit attributes from deceased one
	mo = p_SpawnMobj(x, y, z, mobj.Ftype1)
	mo.Fspawnpoint = mobj.Fspawnpoint
	mo.Fangle = uint32(int32(ANG453) * (int32(mthing.Fangle) / 45))
	if int32(mthing.Foptions)&MTF_AMBUSH != 0 {
		mo.Fflags |= mf_AMBUSH
	}
	mo.Freactiontime = 18
	// remove the old monster,
	p_RemoveMobj(mobj)
}

// C documentation
//
//	//
//	// P_MobjThinker
//	//
func (mobj *mobj_t) ThinkerFunc() {
	p_MobjThinker(mobj)
}
func p_MobjThinker(mobj *mobj_t) {
	// momentum movement
	if mobj.Fmomx != 0 || mobj.Fmomy != 0 || mobj.Fflags&mf_SKULLFLY != 0 {
		p_XYMovement(mobj)
		// FIXME: decent NOP/NULL/Nil function pointer please.
		if mobj.Fthinker.Ffunction == nil {
			return
		} // mobj was removed
	}
	if mobj.Fz != mobj.Ffloorz || mobj.Fmomz != 0 {
		p_ZMovement(mobj)
		// FIXME: decent NOP/NULL/Nil function pointer please.
		if mobj.Fthinker.Ffunction == nil {
			return
		} // mobj was removed
	}
	// cycle through states,
	// calling action functions at transitions
	if mobj.Ftics != -1 {
		mobj.Ftics--
		// you can cycle through multiple states in a tic
		if mobj.Ftics == 0 {
			if p_SetMobjState(mobj, mobj.Fstate.Fnextstate) == 0 {
				return
			}
		} // freed itself
	} else {
		// check for nightmare respawn
		if mobj.Fflags&mf_COUNTKILL == 0 {
			return
		}
		if respawnmonsters == 0 {
			return
		}
		mobj.Fmovecount++
		if mobj.Fmovecount < 12*TICRATE {
			return
		}
		if leveltime&int32(31) != 0 {
			return
		}
		if p_Random() > 4 {
			return
		}
		p_NightmareRespawn(mobj)
	}
}

// C documentation
//
//	//
//	// P_SpawnMobj
//	//
func p_SpawnMobj(x fixed_t, y fixed_t, z fixed_t, type1 mobjtype_t) (r *mobj_t) {
	var info *mobjinfo_t
	mobj := &mobj_t{}
	info = &mobjinfo[type1]
	mobj.Ftype1 = type1
	mobj.Finfo = info
	mobj.Fx = x
	mobj.Fy = y
	mobj.Fradius = info.Fradius
	mobj.Fheight = info.Fheight
	mobj.Fflags = info.Fflags
	mobj.Fhealth = info.Fspawnhealth
	if gameskill != sk_nightmare {
		mobj.Freactiontime = info.Freactiontime
	}
	mobj.Flastlook = p_Random() % MAXPLAYERS
	// do not set the state with p_SetMobjState,
	// because action routines can not be called yet
	st := &states[info.Fspawnstate]
	mobj.Fstate = st
	mobj.Ftics = st.Ftics
	mobj.Fsprite = st.Fsprite
	mobj.Fframe = st.Fframe
	// set subsector and/or block links
	p_SetThingPosition(mobj)
	mobj.Ffloorz = mobj.Fsubsector.Fsector.Ffloorheight
	mobj.Fceilingz = mobj.Fsubsector.Fsector.Fceilingheight
	if z == -1-0x7fffffff {
		mobj.Fz = mobj.Ffloorz
	} else {
		if z == int32(INT_MAX11) {
			mobj.Fz = mobj.Fceilingz - mobj.Finfo.Fheight
		} else {
			mobj.Fz = z
		}
	}
	mobj.Fthinker.Ffunction = mobj
	p_AddThinker(&mobj.Fthinker)
	return mobj
}

func p_RemoveMobj(mobj *mobj_t) {
	if mobj.Fflags&mf_SPECIAL != 0 && mobj.Fflags&mf_DROPPED == 0 && mobj.Ftype1 != mt_INV && mobj.Ftype1 != mt_INS {
		itemrespawnque[iquehead] = mobj.Fspawnpoint
		itemrespawntime[iquehead] = leveltime
		iquehead = (iquehead + 1) & (ITEMQUESIZE - 1)
		// lose one off the end?
		if iquehead == iquetail {
			iquetail = (iquetail + 1) & (ITEMQUESIZE - 1)
		}
	}
	// unlink from sector and block lists
	p_UnsetThingPosition(mobj)
	// stop any playing sound
	s_StopSound(&mobj.degenmobj_t)
	// free block
	p_RemoveThinker(&mobj.Fthinker)
}

// C documentation
//
//	//
//	// P_RespawnSpecials
//	//
func p_RespawnSpecials() {
	var i int32
	var mo *mobj_t
	var mthing *mapthing_t
	var ss *subsector_t
	var x, y, z fixed_t
	// only respawn items in deathmatch
	if deathmatch != 2 {
		return
	} //
	// nothing left to respawn?
	if iquehead == iquetail {
		return
	}
	// wait at least 30 seconds
	if leveltime-itemrespawntime[iquetail] < 30*TICRATE {
		return
	}
	mthing = &itemrespawnque[iquetail]
	x = int32(mthing.Fx) << FRACBITS
	y = int32(mthing.Fy) << FRACBITS
	// spawn a teleport fog at the new spot
	ss = r_PointInSubsector(x, y)
	mo = p_SpawnMobj(x, y, ss.Fsector.Ffloorheight, mt_IFOG)
	s_StartSound(&mo.degenmobj_t, int32(sfx_itmbk))
	// find which type to spawn
	for i := 0; i < NUMMOBJTYPES; i++ {
		if int32(mthing.Ftype1) == mobjinfo[i].Fdoomednum {
			break
		}
	}
	// spawn it
	if mobjinfo[i].Fflags&mf_SPAWNCEILING != 0 {
		z = int32(INT_MAX11)
	} else {
		z = -1 - 0x7fffffff
	}
	mo = p_SpawnMobj(x, y, z, i)
	mo.Fspawnpoint = *mthing
	mo.Fangle = uint32(int32(ANG453) * (int32(mthing.Fangle) / 45))
	// pull it from the que
	iquetail = (iquetail + 1) & (ITEMQUESIZE - 1)
}

// C documentation
//
//	//
//	// P_SpawnPlayer
//	// Called when a player is spawned on the level.
//	// Most of the player structure stays unchanged
//	//  between levels.
//	//
func p_SpawnPlayer(mthing *mapthing_t) {
	var mobj *mobj_t
	var p *player_t
	var x, y, z fixed_t
	if int32(mthing.Ftype1) == 0 {
		return
	}
	// not playing?
	if playeringame[int32(mthing.Ftype1)-1] == 0 {
		return
	}
	p = &players[mthing.Ftype1-1]
	if p.Fplayerstate == Pst_REBORN {
		g_PlayerReborn(int32(mthing.Ftype1) - 1)
	}
	x = int32(mthing.Fx) << FRACBITS
	y = int32(mthing.Fy) << FRACBITS
	z = -1 - 0x7fffffff
	mobj = p_SpawnMobj(x, y, z, mt_PLAYER)
	// set color translations for player sprites
	if int32(mthing.Ftype1) > 1 {
		mobj.Fflags |= (int32(mthing.Ftype1) - 1) << mf_TRANSSHIFT
	}
	mobj.Fangle = uint32(int32(ANG453) * (int32(mthing.Fangle) / 45))
	mobj.Fplayer = p
	mobj.Fhealth = p.Fhealth
	p.Fmo = mobj
	p.Fplayerstate = Pst_LIVE
	p.Frefire = 0
	p.Fmessage = ""
	p.Fdamagecount = 0
	p.Fbonuscount = 0
	p.Fextralight = 0
	p.Ffixedcolormap = 0
	p.Fviewheight = 41 * (1 << FRACBITS)
	// setup gun psprite
	p_SetupPsprites(p)
	// give all cards in death match mode
	if deathmatch != 0 {
		for i := 0; i < NUMCARDS; i++ {
			p.Fcards[i] = 1
		}
	}
	if int32(mthing.Ftype1)-1 == consoleplayer {
		// wake up the status bar
		st_Start()
		// wake up the heads up text
		hu_Start()
	}
}

// C documentation
//
//	//
//	// P_SpawnMapThing
//	// The fields of the mapthing should
//	// already be in host byte order.
//	//
func p_SpawnMapThing(mthing *mapthing_t) {
	var bit, i int32
	var mobj *mobj_t
	var x, y, z fixed_t
	// count deathmatch start positions
	if int32(mthing.Ftype1) == 11 {
		if deathmatch_pos < len(deathmatchstarts) {
			deathmatchstarts[deathmatch_pos] = *mthing
			deathmatch_pos++
		}
		return
	}
	if int32(mthing.Ftype1) <= 0 {
		// Thing type 0 is actually "player -1 start".
		// For some reason, Vanilla Doom accepts/ignores this.
		return
	}
	// check for players specially
	if int32(mthing.Ftype1) <= 4 {
		// save spots for respawning in network games
		playerstarts[int32(mthing.Ftype1)-1] = *mthing
		if deathmatch == 0 {
			p_SpawnPlayer(mthing)
		}
		return
	}
	// check for apropriate skill level
	if netgame == 0 && int32(mthing.Foptions)&int32(16) != 0 {
		return
	}
	if gameskill == sk_baby {
		bit = 1
	} else {
		if gameskill == sk_nightmare {
			bit = 4
		} else {
			bit = 1 << (gameskill - 1)
		}
	}
	if int32(mthing.Foptions)&bit == 0 {
		return
	}
	// find which type to spawn
	for i = 0; i < NUMMOBJTYPES; i++ {
		if int32(mthing.Ftype1) == mobjinfo[i].Fdoomednum {
			break
		}
	}
	if i == NUMMOBJTYPES {
		i_Error("p_SpawnMapThing: Unknown type %d at (%d, %d)", int32(mthing.Ftype1), int32(mthing.Fx), int32(mthing.Fy))
	}
	// don't spawn keycards and players in deathmatch
	if deathmatch != 0 && mobjinfo[i].Fflags&mf_NOTDMATCH != 0 {
		return
	}
	// don't spawn any monsters if -nomonsters
	if nomonsters != 0 && (i == mt_SKULL || mobjinfo[i].Fflags&mf_COUNTKILL != 0) {
		return
	}
	// spawn it
	x = int32(mthing.Fx) << FRACBITS
	y = int32(mthing.Fy) << FRACBITS
	if mobjinfo[i].Fflags&mf_SPAWNCEILING != 0 {
		z = int32(INT_MAX11)
	} else {
		z = -1 - 0x7fffffff
	}
	mobj = p_SpawnMobj(x, y, z, i)
	mobj.Fspawnpoint = *mthing
	if mobj.Ftics > 0 {
		mobj.Ftics = 1 + p_Random()%mobj.Ftics
	}
	if mobj.Fflags&mf_COUNTKILL != 0 {
		totalkills++
	}
	if mobj.Fflags&mf_COUNTITEM != 0 {
		totalitems++
	}
	mobj.Fangle = uint32(int32(ANG453) * (int32(mthing.Fangle) / 45))
	if int32(mthing.Foptions)&MTF_AMBUSH != 0 {
		mobj.Fflags |= mf_AMBUSH
	}
}

func p_SpawnPuff(x fixed_t, y fixed_t, z fixed_t) {
	var th *mobj_t
	z += (p_Random() - p_Random()) << 10
	th = p_SpawnMobj(x, y, z, mt_PUFF)
	th.Fmomz = 1 << FRACBITS
	th.Ftics -= p_Random() & 3
	if th.Ftics < 1 {
		th.Ftics = 1
	}
	// don't make punches spark on the wall
	if attackrange == 64*(1<<FRACBITS) {
		p_SetMobjState(th, s_PUFF3)
	}
}

// C documentation
//
//	//
//	// P_SpawnBlood
//	//
func p_SpawnBlood(x fixed_t, y fixed_t, z fixed_t, damage int32) {
	var th *mobj_t
	z += (p_Random() - p_Random()) << 10
	th = p_SpawnMobj(x, y, z, mt_BLOOD)
	th.Fmomz = 1 << FRACBITS * 2
	th.Ftics -= p_Random() & 3
	if th.Ftics < 1 {
		th.Ftics = 1
	}
	if damage <= 12 && damage >= 9 {
		p_SetMobjState(th, s_BLOOD2)
	} else {
		if damage < 9 {
			p_SetMobjState(th, s_BLOOD3)
		}
	}
}

// C documentation
//
//	//
//	// P_CheckMissileSpawn
//	// Moves the missile forward a bit
//	//  and possibly explodes it right there.
//	//
func p_CheckMissileSpawn(th *mobj_t) {
	th.Ftics -= p_Random() & 3
	if th.Ftics < 1 {
		th.Ftics = 1
	}
	// move a little forward so an angle can
	// be computed if it immediately explodes
	th.Fx += th.Fmomx >> 1
	th.Fy += th.Fmomy >> 1
	th.Fz += th.Fmomz >> 1
	if p_TryMove(th, th.Fx, th.Fy) == 0 {
		p_ExplodeMissile(th)
	}
}

// Certain functions assume that a mobj_t pointer is non-NULL,
// causing a crash in some situations where it is NULL.  Vanilla
// Doom did not crash because of the lack of proper memory
// protection. This function substitutes NULL pointers for
// pointers to a dummy mobj, to avoid a crash.

func p_SubstNullMobj(mobj *mobj_t) *mobj_t {
	if mobj == nil {
		dummy_mobj.Fx = 0
		dummy_mobj.Fy = 0
		dummy_mobj.Fz = 0
		dummy_mobj.Fflags = 0
		mobj = &dummy_mobj
	}
	return mobj
}

var dummy_mobj mobj_t

// C documentation
//
//	//
//	// P_SpawnMissile
//	//
func p_SpawnMissile(source *mobj_t, dest *mobj_t, type1 mobjtype_t) (r *mobj_t) {
	var an angle_t
	var dist int32
	var th *mobj_t
	th = p_SpawnMobj(source.Fx, source.Fy, source.Fz+4*8*(1<<FRACBITS), type1)
	if th.Finfo.Fseesound != 0 {
		s_StartSound(&th.degenmobj_t, th.Finfo.Fseesound)
	}
	th.Ftarget = source // where it came from
	an = r_PointToAngle2(source.Fx, source.Fy, dest.Fx, dest.Fy)
	// fuzzy player
	if dest.Fflags&mf_SHADOW != 0 {
		an += uint32((p_Random() - p_Random()) << 20)
	}
	th.Fangle = an
	an >>= ANGLETOFINESHIFT
	th.Fmomx = fixedMul(th.Finfo.Fspeed, finecosine[an])
	th.Fmomy = fixedMul(th.Finfo.Fspeed, finesine[an])
	dist = p_AproxDistance(dest.Fx-source.Fx, dest.Fy-source.Fy)
	dist = dist / th.Finfo.Fspeed
	if dist < 1 {
		dist = 1
	}
	th.Fmomz = (dest.Fz - source.Fz) / dist
	p_CheckMissileSpawn(th)
	return th
}

// C documentation
//
//	//
//	// P_SpawnPlayerMissile
//	// Tries to aim at a nearby monster
//	//
func p_SpawnPlayerMissile(source *mobj_t, type1 mobjtype_t) {
	var an angle_t
	var slope, x, y, z fixed_t
	var th *mobj_t
	// see which target is to be aimed at
	an = source.Fangle
	slope = p_AimLineAttack(source, an, 16*64*(1<<FRACBITS))
	if linetarget == nil {
		an += uint32(1 << 26)
		slope = p_AimLineAttack(source, an, 16*64*(1<<FRACBITS))
		if linetarget == nil {
			an -= uint32(2 << 26)
			slope = p_AimLineAttack(source, an, 16*64*(1<<FRACBITS))
		}
		if linetarget == nil {
			an = source.Fangle
			slope = 0
		}
	}
	x = source.Fx
	y = source.Fy
	z = source.Fz + 4*8*(1<<FRACBITS)
	th = p_SpawnMobj(x, y, z, type1)
	if th.Finfo.Fseesound != 0 {
		s_StartSound(&th.degenmobj_t, th.Finfo.Fseesound)
	}
	th.Ftarget = source
	th.Fangle = an
	th.Fmomx = fixedMul(th.Finfo.Fspeed, finecosine[an>>ANGLETOFINESHIFT])
	th.Fmomy = fixedMul(th.Finfo.Fspeed, finesine[an>>ANGLETOFINESHIFT])
	th.Fmomz = fixedMul(th.Finfo.Fspeed, slope)
	p_CheckMissileSpawn(th)
}

// C documentation
//
//	//
//	// Move a plat up and down
//	//
func (plat *plat_t) ThinkerFunc() {
	t_PlatRaise(plat)
}
func t_PlatRaise(plat *plat_t) {
	var res result_e
	switch plat.Fstatus {
	case int32(up):
		res = t_MovePlane(plat.Fsector, plat.Fspeed, plat.Fhigh, plat.Fcrush, 0, 1)
		if plat.Ftype1 == int32(raiseAndChange) || plat.Ftype1 == int32(raiseToNearestAndChange) {
			if leveltime&7 == 0 {
				s_StartSound(&plat.Fsector.Fsoundorg, int32(sfx_stnmov))
			}
		}
		if res == int32(crushed) && plat.Fcrush == 0 {
			plat.Fcount = plat.Fwait
			plat.Fstatus = int32(down)
			s_StartSound(&plat.Fsector.Fsoundorg, int32(sfx_pstart))
		} else {
			if res == int32(pastdest) {
				plat.Fcount = plat.Fwait
				plat.Fstatus = int32(waiting)
				s_StartSound(&plat.Fsector.Fsoundorg, int32(sfx_pstop))
				switch plat.Ftype1 {
				case int32(blazeDWUS):
					fallthrough
				case int32(downWaitUpStay):
					p_RemoveActivePlat(plat)
				case int32(raiseAndChange):
					fallthrough
				case int32(raiseToNearestAndChange):
					p_RemoveActivePlat(plat)
				default:
					break
				}
			}
		}
	case int32(down):
		res = t_MovePlane(plat.Fsector, plat.Fspeed, plat.Flow, 0, 0, -1)
		if res == int32(pastdest) {
			plat.Fcount = plat.Fwait
			plat.Fstatus = int32(waiting)
			s_StartSound(&plat.Fsector.Fsoundorg, int32(sfx_pstop))
		}
	case int32(waiting):
		plat.Fcount--
		if plat.Fcount == 0 {
			if plat.Fsector.Ffloorheight == plat.Flow {
				plat.Fstatus = int32(up)
			} else {
				plat.Fstatus = int32(down)
			}
			s_StartSound(&plat.Fsector.Fsoundorg, int32(sfx_pstart))
		}
		fallthrough
	case int32(in_stasis):
		break
	}
}

// C documentation
//
//	//
//	// Do Platforms
//	//  "amount" is only used for SOME platforms.
//	//
func ev_DoPlat(line *line_t, type1 plattype_e, amount int32) int32 {
	var sec *sector_t
	var rtn, secnum, v1 int32
	secnum = -1
	rtn = 0
	//	Activate all <type> plats that are in_stasis
	switch type1 {
	case int32(perpetualRaise):
		p_ActivateInStasis(int32(line.Ftag))
	default:
		break
	}
	for {
		v1 = p_FindSectorFromLineTag(line, secnum)
		secnum = v1
		if !(v1 >= 0) {
			break
		}
		sec = &sectors[secnum]
		if sec.Fspecialdata != nil {
			continue
		}
		// Find lowest & highest floors around sector
		rtn = 1
		platP := &plat_t{}
		p_AddThinker(&platP.Fthinker)
		platP.Ftype1 = type1
		platP.Fsector = sec
		platP.Fsector.Fspecialdata = platP
		platP.Fthinker.Ffunction = platP
		platP.Fcrush = 0
		platP.Ftag = int32(line.Ftag)
		switch type1 {
		case int32(raiseToNearestAndChange):
			platP.Fspeed = 1 << FRACBITS / 2
			sec.Ffloorpic = sides[line.Fsidenum[0]].Fsector.Ffloorpic
			platP.Fhigh = p_FindNextHighestFloor(sec, sec.Ffloorheight)
			platP.Fwait = 0
			platP.Fstatus = int32(up)
			// NO MORE DAMAGE, IF APPLICABLE
			sec.Fspecial = 0
			s_StartSound(&sec.Fsoundorg, int32(sfx_stnmov))
		case int32(raiseAndChange):
			platP.Fspeed = 1 << FRACBITS / 2
			sec.Ffloorpic = sides[line.Fsidenum[0]].Fsector.Ffloorpic
			platP.Fhigh = sec.Ffloorheight + amount*(1<<FRACBITS)
			platP.Fwait = 0
			platP.Fstatus = int32(up)
			s_StartSound(&sec.Fsoundorg, int32(sfx_stnmov))
		case int32(downWaitUpStay):
			platP.Fspeed = 1 << FRACBITS * 4
			platP.Flow = p_FindLowestFloorSurrounding(sec)
			if platP.Flow > sec.Ffloorheight {
				platP.Flow = sec.Ffloorheight
			}
			platP.Fhigh = sec.Ffloorheight
			platP.Fwait = TICRATE * PLATWAIT
			platP.Fstatus = int32(down)
			s_StartSound(&sec.Fsoundorg, int32(sfx_pstart))
		case int32(blazeDWUS):
			platP.Fspeed = 1 << FRACBITS * 8
			platP.Flow = p_FindLowestFloorSurrounding(sec)
			if platP.Flow > sec.Ffloorheight {
				platP.Flow = sec.Ffloorheight
			}
			platP.Fhigh = sec.Ffloorheight
			platP.Fwait = TICRATE * PLATWAIT
			platP.Fstatus = int32(down)
			s_StartSound(&sec.Fsoundorg, int32(sfx_pstart))
		case int32(perpetualRaise):
			platP.Fspeed = 1 << FRACBITS
			platP.Flow = p_FindLowestFloorSurrounding(sec)
			if platP.Flow > sec.Ffloorheight {
				platP.Flow = sec.Ffloorheight
			}
			platP.Fhigh = p_FindHighestFloorSurrounding(sec)
			if platP.Fhigh < sec.Ffloorheight {
				platP.Fhigh = sec.Ffloorheight
			}
			platP.Fwait = TICRATE * PLATWAIT
			platP.Fstatus = p_Random() & 1
			s_StartSound(&sec.Fsoundorg, int32(sfx_pstart))
			break
		}
		p_AddActivePlat(platP)
	}
	return rtn
}

func p_ActivateInStasis(tag int32) {
	for i := 0; i < MAXPLATS; i++ {
		if activeplats[i] != nil && activeplats[i].Ftag == tag && activeplats[i].Fstatus == int32(in_stasis) {
			activeplats[i].Fstatus = activeplats[i].Foldstatus
			activeplats[i].Fthinker.Ffunction = activeplats[i]
		}
	}
}

func ev_StopPlat(line *line_t) {
	for j := 0; j < MAXPLATS; j++ {
		if activeplats[j] != nil && activeplats[j].Fstatus != int32(in_stasis) && activeplats[j].Ftag == int32(line.Ftag) {
			activeplats[j].Foldstatus = activeplats[j].Fstatus
			activeplats[j].Fstatus = int32(in_stasis)
			activeplats[j].Fthinker.Ffunction = nil
		}
	}
}

func p_AddActivePlat(plat *plat_t) {
	for i := 0; i < MAXPLATS; i++ {
		if activeplats[i] == nil {
			activeplats[i] = plat
			return
		}
	}
	i_Error("p_AddActivePlat: no more plats!")
}

func p_RemoveActivePlat(plat *plat_t) {
	for i := 0; i < MAXPLATS; i++ {
		if plat == activeplats[i] {
			activeplats[i].Fsector.Fspecialdata = nil
			p_RemoveThinker(&activeplats[i].Fthinker)
			activeplats[i] = nil
			return
		}
	}
	i_Error("p_RemoveActivePlat: can't find plat!")
}

const ANG1807 = 2147483648
const ANG905 = 1073741824

// Data.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//  Sprite animation.
//

// C documentation
//
//	//
//	// P_SetPsprite
//	//
func p_SetPsprite(player *player_t, position int32, stnum statenum_t) {
	psp := &player.Fpsprites[position]
	for cond := true; cond; cond = psp.Ftics == 0 {
		if stnum == 0 {
			// object removed itself
			psp.Fstate = nil
			break
		}
		state := &states[stnum]
		psp.Fstate = state
		psp.Ftics = state.Ftics // could be 0
		if state.Fmisc1 != 0 {
			// coordinate set
			psp.Fsx = state.Fmisc1 << FRACBITS
			psp.Fsy = state.Fmisc2 << FRACBITS
		}
		// Call action routine.
		// Modified handling.
		if state.Faction != nil {
			state.Faction(player.Fmo, psp)
			if psp.Fstate == nil {
				break
			}
		}
		stnum = psp.Fstate.Fnextstate
	}
	// an initial state of 0 could cycle through
}

// C documentation
//
//	//
//	// P_BringUpWeapon
//	// Starts bringing the pending weapon up
//	// from the bottom of the screen.
//	// Uses player
//	//
func p_BringUpWeapon(player *player_t) {
	var newstate statenum_t
	if player.Fpendingweapon == wp_nochange {
		player.Fpendingweapon = player.Freadyweapon
	}
	if player.Fpendingweapon == wp_chainsaw {
		s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_sawup))
	}
	newstate = weaponinfo[player.Fpendingweapon].Fupstate
	player.Fpendingweapon = wp_nochange
	player.Fpsprites[ps_weapon].Fsy = 128 * (1 << FRACBITS)
	p_SetPsprite(player, int32(ps_weapon), newstate)
}

// C documentation
//
//	//
//	// P_CheckAmmo
//	// Returns true if there is enough ammo to shoot.
//	// If not, selects the next weapon to use.
//	//
func p_CheckAmmo(player *player_t) boolean {
	var ammo ammotype_t
	var count int32
	ammo = weaponinfo[player.Freadyweapon].Fammo
	// Minimal amount for one shot varies.
	if player.Freadyweapon == wp_bfg {
		count = DEH_DEFAULT_BFG_CELLS_PER_SHOT
	} else {
		if player.Freadyweapon == wp_supershotgun {
			count = 2
		} else {
			count = 1
		}
	} // Regular.
	// Some do not need ammunition anyway.
	// Return if current ammunition sufficient.
	if ammo == am_noammo || player.Fammo[ammo] >= count {
		return 1
	}
	// Out of ammo, pick a weapon to change to.
	// Preferences are set here.
	for cond := true; cond; cond = player.Fpendingweapon == wp_nochange {
		if player.Fweaponowned[wp_plasma] != 0 && player.Fammo[am_cell] != 0 && gamemode != shareware {
			player.Fpendingweapon = wp_plasma
		} else {
			if player.Fweaponowned[wp_supershotgun] != 0 && player.Fammo[am_shell] > 2 && gamemode == commercial {
				player.Fpendingweapon = wp_supershotgun
			} else {
				if player.Fweaponowned[wp_chaingun] != 0 && player.Fammo[am_clip] != 0 {
					player.Fpendingweapon = wp_chaingun
				} else {
					if player.Fweaponowned[wp_shotgun] != 0 && player.Fammo[am_shell] != 0 {
						player.Fpendingweapon = wp_shotgun
					} else {
						if player.Fweaponowned[am_clip] != 0 {
							player.Fpendingweapon = wp_pistol
						} else {
							if player.Fweaponowned[wp_chainsaw] != 0 {
								player.Fpendingweapon = wp_chainsaw
							} else {
								if player.Fweaponowned[wp_missile] != 0 && player.Fammo[am_misl] != 0 {
									player.Fpendingweapon = wp_missile
								} else {
									if player.Fweaponowned[wp_bfg] != 0 && player.Fammo[am_cell] > 40 && gamemode != shareware {
										player.Fpendingweapon = wp_bfg
									} else {
										// If everything fails.
										player.Fpendingweapon = wp_fist
									}
								}
							}
						}
					}
				}
			}
		}
	}
	// Now set appropriate weapon overlay.
	p_SetPsprite(player, int32(ps_weapon), weaponinfo[player.Freadyweapon].Fdownstate)
	return 0
}

// C documentation
//
//	//
//	// p_FireWeapon.
//	//
func p_FireWeapon(player *player_t) {
	var newstate statenum_t
	if p_CheckAmmo(player) == 0 {
		return
	}
	p_SetMobjState(player.Fmo, s_PLAY_ATK1)
	newstate = weaponinfo[player.Freadyweapon].Fatkstate
	p_SetPsprite(player, int32(ps_weapon), newstate)
	p_NoiseAlert(player.Fmo, player.Fmo)
}

// C documentation
//
//	//
//	// P_DropWeapon
//	// Player died, so put the weapon away.
//	//
func p_DropWeapon(player *player_t) {
	p_SetPsprite(player, int32(ps_weapon), weaponinfo[player.Freadyweapon].Fdownstate)
}

// C documentation
//
//	//
//	// A_WeaponReady
//	// The player can fire the weapon
//	// or change to another weapon at this time.
//	// Follows after getting weapon up,
//	// or after previous attack/fire sequence.
//	//
func a_WeaponReady(player *player_t, psp *pspdef_t) {
	var angle int32
	var newstate statenum_t
	// get out of attack state
	if player.Fmo.Fstate == &states[s_PLAY_ATK1] || player.Fmo.Fstate == &states[s_PLAY_ATK2] {
		p_SetMobjState(player.Fmo, s_PLAY)
	}
	if player.Freadyweapon == wp_chainsaw && psp.Fstate == &states[s_SAW] {
		s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_sawidl))
	}
	// check for change
	//  if player is dead, put the weapon away
	if player.Fpendingweapon != wp_nochange || player.Fhealth == 0 {
		// change weapon
		//  (pending weapon should allready be validated)
		newstate = weaponinfo[player.Freadyweapon].Fdownstate
		p_SetPsprite(player, int32(ps_weapon), newstate)
		return
	}
	// check for fire
	//  the missile launcher and bfg do not auto fire
	if int32(player.Fcmd.Fbuttons)&bt_ATTACK != 0 {
		if player.Fattackdown == 0 || player.Freadyweapon != wp_missile && player.Freadyweapon != wp_bfg {
			player.Fattackdown = 1
			p_FireWeapon(player)
			return
		}
	} else {
		player.Fattackdown = 0
	}
	// bob the weapon based on movement speed
	angle = 128 * leveltime & (FINEANGLES - 1)
	psp.Fsx = 1<<FRACBITS + fixedMul(player.Fbob, finecosine[angle])
	angle &= FINEANGLES/2 - 1
	psp.Fsy = 32*(1<<FRACBITS) + fixedMul(player.Fbob, finesine[angle])
}

// C documentation
//
//	//
//	// A_ReFire
//	// The player can re-fire the weapon
//	// without lowering it entirely.
//	//
func a_ReFire(player *player_t, psp *pspdef_t) {
	// check for fire
	//  (if a weaponchange is pending, let it go through instead)
	if int32(player.Fcmd.Fbuttons)&bt_ATTACK != 0 && player.Fpendingweapon == wp_nochange && player.Fhealth != 0 {
		player.Frefire++
		p_FireWeapon(player)
	} else {
		player.Frefire = 0
		p_CheckAmmo(player)
	}
}

func a_CheckReload(player *player_t, psp *pspdef_t) {
	p_CheckAmmo(player)
}

// C documentation
//
//	//
//	// A_Lower
//	// Lowers current weapon,
//	//  and changes weapon at bottom.
//	//
func a_Lower(player *player_t, psp *pspdef_t) {
	psp.Fsy += 1 << FRACBITS * 6
	// Is already down.
	if psp.Fsy < 128*(1<<FRACBITS) {
		return
	}
	// Player is dead.
	if player.Fplayerstate == Pst_DEAD {
		psp.Fsy = 128 * (1 << FRACBITS)
		// don't bring weapon back up
		return
	}
	// The old weapon has been lowered off the screen,
	// so change the weapon and start raising it
	if player.Fhealth == 0 {
		// Player is dead, so keep the weapon off screen.
		p_SetPsprite(player, int32(ps_weapon), s_NULL)
		return
	}
	player.Freadyweapon = player.Fpendingweapon
	p_BringUpWeapon(player)
}

// C documentation
//
//	//
//	// A_Raise
//	//
func a_Raise(player *player_t, psp *pspdef_t) {
	var newstate statenum_t
	psp.Fsy -= 1 << FRACBITS * 6
	if psp.Fsy > 32*(1<<FRACBITS) {
		return
	}
	psp.Fsy = 32 * (1 << FRACBITS)
	// The weapon has been raised all the way,
	//  so change to the ready state.
	newstate = weaponinfo[player.Freadyweapon].Freadystate
	p_SetPsprite(player, int32(ps_weapon), newstate)
}

// C documentation
//
//	//
//	// A_GunFlash
//	//
func a_GunFlash(player *player_t, psp *pspdef_t) {
	p_SetMobjState(player.Fmo, s_PLAY_ATK2)
	p_SetPsprite(player, int32(ps_flash), weaponinfo[player.Freadyweapon].Fflashstate)
}

//
// WEAPON ATTACKS
//

// C documentation
//
//	//
//	// A_Punch
//	//
func a_Punch(player *player_t, psp *pspdef_t) {
	var angle angle_t
	var damage, slope int32
	damage = (p_Random()%int32(10) + 1) << 1
	if player.Fpowers[pw_strength] != 0 {
		damage *= 10
	}
	angle = player.Fmo.Fangle
	angle += uint32((p_Random() - p_Random()) << 18)
	slope = p_AimLineAttack(player.Fmo, angle, 64*(1<<FRACBITS))
	p_LineAttack(player.Fmo, angle, 64*(1<<FRACBITS), slope, damage)
	// turn to face target
	if linetarget != nil {
		s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_punch))
		player.Fmo.Fangle = r_PointToAngle2(player.Fmo.Fx, player.Fmo.Fy, linetarget.Fx, linetarget.Fy)
	}
}

// C documentation
//
//	//
//	// A_Saw
//	//
func a_Saw(player *player_t, psp *pspdef_t) {
	var angle angle_t
	var damage, slope int32
	damage = 2 * (p_Random()%int32(10) + 1)
	angle = player.Fmo.Fangle
	angle += uint32((p_Random() - p_Random()) << 18)
	// use meleerange + 1 se the puff doesn't skip the flash
	slope = p_AimLineAttack(player.Fmo, angle, 64*(1<<FRACBITS)+1)
	p_LineAttack(player.Fmo, angle, 64*(1<<FRACBITS)+1, slope, damage)
	if linetarget == nil {
		s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_sawful))
		return
	}
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_sawhit))
	// turn to face target
	angle = r_PointToAngle2(player.Fmo.Fx, player.Fmo.Fy, linetarget.Fx, linetarget.Fy)
	if angle-player.Fmo.Fangle > uint32(ANG1807) {
		if int32(angle-player.Fmo.Fangle) < -ANG905/20 {
			player.Fmo.Fangle = angle + uint32(ANG905/21)
		} else {
			player.Fmo.Fangle -= uint32(ANG905 / 20)
		}
	} else {
		if angle-player.Fmo.Fangle > uint32(ANG905/20) {
			player.Fmo.Fangle = angle - uint32(ANG905/21)
		} else {
			player.Fmo.Fangle += uint32(ANG905 / 20)
		}
	}
	player.Fmo.Fflags |= mf_JUSTATTACKED
}

// Doom does not check the bounds of the ammo array.  As a result,
// it is possible to use an ammo type > 4 that overflows into the
// maxammo array and affects that instead.  Through dehacked, for
// example, it is possible to make a weapon that decreases the max
// number of ammo for another weapon.  Emulate this.

func decreaseAmmo(player *player_t, ammonum ammotype_t, amount int32) {
	if ammonum < NUMAMMO {
		player.Fammo[ammonum] -= amount
	} else {
		player.Fmaxammo[ammonum-NUMAMMO] -= amount
	}
}

// C documentation
//
//	//
//	// A_FireMissile
//	//
func a_FireMissile(player *player_t, psp *pspdef_t) {
	decreaseAmmo(player, weaponinfo[player.Freadyweapon].Fammo, 1)
	p_SpawnPlayerMissile(player.Fmo, mt_ROCKET)
}

// C documentation
//
//	//
//	// A_FireBFG
//	//
func a_FireBFG(player *player_t, psp *pspdef_t) {
	decreaseAmmo(player, weaponinfo[player.Freadyweapon].Fammo, DEH_DEFAULT_BFG_CELLS_PER_SHOT)
	p_SpawnPlayerMissile(player.Fmo, mt_BFG)
}

// C documentation
//
//	//
//	// A_FirePlasma
//	//
func a_FirePlasma(player *player_t, psp *pspdef_t) {
	decreaseAmmo(player, weaponinfo[player.Freadyweapon].Fammo, 1)
	p_SetPsprite(player, int32(ps_flash), weaponinfo[player.Freadyweapon].Fflashstate+p_Random()&1)
	p_SpawnPlayerMissile(player.Fmo, mt_PLASMA)
}

func p_BulletSlope(mo *mobj_t) {
	var an angle_t
	// see which target is to be aimed at
	an = mo.Fangle
	bulletslope = p_AimLineAttack(mo, an, 16*64*(1<<FRACBITS))
	if linetarget == nil {
		an += uint32(1 << 26)
		bulletslope = p_AimLineAttack(mo, an, 16*64*(1<<FRACBITS))
		if linetarget == nil {
			an -= uint32(2 << 26)
			bulletslope = p_AimLineAttack(mo, an, 16*64*(1<<FRACBITS))
		}
	}
}

// C documentation
//
//	//
//	// P_GunShot
//	//
func p_GunShot(mo *mobj_t, accurate boolean) {
	var angle angle_t
	var damage int32
	damage = 5 * (p_Random()%3 + 1)
	angle = mo.Fangle
	if accurate == 0 {
		angle += uint32((p_Random() - p_Random()) << 18)
	}
	p_LineAttack(mo, angle, 32*64*(1<<FRACBITS), bulletslope, damage)
}

// C documentation
//
//	//
//	// A_FirePistol
//	//
func a_FirePistol(player *player_t, psp *pspdef_t) {
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_pistol))
	p_SetMobjState(player.Fmo, s_PLAY_ATK2)
	decreaseAmmo(player, weaponinfo[player.Freadyweapon].Fammo, 1)
	p_SetPsprite(player, int32(ps_flash), weaponinfo[player.Freadyweapon].Fflashstate)
	p_BulletSlope(player.Fmo)
	p_GunShot(player.Fmo, booluint32(player.Frefire == 0))
}

// C documentation
//
//	//
//	// A_FireShotgun
//	//
func a_FireShotgun(player *player_t, psp *pspdef_t) {
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_shotgn))
	p_SetMobjState(player.Fmo, s_PLAY_ATK2)
	decreaseAmmo(player, weaponinfo[player.Freadyweapon].Fammo, 1)
	p_SetPsprite(player, int32(ps_flash), weaponinfo[player.Freadyweapon].Fflashstate)
	p_BulletSlope(player.Fmo)
	for i := 0; i < 7; i++ {
		p_GunShot(player.Fmo, 0)
	}
}

// C documentation
//
//	//
//	// A_FireShotgun2
//	//
func a_FireShotgun2(player *player_t, psp *pspdef_t) {
	var angle angle_t
	var damage int32
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_dshtgn))
	p_SetMobjState(player.Fmo, s_PLAY_ATK2)
	decreaseAmmo(player, weaponinfo[player.Freadyweapon].Fammo, 2)
	p_SetPsprite(player, int32(ps_flash), weaponinfo[player.Freadyweapon].Fflashstate)
	p_BulletSlope(player.Fmo)
	for i := 0; i < 20; i++ {
		damage = 5 * (p_Random()%3 + 1)
		angle = player.Fmo.Fangle
		angle += uint32((p_Random() - p_Random()) << 19)
		p_LineAttack(player.Fmo, angle, 32*64*(1<<FRACBITS), bulletslope+(p_Random()-p_Random())<<5, damage)
	}
}

// C documentation
//
//	//
//	// A_FireCGun
//	//
func a_FireCGun(player *player_t, psp *pspdef_t) {
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_pistol))
	if player.Fammo[weaponinfo[player.Freadyweapon].Fammo] == 0 {
		return
	}
	p_SetMobjState(player.Fmo, s_PLAY_ATK2)
	decreaseAmmo(player, weaponinfo[player.Freadyweapon].Fammo, 1)
	newState := weaponinfo[player.Freadyweapon].Fflashstate + stateIndex(psp.Fstate) - s_CHAIN1
	p_SetPsprite(player, int32(ps_flash), newState)
	p_BulletSlope(player.Fmo)
	p_GunShot(player.Fmo, booluint32(player.Frefire == 0))
}

// C documentation
//
//	//
//	// ?
//	//
func a_Light0(player *player_t, psp *pspdef_t) {
	player.Fextralight = 0
}

func a_Light1(player *player_t, psp *pspdef_t) {
	player.Fextralight = 1
}

func a_Light2(player *player_t, psp *pspdef_t) {
	player.Fextralight = 2
}

// C documentation
//
//	//
//	// A_BFGSpray
//	// Spawn a BFG explosion on every monster in view
//	//
func a_BFGSpray(mo *mobj_t) {
	var an angle_t
	var damage, i, j int32
	// offset angles from its attack angle
	i = 0
	for {
		if i >= 40 {
			break
		}
		an = mo.Fangle - uint32(ANG905/2) + uint32(ANG905/40*i)
		// mo->target is the originator (player)
		//  of the missile
		p_AimLineAttack(mo.Ftarget, an, 16*64*(1<<FRACBITS))
		if linetarget == nil {
			goto _1
		}
		p_SpawnMobj(linetarget.Fx, linetarget.Fy, linetarget.Fz+linetarget.Fheight>>2, mt_EXTRABFG)
		damage = 0
		j = 0
		for {
			if j >= 15 {
				break
			}
			damage += p_Random()&7 + 1
			goto _2
		_2:
			;
			j++
		}
		p_DamageMobj(linetarget, mo.Ftarget, mo.Ftarget, damage)
		goto _1
	_1:
		;
		i++
	}
}

// C documentation
//
//	//
//	// A_BFGsound
//	//
func a_BFGsound(player *player_t, psp *pspdef_t) {
	s_StartSound(&player.Fmo.degenmobj_t, int32(sfx_bfg))
}

// C documentation
//
//	//
//	// P_SetupPsprites
//	// Called at start of level for each player.
//	//
func p_SetupPsprites(player *player_t) {
	var i int32
	// remove all psprites
	i = 0
	for {
		if i >= NUMPSPRITES {
			break
		}
		player.Fpsprites[i].Fstate = nil
		goto _1
	_1:
		;
		i++
	}
	// spawn the gun
	player.Fpendingweapon = player.Freadyweapon
	p_BringUpWeapon(player)
}

// C documentation
//
//	//
//	// P_MovePsprites
//	// Called every tic by player thinking routine.
//	//
func p_MovePsprites(player *player_t) {
	var i int32
	var v2 *state_t
	i = 0
	for {
		if i >= NUMPSPRITES {
			break
		}
		psp := &player.Fpsprites[i]
		// a null state means not active
		v2 = psp.Fstate
		if v2 != nil {
			// drop tic count and possibly change state
			// a -1 tic count never changes
			if psp.Ftics != -1 {
				psp.Ftics--
				if psp.Ftics == 0 {
					p_SetPsprite(player, i, psp.Fstate.Fnextstate)
				}
			}
		}
		goto _1
	_1:
		;
		i++
	}
	player.Fpsprites[ps_flash].Fsx = player.Fpsprites[ps_weapon].Fsx
	player.Fpsprites[ps_flash].Fsy = player.Fpsprites[ps_weapon].Fsy
}

const SAVEGAME_EOF = 29

// Get the filename of a temporary file to write the savegame to.  After
// the file has been successfully saved, it will be renamed to the
// real file.

func p_TempSaveGameFile() string {
	if filename == "" {
		filename = savegamedir + "temp.dsg"
	}
	return filename
}

var filename string

// Get the filename of the save game file to use for the specified slot.

func p_SaveGameFile(slot int32) string {
	return fmt.Sprintf("%sdgsave%d.dsg", savegamedir, slot)
}

// Endian-safe integer read/write functions

func saveg_read8() uint8 {
	var val [1]byte
	if _, err := save_stream.Read(val[:]); err != nil {
		if savegame_error == 0 {
			fprintf_ccgo(os.Stderr, "saveg_read8: Unexpected end of file while reading save game\n")
			savegame_error = 1
		}
	}
	return val[0]
}

func saveg_write8(_value uint8) {
	val := [1]byte{_value}
	if _, err := save_stream.Write(val[:]); err != nil {
		if savegame_error == 0 {
			fprintf_ccgo(os.Stderr, "saveg_write8: Error while writing save game\n")
			savegame_error = 1
		}
	}
}

func saveg_read16() int16 {
	var result int32
	result = int32(saveg_read8())
	result |= int32(saveg_read8()) << 8
	return int16(result)
}

func saveg_write16(value int16) {
	saveg_write8(uint8(int32(value) & 0xff))
	saveg_write8(uint8(int32(value) >> 8 & 0xff))
}

func saveg_read32() int32 {
	var result int32
	result = int32(saveg_read8())
	result |= int32(saveg_read8()) << 8
	result |= int32(saveg_read8()) << 16
	result |= int32(saveg_read8()) << 24
	return result
}

func saveg_write32(value int32) {
	saveg_write8(uint8(value & 0xff))
	saveg_write8(uint8(value >> 8 & 0xff))
	saveg_write8(uint8(value >> int32(16) & 0xff))
	saveg_write8(uint8(value >> int32(24) & 0xff))
}

// Pad to 4-byte boundaries

func saveg_read_pad() {
	var i, padding int32
	pos, _ := save_stream.Seek(0, io.SeekCurrent)
	padding = int32((4 - uint64(pos)&uint64(3)) & uint64(3))
	i = 0
	for {
		if i >= padding {
			break
		}
		saveg_read8()
		goto _1
	_1:
		;
		i++
	}
}

func saveg_write_pad() {
	var i, padding int32
	pos, _ := save_stream.Seek(0, io.SeekCurrent)
	padding = int32((4 - uint64(pos)&uint64(3)) & uint64(3))
	i = 0
	for {
		if i >= padding {
			break
		}
		saveg_write8(0)
		goto _1
	_1:
		;
		i++
	}
}

// Pointers

func saveg_readp() uintptr {
	return uintptr(int64(saveg_read32()))
}

func saveg_writep(p uintptr) {
	saveg_write32(int32(int64(p)))
}

// Enum values are 32-bit integers.

//
// Structure read/write functions
//

//
// mapthing_t
//

func saveg_read_mapthing_t(str *mapthing_t) {
	// short x;
	str.Fx = saveg_read16()
	// short y;
	str.Fy = saveg_read16()
	// short angle;
	str.Fangle = saveg_read16()
	// short type;
	str.Ftype1 = saveg_read16()
	// short options;
	str.Foptions = saveg_read16()
}

func saveg_write_mapthing_t(str *mapthing_t) {
	// short x;
	saveg_write16(str.Fx)
	// short y;
	saveg_write16(str.Fy)
	// short angle;
	saveg_write16(str.Fangle)
	// short type;
	saveg_write16(str.Ftype1)
	// short options;
	saveg_write16(str.Foptions)
}

//
// actionf_t
//

func saveg_read_actionf_t(str *thinker_func_t) {
	// actionf_p1 acp1;
	str = (*thinker_func_t)(unsafe.Pointer(saveg_readp()))
}

func saveg_write_actionf_t(str *thinker_func_t) {
	// actionf_p1 acp1;
	saveg_writep(uintptr(unsafe.Pointer(str)))
}

//
// think_t
//
// This is just an actionf_t.
//

//
// thinker_t
//

func saveg_read_thinker_t(str *thinker_t) {
	// struct thinker_t* prev;
	str.Fprev = (*thinker_t)(unsafe.Pointer(saveg_readp()))
	// struct thinker_t* next;
	str.Fnext = (*thinker_t)(unsafe.Pointer(saveg_readp()))
	// think_t function;

	saveg_read_actionf_t(&str.Ffunction)
}

func saveg_write_thinker_t(str *thinker_t) {
	// struct thinker_t* prev;
	saveg_writep(uintptr(unsafe.Pointer(str.Fprev)))
	// struct thinker_t* next;
	saveg_writep(uintptr(unsafe.Pointer(str.Fnext)))
	// think_t function;
	saveg_write_actionf_t(&str.Ffunction)
}

//
// mobj_t
//

func saveg_read_mobj_t(str *mobj_t) {
	var pl int32
	// thinker_t thinker;
	saveg_read_thinker_t(&str.Fthinker)
	// fixed_t x;
	str.Fx = saveg_read32()
	// fixed_t y;
	str.Fy = saveg_read32()
	// fixed_t z;
	str.Fz = saveg_read32()
	// struct mobj_t* snext;
	str.Fsnext = (*mobj_t)(unsafe.Pointer(saveg_readp()))
	// struct mobj_t* sprev;
	str.Fsprev = (*mobj_t)(unsafe.Pointer(saveg_readp()))
	// angle_t angle;
	str.Fangle = uint32(saveg_read32())
	// spritenum_t sprite;
	str.Fsprite = saveg_read32()
	// int frame;
	str.Fframe = saveg_read32()
	// struct mobj_t* bnext;
	str.Fbnext = (*mobj_t)(unsafe.Pointer(saveg_readp()))
	// struct mobj_t* bprev;
	str.Fbprev = (*mobj_t)(unsafe.Pointer(saveg_readp()))
	// struct subsector_t* subsector;
	str.Fsubsector = (*subsector_t)(unsafe.Pointer(saveg_readp()))
	// fixed_t floorz;
	str.Ffloorz = saveg_read32()
	// fixed_t ceilingz;
	str.Fceilingz = saveg_read32()
	// fixed_t radius;
	str.Fradius = saveg_read32()
	// fixed_t height;
	str.Fheight = saveg_read32()
	// fixed_t momx;
	str.Fmomx = saveg_read32()
	// fixed_t momy;
	str.Fmomy = saveg_read32()
	// fixed_t momz;
	str.Fmomz = saveg_read32()
	// int validcount;
	str.Fvalidcount = saveg_read32()
	// mobjtype_t type;
	str.Ftype1 = saveg_read32()
	// mobjinfo_t* info;
	str.Finfo = (*mobjinfo_t)(unsafe.Pointer(saveg_readp()))
	// int tics;
	str.Ftics = saveg_read32()
	// state_t* state;
	str.Fstate = &states[saveg_read32()]
	// int flags;
	str.Fflags = saveg_read32()
	// int health;
	str.Fhealth = saveg_read32()
	// int movedir;
	str.Fmovedir = saveg_read32()
	// int movecount;
	str.Fmovecount = saveg_read32()
	// struct mobj_t* target;
	str.Ftarget = (*mobj_t)(unsafe.Pointer(saveg_readp()))
	// int reactiontime;
	str.Freactiontime = saveg_read32()
	// int threshold;
	str.Fthreshold = saveg_read32()
	// player_t* player;
	pl = saveg_read32()
	if pl > 0 {
		str.Fplayer = &players[pl-1]
		str.Fplayer.Fmo = str
	} else {
		str.Fplayer = nil
	}
	// int lastlook;
	str.Flastlook = saveg_read32()
	// mapthing_t spawnpoint;
	saveg_read_mapthing_t(&str.Fspawnpoint)
	// struct mobj_t* tracer;
	str.Ftracer = (*mobj_t)(unsafe.Pointer(saveg_readp()))
}

func saveg_write_mobj_t(str *mobj_t) {
	// thinker_t thinker;
	saveg_write_thinker_t(&str.Fthinker)
	// fixed_t x;
	saveg_write32(str.Fx)
	// fixed_t y;
	saveg_write32(str.Fy)
	// fixed_t z;
	saveg_write32(str.Fz)
	// struct mobj_t* snext;
	saveg_writep(uintptr(unsafe.Pointer(str.Fsnext)))
	// struct mobj_t* sprev;
	saveg_writep(uintptr(unsafe.Pointer(str.Fsprev)))
	// angle_t angle;
	saveg_write32(int32(str.Fangle))
	// spritenum_t sprite;
	saveg_write32(str.Fsprite)
	// int frame;
	saveg_write32(str.Fframe)
	// struct mobj_t* bnext;
	saveg_writep(uintptr(unsafe.Pointer(str.Fbnext)))
	// struct mobj_t* bprev;
	saveg_writep(uintptr(unsafe.Pointer(str.Fbprev)))
	// struct subsector_t* subsector;
	saveg_writep(uintptr(unsafe.Pointer(str.Fsubsector)))
	// fixed_t floorz;
	saveg_write32(str.Ffloorz)
	// fixed_t ceilingz;
	saveg_write32(str.Fceilingz)
	// fixed_t radius;
	saveg_write32(str.Fradius)
	// fixed_t height;
	saveg_write32(str.Fheight)
	// fixed_t momx;
	saveg_write32(str.Fmomx)
	// fixed_t momy;
	saveg_write32(str.Fmomy)
	// fixed_t momz;
	saveg_write32(str.Fmomz)
	// int validcount;
	saveg_write32(str.Fvalidcount)
	// mobjtype_t type;
	saveg_write32(str.Ftype1)
	// mobjinfo_t* info;
	saveg_writep(uintptr(unsafe.Pointer(str.Finfo)))
	// int tics;
	saveg_write32(str.Ftics)
	// state_t* state;
	idx := stateIndex(str.Fstate)
	saveg_write32(idx)
	// int flags;
	saveg_write32(str.Fflags)
	// int health;
	saveg_write32(str.Fhealth)
	// int movedir;
	saveg_write32(str.Fmovedir)
	// int movecount;
	saveg_write32(str.Fmovecount)
	// struct mobj_t* target;
	saveg_writep(uintptr(unsafe.Pointer(str.Ftarget)))
	// int reactiontime;
	saveg_write32(str.Freactiontime)
	// int threshold;
	saveg_write32(str.Fthreshold)
	// player_t* player;
	if str.Fplayer != nil {
		idx := int32(playerIndex(str.Fplayer))
		saveg_write32(idx + 1)
	} else {
		saveg_write32(0)
	}
	// int lastlook;
	saveg_write32(str.Flastlook)
	// mapthing_t spawnpoint;
	saveg_write_mapthing_t(&str.Fspawnpoint)
	// struct mobj_t* tracer;
	saveg_writep(uintptr(unsafe.Pointer(str.Ftracer)))
}

//
// ticcmd_t
//

func saveg_read_ticcmd_t(str *ticcmd_t) {
	// signed char forwardmove;
	str.Fforwardmove = int8(saveg_read8())
	// signed char sidemove;
	str.Fsidemove = int8(saveg_read8())
	// short angleturn;
	str.Fangleturn = saveg_read16()
	// short consistancy;
	str.Fconsistancy = uint8(saveg_read16())
	// byte chatchar;
	str.Fchatchar = saveg_read8()
	// byte buttons;
	str.Fbuttons = saveg_read8()
}

func saveg_write_ticcmd_t(str *ticcmd_t) {
	// signed char forwardmove;
	saveg_write8(uint8(str.Fforwardmove))
	// signed char sidemove;
	saveg_write8(uint8(str.Fsidemove))
	// short angleturn;
	saveg_write16(str.Fangleturn)
	// short consistancy;
	saveg_write16(int16(str.Fconsistancy))
	// byte chatchar;
	saveg_write8(str.Fchatchar)
	// byte buttons;
	saveg_write8(str.Fbuttons)
}

//
// pspdef_t
//

func saveg_read_pspdef_t(str *pspdef_t) {
	var state int32
	// state_t* state;
	state = saveg_read32()
	if state > 0 {
		str.Fstate = &states[state]
	} else {
		str.Fstate = nil
	}
	// int tics;
	str.Ftics = saveg_read32()
	// fixed_t sx;
	str.Fsx = saveg_read32()
	// fixed_t sy;
	str.Fsy = saveg_read32()
}

func saveg_write_pspdef_t(str *pspdef_t) {
	// state_t* state;
	if str.Fstate != nil {
		saveg_write32(stateIndex(str.Fstate))
	} else {
		saveg_write32(0)
	}
	// int tics;
	saveg_write32(str.Ftics)
	// fixed_t sx;
	saveg_write32(str.Fsx)
	// fixed_t sy;
	saveg_write32(str.Fsy)
}

//
// player_t
//

func saveg_read_player_t(str *player_t) {
	// mobj_t* mo;
	str.Fmo = (*mobj_t)(unsafe.Pointer(saveg_readp()))
	// playerstate_t playerstate;
	str.Fplayerstate = saveg_read32()
	// ticcmd_t cmd;
	saveg_read_ticcmd_t(&str.Fcmd)
	// fixed_t viewz;
	str.Fviewz = saveg_read32()
	// fixed_t viewheight;
	str.Fviewheight = saveg_read32()
	// fixed_t deltaviewheight;
	str.Fdeltaviewheight = saveg_read32()
	// fixed_t bob;
	str.Fbob = saveg_read32()
	// int health;
	str.Fhealth = saveg_read32()
	// int armorpoints;
	str.Farmorpoints = saveg_read32()
	// int armortype;
	str.Farmortype = saveg_read32()
	// int powers[NUMPOWERS];
	for i := 0; i < NUMPOWERS; i++ {
		str.Fpowers[i] = saveg_read32()
	}
	// boolean cards[NUMCARDS];
	for i := 0; i < NUMCARDS; i++ {
		str.Fcards[i] = uint32(saveg_read32())
	}
	// boolean backpack;
	str.Fbackpack = uint32(saveg_read32())
	// int frags[MAXPLAYERS];
	for i := 0; i < MAXPLAYERS; i++ {
		str.Ffrags[i] = saveg_read32()
	}
	// weapontype_t readyweapon;
	str.Freadyweapon = weapontype_t(saveg_read32())
	// weapontype_t pendingweapon;
	str.Fpendingweapon = weapontype_t(saveg_read32())
	// boolean weaponowned[NUMWEAPONS];
	for i := 0; i < NUMWEAPONS; i++ {
		str.Fweaponowned[i] = uint32(saveg_read32())
	}
	for i := 0; i < NUMAMMO; i++ {
		str.Fammo[i] = saveg_read32()
	}
	// int maxammo[NUMAMMO];
	for i := 0; i < NUMAMMO; i++ {
		str.Fmaxammo[i] = saveg_read32()
	}
	// int attackdown;
	str.Fattackdown = saveg_read32()
	// int usedown;
	str.Fusedown = saveg_read32()
	// int cheats;
	str.Fcheats = saveg_read32()
	// int refire;
	str.Frefire = saveg_read32()
	// int killcount;
	str.Fkillcount = saveg_read32()
	// int itemcount;
	str.Fitemcount = saveg_read32()
	// int secretcount;
	str.Fsecretcount = saveg_read32()
	// char* message;
	str.Fmessage = gostring(saveg_readp())
	// int damagecount;
	str.Fdamagecount = saveg_read32()
	// int bonuscount;
	str.Fbonuscount = saveg_read32()
	// mobj_t* attacker;
	str.Fattacker = (*mobj_t)(unsafe.Pointer(saveg_readp()))
	// int extralight;
	str.Fextralight = saveg_read32()
	// int fixedcolormap;
	str.Ffixedcolormap = saveg_read32()
	// int colormap;
	str.Fcolormap = saveg_read32()
	// pspdef_t psprites[NUMPSPRITES];
	for i := 0; i < NUMPSPRITES; i++ {
		saveg_read_pspdef_t(&str.Fpsprites[i])
	}
	// boolean didsecret;
	str.Fdidsecret = uint32(saveg_read32())
}

func saveg_write_player_t(str *player_t) {
	var i int32
	// mobj_t* mo;
	saveg_writep(uintptr(unsafe.Pointer(str.Fmo)))
	// playerstate_t playerstate;
	saveg_write32(str.Fplayerstate)
	// ticcmd_t cmd;
	saveg_write_ticcmd_t(&str.Fcmd)
	// fixed_t viewz;
	saveg_write32(str.Fviewz)
	// fixed_t viewheight;
	saveg_write32(str.Fviewheight)
	// fixed_t deltaviewheight;
	saveg_write32(str.Fdeltaviewheight)
	// fixed_t bob;
	saveg_write32(str.Fbob)
	// int health;
	saveg_write32(str.Fhealth)
	// int armorpoints;
	saveg_write32(str.Farmorpoints)
	// int armortype;
	saveg_write32(str.Farmortype)
	// int powers[NUMPOWERS];
	i = 0
	for {
		if i >= NUMPOWERS {
			break
		}
		saveg_write32(str.Fpowers[i])
		goto _1
	_1:
		;
		i++
	}
	// boolean cards[NUMCARDS];
	i = 0
	for {
		if i >= NUMCARDS {
			break
		}
		saveg_write32(int32(str.Fcards[i]))
		goto _2
	_2:
		;
		i++
	}
	// boolean backpack;
	saveg_write32(int32(str.Fbackpack))
	// int frags[MAXPLAYERS];
	i = 0
	for {
		if i >= MAXPLAYERS {
			break
		}
		saveg_write32(str.Ffrags[i])
		goto _3
	_3:
		;
		i++
	}
	// weapontype_t readyweapon;
	saveg_write32(int32(str.Freadyweapon))
	// weapontype_t pendingweapon;
	saveg_write32(int32(str.Fpendingweapon))
	// boolean weaponowned[NUMWEAPONS];
	i = 0
	for {
		if i >= NUMWEAPONS {
			break
		}
		saveg_write32(int32(str.Fweaponowned[i]))
		goto _4
	_4:
		;
		i++
	}
	// int ammo[NUMAMMO];
	i = 0
	for {
		if i >= NUMAMMO {
			break
		}
		saveg_write32(str.Fammo[i])
		goto _5
	_5:
		;
		i++
	}
	// int maxammo[NUMAMMO];
	i = 0
	for {
		if i >= NUMAMMO {
			break
		}
		saveg_write32(str.Fmaxammo[i])
		goto _6
	_6:
		;
		i++
	}
	// int attackdown;
	saveg_write32(str.Fattackdown)
	// int usedown;
	saveg_write32(str.Fusedown)
	// int cheats;
	saveg_write32(str.Fcheats)
	// int refire;
	saveg_write32(str.Frefire)
	// int killcount;
	saveg_write32(str.Fkillcount)
	// int itemcount;
	saveg_write32(str.Fitemcount)
	// int secretcount;
	saveg_write32(str.Fsecretcount)
	// char* message;
	if str.Fmessage == "" {
		saveg_writep(0)
	} else {
		saveg_writep(uintptr(unsafe.Pointer(&[]byte(str.Fmessage)[0])))
	}
	// int damagecount;
	saveg_write32(str.Fdamagecount)
	// int bonuscount;
	saveg_write32(str.Fbonuscount)
	// mobj_t* attacker;
	saveg_writep(uintptr(unsafe.Pointer(str.Fattacker)))
	// int extralight;
	saveg_write32(str.Fextralight)
	// int fixedcolormap;
	saveg_write32(str.Ffixedcolormap)
	// int colormap;
	saveg_write32(str.Fcolormap)
	// pspdef_t psprites[NUMPSPRITES];
	i = 0
	for {
		if i >= NUMPSPRITES {
			break
		}
		saveg_write_pspdef_t(&str.Fpsprites[i])
		goto _7
	_7:
		;
		i++
	}
	// boolean didsecret;
	saveg_write32(int32(str.Fdidsecret))
}

//
// ceiling_t
//

func saveg_read_ceiling_t(str *ceiling_t) {
	var sector int32
	// thinker_t thinker;
	saveg_read_thinker_t(&str.Fthinker)
	// ceiling_e type;
	str.Ftype1 = saveg_read32()
	// sector_t* sector;
	sector = saveg_read32()
	str.Fsector = &sectors[sector]
	// fixed_t bottomheight;
	str.Fbottomheight = saveg_read32()
	// fixed_t topheight;
	str.Ftopheight = saveg_read32()
	// fixed_t speed;
	str.Fspeed = saveg_read32()
	// boolean crush;
	str.Fcrush = uint32(saveg_read32())
	// int direction;
	str.Fdirection = saveg_read32()
	// int tag;
	str.Ftag = saveg_read32()
	// int olddirection;
	str.Folddirection = saveg_read32()
}

func saveg_write_ceiling_t(str *ceiling_t) {
	// thinker_t thinker;
	saveg_write_thinker_t(&str.Fthinker)
	// ceiling_e type;
	saveg_write32(str.Ftype1)
	// sector_t* sector;
	saveg_write32(sectorIndex(str.Fsector))
	// fixed_t bottomheight;
	saveg_write32(str.Fbottomheight)
	// fixed_t topheight;
	saveg_write32(str.Ftopheight)
	// fixed_t speed;
	saveg_write32(str.Fspeed)
	// boolean crush;
	saveg_write32(int32(str.Fcrush))
	// int direction;
	saveg_write32(str.Fdirection)
	// int tag;
	saveg_write32(str.Ftag)
	// int olddirection;
	saveg_write32(str.Folddirection)
}

//
// vldoor_t
//

func saveg_read_vldoor_t(str *vldoor_t) {
	var sector int32
	// thinker_t thinker;
	saveg_read_thinker_t(&str.Fthinker)
	// vldoor_e type;
	str.Ftype1 = saveg_read32()
	// sector_t* sector;
	sector = saveg_read32()
	str.Fsector = &sectors[sector]
	// fixed_t topheight;
	str.Ftopheight = saveg_read32()
	// fixed_t speed;
	str.Fspeed = saveg_read32()
	// int direction;
	str.Fdirection = saveg_read32()
	// int topwait;
	str.Ftopwait = saveg_read32()
	// int topcountdown;
	str.Ftopcountdown = saveg_read32()
}

func saveg_write_vldoor_t(str *vldoor_t) {
	// thinker_t thinker;
	saveg_write_thinker_t(&str.Fthinker)
	// vldoor_e type;
	saveg_write32(str.Ftype1)
	// sector_t* sector;
	saveg_write32(sectorIndex(str.Fsector))
	// fixed_t topheight;
	saveg_write32(str.Ftopheight)
	// fixed_t speed;
	saveg_write32(str.Fspeed)
	// int direction;
	saveg_write32(str.Fdirection)
	// int topwait;
	saveg_write32(str.Ftopwait)
	// int topcountdown;
	saveg_write32(str.Ftopcountdown)
}

//
// floormove_t
//

func saveg_read_floormove_t(str *floormove_t) {
	var sector int32
	// thinker_t thinker;
	saveg_read_thinker_t(&str.Fthinker)
	// floor_e type;
	str.Ftype1 = saveg_read32()
	// boolean crush;
	str.Fcrush = uint32(saveg_read32())
	// sector_t* sector;
	sector = saveg_read32()
	str.Fsector = &sectors[sector]
	// int direction;
	str.Fdirection = saveg_read32()
	// int newspecial;
	str.Fnewspecial = saveg_read32()
	// short texture;
	str.Ftexture = saveg_read16()
	// fixed_t floordestheight;
	str.Ffloordestheight = saveg_read32()
	// fixed_t speed;
	str.Fspeed = saveg_read32()
}

func saveg_write_floormove_t(str *floormove_t) {
	// thinker_t thinker;
	saveg_write_thinker_t(&str.Fthinker)
	// floor_e type;
	saveg_write32(str.Ftype1)
	// boolean crush;
	saveg_write32(int32(str.Fcrush))
	// sector_t* sector;
	saveg_write32(sectorIndex(str.Fsector))
	// int direction;
	saveg_write32(str.Fdirection)
	// int newspecial;
	saveg_write32(str.Fnewspecial)
	// short texture;
	saveg_write16(str.Ftexture)
	// fixed_t floordestheight;
	saveg_write32(str.Ffloordestheight)
	// fixed_t speed;
	saveg_write32(str.Fspeed)
}

//
// plat_t
//

func saveg_read_plat_t(str *plat_t) {
	var sector int32
	// thinker_t thinker;
	saveg_read_thinker_t(&str.Fthinker)
	// sector_t* sector;
	sector = saveg_read32()
	str.Fsector = &sectors[sector]
	// fixed_t speed;
	str.Fspeed = saveg_read32()
	// fixed_t low;
	str.Flow = saveg_read32()
	// fixed_t high;
	str.Fhigh = saveg_read32()
	// int wait;
	str.Fwait = saveg_read32()
	// int count;
	str.Fcount = saveg_read32()
	// plat_e status;
	str.Fstatus = saveg_read32()
	// plat_e oldstatus;
	str.Foldstatus = saveg_read32()
	// boolean crush;
	str.Fcrush = uint32(saveg_read32())
	// int tag;
	str.Ftag = saveg_read32()
	// plattype_e type;
	str.Ftype1 = saveg_read32()
}

func saveg_write_plat_t(str *plat_t) {
	// thinker_t thinker;
	saveg_write_thinker_t(&str.Fthinker)
	// sector_t* sector;
	saveg_write32(sectorIndex(str.Fsector))
	// fixed_t speed;
	saveg_write32(str.Fspeed)
	// fixed_t low;
	saveg_write32(str.Flow)
	// fixed_t high;
	saveg_write32(str.Fhigh)
	// int wait;
	saveg_write32(str.Fwait)
	// int count;
	saveg_write32(str.Fcount)
	// plat_e status;
	saveg_write32(str.Fstatus)
	// plat_e oldstatus;
	saveg_write32(str.Foldstatus)
	// boolean crush;
	saveg_write32(int32(str.Fcrush))
	// int tag;
	saveg_write32(str.Ftag)
	// plattype_e type;
	saveg_write32(str.Ftype1)
}

//
// lightflash_t
//

func saveg_read_lightflash_t(str *lightflash_t) {
	var sector int32
	// thinker_t thinker;
	saveg_read_thinker_t(&str.Fthinker)
	// sector_t* sector;
	sector = saveg_read32()
	str.Fsector = &sectors[sector]
	// int count;
	str.Fcount = saveg_read32()
	// int maxlight;
	str.Fmaxlight = saveg_read32()
	// int minlight;
	str.Fminlight = saveg_read32()
	// int maxtime;
	str.Fmaxtime = saveg_read32()
	// int mintime;
	str.Fmintime = saveg_read32()
}

func saveg_write_lightflash_t(str *lightflash_t) {
	// thinker_t thinker;
	saveg_write_thinker_t(&str.Fthinker)
	// sector_t* sector;
	saveg_write32(sectorIndex(str.Fsector))
	// int count;
	saveg_write32(str.Fcount)
	// int maxlight;
	saveg_write32(str.Fmaxlight)
	// int minlight;
	saveg_write32(str.Fminlight)
	// int maxtime;
	saveg_write32(str.Fmaxtime)
	// int mintime;
	saveg_write32(str.Fmintime)
}

//
// strobe_t
//

func saveg_read_strobe_t(str *strobe_t) {
	var sector int32
	// thinker_t thinker;
	saveg_read_thinker_t(&str.Fthinker)
	// sector_t* sector;
	sector = saveg_read32()
	str.Fsector = &sectors[sector]
	// int count;
	str.Fcount = saveg_read32()
	// int minlight;
	str.Fminlight = saveg_read32()
	// int maxlight;
	str.Fmaxlight = saveg_read32()
	// int darktime;
	str.Fdarktime = saveg_read32()
	// int brighttime;
	str.Fbrighttime = saveg_read32()
}

func saveg_write_strobe_t(str *strobe_t) {
	// thinker_t thinker;
	saveg_write_thinker_t(&str.Fthinker)
	// sector_t* sector;
	saveg_write32(sectorIndex(str.Fsector))
	// int count;
	saveg_write32(str.Fcount)
	// int minlight;
	saveg_write32(str.Fminlight)
	// int maxlight;
	saveg_write32(str.Fmaxlight)
	// int darktime;
	saveg_write32(str.Fdarktime)
	// int brighttime;
	saveg_write32(str.Fbrighttime)
}

//
// glow_t
//

func saveg_read_glow_t(str *glow_t) {
	var sector int32
	// thinker_t thinker;
	saveg_read_thinker_t(&str.Fthinker)
	// sector_t* sector;
	sector = saveg_read32()
	str.Fsector = &sectors[sector]
	// int minlight;
	str.Fminlight = saveg_read32()
	// int maxlight;
	str.Fmaxlight = saveg_read32()
	// int direction;
	str.Fdirection = saveg_read32()
}

func saveg_write_glow_t(str *glow_t) {
	// thinker_t thinker;
	saveg_write_thinker_t(&str.Fthinker)
	// sector_t* sector;
	saveg_write32(sectorIndex(str.Fsector))
	// int minlight;
	saveg_write32(str.Fminlight)
	// int maxlight;
	saveg_write32(str.Fmaxlight)
	// int direction;
	saveg_write32(str.Fdirection)
}

//
// Write the header for a savegame
//

func p_WriteSaveGameHeader(description string) {
	var i int
	for i = 0; i < SAVESTRINGSIZE; i++ {
		if i < len(description) {
			saveg_write8(uint8(description[i]))
		} else {
			saveg_write8(0)
		}
	}
	bp := fmt.Sprintf("version %d", g_VanillaVersionCode())
	i = 0
	for {
		if i >= VERSIONSIZE {
			break
		}
		if i < len(bp) {
			saveg_write8(uint8(bp[i]))
		} else {
			saveg_write8(0)
		}
		i++
	}
	saveg_write8(uint8(gameskill))
	saveg_write8(uint8(gameepisode))
	saveg_write8(uint8(gamemap))
	i = 0
	for {
		if i >= MAXPLAYERS {
			break
		}
		saveg_write8(uint8(playeringame[i]))
		goto _4
	_4:
		;
		i++
	}
	saveg_write8(uint8(leveltime >> int32(16) & 0xff))
	saveg_write8(uint8(leveltime >> 8 & 0xff))
	saveg_write8(uint8(leveltime & 0xff))
}

//
// Read the header for a savegame
//

func p_ReadSaveGameHeader() boolean {
	var a, b, c uint8
	var i int32
	// skip the description field
	i = 0
	for {
		if i >= SAVESTRINGSIZE {
			break
		}
		saveg_read8()
		goto _1
	_1:
		;
		i++
	}
	i = 0
	var bp [VERSIONSIZE]byte
	for {
		if i >= VERSIONSIZE {
			break
		}
		bp[i] = saveg_read8()
		i++
	}
	vanilla := fmt.Sprintf("version %d", g_VanillaVersionCode())
	if vanilla != gostring_bytes(bp[:]) {
		return 0
	} // bad version
	gameskill = skill_t(saveg_read8())
	gameepisode = int32(saveg_read8())
	gamemap = int32(saveg_read8())
	i = 0
	for {
		if i >= MAXPLAYERS {
			break
		}
		playeringame[i] = uint32(saveg_read8())
		goto _3
	_3:
		;
		i++
	}
	// get the times
	a = saveg_read8()
	b = saveg_read8()
	c = saveg_read8()
	leveltime = int32(a)<<int32(16) + int32(b)<<8 + int32(c)
	return 1
}

//
// Read the end of file marker.  Returns true if read successfully.
//

func p_ReadSaveGameEOF() boolean {
	var value int32
	value = int32(saveg_read8())
	return booluint32(value == SAVEGAME_EOF)
}

//
// Write the end of file marker
//

func p_WriteSaveGameEOF() {
	saveg_write8(uint8(SAVEGAME_EOF))
}

// C documentation
//
//	//
//	// P_ArchivePlayers
//	//
func p_ArchivePlayers() {
	var i int32
	i = 0
	for {
		if i >= MAXPLAYERS {
			break
		}
		if playeringame[i] == 0 {
			goto _1
		}
		saveg_write_pad()
		saveg_write_player_t(&players[i])
		goto _1
	_1:
		;
		i++
	}
}

// C documentation
//
//	//
//	// P_UnArchivePlayers
//	//
func p_UnArchivePlayers() {
	for i := 0; i < MAXPLAYERS; i++ {
		if playeringame[i] == 0 {
			continue
		}
		saveg_read_pad()
		saveg_read_player_t(&players[i])
		// will be set when unarc thinker
		players[i].Fmo = nil
		players[i].Fmessage = ""
		players[i].Fattacker = nil
	}
}

// C documentation
//
//	//
//	// P_ArchiveWorld
//	//
func p_ArchiveWorld() {
	var si *side_t
	// do sectors
	for i := int32(0); i < numsectors; i++ {
		sec := &sectors[i]
		saveg_write16(int16(sec.Ffloorheight >> FRACBITS))
		saveg_write16(int16(sec.Fceilingheight >> FRACBITS))
		saveg_write16(sec.Ffloorpic)
		saveg_write16(sec.Fceilingpic)
		saveg_write16(sec.Flightlevel)
		saveg_write16(sec.Fspecial) // needed?
		saveg_write16(sec.Ftag)     // needed?
	}
	// do lines
	for i := int32(0); i < numlines; i++ {
		li := &lines[i]
		saveg_write16(li.Fflags)
		saveg_write16(li.Fspecial)
		saveg_write16(li.Ftag)
		for j := 0; j < 2; j++ {
			if li.Fsidenum[j] == -1 {
				continue
			}
			si = &sides[li.Fsidenum[j]]
			saveg_write16(int16(si.Ftextureoffset >> FRACBITS))
			saveg_write16(int16(si.Frowoffset >> FRACBITS))
			saveg_write16(si.Ftoptexture)
			saveg_write16(si.Fbottomtexture)
			saveg_write16(si.Fmidtexture)
		}
	}
}

// C documentation
//
//	//
//	// P_UnArchiveWorld
//	//
func p_UnArchiveWorld() {
	var si *side_t
	// do sectors
	for i := int32(0); i < numsectors; i++ {
		sec := &sectors[i]
		sec.Ffloorheight = int32(saveg_read16()) << FRACBITS
		sec.Fceilingheight = int32(saveg_read16()) << FRACBITS
		sec.Ffloorpic = saveg_read16()
		sec.Fceilingpic = saveg_read16()
		sec.Flightlevel = saveg_read16()
		sec.Fspecial = saveg_read16() // needed?
		sec.Ftag = saveg_read16()     // needed?
		sec.Fspecialdata = nil
		sec.Fsoundtarget = nil
	}
	// do lines
	for i := int32(0); i < numlines; i++ {
		li := &lines[i]
		li.Fflags = saveg_read16()
		li.Fspecial = saveg_read16()
		li.Ftag = saveg_read16()
		for j := 0; j < 2; j++ {
			if li.Fsidenum[j] == -1 {
				continue
			}
			si = &sides[li.Fsidenum[j]]
			si.Ftextureoffset = int32(saveg_read16()) << FRACBITS
			si.Frowoffset = int32(saveg_read16()) << FRACBITS
			si.Ftoptexture = saveg_read16()
			si.Fbottomtexture = saveg_read16()
			si.Fmidtexture = saveg_read16()
		}
	}
}

const tc_end = 0
const tc_mobj = 1

// C documentation
//
//	//
//	// P_ArchiveThinkers
//	//
func p_ArchiveThinkers() {
	var th *thinker_t
	// save off the current thinkers
	th = thinkercap.Fnext
	for {
		if th == &thinkercap {
			break
		}
		if mo, ok := th.Ffunction.(*mobj_t); ok {
			saveg_write8(uint8(tc_mobj))
			saveg_write_pad()
			saveg_write_mobj_t(mo)
			goto _1
		}
		// i_Error ("p_ArchiveThinkers: Unknown thinker function");
		goto _1
	_1:
		;
		th = th.Fnext
	}
	// add a terminating marker
	saveg_write8(uint8(tc_end))
}

// C documentation
//
//	//
//	// P_UnArchiveThinkers
//	//
func p_UnArchiveThinkers() {
	var currentthinker, next *thinker_t
	var mobj *mobj_t
	var tclass uint8
	// remove all the current thinkers
	currentthinker = thinkercap.Fnext
	for currentthinker != &thinkercap {
		next = currentthinker.Fnext
		if mo, ok := currentthinker.Ffunction.(*mobj_t); ok {
			p_RemoveMobj(mo)
		} else {
			//z_Free(uintptr(unsafe.Pointer(currentthinker)))
		}
		currentthinker = next
	}
	p_InitThinkers()
	// read in saved thinkers
	for 1 != 0 {
		tclass = saveg_read8()
		switch int32(tclass) {
		case tc_end:
			return // end of list
		case tc_mobj:
			saveg_read_pad()
			mobj = &mobj_t{}
			saveg_read_mobj_t(mobj)
			mobj.Ftarget = nil
			mobj.Ftracer = nil
			p_SetThingPosition(mobj)
			mobj.Finfo = &mobjinfo[mobj.Ftype1]
			mobj.Ffloorz = mobj.Fsubsector.Fsector.Ffloorheight
			mobj.Fceilingz = mobj.Fsubsector.Fsector.Fceilingheight
			mobj.Fthinker.Ffunction = mobj
			p_AddThinker(&mobj.Fthinker)
		default:
			i_Error("Unknown tclass %d in savegame", tclass)
		}
	}
}

const tc_ceiling = 0
const tc_door = 1
const tc_floor = 2
const tc_plat = 3
const tc_flash = 4
const tc_strobe = 5
const tc_glow = 6
const tc_endspecials = 7

// C documentation
//
//	//
//	// Things to handle:
//	//
//	// t_MoveCeiling, (ceiling_t: sector_t * swizzle), - active list
//	// t_VerticalDoor, (vldoor_t: sector_t * swizzle),
//	// t_MoveFloor, (floormove_t: sector_t * swizzle),
//	// t_LightFlash, (lightflash_t: sector_t * swizzle),
//	// t_StrobeFlash, (strobe_t: sector_t *),
//	// t_Glow, (glow_t: sector_t *),
//	// t_PlatRaise, (plat_t: sector_t *), - active list
//	//
func p_ArchiveSpecials() {
	var i int32
	var th *thinker_t
	// save off the current thinkers
	th = thinkercap.Fnext
	for {
		if th == &thinkercap {
			break
		}
		if th.Ffunction == nil {
			i = 0
			for {
				if i >= MAXCEILINGS {
					break
				}
				if &activeceilings[i].Fthinker == th {
					break
				}
				goto _2
			_2:
				;
				i++
			}
			if i < MAXCEILINGS {
				saveg_write8(uint8(tc_ceiling))
				saveg_write_pad()
				saveg_write_ceiling_t((*ceiling_t)(unsafe.Pointer(th)))
			}
			goto _1
		}
		if ceiling, ok := th.Ffunction.(*ceiling_t); ok {
			saveg_write8(uint8(tc_ceiling))
			saveg_write_pad()
			saveg_write_ceiling_t(ceiling)
			goto _1
		}
		if vldoor, ok := th.Ffunction.(*vldoor_t); ok {
			saveg_write8(uint8(tc_door))
			saveg_write_pad()
			saveg_write_vldoor_t(vldoor)
			goto _1
		}
		if floor, ok := th.Ffunction.(*floormove_t); ok {
			saveg_write8(uint8(tc_floor))
			saveg_write_pad()
			saveg_write_floormove_t(floor)
			goto _1
		}
		if plat, ok := th.Ffunction.(*plat_t); ok {
			saveg_write8(uint8(tc_plat))
			saveg_write_pad()
			saveg_write_plat_t(plat)
			goto _1
		}
		if light, ok := th.Ffunction.(*lightflash_t); ok {
			saveg_write8(uint8(tc_flash))
			saveg_write_pad()
			saveg_write_lightflash_t(light)
			goto _1
		}
		if strobe, ok := th.Ffunction.(*strobe_t); ok {
			saveg_write8(uint8(tc_strobe))
			saveg_write_pad()
			saveg_write_strobe_t(strobe)
			goto _1
		}
		if glow, ok := th.Ffunction.(*glow_t); ok {
			saveg_write8(uint8(tc_glow))
			saveg_write_pad()
			saveg_write_glow_t(glow)
			goto _1
		}
		goto _1
	_1:
		;
		th = th.Fnext
	}
	// add a terminating marker
	saveg_write8(uint8(tc_endspecials))
}

// C documentation
//
//	//
//	// P_UnArchiveSpecials
//	//
func p_UnArchiveSpecials() {
	var tclass uint8
	// read in saved thinkers
	for 1 != 0 {
		tclass = saveg_read8()
		switch int32(tclass) {
		case tc_endspecials:
			return // end of list
		case tc_ceiling:
			saveg_read_pad()
			ceilingP := &ceiling_t{}
			saveg_read_ceiling_t(ceilingP)
			ceilingP.Fsector.Fspecialdata = ceilingP
			if ceilingP.Fthinker.Ffunction != nil {
				ceilingP.Fthinker.Ffunction = ceilingP
			}
			p_AddThinker(&ceilingP.Fthinker)
			p_AddActiveCeiling(ceilingP)
		case tc_door:
			saveg_read_pad()
			doorP := &vldoor_t{}
			saveg_read_vldoor_t(doorP)
			doorP.Fsector.Fspecialdata = doorP
			doorP.Fthinker.Ffunction = doorP
			p_AddThinker(&doorP.Fthinker)
		case tc_floor:
			saveg_read_pad()
			floorP := &floormove_t{}
			saveg_read_floormove_t(floorP)
			floorP.Fsector.Fspecialdata = floorP
			floorP.Fthinker.Ffunction = floorP
			p_AddThinker(&floorP.Fthinker)
		case tc_plat:
			saveg_read_pad()
			platP := &plat_t{}
			saveg_read_plat_t(platP)
			platP.Fsector.Fspecialdata = platP
			if platP.Fthinker.Ffunction != nil {
				platP.Fthinker.Ffunction = platP
			}
			p_AddThinker(&platP.Fthinker)
			p_AddActivePlat(platP)
		case tc_flash:
			saveg_read_pad()
			flashP := &lightflash_t{}
			saveg_read_lightflash_t(flashP)
			flashP.Fthinker.Ffunction = flashP
			p_AddThinker(&flashP.Fthinker)
		case tc_strobe:
			saveg_read_pad()
			strobeP := &strobe_t{}
			saveg_read_strobe_t(strobeP)
			strobeP.Fthinker.Ffunction = strobeP
			p_AddThinker(&strobeP.Fthinker)
		case tc_glow:
			saveg_read_pad()
			glowP := &glow_t{}
			saveg_read_glow_t(glowP)
			glowP.Fthinker.Ffunction = glowP
			p_AddThinker(&glowP.Fthinker)
		default:
			i_Error("P_UnarchiveSpecials:Unknown tclass %d in savegame", tclass)
		}
	}
}

var totallines int32

// C documentation
//
//	//
//	// P_LoadVertexes
//	//
func p_LoadVertexes(lump int32) {
	var data uintptr
	// Determine number of lumps:
	//  total lump length / vertex record length.
	numvertexes = int32(uint64(w_LumpLength(uint32(lump))) / 4)
	// Allocate zone memory for buffer.
	vertexes = make([]vertex_t, numvertexes)
	// Load data into cache.
	data = w_CacheLumpNum(lump)
	ml := unsafe.Slice((*mapvertex_t)(unsafe.Pointer(data)), numvertexes)
	// Copy and convert vertex coordinates,
	// internal representation as fixed.
	for i := int32(0); i < numvertexes; i++ {
		li := &vertexes[i]
		li.Fx = int32(ml[i].Fx) << FRACBITS
		li.Fy = int32(ml[i].Fy) << FRACBITS
	}
	// Free buffer memory.
	w_ReleaseLumpNum(lump)
}

// C documentation
//
//	//
//	// GetSectorAtNullAddress
//	//
func getSectorAtNullAddress() (r *sector_t) {
	if null_sector_is_initialized == 0 {
		null_sector = sector_t{}
		i_GetMemoryValue(0, uintptr(unsafe.Pointer(&null_sector)), 4)
		i_GetMemoryValue(4, uintptr(unsafe.Pointer(&null_sector))+4, 4)
		null_sector_is_initialized = 1
	}
	return &null_sector
}

var null_sector_is_initialized boolean

var null_sector sector_t

// C documentation
//
//	//
//	// P_LoadSegs
//	//
func p_LoadSegs(lump int32) {
	var data uintptr
	var ldef *line_t
	var linedef, side, sidenum int32
	numsegs = int32(uint64(w_LumpLength(uint32(lump))) / 12)
	segs = make([]seg_t, numsegs)
	data = w_CacheLumpNum(lump)
	ml := unsafe.Slice((*mapseg_t)(unsafe.Pointer(data)), numsegs)
	for i := int32(0); i < numsegs; i++ {
		li := &segs[i]
		li.Fv1 = &vertexes[ml[i].Fv1]
		li.Fv2 = &vertexes[ml[i].Fv2]
		li.Fangle = uint32(int32(ml[i].Fangle) << 16)
		li.Foffset = int32(ml[i].Foffset) << 16
		linedef = int32(ml[i].Flinedef)
		ldef = &lines[linedef]
		li.Flinedef = ldef
		side = int32(ml[i].Fside)
		li.Fsidedef = &sides[ldef.Fsidenum[side]]
		li.Ffrontsector = sides[ldef.Fsidenum[side]].Fsector
		if int32(ldef.Fflags)&ml_TWOSIDED != 0 {
			sidenum = int32(ldef.Fsidenum[side^1])
			// If the sidenum is out of range, this may be a "glass hack"
			// impassible window.  Point at side #0 (this may not be
			// the correct Vanilla behavior; however, it seems to work for
			// OTTAWAU.WAD, which is the one place I've seen this trick
			// used).
			if sidenum < 0 || sidenum >= numsides {
				li.Fbacksector = getSectorAtNullAddress()
			} else {
				li.Fbacksector = sides[sidenum].Fsector
			}
		} else {
			li.Fbacksector = nil
		}
	}
	w_ReleaseLumpNum(lump)
}

// C documentation
//
//	//
//	// P_LoadSubsectors
//	//
func p_LoadSubsectors(lump int32) {
	var data uintptr
	numsubsectors = int32(uint64(w_LumpLength(uint32(lump))) / 4)
	subsectors = make([]subsector_t, numsubsectors)
	data = w_CacheLumpNum(lump)
	ms := unsafe.Slice((*mapsubsector_t)(unsafe.Pointer(data)), numsubsectors)
	for i := int32(0); i < numsubsectors; i++ {
		subsectors[i].Fnumlines = ms[i].Fnumsegs
		subsectors[i].Ffirstline = ms[i].Ffirstseg
	}
	w_ReleaseLumpNum(lump)
}

// C documentation
//
//	//
//	// P_LoadSectors
//	//
func p_LoadSectors(lump int32) {
	var data uintptr
	numsectors = int32(uint64(w_LumpLength(uint32(lump))) / uint64(unsafe.Sizeof(mapsector_t{})))
	sectors = make([]sector_t, numsectors)
	data = w_CacheLumpNum(lump)
	mapsectors := unsafe.Slice((*mapsector_t)(unsafe.Pointer(data)), numsectors)
	for i := int32(0); i < numsectors; i++ {
		ms := &mapsectors[i]
		ss := &sectors[i]
		ss.Ffloorheight = int32(ms.Ffloorheight) << FRACBITS
		ss.Fceilingheight = int32(ms.Fceilingheight) << FRACBITS
		ss.Ffloorpic = int16(r_FlatNumForName(gostring_bytes(ms.Ffloorpic[:])))
		ss.Fceilingpic = int16(r_FlatNumForName(gostring_bytes(ms.Fceilingpic[:])))
		ss.Flightlevel = ms.Flightlevel
		ss.Fspecial = ms.Fspecial
		ss.Ftag = ms.Ftag
		ss.Fthinglist = nil
	}
	w_ReleaseLumpNum(lump)
}

// C documentation
//
//	//
//	// P_LoadNodes
//	//
func p_LoadNodes(lump int32) {
	var data uintptr
	numnodes = int32(uint64(w_LumpLength(uint32(lump))) / uint64(unsafe.Sizeof(mapnode_t{})))
	nodes = make([]node_t, numnodes)
	data = w_CacheLumpNum(lump)
	mapnodes := unsafe.Slice((*mapnode_t)(unsafe.Pointer(data)), numnodes)
	for i := 0; i < int(numnodes); i++ {
		no := &nodes[i]
		mn := &mapnodes[i]
		no.Fx = int32(mn.Fx) << FRACBITS
		no.Fy = int32(mn.Fy) << FRACBITS
		no.Fdx = int32(mn.Fdx) << FRACBITS
		no.Fdy = int32(mn.Fdy) << FRACBITS
		for j := 0; j < 2; j++ {
			no.Fchildren[j] = mn.Fchildren[j]
			for k := 0; k < 4; k++ {
				no.Fbbox[j][k] = int32(mn.Fbbox[j][k]) << FRACBITS
			}
		}
	}
	w_ReleaseLumpNum(lump)
}

// C documentation
//
//	//
//	// P_LoadThings
//	//
func p_LoadThings(lump int32) {
	var data uintptr
	var numthings int32
	var spawn boolean
	data = w_CacheLumpNum(lump)
	numthings = int32(uint64(w_LumpLength(uint32(lump))) / uint64(unsafe.Sizeof(mapthing_t{})))
	mthings := unsafe.Slice((*mapthing_t)(unsafe.Pointer(data)), numthings)
	for i := int32(0); i < numthings; i++ {
		mt := &mthings[i]
		spawn = 1
		// Do not spawn cool, new monsters if !commercial
		if gamemode != commercial {
			switch int32(mt.Ftype1) {
			case 68: // Arachnotron
				fallthrough
			case 64: // Archvile
				fallthrough
			case 88: // Boss Brain
				fallthrough
			case 89: // Boss Shooter
				fallthrough
			case 69: // Hell Knight
				fallthrough
			case 67: // Mancubus
				fallthrough
			case 71: // Pain Elemental
				fallthrough
			case 65: // Former Human Commando
				fallthrough
			case 66: // Revenant
				fallthrough
			case 84: // Wolf SS
				spawn = 0
				break
			}
		}
		if spawn == 0 {
			break
		}
		// Do spawn all other stuff.
		bp := &mapthing_t{
			Fx:       mt.Fx,
			Fy:       mt.Fy,
			Fangle:   mt.Fangle,
			Ftype1:   mt.Ftype1,
			Foptions: mt.Foptions,
		}
		p_SpawnMapThing(bp)
	}
	w_ReleaseLumpNum(lump)
}

// C documentation
//
//	//
//	// P_LoadLineDefs
//	// Also counts secret lines for intermissions.
//	//
func p_LoadLineDefs(lump int32) {
	var data uintptr
	var v1, v2, v21, v3 *vertex_t
	var i int32
	numlines = int32(uint64(w_LumpLength(uint32(lump))) / 14)
	lines = make([]line_t, numlines)
	data = w_CacheLumpNum(lump)
	ml := unsafe.Slice((*maplinedef_t)(unsafe.Pointer(data)), numlines)
	for i = 0; i < numlines; i++ {
		ld := &lines[i]
		mld := ml[i]
		ld.Fflags = mld.Fflags
		ld.Fspecial = mld.Fspecial
		ld.Ftag = mld.Ftag
		v21 = &vertexes[mld.Fv1]
		ld.Fv1 = v21
		v1 = v21
		v3 = &vertexes[mld.Fv2]
		ld.Fv2 = v3
		v2 = v3
		ld.Fdx = v2.Fx - v1.Fx
		ld.Fdy = v2.Fy - v1.Fy
		if ld.Fdx == 0 {
			ld.Fslopetype = st_VERTICAL
		} else {
			if ld.Fdy == 0 {
				ld.Fslopetype = st_HORIZONTAL
			} else {
				if fixedDiv(ld.Fdy, ld.Fdx) > 0 {
					ld.Fslopetype = st_POSITIVE
				} else {
					ld.Fslopetype = st_NEGATIVE
				}
			}
		}
		if v1.Fx < v2.Fx {
			ld.Fbbox[BOXLEFT] = v1.Fx
			ld.Fbbox[BOXRIGHT] = v2.Fx
		} else {
			ld.Fbbox[BOXLEFT] = v2.Fx
			ld.Fbbox[BOXRIGHT] = v1.Fx
		}
		if v1.Fy < v2.Fy {
			ld.Fbbox[BOXBOTTOM] = v1.Fy
			ld.Fbbox[BOXTOP] = v2.Fy
		} else {
			ld.Fbbox[BOXBOTTOM] = v2.Fy
			ld.Fbbox[BOXTOP] = v1.Fy
		}
		ld.Fsidenum[0] = mld.Fsidenum[0]
		ld.Fsidenum[1] = mld.Fsidenum[1]
		if ld.Fsidenum[0] != -1 {
			ld.Ffrontsector = sides[ld.Fsidenum[0]].Fsector
		} else {
			ld.Ffrontsector = nil
		}
		if ld.Fsidenum[1] != -1 {
			ld.Fbacksector = sides[ld.Fsidenum[1]].Fsector
		} else {
			ld.Fbacksector = nil
		}
	}
	w_ReleaseLumpNum(lump)
}

// C documentation
//
//	//
//	// P_LoadSideDefs
//	//
func p_LoadSideDefs(lump int32) {
	var data uintptr
	numsides = int32(uint64(w_LumpLength(uint32(lump))) / 30)
	sides = make([]side_t, numsides)
	data = w_CacheLumpNum(lump)
	msd := unsafe.Slice((*mapsidedef_t)(unsafe.Pointer(data)), numsides)
	for i := int32(0); i < numsides; i++ {
		sd := &sides[i]
		sd.Ftextureoffset = int32(msd[i].Ftextureoffset) << FRACBITS
		sd.Frowoffset = int32(msd[i].Frowoffset) << FRACBITS
		sd.Ftoptexture = int16(r_TextureNumForName(gostring_bytes(msd[i].Ftoptexture[:])))
		sd.Fbottomtexture = int16(r_TextureNumForName(gostring_bytes(msd[i].Fbottomtexture[:])))
		sd.Fmidtexture = int16(r_TextureNumForName(gostring_bytes(msd[i].Fmidtexture[:])))
		sd.Fsector = &sectors[msd[i].Fsector]
	}
	w_ReleaseLumpNum(lump)
}

// C documentation
//
//	//
//	// P_LoadBlockMap
//	//
func p_LoadBlockMap(lump int32) {
	rawLump := w_ReadLumpBytes(uint32(lump))
	blockmaplump = make([]int16, len(rawLump)/2)
	for i := 0; i < len(rawLump); i += 2 {
		blockmaplump[i/2] = int16(rawLump[i]) | int16(rawLump[i+1])<<8
	}
	blockmap = blockmaplump[4:]
	// Swap all short integers to native byte ordering.
	// TODO: GORE: We've lost endian fixes here
	// Read the header
	bmaporgx = int32(blockmaplump[0]) << FRACBITS
	bmaporgy = int32(blockmaplump[1]) << FRACBITS
	bmapwidth = int32(blockmaplump[2])
	bmapheight = int32(blockmaplump[3])
	// Clear out mobj chains
	count := int32(uint64(bmapwidth) * uint64(bmapheight))
	blocklinks = make([]*mobj_t, count)
}

// C documentation
//
//	//
//	// P_GroupLines
//	// Builds sector line lists and subsector sector numbers.
//	// Finds block bounding boxes for sectors.
//	//
func p_GroupLines() {
	var block, i, j, v10, v7, v8, v9 int32
	// look up sector number for each subsector
	for i := int32(0); i < numsubsectors; i++ {
		ss := &subsectors[i]
		seg := &segs[ss.Ffirstline]
		ss.Fsector = seg.Fsidedef.Fsector
	}
	// count number of lines in each sector
	totallines = 0
	for i := int32(0); i < numlines; i++ {
		li := &lines[i]
		totallines++
		li.Ffrontsector.Flinecount++
		if li.Fbacksector != nil && li.Fbacksector != li.Ffrontsector {
			li.Fbacksector.Flinecount++
			totallines++
		}
	}
	// build line tables for each sector
	i = 0
	for {
		if i >= numsectors {
			break
		}
		// Assign the line buffer for this sector
		sector := &sectors[i]
		sector.Flines = make([]*line_t, sector.Flinecount)
		// Reset linecount to zero so in the next stage we can count
		// lines into the list.
		sector.Flinecount = 0
		goto _3
	_3:
		;
		i++
	}
	// Assign lines to sectors
	for i := int32(0); i < numlines; i++ {
		li := &lines[i]
		if li.Ffrontsector != nil {
			li.Ffrontsector.Flines[li.Ffrontsector.Flinecount] = li
			li.Ffrontsector.Flinecount++
		}
		if li.Fbacksector != nil && li.Ffrontsector != li.Fbacksector {
			li.Fbacksector.Flines[li.Fbacksector.Flinecount] = li
			li.Fbacksector.Flinecount++
		}
	}
	// Generate bounding boxes for sectors
	i = 0
	for {
		if i >= numsectors {
			break
		}
		var box box_t
		sector := &sectors[i]
		m_ClearBox(&box)
		j = 0
		for {
			if j >= sector.Flinecount {
				break
			}
			li := sector.Flines[j]
			m_AddToBox(&box, li.Fv1.Fx, li.Fv1.Fy)
			m_AddToBox(&box, li.Fv2.Fx, li.Fv2.Fy)
			goto _6
		_6:
			;
			j++
		}
		// set the degenmobj_t to the middle of the bounding box
		sector.Fsoundorg.Fx = (box[BOXRIGHT] + box[BOXLEFT]) / 2
		sector.Fsoundorg.Fy = (box[BOXTOP] + box[BOXBOTTOM]) / 2
		// adjust bounding box to map blocks
		block = (box[BOXTOP] - bmaporgy + 32*(1<<FRACBITS)) >> (FRACBITS + 7)
		if block >= bmapheight {
			v7 = bmapheight - 1
		} else {
			v7 = block
		}
		block = v7
		sector.Fblockbox[BOXTOP] = block
		block = (box[BOXBOTTOM] - bmaporgy - 32*(1<<FRACBITS)) >> (FRACBITS + 7)
		if block < 0 {
			v8 = 0
		} else {
			v8 = block
		}
		block = v8
		sector.Fblockbox[BOXBOTTOM] = block
		block = (box[BOXRIGHT] - bmaporgx + 32*(1<<FRACBITS)) >> (FRACBITS + 7)
		if block >= bmapwidth {
			v9 = bmapwidth - 1
		} else {
			v9 = block
		}
		block = v9
		sector.Fblockbox[BOXRIGHT] = block
		block = (box[BOXLEFT] - bmaporgx - 32*(1<<FRACBITS)) >> (FRACBITS + 7)
		if block < 0 {
			v10 = 0
		} else {
			v10 = block
		}
		block = v10
		sector.Fblockbox[BOXLEFT] = block
		goto _5
	_5:
		;
		i++
	}
}

// Pad the REJECT lump with extra data when the lump is too small,
// to simulate a REJECT buffer overflow in Vanilla Doom.

func padRejectArray(array []uint8, len1 uint32) {
	var byte_num uint32
	var padvalue uint8
	var pos int
	var rejectpad [4]uint32
	// Values to pad the REJECT array with:
	rejectpad = [4]uint32{
		0: uint32((totallines*4+3) & ^3 + 24),
		2: 50,
		3: 0x1d4a11,
	}
	// Copy values from rejectpad into the destination array.
	for i := uint32(0); i < len1 && i < 16; i++ {
		byte_num = i % 4
		array[pos] = uint8(rejectpad[i/4] >> (byte_num * 8) & 0xff)
		pos++
	}
	// We only have a limited pad size.  Print a warning if the
	// REJECT lump is too small.
	if uint64(len1) > 16 {
		fprintf_ccgo(os.Stderr, "padRejectArray: REJECT lump too short to pad! (%d > %d)\n", len1, 16)
		// Pad remaining space with 0 (or 0xff, if specified on command line).
		if m_CheckParm("-reject_pad_with_ff") != 0 {
			padvalue = 0xff
		} else {
			padvalue = 0x00
		}
		for i := uint32(0); i < len1-16; i++ {
			array[16+i] = padvalue
		}
	}
}

func p_LoadReject(lumpnum int32) {
	var lumplen, minlength int32
	// Calculate the size that the REJECT lump *should* be.
	minlength = (numsectors*numsectors + 7) / 8
	// If the lump meets the minimum length, it can be loaded directly.
	// Otherwise, we need to allocate a buffer of the correct size
	// and pad it with appropriate data.
	lumplen = w_LumpLength(uint32(lumpnum))
	if lumplen >= minlength {
		data := w_CacheLumpNum(lumpnum)
		rejectmatrix = unsafe.Slice((*uint8)(unsafe.Pointer(data)), lumplen)
	} else {
		rejectmatrix = w_ReadLumpBytes(uint32(lumpnum))
		rejectmatrix = append(rejectmatrix, make([]uint8, minlength-lumplen)...)
		padRejectArray(rejectmatrix[lumplen:], uint32(minlength-lumplen))
	}
}

// C documentation
//
//	//
//	// P_SetupLevel
//	//
func p_SetupLevel(episode int32, map1 int32, playermask int32, skill skill_t) {
	var i, lumpnum, v1, v2, v3, v5, v6, v8 int32
	v3 = 0
	wminfo.Fmaxfrags = v3
	v2 = v3
	totalsecret = v2
	v1 = v2
	totalitems = v1
	totalkills = v1
	wminfo.Fpartime = 180
	i = 0
	for {
		if i >= MAXPLAYERS {
			break
		}
		v6 = 0
		players[i].Fitemcount = v6
		v5 = v6
		players[i].Fsecretcount = v5
		players[i].Fkillcount = v5
		goto _4
	_4:
		;
		i++
	}
	// Initial height of PointOfView
	// will be set by player think.
	players[consoleplayer].Fviewz = 1
	// Make sure all sounds are stopped before z_FreeTags.
	s_Start()
	// UNUSED W_Profile ();
	p_InitThinkers()
	// find map name
	var bp string
	if gamemode == commercial {
		if map1 < 10 {
			bp = fmt.Sprintf("map0%d", map1)
		} else {
			bp = fmt.Sprintf("map%d", map1)
		}
	} else {
		bp = string([]byte{'E', '0' + byte(episode), 'M', '0' + byte(map1)})
	}
	lumpnum = w_GetNumForName(bp)
	leveltime = 0
	// note: most of this ordering is important
	p_LoadBlockMap(lumpnum + ml_BLOCKMAP)
	p_LoadVertexes(lumpnum + ml_VERTEXES)
	p_LoadSectors(lumpnum + ml_SECTORS)
	p_LoadSideDefs(lumpnum + ml_SIDEDEFS)
	p_LoadLineDefs(lumpnum + ml_LINEDEFS)
	p_LoadSubsectors(lumpnum + ml_SSECTORS)
	p_LoadNodes(lumpnum + ml_NODES)
	p_LoadSegs(lumpnum + ml_SEGS)
	p_GroupLines()
	p_LoadReject(lumpnum + ml_REJECT)
	bodyqueslot = 0
	deathmatch_pos = 0
	p_LoadThings(lumpnum + ml_THINGS)
	// if deathmatch, randomly spawn the active players
	if deathmatch != 0 {
		i = 0
		for {
			if i >= MAXPLAYERS {
				break
			}
			if playeringame[i] != 0 {
				players[i].Fmo = nil
				g_DeathMatchSpawnPlayer(i)
			}
			goto _7
		_7:
			;
			i++
		}
	}
	// clear special respawning que
	v8 = 0
	iquetail = v8
	iquehead = v8
	// set up world state
	p_SpawnSpecials()
	// build subsector connect matrix
	//	UNUSED P_ConnectSubsectors ();
	// preload graphics
	if precache != 0 {
		r_PrecacheLevel()
	}
	//printf ("free memory: 0x%x\n", Z_FreeMemory());
}

// C documentation
//
//	//
//	// P_Init
//	//
func p_Init() {
	p_InitSwitchList()
	p_InitPicAnims()
	r_InitSprites(sprnames)
}

const NF_SUBSECTOR1 = 32768

// C documentation
//
//	//
//	// P_DivlineSide
//	// Returns side 0 (front), 1 (back), or 2 (on).
//	//
func p_DivlineSide(x fixed_t, y fixed_t, node *divline_t) int32 {
	var dx, dy, left, right fixed_t
	if node.Fdx == 0 {
		if x == node.Fx {
			return 2
		}
		if x <= node.Fx {
			return boolint32(node.Fdy > 0)
		}
		return boolint32(node.Fdy < 0)
	}
	if node.Fdy == 0 {
		if x == node.Fy {
			return 2
		}
		if y <= node.Fy {
			return boolint32(node.Fdx < 0)
		}
		return boolint32(node.Fdx > 0)
	}
	dx = x - node.Fx
	dy = y - node.Fy
	left = node.Fdy >> FRACBITS * (dx >> FRACBITS)
	right = dy >> FRACBITS * (node.Fdx >> FRACBITS)
	if right < left {
		return 0
	} // front side
	if left == right {
		return 2
	}
	return 1 // back side
}

// C documentation
//
//	//
//	// P_InterceptVector2
//	// Returns the fractional intercept point
//	// along the first divline.
//	// This is only called by the addthings and addlines traversers.
//	//
func p_InterceptVector2(v2 *divline_t, v1 *divline_t) fixed_t {
	var den, frac, num fixed_t
	den = fixedMul(v1.Fdy>>8, v2.Fdx) - fixedMul(v1.Fdx>>8, v2.Fdy)
	if den == 0 {
		return 0
	}
	//	i_Error ("p_InterceptVector: parallel");
	num = fixedMul((v1.Fx-v2.Fx)>>8, v1.Fdy) + fixedMul((v2.Fy-v1.Fy)>>8, v1.Fdx)
	frac = fixedDiv(num, den)
	return frac
}

// C documentation
//
//	//
//	// P_CrossSubsector
//	// Returns true
//	//  if strace crosses the given subsector successfully.
//	//
func p_CrossSubsector(num int32) boolean {
	var divline divline_t
	var v1, v2 *vertex_t
	var line *line_t
	var front, back *sector_t
	var count, s1, s2 int32
	var frac, openbottom, opentop, slope fixed_t
	if num >= numsubsectors {
		i_Error("p_CrossSubsector: ss %d with numss = %d", num, numsubsectors)
	}
	sub := &subsectors[num]
	// check lines
	count = int32(sub.Fnumlines)
	for i := sub.Ffirstline; ; i++ {
		if count == 0 {
			break
		}
		seg := &segs[sub.Ffirstline]
		line = seg.Flinedef
		// allready checked other side?
		if line.Fvalidcount == validcount {
			goto _1
		}
		line.Fvalidcount = validcount
		v1 = line.Fv1
		v2 = line.Fv2
		s1 = p_DivlineSide(v1.Fx, v1.Fy, &strace)
		s2 = p_DivlineSide(v2.Fx, v2.Fy, &strace)
		// line isn't crossed?
		if s1 == s2 {
			goto _1
		}
		divline.Fx = v1.Fx
		divline.Fy = v1.Fy
		divline.Fdx = v2.Fx - v1.Fx
		divline.Fdy = v2.Fy - v1.Fy
		s1 = p_DivlineSide(strace.Fx, strace.Fy, &divline)
		s2 = p_DivlineSide(t2x, t2y, &divline)
		// line isn't crossed?
		if s1 == s2 {
			goto _1
		}
		// Backsector may be NULL if this is an "impassible
		// glass" hack line.
		if line.Fbacksector == nil {
			return 0
		}
		// stop because it is not two sided anyway
		// might do this after updating validcount?
		if int32(line.Fflags)&ml_TWOSIDED == 0 {
			return 0
		}
		// crosses a two sided line
		front = seg.Ffrontsector
		back = seg.Fbacksector
		// no wall to block sight with?
		if front.Ffloorheight == back.Ffloorheight && front.Fceilingheight == back.Fceilingheight {
			goto _1
		}
		// possible occluder
		// because of ceiling height differences
		if front.Fceilingheight < back.Fceilingheight {
			opentop = front.Fceilingheight
		} else {
			opentop = back.Fceilingheight
		}
		// because of ceiling height differences
		if front.Ffloorheight > back.Ffloorheight {
			openbottom = front.Ffloorheight
		} else {
			openbottom = back.Ffloorheight
		}
		// quick test for totally closed doors
		if openbottom >= opentop {
			return 0
		} // stop
		frac = p_InterceptVector2(&strace, &divline)
		if front.Ffloorheight != back.Ffloorheight {
			slope = fixedDiv(openbottom-sightzstart, frac)
			if slope > bottomslope {
				bottomslope = slope
			}
		}
		if front.Fceilingheight != back.Fceilingheight {
			slope = fixedDiv(opentop-sightzstart, frac)
			if slope < topslope {
				topslope = slope
			}
		}
		if topslope <= bottomslope {
			return 0
		} // stop
		goto _1
	_1:
		;
		count--
	}
	// passed the subsector ok
	return 1
}

// C documentation
//
//	//
//	// P_CrossBSPNode
//	// Returns true
//	//  if strace crosses the given node successfully.
//	//
func p_CrossBSPNode(bspnum int32) boolean {
	var side int32
	if bspnum&int32(NF_SUBSECTOR1) != 0 {
		if bspnum == -1 {
			return p_CrossSubsector(0)
		} else {
			return p_CrossSubsector(bspnum & ^NF_SUBSECTOR1)
		}
	}
	bsp := &nodes[bspnum]
	// decide which side the start point is on
	side = p_DivlineSide(strace.Fx, strace.Fy, &bsp.divline_t)
	if side == 2 {
		side = 0
	} // an "on" should cross both sides
	// cross the starting side
	if p_CrossBSPNode(int32(bsp.Fchildren[side])) == 0 {
		return 0
	}
	// the partition plane is crossed here
	if side == p_DivlineSide(t2x, t2y, &bsp.divline_t) {
		// the line doesn't touch the other side
		return 1
	}
	// cross the ending side
	return p_CrossBSPNode(int32(bsp.Fchildren[side^1]))
}

// C documentation
//
//	//
//	// P_CheckSight
//	// Returns true
//	//  if a straight line between t1 and t2 is unobstructed.
//	// Uses REJECT.
//	//
func p_CheckSight(t1 *mobj_t, t2 *mobj_t) boolean {
	var bitnum, bytenum, pnum, s1, s2 int32
	// First check for trivial rejection.
	// Determine subsector entries in REJECT table.
	s1 = sectorIndex(t1.Fsubsector.Fsector)
	s2 = sectorIndex(t2.Fsubsector.Fsector)
	pnum = s1*numsectors + s2
	bytenum = pnum >> 3
	bitnum = 1 << (pnum & 7)
	// Check in REJECT table.
	if int32(rejectmatrix[bytenum])&bitnum != 0 {
		sightcounts[0]++
		// can't possibly be connected
		return 0
	}
	// An unobstructed LOS is possible.
	// Now look from eyes of t1 to any part of t2.
	sightcounts[1]++
	validcount++
	sightzstart = t1.Fz + t1.Fheight - t1.Fheight>>2
	topslope = t2.Fz + t2.Fheight - sightzstart
	bottomslope = t2.Fz - sightzstart
	strace.Fx = t1.Fx
	strace.Fy = t1.Fy
	t2x = t2.Fx
	t2y = t2.Fy
	strace.Fdx = t2.Fx - t1.Fx
	strace.Fdy = t2.Fy - t1.Fy
	// the head node is the last node output
	return p_CrossBSPNode(numnodes - 1)
}

const DONUT_FLOORHEIGHT_DEFAULT = 0
const DONUT_FLOORPIC_DEFAULT = 22
const INT_MAX13 = 2147483647
const MAXLINEANIMS = 64
const MAX_ADJOINING_SECTORS = 20

// State.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// Data.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

// C documentation
//
//	//
//	// Animating textures and planes
//	// There is another anim_t used in wi_stuff, unrelated.
//	//
type anim_t struct {
	Fistexture boolean
	Fpicnum    int32
	Fbasepic   int32
	Fnumpics   int32
	Fspeed     int32
}

// C documentation
//
//	//
//	//      source animation definition
//	//
type animdef_t struct {
	Fistexture int32
	Fendname   string
	Fstartname string
	Fspeed     int32
}

func init() {
	animdefs = [23]animdef_t{
		0: {
			Fendname:   "NUKAGE3",
			Fstartname: "NUKAGE1",
			Fspeed:     8,
		},
		1: {
			Fendname:   "FWATER4",
			Fstartname: "FWATER1",
			Fspeed:     8,
		},
		2: {
			Fendname:   "SWATER4",
			Fstartname: "SWATER1",
			Fspeed:     8,
		},
		3: {
			Fendname:   "LAVA4",
			Fstartname: "LAVA1",
			Fspeed:     8,
		},
		4: {
			Fendname:   "BLOOD3",
			Fstartname: "BLOOD1",
			Fspeed:     8,
		},
		5: {
			Fendname:   "RROCK08",
			Fstartname: "RROCK05",
			Fspeed:     8,
		},
		6: {
			Fendname:   "SLIME04",
			Fstartname: "SLIME01",
			Fspeed:     8,
		},
		7: {
			Fendname:   "SLIME08",
			Fstartname: "SLIME05",
			Fspeed:     8,
		},
		8: {
			Fendname:   "SLIME12",
			Fstartname: "SLIME09",
			Fspeed:     8,
		},
		9: {
			Fistexture: 1,
			Fendname:   "BLODGR4",
			Fstartname: "BLODGR1",
			Fspeed:     8,
		},
		10: {
			Fistexture: 1,
			Fendname:   "SLADRIP3",
			Fstartname: "SLADRIP1",
			Fspeed:     8,
		},
		11: {
			Fistexture: 1,
			Fendname:   "BLODRIP4",
			Fstartname: "BLODRIP1",
			Fspeed:     8,
		},
		12: {
			Fistexture: 1,
			Fendname:   "FIREWALL",
			Fstartname: "FIREWALA",
			Fspeed:     8,
		},
		13: {
			Fistexture: 1,
			Fendname:   "GSTFONT3",
			Fstartname: "GSTFONT1",
			Fspeed:     8,
		},
		14: {
			Fistexture: 1,
			Fendname:   "FIRELAVA",
			Fstartname: "FIRELAV3",
			Fspeed:     8,
		},
		15: {
			Fistexture: 1,
			Fendname:   "FIREMAG3",
			Fstartname: "FIREMAG1",
			Fspeed:     8,
		},
		16: {
			Fistexture: 1,
			Fendname:   "FIREBLU2",
			Fstartname: "FIREBLU1",
			Fspeed:     8,
		},
		17: {
			Fistexture: 1,
			Fendname:   "ROCKRED3",
			Fstartname: "ROCKRED1",
			Fspeed:     8,
		},
		18: {
			Fistexture: 1,
			Fendname:   "BFALL4",
			Fstartname: "BFALL1",
			Fspeed:     8,
		},
		19: {
			Fistexture: 1,
			Fendname:   "SFALL4",
			Fstartname: "SFALL1",
			Fspeed:     8,
		},
		20: {
			Fistexture: 1,
			Fendname:   "WFALL4",
			Fstartname: "WFALL1",
			Fspeed:     8,
		},
		21: {
			Fistexture: 1,
			Fendname:   "DBRAIN4",
			Fstartname: "DBRAIN1",
			Fspeed:     8,
		},
		22: {
			Fistexture: -1,
		},
	}
}

func p_InitPicAnims() {
	var endname, startname string
	//	Init animation
	lastanim = &anims[0]
	animPos := 0
	for i := 0; ; i++ {
		if animdefs[i].Fistexture == -1 {
			break
		}
		startname = animdefs[i].Fstartname
		endname = animdefs[i].Fendname
		if animdefs[i].Fistexture != 0 {
			// different episode ?
			if r_CheckTextureNumForName(startname) == -1 {
				continue
			}
			lastanim.Fpicnum = r_TextureNumForName(endname)
			lastanim.Fbasepic = r_TextureNumForName(startname)
		} else {
			if w_CheckNumForName(startname) == -1 {
				continue
			}
			lastanim.Fpicnum = r_FlatNumForName(endname)
			lastanim.Fbasepic = r_FlatNumForName(startname)
		}
		lastanim.Fistexture = uint32(animdefs[i].Fistexture)
		lastanim.Fnumpics = lastanim.Fpicnum - lastanim.Fbasepic + 1
		if lastanim.Fnumpics < 2 {
			i_Error("p_InitPicAnims: bad cycle from %s to %s", startname, endname)
		}
		lastanim.Fspeed = animdefs[i].Fspeed
		animPos++
		lastanim = &anims[animPos]
	}
}

//
// UTILITIES
//

// C documentation
//
//	//
//	// getSide()
//	// Will return a side_t*
//	//  given the number of the current sector,
//	//  the line number, and the side (0/1) that you want.
//	//
func getSide(currentSector int32, line int32, side int32) *side_t {
	sec := &sectors[currentSector]
	linePtr := sec.Flines[line]
	return &sides[linePtr.Fsidenum[side]]
}

// C documentation
//
//	//
//	// getSector()
//	// Will return a sector_t*
//	//  given the number of the current sector,
//	//  the line number and the side (0/1) that you want.
//	//
func getSector(currentSector int32, line int32, side int32) (r *sector_t) {
	sidePtr := getSide(currentSector, line, side)
	return sidePtr.Fsector
}

// C documentation
//
//	//
//	// twoSided()
//	// Given the sector number and the line number,
//	//  it will tell you whether the line is two-sided or not.
//	//
func twoSided(sector int32, line int32) int32 {
	sec := &sectors[sector]
	return int32(sec.Flines[line].Fflags & ml_TWOSIDED)
}

// C documentation
//
//	//
//	// getNextSector()
//	// Return sector_t * of sector next to current.
//	// NULL if not two-sided line
//	//
func getNextSector(line *line_t, sec *sector_t) (r *sector_t) {
	if int32(line.Fflags)&ml_TWOSIDED == 0 {
		return nil
	}
	if line.Ffrontsector == sec {
		return line.Fbacksector
	}
	return line.Ffrontsector
}

// C documentation
//
//	//
//	// p_FindLowestFloorSurrounding()
//	// FIND LOWEST FLOOR HEIGHT IN SURROUNDING SECTORS
//	//
func p_FindLowestFloorSurrounding(sec *sector_t) fixed_t {
	var check *line_t
	var other *sector_t
	var floor fixed_t
	var i int32
	floor = sec.Ffloorheight
	i = 0
	for {
		if i >= sec.Flinecount {
			break
		}
		check = sec.Flines[i]
		other = getNextSector(check, sec)
		if other == nil {
			goto _1
		}
		if other.Ffloorheight < floor {
			floor = other.Ffloorheight
		}
		goto _1
	_1:
		;
		i++
	}
	return floor
}

// C documentation
//
//	//
//	// p_FindHighestFloorSurrounding()
//	// FIND HIGHEST FLOOR HEIGHT IN SURROUNDING SECTORS
//	//
func p_FindHighestFloorSurrounding(sec *sector_t) fixed_t {
	var check *line_t
	var other *sector_t
	var floor fixed_t
	var i int32
	floor = -500 * (1 << FRACBITS)
	i = 0
	for {
		if i >= sec.Flinecount {
			break
		}
		check = sec.Flines[i]
		other = getNextSector(check, sec)
		if other == nil {
			goto _1
		}
		if other.Ffloorheight > floor {
			floor = other.Ffloorheight
		}
		goto _1
	_1:
		;
		i++
	}
	return floor
}

//
// P_FindNextHighestFloor
// FIND NEXT HIGHEST FLOOR IN SURROUNDING SECTORS
// Note: this should be doable w/o a fixed array.

// Thanks to entryway for the Vanilla overflow emulation.

// 20 adjoining sectors max!

func p_FindNextHighestFloor(sec *sector_t, currentheight int32) fixed_t {
	var check *line_t
	var other *sector_t
	var h, i, min, v2 int32
	var height fixed_t
	var heightlist [22]fixed_t
	height = currentheight
	i = 0
	h = 0
	for {
		if i >= sec.Flinecount {
			break
		}
		check = sec.Flines[i]
		other = getNextSector(check, sec)
		if other == nil {
			goto _1
		}
		if other.Ffloorheight > height {
			// Emulation of memory (stack) overflow
			if h == MAX_ADJOINING_SECTORS+1 {
				height = other.Ffloorheight
			} else {
				if h == MAX_ADJOINING_SECTORS+2 {
					// Fatal overflow: game crashes at 22 textures
					i_Error("Sector with more than 22 adjoining sectors. Vanilla will crash here")
				}
			}
			v2 = h
			h++
			heightlist[v2] = other.Ffloorheight
		}
		goto _1
	_1:
		;
		i++
	}
	// Find lowest height in list
	if h == 0 {
		return currentheight
	}
	min = heightlist[0]
	// Range checking?
	i = 1
	for {
		if i >= h {
			break
		}
		if heightlist[i] < min {
			min = heightlist[i]
		}
		goto _3
	_3:
		;
		i++
	}
	return min
}

// C documentation
//
//	//
//	// FIND LOWEST CEILING IN THE SURROUNDING SECTORS
//	//
func p_FindLowestCeilingSurrounding(sec *sector_t) fixed_t {
	var check *line_t
	var other *sector_t
	var height fixed_t
	var i int32
	height = int32(INT_MAX13)
	i = 0
	for {
		if i >= sec.Flinecount {
			break
		}
		check = sec.Flines[i]
		other = getNextSector(check, sec)
		if other == nil {
			goto _1
		}
		if other.Fceilingheight < height {
			height = other.Fceilingheight
		}
		goto _1
	_1:
		;
		i++
	}
	return height
}

// C documentation
//
//	//
//	// FIND HIGHEST CEILING IN THE SURROUNDING SECTORS
//	//
func p_FindHighestCeilingSurrounding(sec *sector_t) fixed_t {
	var check *line_t
	var other *sector_t
	var height fixed_t
	var i int32
	height = 0
	i = 0
	for {
		if i >= sec.Flinecount {
			break
		}
		check = sec.Flines[i]
		other = getNextSector(check, sec)
		if other == nil {
			goto _1
		}
		if other.Fceilingheight > height {
			height = other.Fceilingheight
		}
		goto _1
	_1:
		;
		i++
	}
	return height
}

// C documentation
//
//	//
//	// RETURN NEXT SECTOR # THAT LINE TAG REFERS TO
//	//
func p_FindSectorFromLineTag(line *line_t, start int32) int32 {
	var i int32
	i = start + 1
	for {
		if i >= numsectors {
			break
		}
		if sectors[i].Ftag == line.Ftag {
			return i
		}
		goto _1
	_1:
		;
		i++
	}
	return -1
}

// C documentation
//
//	//
//	// Find minimum light from an adjacent sector
//	//
func p_FindMinSurroundingLight(sector *sector_t, max int32) int32 {
	var line *line_t
	var check *sector_t
	var i, min int32
	min = max
	i = 0
	for {
		if i >= sector.Flinecount {
			break
		}
		line = sector.Flines[i]
		check = getNextSector(line, sector)
		if check == nil {
			goto _1
		}
		if int32(check.Flightlevel) < min {
			min = int32(check.Flightlevel)
		}
		goto _1
	_1:
		;
		i++
	}
	return min
}

//
// EVENTS
// Events are operations triggered by using, crossing,
// or shooting special lines, or by timed thinkers.
//

// C documentation
//
//	//
//	// p_CrossSpecialLine - TRIGGER
//	// Called every time a thing origin is about
//	//  to cross a line with a non 0 special.
//	//
func p_CrossSpecialLine(linenum int32, side int32, thing *mobj_t) {
	var ok int32
	line := &lines[linenum]
	//	Triggers that other things can activate
	if thing.Fplayer == nil {
		// Things that should NOT trigger specials...
		switch thing.Ftype1 {
		case mt_ROCKET:
			fallthrough
		case mt_PLASMA:
			fallthrough
		case mt_BFG:
			fallthrough
		case mt_TROOPSHOT:
			fallthrough
		case mt_HEADSHOT:
			fallthrough
		case mt_BRUISERSHOT:
			return
		default:
			break
		}
		ok = 0
		switch int32(line.Fspecial) {
		case 39: // TELEPORT TRIGGER
			fallthrough
		case 97: // TELEPORT RETRIGGER
			fallthrough
		case 125: // TELEPORT MONSTERONLY TRIGGER
			fallthrough
		case 126: // TELEPORT MONSTERONLY RETRIGGER
			fallthrough
		case 4: // RAISE DOOR
			fallthrough
		case 10: // PLAT DOWN-WAIT-UP-STAY TRIGGER
			fallthrough
		case 88: // PLAT DOWN-WAIT-UP-STAY RETRIGGER
			ok = 1
			break
		}
		if ok == 0 {
			return
		}
	}
	// Note: could use some const's here.
	switch int32(line.Fspecial) {
	// TRIGGERS.
	// All from here to RETRIGGERS.
	case 2:
		// Open Door
		ev_DoDoor(line, int32(vld_open))
		line.Fspecial = 0
	case 3:
		// Close Door
		ev_DoDoor(line, int32(vld_close))
		line.Fspecial = 0
	case 4:
		// Raise Door
		ev_DoDoor(line, int32(vld_normal))
		line.Fspecial = 0
	case 5:
		// Raise Floor
		ev_DoFloor(line, int32(raiseFloor))
		line.Fspecial = 0
	case 6:
		// Fast Ceiling Crush & Raise
		ev_DoCeiling(line, int32(fastCrushAndRaise))
		line.Fspecial = 0
	case 8:
		// Build Stairs
		ev_BuildStairs(line, int32(build8))
		line.Fspecial = 0
	case 10:
		// PlatDownWaitUp
		ev_DoPlat(line, int32(downWaitUpStay), 0)
		line.Fspecial = 0
	case 12:
		// Light Turn On - brightest near
		ev_LightTurnOn(line, 0)
		line.Fspecial = 0
	case 13:
		// Light Turn On 255
		ev_LightTurnOn(line, 255)
		line.Fspecial = 0
	case 16:
		// Close Door 30
		ev_DoDoor(line, int32(vld_close30ThenOpen))
		line.Fspecial = 0
	case 17:
		// Start Light Strobing
		ev_StartLightStrobing(line)
		line.Fspecial = 0
	case 19:
		// Lower Floor
		ev_DoFloor(line, int32(lowerFloor))
		line.Fspecial = 0
	case 22:
		// Raise floor to nearest height and change texture
		ev_DoPlat(line, int32(raiseToNearestAndChange), 0)
		line.Fspecial = 0
	case 25:
		// Ceiling Crush and Raise
		ev_DoCeiling(line, int32(crushAndRaise))
		line.Fspecial = 0
	case 30:
		// Raise floor to shortest texture height
		//  on either side of lines.
		ev_DoFloor(line, int32(raiseToTexture))
		line.Fspecial = 0
	case 35:
		// Lights Very Dark
		ev_LightTurnOn(line, 35)
		line.Fspecial = 0
	case 36:
		// Lower Floor (TURBO)
		ev_DoFloor(line, int32(turboLower))
		line.Fspecial = 0
	case 37:
		// LowerAndChange
		ev_DoFloor(line, int32(lowerAndChange))
		line.Fspecial = 0
	case 38:
		// Lower Floor To Lowest
		ev_DoFloor(line, int32(lowerFloorToLowest))
		line.Fspecial = 0
	case 39:
		// TELEPORT!
		ev_Teleport(line, side, thing)
		line.Fspecial = 0
	case 40:
		// RaiseCeilingLowerFloor
		ev_DoCeiling(line, int32(raiseToHighest))
		ev_DoFloor(line, int32(lowerFloorToLowest))
		line.Fspecial = 0
	case 44:
		// Ceiling Crush
		ev_DoCeiling(line, int32(lowerAndCrush))
		line.Fspecial = 0
	case 52:
		// EXIT!
		g_ExitLevel()
	case 53:
		// Perpetual Platform Raise
		ev_DoPlat(line, int32(perpetualRaise), 0)
		line.Fspecial = 0
	case 54:
		// Platform Stop
		ev_StopPlat(line)
		line.Fspecial = 0
	case 56:
		// Raise Floor Crush
		ev_DoFloor(line, int32(raiseFloorCrush))
		line.Fspecial = 0
	case 57:
		// Ceiling Crush Stop
		ev_CeilingCrushStop(line)
		line.Fspecial = 0
	case 58:
		// Raise Floor 24
		ev_DoFloor(line, int32(raiseFloor24))
		line.Fspecial = 0
	case 59:
		// Raise Floor 24 And Change
		ev_DoFloor(line, int32(raiseFloor24AndChange))
		line.Fspecial = 0
	case 104:
		// Turn lights off in sector(tag)
		ev_TurnTagLightsOff(line)
		line.Fspecial = 0
	case 108:
		// Blazing Door Raise (faster than TURBO!)
		ev_DoDoor(line, int32(vld_blazeRaise))
		line.Fspecial = 0
	case 109:
		// Blazing Door Open (faster than TURBO!)
		ev_DoDoor(line, int32(vld_blazeOpen))
		line.Fspecial = 0
	case 100:
		// Build Stairs Turbo 16
		ev_BuildStairs(line, int32(turbo16))
		line.Fspecial = 0
	case 110:
		// Blazing Door Close (faster than TURBO!)
		ev_DoDoor(line, int32(vld_blazeClose))
		line.Fspecial = 0
	case 119:
		// Raise floor to nearest surr. floor
		ev_DoFloor(line, int32(raiseFloorToNearest))
		line.Fspecial = 0
	case 121:
		// Blazing PlatDownWaitUpStay
		ev_DoPlat(line, int32(blazeDWUS), 0)
		line.Fspecial = 0
	case 124:
		// Secret EXIT
		g_SecretExitLevel()
	case 125:
		// TELEPORT MonsterONLY
		if thing.Fplayer == nil {
			ev_Teleport(line, side, thing)
			line.Fspecial = 0
		}
	case 130:
		// Raise Floor Turbo
		ev_DoFloor(line, int32(raiseFloorTurbo))
		line.Fspecial = 0
	case 141:
		// Silent Ceiling Crush & Raise
		ev_DoCeiling(line, int32(silentCrushAndRaise))
		line.Fspecial = 0
		break
		// RETRIGGERS.  All from here till end.
		fallthrough
	case 72:
		// Ceiling Crush
		ev_DoCeiling(line, int32(lowerAndCrush))
	case 73:
		// Ceiling Crush and Raise
		ev_DoCeiling(line, int32(crushAndRaise))
	case 74:
		// Ceiling Crush Stop
		ev_CeilingCrushStop(line)
	case 75:
		// Close Door
		ev_DoDoor(line, int32(vld_close))
	case 76:
		// Close Door 30
		ev_DoDoor(line, int32(vld_close30ThenOpen))
	case 77:
		// Fast Ceiling Crush & Raise
		ev_DoCeiling(line, int32(fastCrushAndRaise))
	case 79:
		// Lights Very Dark
		ev_LightTurnOn(line, 35)
	case 80:
		// Light Turn On - brightest near
		ev_LightTurnOn(line, 0)
	case 81:
		// Light Turn On 255
		ev_LightTurnOn(line, 255)
	case 82:
		// Lower Floor To Lowest
		ev_DoFloor(line, int32(lowerFloorToLowest))
	case 83:
		// Lower Floor
		ev_DoFloor(line, int32(lowerFloor))
	case 84:
		// LowerAndChange
		ev_DoFloor(line, int32(lowerAndChange))
	case 86:
		// Open Door
		ev_DoDoor(line, int32(vld_open))
	case 87:
		// Perpetual Platform Raise
		ev_DoPlat(line, int32(perpetualRaise), 0)
	case 88:
		// PlatDownWaitUp
		ev_DoPlat(line, int32(downWaitUpStay), 0)
	case 89:
		// Platform Stop
		ev_StopPlat(line)
	case 90:
		// Raise Door
		ev_DoDoor(line, int32(vld_normal))
	case 91:
		// Raise Floor
		ev_DoFloor(line, int32(raiseFloor))
	case 92:
		// Raise Floor 24
		ev_DoFloor(line, int32(raiseFloor24))
	case 93:
		// Raise Floor 24 And Change
		ev_DoFloor(line, int32(raiseFloor24AndChange))
	case 94:
		// Raise Floor Crush
		ev_DoFloor(line, int32(raiseFloorCrush))
	case 95:
		// Raise floor to nearest height
		// and change texture.
		ev_DoPlat(line, int32(raiseToNearestAndChange), 0)
	case 96:
		// Raise floor to shortest texture height
		// on either side of lines.
		ev_DoFloor(line, int32(raiseToTexture))
	case 97:
		// TELEPORT!
		ev_Teleport(line, side, thing)
	case 98:
		// Lower Floor (TURBO)
		ev_DoFloor(line, int32(turboLower))
	case 105:
		// Blazing Door Raise (faster than TURBO!)
		ev_DoDoor(line, int32(vld_blazeRaise))
	case 106:
		// Blazing Door Open (faster than TURBO!)
		ev_DoDoor(line, int32(vld_blazeOpen))
	case 107:
		// Blazing Door Close (faster than TURBO!)
		ev_DoDoor(line, int32(vld_blazeClose))
	case 120:
		// Blazing PlatDownWaitUpStay.
		ev_DoPlat(line, int32(blazeDWUS), 0)
	case 126:
		// TELEPORT MonsterONLY.
		if thing.Fplayer == nil {
			ev_Teleport(line, side, thing)
		}
	case 128:
		// Raise To Nearest Floor
		ev_DoFloor(line, int32(raiseFloorToNearest))
	case 129:
		// Raise Floor Turbo
		ev_DoFloor(line, int32(raiseFloorTurbo))
		break
	}
}

// C documentation
//
//	//
//	// p_ShootSpecialLine - IMPACT SPECIALS
//	// Called when a thing shoots a special line.
//	//
func p_ShootSpecialLine(thing *mobj_t, line *line_t) {
	var ok int32
	//	Impacts that other things can activate.
	if thing.Fplayer == nil {
		ok = 0
		switch int32(line.Fspecial) {
		case 46:
			// OPEN DOOR IMPACT
			ok = 1
			break
		}
		if ok == 0 {
			return
		}
	}
	switch int32(line.Fspecial) {
	case 24:
		// RAISE FLOOR
		ev_DoFloor(line, int32(raiseFloor))
		p_ChangeSwitchTexture(line, 0)
	case 46:
		// OPEN DOOR
		ev_DoDoor(line, int32(vld_open))
		p_ChangeSwitchTexture(line, 1)
	case 47:
		// RAISE FLOOR NEAR AND CHANGE
		ev_DoPlat(line, int32(raiseToNearestAndChange), 0)
		p_ChangeSwitchTexture(line, 0)
		break
	}
}

// C documentation
//
//	//
//	// P_PlayerInSpecialSector
//	// Called every tic frame
//	//  that the player origin is in a special sector
//	//
func p_PlayerInSpecialSector(player *player_t) {
	var sector *sector_t
	sector = player.Fmo.Fsubsector.Fsector
	// Falling, not all the way down yet?
	if player.Fmo.Fz != sector.Ffloorheight {
		return
	}
	// Has hitten ground.
	switch int32(sector.Fspecial) {
	case 5:
		// HELLSLIME DAMAGE
		if player.Fpowers[pw_ironfeet] == 0 {
			if leveltime&0x1f == 0 {
				p_DamageMobj(player.Fmo, nil, nil, 10)
			}
		}
	case 7:
		// NUKAGE DAMAGE
		if player.Fpowers[pw_ironfeet] == 0 {
			if leveltime&0x1f == 0 {
				p_DamageMobj(player.Fmo, nil, nil, 5)
			}
		}
	case 16:
		// SUPER HELLSLIME DAMAGE
		fallthrough
	case 4:
		// STROBE HURT
		if player.Fpowers[pw_ironfeet] == 0 || p_Random() < 5 {
			if leveltime&0x1f == 0 {
				p_DamageMobj(player.Fmo, nil, nil, 20)
			}
		}
	case 9:
		// SECRET SECTOR
		player.Fsecretcount++
		sector.Fspecial = 0
	case 11:
		// EXIT SUPER DAMAGE! (for E1M8 finale)
		player.Fcheats &= ^CF_GODMODE
		if leveltime&0x1f == 0 {
			p_DamageMobj(player.Fmo, nil, nil, 20)
		}
		if player.Fhealth <= 10 {
			g_ExitLevel()
		}
	default:
		i_Error("p_PlayerInSpecialSector: unknown special %d", int32(sector.Fspecial))
		break
	}
}

func p_UpdateSpecials() {
	var i, pic int32
	//	LEVEL TIMER
	if levelTimer == 1 {
		levelTimeCount--
		if levelTimeCount == 0 {
			g_ExitLevel()
		}
	}
	//	ANIMATE FLATS AND TEXTURES GLOBALLY
	for pos := 0; ; pos++ {
		anim := &anims[pos]
		if anim == lastanim {
			break
		}
		i = anim.Fbasepic
		for {
			if i >= anim.Fbasepic+anim.Fnumpics {
				break
			}
			pic = anim.Fbasepic + (leveltime/anim.Fspeed+i)%anim.Fnumpics
			if anim.Fistexture != 0 {
				texturetranslation[i] = pic
			} else {
				flattranslation[i] = pic
			}
			i++
		}

	}
	//	ANIMATE LINE SPECIALS
	i = 0
	for {
		if i >= int32(numlinespecials) {
			break
		}
		line := linespeciallist[i]
		switch int32(line.Fspecial) {
		case 48:
			// EFFECT FIRSTCOL SCROLL +
			sides[line.Fsidenum[0]].Ftextureoffset += 1 << FRACBITS
			break
		}
		goto _3
	_3:
		;
		i++
	}
	//	DO BUTTONS
	i = 0
	for {
		if i >= MAXBUTTONS {
			break
		}
		if buttonlist[i].Fbtimer != 0 {
			buttonlist[i].Fbtimer--
			if buttonlist[i].Fbtimer == 0 {
				switch buttonlist[i].Fwhere {
				case int32(top):
					sides[buttonlist[i].Fline.Fsidenum[0]].Ftoptexture = int16(buttonlist[i].Fbtexture)
				case int32(middle):
					sides[buttonlist[i].Fline.Fsidenum[0]].Fmidtexture = int16(buttonlist[i].Fbtexture)
				case int32(bottom):
					sides[buttonlist[i].Fline.Fsidenum[0]].Fbottomtexture = int16(buttonlist[i].Fbtexture)
					break
				}
				s_StartSound(buttonlist[i].Fsoundorg, int32(sfx_swtchn))
				buttonlist[i] = button_t{}
			}
		}
		goto _4
	_4:
		;
		i++
	}
}

//
// Donut overrun emulation
//
// Derived from the code from PrBoom+.  Thanks go to Andrey Budko (entryway)
// as usual :-)
//

func donutOverrun(s3_floorheight *fixed_t, s3_floorpic *int16, line *line_t, pillar_sector *sector_t) {
	var p int32
	if first != 0 {
		// This is the first time we have had an overrun.
		first = 0
		// Default values
		tmp_s3_floorheight = DONUT_FLOORHEIGHT_DEFAULT
		tmp_s3_floorpic = DONUT_FLOORPIC_DEFAULT
		//!
		// @category compat
		// @arg <x> <y>
		//
		// Use the specified magic values when emulating behavior caused
		// by memory overruns from improperly constructed donuts.
		// In Vanilla Doom this can differ depending on the operating
		// system.  The default (if this option is not specified) is to
		// emulate the behavior when running under Windows 98.
		p = m_CheckParmWithArgs("-donut", 2)
		if p > 0 {
			// Dump of needed memory: (fixed_t)0000:0000 and (short)0000:0008
			//
			// C:\>debug
			// -d 0:0
			//
			// DOS 6.22:
			// 0000:0000    (57 92 19 00) F4 06 70 00-(16 00)
			// DOS 7.1:
			// 0000:0000    (9E 0F C9 00) 65 04 70 00-(16 00)
			// Win98:
			// 0000:0000    (00 00 00 00) 65 04 70 00-(16 00)
			// DOSBox under XP:
			// 0000:0000    (00 00 00 F1) ?? ?? ?? 00-(07 00)
			v, _ := strconv.Atoi(myargs[p+1])
			tmp_s3_floorheight = int32(v)
			v, _ = strconv.Atoi(myargs[p+2])
			tmp_s3_floorpic = int32(v)
			if tmp_s3_floorpic >= numflats {
				fprintf_ccgo(os.Stderr, "donutOverrun: The second parameter for \"-donut\" switch should be greater than 0 and less than number of flats (%d). Using default value (%d) instead. \n", numflats, DONUT_FLOORPIC_DEFAULT)
				tmp_s3_floorpic = DONUT_FLOORPIC_DEFAULT
			}
		}
	}
	/*
	   fprintf(stderr,
	           "Linedef: %d; Sector: %d; "
	           "New floor height: %d; New floor pic: %d\n",
	           line->iLineID, pillar_sector->iSectorID,
	           tmp_s3_floorheight >> 16, tmp_s3_floorpic);
	*/
	*s3_floorheight = tmp_s3_floorheight
	*s3_floorpic = int16(tmp_s3_floorpic)
}

var first = 1

var tmp_s3_floorheight int32

var tmp_s3_floorpic int32

// C documentation
//
//	//
//	// Special Stuff that can not be categorized
//	//
func ev_DoDonut(line *line_t) int32 {
	var s1, s2, s3 *sector_t
	var rtn, secnum, v1 int32
	secnum = -1
	rtn = 0
	for {
		v1 = p_FindSectorFromLineTag(line, secnum)
		secnum = v1
		if !(v1 >= 0) {
			break
		}
		s1 = &sectors[secnum]
		// ALREADY MOVING?  IF SO, KEEP GOING...
		if s1.Fspecialdata != nil {
			continue
		}
		rtn = 1
		s2 = getNextSector(s1.Flines[0], s1)
		// Vanilla Doom does not check if the linedef is one sided.  The
		// game does not crash, but reads invalid memory and causes the
		// sector floor to move "down" to some unknown height.
		// DOSbox prints a warning about an invalid memory access.
		//
		// I'm not sure exactly what invalid memory is being read.  This
		// isn't something that should be done, anyway.
		// Just print a warning and return.
		if s2 == nil {
			fprintf_ccgo(os.Stderr, "ev_DoDonut: linedef had no second sidedef! Unexpected behavior may occur in Vanilla Doom. \n")
			break
		}
		for i := int32(0); i < s2.Flinecount; i++ {
			s3 = s2.Flines[i].Fbacksector
			if s3 == s1 {
				continue
			}
			var floorpic int16
			var floorheight fixed_t
			if s3 == nil {
				// e6y
				// s3 is NULL, so
				// s3->floorheight is an int at 0000:0000
				// s3->floorpic is a short at 0000:0008
				// Trying to emulate
				fprintf_ccgo(os.Stderr, "ev_DoDonut: WARNING: emulating buffer overrun due to NULL back sector. Unexpected behavior may occur in Vanilla Doom.\n")
				donutOverrun(&floorheight, &floorpic, line, s1)
			} else {
				floorheight = s3.Ffloorheight
				floorpic = s3.Ffloorpic
			}
			//	Spawn rising slime
			floorP := &floormove_t{}
			p_AddThinker(&floorP.Fthinker)
			s2.Fspecialdata = floorP
			floorP.Fthinker.Ffunction = floorP
			floorP.Ftype1 = int32(donutRaise)
			floorP.Fcrush = 0
			floorP.Fdirection = 1
			floorP.Fsector = s2
			floorP.Fspeed = 1 << FRACBITS / 2
			floorP.Ftexture = floorpic
			floorP.Fnewspecial = 0
			floorP.Ffloordestheight = floorheight
			//	Spawn lowering donut-hole
			floorP = &floormove_t{}
			p_AddThinker(&floorP.Fthinker)
			s1.Fspecialdata = floorP
			floorP.Fthinker.Ffunction = floorP
			floorP.Ftype1 = int32(lowerFloor)
			floorP.Fcrush = 0
			floorP.Fdirection = -1
			floorP.Fsector = s1
			floorP.Fspeed = 1 << FRACBITS / 2
			floorP.Ffloordestheight = floorheight
			break
		}
	}
	return rtn
}

// C documentation
//
//	// Parses command line parameters.
func p_SpawnSpecials() {
	var i int32
	// See if -TIMER was specified.
	if timelimit > 0 && deathmatch != 0 {
		levelTimer = 1
		levelTimeCount = timelimit * 60 * TICRATE
	} else {
		levelTimer = 0
	}
	//	Init special SECTORs.
	i = 0
	for {
		if i >= numsectors {
			break
		}
		sector := &sectors[i]
		if sector.Fspecial == 0 {
			goto _1
		}
		switch int32(sector.Fspecial) {
		case 1:
			// FLICKERING LIGHTS
			p_SpawnLightFlash(sector)
		case 2:
			// STROBE FAST
			p_SpawnStrobeFlash(sector, FASTDARK, 0)
		case 3:
			// STROBE SLOW
			p_SpawnStrobeFlash(sector, SLOWDARK, 0)
		case 4:
			// STROBE FAST/DEATH SLIME
			p_SpawnStrobeFlash(sector, FASTDARK, 0)
			sector.Fspecial = 4
		case 8:
			// GLOWING LIGHT
			p_SpawnGlowingLight(sector)
		case 9:
			// SECRET SECTOR
			totalsecret++
		case 10:
			// DOOR CLOSE IN 30 SECONDS
			p_SpawnDoorCloseIn30(sector)
		case 12:
			// SYNC STROBE SLOW
			p_SpawnStrobeFlash(sector, SLOWDARK, 1)
		case 13:
			// SYNC STROBE FAST
			p_SpawnStrobeFlash(sector, FASTDARK, 1)
		case 14:
			// DOOR RAISE IN 5 MINUTES
			p_SpawnDoorRaiseIn5Mins(sector, i)
		case 17:
			p_SpawnFireFlicker(sector)
			break
		}
		goto _1
	_1:
		;
		i++
	}
	//	Init line EFFECTs
	numlinespecials = 0
	i = 0
	for {
		if i >= numlines {
			break
		}
		switch int32(lines[i].Fspecial) {
		case 48:
			if int32(numlinespecials) >= MAXLINEANIMS {
				i_Error("Too many scrolling wall linedefs! (Vanilla limit is 64)")
			}
			// EFFECT FIRSTCOL SCROLL+
			linespeciallist[numlinespecials] = &lines[i]
			numlinespecials++
			break
		}
		goto _2
	_2:
		;
		i++
	}
	clear(activeceilings[:])
	clear(activeplats[:])
	clear(buttonlist[:])
	// UNUSED: no horizonal sliders.
	//	P_InitSlidingDoorFrames();
}

func init() {
	alphSwitchList = [41]switchlist_t{
		0: {
			Fname1:   "SW1BRCOM",
			Fname2:   "SW2BRCOM",
			Fepisode: 1,
		},
		1: {
			Fname1:   "SW1BRN1",
			Fname2:   "SW2BRN1",
			Fepisode: 1,
		},
		2: {
			Fname1:   "SW1BRN2",
			Fname2:   "SW2BRN2",
			Fepisode: 1,
		},
		3: {
			Fname1:   "SW1BRNGN",
			Fname2:   "SW2BRNGN",
			Fepisode: 1,
		},
		4: {
			Fname1:   "SW1BROWN",
			Fname2:   "SW2BROWN",
			Fepisode: 1,
		},
		5: {
			Fname1:   "SW1COMM",
			Fname2:   "SW2COMM",
			Fepisode: 1,
		},
		6: {
			Fname1:   "SW1COMP",
			Fname2:   "SW2COMP",
			Fepisode: 1,
		},
		7: {
			Fname1:   "SW1DIRT",
			Fname2:   "SW2DIRT",
			Fepisode: 1,
		},
		8: {
			Fname1:   "SW1EXIT",
			Fname2:   "SW2EXIT",
			Fepisode: 1,
		},
		9: {
			Fname1:   "SW1GRAY",
			Fname2:   "SW2GRAY",
			Fepisode: 1,
		},
		10: {
			Fname1:   "SW1GRAY1",
			Fname2:   "SW2GRAY1",
			Fepisode: 1,
		},
		11: {
			Fname1:   "SW1METAL",
			Fname2:   "SW2METAL",
			Fepisode: 1,
		},
		12: {
			Fname1:   "SW1PIPE",
			Fname2:   "SW2PIPE",
			Fepisode: 1,
		},
		13: {
			Fname1:   "SW1SLAD",
			Fname2:   "SW2SLAD",
			Fepisode: 1,
		},
		14: {
			Fname1:   "SW1STARG",
			Fname2:   "SW2STARG",
			Fepisode: 1,
		},
		15: {
			Fname1:   "SW1STON1",
			Fname2:   "SW2STON1",
			Fepisode: 1,
		},
		16: {
			Fname1:   "SW1STON2",
			Fname2:   "SW2STON2",
			Fepisode: 1,
		},
		17: {
			Fname1:   "SW1STONE",
			Fname2:   "SW2STONE",
			Fepisode: 1,
		},
		18: {
			Fname1:   "SW1STRTN",
			Fname2:   "SW2STRTN",
			Fepisode: 1,
		},
		19: {
			Fname1:   "SW1BLUE",
			Fname2:   "SW2BLUE",
			Fepisode: 2,
		},
		20: {
			Fname1:   "SW1CMT",
			Fname2:   "SW2CMT",
			Fepisode: 2,
		},
		21: {
			Fname1:   "SW1GARG",
			Fname2:   "SW2GARG",
			Fepisode: 2,
		},
		22: {
			Fname1:   "SW1GSTON",
			Fname2:   "SW2GSTON",
			Fepisode: 2,
		},
		23: {
			Fname1:   "SW1HOT",
			Fname2:   "SW2HOT",
			Fepisode: 2,
		},
		24: {
			Fname1:   "SW1LION",
			Fname2:   "SW2LION",
			Fepisode: 2,
		},
		25: {
			Fname1:   "SW1SATYR",
			Fname2:   "SW2SATYR",
			Fepisode: 2,
		},
		26: {
			Fname1:   "SW1SKIN",
			Fname2:   "SW2SKIN",
			Fepisode: 2,
		},
		27: {
			Fname1:   "SW1VINE",
			Fname2:   "SW2VINE",
			Fepisode: 2,
		},
		28: {
			Fname1:   "SW1WOOD",
			Fname2:   "SW2WOOD",
			Fepisode: 2,
		},
		29: {
			Fname1:   "SW1PANEL",
			Fname2:   "SW2PANEL",
			Fepisode: 3,
		},
		30: {
			Fname1:   "SW1ROCK",
			Fname2:   "SW2ROCK",
			Fepisode: 3,
		},
		31: {
			Fname1:   "SW1MET2",
			Fname2:   "SW2MET2",
			Fepisode: 3,
		},
		32: {
			Fname1:   "SW1WDMET",
			Fname2:   "SW2WDMET",
			Fepisode: 3,
		},
		33: {
			Fname1:   "SW1BRIK",
			Fname2:   "SW2BRIK",
			Fepisode: 3,
		},
		34: {
			Fname1:   "SW1MOD1",
			Fname2:   "SW2MOD1",
			Fepisode: 3,
		},
		35: {
			Fname1:   "SW1ZIM",
			Fname2:   "SW2ZIM",
			Fepisode: 3,
		},
		36: {
			Fname1:   "SW1STON6",
			Fname2:   "SW2STON6",
			Fepisode: 3,
		},
		37: {
			Fname1:   "SW1TEK",
			Fname2:   "SW2TEK",
			Fepisode: 3,
		},
		38: {
			Fname1:   "SW1MARB",
			Fname2:   "SW2MARB",
			Fepisode: 3,
		},
		39: {
			Fname1:   "SW1SKULL",
			Fname2:   "SW2SKULL",
			Fepisode: 3,
		},
		40: {
			Fname1: "",
			Fname2: "",
		},
	}
}

// C documentation
//
//	//
//	// P_InitSwitchList
//	// Only called at game initialization.
//	//
func p_InitSwitchList() {
	var episode, i, index, v2, v3 int32
	episode = 1
	if gamemode == registered || gamemode == retail {
		episode = 2
	} else {
		if gamemode == commercial {
			episode = 3
		}
	}
	index = 0
	i = 0
	for {
		if i >= MAXSWITCHES {
			break
		}
		if alphSwitchList[i].Fepisode == 0 {
			numswitches = index / 2
			switchlist[index] = -1
			break
		}
		if int32(alphSwitchList[i].Fepisode) <= episode {
			v2 = index
			index++
			switchlist[v2] = r_TextureNumForName(alphSwitchList[i].Fname2)
			v3 = index
			index++
			switchlist[v3] = r_TextureNumForName(alphSwitchList[i].Fname1)
		}
		goto _1
	_1:
		;
		i++
	}
}

// C documentation
//
//	//
//	// Start a button counting down till it turns off.
//	//
func p_StartButton(line *line_t, w bwhere_e, texture int32, time int32) {
	var i int32
	// See if button is already pressed
	i = 0
	for {
		if i >= MAXBUTTONS {
			break
		}
		if buttonlist[i].Fbtimer != 0 && buttonlist[i].Fline == line {
			return
		}
		goto _1
	_1:
		;
		i++
	}
	i = 0
	for {
		if i >= MAXBUTTONS {
			break
		}
		if buttonlist[i].Fbtimer == 0 {
			buttonlist[i].Fline = line
			buttonlist[i].Fwhere = w
			buttonlist[i].Fbtexture = texture
			buttonlist[i].Fbtimer = time
			buttonlist[i].Fsoundorg = &line.Ffrontsector.Fsoundorg
			return
		}
		goto _2
	_2:
		;
		i++
	}
	i_Error("p_StartButton: no button slots left!")
}

// C documentation
//
//	//
//	// Function that changes wall texture.
//	// Tell it if switch is ok to use again (1=yes, it's a button).
//	//
func p_ChangeSwitchTexture(line *line_t, useAgain int32) {
	var i, sound, texBot, texMid, texTop int32
	if useAgain == 0 {
		line.Fspecial = 0
	}
	texTop = int32(sides[line.Fsidenum[0]].Ftoptexture)
	texMid = int32(sides[line.Fsidenum[0]].Fmidtexture)
	texBot = int32(sides[line.Fsidenum[0]].Fbottomtexture)
	sound = int32(sfx_swtchn)
	// EXIT SWITCH?
	if int32(line.Fspecial) == 11 {
		sound = int32(sfx_swtchx)
	}
	i = 0
	for {
		if i >= numswitches*2 {
			break
		}
		if switchlist[i] == texTop {
			s_StartSound(buttonlist[0].Fsoundorg, sound)
			sides[line.Fsidenum[0]].Ftoptexture = int16(switchlist[i^1])
			if useAgain != 0 {
				p_StartButton(line, int32(top), switchlist[i], BUTTONTIME)
			}
			return
		} else {
			if switchlist[i] == texMid {
				s_StartSound(buttonlist[0].Fsoundorg, sound)
				sides[line.Fsidenum[0]].Fmidtexture = int16(switchlist[i^1])
				if useAgain != 0 {
					p_StartButton(line, int32(middle), switchlist[i], BUTTONTIME)
				}
				return
			} else {
				if switchlist[i] == texBot {
					s_StartSound(buttonlist[0].Fsoundorg, sound)
					sides[line.Fsidenum[0]].Fbottomtexture = int16(switchlist[i^1])
					if useAgain != 0 {
						p_StartButton(line, int32(bottom), switchlist[i], BUTTONTIME)
					}
					return
				}
			}
		}
		goto _1
	_1:
		;
		i++
	}
}

// C documentation
//
//	//
//	// P_UseSpecialLine
//	// Called when a thing uses a special line.
//	// Only the front sides of lines are usable.
//	//
func p_UseSpecialLine(thing *mobj_t, line *line_t, side int32) boolean {
	// Err...
	// Use the back sides of VERY SPECIAL lines...
	if side != 0 {
		switch int32(line.Fspecial) {
		case 124:
			// Sliding door open&close
			// UNUSED?
		default:
			return 0
			break
		}
	}
	// Switches that other things can activate.
	if thing.Fplayer == nil {
		// never open secret doors
		if int32(line.Fflags)&ml_SECRET != 0 {
			return 0
		}
		switch int32(line.Fspecial) {
		case 1: // MANUAL DOOR RAISE
			fallthrough
		case 32: // MANUAL BLUE
			fallthrough
		case 33: // MANUAL RED
			fallthrough
		case 34: // MANUAL YELLOW
		default:
			return 0
			break
		}
	}
	// do something
	switch int32(line.Fspecial) {
	// MANUALS
	case 1: // Vertical Door
		fallthrough
	case 26: // Blue Door/Locked
		fallthrough
	case 27: // Yellow Door /Locked
		fallthrough
	case 28: // Red Door /Locked
		fallthrough
	case 31: // Manual door open
		fallthrough
	case 32: // Blue locked door open
		fallthrough
	case 33: // Red locked door open
		fallthrough
	case 34: // Yellow locked door open
		fallthrough
	case 117: // Blazing door raise
		fallthrough
	case 118: // Blazing door open
		ev_VerticalDoor(line, thing)
		break
		//UNUSED - Door Slide Open&Close
		// case 124:
		// EV_SlidingDoor (line, thing);
		// break;
		// SWITCHES
		fallthrough
	case 7:
		// Build Stairs
		if ev_BuildStairs(line, int32(build8)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 9:
		// Change Donut
		if ev_DoDonut(line) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 11:
		// Exit level
		p_ChangeSwitchTexture(line, 0)
		g_ExitLevel()
	case 14:
		// Raise Floor 32 and change texture
		if ev_DoPlat(line, int32(raiseAndChange), 32) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 15:
		// Raise Floor 24 and change texture
		if ev_DoPlat(line, int32(raiseAndChange), 24) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 18:
		// Raise Floor to next highest floor
		if ev_DoFloor(line, int32(raiseFloorToNearest)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 20:
		// Raise Plat next highest floor and change texture
		if ev_DoPlat(line, int32(raiseToNearestAndChange), 0) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 21:
		// PlatDownWaitUpStay
		if ev_DoPlat(line, int32(downWaitUpStay), 0) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 23:
		// Lower Floor to Lowest
		if ev_DoFloor(line, int32(lowerFloorToLowest)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 29:
		// Raise Door
		if ev_DoDoor(line, int32(vld_normal)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 41:
		// Lower Ceiling to Floor
		if ev_DoCeiling(line, int32(lowerToFloor)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 71:
		// Turbo Lower Floor
		if ev_DoFloor(line, int32(turboLower)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 49:
		// Ceiling Crush And Raise
		if ev_DoCeiling(line, int32(crushAndRaise)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 50:
		// Close Door
		if ev_DoDoor(line, int32(vld_close)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 51:
		// Secret EXIT
		p_ChangeSwitchTexture(line, 0)
		g_SecretExitLevel()
	case 55:
		// Raise Floor Crush
		if ev_DoFloor(line, int32(raiseFloorCrush)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 101:
		// Raise Floor
		if ev_DoFloor(line, int32(raiseFloor)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 102:
		// Lower Floor to Surrounding floor height
		if ev_DoFloor(line, int32(lowerFloor)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 103:
		// Open Door
		if ev_DoDoor(line, int32(vld_open)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 111:
		// Blazing Door Raise (faster than TURBO!)
		if ev_DoDoor(line, int32(vld_blazeRaise)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 112:
		// Blazing Door Open (faster than TURBO!)
		if ev_DoDoor(line, int32(vld_blazeOpen)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 113:
		// Blazing Door Close (faster than TURBO!)
		if ev_DoDoor(line, int32(vld_blazeClose)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 122:
		// Blazing PlatDownWaitUpStay
		if ev_DoPlat(line, int32(blazeDWUS), 0) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 127:
		// Build Stairs Turbo 16
		if ev_BuildStairs(line, int32(turbo16)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 131:
		// Raise Floor Turbo
		if ev_DoFloor(line, int32(raiseFloorTurbo)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 133:
		// BlzOpenDoor BLUE
		fallthrough
	case 135:
		// BlzOpenDoor RED
		fallthrough
	case 137:
		// BlzOpenDoor YELLOW
		if ev_DoLockedDoor(line, int32(vld_blazeOpen), thing) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
	case 140:
		// Raise Floor 512
		if ev_DoFloor(line, int32(raiseFloor512)) != 0 {
			p_ChangeSwitchTexture(line, 0)
		}
		break
		// BUTTONS
		fallthrough
	case 42:
		// Close Door
		if ev_DoDoor(line, int32(vld_close)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 43:
		// Lower Ceiling to Floor
		if ev_DoCeiling(line, int32(lowerToFloor)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 45:
		// Lower Floor to Surrounding floor height
		if ev_DoFloor(line, int32(lowerFloor)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 60:
		// Lower Floor to Lowest
		if ev_DoFloor(line, int32(lowerFloorToLowest)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 61:
		// Open Door
		if ev_DoDoor(line, int32(vld_open)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 62:
		// PlatDownWaitUpStay
		if ev_DoPlat(line, int32(downWaitUpStay), 1) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 63:
		// Raise Door
		if ev_DoDoor(line, int32(vld_normal)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 64:
		// Raise Floor to ceiling
		if ev_DoFloor(line, int32(raiseFloor)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 66:
		// Raise Floor 24 and change texture
		if ev_DoPlat(line, int32(raiseAndChange), 24) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 67:
		// Raise Floor 32 and change texture
		if ev_DoPlat(line, int32(raiseAndChange), 32) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 65:
		// Raise Floor Crush
		if ev_DoFloor(line, int32(raiseFloorCrush)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 68:
		// Raise Plat to next highest floor and change texture
		if ev_DoPlat(line, int32(raiseToNearestAndChange), 0) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 69:
		// Raise Floor to next highest floor
		if ev_DoFloor(line, int32(raiseFloorToNearest)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 70:
		// Turbo Lower Floor
		if ev_DoFloor(line, int32(turboLower)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 114:
		// Blazing Door Raise (faster than TURBO!)
		if ev_DoDoor(line, int32(vld_blazeRaise)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 115:
		// Blazing Door Open (faster than TURBO!)
		if ev_DoDoor(line, int32(vld_blazeOpen)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 116:
		// Blazing Door Close (faster than TURBO!)
		if ev_DoDoor(line, int32(vld_blazeClose)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 123:
		// Blazing PlatDownWaitUpStay
		if ev_DoPlat(line, int32(blazeDWUS), 0) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 132:
		// Raise Floor Turbo
		if ev_DoFloor(line, int32(raiseFloorTurbo)) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 99:
		// BlzOpenDoor BLUE
		fallthrough
	case 134:
		// BlzOpenDoor RED
		fallthrough
	case 136:
		// BlzOpenDoor YELLOW
		if ev_DoLockedDoor(line, int32(vld_blazeOpen), thing) != 0 {
			p_ChangeSwitchTexture(line, 1)
		}
	case 138:
		// Light Turn On
		ev_LightTurnOn(line, 255)
		p_ChangeSwitchTexture(line, 1)
	case 139:
		// Light Turn Off
		ev_LightTurnOn(line, 35)
		p_ChangeSwitchTexture(line, 1)
		break
	}
	return 1
}

// Data.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

// State.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// C documentation
//
//	//
//	// TELEPORTATION
//	//
func ev_Teleport(line *line_t, side int32, thing *mobj_t) int32 {
	var an uint32
	var fog *mobj_t
	var thinker *thinker_t
	var sector *sector_t
	var i int32
	var tag int16
	var oldx, oldy, oldz, v3, v4 fixed_t
	// don't teleport missiles
	if thing.Fflags&mf_MISSILE != 0 {
		return 0
	}
	// Don't teleport if hit back of line,
	//  so you can get out of teleporter.
	if side == 1 {
		return 0
	}
	tag = line.Ftag
	i = 0
	for {
		if i >= numsectors {
			break
		}
		if sectors[i].Ftag == tag {
			thinker = thinkercap.Fnext
			thinker = thinkercap.Fnext
			for {
				if thinker == &thinkercap {
					break
				}
				// not a mobj
				m, ok := thinker.Ffunction.(*mobj_t)
				if !ok {
					goto _2
				}
				// not a teleportman
				if m.Ftype1 != mt_TELEPORTMAN {
					goto _2
				}
				sector = m.Fsubsector.Fsector
				// wrong sector
				if sectorIndex(sector) != i {
					goto _2
				}
				oldx = thing.Fx
				oldy = thing.Fy
				oldz = thing.Fz
				if p_TeleportMove(thing, m.Fx, m.Fy) == 0 {
					return 0
				}
				// The first Final Doom executable does not set thing->z
				// when teleporting. This quirk is unique to this
				// particular version; the later version included in
				// some versions of the Id Anthology fixed this.
				if gameversion != exe_final {
					thing.Fz = thing.Ffloorz
				}
				if thing.Fplayer != nil {
					thing.Fplayer.Fviewz = thing.Fz + thing.Fplayer.Fviewheight
				}
				// spawn teleport fog at source and destination
				fog = p_SpawnMobj(oldx, oldy, oldz, mt_TFOG)
				s_StartSound(&fog.degenmobj_t, int32(sfx_telept))
				an = m.Fangle >> ANGLETOFINESHIFT
				fog = p_SpawnMobj(m.Fx+int32(20)*finecosine[an], m.Fy+int32(20)*finesine[an], thing.Fz, mt_TFOG)
				// emit sound, where?
				s_StartSound(&fog.degenmobj_t, int32(sfx_telept))
				// don't move for a bit
				if thing.Fplayer != nil {
					thing.Freactiontime = 18
				}
				thing.Fangle = m.Fangle
				v4 = 0
				thing.Fmomz = v4
				v3 = v4
				thing.Fmomy = v3
				thing.Fmomx = v3
				return 1
				goto _2
			_2:
				;
				thinker = thinker.Fnext
			}
		}
		goto _1
	_1:
		;
		i++
	}
	return 0
}

// C documentation
//
//	//
//	// P_InitThinkers
//	//
func p_InitThinkers() {
	thinkercap.Fnext = &thinkercap
	thinkercap.Fprev = &thinkercap
}

// C documentation
//
//	//
//	// P_AddThinker
//	// Adds a new thinker at the end of the list.
//	//
func p_AddThinker(thinker *thinker_t) {
	thinkercap.Fprev.Fnext = thinker
	thinker.Fnext = &thinkercap
	thinker.Fprev = thinkercap.Fprev
	thinkercap.Fprev = thinker
}

// C documentation
//
//	//
//	// P_RemoveThinker
//	// Deallocation is lazy -- it will not actually be freed
//	// until its thinking turn comes up.
//	//
func p_RemoveThinker(thinker *thinker_t) {
	// FIXME: NOP.
	thinker.Ffunction = nil
}

// C documentation
//
//	//
//	// P_RunThinkers
//	//
func p_RunThinkers() {
	var currentthinker *thinker_t
	currentthinker = thinkercap.Fnext
	for currentthinker != &thinkercap {
		if currentthinker.Ffunction == nil {
			// time to remove it
			currentthinker.Fnext.Fprev = currentthinker.Fprev
			currentthinker.Fprev.Fnext = currentthinker.Fnext
		} else {
			currentthinker.Ffunction.ThinkerFunc()
		}
		currentthinker = currentthinker.Fnext
	}
}

//
// P_Ticker
//

func p_Ticker() {
	var i int32
	// run the tic
	if paused != 0 {
		return
	}
	// pause if in menu and at least one tic has been run
	if netgame == 0 && menuactive != 0 && demoplayback == 0 && players[consoleplayer].Fviewz != 1 {
		return
	}
	i = 0
	for {
		if i >= MAXPLAYERS {
			break
		}
		if playeringame[i] != 0 {
			p_PlayerThink(&players[i])
		}
		goto _1
	_1:
		;
		i++
	}
	p_RunThinkers()
	p_UpdateSpecials()
	p_RespawnSpecials()
	// for par times
	leveltime++
}

const ANG1809 = 2147483648
const ANG907 = 1073741824
const INVERSECOLORMAP = 32
const MAXBOB = 1048576

// C documentation
//
//	//
//	// P_Thrust
//	// Moves the given origin along a given angle.
//	//
func p_Thrust(player *player_t, angle angle_t, move fixed_t) {
	angle >>= ANGLETOFINESHIFT
	player.Fmo.Fmomx += fixedMul(move, finecosine[angle])
	player.Fmo.Fmomy += fixedMul(move, finesine[angle])
}

// C documentation
//
//	//
//	// P_CalcHeight
//	// Calculate the walking / running height adjustment
//	//
func p_CalcHeight(player *player_t) {
	var angle int32
	var bob fixed_t
	// Regular movement bobbing
	// (needs to be calculated for gun swing
	// even if not on ground)
	// OPTIMIZE: tablify angle
	// Note: a LUT allows for effects
	//  like a ramp with low health.
	player.Fbob = fixedMul(player.Fmo.Fmomx, player.Fmo.Fmomx) + fixedMul(player.Fmo.Fmomy, player.Fmo.Fmomy)
	player.Fbob >>= 2
	if player.Fbob > MAXBOB {
		player.Fbob = MAXBOB
	}
	if player.Fcheats&CF_NOMOMENTUM != 0 || onground == 0 {
		player.Fviewz = player.Fmo.Fz + 41*(1<<FRACBITS)
		if player.Fviewz > player.Fmo.Fceilingz-4*(1<<FRACBITS) {
			player.Fviewz = player.Fmo.Fceilingz - 4*(1<<FRACBITS)
		}
		player.Fviewz = player.Fmo.Fz + player.Fviewheight
		return
	}
	angle = FINEANGLES / 20 * leveltime & (FINEANGLES - 1)
	bob = fixedMul(player.Fbob/2, finesine[angle])
	// move viewheight
	if player.Fplayerstate == Pst_LIVE {
		player.Fviewheight += player.Fdeltaviewheight
		if player.Fviewheight > 41*(1<<FRACBITS) {
			player.Fviewheight = 41 * (1 << FRACBITS)
			player.Fdeltaviewheight = 0
		}
		if player.Fviewheight < 41*(1<<FRACBITS)/2 {
			player.Fviewheight = 41 * (1 << FRACBITS) / 2
			if player.Fdeltaviewheight <= 0 {
				player.Fdeltaviewheight = 1
			}
		}
		if player.Fdeltaviewheight != 0 {
			player.Fdeltaviewheight += 1 << FRACBITS / 4
			if player.Fdeltaviewheight == 0 {
				player.Fdeltaviewheight = 1
			}
		}
	}
	player.Fviewz = player.Fmo.Fz + player.Fviewheight + bob
	if player.Fviewz > player.Fmo.Fceilingz-4*(1<<FRACBITS) {
		player.Fviewz = player.Fmo.Fceilingz - 4*(1<<FRACBITS)
	}
}

// C documentation
//
//	//
//	// P_MovePlayer
//	//
func p_MovePlayer(player *player_t) {
	cmd := &player.Fcmd
	player.Fmo.Fangle += uint32(int32(cmd.Fangleturn) << 16)
	// Do not let the player control movement
	//  if not onground.
	onground = booluint32(player.Fmo.Fz <= player.Fmo.Ffloorz)
	if cmd.Fforwardmove != 0 && onground != 0 {
		p_Thrust(player, player.Fmo.Fangle, int32(cmd.Fforwardmove)*int32(2048))
	}
	if cmd.Fsidemove != 0 && onground != 0 {
		p_Thrust(player, player.Fmo.Fangle-uint32(ANG907), int32(cmd.Fsidemove)*int32(2048))
	}
	if (cmd.Fforwardmove != 0 || cmd.Fsidemove != 0) && player.Fmo.Fstate == &states[s_PLAY] {
		p_SetMobjState(player.Fmo, s_PLAY_RUN1)
	}
}

//
// P_DeathThink
// Fall on your face when dying.
// Decrease POV height to floor height.
//

func p_DeathThink(player *player_t) {
	var angle, delta angle_t
	p_MovePsprites(player)
	// fall to the ground
	if player.Fviewheight > 6*(1<<FRACBITS) {
		player.Fviewheight -= 1 << FRACBITS
	}
	if player.Fviewheight < 6*(1<<FRACBITS) {
		player.Fviewheight = 6 * (1 << FRACBITS)
	}
	player.Fdeltaviewheight = 0
	onground = booluint32(player.Fmo.Fz <= player.Fmo.Ffloorz)
	p_CalcHeight(player)
	if player.Fattacker != nil && player.Fattacker != player.Fmo {
		angle = r_PointToAngle2(player.Fmo.Fx, player.Fmo.Fy, player.Fattacker.Fx, player.Fattacker.Fy)
		delta = angle - player.Fmo.Fangle
		if delta < uint32(ANG907/18) || delta > uint32(-(ANG907/18)&0xffff_ffff) {
			// Looking at killer,
			//  so fade damage flash down.
			player.Fmo.Fangle = angle
			if player.Fdamagecount != 0 {
				player.Fdamagecount--
			}
		} else {
			if delta < uint32(ANG1809) {
				player.Fmo.Fangle += uint32(ANG907 / 18)
			} else {
				player.Fmo.Fangle -= uint32(ANG907 / 18)
			}
		}
	} else {
		if player.Fdamagecount != 0 {
			player.Fdamagecount--
		}
	}
	if int32(player.Fcmd.Fbuttons)&bt_USE != 0 {
		player.Fplayerstate = Pst_REBORN
	}
}

// C documentation
//
//	//
//	// P_PlayerThink
//	//
func p_PlayerThink(player *player_t) {
	var newweapon weapontype_t
	// fixme: do this in the cheat code
	if player.Fcheats&CF_NOCLIP != 0 {
		player.Fmo.Fflags |= mf_NOCLIP
	} else {
		player.Fmo.Fflags &^= mf_NOCLIP
	}
	// chain saw run forward
	cmd := &player.Fcmd
	if player.Fmo.Fflags&mf_JUSTATTACKED != 0 {
		cmd.Fangleturn = 0
		cmd.Fforwardmove = int8(0xc800 / 512)
		cmd.Fsidemove = 0
		player.Fmo.Fflags &^= mf_JUSTATTACKED
	}
	if player.Fplayerstate == Pst_DEAD {
		p_DeathThink(player)
		return
	}
	// Move around.
	// Reactiontime is used to prevent movement
	//  for a bit after a teleport.
	if player.Fmo.Freactiontime != 0 {
		player.Fmo.Freactiontime--
	} else {
		p_MovePlayer(player)
	}
	p_CalcHeight(player)
	if player.Fmo.Fsubsector.Fsector.Fspecial != 0 {
		p_PlayerInSpecialSector(player)
	}
	// Check for weapon change.
	// A special event has no other buttons.
	if int32(cmd.Fbuttons)&bt_SPECIAL != 0 {
		cmd.Fbuttons = 0
	}
	if int32(cmd.Fbuttons)&bt_CHANGE != 0 {
		// The actual changing of the weapon is done
		//  when the weapon psprite can do it
		//  (read: not in the middle of an attack).
		newweapon = weapontype_t(int32(cmd.Fbuttons) & bt_WEAPONMASK >> bt_WEAPONSHIFT)
		if newweapon == wp_fist && player.Fweaponowned[wp_chainsaw] != 0 && !(player.Freadyweapon == wp_chainsaw && player.Fpowers[pw_strength] != 0) {
			newweapon = wp_chainsaw
		}
		if gamemode == commercial && newweapon == wp_shotgun && player.Fweaponowned[wp_supershotgun] != 0 && player.Freadyweapon != wp_supershotgun {
			newweapon = wp_supershotgun
		}
		if player.Fweaponowned[newweapon] != 0 && newweapon != player.Freadyweapon {
			// Do not go to plasma or BFG in shareware,
			//  even if cheated.
			if newweapon != wp_plasma && newweapon != wp_bfg || gamemode != shareware {
				player.Fpendingweapon = newweapon
			}
		}
	}
	// check for use
	if int32(cmd.Fbuttons)&bt_USE != 0 {
		if player.Fusedown == 0 {
			p_UseLines(player)
			player.Fusedown = 1
		}
	} else {
		player.Fusedown = 0
	}
	// cycle psprites
	p_MovePsprites(player)
	// Counters, time dependend power ups.
	// Strength counts up to diminish fade.
	if player.Fpowers[pw_strength] != 0 {
		player.Fpowers[pw_strength]++
	}
	if player.Fpowers[pw_invulnerability] != 0 {
		player.Fpowers[pw_invulnerability]--
	}
	if player.Fpowers[pw_invisibility] != 0 {
		player.Fpowers[pw_invisibility]--
		if player.Fpowers[pw_invisibility] == 0 {
			player.Fmo.Fflags &^= mf_SHADOW
		}
	}
	if player.Fpowers[pw_infrared] != 0 {
		player.Fpowers[pw_infrared]--
	}
	if player.Fpowers[pw_ironfeet] != 0 {
		player.Fpowers[pw_ironfeet]--
	}
	if player.Fdamagecount != 0 {
		player.Fdamagecount--
	}
	if player.Fbonuscount != 0 {
		player.Fbonuscount--
	}
	// Handling colormaps.
	if player.Fpowers[pw_invulnerability] != 0 {
		if player.Fpowers[pw_invulnerability] > 4*32 || player.Fpowers[pw_invulnerability]&8 != 0 {
			player.Ffixedcolormap = INVERSECOLORMAP
		} else {
			player.Ffixedcolormap = 0
		}
	} else {
		if player.Fpowers[pw_infrared] != 0 {
			if player.Fpowers[pw_infrared] > 4*32 || player.Fpowers[pw_infrared]&8 != 0 {
				// almost full bright
				player.Ffixedcolormap = 1
			} else {
				player.Ffixedcolormap = 0
			}
		} else {
			player.Ffixedcolormap = 0
		}
	}
}

const NF_SUBSECTOR3 = 32768

// C documentation
//
//	//
//	// R_ClearDrawSegs
//	//
func r_ClearDrawSegs() {
	ds_index = 0
}

// C documentation
//
//	//
//	// ClipWallSegment
//	// Clips the given range of columns
//	// and includes it in the new clip list.
//	//
type cliprange_t struct {
	Ffirst int32
	Flast  int32
}

// C documentation
//
//	//
//	// R_ClipSolidWallSegment
//	// Does handle solid walls,
//	//  e.g. single sided LineDefs (middle texture)
//	//  that entirely block the view.
//	//
func r_ClipSolidWallSegment(first int32, last int32) {
	var v1 int
	var next, start int
	// Find the first range that touches the range
	//  (adjacent pixels are touching).
	for start = 0; solidsegs[start].Flast < first-1; start++ {
	}
	if first < solidsegs[start].Ffirst {
		if last < solidsegs[start].Ffirst-1 {
			// Post is entirely visible (above start),
			//  so insert a new clippost.
			r_StoreWallRange(first, last)
			next = newend
			newend++
			for next != start {
				solidsegs[next] = solidsegs[next-1]
				next--
			}
			solidsegs[next].Ffirst = first
			solidsegs[next].Flast = last
			return
		}
		// There is a fragment above *start.
		r_StoreWallRange(first, solidsegs[start].Ffirst-1)
		// Now adjust the clip size.
		solidsegs[start].Ffirst = first
	}
	// Bottom contained in start?
	if last <= solidsegs[start].Flast {
		return
	}
	next = start
	for last >= solidsegs[next+1].Ffirst-1 {
		// There is a fragment between two posts.
		r_StoreWallRange(solidsegs[next].Flast+1, solidsegs[next+1].Ffirst-1)
		next++
		if last <= solidsegs[next].Flast {
			// Bottom is contained in next.
			// Adjust the clip size.
			solidsegs[start].Flast = solidsegs[next].Flast
			goto crunch
		}
	}
	// There is a fragment after *next.
	r_StoreWallRange(solidsegs[next].Flast+1, last)
	// Adjust the clip size.
	solidsegs[start].Flast = last
	// Remove start+1 to next from the clip list,
	// because start now covers their area.
	goto crunch
crunch:
	;
	if next == start {
		// Post just extended past the bottom of one post.
		return
	}
	for {
		v1 = next
		next++
		if v1 == newend {
			break
		}
		// Remove a post.
		start++
		solidsegs[start] = solidsegs[next]
	}
	newend = start + 1
}

// C documentation
//
//	//
//	// R_ClipPassWallSegment
//	// Clips the given range of columns,
//	//  but does not includes it in the clip list.
//	// Does handle windows,
//	//  e.g. LineDefs with upper and lower texture.
//	//
func r_ClipPassWallSegment(first int32, last int32) {
	var start int
	// Find the first range that touches the range
	//  (adjacent pixels are touching).
	for start = 0; solidsegs[start].Flast < first-1; start++ {
	}
	if first < solidsegs[start].Ffirst {
		if last < solidsegs[start].Ffirst-1 {
			// Post is entirely visible (above start).
			r_StoreWallRange(first, last)
			return
		}
		// There is a fragment above *start.
		r_StoreWallRange(first, solidsegs[start].Ffirst-1)
	}
	// Bottom contained in start?
	if last <= solidsegs[start].Flast {
		return
	}
	for last >= solidsegs[start+1].Ffirst-1 {
		// There is a fragment between two posts.
		r_StoreWallRange(solidsegs[start].Flast+1, solidsegs[start+1].Ffirst-1)
		start++
		if last <= solidsegs[start].Flast {
			return
		}
	}
	// There is a fragment after *next.
	r_StoreWallRange(solidsegs[start].Flast+1, last)
}

// C documentation
//
//	//
//	// R_ClearClipSegs
//	//
func r_ClearClipSegs() {
	solidsegs[0].Ffirst = -int32(0x7fffffff)
	solidsegs[0].Flast = -1
	solidsegs[1].Ffirst = viewwidth
	solidsegs[1].Flast = 0x7fffffff
	newend = 2
}

// C documentation
//
//	//
//	// R_AddLine
//	// Clips the given segment
//	// and adds any visible pieces to the line list.
//	//
func r_AddLine(line *seg_t) {
	var angle1, angle2, span, tspan angle_t
	var x1, x2 int32
	curline = line
	// OPTIMIZE: quickly reject orthogonal back sides.
	angle1 = r_PointToAngle(line.Fv1.Fx, line.Fv1.Fy)
	angle2 = r_PointToAngle(line.Fv2.Fx, line.Fv2.Fy)
	// Clip to view edges.
	// OPTIMIZE: make constant out of 2*clipangle (FIELDOFVIEW).
	span = angle1 - angle2
	// Back side? I.e. backface culling?
	if span >= uint32(ANG1809) {
		return
	}
	// Global angle needed by segcalc.
	rw_angle1 = int32(angle1)
	angle1 -= viewangle
	angle2 -= viewangle
	tspan = angle1 + clipangle
	if tspan > 2*clipangle {
		tspan -= 2 * clipangle
		// Totally off the left edge?
		if tspan >= span {
			return
		}
		angle1 = clipangle
	}
	tspan = clipangle - angle2
	if tspan > 2*clipangle {
		tspan -= 2 * clipangle
		// Totally off the left edge?
		if tspan >= span {
			return
		}
		angle2 = -clipangle
	}
	// The seg is in the view range,
	// but not necessarily visible.
	angle1 = (angle1 + uint32(ANG907)) >> ANGLETOFINESHIFT
	angle2 = (angle2 + uint32(ANG907)) >> ANGLETOFINESHIFT
	x1 = viewangletox[angle1]
	x2 = viewangletox[angle2]
	// Does not cross a pixel?
	if x1 == x2 {
		return
	}
	backsector = line.Fbacksector
	// Single sided line?
	if backsector == nil {
		goto clipsolid
	}
	// Closed door.
	if backsector.Fceilingheight <= frontsector.Ffloorheight || backsector.Ffloorheight >= frontsector.Fceilingheight {
		goto clipsolid
	}
	// Window.
	if backsector.Fceilingheight != frontsector.Fceilingheight || backsector.Ffloorheight != frontsector.Ffloorheight {
		goto clippass
	}
	// Reject empty lines used for triggers
	//  and special events.
	// Identical floor and ceiling on both sides,
	// identical light levels on both sides,
	// and no middle texture.
	if int32(backsector.Fceilingpic) == int32(frontsector.Fceilingpic) && int32(backsector.Ffloorpic) == int32(frontsector.Ffloorpic) && int32(backsector.Flightlevel) == int32(frontsector.Flightlevel) && int32(curline.Fsidedef.Fmidtexture) == 0 {
		return
	}
	goto clippass
clippass:
	;
	r_ClipPassWallSegment(x1, x2-1)
	return
	goto clipsolid
clipsolid:
	;
	r_ClipSolidWallSegment(x1, x2-1)
}

func init() {
	checkcoord = [12][4]int32{
		0: {
			0: 3,
			2: 2,
			3: 1,
		},
		1: {
			0: 3,
			2: 2,
		},
		2: {
			0: 3,
			1: 1,
			2: 2,
		},
		3: {},
		4: {
			0: 2,
			2: 2,
			3: 1,
		},
		5: {},
		6: {
			0: 3,
			1: 1,
			2: 3,
		},
		7: {},
		8: {
			0: 2,
			2: 3,
			3: 1,
		},
		9: {
			0: 2,
			1: 1,
			2: 3,
			3: 1,
		},
		10: {
			0: 2,
			1: 1,
			2: 3,
		},
	}
}

func r_CheckBBox(bspcoord *box_t) boolean {
	var angle1, angle2, span, tspan angle_t
	var boxpos, boxx, boxy, sx1, sx2 int32
	var start int
	var x1, x2, y1, y2 fixed_t
	// Find the corners of the box
	// that define the edges from current viewpoint.
	if viewx <= bspcoord[BOXLEFT] {
		boxx = 0
	} else {
		if viewx < bspcoord[BOXRIGHT] {
			boxx = 1
		} else {
			boxx = 2
		}
	}
	if viewy >= bspcoord[BOXTOP] {
		boxy = 0
	} else {
		if viewy > bspcoord[BOXBOTTOM] {
			boxy = 1
		} else {
			boxy = 2
		}
	}
	boxpos = boxy<<2 + boxx
	if boxpos == 5 {
		return 1
	}

	x1 = bspcoord[checkcoord[boxpos][0]]
	y1 = bspcoord[checkcoord[boxpos][1]]
	x2 = bspcoord[checkcoord[boxpos][2]]
	y2 = bspcoord[checkcoord[boxpos][3]]

	// check clip list for an open space
	angle1 = r_PointToAngle(x1, y1) - viewangle
	angle2 = r_PointToAngle(x2, y2) - viewangle
	span = angle1 - angle2
	// Sitting on a line?
	if span >= uint32(ANG1809) {
		return 1
	}
	tspan = angle1 + clipangle
	if tspan > 2*clipangle {
		tspan -= 2 * clipangle
		// Totally off the left edge?
		if tspan >= span {
			return 0
		}
		angle1 = clipangle
	}
	tspan = clipangle - angle2
	if tspan > 2*clipangle {
		tspan -= 2 * clipangle
		// Totally off the left edge?
		if tspan >= span {
			return 0
		}
		angle2 = -clipangle
	}
	// Find the first clippost
	//  that touches the source post
	//  (adjacent pixels are touching).
	angle1 = (angle1 + uint32(ANG907)) >> ANGLETOFINESHIFT
	angle2 = (angle2 + uint32(ANG907)) >> ANGLETOFINESHIFT
	sx1 = viewangletox[angle1]
	sx2 = viewangletox[angle2]
	// Does not cross a pixel.
	if sx1 == sx2 {
		return 0
	}
	sx2--
	for start = 0; solidsegs[start].Flast < sx2; start++ {
	}
	if sx1 >= solidsegs[start].Ffirst && sx2 <= solidsegs[start].Flast {
		// The clippost contains the new span.
		return 0
	}
	return 1
}

// C documentation
//
//	//
//	// R_Subsector
//	// Determine floor/ceiling planes.
//	// Add sprites of things in sector.
//	// Draw one or more line segments.
//	//
func r_Subsector(num int32) {
	var count, v1 int32
	if num >= numsubsectors {
		i_Error("r_Subsector: ss %d with numss = %d", num, numsubsectors)
	}
	sub := &subsectors[num]
	frontsector = sub.Fsector
	count = int32(sub.Fnumlines)
	if frontsector.Ffloorheight < viewz {
		floorplane = r_FindPlane(frontsector.Ffloorheight, int32(frontsector.Ffloorpic), int32(frontsector.Flightlevel))
	} else {
		floorplane = nil
	}
	if frontsector.Fceilingheight > viewz || int32(frontsector.Fceilingpic) == skyflatnum {
		ceilingplane = r_FindPlane(frontsector.Fceilingheight, int32(frontsector.Fceilingpic), int32(frontsector.Flightlevel))
	} else {
		ceilingplane = nil
	}
	r_AddSprites(frontsector)
	for i := sub.Ffirstline; ; i++ {
		v1 = count
		count--
		if v1 == 0 {
			break
		}
		line := &segs[i]
		r_AddLine(line)
	}
}

// C documentation
//
//	//
//	// RenderBSPNode
//	// Renders all subsectors below a given node,
//	//  traversing subtree recursively.
//	// Just call with BSP root.
func r_RenderBSPNode(bspnum int32) {
	var bsp *node_t
	var side int32
	// Found a subsector?
	if bspnum&int32(NF_SUBSECTOR3) != 0 {
		if bspnum == -1 {
			r_Subsector(0)
		} else {
			r_Subsector(bspnum & ^NF_SUBSECTOR3)
		}
		return
	}
	bsp = &nodes[bspnum]
	// Decide which side the view point is on.
	side = r_PointOnSide(viewx, viewy, &nodes[bspnum])
	// Recursively divide front space.
	r_RenderBSPNode(int32(bsp.Fchildren[side]))
	// Possibly divide back space.
	if r_CheckBBox(&bsp.Fbbox[side^1]) != 0 {
		r_RenderBSPNode(int32(bsp.Fchildren[side^1]))
	}
}

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//  Refresh module, data I/O, caching, retrieval of graphics
//  by name.
//

//
// Graphics.
// DOOM graphics for walls and sprites
// is stored in vertical runs of opaque pixels (posts).
// A column is composed of zero or more posts,
// a patch or sprite is composed of zero or more columns.
//

// C documentation
//
//	//
//	// Texture definition.
//	// Each texture is composed of one or more patches,
//	// with patches being lumps stored in the WAD.
//	// The lumps are referenced by number, and patched
//	// into the rectangular texture space using origin
//	// and possibly other attributes.
//	//
type mappatch_t struct {
	Foriginx  int16
	Foriginy  int16
	Fpatch    int16
	Fstepdir  int16
	Fcolormap int16
}

// C documentation
//
//	//
//	// Texture definition.
//	// A DOOM wall texture is a list of patches
//	// which are to be combined in a predefined order.
//	//
type maptexture_t struct {
	Fname       [8]byte
	Fmasked     int32
	Fwidth      int16
	Fheight     int16
	Fobsolete   int32
	Fpatchcount int16
	Fpatches    [1]mappatch_t
}

// C documentation
//
//	// A single patch from a texture definition,
//	//  basically a rectangular area within
//	//  the texture rectangle.
type texpatch_t struct {
	Foriginx int16
	Foriginy int16
	Fpatch   int32
}

// A maptexturedef_t describes a rectangular texture,
//  which is composed of one or more mappatch_t structures
//  that arrange graphic patches.

type texture_t struct {
	Fname       [8]byte
	Fwidth      int16
	Fheight     int16
	Findex      int32
	Fnext       *texture_t
	Fpatchcount int16
	Fpatches    []texpatch_t
}

//
// MAPTEXTURE_T CACHING
// When a texture is first needed,
//  it counts the number of composite columns
//  required in the texture and allocates space
//  for a column directory and any new columns.
// The directory will simply point inside other patches
//  if there is only one patch in a given column,
//  but any columns with multiple patches
//  will have new column_ts generated.
//

// C documentation
//
//	//
//	// R_DrawColumnInCache
//	// Clip and draw a column
//	//  from a patch into a cached post.
//	//
func r_DrawColumnInCache(patch *column_t, cache []byte, originy int32, cacheheight int32) {
	for int32(patch.Ftopdelta) != 0xff {
		source := patch.Data()
		count := int32(len(source))
		position := originy + int32(patch.Ftopdelta)
		if position < 0 {
			count += position
			position = 0
		}
		if position+count > cacheheight {
			count = cacheheight - position
		}
		if count > 0 {
			copy(cache[position:], source[:count])
		}
		patch = patch.Next()
	}
}

// C documentation
//
//	//
//	// R_GenerateComposite
//	// Using the texture definition,
//	//  the composite texture is created from the patches,
//	//  and each column is cached.
//	//
func r_GenerateComposite(texnum int32) {
	var collump []int16
	var colofs []uint16
	var i, x, x1, x2 int32
	var texture *texture_t
	texture = textures[texnum]
	texturecomposite[texnum] = make([]byte, texturecompositesize[texnum])
	collump = texturecolumnlump[texnum]
	colofs = texturecolumnofs[texnum]
	// Composite the columns together.
	i = 0
	for {
		if i >= int32(texture.Fpatchcount) {
			break
		}
		patch := &texture.Fpatches[i]
		realpatch := w_CacheLumpNumT(patch.Fpatch)
		x1 = int32(patch.Foriginx)
		x2 = x1 + int32(realpatch.Fwidth)
		if x1 < 0 {
			x = 0
		} else {
			x = x1
		}
		if x2 > int32(texture.Fwidth) {
			x2 = int32(texture.Fwidth)
		}
		for ; x < x2; x++ {
			// Column does not have multiple patches?
			if collump[x] >= 0 {
				continue
			}
			patchcol := realpatch.GetColumn(x - x1)
			r_DrawColumnInCache(patchcol, texturecomposite[texnum][colofs[x]:], int32(patch.Foriginy), int32(texture.Fheight))
		}
		goto _1
	_1:
		;
		i++
	}
	// Now that the texture has been built in column cache,
	//  it is purgable from zone memory.
}

// C documentation
//
//	//
//	// R_GenerateLookup
//	//
func r_GenerateLookup(texnum int32) {
	var realpatch *patch_t
	var colofs []uint16
	var collump []int16
	var texture *texture_t
	var i, x, x1, x2 int32
	texture = textures[texnum]
	// Composited texture not created yet.
	texturecomposite[texnum] = nil
	texturecompositesize[texnum] = 0
	collump = texturecolumnlump[texnum]
	colofs = texturecolumnofs[texnum]
	// Now count the number of columns
	//  that are covered by more than one patch.
	// Fill in the lump / offset, so columns
	//  with only a single patch are all done.
	widths := make([]int32, texture.Fwidth)
	i = 0
	for {
		if i >= int32(texture.Fpatchcount) {
			break
		}
		patch := &texture.Fpatches[i]
		realpatch = w_CacheLumpNumT(patch.Fpatch)
		x1 = int32(patch.Foriginx)
		x2 = x1 + int32(realpatch.Fwidth)
		if x1 < 0 {
			x = 0
		} else {
			x = x1
		}
		if x2 > int32(texture.Fwidth) {
			x2 = int32(texture.Fwidth)
		}
		for {
			if x >= x2 {
				break
			}
			widths[x]++
			collump[x] = int16(patch.Fpatch)
			colofs[x] = uint16(realpatch.Fcolumnofs[x-x1] + 3)
			goto _2
		_2:
			;
			x++
		}
		goto _1
	_1:
		;
		i++
	}
	x = 0
	for {
		if x >= int32(texture.Fwidth) {
			break
		}
		if widths[x] == 0 {
			fprintf_ccgo(os.Stdout, "r_GenerateLookup: column without a patch (%s)\n", gostring_bytes(texture.Fname[:]))
			return
		}
		// i_Error ("r_GenerateLookup: column without a patch");
		if int32(widths[x]) > 1 {
			// Use the cached block.
			collump[x] = -1
			colofs[x] = uint16(texturecompositesize[texnum])
			if texturecompositesize[texnum] > 0x10000-int32(texture.Fheight) {
				i_Error("r_GenerateLookup: texture %d is >64k", texnum)
			}
			texturecompositesize[texnum] += int32(texture.Fheight)
		}
		goto _3
	_3:
		;
		x++
	}
}

// C documentation
//
//	//
//	// R_GetColumn
//	//
func r_GetColumn(tex int32, col int32) uintptr {
	var lump, ofs int32
	col &= texturewidthmask[tex]
	lump = int32(texturecolumnlump[tex][col])
	ofs = int32(texturecolumnofs[tex][col])
	if lump > 0 {
		return w_CacheLumpNum(lump) + uintptr(ofs)
	}
	if texturecomposite[tex] == nil {
		r_GenerateComposite(tex)
	}
	return *(*uintptr)(unsafe.Pointer(&texturecomposite[tex])) + uintptr(ofs)
}

func generateTextureHashTable() {
	var i, key int32
	var rover **texture_t
	textures_hashtable = make([]*texture_t, numtextures)
	// Add all textures to hash table
	i = 0
	for {
		if i >= numtextures {
			break
		}
		// Store index
		textures[i].Findex = i
		// Vanilla Doom does a linear search of the texures array
		// and stops at the first entry it finds.  If there are two
		// entries with the same name, the first one in the array
		// wins. The new entry must therefore be added at the end
		// of the hash chain, so that earlier entries win.
		key = int32(w_LumpNameHash(gostring_bytes(textures[i].Fname[:])) % uint32(numtextures))
		rover = &textures_hashtable[key]
		for *rover != nil {
			rover = &(*rover).Fnext
		}
		// Hook into hash table
		textures[i].Fnext = nil
		*rover = textures[i]
		goto _1
	_1:
		;
		i++
	}
}

// C documentation
//
//	//
//	// R_InitTextures
//	// Initializes the texture list
//	//  with the textures from the world map.
//	//
func r_InitTextures() {
	var directory, maptex, maptex2, mpatch, name_p, names, v2 uintptr
	var i, j, maxoff, maxoff2, nummappatches, numtextures1, numtextures2, offset, temp1, temp2, temp3, totalwidth int32
	// Load the patch names from pnames.lmp.
	names = w_CacheLumpName("PNAMES")
	nummappatches = *(*int32)(unsafe.Pointer(names))
	name_p = names + uintptr(4)
	patchlookup := make([]int32, nummappatches)
	i = 0
	for {
		if i >= nummappatches {
			break
		}
		patchlookup[i] = w_CheckNumForName(gostring_n(name_p+uintptr(i*8), 8))
		goto _1
	_1:
		;
		i++
	}
	w_ReleaseLumpName("PNAMES")
	// Load the map texture definitions from textures.lmp.
	// The data is contained in one or two lumps,
	//  TEXTURE1 for shareware, plus TEXTURE2 for commercial.
	v2 = w_CacheLumpName("TEXTURE1")
	maptex = v2
	numtextures1 = *(*int32)(unsafe.Pointer(maptex))
	maxoff = w_LumpLength(uint32(w_GetNumForName("TEXTURE1")))
	directory = maptex + uintptr(1)*4
	if w_CheckNumForName("TEXTURE2") != -1 {
		maptex2 = w_CacheLumpName("TEXTURE2")
		numtextures2 = *(*int32)(unsafe.Pointer(maptex2))
		maxoff2 = w_LumpLength(uint32(w_GetNumForName("TEXTURE2")))
	} else {
		maptex2 = 0
		numtextures2 = 0
		maxoff2 = 0
	}
	numtextures = numtextures1 + numtextures2
	textures = make([]*texture_t, numtextures)
	texturecolumnlump = make([][]int16, numtextures)
	texturecolumnofs = make([][]uint16, numtextures)
	texturecomposite = make([][]byte, numtextures)
	texturecompositesize = make([]int32, numtextures)
	texturewidthmask = make([]int32, numtextures)
	textureheight = make([]fixed_t, numtextures)
	totalwidth = 0
	//	Really complex printing shit...
	temp1 = w_GetNumForName("S_START") // P_???????
	temp2 = w_GetNumForName("S_END") - 1
	temp3 = (temp2-temp1+int32(63))/int32(64) + (numtextures+int32(63))/int32(64)
	// If stdout is a real console, use the classic vanilla "filling
	// up the box" effect, which uses backspace to "step back" inside
	// the box.  If stdout is a file, don't draw the box.
	if i_ConsoleStdout() != 0 {
		fprintf_ccgo(os.Stdout, "[")
		i = 0
		for {
			if i >= temp3+9 {
				break
			}
			fprintf_ccgo(os.Stdout, " ")
			goto _3
		_3:
			;
			i++
		}
		fprintf_ccgo(os.Stdout, "]")
		i = 0
		for {
			if i >= temp3+int32(10) {
				break
			}
			fprintf_ccgo(os.Stdout, "\x08")
			goto _4
		_4:
			;
			i++
		}
	}
	i = 0
	for {
		if i >= numtextures {
			break
		}
		if i&63 == 0 {
			fprintf_ccgo(os.Stdout, ".")
		}
		if i == numtextures1 {
			// Start looking in second texture file.
			maptex = maptex2
			maxoff = maxoff2
			directory = maptex + uintptr(1)*4
		}
		offset = *(*int32)(unsafe.Pointer(directory))
		if offset > maxoff {
			i_Error("r_InitTextures: bad texture directory")
		}
		mtexture := (*maptexture_t)(unsafe.Pointer(maptex + uintptr(offset)))
		texture := &texture_t{
			Fpatches: make([]texpatch_t, mtexture.Fpatchcount),
		}
		textures[i] = texture
		texture.Fwidth = mtexture.Fwidth
		texture.Fheight = mtexture.Fheight
		texture.Fpatchcount = mtexture.Fpatchcount
		copy(texture.Fname[:], mtexture.Fname[:])
		mpatch = uintptr(unsafe.Pointer(&mtexture.Fpatches[0]))
		j = 0
		for {
			if j >= int32(texture.Fpatchcount) {
				break
			}
			patch := &texture.Fpatches[j]
			patch.Foriginx = (*mappatch_t)(unsafe.Pointer(mpatch)).Foriginx
			patch.Foriginy = (*mappatch_t)(unsafe.Pointer(mpatch)).Foriginy
			patch.Fpatch = patchlookup[(*mappatch_t)(unsafe.Pointer(mpatch)).Fpatch]
			if patch.Fpatch == -1 {
				i_Error("r_InitTextures: Missing patch in texture %s", gostring_bytes(texture.Fname[:]))
			}
			goto _7
		_7:
			;
			j++
			mpatch += 10
		}
		texturecolumnlump[i] = make([]int16, texture.Fwidth)
		texturecolumnofs[i] = make([]uint16, texture.Fwidth)
		j = 1
		for j*2 <= int32(texture.Fwidth) {
			j <<= 1
		}
		texturewidthmask[i] = j - 1
		textureheight[i] = int32(texture.Fheight) << FRACBITS
		totalwidth += int32(texture.Fwidth)
		goto _5
	_5:
		;
		i++
		directory += 4
	}
	w_ReleaseLumpName("TEXTURE1")
	if maptex2 != 0 {
		w_ReleaseLumpName("TEXTURE2")
	}
	// Precalculate whatever possible.
	i = 0
	for {
		if i >= numtextures {
			break
		}
		r_GenerateLookup(i)
		goto _8
	_8:
		;
		i++
	}
	// Create translation table for global animation.
	texturetranslation = make([]int32, numtextures+1)
	i = 0
	for {
		if i >= numtextures {
			break
		}
		texturetranslation[i] = i
		goto _9
	_9:
		;
		i++
	}
	generateTextureHashTable()
}

// C documentation
//
//	//
//	// R_InitFlats
//	//
func r_InitFlats() {
	var i int32
	firstflat = w_GetNumForName("F_START") + 1
	lastflat = w_GetNumForName("F_END") - 1
	numflats = lastflat - firstflat + 1
	// Create translation table for global animation.
	flattranslation = make([]int32, numflats+1)
	i = 0
	for {
		if i >= numflats {
			break
		}
		flattranslation[i] = i
		goto _1
	_1:
		;
		i++
	}
}

// C documentation
//
//	//
//	// R_InitSpriteLumps
//	// Finds the width and hoffset of all sprites in the wad,
//	//  so the sprite does not need to be cached completely
//	//  just for having the header info ready during rendering.
//	//
func r_InitSpriteLumps() {
	var i int32
	firstspritelump = w_GetNumForName("S_START") + 1
	lastspritelump = w_GetNumForName("S_END") - 1
	numspritelumps = lastspritelump - firstspritelump + 1
	spritewidth = make([]fixed_t, numspritelumps)
	spriteoffset = make([]fixed_t, numspritelumps)
	spritetopoffset = make([]fixed_t, numspritelumps)
	i = 0
	for {
		if i >= numspritelumps {
			break
		}
		if i&63 == 0 {
			fprintf_ccgo(os.Stdout, ".")
		}
		patch := w_CacheLumpNumT[*patch_t](firstspritelump + i)
		spritewidth[i] = int32(patch.Fwidth) << FRACBITS
		spriteoffset[i] = int32(patch.Fleftoffset) << FRACBITS
		spritetopoffset[i] = int32(patch.Ftopoffset) << FRACBITS
		goto _1
	_1:
		;
		i++
	}
}

// C documentation
//
//	//
//	// R_InitColormaps
//	//
func r_InitColormaps() {
	var lump int32
	var size int32
	var data uintptr
	// Load in the light tables,
	//  256 byte align tables.
	lump = w_GetNumForName("COLORMAP")
	size = w_LumpLength(uint32(lump))
	data = w_CacheLumpNum(lump)
	colormaps = unsafe.Slice((*lighttable_t)(unsafe.Pointer(data)), size)
}

// C documentation
//
//	//
//	// R_InitData
//	// Locates all the lumps
//	//  that will be used by all views
//	// Must be called after W_Init.
//	//
func r_InitData() {
	r_InitTextures()
	fprintf_ccgo(os.Stdout, ".")
	r_InitFlats()
	fprintf_ccgo(os.Stdout, ".")
	r_InitSpriteLumps()
	fprintf_ccgo(os.Stdout, ".")
	r_InitColormaps()
}

// C documentation
//
//	//
//	// R_FlatNumForName
//	// Retrieval, get a flat number for a flat name.
//	//
func r_FlatNumForName(name string) int32 {
	var i int32
	i = w_CheckNumForName(name)
	if i == -1 {
		i_Error("r_FlatNumForName: %s not found", name)
	}
	return i - firstflat
}

// C documentation
//
//	//
//	// R_CheckTextureNumForName
//	// Check whether texture is available.
//	// Filter out NoTexture indicator.
//	//
func r_CheckTextureNumForName(name string) int32 {
	var key int32
	var texture *texture_t
	// "NoTexture" marker.
	if name[0] == '-' {
		return 0
	}
	key = int32(w_LumpNameHash(name) % uint32(numtextures))
	texture = textures_hashtable[key]
	for texture != nil {
		if strings.EqualFold(gostring_bytes(texture.Fname[:]), name) {
			return texture.Findex
		}
		texture = texture.Fnext
	}
	return -1
}

// C documentation
//
//	//
//	// R_TextureNumForName
//	// Calls r_CheckTextureNumForName,
//	//  aborts with error message.
//	//
func r_TextureNumForName(name string) int32 {
	var i int32
	i = r_CheckTextureNumForName(name)
	if i == -1 {
		i_Error("r_TextureNumForName: %s not found", name)
	}
	return i
}

func r_PrecacheLevel() {
	var texture *texture_t
	var sf *spriteframe_t
	var th *thinker_t
	var i, j, k, lump int32
	if demoplayback != 0 {
		return
	}
	// Precache flats.
	flatpresent := make([]int8, numflats)
	i = 0
	for {
		if i >= numsectors {
			break
		}
		sector := &sectors[i]
		flatpresent[sector.Ffloorpic] = 1
		flatpresent[sector.Fceilingpic] = 1
		goto _1
	_1:
		;
		i++
	}
	i = 0
	for {
		if i >= numflats {
			break
		}
		if flatpresent[i] != 0 {
			lump = firstflat + i
			w_CacheLumpNum(lump)
		}
		goto _2
	_2:
		;
		i++
	}
	// Precache textures.
	texturepresent := make([]int8, numtextures)
	i = 0
	for {
		if i >= numsides {
			break
		}
		texturepresent[sides[i].Ftoptexture] = 1
		texturepresent[sides[i].Fmidtexture] = 1
		texturepresent[sides[i].Fbottomtexture] = 1
		goto _3
	_3:
		;
		i++
	}
	// Sky texture is always present.
	// Note that F_SKY1 is the name used to
	//  indicate a sky floor/ceiling as a flat,
	//  while the sky texture is stored like
	//  a wall texture, with an episode dependend
	//  name.
	texturepresent[skytexture] = 1
	i = 0
	for {
		if i >= numtextures {
			break
		}
		if texturepresent[i] != 0 {
			goto _4
		}
		texture = textures[i]
		j = 0
		for {
			if j >= int32(texture.Fpatchcount) {
				break
			}
			lump = texture.Fpatches[j].Fpatch
			w_CacheLumpNum(lump)
			goto _5
		_5:
			;
			j++
		}
		goto _4
	_4:
		;
		i++
	}
	// Precache sprites.
	spritepresent := make([]int8, numsprites)
	th = thinkercap.Fnext
	for {
		if th == &thinkercap {
			break
		}
		if mo, ok := th.Ffunction.(*mobj_t); ok {
			spritepresent[mo.Fsprite] = 1
		}
		goto _6
	_6:
		;
		th = th.Fnext
	}
	i = 0
	for {
		if i >= numsprites {
			break
		}
		if spritepresent[i] == 0 {
			goto _7
		}
		j = 0
		for {
			if j >= sprites[i].Fnumframes {
				break
			}
			sf = &sprites[i].Fspriteframes[j]
			k = 0
			for {
				if k >= 8 {
					break
				}
				lump = firstspritelump + int32(sf.Flump[k])
				w_CacheLumpNum(lump)
				goto _9
			_9:
				;
				k++
			}
			goto _8
		_8:
			;
			j++
		}
		goto _7
	_7:
		;
		i++
	}
}

const FUZZTABLE = 50
const SBARHEIGHT = 32

// Backing buffer containing the bezel drawn around the screen and
// surrounding background.

var background_buffer []byte

// C documentation
//
//	//
//	// A column is a vertical slice/span from a wall texture that,
//	//  given the DOOM style restrictions on the view orientation,
//	//  will always have constant z depth.
//	// Thus a special case loop for very fast rendering can
//	//  be used. It has also been used with Wolfenstein 3D.
//	//
func r_DrawColumn() {
	var count, v1 int32
	var dest int32
	var frac, fracstep fixed_t
	count = dc_yh - dc_yl
	// Zero length, column does not exceed a pixel.
	if count < 0 {
		return
	}
	if uint32(dc_x) >= SCREENWIDTH || dc_yl < 0 || dc_yh >= SCREENHEIGHT {
		i_Error("r_DrawColumn: %d to %d at %d", dc_yl, dc_yh, dc_x)
	}
	// Framebuffer destination address.
	// Use ylookup LUT to avoid multiply with ScreenWidth.
	// Use columnofs LUT for subwindows?
	dest = ylookup[dc_yl] + (columnofs[dc_x])
	// Determine scaling,
	//  which is the only mapping to be done.
	fracstep = dc_iscale
	frac = dc_texturemid + (dc_yl-centery)*fracstep
	// Inner loop that does the actual texture mapping,
	//  e.g. a DDA-lile scaling.
	// This is as fast as it gets.
	for {
		// Re-map color indices from wall texture column
		//  using a lighting/special effects LUT.
		I_VideoBuffer[dest] = dc_colormap[*(*uint8)(unsafe.Pointer(dc_source + uintptr(frac>>FRACBITS&int32(127))))]
		dest += SCREENWIDTH
		frac += fracstep
		goto _2
	_2:
		;
		v1 = count
		count--
		if v1 == 0 {
			break
		}
	}
}

// UNUSED.
// Loop unrolled.

func r_DrawColumnLow() {
	var count, x, v1 int32
	var dest, dest2 int32
	var frac, fracstep fixed_t
	var v3 uint8
	count = dc_yh - dc_yl
	// Zero length.
	if count < 0 {
		return
	}
	if uint32(dc_x) >= SCREENWIDTH || dc_yl < 0 || dc_yh >= SCREENHEIGHT {
		i_Error("r_DrawColumn: %d to %d at %d", dc_yl, dc_yh, dc_x)
	}
	//	dccount++;
	// Blocky mode, need to multiply by 2.
	x = dc_x << 1
	dest = ylookup[dc_yl] + (columnofs[x])
	dest2 = ylookup[dc_yl] + (columnofs[x+1])
	fracstep = dc_iscale
	frac = dc_texturemid + (dc_yl-centery)*fracstep
	for {
		// Hack. Does not work corretly.
		v3 = dc_colormap[*(*uint8)(unsafe.Pointer(dc_source + uintptr(frac>>FRACBITS&int32(127))))]
		I_VideoBuffer[dest] = v3
		I_VideoBuffer[dest2] = v3
		dest += SCREENWIDTH
		dest2 += SCREENWIDTH
		frac += fracstep
		goto _2
	_2:
		;
		v1 = count
		count--
		if v1 == 0 {
			break
		}
	}
}

func init() {
	fuzzoffset = [50]int32{
		0:  SCREENWIDTH,
		1:  -SCREENWIDTH,
		2:  SCREENWIDTH,
		3:  -SCREENWIDTH,
		4:  SCREENWIDTH,
		5:  SCREENWIDTH,
		6:  -SCREENWIDTH,
		7:  SCREENWIDTH,
		8:  SCREENWIDTH,
		9:  -SCREENWIDTH,
		10: SCREENWIDTH,
		11: SCREENWIDTH,
		12: SCREENWIDTH,
		13: -SCREENWIDTH,
		14: SCREENWIDTH,
		15: SCREENWIDTH,
		16: SCREENWIDTH,
		17: -SCREENWIDTH,
		18: -SCREENWIDTH,
		19: -SCREENWIDTH,
		20: -SCREENWIDTH,
		21: SCREENWIDTH,
		22: -SCREENWIDTH,
		23: -SCREENWIDTH,
		24: SCREENWIDTH,
		25: SCREENWIDTH,
		26: SCREENWIDTH,
		27: SCREENWIDTH,
		28: -SCREENWIDTH,
		29: SCREENWIDTH,
		30: -SCREENWIDTH,
		31: SCREENWIDTH,
		32: SCREENWIDTH,
		33: -SCREENWIDTH,
		34: -SCREENWIDTH,
		35: SCREENWIDTH,
		36: SCREENWIDTH,
		37: -SCREENWIDTH,
		38: -SCREENWIDTH,
		39: -SCREENWIDTH,
		40: -SCREENWIDTH,
		41: SCREENWIDTH,
		42: SCREENWIDTH,
		43: SCREENWIDTH,
		44: SCREENWIDTH,
		45: -SCREENWIDTH,
		46: SCREENWIDTH,
		47: SCREENWIDTH,
		48: -SCREENWIDTH,
		49: SCREENWIDTH,
	}
}

// C documentation
//
//	//
//	// Framebuffer postprocessing.
//	// Creates a fuzzy image by copying pixels
//	//  from adjacent ones to left and right.
//	// Used with an all black colormap, this
//	//  could create the SHADOW effect,
//	//  i.e. spectres and invisible players.
//	//
func r_DrawFuzzColumn() {
	var count, v1, v3 int32
	var dest int32
	var frac, fracstep fixed_t
	// Adjust borders. Low...
	if dc_yl == 0 {
		dc_yl = 1
	}
	// .. and high.
	if dc_yh == viewheight-1 {
		dc_yh = viewheight - 2
	}
	count = dc_yh - dc_yl
	// Zero length.
	if count < 0 {
		return
	}
	if uint32(dc_x) >= SCREENWIDTH || dc_yl < 0 || dc_yh >= SCREENHEIGHT {
		i_Error("r_DrawFuzzColumn: %d to %d at %d", dc_yl, dc_yh, dc_x)
	}
	dest = ylookup[dc_yl] + (columnofs[dc_x])
	// Looks familiar.
	fracstep = dc_iscale
	frac = dc_texturemid + (dc_yl-centery)*fracstep
	// Looks like an attempt at dithering,
	//  using the colormap #6 (of 0-31, a bit
	//  brighter than average).
	for {
		// Lookup framebuffer, and retrieve
		//  a pixel that is either one column
		//  left or right of the current one.
		// Add index from colormap to index.
		I_VideoBuffer[dest] = colormaps[6*256+int(I_VideoBuffer[dest+fuzzoffset[fuzzpos]])]
		// Clamp table lookup index.
		fuzzpos++
		v3 = fuzzpos
		if v3 == FUZZTABLE {
			fuzzpos = 0
		}
		dest += SCREENWIDTH
		frac += fracstep
		goto _2
	_2:
		;
		v1 = count
		count--
		if v1 == 0 {
			break
		}
	}
}

// low detail mode version
func r_DrawFuzzColumnLow() {
	var count, x, v1, v3 int32
	var dest, dest2 int32
	var frac, fracstep fixed_t
	// Adjust borders. Low...
	if dc_yl == 0 {
		dc_yl = 1
	}
	// .. and high.
	if dc_yh == viewheight-1 {
		dc_yh = viewheight - 2
	}
	count = dc_yh - dc_yl
	// Zero length.
	if count < 0 {
		return
	}
	// low detail mode, need to multiply by 2
	x = dc_x << 1
	if uint32(x) >= SCREENWIDTH || dc_yl < 0 || dc_yh >= SCREENHEIGHT {
		i_Error("r_DrawFuzzColumn: %d to %d at %d", dc_yl, dc_yh, dc_x)
	}
	dest = ylookup[dc_yl] + (columnofs[x])
	dest2 = ylookup[dc_yl] + (columnofs[x+1])
	// Looks familiar.
	fracstep = dc_iscale
	frac = dc_texturemid + (dc_yl-centery)*fracstep
	// Looks like an attempt at dithering,
	//  using the colormap #6 (of 0-31, a bit
	//  brighter than average).
	for {
		// Lookup framebuffer, and retrieve
		//  a pixel that is either one column
		//  left or right of the current one.
		// Add index from colormap to index.
		I_VideoBuffer[dest] = colormaps[6*256+int(I_VideoBuffer[dest+fuzzoffset[fuzzpos]])]
		I_VideoBuffer[dest2] = colormaps[6*256+int(I_VideoBuffer[dest2+fuzzoffset[fuzzpos]])]
		// Clamp table lookup index.
		fuzzpos++
		v3 = fuzzpos
		if v3 == FUZZTABLE {
			fuzzpos = 0
		}
		dest += SCREENWIDTH
		dest2 += SCREENWIDTH
		frac += fracstep
		goto _2
	_2:
		;
		v1 = count
		count--
		if v1 == 0 {
			break
		}
	}
}

func r_DrawTranslatedColumn() {
	var count, v1 int32
	var dest int32
	var frac, fracstep fixed_t
	count = dc_yh - dc_yl
	if count < 0 {
		return
	}
	if uint32(dc_x) >= SCREENWIDTH || dc_yl < 0 || dc_yh >= SCREENHEIGHT {
		i_Error("r_DrawColumn: %d to %d at %d", dc_yl, dc_yh, dc_x)
	}
	dest = ylookup[dc_yl] + (columnofs[dc_x])
	// Looks familiar.
	fracstep = dc_iscale
	frac = dc_texturemid + (dc_yl-centery)*fracstep
	// Here we do an additional index re-mapping.
	for {
		// Translation tables are used
		//  to map certain colorramps to other ones,
		//  used with PLAY sprites.
		// Thus the "green" ramp of the player 0 sprite
		//  is mapped to gray, red, black/indigo.
		I_VideoBuffer[dest] = dc_colormap[dc_translation[*(*uint8)(unsafe.Pointer(dc_source + uintptr(frac>>FRACBITS)))]]
		dest += SCREENWIDTH
		frac += fracstep
		goto _2
	_2:
		;
		v1 = count
		count--
		if v1 == 0 {
			break
		}
	}
}

func r_DrawTranslatedColumnLow() {
	var count, x, v1 int32
	var dest, dest2 int32
	var frac, fracstep fixed_t
	count = dc_yh - dc_yl
	if count < 0 {
		return
	}
	// low detail, need to scale by 2
	x = dc_x << 1
	if uint32(x) >= SCREENWIDTH || dc_yl < 0 || dc_yh >= SCREENHEIGHT {
		i_Error("r_DrawColumn: %d to %d at %d", dc_yl, dc_yh, x)
	}
	dest = ylookup[dc_yl] + (columnofs[x])
	dest2 = ylookup[dc_yl] + (columnofs[x+1])
	// Looks familiar.
	fracstep = dc_iscale
	frac = dc_texturemid + (dc_yl-centery)*fracstep
	// Here we do an additional index re-mapping.
	for {
		// Translation tables are used
		//  to map certain colorramps to other ones,
		//  used with PLAY sprites.
		// Thus the "green" ramp of the player 0 sprite
		//  is mapped to gray, red, black/indigo.
		I_VideoBuffer[dest] = dc_colormap[dc_translation[*(*uint8)(unsafe.Pointer(dc_source + uintptr(frac>>FRACBITS)))]]
		I_VideoBuffer[dest2] = dc_colormap[dc_translation[*(*uint8)(unsafe.Pointer(dc_source + uintptr(frac>>FRACBITS)))]]
		dest += SCREENWIDTH
		dest2 += SCREENWIDTH
		frac += fracstep
		goto _2
	_2:
		;
		v1 = count
		count--
		if v1 == 0 {
			break
		}
	}
}

// C documentation
//
//	//
//	// R_InitTranslationTables
//	// Creates the translation tables to map
//	//  the green color ramp to gray, brown, red.
//	// Assumes a given structure of the PLAYPAL.
//	// Could be read from a lump instead.
//	//
func r_InitTranslationTables() {
	var i int32
	translationtables = make([]byte, 256*3)
	// translate just the 16 green colors
	i = 0
	for {
		if i >= 256 {
			break
		}
		if i >= 0x70 && i <= 0x7f {
			// map green ramp to gray, brown, red
			translationtables[i] = uint8(0x60 + i&int32(0xf))
			translationtables[i+256] = uint8(0x40 + i&int32(0xf))
			translationtables[i+512] = uint8(0x20 + i&int32(0xf))
		} else {
			// Keep all other colors as is.
			translationtables[i+512] = uint8(i)
			translationtables[i+256] = uint8(i)
			translationtables[i] = uint8(i)
		}
		goto _1
	_1:
		;
		i++
	}
}

// C documentation
//
//	//
//	// Draws the actual span.
func r_DrawSpan() {
	var count, spot, v1, dest int32
	var position, step, xtemp, ytemp uint32
	if ds_x2 < ds_x1 || ds_x1 < 0 || ds_x2 >= SCREENWIDTH || uint32(ds_y) > SCREENHEIGHT {
		i_Error("r_DrawSpan: %d to %d at %d", ds_x1, ds_x2, ds_y)
	}
	//	dscount++;
	// Pack position and step variables into a single 32-bit integer,
	// with x in the top 16 bits and y in the bottom 16 bits.  For
	// each 16-bit part, the top 6 bits are the integer part and the
	// bottom 10 bits are the fractional part of the pixel position.
	position = uint32(ds_xfrac<<10)&0xffff0000 | uint32(ds_yfrac>>6&0x0000ffff)
	step = uint32(ds_xstep<<10)&0xffff0000 | uint32(ds_ystep>>6&0x0000ffff)
	dest = ylookup[ds_y] + columnofs[ds_x1]
	// We do not check for zero spans here?
	count = ds_x2 - ds_x1
	for {
		// Calculate current texture index in u,v.
		ytemp = position >> 4 & 0x0fc0
		xtemp = position >> 26
		spot = int32(xtemp | ytemp)
		// Lookup pixel from flat texture tile,
		//  re-index using light/colormap.
		I_VideoBuffer[dest] = ds_colormap[ds_source[spot]]
		dest++
		position += step
		goto _2
	_2:
		;
		v1 = count
		count--
		if v1 == 0 {
			break
		}
	}
}

// UNUSED.
// Loop unrolled by 4.

// C documentation
//
//	//
//	// Again..
//	//
func r_DrawSpanLow() {
	var count, spot, v1, dest int32
	var position, step, xtemp, ytemp uint32
	if ds_x2 < ds_x1 || ds_x1 < 0 || ds_x2 >= SCREENWIDTH || uint32(ds_y) > SCREENHEIGHT {
		i_Error("r_DrawSpan: %d to %d at %d", ds_x1, ds_x2, ds_y)
	}
	//	dscount++;
	position = uint32(ds_xfrac<<10)&0xffff0000 | uint32(ds_yfrac>>6&0x0000ffff)
	step = uint32(ds_xstep<<10)&0xffff0000 | uint32(ds_ystep>>6&0x0000ffff)
	count = ds_x2 - ds_x1
	// Blocky mode, need to multiply by 2.
	ds_x1 <<= 1
	ds_x2 <<= 1
	dest = ylookup[ds_y] + columnofs[ds_x1]
	for {
		// Calculate current texture index in u,v.
		ytemp = position >> 4 & 0x0fc0
		xtemp = position >> 26
		spot = int32(xtemp | ytemp)
		// Lowres/blocky mode does it twice,
		//  while scale is adjusted appropriately.
		I_VideoBuffer[dest] = ds_colormap[ds_source[spot]]
		dest++
		I_VideoBuffer[dest] = ds_colormap[ds_source[spot]]
		dest++
		position += step
		goto _2
	_2:
		;
		v1 = count
		count--
		if v1 == 0 {
			break
		}
	}
}

// C documentation
//
//	//
//	// R_InitBuffer
//	// Creats lookup tables that avoid
//	//  multiplies and other hazzles
//	//  for getting the framebuffer address
//	//  of a pixel to draw.
//	//
func r_InitBuffer(width int32, height int32) {
	var i int32
	// Handle resize,
	//  e.g. smaller view windows
	//  with border and/or status bar.
	viewwindowx = (SCREENWIDTH - width) >> 1
	// Column offset. For windows.
	i = 0
	for {
		if i >= width {
			break
		}
		columnofs[i] = viewwindowx + i
		goto _1
	_1:
		;
		i++
	}
	// Samw with base row offset.
	if width == SCREENWIDTH {
		viewwindowy = 0
	} else {
		viewwindowy = (SCREENHEIGHT - SBARHEIGHT - height) >> 1
	}
	// Preclaculate all row offsets.
	i = 0
	for {
		if i >= height {
			break
		}
		ylookup[i] = (i + viewwindowy) * SCREENWIDTH
		goto _2
	_2:
		;
		i++
	}
}

// C documentation
//
//	//
//	// R_FillBackScreen
//	// Fills the back screen with a pattern
//	//  for variable screen sizes
//	// Also draws a beveled edge.
//	//
func r_FillBackScreen() {
	var src []byte
	var name, name1, name2 string
	var patch *patch_t
	var x, y int32
	// DOOM border patch.
	name1 = "FLOOR7_2"
	// DOOM II border patch.
	name2 = "GRNROCK"
	// If we are running full screen, there is no need to do any of this,
	// and the background buffer can be freed if it was previously in use.
	if scaledviewwidth == SCREENWIDTH {
		if background_buffer != nil {
			background_buffer = nil
		}
		return
	}
	// Allocate the background buffer if necessary
	if background_buffer == nil {
		background_buffer = make([]byte, SCREENWIDTH*(SCREENHEIGHT-SBARHEIGHT))
	}
	if gamemode == commercial {
		name = name2
	} else {
		name = name1
	}
	src = w_CacheLumpNameBytes(name)
	destPos := 0
	y = 0
	for {
		if y >= SCREENHEIGHT-SBARHEIGHT {
			break
		}
		x = 0
		for {
			if x >= SCREENWIDTH/64 {
				break
			}
			copy(background_buffer[destPos:destPos+64], src[uintptr(y&63<<6):])
			destPos += 64
			goto _2
		_2:
			;
			x++
		}
		if SCREENWIDTH&63 != 0 {
			length := SCREENWIDTH & 63
			copy(background_buffer[destPos:destPos+length], src[uintptr(y&63<<6):])
			destPos += length
		}
		goto _1
	_1:
		;
		y++
	}
	// Draw screen and bezel; this is done to a separate screen buffer.
	v_UseBuffer(background_buffer)
	patch = w_CacheLumpNameT("brdr_t")
	x = 0
	for {
		if x >= scaledviewwidth {
			break
		}
		v_DrawPatch(viewwindowx+x, viewwindowy-8, patch)
		goto _3
	_3:
		;
		x += 8
	}
	patch = w_CacheLumpNameT("brdr_b")
	x = 0
	for {
		if x >= scaledviewwidth {
			break
		}
		v_DrawPatch(viewwindowx+x, viewwindowy+viewheight, patch)
		goto _4
	_4:
		;
		x += 8
	}
	patch = w_CacheLumpNameT("brdr_l")
	y = 0
	for {
		if y >= viewheight {
			break
		}
		v_DrawPatch(viewwindowx-8, viewwindowy+y, patch)
		goto _5
	_5:
		;
		y += 8
	}
	patch = w_CacheLumpNameT("brdr_r")
	y = 0
	for {
		if y >= viewheight {
			break
		}
		v_DrawPatch(viewwindowx+scaledviewwidth, viewwindowy+y, patch)
		goto _6
	_6:
		;
		y += 8
	}
	// Draw beveled edge.
	v_DrawPatch(viewwindowx-8, viewwindowy-8, w_CacheLumpNameT("brdr_tl"))
	v_DrawPatch(viewwindowx+scaledviewwidth, viewwindowy-8, w_CacheLumpNameT("brdr_tr"))
	v_DrawPatch(viewwindowx-8, viewwindowy+viewheight, w_CacheLumpNameT("brdr_bl"))
	v_DrawPatch(viewwindowx+scaledviewwidth, viewwindowy+viewheight, w_CacheLumpNameT("brdr_br"))
	v_RestoreBuffer()
}

// C documentation
//
//	//
//	// Copy a screen buffer.
//	//
func r_VideoErase(ofs uint32, count int32) {
	// LFB copy.
	// This might not be a good idea if memcpy
	//  is not optiomal, e.g. byte by byte on
	//  a 32bit CPU, as GNU GCC/Linux libc did
	//  at one point.
	if background_buffer != nil {
		copy(I_VideoBuffer[ofs:], background_buffer[ofs:ofs+uint32(count)])
	}
}

// C documentation
//
//	//
//	// R_DrawViewBorder
//	// Draws the border around the view
//	//  for different size windows?
//	//
func r_DrawViewBorder() {
	var i, ofs, side, top int32
	if scaledviewwidth == SCREENWIDTH {
		return
	}
	top = (SCREENHEIGHT - SBARHEIGHT - viewheight) / 2
	side = (SCREENWIDTH - scaledviewwidth) / 2
	// copy top and one line of left side
	r_VideoErase(0, top*SCREENWIDTH+side)
	// copy one line of right side and bottom
	ofs = (viewheight+top)*SCREENWIDTH - side
	r_VideoErase(uint32(ofs), top*SCREENWIDTH+side)
	// copy sides using wraparound
	ofs = top*SCREENWIDTH + SCREENWIDTH - side
	side <<= 1
	i = 1
	for {
		if i >= viewheight {
			break
		}
		r_VideoErase(uint32(ofs), side)
		ofs += SCREENWIDTH
		goto _1
	_1:
		;
		i++
	}
	// ?
	v_MarkRect(0, 0, SCREENWIDTH, SCREENHEIGHT-SBARHEIGHT)
}

const ANG18011 = 2147483648
const ANG2705 = 3221225472
const ANG909 = 1073741824
const DISTMAP = 2
const FIELDOFVIEW = 2048
const NF_SUBSECTOR5 = 32768

func init() {
	validcount = 1
}

// C documentation
//
//	//
//	// R_PointOnSide
//	// Traverse BSP (sub) tree,
//	//  check point against partition plane.
//	// Returns side 0 (front) or 1 (back).
//	//
func r_PointOnSide(x fixed_t, y fixed_t, node *node_t) int32 {
	var dx, dy, left, right fixed_t
	if node.Fdx == 0 {
		if x <= node.Fx {
			return boolint32(node.Fdy > 0)
		}
		return boolint32(node.Fdy < 0)
	}
	if node.Fdy == 0 {
		if y <= node.Fy {
			return boolint32(node.Fdx < 0)
		}
		return boolint32(node.Fdx > 0)
	}
	dx = x - node.Fx
	dy = y - node.Fy
	// Try to quickly decide by looking at sign bits.
	if uint32(node.Fdy^node.Fdx^dx^dy)&0x80000000 != 0 {
		if uint32(node.Fdy^dx)&0x80000000 != 0 {
			// (left is negative)
			return 1
		}
		return 0
	}
	left = fixedMul(node.Fdy>>FRACBITS, dx)
	right = fixedMul(dy, node.Fdx>>FRACBITS)
	if right < left {
		// front side
		return 0
	}
	// back side
	return 1
}

func r_PointOnSegSide(x fixed_t, y fixed_t, line *seg_t) int32 {
	var dx, dy, ldx, ldy, left, lx, ly, right fixed_t
	lx = line.Fv1.Fx
	ly = line.Fv1.Fy
	ldx = line.Fv2.Fx - lx
	ldy = line.Fv2.Fy - ly
	if ldx == 0 {
		if x <= lx {
			return boolint32(ldy > 0)
		}
		return boolint32(ldy < 0)
	}
	if ldy == 0 {
		if y <= ly {
			return boolint32(ldx < 0)
		}
		return boolint32(ldx > 0)
	}
	dx = x - lx
	dy = y - ly
	// Try to quickly decide by looking at sign bits.
	if uint32(ldy^ldx^dx^dy)&0x80000000 != 0 {
		if uint32(ldy^dx)&0x80000000 != 0 {
			// (left is negative)
			return 1
		}
		return 0
	}
	left = fixedMul(ldy>>FRACBITS, dx)
	right = fixedMul(dy, ldx>>FRACBITS)
	if right < left {
		// front side
		return 0
	}
	// back side
	return 1
}

//
// R_PointToAngle
// To get a global angle from cartesian coordinates,
//  the coordinates are flipped until they are in
//  the first octant of the coordinate system, then
//  the y (<=x) is scaled and divided by x to get a
//  tangent (slope) value which is looked up in the
//  tantoangle[] table.

//

func r_PointToAngle(x fixed_t, y fixed_t) angle_t {
	x -= viewx
	y -= viewy
	if x == 0 && y == 0 {
		return 0
	}
	if x >= 0 {
		// x >=0
		if y >= 0 {
			// y>= 0
			if x > y {
				// octant 0
				return tantoangle[slopeDiv(uint32(y), uint32(x))]
			} else {
				// octant 1
				return uint32(ANG909-1) - tantoangle[slopeDiv(uint32(x), uint32(y))]
			}
		} else {
			// y<0
			y = -y
			if x > y {
				// octant 8
				return -tantoangle[slopeDiv(uint32(y), uint32(x))]
			} else {
				// octant 7
				return uint32(ANG2705) + tantoangle[slopeDiv(uint32(x), uint32(y))]
			}
		}
	} else {
		// x<0
		x = -x
		if y >= 0 {
			// y>= 0
			if x > y {
				// octant 3
				return uint32(ANG18011) - 1 - tantoangle[slopeDiv(uint32(y), uint32(x))]
			} else {
				// octant 2
				return uint32(ANG909) + tantoangle[slopeDiv(uint32(x), uint32(y))]
			}
		} else {
			// y<0
			y = -y
			if x > y {
				// octant 4
				return uint32(ANG18011) + tantoangle[slopeDiv(uint32(y), uint32(x))]
			} else {
				// octant 5
				return uint32(ANG2705) - 1 - tantoangle[slopeDiv(uint32(x), uint32(y))]
			}
		}
	}
	return 0
}

func r_PointToAngle2(x1 fixed_t, y1 fixed_t, x2 fixed_t, y2 fixed_t) angle_t {
	viewx = x1
	viewy = y1
	return r_PointToAngle(x2, y2)
}

func r_PointToDist(x fixed_t, y fixed_t) fixed_t {
	var angle int32
	var dist, dx, dy, frac, temp fixed_t
	dx = xabs(x - viewx)
	dy = xabs(y - viewy)
	if dy > dx {
		temp = dx
		dx = dy
		dy = temp
	}
	// Fix crashes in udm1.wad
	if dx != 0 {
		frac = fixedDiv(dy, dx)
	} else {
		frac = 0
	}
	angle = int32((tantoangle[frac>>(FRACBITS-SLOPEBITS)] + uint32(ANG909)) >> ANGLETOFINESHIFT)
	// use as cosine
	dist = fixedDiv(dx, finesine[angle])
	return dist
}

// C documentation
//
//	//
//	// R_InitPointToAngle
//	//
func r_InitPointToAngle() {
	// UNUSED - now getting from tables.c
}

// C documentation
//
//	//
//	// R_ScaleFromGlobalAngle
//	// Returns the texture mapping scale
//	//  for the current line (horizontal span)
//	//  at the given angle.
//	// rw_distance must be calculated first.
//	//
func r_ScaleFromGlobalAngle(visangle angle_t) fixed_t {
	var anglea, angleb angle_t
	var den, sinea, sineb int32
	var num, scale fixed_t
	// UNUSED
	anglea = uint32(ANG909) + (visangle - viewangle)
	angleb = uint32(ANG909) + (visangle - rw_normalangle)
	// both sines are allways positive
	sinea = finesine[anglea>>ANGLETOFINESHIFT]
	sineb = finesine[angleb>>ANGLETOFINESHIFT]
	num = fixedMul(projection, sineb) << detailshift
	den = fixedMul(rw_distance, sinea)
	if den > num>>int32(16) {
		scale = fixedDiv(num, den)
		if scale > 64*(1<<FRACBITS) {
			scale = 64 * (1 << FRACBITS)
		} else {
			if scale < 256 {
				scale = 256
			}
		}
	} else {
		scale = 64 * (1 << FRACBITS)
	}
	return scale
}

// C documentation
//
//	//
//	// R_InitTables
//	//
func r_InitTables() {
	// UNUSED: now getting from tables.c
}

// C documentation
//
//	//
//	// R_InitTextureMapping
//	//
func r_InitTextureMapping() {
	var focallength fixed_t
	var t int32
	// Use tangent table to generate viewangletox:
	//  viewangletox will give the next greatest x
	//  after the view angle.
	//
	// Calc focallength
	//  so FIELDOFVIEW angles covers SCREENWIDTH.
	focallength = fixedDiv(centerxfrac, finetangent[FINEANGLES/4+FIELDOFVIEW/2])
	for i := 0; i < FINEANGLES/2; i++ {
		if finetangent[i] > 1<<FRACBITS*2 {
			t = -1
		} else {
			if finetangent[i] < -(1<<FRACBITS)*2 {
				t = viewwidth + 1
			} else {
				t = fixedMul(finetangent[i], focallength)
				t = (centerxfrac - t + 1<<FRACBITS - 1) >> FRACBITS
				if t < -1 {
					t = -1
				} else {
					if t > viewwidth+1 {
						t = viewwidth + 1
					}
				}
			}
		}
		viewangletox[i] = t
	}
	// Scan viewangletox[] to generate xtoviewangle[]:
	//  xtoviewangle will give the smallest view angle
	//  that maps to x.
	for x := int32(0); x < viewwidth; x++ {
		i := int32(0)
		for viewangletox[i] > x {
			i++
		}
		xtoviewangle[x] = uint32(i<<ANGLETOFINESHIFT - int32(ANG909))
	}
	// Take out the fencepost cases from viewangletox.
	for i := range FINEANGLES / 2 {
		t = fixedMul(finetangent[i], focallength)
		t = centerx - t
		if viewangletox[i] == -1 {
			viewangletox[i] = 0
		} else {
			if viewangletox[i] == viewwidth+1 {
				viewangletox[i] = viewwidth
			}
		}
	}
	clipangle = xtoviewangle[0]
}

//
// R_InitLightTables
// Only inits the zlight table,
//  because the scalelight table changes with view size.
//

func r_InitLightTables() {
	var level, scale, startmap int32
	// Calculate the light levels to use
	//  for each level / distance combination.
	for i := range int32(LIGHTLEVELS) {
		startmap = (LIGHTLEVELS - 1 - i) * 2 * NUMCOLORMAPS / LIGHTLEVELS
		for j := range int32(MAXLIGHTZ) {
			scale = fixedDiv(SCREENWIDTH/2*(1<<FRACBITS), (j+1)<<LIGHTZSHIFT)
			scale >>= LIGHTSCALESHIFT
			level = startmap - scale/DISTMAP
			if level < 0 {
				level = 0
			}
			if level >= NUMCOLORMAPS {
				level = NUMCOLORMAPS - 1
			}
			zlight[i][j] = colormaps[level*int32(256):]
		}
	}
}

func r_SetViewSize(blocks int32, detail int32) {
	setsizeneeded = 1
	setblocks = blocks
	setdetail = detail
}

// C documentation
//
//	//
//	// R_ExecuteSetViewSize
//	//
func r_ExecuteSetViewSize() {
	var cosadj, dy fixed_t
	var level, startmap int32
	setsizeneeded = 0
	if setblocks == 11 {
		scaledviewwidth = SCREENWIDTH
		viewheight = SCREENHEIGHT
	} else {
		scaledviewwidth = setblocks * 32
		viewheight = setblocks * 168 / 10 & ^7
	}
	detailshift = setdetail
	viewwidth = scaledviewwidth >> detailshift
	centery = viewheight / 2
	centerx = viewwidth / 2
	centerxfrac = centerx << FRACBITS
	centeryfrac = centery << FRACBITS
	projection = centerxfrac
	if detailshift == 0 {
		basecolfunc = r_DrawColumn
		colfunc = r_DrawColumn
		fuzzcolfunc = r_DrawFuzzColumn
		transcolfunc = r_DrawTranslatedColumn
		spanfunc = r_DrawSpan
	} else {
		basecolfunc = r_DrawColumnLow
		colfunc = r_DrawColumnLow
		fuzzcolfunc = r_DrawFuzzColumnLow
		transcolfunc = r_DrawTranslatedColumnLow
		spanfunc = r_DrawSpanLow
	}
	r_InitBuffer(scaledviewwidth, viewheight)
	r_InitTextureMapping()
	// psprite scales
	pspritescale = 1 << FRACBITS * viewwidth / SCREENWIDTH
	pspriteiscale = 1 << FRACBITS * SCREENWIDTH / viewwidth
	// thing clipping
	for i := range viewwidth {
		screenheightarray[i] = int16(viewheight)
	}
	// planes
	for i := range viewheight {
		dy = (i-viewheight/2)<<FRACBITS + 1<<FRACBITS/2
		dy = xabs(dy)
		yslope[i] = fixedDiv(viewwidth<<detailshift/2*(1<<FRACBITS), dy)
	}
	for i := range viewwidth {
		cosadj = xabs(finecosine[xtoviewangle[i]>>ANGLETOFINESHIFT])
		distscale[i] = fixedDiv(1<<FRACBITS, cosadj)
	}
	// Calculate the light levels to use
	//  for each level / scale combination.
	for i := range int32(LIGHTLEVELS) {
		startmap = (LIGHTLEVELS - 1 - i) * 2 * NUMCOLORMAPS / LIGHTLEVELS
		for j := range int32(MAXLIGHTSCALE) {
			level = startmap - j*SCREENWIDTH/(viewwidth<<detailshift)/DISTMAP
			if level < 0 {
				level = 0
			}
			if level >= NUMCOLORMAPS {
				level = NUMCOLORMAPS - 1
			}
			scalelight[i][j] = colormaps[level*256:]
		}
	}
}

//
// R_Init
//

func r_Init() {
	r_InitData()
	fprintf_ccgo(os.Stdout, ".")
	r_InitPointToAngle()
	fprintf_ccgo(os.Stdout, ".")
	r_InitTables()
	// viewwidth / viewheight / detailLevel are set by the defaults
	fprintf_ccgo(os.Stdout, ".")
	r_SetViewSize(screenblocks, detailLevel)
	r_InitPlanes()
	fprintf_ccgo(os.Stdout, ".")
	r_InitLightTables()
	fprintf_ccgo(os.Stdout, ".")
	r_InitSkyMap()
	r_InitTranslationTables()
	fprintf_ccgo(os.Stdout, ".")
}

// C documentation
//
//	//
//	// R_PointInSubsector
//	//
func r_PointInSubsector(x fixed_t, y fixed_t) *subsector_t {
	var nodenum, side int32
	// single subsector is a special case
	if numnodes == 0 {
		return &subsectors[0]
	}
	nodenum = numnodes - 1
	for nodenum&NF_SUBSECTOR5 == 0 {
		node := &nodes[nodenum]
		side = r_PointOnSide(x, y, node)
		nodenum = int32(node.Fchildren[side])
	}
	return &subsectors[nodenum & ^NF_SUBSECTOR5]
}

// C documentation
//
//	//
//	// R_SetupFrame
//	//
func r_SetupFrame(player *player_t) {
	viewplayer = player
	viewx = player.Fmo.Fx
	viewy = player.Fmo.Fy
	viewangle = player.Fmo.Fangle + viewangleoffset
	extralight = player.Fextralight
	viewz = player.Fviewz
	viewsin = finesine[viewangle>>ANGLETOFINESHIFT]
	viewcos = finecosine[viewangle>>ANGLETOFINESHIFT]
	if player.Ffixedcolormap != 0 {
		fixedcolormap = colormaps[player.Ffixedcolormap*int32(256):]
		walllights = scalelightfixed
		for i := range MAXLIGHTSCALE {
			scalelightfixed[i] = fixedcolormap
		}
	} else {
		fixedcolormap = nil
	}
	validcount++
}

// C documentation
//
//	//
//	// R_RenderView
//	//
func r_RenderPlayerView(player *player_t) {
	r_SetupFrame(player)
	// Clear buffers.
	r_ClearClipSegs()
	r_ClearDrawSegs()
	r_ClearPlanes()
	r_ClearSprites()
	// check for new console commands.
	netUpdate()
	// The head node is the last node output.
	r_RenderBSPNode(numnodes - 1)
	// Check for new console commands.
	netUpdate()
	r_DrawPlanes()
	// Check for new console commands.
	netUpdate()
	r_DrawMasked()
	// Check for new console commands.
	netUpdate()
}

// C documentation
//
//	//
//	// R_InitPlanes
//	// Only at game startup.
//	//
func r_InitPlanes() {
	// Doh!
}

// C documentation
//
//	//
//	// R_MapPlane
//	//
//	// Uses global vars:
//	//  planeheight
//	//  ds_source
//	//  basexscale
//	//  baseyscale
//	//  viewx
//	//  viewy
//	//
//	// BASIC PRIMITIVE
//	//
func r_MapPlane(y int32, x1 int32, x2 int32) {
	var angle angle_t
	var distance, length, v1, v2, v3 fixed_t
	var index uint32
	if x2 < x1 || x1 < 0 || x2 >= viewwidth || y > viewheight {
		i_Error("r_MapPlane: %d, %d at %d", x1, x2, y)
	}
	if planeheight != cachedheight[y] {
		cachedheight[y] = planeheight
		v1 = fixedMul(planeheight, yslope[y])
		cacheddistance[y] = v1
		distance = v1
		v2 = fixedMul(distance, basexscale)
		cachedxstep[y] = v2
		ds_xstep = v2
		v3 = fixedMul(distance, baseyscale)
		cachedystep[y] = v3
		ds_ystep = v3
	} else {
		distance = cacheddistance[y]
		ds_xstep = cachedxstep[y]
		ds_ystep = cachedystep[y]
	}
	length = fixedMul(distance, distscale[x1])
	angle = (viewangle + xtoviewangle[x1]) >> ANGLETOFINESHIFT
	ds_xfrac = viewx + fixedMul(finecosine[angle], length)
	ds_yfrac = -viewy - fixedMul(finesine[angle], length)
	if fixedcolormap != nil {
		ds_colormap = fixedcolormap
	} else {
		index = uint32(distance >> LIGHTZSHIFT)
		if index >= MAXLIGHTZ {
			index = uint32(MAXLIGHTZ - 1)
		}
		ds_colormap = planezlight[index]
	}
	ds_y = y
	ds_x1 = x1
	ds_x2 = x2
	// high or low detail
	spanfunc()
}

// C documentation
//
//	//
//	// R_ClearPlanes
//	// At begining of frame.
//	//
func r_ClearPlanes() {
	var angle angle_t
	// opening / clipping determination
	for i := range viewwidth {
		floorclip[i] = int16(viewheight)
		ceilingclip[i] = int16(-1)
	}
	lastvisplane_index = 0
	lastopening = uintptr(unsafe.Pointer(&openings))
	// texture calculation
	clear(cachedheight[:])
	// left to right mapping
	angle = (viewangle - uint32(ANG909)) >> ANGLETOFINESHIFT
	// scale will be unit scale at SCREENWIDTH/2 distance
	basexscale = fixedDiv(finecosine[angle], centerxfrac)
	baseyscale = -fixedDiv(finesine[angle], centerxfrac)
}

// C documentation
//
//	//
//	// R_FindPlane
//	//
func r_FindPlane(height fixed_t, picnum int32, lightlevel int32) *visplane_t {
	if picnum == skyflatnum {
		height = 0 // all skys map together
		lightlevel = 0
	}
	for i := 0; i < lastvisplane_index; i++ {
		check := &visplanes[i]
		if height == check.Fheight && picnum == check.Fpicnum && lightlevel == check.Flightlevel {
			return check
		}
	}
	if lastvisplane_index >= len(visplanes)-1 {
		i_Error("r_FindPlane: no more visplanes")
	}
	check := &visplanes[lastvisplane_index]
	check.Fheight = height
	check.Fpicnum = picnum
	check.Flightlevel = lightlevel
	check.Fminx = SCREENWIDTH
	check.Fmaxx = -1
	for i := 0; i < 320; i++ {
		check.Ftop[i] = 0xff
	}
	lastvisplane_index++
	return check
}

// C documentation
//
//	//
//	// R_CheckPlane
//	//
func r_CheckPlane(pl *visplane_t, start int32, stop int32) *visplane_t {
	var intrh, intrl, unionh, unionl, x int32
	if start < pl.Fminx {
		intrl = pl.Fminx
		unionl = start
	} else {
		unionl = pl.Fminx
		intrl = start
	}
	if stop > pl.Fmaxx {
		intrh = pl.Fmaxx
		unionh = stop
	} else {
		unionh = pl.Fmaxx
		intrh = stop
	}
	x = intrl
	for {
		if !(x <= intrh) {
			break
		}
		if pl.Ftop[x] != 0xff {
			break
		}
		goto _1
	_1:
		;
		x++
	}
	if x > intrh {
		pl.Fminx = unionl
		pl.Fmaxx = unionh
		// use the same one
		return pl
	}
	// make a new visplane
	newPl := &visplanes[lastvisplane_index]
	newPl.Fheight = pl.Fheight
	newPl.Fpicnum = pl.Fpicnum
	newPl.Flightlevel = pl.Flightlevel
	newPl.Fminx = start
	newPl.Fmaxx = stop
	for i := 0; i < 320; i++ {
		newPl.Ftop[i] = 0xff
	}
	lastvisplane_index++
	return newPl
}

// C documentation
//
//	//
//	// R_DrawPlanes
//	// At the end of each frame.
//	//
func r_DrawPlanes() {
	var angle, b1, b2, light, lumpnum, stop, t1, t2, x int32
	if ds_index >= len(drawsegs) {
		i_Error("r_DrawPlanes: drawsegs overflow (%d)", ds_index)
	}
	if lastvisplane_index >= len(visplanes)-1 {
		i_Error("r_DrawPlanes: visplane overflow (%d)", lastvisplane_index)
	}
	if (int64(lastopening)-int64(uintptr(unsafe.Pointer(&openings))))/2 > int64(SCREENWIDTH*64) {
		i_Error("r_DrawPlanes: opening overflow (%d)", (int64(lastopening)-int64(uintptr(unsafe.Pointer(&openings))))/2)
	}
	for i := 0; i < lastvisplane_index; i++ {
		pl := &visplanes[i]
		if pl.Fminx > pl.Fmaxx {
			continue
		}
		// sky flat
		if pl.Fpicnum == skyflatnum {
			dc_iscale = pspriteiscale >> detailshift
			// Sky is allways drawn full bright,
			//  i.e. colormaps[0] is used.
			// Because of this hack, sky is not affected
			//  by INVUL inverse mapping.
			dc_colormap = colormaps
			dc_texturemid = skytexturemid
			x = pl.Fminx
			for {
				if !(x <= pl.Fmaxx) {
					break
				}
				dc_yl = int32(pl.Ftop[x])
				dc_yh = int32(pl.Fbottom[x])
				if dc_yl <= dc_yh {
					angle = int32((viewangle + xtoviewangle[x]) >> ANGLETOSKYSHIFT)
					dc_x = x
					dc_source = r_GetColumn(skytexture, angle)
					colfunc()
				}
				goto _2
			_2:
				;
				x++
			}
			continue
		}
		// regular flat
		lumpnum = firstflat + flattranslation[pl.Fpicnum]
		ds_source = w_CacheLumpNumBytes(lumpnum)
		planeheight = xabs(pl.Fheight - viewz)
		light = pl.Flightlevel>>LIGHTSEGSHIFT + extralight
		if light >= LIGHTLEVELS {
			light = LIGHTLEVELS - 1
		}
		if light < 0 {
			light = 0
		}
		planezlight = zlight[light][:]
		if int(pl.Fmaxx+1) < len(pl.Ftop) {
			pl.Ftop[pl.Fmaxx+1] = 0xff
		}
		if pl.Fminx-1 >= 0 {
			pl.Ftop[pl.Fminx-1] = 0xff
		}
		stop = pl.Fmaxx + 1
		x = pl.Fminx
		for {
			if !(x <= stop) {
				break
			}
			if x-1 >= 0 {
				t1 = int32(pl.Ftop[x-1])
				b1 = int32(pl.Fbottom[x-1])
			} else {
				t1 = 0xff
				b1 = -1
			}
			if x < int32(len(pl.Ftop)) {
				t2 = int32(pl.Ftop[x])
				b2 = int32(pl.Fbottom[x])
			} else {
				t2 = 0xff
				b2 = -1
			}
			for t1 < t2 && t1 <= b1 {
				r_MapPlane(t1, spanstart[t1], x-1)
				t1++
			}
			for b1 > b2 && b1 >= t1 {
				r_MapPlane(b1, spanstart[b1], x-1)
				b1--
			}
			for t2 < t1 && t2 <= b2 {
				spanstart[t2] = x
				t2++
			}
			for b2 > b1 && b2 >= t2 {
				spanstart[b2] = x
				b2--
			}
			goto _3
		_3:
			;
			x++
		}
		w_ReleaseLumpNum(lumpnum)
	}
}

const ANG18013 = 2147483648
const HEIGHTBITS = 12
const INT_MAX15 = 2147483647
const SHRT_MAX1 = 32767

// C documentation
//
//	//
//	// R_RenderMaskedSegRange
//	//
func r_RenderMaskedSegRange(ds *drawseg_t, x1 int32, x2 int32) {
	var col *column_t
	var index uint32
	var lightnum, texnum, v1, v2 int32
	// Calculate light table.
	// Use different light tables
	//   for horizontal / vertical / diagonal. Diagonal?
	// OPTIMIZE: get rid of LIGHTSEGSHIFT globally
	curline = ds.Fcurline
	frontsector = curline.Ffrontsector
	backsector = curline.Fbacksector
	texnum = texturetranslation[curline.Fsidedef.Fmidtexture]
	lightnum = int32(frontsector.Flightlevel)>>LIGHTSEGSHIFT + extralight
	if curline.Fv1.Fy == curline.Fv2.Fy {
		lightnum--
	} else {
		if curline.Fv1.Fx == curline.Fv2.Fx {
			lightnum++
		}
	}
	if lightnum < 0 {
		walllights = scalelight[0]
	} else {
		if lightnum >= LIGHTLEVELS {
			walllights = scalelight[LIGHTLEVELS-1]
		} else {
			walllights = scalelight[lightnum]
		}
	}
	maskedtexturecol = ds.Fmaskedtexturecol
	rw_scalestep = ds.Fscalestep
	spryscale = ds.Fscale1 + (x1-ds.Fx1)*rw_scalestep
	mfloorclip = ds.Fsprbottomclip
	mceilingclip = ds.Fsprtopclip
	// find positioning
	if int32(curline.Flinedef.Fflags)&ml_DONTPEGBOTTOM != 0 {
		if frontsector.Ffloorheight > backsector.Ffloorheight {
			v1 = frontsector.Ffloorheight
		} else {
			v1 = backsector.Ffloorheight
		}
		dc_texturemid = v1
		dc_texturemid = dc_texturemid + textureheight[texnum] - viewz
	} else {
		if frontsector.Fceilingheight < backsector.Fceilingheight {
			v2 = frontsector.Fceilingheight
		} else {
			v2 = backsector.Fceilingheight
		}
		dc_texturemid = v2
		dc_texturemid = dc_texturemid - viewz
	}
	dc_texturemid += curline.Fsidedef.Frowoffset
	if fixedcolormap != nil {
		dc_colormap = fixedcolormap
	}
	// draw the columns
	dc_x = x1
	for {
		if !(dc_x <= x2) {
			break
		}
		// calculate lighting
		if int32(*(*int16)(unsafe.Pointer(maskedtexturecol + uintptr(dc_x)*2))) != int32(SHRT_MAX1) {
			if fixedcolormap == nil {
				index = uint32(spryscale >> LIGHTSCALESHIFT)
				if index >= MAXLIGHTSCALE {
					index = uint32(MAXLIGHTSCALE - 1)
				}
				dc_colormap = walllights[index]
			}
			sprtopscreen = centeryfrac - fixedMul(dc_texturemid, spryscale)
			dc_iscale = int32(0xffffffff / uint32(spryscale))
			// draw the texture
			col = (*column_t)(unsafe.Pointer(r_GetColumn(texnum, int32(*(*int16)(unsafe.Pointer(maskedtexturecol + uintptr(dc_x)*2)))) - uintptr(3)))
			r_DrawMaskedColumn(col)
			*(*int16)(unsafe.Pointer(maskedtexturecol + uintptr(dc_x)*2)) = int16(SHRT_MAX1)
		}
		spryscale += rw_scalestep
		goto _3
	_3:
		;
		dc_x++
	}
}

//
// R_RenderSegLoop
// Draws zero, one, or two textures (and possibly a masked
//  texture) for walls.
// Can draw or mark the starting pixel of floor and ceiling
//  textures.
// CALLED: CORE LOOPING ROUTINE.
//

func r_RenderSegLoop() {
	var angle angle_t
	var bottom, mid, top, yh, yl int32
	var ceilingclip_pos, floorclip_pos int32
	var index uint32
	var texturecolumn fixed_t
	ceilingclip_pos = rw_x
	floorclip_pos = rw_x
	for {
		if rw_x >= rw_stopx {
			break
		}
		// mark floor / ceiling areas
		yl = (topfrac + 1<<HEIGHTBITS - 1) >> HEIGHTBITS
		// no space above wall?
		if yl < int32(ceilingclip[ceilingclip_pos])+1 {
			yl = int32(ceilingclip[ceilingclip_pos]) + 1
		}
		if markceiling != 0 {
			top = int32(ceilingclip[rw_x]) + 1
			bottom = yl - 1
			if bottom >= int32(floorclip[floorclip_pos]) {
				bottom = int32(floorclip[floorclip_pos]) - 1
			}
			if top <= bottom {
				ceilingplane.Ftop[rw_x] = uint8(top)
				ceilingplane.Fbottom[rw_x] = uint8(bottom)
			}
		}
		yh = bottomfrac >> HEIGHTBITS
		if yh >= int32(floorclip[floorclip_pos]) {
			yh = int32(floorclip[floorclip_pos]) - 1
		}
		if markfloor != 0 {
			top = yh + 1
			bottom = int32(floorclip[floorclip_pos]) - 1
			if top <= int32(ceilingclip[ceilingclip_pos]) {
				top = int32(ceilingclip[ceilingclip_pos]) + 1
			}
			if top <= bottom {
				floorplane.Ftop[rw_x] = uint8(top)
				floorplane.Fbottom[rw_x] = uint8(bottom)
			}
		}
		// texturecolumn and lighting are independent of wall tiers
		if segtextured != 0 {
			// calculate texture offset
			angle = (rw_centerangle + xtoviewangle[rw_x]) >> ANGLETOFINESHIFT
			if angle >= FINEANGLES/2 { // DSB-23
				angle = 0
			}
			texturecolumn = rw_offset - fixedMul(finetangent[angle], rw_distance)
			texturecolumn >>= FRACBITS
			// calculate lighting
			index = uint32(rw_scale >> LIGHTSCALESHIFT)
			if index >= MAXLIGHTSCALE {
				index = uint32(MAXLIGHTSCALE - 1)
			}
			dc_colormap = walllights[index]
			dc_x = rw_x
			dc_iscale = int32(0xffffffff / uint32(rw_scale))
		} else {
			// purely to shut up the compiler
			texturecolumn = 0
		}
		// draw the wall tiers
		if midtexture != 0 {
			// single sided line
			dc_yl = yl
			dc_yh = yh
			dc_texturemid = rw_midtexturemid
			dc_source = r_GetColumn(midtexture, texturecolumn)
			colfunc()
			ceilingclip[ceilingclip_pos] = int16(viewheight)
			floorclip[floorclip_pos] = int16(-1)
		} else {
			// two sided line
			if toptexture != 0 {
				// top wall
				mid = pixhigh >> HEIGHTBITS
				pixhigh += pixhighstep
				if mid >= int32(floorclip[floorclip_pos]) {
					mid = int32(floorclip[floorclip_pos]) - 1
				}
				if mid >= yl {
					dc_yl = yl
					dc_yh = mid
					dc_texturemid = rw_toptexturemid
					dc_source = r_GetColumn(toptexture, texturecolumn)
					colfunc()
					ceilingclip[ceilingclip_pos] = int16(mid)
				} else {
					ceilingclip[ceilingclip_pos] = int16(yl - 1)
				}
			} else {
				// no top wall
				if markceiling != 0 {
					ceilingclip[ceilingclip_pos] = int16(yl - 1)
				}
			}
			if bottomtexture != 0 {
				// bottom wall
				mid = (pixlow + 1<<HEIGHTBITS - 1) >> HEIGHTBITS
				pixlow += pixlowstep
				// no space above wall?
				if mid <= int32(ceilingclip[ceilingclip_pos]) {
					mid = int32(ceilingclip[ceilingclip_pos]) + 1
				}
				if mid <= yh {
					dc_yl = mid
					dc_yh = yh
					dc_texturemid = rw_bottomtexturemid
					dc_source = r_GetColumn(bottomtexture, texturecolumn)
					colfunc()
					floorclip[floorclip_pos] = int16(mid)
				} else {
					floorclip[floorclip_pos] = int16(yh + 1)
				}
			} else {
				// no bottom wall
				if markfloor != 0 {
					floorclip[floorclip_pos] = int16(yh + 1)
				}
			}
			if maskedtexture != 0 {
				// save texturecol
				//  for backdrawing of masked mid texture
				*(*int16)(unsafe.Pointer(maskedtexturecol + uintptr(rw_x)*2)) = int16(texturecolumn)
			}
		}
		rw_scale += rw_scalestep
		topfrac += topstep
		bottomfrac += bottomstep
		ceilingclip_pos++
		floorclip_pos++
		goto _1
	_1:
		;
		rw_x++
	}
}

// C documentation
//
//	//
//	// R_StoreWallRange
//	// A wall segment will be drawn
//	//  between start and stop pixels (inclusive).
//	//
func r_StoreWallRange(start int32, stop int32) {
	var distangle, offsetangle angle_t
	var hyp, sineval, vtop, v3, v4 fixed_t
	var lightnum, v2, v5, v6 int32
	var v10, v7, v8 boolean
	var v11, v9 uintptr
	// don't overflow and crash
	if ds_index >= len(drawsegs) {
		return
	}
	if start >= viewwidth || start > stop {
		i_Error("Bad R_RenderWallRange: %d to %d", start, stop)
	}
	sidedef = curline.Fsidedef
	linedef = curline.Flinedef
	// mark the segment as visible for auto map
	linedef.Fflags |= ml_MAPPED
	// calculate rw_distance for scale calculation
	rw_normalangle = curline.Fangle + uint32(ANG909)
	offsetangle = uint32(xabs(int32(rw_normalangle - uint32(rw_angle1))))
	if offsetangle > uint32(ANG909) {
		offsetangle = uint32(ANG909)
	}
	distangle = uint32(ANG909) - offsetangle
	hyp = r_PointToDist(curline.Fv1.Fx, curline.Fv1.Fy)
	sineval = finesine[distangle>>ANGLETOFINESHIFT]
	rw_distance = fixedMul(hyp, sineval)
	v2 = start
	rw_x = v2
	drawsegs[ds_index].Fx1 = v2
	drawsegs[ds_index].Fx2 = stop
	drawsegs[ds_index].Fcurline = curline
	rw_stopx = stop + 1
	// calculate scale at both ends and step
	v3 = r_ScaleFromGlobalAngle(viewangle + xtoviewangle[start])
	rw_scale = v3
	drawsegs[ds_index].Fscale1 = v3
	if stop > start {
		drawsegs[ds_index].Fscale2 = r_ScaleFromGlobalAngle(viewangle + xtoviewangle[stop])
		v4 = (drawsegs[ds_index].Fscale2 - rw_scale) / (stop - start)
		rw_scalestep = v4
		drawsegs[ds_index].Fscalestep = v4
	} else {
		// UNUSED: try to fix the stretched line bug
		drawsegs[ds_index].Fscale2 = drawsegs[ds_index].Fscale1
	}
	// calculate texture boundaries
	//  and decide if floor / ceiling marks are needed
	worldtop = frontsector.Fceilingheight - viewz
	worldbottom = frontsector.Ffloorheight - viewz
	v7 = 0
	maskedtexture = v7
	v6 = int32(v7)
	bottomtexture = v6
	v5 = v6
	toptexture = v5
	midtexture = v5
	drawsegs[ds_index].Fmaskedtexturecol = 0
	if backsector == nil {
		// single sided line
		midtexture = texturetranslation[sidedef.Fmidtexture]
		// a single sided line is terminal, so it must mark ends
		v8 = 1
		markceiling = v8
		markfloor = v8
		if int32(linedef.Fflags)&ml_DONTPEGBOTTOM != 0 {
			vtop = frontsector.Ffloorheight + textureheight[sidedef.Fmidtexture]
			// bottom of texture at bottom
			rw_midtexturemid = vtop - viewz
		} else {
			// top of texture at top
			rw_midtexturemid = worldtop
		}
		rw_midtexturemid += sidedef.Frowoffset
		drawsegs[ds_index].Fsilhouette = SIL_BOTH
		drawsegs[ds_index].Fsprtopclip = uintptr(unsafe.Pointer(&screenheightarray))
		drawsegs[ds_index].Fsprbottomclip = uintptr(unsafe.Pointer(&negonearray))
		drawsegs[ds_index].Fbsilheight = int32(INT_MAX15)
		drawsegs[ds_index].Ftsilheight = -1 - 0x7fffffff
	} else {
		// two sided line
		v9 = 0
		drawsegs[ds_index].Fsprbottomclip = v9
		drawsegs[ds_index].Fsprtopclip = v9
		drawsegs[ds_index].Fsilhouette = 0
		if frontsector.Ffloorheight > backsector.Ffloorheight {
			drawsegs[ds_index].Fsilhouette = SIL_BOTTOM
			drawsegs[ds_index].Fbsilheight = frontsector.Ffloorheight
		} else {
			if backsector.Ffloorheight > viewz {
				drawsegs[ds_index].Fsilhouette = SIL_BOTTOM
				drawsegs[ds_index].Fbsilheight = int32(INT_MAX15)
				// ds_p->sprbottomclip = negonearray;
			}
		}
		if frontsector.Fceilingheight < backsector.Fceilingheight {
			drawsegs[ds_index].Fsilhouette |= SIL_TOP
			drawsegs[ds_index].Ftsilheight = frontsector.Fceilingheight
		} else {
			if backsector.Fceilingheight < viewz {
				drawsegs[ds_index].Fsilhouette |= SIL_TOP
				drawsegs[ds_index].Ftsilheight = -1 - 0x7fffffff
				// ds_p->sprtopclip = screenheightarray;
			}
		}
		if backsector.Fceilingheight <= frontsector.Ffloorheight {
			drawsegs[ds_index].Fsprbottomclip = uintptr(unsafe.Pointer(&negonearray))
			drawsegs[ds_index].Fbsilheight = int32(INT_MAX15)
			drawsegs[ds_index].Fsilhouette |= SIL_BOTTOM
		}
		if backsector.Ffloorheight >= frontsector.Fceilingheight {
			drawsegs[ds_index].Fsprtopclip = uintptr(unsafe.Pointer(&screenheightarray))
			drawsegs[ds_index].Ftsilheight = -1 - 0x7fffffff
			drawsegs[ds_index].Fsilhouette |= SIL_TOP
		}
		worldhigh = backsector.Fceilingheight - viewz
		worldlow = backsector.Ffloorheight - viewz
		// hack to allow height changes in outdoor areas
		if int32(frontsector.Fceilingpic) == skyflatnum && int32(backsector.Fceilingpic) == skyflatnum {
			worldtop = worldhigh
		}
		if worldlow != worldbottom || int32(backsector.Ffloorpic) != int32(frontsector.Ffloorpic) || int32(backsector.Flightlevel) != int32(frontsector.Flightlevel) {
			markfloor = 1
		} else {
			// same plane on both sides
			markfloor = 0
		}
		if worldhigh != worldtop || int32(backsector.Fceilingpic) != int32(frontsector.Fceilingpic) || int32(backsector.Flightlevel) != int32(frontsector.Flightlevel) {
			markceiling = 1
		} else {
			// same plane on both sides
			markceiling = 0
		}
		if backsector.Fceilingheight <= frontsector.Ffloorheight || backsector.Ffloorheight >= frontsector.Fceilingheight {
			// closed door
			v10 = 1
			markfloor = v10
			markceiling = v10
		}
		if worldhigh < worldtop {
			// top texture
			toptexture = texturetranslation[sidedef.Ftoptexture]
			if int32(linedef.Fflags)&ml_DONTPEGTOP != 0 {
				// top of texture at top
				rw_toptexturemid = worldtop
			} else {
				vtop = backsector.Fceilingheight + textureheight[sidedef.Ftoptexture]
				// bottom of texture
				rw_toptexturemid = vtop - viewz
			}
		}
		if worldlow > worldbottom {
			// bottom texture
			bottomtexture = texturetranslation[sidedef.Fbottomtexture]
			if int32(linedef.Fflags)&ml_DONTPEGBOTTOM != 0 {
				// bottom of texture at bottom
				// top of texture at top
				rw_bottomtexturemid = worldtop
			} else { // top of texture at top
				rw_bottomtexturemid = worldlow
			}
		}
		rw_toptexturemid += sidedef.Frowoffset
		rw_bottomtexturemid += sidedef.Frowoffset
		// allocate space for masked texture tables
		if sidedef.Fmidtexture != 0 {
			// masked midtexture
			maskedtexture = 1
			v11 = lastopening - uintptr(rw_x)*2
			maskedtexturecol = v11
			drawsegs[ds_index].Fmaskedtexturecol = v11
			lastopening += uintptr(rw_stopx-rw_x) * 2
		}
	}
	// calculate rw_offset (only needed for textured lines)
	segtextured = uint32(midtexture|toptexture|bottomtexture) | maskedtexture
	if segtextured != 0 {
		offsetangle = rw_normalangle - uint32(rw_angle1)
		if offsetangle > uint32(ANG18013) {
			offsetangle = -offsetangle
		}
		if offsetangle > uint32(ANG909) {
			offsetangle = uint32(ANG909)
		}
		sineval = finesine[offsetangle>>ANGLETOFINESHIFT]
		rw_offset = fixedMul(hyp, sineval)
		if rw_normalangle-uint32(rw_angle1) < uint32(ANG18013) {
			rw_offset = -rw_offset
		}
		rw_offset += sidedef.Ftextureoffset + curline.Foffset
		rw_centerangle = uint32(ANG909) + viewangle - rw_normalangle
		// calculate light table
		//  use different light tables
		//  for horizontal / vertical / diagonal
		// OPTIMIZE: get rid of LIGHTSEGSHIFT globally
		if fixedcolormap == nil {
			lightnum = int32(frontsector.Flightlevel)>>LIGHTSEGSHIFT + extralight
			if curline.Fv1.Fy == curline.Fv2.Fy {
				lightnum--
			} else {
				if curline.Fv1.Fx == curline.Fv2.Fx {
					lightnum++
				}
			}
			if lightnum < 0 {
				walllights = scalelight[0]
			} else {
				if lightnum >= LIGHTLEVELS {
					walllights = scalelight[LIGHTLEVELS-1]
				} else {
					walllights = scalelight[lightnum]
				}
			}
		}
	}
	// if a floor / ceiling plane is on the wrong side
	//  of the view plane, it is definitely invisible
	//  and doesn't need to be marked.
	if frontsector.Ffloorheight >= viewz {
		// above view plane
		markfloor = 0
	}
	if frontsector.Fceilingheight <= viewz && int32(frontsector.Fceilingpic) != skyflatnum {
		// below view plane
		markceiling = 0
	}
	// calculate incremental stepping values for texture edges
	worldtop >>= 4
	worldbottom >>= 4
	topstep = -fixedMul(rw_scalestep, worldtop)
	topfrac = centeryfrac>>4 - fixedMul(worldtop, rw_scale)
	bottomstep = -fixedMul(rw_scalestep, worldbottom)
	bottomfrac = centeryfrac>>4 - fixedMul(worldbottom, rw_scale)
	if backsector != nil {
		worldhigh >>= 4
		worldlow >>= 4
		if worldhigh < worldtop {
			pixhigh = centeryfrac>>4 - fixedMul(worldhigh, rw_scale)
			pixhighstep = -fixedMul(rw_scalestep, worldhigh)
		}
		if worldlow > worldbottom {
			pixlow = centeryfrac>>4 - fixedMul(worldlow, rw_scale)
			pixlowstep = -fixedMul(rw_scalestep, worldlow)
		}
	}
	// render it
	if markceiling != 0 {
		ceilingplane = r_CheckPlane(ceilingplane, rw_x, rw_stopx-1)
	}
	if markfloor != 0 {
		floorplane = r_CheckPlane(floorplane, rw_x, rw_stopx-1)
	}
	r_RenderSegLoop()
	// save sprite clipping info
	if (drawsegs[ds_index].Fsilhouette&SIL_TOP != 0 || maskedtexture != 0) && drawsegs[ds_index].Fsprtopclip == 0 {
		xmemcpy(lastopening, uintptr(unsafe.Pointer(&ceilingclip))+uintptr(start)*2, uint64(2*(rw_stopx-start)))
		drawsegs[ds_index].Fsprtopclip = lastopening - uintptr(start)*2
		lastopening += uintptr(rw_stopx-start) * 2
	}
	if (drawsegs[ds_index].Fsilhouette&SIL_BOTTOM != 0 || maskedtexture != 0) && drawsegs[ds_index].Fsprbottomclip == 0 {
		xmemcpy(lastopening, uintptr(unsafe.Pointer(&floorclip))+uintptr(start)*2, uint64(2*(rw_stopx-start)))
		drawsegs[ds_index].Fsprbottomclip = lastopening - uintptr(start)*2
		lastopening += uintptr(rw_stopx-start) * 2
	}
	if maskedtexture != 0 && drawsegs[ds_index].Fsilhouette&SIL_TOP == 0 {
		drawsegs[ds_index].Fsilhouette |= SIL_TOP
		drawsegs[ds_index].Ftsilheight = -1 - 0x7fffffff
	}
	if maskedtexture != 0 && drawsegs[ds_index].Fsilhouette&SIL_BOTTOM == 0 {
		drawsegs[ds_index].Fsilhouette |= SIL_BOTTOM
		drawsegs[ds_index].Fbsilheight = int32(INT_MAX15)
	}
	ds_index++
}

// C documentation
//
//	//
//	// R_InitSkyMap
//	// Called whenever the view size changes.
//	//
func r_InitSkyMap() {
	// skyflatnum = r_FlatNumForName ( SKYFLATNAME );
	skytexturemid = 100 * (1 << FRACBITS)
}

const ANG455 = 536870912
const BASEYCENTER = 100
const FF_FRAMEMASK3 = 32767
const FF_FULLBRIGHT1 = 32768
const INT_MAX17 = 2147483647

// C documentation
//
//	//
//	// R_InstallSpriteLump
//	// Local function for r_InitSprites.
//	//
func r_InstallSpriteLump(spritename string, lump int32, frame uint32, rotation uint32, flipped boolean) {
	if frame >= 29 || rotation > 8 {
		i_Error("r_InstallSpriteLump: Bad frame characters in lump %d", lump)
	}
	if int32(frame) > maxframe {
		maxframe = int32(frame)
	}
	if rotation == 0 {
		// the lump should be used for all rotations
		if sprtemp[frame].Frotate == 0 {
			i_Error("r_InitSprites: Sprite %s frame %c has multip rot=0 lump", spritename, 'A'+frame)
		}
		if sprtemp[frame].Frotate == 1 {
			i_Error("r_InitSprites: Sprite %s frame %c has rotations and a rot=0 lump", spritename, 'A'+frame)
		}
		sprtemp[frame].Frotate = 0
		for r := range 8 {
			sprtemp[frame].Flump[r] = int16(lump - firstspritelump)
			sprtemp[frame].Fflip[r] = uint8(flipped)
		}
		return
	}
	// the lump is only used for one rotation
	if sprtemp[frame].Frotate == 0 {
		i_Error("r_InitSprites: Sprite %s frame %c has rotations and a rot=0 lump", spritename, 'A'+frame)
	}
	sprtemp[frame].Frotate = 1
	// make 0 based
	rotation--
	if int32(sprtemp[frame].Flump[rotation]) != -1 {
		i_Error("r_InitSprites: Sprite %s : %c : %c has two lumps mapped to it", spritename, 'A'+frame, '1'+rotation)
	}
	sprtemp[frame].Flump[rotation] = int16(lump - firstspritelump)
	sprtemp[frame].Fflip[rotation] = uint8(flipped)
}

// C documentation
//
//	//
//	// R_InitSpriteDefs
//	// Pass a null terminated list of sprite names
//	//  (4 chars exactly) to be used.
//	// Builds the sprite rotation matrixes to account
//	//  for horizontally flipped sprites.
//	// Will report an error if the lumps are inconsistant.
//	// Only called at startup.
//	//
//	// Sprite lump names are 4 characters for the actor,
//	//  a letter for the frame, and a number for the rotation.
//	// A sprite that is flippable will have an additional
//	//  letter/number appended.
//	// The rotation character can be 0 to signify no rotations.
//	//
func r_InitSpriteDefs(namelist []string) {
	var end, frame, l, patched, rotation, start int32
	// count the number of sprite names
	numsprites = int32(len(namelist))
	if numsprites == 0 {
		return
	}
	sprites = make([]spritedef_t, int(numsprites))
	start = firstspritelump - 1
	end = lastspritelump + 1
	// scan all the lump names for each of the names,
	//  noting the highest frame letter.
	// Just compare 4 characters as ints
	for i := range numsprites {

		spritename := namelist[i][:4]
		for i := range sprtemp {
			sprtemp[i].Frotate = 0xff
			for j := range sprtemp[i].Flump {
				sprtemp[i].Flump[j] = -1
			}
			for j := range sprtemp[i].Fflip {
				sprtemp[i].Fflip[j] = 0xff
			}
		}
		maxframe = -1
		// scan the lumps,
		//  filling in the frames for whatever is found
		l = start + 1
		for {
			if l >= end {
				break
			}

			if strings.EqualFold(lumpinfo[l].Name()[:4], spritename[:4]) {
				frame = int32(lumpinfo[l].Fname[4] - 'A')
				rotation = int32(lumpinfo[l].Fname[5] - '0')
				if modifiedgame != 0 {
					patched = w_GetNumForName(lumpinfo[l].Name())
				} else {
					patched = l
				}
				r_InstallSpriteLump(spritename, patched, uint32(frame), uint32(rotation), 0)
				if lumpinfo[l].Fname[6] != 0 {
					frame = int32(lumpinfo[l].Fname[6] - 'A')
					rotation = int32(lumpinfo[l].Fname[7] - '0')
					r_InstallSpriteLump(spritename, l, uint32(frame), uint32(rotation), 1)
				}
			}
			l++
		}
		// check the frames that were found for completeness
		if maxframe == -1 {
			sprites[i].Fnumframes = 0
			continue
		}
		maxframe++
		for frame := range maxframe {
			switch int32(sprtemp[frame].Frotate) {
			case -1:
				// no rotations were found for that frame at all
				i_Error("r_InitSprites: No patches found for %s frame %c", spritename, frame+'A')
				break
			case 0:
				break
			case 1:
				// must have all 8 frames
				for rotation := range 8 {
					if int32(sprtemp[frame].Flump[rotation]) == -1 {
						i_Error("r_InitSprites: Sprite %s frame %c is missing rotations", spritename, frame+'A')
					}
				}
			}
		}
		// allocate space for the frames present and copy sprtemp to it
		sprites[i].Fnumframes = maxframe
		sprites[i].Fspriteframes = make([]spriteframe_t, maxframe)
		copy(sprites[i].Fspriteframes, sprtemp[:maxframe])
	}
}

// C documentation
//
//	//
//	// R_InitSprites
//	// Called at program start.
//	//
func r_InitSprites(namelist []string) {
	for i := range SCREENWIDTH {
		negonearray[i] = int16(-1)
	}
	r_InitSpriteDefs(namelist)
}

// C documentation
//
//	//
//	// R_ClearSprites
//	// Called at frame start.
//	//
func r_ClearSprites() {
	vissprite_n = 0
}

func r_NewVisSprite() *vissprite_t {
	if vissprite_n == len(vissprites) {
		return &overflowsprite
	}
	vissprite_n++
	return &vissprites[vissprite_n-1]
}

func r_DrawMaskedColumn(column *column_t) {
	var basetexturemid fixed_t
	var bottomscreen, topscreen int32
	basetexturemid = dc_texturemid
	for {
		if int32(column.Ftopdelta) == 0xff {
			break
		}
		// calculate unclipped screen coordinates
		//  for post
		topscreen = sprtopscreen + spryscale*int32(column.Ftopdelta)
		bottomscreen = topscreen + spryscale*int32(column.Flength)
		dc_yl = (topscreen + 1<<FRACBITS - 1) >> FRACBITS
		dc_yh = (bottomscreen - 1) >> FRACBITS
		if dc_yh >= int32(*(*int16)(unsafe.Pointer(mfloorclip + uintptr(dc_x)*2))) {
			dc_yh = int32(*(*int16)(unsafe.Pointer(mfloorclip + uintptr(dc_x)*2))) - 1
		}
		if dc_yl <= int32(*(*int16)(unsafe.Pointer(mceilingclip + uintptr(dc_x)*2))) {
			dc_yl = int32(*(*int16)(unsafe.Pointer(mceilingclip + uintptr(dc_x)*2))) + 1
		}
		if dc_yl <= dc_yh {
			dc_source = (uintptr)(unsafe.Pointer(column)) + uintptr(3)
			dc_texturemid = basetexturemid - int32(column.Ftopdelta)<<FRACBITS
			// dc_source = (byte *)column + 3 - column->topdelta;
			// Drawn by either R_DrawColumn
			//  or (SHADOW) r_DrawFuzzColumn.
			colfunc()
		}
		column = column.Next()
		goto _1
	_1:
	}
	dc_texturemid = basetexturemid
}

// C documentation
//
//	//
//	// R_DrawVisSprite
//	//  mfloorclip and mceilingclip should also be set.
//	//
func r_DrawVisSprite(vis *vissprite_t, x1 int32, x2 int32) {
	var patch *patch_t
	var frac fixed_t
	var texturecolumn int32
	patch = w_CacheLumpNumT(vis.Fpatch + firstspritelump)
	dc_colormap = vis.Fcolormap
	if dc_colormap == nil {
		// NULL colormap = shadow draw
		colfunc = fuzzcolfunc
	} else {
		if vis.Fmobjflags&mf_TRANSLATION != 0 {
			colfunc = transcolfunc
			dc_translation = translationtables[-256+(vis.Fmobjflags&mf_TRANSLATION>>(mf_TRANSSHIFT-8)):]
		}
	}
	dc_iscale = xabs(vis.Fxiscale) >> detailshift
	dc_texturemid = vis.Ftexturemid
	frac = vis.Fstartfrac
	spryscale = vis.Fscale
	sprtopscreen = centeryfrac - fixedMul(dc_texturemid, spryscale)
	dc_x = vis.Fx1
	for {
		if !(dc_x <= vis.Fx2) {
			break
		}
		texturecolumn = frac >> FRACBITS
		if texturecolumn < 0 || texturecolumn >= int32(patch.Fwidth) {
			i_Error("R_DrawSpriteRange: bad texturecolumn")
		}
		r_DrawMaskedColumn(patch.GetColumn(texturecolumn))
		goto _1
	_1:
		;
		dc_x++
		frac += vis.Fxiscale
	}
	colfunc = basecolfunc
}

// C documentation
//
//	//
//	// R_ProjectSprite
//	// Generates a vissprite for a thing
//	//  if it might be visible.
//	//
func r_ProjectSprite(thing *mobj_t) {
	var ang angle_t
	var flip boolean
	var gxt, gyt, iscale, tr_x, tr_y, tx, tz, xscale fixed_t
	var index, lump, x1, x2, v1, v2 int32
	var rot uint32
	var sprdef *spritedef_t
	var sprframe *spriteframe_t
	var vis *vissprite_t
	// transform the origin point
	tr_x = thing.Fx - viewx
	tr_y = thing.Fy - viewy
	gxt = fixedMul(tr_x, viewcos)
	gyt = -fixedMul(tr_y, viewsin)
	tz = gxt - gyt
	// thing is behind view plane?
	if tz < 1<<FRACBITS*4 {
		return
	}
	xscale = fixedDiv(projection, tz)
	gxt = -fixedMul(tr_x, viewsin)
	gyt = fixedMul(tr_y, viewcos)
	tx = -(gyt + gxt)
	// too far off the side?
	if xabs(tx) > tz<<2 {
		return
	}
	// decide which patch to use for sprite relative to player
	if uint32(thing.Fsprite) >= uint32(numsprites) {
		i_Error("r_ProjectSprite: invalid sprite number %d ", thing.Fsprite)
	}
	sprdef = &sprites[thing.Fsprite]
	if thing.Fframe&int32(FF_FRAMEMASK3) >= sprdef.Fnumframes {
		i_Error("r_ProjectSprite: invalid sprite frame %d : %d ", thing.Fsprite, thing.Fframe)
	}
	sprframe = &sprdef.Fspriteframes[thing.Fframe&int32(FF_FRAMEMASK3)]
	if sprframe.Frotate != 0 {
		// choose a different rotation based on player view
		ang = r_PointToAngle(thing.Fx, thing.Fy)
		rot = (ang - thing.Fangle + uint32(ANG455/2)*9) >> 29
		lump = int32(sprframe.Flump[rot])
		flip = uint32(sprframe.Fflip[rot])
	} else {
		// use single rotation for all views
		lump = int32(sprframe.Flump[0])
		flip = uint32(sprframe.Fflip[0])
	}
	// calculate edges of the shape
	tx -= spriteoffset[lump]
	x1 = (centerxfrac + fixedMul(tx, xscale)) >> FRACBITS
	// off the right side?
	if x1 > viewwidth {
		return
	}
	tx += spritewidth[lump]
	x2 = (centerxfrac+fixedMul(tx, xscale))>>FRACBITS - 1
	// off the left side
	if x2 < 0 {
		return
	}
	// store information in a vissprite
	vis = r_NewVisSprite()
	vis.Fmobjflags = thing.Fflags
	vis.Fscale = xscale << detailshift
	vis.Fgx = thing.Fx
	vis.Fgy = thing.Fy
	vis.Fgz = thing.Fz
	vis.Fgzt = thing.Fz + spritetopoffset[lump]
	vis.Ftexturemid = vis.Fgzt - viewz
	if x1 < 0 {
		v1 = 0
	} else {
		v1 = x1
	}
	vis.Fx1 = v1
	if x2 >= viewwidth {
		v2 = viewwidth - 1
	} else {
		v2 = x2
	}
	vis.Fx2 = v2
	iscale = fixedDiv(1<<FRACBITS, xscale)
	if flip != 0 {
		vis.Fstartfrac = spritewidth[lump] - 1
		vis.Fxiscale = -iscale
	} else {
		vis.Fstartfrac = 0
		vis.Fxiscale = iscale
	}
	if vis.Fx1 > x1 {

		vis.Fstartfrac += vis.Fxiscale * (vis.Fx1 - x1)
	}
	vis.Fpatch = lump
	// get light level
	if thing.Fflags&mf_SHADOW != 0 {
		// shadow draw
		vis.Fcolormap = nil
	} else {
		if fixedcolormap != nil {
			// fixed map
			vis.Fcolormap = fixedcolormap
		} else {
			if thing.Fframe&int32(FF_FULLBRIGHT1) != 0 {
				// full bright
				vis.Fcolormap = colormaps
			} else {
				// diminished light
				index = xscale >> (LIGHTSCALESHIFT - detailshift)
				if index >= MAXLIGHTSCALE {
					index = MAXLIGHTSCALE - 1
				}
				vis.Fcolormap = spritelights[index]
			}
		}
	}
}

// C documentation
//
//	//
//	// R_AddSprites
//	// During BSP traversal, this adds sprites by sector.
//	//
func r_AddSprites(sec *sector_t) {
	var lightnum int32
	var thing *mobj_t
	// BSP is traversed by subsector.
	// A sector might have been split into several
	//  subsectors during BSP building.
	// Thus we check whether its already added.
	if sec.Fvalidcount == validcount {
		return
	}
	// Well, now it will be done.
	sec.Fvalidcount = validcount
	lightnum = int32(sec.Flightlevel)>>LIGHTSEGSHIFT + extralight
	if lightnum < 0 {
		spritelights = scalelight[0]
	} else {
		if lightnum >= LIGHTLEVELS {
			spritelights = scalelight[LIGHTLEVELS-1]
		} else {
			spritelights = scalelight[lightnum]
		}
	}
	// Handle all things in sector.
	thing = sec.Fthinglist
	for {
		if thing == nil {
			break
		}
		r_ProjectSprite(thing)
		goto _1
	_1:
		;
		thing = thing.Fsnext
	}
}

// C documentation
//
//	//
//	// R_DrawPSprite
//	//
func r_DrawPSprite(psp *pspdef_t) {
	var flip boolean
	var lump, x1, x2, v1, v2 int32
	var sprdef *spritedef_t
	var sprframe *spriteframe_t
	var vis *vissprite_t
	var tx fixed_t
	// decide which patch to use
	if uint32(psp.Fstate.Fsprite) >= uint32(numsprites) {
		i_Error("r_ProjectSprite: invalid sprite number %d ", psp.Fstate.Fsprite)
	}
	sprdef = &sprites[psp.Fstate.Fsprite]
	if psp.Fstate.Fframe&int32(FF_FRAMEMASK3) >= sprdef.Fnumframes {
		i_Error("r_ProjectSprite: invalid sprite frame %d : %d ", psp.Fstate.Fsprite, psp.Fstate.Fframe)
	}
	sprframe = &sprdef.Fspriteframes[psp.Fstate.Fframe&int32(FF_FRAMEMASK3)]
	lump = int32(sprframe.Flump[0])
	flip = uint32(sprframe.Fflip[0])
	// calculate edges of the shape
	tx = psp.Fsx - 160*(1<<FRACBITS)
	tx -= spriteoffset[lump]
	x1 = (centerxfrac + fixedMul(tx, pspritescale)) >> FRACBITS
	// off the right side
	if x1 > viewwidth {
		return
	}
	tx += spritewidth[lump]
	x2 = (centerxfrac+fixedMul(tx, pspritescale))>>FRACBITS - 1
	// off the left side
	if x2 < 0 {
		return
	}
	// store information in a vissprite
	vis = &vissprite_t{}
	vis.Fmobjflags = 0
	vis.Ftexturemid = BASEYCENTER<<FRACBITS + 1<<FRACBITS/2 - (psp.Fsy - spritetopoffset[lump])
	if x1 < 0 {
		v1 = 0
	} else {
		v1 = x1
	}
	vis.Fx1 = v1
	if x2 >= viewwidth {
		v2 = viewwidth - 1
	} else {
		v2 = x2
	}
	vis.Fx2 = v2
	vis.Fscale = pspritescale << detailshift
	if flip != 0 {
		vis.Fxiscale = -pspriteiscale
		vis.Fstartfrac = spritewidth[lump] - 1
	} else {
		vis.Fxiscale = pspriteiscale
		vis.Fstartfrac = 0
	}
	if vis.Fx1 > x1 {
		vis.Fstartfrac += vis.Fxiscale * fixed_t(vis.Fx1-x1)
	}
	vis.Fpatch = lump
	if viewplayer.Fpowers[pw_invisibility] > 4*32 || viewplayer.Fpowers[pw_invisibility]&8 != 0 {
		// shadow draw
		vis.Fcolormap = nil
	} else {
		if fixedcolormap != nil {
			// fixed color
			vis.Fcolormap = fixedcolormap
		} else {
			if psp.Fstate.Fframe&int32(FF_FULLBRIGHT1) != 0 {
				// full bright
				vis.Fcolormap = colormaps
			} else {
				// local light
				vis.Fcolormap = spritelights[MAXLIGHTSCALE-1]
			}
		}
	}
	r_DrawVisSprite(vis, vis.Fx1, vis.Fx2)
}

// C documentation
//
//	//
//	// R_DrawPlayerSprites
//	//
func r_DrawPlayerSprites() {
	var lightnum int32
	// get light level
	lightnum = int32(viewplayer.Fmo.Fsubsector.Fsector.Flightlevel)>>LIGHTSEGSHIFT + extralight
	if lightnum < 0 {
		spritelights = scalelight[0]
	} else {
		if lightnum >= LIGHTLEVELS {
			spritelights = scalelight[LIGHTLEVELS-1]
		} else {
			spritelights = scalelight[lightnum]
		}
	}
	// clip to screen bounds
	mfloorclip = uintptr(unsafe.Pointer(&screenheightarray))
	mceilingclip = uintptr(unsafe.Pointer(&negonearray))
	// add all active psprites
	for i := range NUMPSPRITES {
		psp := &viewplayer.Fpsprites[i]
		if psp.Fstate != nil {
			r_DrawPSprite(psp)
		}
	}
}

func r_SortVisSprites() {
	bp := &vissprite_t{}
	var best, ds *vissprite_t
	var bestscale fixed_t
	var count int32
	count = int32(vissprite_n)
	bp.Fprev = bp
	bp.Fnext = bp
	if count == 0 {
		return
	}
	for i := 0; i < vissprite_n; i++ {
		ds := &vissprites[i]
		if i < len(vissprites)-1 {
			ds.Fnext = &vissprites[i+1]
		} else {
			ds.Fnext = nil
		}
		if i > 0 {
			ds.Fprev = &vissprites[i-1]
		} else {
			ds.Fprev = nil
		}
	}
	vissprites[0].Fprev = bp
	bp.Fnext = &vissprites[0]
	vissprites[vissprite_n-1].Fnext = bp
	bp.Fprev = &vissprites[vissprite_n-1]
	// pull the vissprites out by scale
	vsprsortedhead.Fprev = &vsprsortedhead
	vsprsortedhead.Fnext = &vsprsortedhead
	for range count {
		bestscale = int32(INT_MAX17)
		best = (*bp).Fnext
		ds = (*bp).Fnext
		for {
			if ds == bp {
				break
			}
			if ds.Fscale < bestscale {
				bestscale = ds.Fscale
				best = ds
			}
			goto _5
		_5:
			;
			ds = ds.Fnext
		}
		best.Fnext.Fprev = best.Fprev
		best.Fprev.Fnext = best.Fnext
		best.Fnext = &vsprsortedhead
		best.Fprev = vsprsortedhead.Fprev
		vsprsortedhead.Fprev.Fnext = best
		vsprsortedhead.Fprev = best
	}
}

// C documentation
//
//	//
//	// R_DrawSprite
//	//
var clipbot [320]int16
var cliptop [320]int16

func r_DrawSprite(spr *vissprite_t) {
	var ds int
	var lowscale, scale fixed_t
	var r1, r2, silhouette, v4, v5 int32
	var v2 int16
	for x := spr.Fx1; x <= spr.Fx2; x++ {
		v2 = int16(-2)
		cliptop[x] = v2
		clipbot[x] = v2
	}
	// Scan drawsegs from end to start for obscuring segs.
	// The first drawseg that has a greater scale
	//  is the clip seg.
	ds = ds_index - 1
	for {
		if !(ds >= 0) {
			break
		}
		// determine if the drawseg obscures the sprite
		if drawsegs[ds].Fx1 > spr.Fx2 || drawsegs[ds].Fx2 < spr.Fx1 || drawsegs[ds].Fsilhouette == 0 && drawsegs[ds].Fmaskedtexturecol == 0 {
			// does not cover sprite
			goto _3
		}
		if drawsegs[ds].Fx1 < spr.Fx1 {
			v4 = spr.Fx1
		} else {
			v4 = drawsegs[ds].Fx1
		}
		r1 = v4
		if drawsegs[ds].Fx2 > spr.Fx2 {
			v5 = spr.Fx2
		} else {
			v5 = drawsegs[ds].Fx2
		}
		r2 = v5
		if drawsegs[ds].Fscale1 > drawsegs[ds].Fscale2 {
			lowscale = drawsegs[ds].Fscale2
			scale = drawsegs[ds].Fscale1
		} else {
			lowscale = drawsegs[ds].Fscale1
			scale = drawsegs[ds].Fscale2
		}
		if scale < spr.Fscale || lowscale < spr.Fscale && r_PointOnSegSide(spr.Fgx, spr.Fgy, drawsegs[ds].Fcurline) == 0 {
			// masked mid texture?
			if drawsegs[ds].Fmaskedtexturecol != 0 {
				r_RenderMaskedSegRange(&drawsegs[ds], r1, r2)
			}
			// seg is behind sprite
			goto _3
		}
		// clip this piece of the sprite
		silhouette = drawsegs[ds].Fsilhouette
		if spr.Fgz >= drawsegs[ds].Fbsilheight {
			silhouette &= ^SIL_BOTTOM
		}
		if spr.Fgzt <= drawsegs[ds].Ftsilheight {
			silhouette &= ^SIL_TOP
		}
		if silhouette == 1 {
			// bottom sil
			for x := r1; x <= r2; x++ {
				if int32(clipbot[x]) == -2 {
					clipbot[x] = *(*int16)(unsafe.Pointer(drawsegs[ds].Fsprbottomclip + uintptr(x)*2))
				}
			}
		} else {
			if silhouette == 2 {
				// top sil
				for x := r1; x <= r2; x++ {
					if int32(cliptop[x]) == -2 {
						cliptop[x] = *(*int16)(unsafe.Pointer(drawsegs[ds].Fsprtopclip + uintptr(x)*2))
					}
				}
			} else {
				if silhouette == 3 {
					// both
					for x := r1; x <= r2; x++ {
						if int32(clipbot[x]) == -2 {
							clipbot[x] = *(*int16)(unsafe.Pointer(drawsegs[ds].Fsprbottomclip + uintptr(x)*2))
						}
						if int32(cliptop[x]) == -2 {
							cliptop[x] = *(*int16)(unsafe.Pointer(drawsegs[ds].Fsprtopclip + uintptr(x)*2))
						}
					}
				}
			}
		}
		goto _3
	_3:
		;
		ds -= 1
	}
	// all clipping has been performed, so draw the sprite
	// check for unclipped columns
	for x := spr.Fx1; x <= spr.Fx2; x++ {
		if int32(clipbot[x]) == -2 {
			clipbot[x] = int16(viewheight)
		}
		if int32(cliptop[x]) == -2 {
			cliptop[x] = int16(-1)
		}
	}
	mfloorclip = uintptr(unsafe.Pointer(&clipbot))
	mceilingclip = uintptr(unsafe.Pointer(&cliptop))
	r_DrawVisSprite(spr, spr.Fx1, spr.Fx2)
}

// C documentation
//
//	//
//	// R_DrawMasked
//	//
func r_DrawMasked() {
	var spr *vissprite_t
	r_SortVisSprites()
	if vissprite_n > 0 {
		// draw all vissprites back to front
		spr = vsprsortedhead.Fnext
		for {
			if spr == &vsprsortedhead {
				break
			}
			r_DrawSprite(spr)
			goto _1
		_1:
			;
			spr = spr.Fnext
		}
	}
	// render any remaining masked mid textures
	ds := ds_index - 1
	for {
		if ds < 0 {
			break
		}
		if drawsegs[ds].Fmaskedtexturecol != 0 {
			r_RenderMaskedSegRange(&drawsegs[ds], drawsegs[ds].Fx1, drawsegs[ds].Fx2)
		}
		goto _2
	_2:
		;
		ds -= 1
	}
	// draw the psprites on top of everything
	//  but does not draw on side views
	if viewangleoffset == 0 {
		r_DrawPlayerSprites()
	}
}

// C documentation
//
//	/* Update the message digest with the contents
//	 * of INBUF with length INLEN.
//	 */
func sha1_Update(sha hash.Hash, buf []byte) {
	sha.Write(buf)
}

func sha1_UpdateInt32(sha hash.Hash, val uint32) {
	var bp [4]byte
	bp[0] = uint8(val >> 24 & 0xff)
	bp[1] = uint8(val >> 16 & 0xff)
	bp[2] = uint8(val >> 8 & 0xff)
	bp[3] = uint8(val & 0xff)
	sha1_Update(sha, bp[:])
}

func sha1_UpdateString(sha hash.Hash, str string) {
	sha1_Update(sha, []byte(str))
	sha1_Update(sha, []byte{0}) // Null-terminate the string
}

func init() {
	S_music = [68]musicinfo_t{
		0: {},
		1: {
			Fname: "e1m1",
		},
		2: {
			Fname: "e1m2",
		},
		3: {
			Fname: "e1m3",
		},
		4: {
			Fname: "e1m4",
		},
		5: {
			Fname: "e1m5",
		},
		6: {
			Fname: "e1m6",
		},
		7: {
			Fname: "e1m7",
		},
		8: {
			Fname: "e1m8",
		},
		9: {
			Fname: "e1m9",
		},
		10: {
			Fname: "e2m1",
		},
		11: {
			Fname: "e2m2",
		},
		12: {
			Fname: "e2m3",
		},
		13: {
			Fname: "e2m4",
		},
		14: {
			Fname: "e2m5",
		},
		15: {
			Fname: "e2m6",
		},
		16: {
			Fname: "e2m7",
		},
		17: {
			Fname: "e2m8",
		},
		18: {
			Fname: "e2m9",
		},
		19: {
			Fname: "e3m1",
		},
		20: {
			Fname: "e3m2",
		},
		21: {
			Fname: "e3m3",
		},
		22: {
			Fname: "e3m4",
		},
		23: {
			Fname: "e3m5",
		},
		24: {
			Fname: "e3m6",
		},
		25: {
			Fname: "e3m7",
		},
		26: {
			Fname: "e3m8",
		},
		27: {
			Fname: "e3m9",
		},
		28: {
			Fname: "inter",
		},
		29: {
			Fname: "intro",
		},
		30: {
			Fname: "bunny",
		},
		31: {
			Fname: "victor",
		},
		32: {
			Fname: "introa",
		},
		33: {
			Fname: "runnin",
		},
		34: {
			Fname: "stalks",
		},
		35: {
			Fname: "countd",
		},
		36: {
			Fname: "betwee",
		},
		37: {
			Fname: "doom",
		},
		38: {
			Fname: "the_da",
		},
		39: {
			Fname: "shawn",
		},
		40: {
			Fname: "ddtblu",
		},
		41: {
			Fname: "in_cit",
		},
		42: {
			Fname: "dead",
		},
		43: {
			Fname: "stlks2",
		},
		44: {
			Fname: "theda2",
		},
		45: {
			Fname: "doom2",
		},
		46: {
			Fname: "ddtbl2",
		},
		47: {
			Fname: "runni2",
		},
		48: {
			Fname: "dead2",
		},
		49: {
			Fname: "stlks3",
		},
		50: {
			Fname: "romero",
		},
		51: {
			Fname: "shawn2",
		},
		52: {
			Fname: "messag",
		},
		53: {
			Fname: "count2",
		},
		54: {
			Fname: "ddtbl3",
		},
		55: {
			Fname: "ampie",
		},
		56: {
			Fname: "theda3",
		},
		57: {
			Fname: "adrian",
		},
		58: {
			Fname: "messg2",
		},
		59: {
			Fname: "romer2",
		},
		60: {
			Fname: "tense",
		},
		61: {
			Fname: "shawn3",
		},
		62: {
			Fname: "openin",
		},
		63: {
			Fname: "evil",
		},
		64: {
			Fname: "ultima",
		},
		65: {
			Fname: "read_m",
		},
		66: {
			Fname: "dm2ttl",
		},
		67: {
			Fname: "dm2int",
		},
	}
}

func init() {
	S_sfx = [109]sfxinfo_t{
		0: {
			Fname:        "none",
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		1: {
			Fname:        "pistol",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		2: {
			Fname:        "shotgn",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		3: {
			Fname:        "sgcock",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		4: {
			Fname:        "dshtgn",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		5: {
			Fname:        "dbopn",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		6: {
			Fname:        "dbcls",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		7: {
			Fname:        "dbload",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		8: {
			Fname:        "plasma",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		9: {
			Fname:        "bfg",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		10: {
			Fname:        "sawup",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		11: {
			Fname:        "sawidl",
			Fpriority:    118,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		12: {
			Fname:        "sawful",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		13: {
			Fname:        "sawhit",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		14: {
			Fname:        "rlaunc",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		15: {
			Fname:        "rxplod",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		16: {
			Fname:        "firsht",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		17: {
			Fname:        "firxpl",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		18: {
			Fname:        "pstart",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		19: {
			Fname:        "pstop",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		20: {
			Fname:        "doropn",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		21: {
			Fname:        "dorcls",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		22: {
			Fname:        "stnmov",
			Fpriority:    119,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		23: {
			Fname:        "swtchn",
			Fpriority:    78,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		24: {
			Fname:        "swtchx",
			Fpriority:    78,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		25: {
			Fname:        "plpain",
			Fpriority:    96,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		26: {
			Fname:        "dmpain",
			Fpriority:    96,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		27: {
			Fname:        "popain",
			Fpriority:    96,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		28: {
			Fname:        "vipain",
			Fpriority:    96,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		29: {
			Fname:        "mnpain",
			Fpriority:    96,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		30: {
			Fname:        "pepain",
			Fpriority:    96,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		31: {
			Fname:        "slop",
			Fpriority:    78,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		32: {
			Fname:        "itemup",
			Fpriority:    78,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		33: {
			Fname:        "wpnup",
			Fpriority:    78,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		34: {
			Fname:        "oof",
			Fpriority:    96,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		35: {
			Fname:        "telept",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		36: {
			Fname:        "posit1",
			Fpriority:    98,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		37: {
			Fname:        "posit2",
			Fpriority:    98,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		38: {
			Fname:        "posit3",
			Fpriority:    98,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		39: {
			Fname:        "bgsit1",
			Fpriority:    98,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		40: {
			Fname:        "bgsit2",
			Fpriority:    98,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		41: {
			Fname:        "sgtsit",
			Fpriority:    98,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		42: {
			Fname:        "cacsit",
			Fpriority:    98,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		43: {
			Fname:        "brssit",
			Fpriority:    94,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		44: {
			Fname:        "cybsit",
			Fpriority:    92,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		45: {
			Fname:        "spisit",
			Fpriority:    90,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		46: {
			Fname:        "bspsit",
			Fpriority:    90,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		47: {
			Fname:        "kntsit",
			Fpriority:    90,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		48: {
			Fname:        "vilsit",
			Fpriority:    90,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		49: {
			Fname:        "mansit",
			Fpriority:    90,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		50: {
			Fname:        "pesit",
			Fpriority:    90,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		51: {
			Fname:        "sklatk",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		52: {
			Fname:        "sgtatk",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		53: {
			Fname:        "skepch",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		54: {
			Fname:        "vilatk",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		55: {
			Fname:        "claw",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		56: {
			Fname:        "skeswg",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		57: {
			Fname:        "pldeth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		58: {
			Fname:        "pdiehi",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		59: {
			Fname:        "podth1",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		60: {
			Fname:        "podth2",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		61: {
			Fname:        "podth3",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		62: {
			Fname:        "bgdth1",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		63: {
			Fname:        "bgdth2",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		64: {
			Fname:        "sgtdth",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		65: {
			Fname:        "cacdth",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		66: {
			Fname:        "skldth",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		67: {
			Fname:        "brsdth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		68: {
			Fname:        "cybdth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		69: {
			Fname:        "spidth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		70: {
			Fname:        "bspdth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		71: {
			Fname:        "vildth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		72: {
			Fname:        "kntdth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		73: {
			Fname:        "pedth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		74: {
			Fname:        "skedth",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		75: {
			Fname:        "posact",
			Fpriority:    120,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		76: {
			Fname:        "bgact",
			Fpriority:    120,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		77: {
			Fname:        "dmact",
			Fpriority:    120,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		78: {
			Fname:        "bspact",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		79: {
			Fname:        "bspwlk",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		80: {
			Fname:        "vilact",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		81: {
			Fname:        "noway",
			Fpriority:    78,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		82: {
			Fname:        "barexp",
			Fpriority:    60,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		83: {
			Fname:        "punch",
			Fpriority:    64,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		84: {
			Fname:        "hoof",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		85: {
			Fname:        "metal",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		86: {
			Fname:        "chgun",
			Fpriority:    64,
			Flink:        &S_sfx[sfx_pistol],
			Fpitch:       150,
			Fnumchannels: -1,
		},
		87: {
			Fname:        "tink",
			Fpriority:    60,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		88: {
			Fname:        "bdopn",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		89: {
			Fname:        "bdcls",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		90: {
			Fname:        "itmbk",
			Fpriority:    100,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		91: {
			Fname:        "flame",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		92: {
			Fname:        "flamst",
			Fpriority:    32,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		93: {
			Fname:        "getpow",
			Fpriority:    60,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		94: {
			Fname:        "bospit",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		95: {
			Fname:        "boscub",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		96: {
			Fname:        "bossit",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		97: {
			Fname:        "bospn",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		98: {
			Fname:        "bosdth",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		99: {
			Fname:        "manatk",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		100: {
			Fname:        "mandth",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		101: {
			Fname:        "sssit",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		102: {
			Fname:        "ssdth",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		103: {
			Fname:        "keenpn",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		104: {
			Fname:        "keendt",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		105: {
			Fname:        "skeact",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		106: {
			Fname:        "skesit",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		107: {
			Fname:        "skeatk",
			Fpriority:    70,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
		108: {
			Fname:        "radio",
			Fpriority:    60,
			Fpitch:       -1,
			Fvolume:      -1,
			Fnumchannels: -1,
		},
	}
}

const MAX_CAPTURES = 32

// Array of end-of-level statistics that have been captured.

var captured_stats [32]wbstartstruct_t
var num_captured_stats int32 = 0

func statCopy(stats *wbstartstruct_t) {
	if m_ParmExists("-statdump") != 0 && num_captured_stats < MAX_CAPTURES {
		captured_stats[num_captured_stats] = *stats
		num_captured_stats++
	}
}

func statDump() {
}

type st_number_t struct {
	Fx      int32
	Fy      int32
	Fwidth  int32
	Foldnum int32
	Fnum    *int32
	Fon     *boolean
	Fp      []*patch_t
	Fdata   weapontype_t
}

type st_percent_t struct {
	Fn st_number_t
	Fp *patch_t
}

type st_multicon_t struct {
	Fx       int32
	Fy       int32
	Foldinum int32
	Finum    *int32
	Fon      *boolean
	Fp       []*patch_t
	Fdata    int32
}

type st_binicon_t struct {
	Fx      int32
	Fy      int32
	Foldval boolean
	Fval    *boolean
	Fon     *boolean
	Fp      *patch_t
	Fdata   int32
}

func stlib_init() {
	sttminus = w_CacheLumpNameT("STTMINUS")
}

// C documentation
//
//	// ?
func stlib_initNum(st *st_number_t, x int32, y int32, pl []*patch_t, num *int32, on *boolean, width int32) {
	st.Fx = x
	st.Fy = y
	st.Foldnum = 0
	st.Fwidth = width
	st.Fnum = num
	st.Fon = on
	st.Fp = pl
}

// C documentation
//
//	//
//	// A fairly efficient way to draw a number
//	//  based on differences from the old number.
//	// Note: worth the trouble?
//	//
func stlib_drawNum(n *st_number_t, refresh boolean) {
	var h, neg, num, numdigits, w, x, v1 int32
	var v2 bool
	numdigits = n.Fwidth
	num = *n.Fnum
	w = int32(n.Fp[0].Fwidth)
	h = int32(n.Fp[0].Fheight)
	// [crispy] redraw only if necessary
	if n.Foldnum == num && refresh == 0 {
		return
	}
	n.Foldnum = *n.Fnum
	neg = boolint32(num < 0)
	if neg != 0 {
		if numdigits == 2 && num < -9 {
			num = -9
		} else {
			if numdigits == 3 && num < -99 {
				num = -99
			}
		}
		num = -num
	}
	// clear the area
	x = n.Fx - numdigits*w
	if n.Fy-(SCREENHEIGHT-st_HEIGHT) < 0 {
		i_Error("drawNum: n->y - st_Y < 0")
	}
	v_CopyRect(x, n.Fy-(SCREENHEIGHT-st_HEIGHT), st_backing_screen, w*numdigits, h, x, n.Fy)
	// if non-number, do not draw it
	if num == 1994 {
		return
	}
	x = n.Fx
	// in the special case of 0, you draw 0
	if num == 0 {
		v_DrawPatch(x-w, n.Fy, n.Fp[0])
	}
	// draw the new number
	for {
		if v2 = num != 0; v2 {
			v1 = numdigits
			numdigits--
		}
		if !(v2 && v1 != 0) {
			break
		}
		x -= w
		v_DrawPatch(x, n.Fy, n.Fp[num%10])
		num /= 10
	}
	// draw a minus sign if necessary
	if neg != 0 {
		v_DrawPatch(x-8, n.Fy, sttminus)
	}
}

// C documentation
//
//	//
func stlib_updateNum(n *st_number_t, refresh boolean) {
	if *n.Fon != 0 {
		stlib_drawNum(n, refresh)
	}
}

// C documentation
//
//	//
func stlib_initPercent(st *st_percent_t, x int32, y int32, pl []*patch_t, num *int32, on *boolean, percent *patch_t) {
	stlib_initNum(&st.Fn, x, y, pl, num, on, 3)
	st.Fp = percent
}

func stlib_updatePercent(per *st_percent_t, refresh int32) {
	if refresh != 0 && *per.Fn.Fon != 0 {
		v_DrawPatch(per.Fn.Fx, per.Fn.Fy, per.Fp)
	}
	stlib_updateNum(&per.Fn, uint32(refresh))
}

func stlib_initMultIcon(st *st_multicon_t, x int32, y int32, il []*patch_t, inum *int32, on *boolean) {
	st.Fx = x
	st.Fy = y
	st.Foldinum = -1
	st.Finum = inum
	st.Fon = on
	st.Fp = il
}

func stlib_updateMultIcon(mi *st_multicon_t, refresh boolean) {
	var h, w, x, y int32
	if *mi.Fon != 0 && (mi.Foldinum != *mi.Finum || refresh != 0) && *mi.Finum != -1 {
		if mi.Foldinum != -1 {
			x = mi.Fx - int32(mi.Fp[mi.Foldinum].Fleftoffset)
			y = mi.Fy - int32(mi.Fp[mi.Foldinum].Ftopoffset)
			w = int32(mi.Fp[mi.Foldinum].Fwidth)
			h = int32(mi.Fp[mi.Foldinum].Fheight)
			if y-(SCREENHEIGHT-st_HEIGHT) < 0 {
				i_Error("updateMultIcon: y - st_Y < 0")
			}
			v_CopyRect(x, y-(SCREENHEIGHT-st_HEIGHT), st_backing_screen, w, h, x, y)
		}
		v_DrawPatch(mi.Fx, mi.Fy, mi.Fp[*mi.Finum])
		mi.Foldinum = *mi.Finum
	}
}

func stlib_initBinIcon(st *st_binicon_t, x int32, y int32, i *patch_t, val *boolean, on *boolean) {
	st.Fx = x
	st.Fy = y
	st.Foldval = 0
	st.Fval = val
	st.Fon = on
	st.Fp = i
}

func stlib_updateBinIcon(bi *st_binicon_t, refresh boolean) {
	var h, w, x, y int32
	if *bi.Fon != 0 && (bi.Foldval != *bi.Fval || refresh != 0) {
		x = bi.Fx - int32(bi.Fp.Fleftoffset)
		y = bi.Fy - int32(bi.Fp.Ftopoffset)
		w = int32(bi.Fp.Fwidth)
		h = int32(bi.Fp.Fheight)
		if y-(SCREENHEIGHT-st_HEIGHT) < 0 {
			i_Error("updateBinIcon: y - st_Y < 0")
		}
		if *bi.Fval != 0 {
			v_DrawPatch(bi.Fx, bi.Fy, bi.Fp)
		} else {
			v_CopyRect(x, y-(SCREENHEIGHT-st_HEIGHT), st_backing_screen, w, h, x, y)
		}
		bi.Foldval = *bi.Fval
	}
}

const ANG18015 = 2147483648
const ANG457 = 536870912
const NUMBONUSPALS = 4
const NUMREDPALS = 8
const RADIATIONPAL = 13
const STARTBONUSPALS = 9
const STARTREDPALS = 1
const st_AMMO0WIDTH = 3
const st_AMMO0X = 288
const st_AMMO0Y = 173
const st_AMMO1X = 288
const st_AMMO1Y = 179
const st_AMMO2X = 288
const st_AMMO2Y = 191
const st_AMMO3X = 288
const st_AMMO3Y = 185
const st_AMMOWIDTH = 3
const st_AMMOX = 44
const st_AMMOY = 171
const st_ARMORX = 221
const st_ARMORY = 171
const st_ARMSBGX = 104
const st_ARMSBGY = 168
const st_ARMSX = 111
const st_ARMSXSPACE = 12
const st_ARMSY = 172
const st_ARMSYSPACE = 10
const st_FACESX = 143
const st_FACESY = 168
const st_FRAGSWIDTH = 2
const st_FRAGSX = 138
const st_FRAGSY = 171
const st_FX = 143
const st_HEALTHX = 90
const st_HEALTHY = 171
const st_KEY0X = 239
const st_KEY0Y = 171
const st_KEY1X = 239
const st_KEY1Y = 181
const st_KEY2X = 239
const st_KEY2Y = 191
const st_MAXAMMO0WIDTH = 3
const st_MAXAMMO0X = 314
const st_MAXAMMO0Y = 173
const st_MAXAMMO1X = 314
const st_MAXAMMO1Y = 179
const st_MAXAMMO2X = 314
const st_MAXAMMO2Y = 191
const st_MAXAMMO3X = 314
const st_MAXAMMO3Y = 185
const st_MUCHPAIN = 20
const st_NUMPAINFACES = 5
const st_NUMSPECIALFACES = 3
const st_NUMSTRAIGHTFACES = 3
const st_NUMTURNFACES = 2
const st_X = 0

// C documentation
//
//	// main player in game
var plyr *player_t

// C documentation
//
//	// st_Start() has just been called
var st_firsttime boolean

// C documentation
//
//	// lump number for PLAYPAL
var lu_palette int32

// C documentation
//
//	// whether left-side main status bar is active
var st_statusbaron boolean

// C documentation
//
//	// !deathmatch
var st_notdeathmatch boolean

// C documentation
//
//	// !deathmatch && st_statusbaron
var st_armson boolean

// C documentation
//
//	// !deathmatch
var st_fragson boolean

// C documentation
//
//	// main bar left
var sbar *patch_t

// C documentation
//
//	// 0-9, tall numbers
var tallnum [10]*patch_t

// C documentation
//
//	// tall % sign
var tallpercent *patch_t

// C documentation
//
//	// 0-9, short, yellow (,different!) numbers
var shortnum [10]*patch_t

// C documentation
//
//	// 3 key-cards, 3 skulls
var keys [6]*patch_t

// C documentation
//
//	// face status patches
var faces [42]*patch_t

// C documentation
//
//	// face background
var faceback *patch_t

// C documentation
//
//	// main bar right
var armsbg *patch_t

// C documentation
//
//	// weapon ownership patches
var arms [6][2]*patch_t

// C documentation
//
//	// ready-weapon widget
var w_ready st_number_t

// C documentation
//
//	// in deathmatch only, summary of frags stats
var w_frags st_number_t

// C documentation
//
//	// health widget
var w_health st_percent_t

// C documentation
//
//	// arms background
var w_armsbg st_binicon_t

// C documentation
//
//	// weapon ownership widgets
var w_arms [6]st_multicon_t

// C documentation
//
//	// face status widget
var w_faces st_multicon_t

// C documentation
//
//	// keycard widgets
var w_keyboxes [3]st_multicon_t

// C documentation
//
//	// armor widget
var w_armor st_percent_t

// C documentation
//
//	// ammo widgets
var w_ammo [4]st_number_t

// C documentation
//
//	// max ammo widgets
var w_maxammo [4]st_number_t

// C documentation
//
//	// number of frags so far in deathmatch
var st_fragscount int32

// C documentation
//
//	// used to use appopriately pained face
var st_oldhealth int32 = -1

// C documentation
//
//	// used for evil grin
var oldweaponsowned [9]boolean

// C documentation
//
//	// count until face changes
var st_facecount int32 = 0

// C documentation
//
//	// current face index, used by w_faces
var st_faceindex int32 = 0

// C documentation
//
//	// holds key-type for each key box on bar
var keyboxes [3]int32

// C documentation
//
//	// a random number per tick
var st_randomnumber int32

func init() {
	cheat_mus = cheatseq_t{
		Fsequence:        "idmus",
		Fsequence_len:    6 - 1,
		Fparameter_chars: 2,
		Fparameter_buf:   [5]byte{},
	}
}

func init() {
	cheat_god = cheatseq_t{
		Fsequence:      "iddqd",
		Fsequence_len:  6 - 1,
		Fparameter_buf: [5]byte{},
	}
}

func init() {
	cheat_ammo = cheatseq_t{
		Fsequence:      "idkfa",
		Fsequence_len:  6 - 1,
		Fparameter_buf: [5]byte{},
	}
}

func init() {
	cheat_ammonokey = cheatseq_t{
		Fsequence:      "idfa",
		Fsequence_len:  5 - 1,
		Fparameter_buf: [5]byte{},
	}
}

func init() {
	cheat_noclip = cheatseq_t{
		Fsequence:      "idspispopd",
		Fsequence_len:  11 - 1,
		Fparameter_buf: [5]byte{},
	}
}

func init() {
	cheat_commercial_noclip = cheatseq_t{
		Fsequence:      "idclip",
		Fsequence_len:  7 - 1,
		Fparameter_buf: [5]byte{},
	}
}

func init() {
	cheat_powerup = [7]cheatseq_t{
		0: {
			Fsequence:      "idbeholdv",
			Fsequence_len:  10 - 1,
			Fparameter_buf: [5]byte{},
		},
		1: {
			Fsequence:      "idbeholds",
			Fsequence_len:  10 - 1,
			Fparameter_buf: [5]byte{},
		},
		2: {
			Fsequence:      "idbeholdi",
			Fsequence_len:  10 - 1,
			Fparameter_buf: [5]byte{},
		},
		3: {
			Fsequence:      "idbeholdr",
			Fsequence_len:  10 - 1,
			Fparameter_buf: [5]byte{},
		},
		4: {
			Fsequence:      "idbeholda",
			Fsequence_len:  10 - 1,
			Fparameter_buf: [5]byte{},
		},
		5: {
			Fsequence:      "idbeholdl",
			Fsequence_len:  10 - 1,
			Fparameter_buf: [5]byte{},
		},
		6: {
			Fsequence:      "idbehold",
			Fsequence_len:  9 - 1,
			Fparameter_buf: [5]byte{},
		},
	}
}

func init() {
	cheat_choppers = cheatseq_t{
		Fsequence:      "idchoppers",
		Fsequence_len:  11 - 1,
		Fparameter_buf: [5]byte{},
	}
}

func init() {
	cheat_clev = cheatseq_t{
		Fsequence:        "idclev",
		Fsequence_len:    7 - 1,
		Fparameter_chars: 2,
		Fparameter_buf:   [5]byte{},
	}
}

func init() {
	cheat_mypos = cheatseq_t{
		Fsequence:      "idmypos",
		Fsequence_len:  8 - 1,
		Fparameter_buf: [5]byte{},
	}
}

func st_refreshBackground() {
	if st_statusbaron != 0 {
		v_UseBuffer(st_backing_screen)
		v_DrawPatch(st_X, 0, sbar)
		if netgame != 0 {
			v_DrawPatch(st_FX, 0, faceback)
		}
		v_RestoreBuffer()
		v_CopyRect(st_X, 0, st_backing_screen, SCREENWIDTH, st_HEIGHT, st_X, SCREENHEIGHT-st_HEIGHT)
	}
}

// C documentation
//
//	// Respond to keyboard input events,
//	//  intercept cheats.
func st_Responder(ev *event_t) boolean {
	var epsd, map1, musnum int32
	var v6, v8 gamemission_t
	var v10 bool
	// Filter automap on/off.
	if ev.Ftype1 == Ev_keyup && uint32(ev.Fdata1)&0xffff0000 == 'a'<<24+'m'<<16 {
		switch ev.Fdata1 {
		case 'a'<<24 + 'm'<<16 | 'e'<<8:
			st_firsttime = 1
		case 'a'<<24 + 'm'<<16 | 'x'<<8:
			//	fprintf(stderr, "AM exited\n");
			break
		}
	} else {
		if ev.Ftype1 == Ev_keydown {
			if netgame == 0 && gameskill != sk_nightmare {
				// 'dqd' cheat for toggleable god mode
				if cht_CheckCheat(&cheat_god, int8(ev.Fdata2)) != 0 {
					plyr.Fcheats ^= CF_GODMODE
					if plyr.Fcheats&CF_GODMODE != 0 {
						if plyr.Fmo != nil {
							plyr.Fmo.Fhealth = 100
						}
						plyr.Fhealth = DEH_DEFAULT_GOD_MODE_HEALTH
						plyr.Fmessage = "Degreelessness Mode On"
					} else {
						plyr.Fmessage = "Degreelessness Mode Off"
					}
				} else {
					if cht_CheckCheat(&cheat_ammonokey, int8(ev.Fdata2)) != 0 {
						plyr.Farmorpoints = DEH_DEFAULT_IDFA_ARMOR
						plyr.Farmortype = DEH_DEFAULT_IDFA_ARMOR_CLASS
						for i := range NUMWEAPONS {
							plyr.Fweaponowned[i] = 1
						}
						for i := range NUMAMMO {
							plyr.Fammo[i] = plyr.Fmaxammo[i]
						}
						plyr.Fmessage = "Ammo (no keys) Added"
					} else {
						if cht_CheckCheat(&cheat_ammo, int8(ev.Fdata2)) != 0 {
							plyr.Farmorpoints = DEH_DEFAULT_IDKFA_ARMOR
							plyr.Farmortype = DEH_DEFAULT_IDKFA_ARMOR_CLASS
							for i := range NUMWEAPONS {
								plyr.Fweaponowned[i] = 1
							}
							for i := range NUMAMMO {
								plyr.Fammo[i] = plyr.Fmaxammo[i]
							}
							for i := range NUMCARDS {
								plyr.Fcards[i] = 1
							}
							plyr.Fmessage = "Very Happy Ammo Added"
						} else {
							if cht_CheckCheat(&cheat_mus, int8(ev.Fdata2)) != 0 {
								plyr.Fmessage = "Music Change"
								var param [5]byte
								cht_GetParam(&cheat_mus, param[:])
								// Note: The original v1.9 had a bug that tried to play back
								// the Doom II music regardless of gamemode.  This was fixed
								// in the Ultimate Doom executable so that it would work for
								// the Doom 1 music as well.
								if gamemode == commercial || gameversion < exe_ultimate {
									musnum = int32(mus_runnin) + (int32((param[0]))-'0')*int32(10) + int32(param[1]) - '0' - 1
									if (int32(param[0])-'0')*int32(10)+int32(param[1])-'0' > 35 {
										plyr.Fmessage = "IMPOSSIBLE SELECTION"
									} else {
										s_ChangeMusic(musnum, 1)
									}
								} else {
									musnum = int32(mus_e1m1) + (int32(param[0])-'1')*9 + (int32(param[1]) - '1')
									if (int32(param[0])-'1')*9+int32(param[1])-'1' > 31 {
										plyr.Fmessage = "IMPOSSIBLE SELECTION"
									} else {
										s_ChangeMusic(musnum, 1)
									}
								}
							} else {
								if gamemission == pack_chex {
									v6 = doom
								} else {
									if gamemission == pack_hacx {
										v6 = doom2
									} else {
										v6 = gamemission
									}
								}
								if v10 = v6 == doom && cht_CheckCheat(&cheat_noclip, int8(ev.Fdata2)) != 0; !v10 {
									if gamemission == pack_chex {
										v8 = doom
									} else {
										if gamemission == pack_hacx {
											v8 = doom2
										} else {
											v8 = gamemission
										}
									}
								}
								if v10 || v8 != doom && cht_CheckCheat(&cheat_commercial_noclip, int8(ev.Fdata2)) != 0 {
									// Noclip cheat.
									// For Doom 1, use the idspipsopd cheat; for all others, use
									// idclip
									plyr.Fcheats ^= CF_NOCLIP
									if plyr.Fcheats&CF_NOCLIP != 0 {
										plyr.Fmessage = "No Clipping Mode ON"
									} else {
										plyr.Fmessage = "No Clipping Mode OFF"
									}
								}
							}
						}
					}
				}
				// 'behold?' power-up cheats
				for i := range int32(6) {
					if cht_CheckCheat(&cheat_powerup[i], int8(ev.Fdata2)) != 0 {
						if plyr.Fpowers[i] == 0 {
							p_GivePower(plyr, i)
						} else {
							if i != int32(pw_strength) {
								plyr.Fpowers[i] = 1
							} else {
								plyr.Fpowers[i] = 0
							}
						}
						plyr.Fmessage = "Power-up Toggled"
					}
				}
				// 'behold' power-up menu
				if cht_CheckCheat(&cheat_powerup[6], int8(ev.Fdata2)) != 0 {
					plyr.Fmessage = "inVuln, Str, Inviso, Rad, Allmap, or Lite-amp"
				} else {
					if cht_CheckCheat(&cheat_choppers, int8(ev.Fdata2)) != 0 {
						plyr.Fweaponowned[wp_chainsaw] = 1
						plyr.Fpowers[pw_invulnerability] = 1
						plyr.Fmessage = "... doesn't suck - GM"
					} else {
						if cht_CheckCheat(&cheat_mypos, int8(ev.Fdata2)) != 0 {
							plyr.Fmessage = fmt.Sprintf("ang=0x%x;x,y=(0x%x,0x%x)", players[consoleplayer].Fmo.Fangle, players[consoleplayer].Fmo.Fx, players[consoleplayer].Fmo.Fy)
						}
					}
				}
			}
			// 'clev' change-level cheat
			if netgame == 0 && cht_CheckCheat(&cheat_clev, int8(ev.Fdata2)) != 0 {
				var param [5]byte
				cht_GetParam(&cheat_clev, param[:])
				if gamemode == commercial {
					epsd = 1
					map1 = (int32(param[0])-'0')*int32(10) + int32(param[1]) - '0'
				} else {
					epsd = int32(param[0]) - '0'
					map1 = int32(param[1]) - '0'
				}
				// Chex.exe always warps to episode 1.
				if gameversion == exe_chex {
					epsd = 1
				}
				// Catch invalid maps.
				if epsd < 1 {
					return 0
				}
				if map1 < 1 {
					return 0
				}
				// Ohmygod - this is not going to work.
				if gamemode == retail && (epsd > 4 || map1 > 9) {
					return 0
				}
				if gamemode == registered && (epsd > 3 || map1 > 9) {
					return 0
				}
				if gamemode == shareware && (epsd > 1 || map1 > 9) {
					return 0
				}
				// The source release has this check as map > 34. However, Vanilla
				// Doom allows IDCLEV up to MAP40 even though it normally crashes.
				if gamemode == commercial && (epsd > 1 || map1 > 40) {
					return 0
				}
				// So be it.
				plyr.Fmessage = "Changing Level..."
				g_DeferedInitNew(gameskill, epsd, map1)
			}
		}
	}
	return 0
}

func st_calcPainOffset() int32 {
	var health, v1 int32
	if plyr.Fhealth > 100 {
		v1 = 100
	} else {
		v1 = plyr.Fhealth
	}
	health = v1
	if health != oldhealth {
		lastcalc = (st_NUMSTRAIGHTFACES + st_NUMTURNFACES + st_NUMSPECIALFACES) * ((int32(100) - health) * st_NUMPAINFACES / 101)
		oldhealth = health
	}
	return lastcalc
}

var lastcalc int32

var oldhealth int32 = -1

// C documentation
//
//	//
//	// This is a not-very-pretty routine which handles
//	//  the face states and their timing.
//	// the precedence of expressions is:
//	//  dead > evil grin > turned head > straight ahead
//	//
func st_updateFaceWidget() {
	var badguyangle, diffang angle_t
	var doevilgrin boolean
	var i, v2 int32
	if priority < 10 {
		// dead
		if plyr.Fhealth == 0 {
			priority = 9
			st_faceindex = st_NUMPAINFACES*(st_NUMSTRAIGHTFACES+st_NUMTURNFACES+st_NUMSPECIALFACES) + 1
			st_facecount = 1
		}
	}
	if priority < 9 {
		if plyr.Fbonuscount != 0 {
			// picking up bonus
			doevilgrin = 0
			for i := range NUMWEAPONS {
				if oldweaponsowned[i] != plyr.Fweaponowned[i] {
					doevilgrin = 1
					oldweaponsowned[i] = plyr.Fweaponowned[i]
				}
			}
			if doevilgrin != 0 {
				// evil grin if just picked up weapon
				priority = 8
				st_facecount = 2 * TICRATE
				st_faceindex = st_calcPainOffset() + (st_NUMSTRAIGHTFACES + st_NUMTURNFACES + 1)
			}
		}
	}
	if priority < 8 {
		if plyr.Fdamagecount != 0 && plyr.Fattacker != nil && plyr.Fattacker != plyr.Fmo {
			// being attacked
			priority = 7
			if plyr.Fhealth-st_oldhealth > st_MUCHPAIN {
				st_facecount = 1 * TICRATE
				st_faceindex = st_calcPainOffset() + (st_NUMSTRAIGHTFACES + st_NUMTURNFACES)
			} else {
				badguyangle = r_PointToAngle2(plyr.Fmo.Fx, plyr.Fmo.Fy, plyr.Fattacker.Fx, plyr.Fattacker.Fy)
				if badguyangle > plyr.Fmo.Fangle {
					// whether right or left
					diffang = badguyangle - plyr.Fmo.Fangle
					i = boolint32(diffang > uint32(ANG18015))
				} else {
					// whether left or right
					diffang = plyr.Fmo.Fangle - badguyangle
					i = boolint32(diffang <= uint32(ANG18015))
				} // confusing, aint it?
				st_facecount = 1 * TICRATE
				st_faceindex = st_calcPainOffset()
				if diffang < uint32(ANG457) {
					// head-on
					st_faceindex += st_NUMSTRAIGHTFACES + st_NUMTURNFACES + 1 + 1
				} else {
					if i != 0 {
						// turn face right
						st_faceindex += st_NUMSTRAIGHTFACES
					} else {
						// turn face left
						st_faceindex += st_NUMSTRAIGHTFACES + 1
					}
				}
			}
		}
	}
	if priority < 7 {
		// getting hurt because of your own damn stupidity
		if plyr.Fdamagecount != 0 {
			if plyr.Fhealth-st_oldhealth > st_MUCHPAIN {
				priority = 7
				st_facecount = 1 * TICRATE
				st_faceindex = st_calcPainOffset() + (st_NUMSTRAIGHTFACES + st_NUMTURNFACES)
			} else {
				priority = 6
				st_facecount = 1 * TICRATE
				st_faceindex = st_calcPainOffset() + (st_NUMSTRAIGHTFACES + st_NUMTURNFACES + 1 + 1)
			}
		}
	}
	if priority < 6 {
		// rapid firing
		if plyr.Fattackdown != 0 {
			if lastattackdown == -1 {
				lastattackdown = 2 * TICRATE
			} else {
				lastattackdown--
				v2 = lastattackdown
				if v2 == 0 {
					priority = 5
					st_faceindex = st_calcPainOffset() + (st_NUMSTRAIGHTFACES + st_NUMTURNFACES + 1 + 1)
					st_facecount = 1
					lastattackdown = 1
				}
			}
		} else {
			lastattackdown = -1
		}
	}
	if priority < 5 {
		// invulnerability
		if plyr.Fcheats&CF_GODMODE != 0 || plyr.Fpowers[pw_invulnerability] != 0 {
			priority = 4
			st_faceindex = st_NUMPAINFACES * (st_NUMSTRAIGHTFACES + st_NUMTURNFACES + st_NUMSPECIALFACES)
			st_facecount = 1
		}
	}
	// look left or look right if the facecount has timed out
	if st_facecount == 0 {
		st_faceindex = st_calcPainOffset() + st_randomnumber%3
		st_facecount = TICRATE / 2
		priority = 0
	}
	st_facecount--
}

var lastattackdown int32 = -1

var priority int32

func st_updateWidgets() {
	var v2 int32
	// must redirect the pointer if the ready weapon has changed.
	//  if (w_ready.data != plyr->readyweapon)
	//  {
	if weaponinfo[plyr.Freadyweapon].Fammo == am_noammo {
		w_ready.Fnum = &largeammo
	} else {
		w_ready.Fnum = &plyr.Fammo[weaponinfo[plyr.Freadyweapon].Fammo]
	}
	//{
	// static int tic=0;
	// static int dir=-1;
	// if (!(tic&15))
	//   plyr->ammo[weaponinfo[plyr->readyweapon].ammo]+=dir;
	// if (plyr->ammo[weaponinfo[plyr->readyweapon].ammo] == -100)
	//   dir = 1;
	// tic++;
	// }
	w_ready.Fdata = plyr.Freadyweapon
	// if (*w_ready.on)
	//  stlib_updateNum(&w_ready, true);
	// refresh weapon change
	//  }
	// update keycard multiple widgets
	for i := range int32(3) {
		if plyr.Fcards[i] != 0 {
			v2 = i
		} else {
			v2 = -1
		}
		keyboxes[i] = v2
		if plyr.Fcards[i+3] != 0 {
			keyboxes[i] = i + 3
		}
	}
	// refresh everything if this is him coming back to life
	st_updateFaceWidget()
	// used by the w_armsbg widget
	st_notdeathmatch = booluint32(deathmatch == 0)
	// used by w_arms[] widgets
	st_armson = booluint32(st_statusbaron != 0 && deathmatch == 0)
	// used by w_frags widget
	st_fragson = booluint32(deathmatch != 0 && st_statusbaron != 0)
	st_fragscount = 0
	for i := range int32(MAXPLAYERS) {
		if i != consoleplayer {
			st_fragscount += plyr.Ffrags[i]
		} else {
			st_fragscount -= plyr.Ffrags[i]
		}
	}
}

var largeammo int32 = 1994

func st_Ticker() {
	st_randomnumber = m_Random()
	st_updateWidgets()
	st_oldhealth = plyr.Fhealth
}

var st_palette int32 = 0

func st_doPaletteStuff() {
	var bzc, cnt, palette int32
	var pal []byte
	cnt = plyr.Fdamagecount
	if plyr.Fpowers[pw_strength] != 0 {
		// slowly fade the berzerk out
		bzc = 12 - plyr.Fpowers[pw_strength]>>6
		if bzc > cnt {
			cnt = bzc
		}
	}
	if cnt != 0 {
		if cnt >= 4 {
			palette = (cnt + 7) >> 3
			if palette >= NUMREDPALS {
				palette = NUMREDPALS - 1
			}
		} else {
			palette = 0
		}
		palette += STARTREDPALS
	} else {
		if plyr.Fbonuscount != 0 {
			if plyr.Fbonuscount >= 4 {
				palette = (plyr.Fbonuscount + 7) >> 3
				if palette >= NUMBONUSPALS {
					palette = NUMBONUSPALS - 1
				}
			} else {
				palette = 0
			}
			palette += STARTBONUSPALS
		} else {
			if plyr.Fpowers[pw_ironfeet] > 4*32 || plyr.Fpowers[pw_ironfeet]&8 != 0 {
				palette = RADIATIONPAL
			} else {
				palette = 0
			}
		}
	}
	// In Chex Quest, the player never sees red.  Instead, the
	// radiation suit palette is used to tint the screen green,
	// as though the player is being covered in goo by an
	// attacking flemoid.
	if gameversion == exe_chex && palette >= STARTREDPALS && palette < STARTREDPALS+NUMREDPALS {
		palette = RADIATIONPAL
	}
	if palette != st_palette {
		st_palette = palette
		pal = w_CacheLumpNumBytes(lu_palette)[palette*768:]
		i_SetPalette(pal)
	}
}

func st_drawWidgets(refresh boolean) {
	// used by w_arms[] widgets
	st_armson = booluint32(st_statusbaron != 0 && deathmatch == 0)
	// used by w_frags widget
	st_fragson = booluint32(deathmatch != 0 && st_statusbaron != 0)
	stlib_updateNum(&w_ready, refresh)
	for i := 0; i < 4; i++ {
		stlib_updateNum(&w_ammo[i], refresh)
		stlib_updateNum(&w_maxammo[i], refresh)
	}
	stlib_updatePercent(&w_health, int32(refresh))
	stlib_updatePercent(&w_armor, int32(refresh))
	stlib_updateBinIcon(&w_armsbg, refresh)
	for i := 0; i < 6; i++ {
		stlib_updateMultIcon(&w_arms[i], refresh)
	}
	stlib_updateMultIcon(&w_faces, refresh)
	for i := 0; i < 3; i++ {
		stlib_updateMultIcon(&w_keyboxes[i], refresh)
	}
	stlib_updateNum(&w_frags, refresh)
}

func st_doRefresh() {
	st_firsttime = 0
	// draw status bar background to off-screen buff
	st_refreshBackground()
	// and refresh all widgets
	st_drawWidgets(1)
}

func st_diffDraw() {
	// update all widgets
	st_drawWidgets(0)
}

func st_Drawer(fullscreen boolean, refresh boolean) {
	st_statusbaron = booluint32(fullscreen == 0 || automapactive != 0)
	st_firsttime = booluint32(st_firsttime != 0 || refresh != 0)
	// Do red-/gold-shifts from damage/items
	st_doPaletteStuff()
	// If just after st_Start(), refresh all
	if st_firsttime != 0 {
		st_doRefresh()
	} else {
		st_diffDraw()
	}
}

// Iterates through all graphics to be loaded or unloaded, along with
// the variable they use, invoking the specified callback function.

func st_loadUnloadGraphics(callback func(string, **patch_t)) {
	var facenum int32
	// Load the numbers, tall and short
	for i := range 10 {
		bp := fmt.Sprintf("STTNUM%d", i)
		callback(bp, &tallnum[i])
		bp = fmt.Sprintf("STYSNUM%d", i)
		callback(bp, &shortnum[i])
	}
	// Load percent key.
	//Note: why not load STMINUS here, too?
	callback("STTPRCNT", &tallpercent)
	// key cards
	for i := range NUMCARDS {
		name := fmt.Sprintf("STKEYS%d", i)
		callback(name, &keys[i])
	}
	// arms background
	callback("STARMS", &armsbg)
	// arms ownership widgets
	for i := range 6 {
		name := fmt.Sprintf("STGNUM%d", i+2)
		// gray #
		callback(name, &arms[i][0])
		// yellow #
		arms[i][1] = shortnum[i+2]
	}
	// face backgrounds for different color players
	name := fmt.Sprintf("STFB%d", consoleplayer)
	callback(name, &faceback)
	// status bar background bits
	callback("STBAR", &sbar)
	// face states
	facenum = 0
	for i := range st_NUMPAINFACES {
		for j := range st_NUMSTRAIGHTFACES {
			name := fmt.Sprintf("STFST%d%d", i, j)
			callback(name, &faces[facenum])
			facenum++
		}
		name := fmt.Sprintf("STFTR%d0", i) // turn right
		callback(name, &faces[facenum])
		facenum++
		name = fmt.Sprintf("STFTL%d0", i) // turn left
		callback(name, &faces[facenum])
		facenum++
		name = fmt.Sprintf("STFOUCH%d", i) // ouch!
		callback(name, &faces[facenum])
		facenum++
		name = fmt.Sprintf("STFEVL%d", i) // evil grin ;)
		callback(name, &faces[facenum])
		facenum++
		name = fmt.Sprintf("STFKILL%d", i) // pissed off
		callback(name, &faces[facenum])
		facenum++
	}
	callback("STFGOD0", &faces[facenum])
	facenum++
	callback("STFDEAD0", &faces[facenum])
	facenum++
}

func st_loadCallback(lumpname string, variable **patch_t) {
	*variable = w_CacheLumpNameT(lumpname)
}

func st_loadGraphics() {
	st_loadUnloadGraphics(st_loadCallback)
}

func st_loadData() {
	lu_palette = w_GetNumForName("PLAYPAL")
	st_loadGraphics()
}

func st_initData() {
	st_firsttime = 1
	plyr = &players[consoleplayer]
	st_statusbaron = 1
	st_faceindex = 0
	st_palette = -1
	st_oldhealth = -1
	for i := range NUMWEAPONS {
		oldweaponsowned[i] = plyr.Fweaponowned[i]
	}
	for i := range keyboxes {
		keyboxes[i] = -1
	}
	stlib_init()
}

func st_createWidgets() {
	// ready weapon ammo
	var fnum *int32
	if weaponinfo[plyr.Freadyweapon].Fammo < 0 || weaponinfo[plyr.Freadyweapon].Fammo >= NUMAMMO {
		fnum = &largeammo
	} else {
		fnum = &plyr.Fammo[weaponinfo[plyr.Freadyweapon].Fammo]
	}
	stlib_initNum(&w_ready, st_AMMOX, st_AMMOY, tallnum[:], fnum, &st_statusbaron, st_AMMOWIDTH)
	// the last weapon type
	w_ready.Fdata = plyr.Freadyweapon
	// health percentage
	stlib_initPercent(&w_health, st_HEALTHX, st_HEALTHY, tallnum[:], &plyr.Fhealth, &st_statusbaron, tallpercent)
	// arms background
	stlib_initBinIcon(&w_armsbg, st_ARMSBGX, st_ARMSBGY, armsbg, &st_notdeathmatch, &st_statusbaron)
	// weapons owned
	for i := int32(0); i < 6; i++ {
		stlib_initMultIcon(&w_arms[i], st_ARMSX+i%3*st_ARMSXSPACE, st_ARMSY+i/3*st_ARMSYSPACE, arms[i][:], (*int32)(unsafe.Pointer(&plyr.Fweaponowned[i+1])), &st_armson)
	}
	// frags sum
	stlib_initNum(&w_frags, st_FRAGSX, st_FRAGSY, tallnum[:], &st_fragscount, &st_fragson, st_FRAGSWIDTH)
	// faces
	stlib_initMultIcon(&w_faces, st_FACESX, st_FACESY, faces[:], &st_faceindex, &st_statusbaron)
	// armor percentage - should be colored later
	stlib_initPercent(&w_armor, st_ARMORX, st_ARMORY, tallnum[:], &plyr.Farmorpoints, &st_statusbaron, tallpercent)
	// keyboxes 0-2
	stlib_initMultIcon(&w_keyboxes[0], st_KEY0X, st_KEY0Y, keys[:], &keyboxes[0], &st_statusbaron)
	stlib_initMultIcon(&w_keyboxes[1], st_KEY1X, st_KEY1Y, keys[:], &keyboxes[1], &st_statusbaron)
	stlib_initMultIcon(&w_keyboxes[2], st_KEY2X, st_KEY2Y, keys[:], &keyboxes[2], &st_statusbaron)
	// ammo count (all four kinds)
	stlib_initNum(&w_ammo[0], st_AMMO0X, st_AMMO0Y, shortnum[:], &plyr.Fammo[0], &st_statusbaron, st_AMMO0WIDTH)
	stlib_initNum(&w_ammo[1], st_AMMO1X, st_AMMO1Y, shortnum[:], &plyr.Fammo[1], &st_statusbaron, st_AMMO0WIDTH)
	stlib_initNum(&w_ammo[2], st_AMMO2X, st_AMMO2Y, shortnum[:], &plyr.Fammo[2], &st_statusbaron, st_AMMO0WIDTH)
	stlib_initNum(&w_ammo[3], st_AMMO3X, st_AMMO3Y, shortnum[:], &plyr.Fammo[3], &st_statusbaron, st_AMMO0WIDTH)
	// max ammo count (all four kinds)
	stlib_initNum(&w_maxammo[0], st_MAXAMMO0X, st_MAXAMMO0Y, shortnum[:], &plyr.Fmaxammo[0], &st_statusbaron, st_MAXAMMO0WIDTH)
	stlib_initNum(&w_maxammo[1], st_MAXAMMO1X, st_MAXAMMO1Y, shortnum[:], &plyr.Fmaxammo[1], &st_statusbaron, st_MAXAMMO0WIDTH)
	stlib_initNum(&w_maxammo[2], st_MAXAMMO2X, st_MAXAMMO2Y, shortnum[:], &plyr.Fmaxammo[2], &st_statusbaron, st_MAXAMMO0WIDTH)
	stlib_initNum(&w_maxammo[3], st_MAXAMMO3X, st_MAXAMMO3Y, shortnum[:], &plyr.Fmaxammo[3], &st_statusbaron, st_MAXAMMO0WIDTH)
}

var st_stopped int32 = 1

func st_Start() {
	if st_stopped == 0 {
		st_Stop()
	}
	st_initData()
	st_createWidgets()
	st_stopped = 0
}

func st_Stop() {
	if st_stopped != 0 {
		return
	}
	i_SetPalette(w_CacheLumpNumBytes(lu_palette))
	st_stopped = 1
}

func st_Init() {
	st_loadData()
	st_backing_screen = make([]byte, SCREENWIDTH*st_HEIGHT)
}

const NORM_SEP = 128

//
// This is used to get the local FILE:LINE info from CPP
// prior to really call the function in question.
//

// when to clip out sounds
// Does not fit the large outdoor areas.

// Distance tp origin when sounds should be maxed out.
// This should relate to movement clipping resolution
// (see BLOCKMAP handling).
// In the source code release: (160*FRACUNIT).  Changed back to the
// Vanilla value of 200 (why was this changed?)

// The range over which sound attenuates

// Stereo separation

type channel_t struct {
	Fsfxinfo *sfxinfo_t
	Forigin  *degenmobj_t
	Fhandle  int32
}

// The set of channels available

var channels []channel_t

func init() {
	sfxVolume = 8
	musicVolume = 8
	snd_channels = 8
}

// Internal volume level, ranging from 0-127

var snd_SfxVolume int32

// Whether songs are mus_paused

var mus_paused boolean

// Music currently being played

var mus_playing *musicinfo_t

//
// Initializes sound stuff, including volume
// Sets channels, SFX and music volume,
//  allocates channel buffer, sets S_sfx lookup.
//

func s_Init(sfxVolume int32, musicVolume int32) {
	var v3 int32
	i_PrecacheSounds(S_sfx[:])
	s_SetSfxVolume(sfxVolume)
	s_SetMusicVolume(musicVolume)
	// Allocating the internal channels for mixing
	// (the maximum numer of sounds rendered
	// simultaneously) within zone memory.
	channels = make([]channel_t, snd_channels)
	// no sounds are playing, and they are not mus_paused
	mus_paused = 0
	// Note that sounds have not been cached (yet).
	for i := 1; i < NUMSFX; i++ {
		v3 = -1
		S_sfx[i].Fusefulness = v3
		S_sfx[i].Flumpnum = v3
	}
	i_AtExit(s_Shutdown, 1)
}

func s_Shutdown() {
	i_ShutdownSound()
	i_ShutdownMusic()
}

func s_StopChannel(cnum int32) {
	c := &channels[cnum]
	if c.Fsfxinfo != nil {
		// stop the sound playing
		if i_SoundIsPlaying(c.Fhandle) != 0 {
			i_StopSound(c.Fhandle)
		}
		// check to see if other channels are playing the sound
		for i := range snd_channels {
			if cnum != i && c.Fsfxinfo == channels[i].Fsfxinfo {
				break
			}
		}
		// degrade usefulness of sound data
		c.Fsfxinfo.Fusefulness--
		c.Fsfxinfo = nil
	}
}

//
// Per level startup code.
// Kills playing sounds at start of level,
//  determines music if any, changes music.
//

func s_Start() {
	var mnum int32
	var spmus [9]int32
	// kill all playing sounds at start of level
	//  (trust me - a good idea)
	for cnum := range snd_channels {
		if channels[cnum].Fsfxinfo != nil {
			s_StopChannel(cnum)
		}
	}
	// start new music for the level
	mus_paused = 0
	if gamemode == commercial {
		mnum = int32(mus_runnin) + gamemap - 1
	} else {
		spmus = [9]int32{
			0: int32(mus_e3m4),
			1: int32(mus_e3m2),
			2: int32(mus_e3m3),
			3: int32(mus_e1m5),
			4: int32(mus_e2m7),
			5: int32(mus_e2m4),
			6: int32(mus_e2m6),
			7: int32(mus_e2m5),
			8: int32(mus_e1m9),
		}
		if gameepisode < 4 {
			mnum = int32(mus_e1m1) + (gameepisode-1)*9 + gamemap - 1
		} else {
			mnum = spmus[gamemap-1]
		}
	}
	s_ChangeMusic(mnum, 1)
}

func s_StopSound(origin *degenmobj_t) {
	for cnum := range snd_channels {
		if channels[cnum].Fsfxinfo != nil && channels[cnum].Forigin == origin {
			s_StopChannel(cnum)
			break
		}
	}
}

//
// s_GetChannel :
//   If none available, return -1.  Otherwise channel #.
//

func s_GetChannel(origin *degenmobj_t, sfxinfo *sfxinfo_t) int32 {
	var cnum int32
	// Find an open channel
	for cnum = range snd_channels {
		if channels[cnum].Fsfxinfo == nil {
			break
		} else {
			if origin != nil && channels[cnum].Forigin == origin {
				s_StopChannel(cnum)
				break
			}
		}
	}
	// None available
	if cnum == snd_channels {
		// Look for lower priority
		for cnum = range snd_channels {
			if channels[cnum].Fsfxinfo.Fpriority >= sfxinfo.Fpriority {
				break
			}
		}
		if cnum == snd_channels {
			// FUCK!  No lower priority.  Sorry, Charlie.
			return -1
		} else {
			// Otherwise, kick out lower priority.
			s_StopChannel(cnum)
		}
	}
	c := &channels[cnum]
	// channel is decided to be cnum.
	c.Fsfxinfo = sfxinfo
	c.Forigin = origin
	return cnum
}

//
// Changes volume and stereo-separation variables
//  from the norm of a sound effect to be played.
// If the sound is not audible, returns a 0.
// Otherwise, modifies parameters and returns 1.
//

func s_AdjustSoundParams(listener *degenmobj_t, source *degenmobj_t, vol *int32, sep *int32) int32 {
	var adx, ady, approx_dist fixed_t
	var angle angle_t
	var v1 int32
	// calculate the distance to sound origin
	//  and clip it if necessary
	adx = xabs(listener.Fx - source.Fx)
	ady = xabs(listener.Fy - source.Fy)
	// From _GG1_ p.428. Appox. eucledian distance fast.
	if adx < ady {
		v1 = adx
	} else {
		v1 = ady
	}
	approx_dist = adx + ady - v1>>1
	if gamemap != 8 && approx_dist > 1200*(1<<FRACBITS) {
		return 0
	}
	// angle of source to listener
	angle = r_PointToAngle2(listener.Fx, listener.Fy, source.Fx, source.Fy)
	// TODO: Andre/GORE: Is this a safe cast? Can we guarantee this isn't just a degenmobj_t?
	mo := (*mobj_t)(unsafe.Pointer(listener))
	if angle > mo.Fangle {
		angle = angle - mo.Fangle
	} else {
		angle = angle + (0xffffffff - mo.Fangle)
	}
	angle >>= ANGLETOFINESHIFT
	// stereo separation
	*sep = 128 - fixedMul(96*(1<<FRACBITS), finesine[angle])>>FRACBITS
	// volume calculation
	if approx_dist < 200*(1<<FRACBITS) {
		*vol = snd_SfxVolume
	} else {
		if gamemap == 8 {
			if approx_dist > 1200*(1<<FRACBITS) {
				approx_dist = 1200 * (1 << FRACBITS)
			}
			*vol = 15 + (snd_SfxVolume-int32(15))*((1200*(1<<FRACBITS)-approx_dist)>>FRACBITS)/((1200*(1<<FRACBITS)-200*(1<<FRACBITS))>>FRACBITS)
		} else {
			// distance effect
			*vol = snd_SfxVolume * ((1200*(1<<FRACBITS) - approx_dist) >> FRACBITS) / ((1200*(1<<FRACBITS) - 200*(1<<FRACBITS)) >> FRACBITS)
		}
	}
	return boolint32(*vol > 0)
}

func s_StartSound(origin *degenmobj_t, sfx_id int32) {
	var channel, volume int32
	var cnum, rc int32
	var sfx *sfxinfo_t
	volume = snd_SfxVolume
	// check for bogus sound #
	if sfx_id < 1 || sfx_id > NUMSFX {
		i_Error("Bad sfx #: %d", sfx_id)
	}
	sfx = &S_sfx[sfx_id]
	// Initialize sound parameters
	if sfx.Flink != nil {
		volume += sfx.Fvolume
		if volume < 1 {
			return
		}
		if volume > snd_SfxVolume {
			volume = snd_SfxVolume
		}
	}
	// Check to see if it is audible,
	//  and if not, modify the params
	if origin != nil && origin != &players[consoleplayer].Fmo.degenmobj_t {
		rc = s_AdjustSoundParams(&players[consoleplayer].Fmo.degenmobj_t, origin, &volume, &channel)
		if origin.Fx == players[consoleplayer].Fmo.Fx && origin.Fy == players[consoleplayer].Fmo.Fy {
			channel = NORM_SEP
		}
		if rc == 0 {
			return
		}
	} else {
		channel = NORM_SEP
	}
	// kill old sound
	s_StopSound(origin)
	// try to find a channel
	cnum = s_GetChannel(origin, sfx)
	if cnum < 0 {
		return
	}
	// increase the usefulness
	sfx.Fusefulness = max(1, sfx.Fusefulness+1)
	if sfx.Flumpnum < 0 {
		sfx.Flumpnum = i_GetSfxLumpNum(sfx)
	}
	channels[cnum].Fhandle = i_StartSound(sfx, cnum, volume, channel)
}

//
// Stop and resume music, during game PAUSE.
//

func s_PauseSound() {
	if mus_playing != nil && mus_paused == 0 {
		i_PauseSong()
		mus_paused = 1
	}
}

func s_ResumeSound() {
	if mus_playing != nil && mus_paused != 0 {
		i_ResumeSong()
		mus_paused = 0
	}
}

//
// Updates music & sounds
//

func s_UpdateSounds(listener *degenmobj_t) {
	var volume, channel int32
	var audible int32
	var sfx *sfxinfo_t
	i_UpdateSound()
	for cnum := range snd_channels {
		c := &channels[cnum]
		sfx = c.Fsfxinfo
		if sfx != nil {
			if i_SoundIsPlaying(c.Fhandle) != 0 {
				// initialize parameters
				volume = snd_SfxVolume
				channel = NORM_SEP
				if sfx.Flink != nil {
					volume += sfx.Fvolume
					if volume < 1 {
						s_StopChannel(cnum)
						continue
					} else {
						if volume > snd_SfxVolume {
							volume = snd_SfxVolume

						}
					}
				}
				// check non-local sounds for distance clipping
				//  or modify their params
				if c.Forigin != nil && listener != c.Forigin {
					audible = s_AdjustSoundParams(listener, c.Forigin, &volume, &channel)
					if audible == 0 {
						s_StopChannel(cnum)
					} else {
						i_UpdateSoundParams(c.Fhandle, volume, channel)
					}
				}
			} else {
				// if channel is allocated but sound has stopped,
				//  free it
				s_StopChannel(cnum)
			}
		}
	}
}

func s_SetMusicVolume(volume int32) {
	if volume < 0 || volume > 127 {
		i_Error("Attempt to set music volume at %d", volume)
	}
	i_SetMusicVolume(volume)
}

func s_SetSfxVolume(volume int32) {
	if volume < 0 || volume > 127 {
		i_Error("Attempt to set sfx volume at %d", volume)
	}
	snd_SfxVolume = volume
}

//
// Starts some music with the music id found in sounds.h.
//

func s_StartMusic(m_id int32) {
	s_ChangeMusic(m_id, 0)
}

func s_ChangeMusic(musicnum int32, looping int32) {
	var handle uintptr
	var music *musicinfo_t
	// The Doom IWAD file has two versions of the intro music: d_intro
	// and d_introa.  The latter is used for OPL playback.
	if musicnum == int32(mus_intro) && (snd_musicdevice == SNDDEVICE_ADLIB || snd_musicdevice == SNDDEVICE_SB) {
		musicnum = int32(mus_introa)
	}
	if musicnum <= int32(mus_None) || musicnum >= NUMMUSIC {
		i_Error("Bad music number %d", musicnum)
	} else {
		music = &S_music[musicnum]
	}
	if mus_playing == music {
		return
	}
	// shutdown old music
	s_StopMusic()
	// get lumpnum if neccessary
	if music.Flumpnum == 0 {
		bp := fmt.Sprintf("d_%s", music.Fname)
		music.Flumpnum = w_GetNumForName(bp)
	}
	music.Fdata = w_CacheLumpNumBytes(music.Flumpnum)
	handle = i_RegisterSong(music.Fdata)
	music.Fhandle = handle
	i_PlaySong(handle, uint32(looping))
	mus_playing = music
}

func s_StopMusic() {
	if mus_playing != nil {
		if mus_paused != 0 {
			i_ResumeSong()
		}
		i_StopSong()
		i_UnRegisterSong(mus_playing.Fhandle)
		w_ReleaseLumpNum(mus_playing.Flumpnum)
		mus_playing.Fdata = nil
		mus_playing = nil
	}
}

// to get a global angle from cartesian coordinates, the coordinates are
// flipped until they are in the first octant of the coordinate system, then
// the y (<=x) is scaled and divided by x to get a tangent (slope) value
// which is looked up in the tantoangle[] table.  The +1 size is to handle
// the case when x==y without additional checking.

func slopeDiv(num uint32, den uint32) int32 {
	var ans uint32
	if den < 512 {
		return SLOPERANGE
	} else {
		ans = num << 3 / (den >> 8)
		if ans <= SLOPERANGE {
			return int32(ans)
		} else {
			return SLOPERANGE
		}
	}
}

func init() {
	// Calculate all the various tables

	finetangent = [4096]fixed_t{}
	finesine = [10240]fixed_t{}

	for i := range finetangent {
		a := (float64(i-FINEANGLES/4) + 0.5) * (math.Pi * 2 / FINEANGLES)
		fv := fixed_t(math.Tan(a) * FRACUNIT)
		t := fv
		finetangent[i] = t
	}
	for i := range finesine {
		a := (float64(i) + 0.5) * (math.Pi * 2 / FINEANGLES)
		t := fixed_t(math.Sin(a) * FRACUNIT)
		finesine[i] = t
	}

	finecosine = finesine[FINEANGLES/4:]
	//finecosine = uintptr(unsafe.Pointer(&finesine)) + uintptr(FINEANGLES/4)*4

	tantoangle = [2049]angle_t{}
	for i := range len(tantoangle) {
		f := math.Atan(float64(i)/float64(SLOPERANGE)) / (math.Pi * 2)
		t := uint32(f * 0xffffffff)
		tantoangle[i] = angle_t(t)
	}
}

const MOUSE_SPEED_BOX_HEIGHT = 9
const MOUSE_SPEED_BOX_WIDTH = 120

// The screen buffer that the v_video.c code draws to.

var dest_screen []byte

// C documentation
//
//	//
//	// V_MarkRect
//	//
func v_MarkRect(x int32, y int32, width int32, height int32) {
	// If we are temporarily using an alternate screen, do not
	// affect the update box.
	if reflect.DeepEqual(dest_screen, I_VideoBuffer) {
		m_AddToBox(&dirtybox, x, y)
		m_AddToBox(&dirtybox, x+width-1, y+height-1)
	}
}

// C documentation
//
//	//
//	// V_CopyRect
//	//
func v_CopyRect(srcx int32, srcy int32, source []byte, width int32, height int32, destx int32, desty int32) {
	if srcx < 0 || srcx+width > SCREENWIDTH || srcy < 0 || srcy+height > SCREENHEIGHT || destx < 0 || destx+width > SCREENWIDTH || desty < 0 || desty+height > SCREENHEIGHT {
		i_Error("Bad v_CopyRect")
	}
	v_MarkRect(destx, desty, width, height)
	srcPos := SCREENWIDTH*srcy + srcx
	destPos := SCREENWIDTH*desty + destx
	for {
		if height <= 0 {
			break
		}
		copy(dest_screen[destPos:destPos+width], source[srcPos:srcPos+width])
		srcPos += SCREENWIDTH
		destPos += SCREENWIDTH
		goto _1
	_1:
		;
		height--
	}
}

//
// V_DrawPatch
// Masks a column based masked pic to the screen.
//

func v_DrawPatch(x int32, y int32, patch *patch_t) {
	var col, w int32
	y -= int32(patch.Ftopoffset)
	x -= int32(patch.Fleftoffset)
	if x < 0 || x+int32(patch.Fwidth) > SCREENWIDTH || y < 0 || y+int32(patch.Fheight) > SCREENHEIGHT {
		i_Error("Bad v_DrawPatch x=%d y=%d patch.width=%d patch.height=%d topoffset=%d leftoffset=%d", x, y, int32(patch.Fwidth), int32(patch.Fheight), int32(patch.Ftopoffset), int32(patch.Fleftoffset))
	}
	v_MarkRect(x, y, int32(patch.Fwidth), int32(patch.Fheight))
	col = 0
	w = int32(patch.Fwidth)
	for {
		if col >= w {
			break
		}
		column := patch.GetColumn(col)
		// step through the posts in a column
		for int32(column.Ftopdelta) != 0xff {
			source := column.Data()
			pos := (y * SCREENWIDTH) + x + col + int32(column.Ftopdelta)*SCREENWIDTH
			for i := int32(0); i < int32(column.Flength); i++ {
				dest_screen[pos] = source[i]
				pos += SCREENWIDTH
			}
			column = column.Next()
		}
		goto _1
	_1:
		;
		col++
	}
}

//
// V_DrawPatchFlipped
// Masks a column based masked pic to the screen.
// Flips horizontally, e.g. to mirror face.
//

func v_DrawPatchFlipped(x int32, y int32, patch *patch_t) {
	var col, w int32
	//var dest, desttop uintptr
	y -= int32(patch.Ftopoffset)
	x -= int32(patch.Fleftoffset)
	if x < 0 || x+int32(patch.Fwidth) > SCREENWIDTH || y < 0 || y+int32(patch.Fheight) > SCREENHEIGHT {
		i_Error("Bad v_DrawPatchFlipped")
	}
	v_MarkRect(x, y, int32(patch.Fwidth), int32(patch.Fheight))
	col = 0
	destTop := y*SCREENWIDTH + x
	w = int32(patch.Fwidth)
	for {
		if col >= w {
			break
		}
		column := patch.GetColumn(w - 1 - col)
		// step through the posts in a column
		for int32(column.Ftopdelta) != 0xff {
			source := column.Data()
			destPos := destTop + int32(column.Ftopdelta)*SCREENWIDTH
			for i := uint8(0); i < column.Flength; i++ {
				dest_screen[destPos+SCREENWIDTH*int32(i)] = source[i]
			}
			column = column.Next()
		}
		goto _1
	_1:
		;
		col++
		destTop++
	}
}

//
// V_DrawPatchDirect
// Draws directly to the screen on the pc.
//

func v_DrawPatchDirect(x int32, y int32, patch *patch_t) {
	v_DrawPatch(x, y, patch)
}

//
// V_DrawBlock
// Draw a linear block of pixels into the view buffer.
//

func v_DrawBlock(x int32, y int32, width int32, height int32, src []byte) {
	var v1 int32
	var pos int32
	if x < 0 || x+width > SCREENWIDTH || y < 0 || y+height > SCREENHEIGHT {
		i_Error("Bad v_DrawBlock")
	}
	v_MarkRect(x, y, width, height)
	destPos := y*SCREENWIDTH + x
	for {
		v1 = height
		height--
		if v1 == 0 {
			break
		}
		copy(dest_screen[destPos:destPos+width], src)
		pos += width
		destPos += SCREENWIDTH
	}
}

func v_DrawFilledBox(x int32, y int32, w int32, h int32, c int32) {
	var x1 int32
	pos := SCREENWIDTH*y + x
	for y1 := int32(0); y1 < h; y1++ {
		for x := int32(0); x < w; x++ {
			I_VideoBuffer[pos+x1] = uint8(c)
		}
		pos += SCREENWIDTH
	}
}

func v_DrawHorizLine(x int32, y int32, w int32, c int32) {
	pos := SCREENWIDTH*y + x
	for x1 := int32(0); x1 < w; x1++ {
		I_VideoBuffer[pos+x1] = uint8(c)
	}
}

func v_DrawVertLine(x int32, y int32, h int32, c int32) {
	pos := SCREENWIDTH*y + x
	for y1 := int32(0); y1 < h; y1++ {
		I_VideoBuffer[pos] = uint8(c)
		pos += SCREENWIDTH
	}
}

func v_DrawBox(x int32, y int32, w int32, h int32, c int32) {
	v_DrawHorizLine(x, y, w, c)
	v_DrawHorizLine(x, y+h-1, w, c)
	v_DrawVertLine(x, y, h, c)
	v_DrawVertLine(x+w-1, y, h, c)
}

// C documentation
//
//	//
//	// V_Init
//	//
func v_Init() {
	// no-op!
	// There used to be separate screens that could be drawn to; these are
	// now handled in the upper layers.
}

// Set the buffer that the code draws to.

func v_UseBuffer(buffer []byte) {
	dest_screen = buffer
}

// Restore screen buffer to the i_video screen buffer.

func v_RestoreBuffer() {
	dest_screen = I_VideoBuffer
}

func v_DrawMouseSpeedBox(speed int32) {
	var bgcolor, black, bordercolor, box_x, box_y, linelen, original_speed, red, redline_x, white, yellow int32
	// Get palette indices for colors for widget. These depend on the
	// palette of the game being played.
	bgcolor = i_GetPaletteIndex(0x77, 0x77, 0x77)
	bordercolor = i_GetPaletteIndex(0x55, 0x55, 0x55)
	red = i_GetPaletteIndex(0xff, 0x00, 0x00)
	black = i_GetPaletteIndex(0x00, 0x00, 0x00)
	yellow = i_GetPaletteIndex(0xff, 0xff, 0x00)
	white = i_GetPaletteIndex(0xff, 0xff, 0xff)
	// If the mouse is turned off or acceleration is turned off, don't
	// draw the box at all.
	if usemouse == 0 || math.Abs(float64(mouse_acceleration-1)) < float64(0.01) {
		return
	}
	// Calculate box position
	box_x = SCREENWIDTH - MOUSE_SPEED_BOX_WIDTH - 10
	box_y = 15
	v_DrawFilledBox(box_x, box_y, MOUSE_SPEED_BOX_WIDTH, MOUSE_SPEED_BOX_HEIGHT, bgcolor)
	v_DrawBox(box_x, box_y, MOUSE_SPEED_BOX_WIDTH, MOUSE_SPEED_BOX_HEIGHT, bordercolor)
	// Calculate the position of the red line.  This is 1/3 of the way
	// along the box.
	redline_x = MOUSE_SPEED_BOX_WIDTH / 3
	// Undo acceleration and get back the original mouse speed
	if speed < mouse_threshold {
		original_speed = speed
	} else {
		original_speed = speed - mouse_threshold
		original_speed = int32(float32(original_speed) / mouse_acceleration)
		original_speed += mouse_threshold
	}
	// Calculate line length
	linelen = original_speed * redline_x / mouse_threshold
	// Draw horizontal "thermometer"
	if linelen > MOUSE_SPEED_BOX_WIDTH-1 {
		linelen = MOUSE_SPEED_BOX_WIDTH - 1
	}
	v_DrawHorizLine(box_x+1, box_y+4, MOUSE_SPEED_BOX_WIDTH-2, black)
	if linelen < redline_x {
		v_DrawHorizLine(box_x+1, box_y+MOUSE_SPEED_BOX_HEIGHT/2, linelen, white)
	} else {
		v_DrawHorizLine(box_x+1, box_y+MOUSE_SPEED_BOX_HEIGHT/2, redline_x, white)
		v_DrawHorizLine(box_x+redline_x, box_y+MOUSE_SPEED_BOX_HEIGHT/2, linelen-redline_x, yellow)
	}
	// Draw red line
	v_DrawVertLine(box_x+redline_x, box_y+1, MOUSE_SPEED_BOX_HEIGHT-2, red)
}

const DM_KILLERSX = 10
const DM_KILLERSY = 100
const DM_MATRIXX = 42
const DM_MATRIXY = 68
const DM_SPACINGX = 40
const DM_TOTALSX = 269
const DM_VICTIMSX = 5
const DM_VICTIMSY = 50
const NG_SPACINGX = 64
const NG_STATSY = 50
const NUMMAPS = 9
const SHOWNEXTLOCDELAY = 4
const SP_STATSX = 50
const SP_STATSY = 50
const SP_TIMEX = 16
const WI_SPACINGY = 33
const WI_TITLEY = 2

//
// Data needed to add patches to full screen intermission pics.
// Patches are statistics messages, and animations.
// Loads of by-pixel layout and placement, offsets etc.
//

//
// Different vetween registered DOOM (1994) and
//  Ultimate DOOM - Final edition (retail, 1995?).
// This is supposedly ignored for commercial
//  release (aka DOOM II), which had 34 maps
//  in one episode. So there.

// in tics
//U #define PAUSELEN		(TICRATE*2)
//U #define SCORESTEP		100
//U #define ANIMPERIOD		32
// pixel distance from "(YOU)" to "PLAYER N"
//U #define STARDIST		10
//U #define WK 1

// GLOBAL LOCATIONS

// SINGPLE-PLAYER STUFF

// NET GAME STUFF

// DEATHMATCH STUFF

type animenum_t = int32

const ANIM_ALWAYS = 0
const ANIM_RANDOM = 1
const ANIM_LEVEL = 2

type point_t struct {
	Fx int32
	Fy int32
}

// C documentation
//
//	//
//	// Animation.
//	// There is another anim_t used in p_spec.
//	//
type anim_t1 struct {
	Ftype1     animenum_t
	Fperiod    int32
	Fnanims    int32
	Floc       point_t
	Fdata1     int32
	Fdata2     int32
	Fp         [3]*patch_t
	Fnexttic   int32
	Flastdrawn int32
	Fctr       int32
	Fstate     int32
}

var lnodes = [4][9]point_t{
	0: {
		0: {
			Fx: 185,
			Fy: 164,
		},
		1: {
			Fx: 148,
			Fy: 143,
		},
		2: {
			Fx: 69,
			Fy: 122,
		},
		3: {
			Fx: 209,
			Fy: 102,
		},
		4: {
			Fx: 116,
			Fy: 89,
		},
		5: {
			Fx: 166,
			Fy: 55,
		},
		6: {
			Fx: 71,
			Fy: 56,
		},
		7: {
			Fx: 135,
			Fy: 29,
		},
		8: {
			Fx: 71,
			Fy: 24,
		},
	},
	1: {
		0: {
			Fx: 254,
			Fy: 25,
		},
		1: {
			Fx: 97,
			Fy: 50,
		},
		2: {
			Fx: 188,
			Fy: 64,
		},
		3: {
			Fx: 128,
			Fy: 78,
		},
		4: {
			Fx: 214,
			Fy: 92,
		},
		5: {
			Fx: 133,
			Fy: 130,
		},
		6: {
			Fx: 208,
			Fy: 136,
		},
		7: {
			Fx: 148,
			Fy: 140,
		},
		8: {
			Fx: 235,
			Fy: 158,
		},
	},
	2: {
		0: {
			Fx: 156,
			Fy: 168,
		},
		1: {
			Fx: 48,
			Fy: 154,
		},
		2: {
			Fx: 174,
			Fy: 95,
		},
		3: {
			Fx: 265,
			Fy: 75,
		},
		4: {
			Fx: 130,
			Fy: 48,
		},
		5: {
			Fx: 279,
			Fy: 23,
		},
		6: {
			Fx: 198,
			Fy: 48,
		},
		7: {
			Fx: 140,
			Fy: 25,
		},
		8: {
			Fx: 281,
			Fy: 136,
		},
	},
}

//
// Animation locations for episode 0 (1).
// Using patches saves a lot of space,
//  as they replace 320x200 full screen frames.
//

var epsd0animinfo = [10]anim_t1{
	0: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 224,
			Fy: 104,
		},
	},
	1: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 184,
			Fy: 160,
		},
	},
	2: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 112,
			Fy: 136,
		},
	},
	3: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 72,
			Fy: 112,
		},
	},
	4: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 88,
			Fy: 96,
		},
	},
	5: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 64,
			Fy: 48,
		},
	},
	6: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 192,
			Fy: 40,
		},
	},
	7: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 136,
			Fy: 16,
		},
	},
	8: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 80,
			Fy: 16,
		},
	},
	9: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 64,
			Fy: 24,
		},
	},
}

var epsd1animinfo = [9]anim_t1{
	0: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 1,
		Floc: point_t{
			Fx: 128,
			Fy: 136,
		},
		Fdata1: 1,
	},
	1: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 1,
		Floc: point_t{
			Fx: 128,
			Fy: 136,
		},
		Fdata1: 2,
	},
	2: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 1,
		Floc: point_t{
			Fx: 128,
			Fy: 136,
		},
		Fdata1: 3,
	},
	3: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 1,
		Floc: point_t{
			Fx: 128,
			Fy: 136,
		},
		Fdata1: 4,
	},
	4: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 1,
		Floc: point_t{
			Fx: 128,
			Fy: 136,
		},
		Fdata1: 5,
	},
	5: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 1,
		Floc: point_t{
			Fx: 128,
			Fy: 136,
		},
		Fdata1: 6,
	},
	6: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 1,
		Floc: point_t{
			Fx: 128,
			Fy: 136,
		},
		Fdata1: 7,
	},
	7: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 192,
			Fy: 144,
		},
		Fdata1: 8,
	},
	8: {
		Ftype1:  ANIM_LEVEL,
		Fperiod: TICRATE / 3,
		Fnanims: 1,
		Floc: point_t{
			Fx: 128,
			Fy: 136,
		},
		Fdata1: 8,
	},
}

var epsd2animinfo = [6]anim_t1{
	0: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 104,
			Fy: 168,
		},
	},
	1: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 40,
			Fy: 136,
		},
	},
	2: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 160,
			Fy: 96,
		},
	},
	3: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 104,
			Fy: 80,
		},
	},
	4: {
		Fperiod: TICRATE / 3,
		Fnanims: 3,
		Floc: point_t{
			Fx: 120,
			Fy: 32,
		},
	},
	5: {
		Fperiod: TICRATE / 4,
		Fnanims: 3,
		Floc: point_t{
			Fx: 40,
		},
	},
}

var NUMANIMS = [4]int32{
	0: int32(len(epsd0animinfo)),
	1: int32(len(epsd1animinfo)),
	2: int32(len(epsd2animinfo)),
}

var anims1 = [4][]anim_t1{
	0: epsd0animinfo[:],
	1: epsd1animinfo[:],
	2: epsd2animinfo[:],
}

//
// GENERAL DATA
//

//
// Locally used stuff.
//

// States for single-player

// in seconds
//#define SHOWLASTLOCDELAY	SHOWNEXTLOCDELAY

// C documentation
//
//	// used to accelerate or skip a stage
var acceleratestage int32

// C documentation
//
//	// wbs->pnum
var me int32

// C documentation
//
//	// specifies current state
var state stateenum_t

// C documentation
//
//	// contains information passed into intermission
var wbs *wbstartstruct_t

var plrs []wbplayerstruct_t // wbs->plyr[]

// C documentation
//
//	// used for general timing
var cnt int32

// C documentation
//
//	// used for timing of background animation
var bcnt int32

var cnt_kills [4]int32
var cnt_items [4]int32
var cnt_secret [4]int32
var cnt_time int32
var cnt_par int32
var cnt_pause int32

// C documentation
//
//	// # of commercial levels
var NUMCMAPS int32

//
//	GRAPHICS
//

// C documentation
//
//	// You Are Here graphic
var yah = [3]*patch_t{}

// C documentation
//
//	// splat
var splat = [2]*patch_t{}

// C documentation
//
//	// %, : graphics
var percent *patch_t
var colon *patch_t

// C documentation
//
//	// 0-9 graphic
var num [10]*patch_t

// C documentation
//
//	// minus sign
var wiminus *patch_t

// C documentation
//
//	// "Finished!" graphics
var finished *patch_t

// C documentation
//
//	// "Entering" graphic
var entering *patch_t

// C documentation
//
//	// "secret"
var sp_secret *patch_t

// C documentation
//
//	// "Kills", "Scrt", "Items", "Frags"
var kills *patch_t
var secret *patch_t
var items *patch_t
var frags *patch_t

// C documentation
//
//	// Time sucks.
var timepatch *patch_t
var par *patch_t
var sucks *patch_t

// C documentation
//
//	// "killers", "victims"
var killers *patch_t
var victims *patch_t

// C documentation
//
//	// "Total", your face, your dead face
var total *patch_t
var star *patch_t
var bstar *patch_t

// C documentation
//
//	// "red P[1..MAXPLAYERS]"
var p [4]*patch_t

// C documentation
//
//	// "gray P[1..MAXPLAYERS]"
var bp [4]*patch_t

// C documentation
//
//	// Name graphics of each level (centered)
var lnames []*patch_t

// C documentation
//
//	// Buffer storing the backdrop
var background *patch_t

//
// CODE
//

// C documentation
//
//	// slam background
func wi_slamBackground() {
	v_DrawPatch(0, 0, background)
}

// C documentation
//
//	// Draws "<Levelname> Finished!"
func wi_drawLF() {
	var y int32
	y = WI_TITLEY
	if gamemode != commercial || wbs.Flast < NUMCMAPS {
		// draw <LevelName>
		v_DrawPatch((SCREENWIDTH-int32(lnames[wbs.Flast].Fwidth))/2, y, lnames[wbs.Flast])
		// draw "Finished!"
		y += 5 * int32(lnames[wbs.Flast].Fheight) / 4
		v_DrawPatch((SCREENWIDTH-int32(finished.Fwidth))/2, y, finished)
	} else {
		if wbs.Flast == NUMCMAPS {
			// MAP33 - nothing is displayed!
		} else {
			if wbs.Flast > NUMCMAPS {
				// > MAP33.  Doom bombs out here with a Bad v_DrawPatch error.
				// I'm pretty sure that doom2.exe is just reading into random
				// bits of memory at this point, but let's try to be accurate
				// anyway.  This deliberately triggers a v_DrawPatch error.
				bp := patch_t{
					Fwidth:      int16(SCREENWIDTH),
					Fheight:     int16(SCREENHEIGHT),
					Fleftoffset: 1,
					Ftopoffset:  1,
				}
				v_DrawPatch(0, y, &bp)
			}
		}
	}
}

// C documentation
//
//	// Draws "Entering <LevelName>"
func wi_drawEL() {
	var y int32
	y = WI_TITLEY
	// draw "Entering"
	v_DrawPatch((SCREENWIDTH-int32(entering.Fwidth))/2, y, entering)
	// draw level
	y += 5 * int32(lnames[wbs.Fnext].Fheight) / 4
	v_DrawPatch((SCREENWIDTH-int32(lnames[wbs.Fnext].Fwidth))/2, y, lnames[wbs.Fnext])
}

func wi_drawOnLnode(n int32, c []*patch_t) {
	var bottom, i, left, right, top int32
	var fits boolean
	fits = 0
	i = 0
	for cond := true; cond; cond = fits == 0 && i != 2 && c[i] != nil {
		left = lnodes[wbs.Fepsd][n].Fx - int32(c[i].Fleftoffset)
		top = lnodes[wbs.Fepsd][n].Fy - int32(c[i].Ftopoffset)
		right = left + int32(c[i].Fwidth)
		bottom = top + int32(c[i].Fheight)
		if left >= 0 && right < SCREENWIDTH && top >= 0 && bottom < SCREENHEIGHT {
			fits = 1
		} else {
			i++
		}
	}
	if fits != 0 && i < 2 {
		v_DrawPatch(lnodes[wbs.Fepsd][n].Fx, lnodes[wbs.Fepsd][n].Fy, c[i])
	} else {
		// DEBUG
		fprintf_ccgo(os.Stdout, "Could not place patch on level %d", n+1)
	}
}

func wi_initAnimatedBack() {
	var a *anim_t1
	if gamemode == commercial {
		return
	}
	if wbs.Fepsd > 2 {
		return
	}
	for i := range NUMANIMS[wbs.Fepsd] {
		a = &anims1[wbs.Fepsd][i]
		// init variables
		a.Fctr = -1
		// specify the next time to draw it
		if a.Ftype1 == ANIM_ALWAYS {
			a.Fnexttic = bcnt + 1 + m_Random()%a.Fperiod
		} else {
			if a.Ftype1 == ANIM_RANDOM {
				a.Fnexttic = bcnt + 1 + a.Fdata2 + m_Random()%a.Fdata1
			} else {
				if a.Ftype1 == ANIM_LEVEL {
					a.Fnexttic = bcnt + 1
				}
			}
		}
	}
}

func wi_updateAnimatedBack() {
	var a *anim_t1
	if gamemode == commercial {
		return
	}
	if wbs.Fepsd > 2 {
		return
	}
	for i := range NUMANIMS[wbs.Fepsd] {
		a = &anims1[wbs.Fepsd][i]
		if bcnt == a.Fnexttic {
			switch a.Ftype1 {
			case ANIM_ALWAYS:
				a.Fctr++
				if a.Fctr >= a.Fnanims {
					a.Fctr = 0
				}
				a.Fnexttic = bcnt + a.Fperiod
			case ANIM_RANDOM:
				a.Fctr++
				if a.Fctr == a.Fnanims {
					a.Fctr = -1
					a.Fnexttic = bcnt + a.Fdata2 + m_Random()%a.Fdata1
				} else {
					a.Fnexttic = bcnt + a.Fperiod
				}
			case ANIM_LEVEL:
				// gawd-awful hack for level anims
				if !(state == StatCount && i == 7) && wbs.Fnext == a.Fdata1 {
					a.Fctr++
					if a.Fctr == a.Fnanims {
						a.Fctr--
					}
					a.Fnexttic = bcnt + a.Fperiod
				}
				break
			}
		}
	}
}

func wi_drawAnimatedBack() {
	var a *anim_t1
	if gamemode == commercial {
		return
	}
	if wbs.Fepsd > 2 {
		return
	}
	for i := range NUMANIMS[wbs.Fepsd] {
		a = &anims1[wbs.Fepsd][i]
		if a.Fctr >= 0 {
			v_DrawPatch(a.Floc.Fx, a.Floc.Fy, a.Fp[a.Fctr])
		}
	}
}

//
// Draws a number.
// If digits > 0, then use that many digits minimum,
//  otherwise only use as many as necessary.
// Returns new x position.
//

func wi_drawNum(x int32, y int32, n int32, digits int32) int32 {
	var fontwidth, neg, temp, v1 int32
	fontwidth = int32(num[0].Fwidth)
	if digits < 0 {
		if n == 0 {
			// make variable-length zeros 1 digit long
			digits = 1
		} else {
			// figure out # of digits in #
			digits = 0
			temp = n
			for temp != 0 {
				temp /= 10
				digits++
			}
		}
	}
	neg = boolint32(n < 0)
	if neg != 0 {
		n = -n
	}
	// if non-number, do not draw it
	if n == 1994 {
		return 0
	}
	// draw the new number
	for {
		v1 = digits
		digits--
		if v1 == 0 {
			break
		}
		x -= fontwidth
		v_DrawPatch(x, y, num[n%int32(10)])
		n /= 10
	}
	// draw a minus sign if necessary
	if neg != 0 {
		x -= 8
		v_DrawPatch(x, y, wiminus)
	}
	return x
}

func wi_drawPercent(x int32, y int32, p int32) {
	if p < 0 {
		return
	}
	v_DrawPatch(x, y, percent)
	wi_drawNum(x, y, p, -1)
}

// C documentation
//
//	//
//	// Display level completion time and par,
//	//  or "sucks" message if overflow.
//	//
func wi_drawTime(x int32, y int32, t int32) {
	var div, n int32
	if t < 0 {
		return
	}
	if t <= 61*59 {
		div = 1
		for cond := true; cond; cond = t/div != 0 {
			n = t / div % 60
			x = wi_drawNum(x, y, n, 2) - int32(colon.Fwidth)
			div *= 60
			// draw
			if div == 60 || t/div != 0 {
				v_DrawPatch(x, y, colon)
			}
		}
	} else {
		// "sucks"
		v_DrawPatch(x-int32(sucks.Fwidth), y, sucks)
	}
}

func wi_End() {
	wi_unloadData()
}

func wi_initNoState() {
	state = NoState
	acceleratestage = 0
	cnt = 10
}

func wi_updateNoState() {
	var v1 int32
	wi_updateAnimatedBack()
	cnt--
	v1 = cnt
	if v1 == 0 {
		// Don't call wi_End yet.  g_WorldDone doesnt immediately
		// change gamestate, so wi_Drawer is still going to get
		// run until that happens.  If we do that after WI_End
		// (which unloads all the graphics), we're in trouble.
		//wi_End();
		g_WorldDone()
	}
}

var snl_pointeron uint32 = 0

func wi_initShowNextLoc() {
	state = ShowNextLoc
	acceleratestage = 0
	cnt = SHOWNEXTLOCDELAY * TICRATE
	wi_initAnimatedBack()
}

func wi_updateShowNextLoc() {
	var v1 int32
	wi_updateAnimatedBack()
	cnt--
	v1 = cnt
	if v1 == 0 || acceleratestage != 0 {
		wi_initNoState()
	} else {
		snl_pointeron = booluint32(cnt&31 < 20)
	}
}

func wi_drawShowNextLoc() {
	var last, v1 int32
	wi_slamBackground()
	// draw animated background
	wi_drawAnimatedBack()
	if gamemode != commercial {
		if wbs.Fepsd > 2 {
			wi_drawEL()
			return
		}
		if wbs.Flast == 8 {
			v1 = wbs.Fnext - 1
		} else {
			v1 = wbs.Flast
		}
		last = v1
		// draw a splat on taken cities.
		for i := int32(0); i <= last; i++ {
			wi_drawOnLnode(i, splat[:])
		}
		// splat the secret level?
		if wbs.Fdidsecret != 0 {
			wi_drawOnLnode(8, splat[:])
		}
		// draw flashing ptr
		if snl_pointeron != 0 {
			wi_drawOnLnode(wbs.Fnext, yah[:])
		}
	}
	// draws which level you are entering..
	if gamemode != commercial || wbs.Fnext != 30 {
		wi_drawEL()
	}
}

func wi_drawNoState() {
	snl_pointeron = 1
	wi_drawShowNextLoc()
}

func wi_fragSum(playernum int32) int32 {
	var frags int32
	frags = 0
	for i := range int32(MAXPLAYERS) {
		if playeringame[i] != 0 && i != playernum {
			frags += plrs[playernum].Ffrags[i]
		}
	}
	// JDC hack - negative frags.
	frags -= plrs[playernum].Ffrags[playernum]
	// UNUSED if (frags < 0)
	// 	frags = 0;
	return frags
}

var dm_state int32
var dm_frags [4][4]int32
var dm_totals [4]int32

func wi_initDeathmatchStats() {
	state = StatCount
	acceleratestage = 0
	dm_state = 1
	cnt_pause = TICRATE
	for i := range MAXPLAYERS {
		if playeringame[i] != 0 {
			for j := range MAXPLAYERS {
				if playeringame[j] != 0 {
					dm_frags[i][j] = 0
				}
			}
			dm_totals[i] = 0
		}
	}
	wi_initAnimatedBack()
}

func wi_updateDeathmatchStats() {
	var v5 int32
	var stillticking boolean
	wi_updateAnimatedBack()
	if acceleratestage != 0 && dm_state != 4 {
		acceleratestage = 0
		for i := range int32(MAXPLAYERS) {
			if playeringame[i] != 0 {
				for j := range MAXPLAYERS {
					if playeringame[j] != 0 {
						dm_frags[i][j] = plrs[i].Ffrags[j]
					}
				}
				dm_totals[i] = wi_fragSum(i)
			}
		}
		s_StartSound(nil, int32(sfx_barexp))
		dm_state = 4
	}
	if dm_state == 2 {
		if bcnt&3 == 0 {
			s_StartSound(nil, int32(sfx_pistol))
		}
		stillticking = 0
		for i := range int32(MAXPLAYERS) {
			if playeringame[i] != 0 {
				for j := range MAXPLAYERS {
					if playeringame[j] != 0 && dm_frags[i][j] != plrs[i].Ffrags[j] {
						if plrs[i].Ffrags[j] < 0 {
							dm_frags[i][j]--
						} else {
							dm_frags[i][j]++
						}
						if dm_frags[i][j] > 99 {
							dm_frags[i][j] = 99
						}
						if dm_frags[i][j] < -99 {
							dm_frags[i][j] = -99
						}
						stillticking = 1
					}
				}
				dm_totals[i] = wi_fragSum(i)
				if dm_totals[i] > 99 {
					dm_totals[i] = 99
				}
				if dm_totals[i] < -99 {
					dm_totals[i] = -99
				}
			}
		}
		if stillticking == 0 {
			s_StartSound(nil, int32(sfx_barexp))
			dm_state++
		}
	} else {
		if dm_state == 4 {
			if acceleratestage != 0 {
				s_StartSound(nil, int32(sfx_slop))
				if gamemode == commercial {
					wi_initNoState()
				} else {
					wi_initShowNextLoc()
				}
			}
		} else {
			if dm_state&1 != 0 {
				cnt_pause--
				v5 = cnt_pause
				if v5 == 0 {
					dm_state++
					cnt_pause = TICRATE
				}
			}
		}
	}
}

func wi_drawDeathmatchStats() {
	var w, x, y int32
	wi_slamBackground()
	// draw animated background
	wi_drawAnimatedBack()
	wi_drawLF()
	// draw stat titles (top line)
	v_DrawPatch(DM_TOTALSX-int32(total.Fwidth)/2, DM_MATRIXY-WI_SPACINGY+10, total)
	v_DrawPatch(DM_KILLERSX, DM_KILLERSY, killers)
	v_DrawPatch(DM_VICTIMSX, DM_VICTIMSY, victims)
	// draw P?
	x = DM_MATRIXX + DM_SPACINGX
	y = DM_MATRIXY
	for i := range int32(MAXPLAYERS) {
		if playeringame[i] != 0 {
			v_DrawPatch(x-int32(p[i].Fwidth)/2, DM_MATRIXY-WI_SPACINGY, p[i])
			v_DrawPatch(DM_MATRIXX-int32(p[i].Fwidth)/2, y, p[i])
			if i == me {
				v_DrawPatch(x-int32(p[i].Fwidth)/2, DM_MATRIXY-WI_SPACINGY, bstar)
				v_DrawPatch(DM_MATRIXX-int32(p[i].Fwidth)/2, y, star)
			}
		} else {
			// v_DrawPatch(x-SHORT(bp[i]->width)/2,
			//   DM_MATRIXY - WI_SPACINGY, bp[i]);
			// v_DrawPatch(DM_MATRIXX-SHORT(bp[i]->width)/2,
			//   y, bp[i]);
		}
		x += DM_SPACINGX
		y += WI_SPACINGY
	}
	// draw stats
	y = DM_MATRIXY + 10
	w = int32(num[0].Fwidth)
	for i := range MAXPLAYERS {
		x = DM_MATRIXX + DM_SPACINGX
		if playeringame[i] != 0 {
			for j := range MAXPLAYERS {
				if playeringame[j] != 0 {
					wi_drawNum(x+w, y, dm_frags[i][j], 2)
				}
				x += DM_SPACINGX
			}
			wi_drawNum(DM_TOTALSX+w, y, dm_totals[i], 2)
		}
		y += WI_SPACINGY
	}
}

var cnt_frags [4]int32
var dofrags int32
var ng_state int32

func wi_initNetgameStats() {
	var v2, v3, v4 int32
	state = StatCount
	acceleratestage = 0
	ng_state = 1
	cnt_pause = TICRATE
	for i := range int32(MAXPLAYERS) {
		if playeringame[i] == 0 {
			continue
		}
		v4 = 0
		cnt_frags[i] = v4
		v3 = v4
		cnt_secret[i] = v3
		v2 = v3
		cnt_items[i] = v2
		cnt_kills[i] = v2
		dofrags += wi_fragSum(i)
	}
	dofrags = boolint32(dofrags != 0)
	wi_initAnimatedBack()
}

func wi_updateNetgameStats() {
	var fsum, v6, v7 int32
	var stillticking boolean
	wi_updateAnimatedBack()
	if acceleratestage != 0 && ng_state != 10 {
		acceleratestage = 0
		for i := range int32(MAXPLAYERS) {
			if playeringame[i] == 0 {
				continue
			}
			cnt_kills[i] = plrs[i].Fskills * 100 / wbs.Fmaxkills
			cnt_items[i] = plrs[i].Fsitems * 100 / wbs.Fmaxitems
			cnt_secret[i] = plrs[i].Fssecret * 100 / wbs.Fmaxsecret
			if dofrags != 0 {
				cnt_frags[i] = wi_fragSum(i)
			}
		}
		s_StartSound(nil, int32(sfx_barexp))
		ng_state = 10
	}
	if ng_state == 2 {
		if bcnt&3 == 0 {
			s_StartSound(nil, int32(sfx_pistol))
		}
		stillticking = 0
		for i := range MAXPLAYERS {
			if playeringame[i] == 0 {
				continue
			}
			cnt_kills[i] += 2
			if cnt_kills[i] >= plrs[i].Fskills*int32(100)/wbs.Fmaxkills {
				cnt_kills[i] = plrs[i].Fskills * 100 / wbs.Fmaxkills
			} else {
				stillticking = 1
			}
		}
		if stillticking == 0 {
			s_StartSound(nil, int32(sfx_barexp))
			ng_state++
		}
	} else {
		if ng_state == 4 {
			if bcnt&3 == 0 {
				s_StartSound(nil, int32(sfx_pistol))
			}
			stillticking = 0
			for i := range MAXPLAYERS {
				if playeringame[i] == 0 {
					continue
				}
				cnt_items[i] += 2
				if cnt_items[i] >= plrs[i].Fsitems*int32(100)/wbs.Fmaxitems {
					cnt_items[i] = plrs[i].Fsitems * 100 / wbs.Fmaxitems
				} else {
					stillticking = 1
				}
			}
			if stillticking == 0 {
				s_StartSound(nil, int32(sfx_barexp))
				ng_state++
			}
		} else {
			if ng_state == 6 {
				if bcnt&3 == 0 {
					s_StartSound(nil, int32(sfx_pistol))
				}
				stillticking = 0
				for i := range MAXPLAYERS {
					if playeringame[i] == 0 {
						continue
					}
					cnt_secret[i] += 2
					if cnt_secret[i] >= plrs[i].Fssecret*int32(100)/wbs.Fmaxsecret {
						cnt_secret[i] = plrs[i].Fssecret * 100 / wbs.Fmaxsecret
					} else {
						stillticking = 1
					}
				}
				if stillticking == 0 {
					s_StartSound(nil, int32(sfx_barexp))
					ng_state += 1 + 2*boolint32(dofrags == 0)
				}
			} else {
				if ng_state == 8 {
					if bcnt&3 == 0 {
						s_StartSound(nil, int32(sfx_pistol))
					}
					stillticking = 0
					for i := range int32(MAXPLAYERS) {
						if playeringame[i] == 0 {
							continue
						}
						cnt_frags[i] += 1
						v6 = wi_fragSum(i)
						fsum = v6
						if cnt_frags[i] >= v6 {
							cnt_frags[i] = fsum
						} else {
							stillticking = 1
						}
					}
					if stillticking == 0 {
						s_StartSound(nil, int32(sfx_pldeth))
						ng_state++
					}
				} else {
					if ng_state == 10 {
						if acceleratestage != 0 {
							s_StartSound(nil, int32(sfx_sgcock))
							if gamemode == commercial {
								wi_initNoState()
							} else {
								wi_initShowNextLoc()
							}
						}
					} else {
						if ng_state&1 != 0 {
							cnt_pause--
							v7 = cnt_pause
							if v7 == 0 {
								ng_state++
								cnt_pause = TICRATE
							}
						}
					}
				}
			}
		}
	}
}

func wi_drawNetgameStats() {
	var pwidth, x, y int32
	pwidth = int32(percent.Fwidth)
	wi_slamBackground()
	// draw animated background
	wi_drawAnimatedBack()
	wi_drawLF()
	// draw stat titles (top line)
	v_DrawPatch(32+int32(star.Fwidth)/2+int32(32)*boolint32(dofrags == 0)+NG_SPACINGX-int32(kills.Fwidth), NG_STATSY, kills)
	v_DrawPatch(32+int32(star.Fwidth)/2+int32(32)*boolint32(dofrags == 0)+2*NG_SPACINGX-int32(items.Fwidth), NG_STATSY, items)
	v_DrawPatch(32+int32(star.Fwidth)/2+int32(32)*boolint32(dofrags == 0)+3*NG_SPACINGX-int32(secret.Fwidth), NG_STATSY, secret)
	if dofrags != 0 {
		v_DrawPatch(32+int32(star.Fwidth)/2+int32(32)*boolint32(dofrags == 0)+4*NG_SPACINGX-int32(frags.Fwidth), NG_STATSY, frags)
	}
	// draw stats
	y = NG_STATSY + int32(kills.Fheight)
	for i := range int32(MAXPLAYERS) {
		if playeringame[i] == 0 {
			continue
		}
		x = 32 + int32(star.Fwidth)/2 + 32*boolint32(dofrags == 0)
		v_DrawPatch(x-int32(p[i].Fwidth), y, p[i])
		if i == me {
			v_DrawPatch(x-int32(p[i].Fwidth), y, star)
		}
		x += NG_SPACINGX
		wi_drawPercent(x-pwidth, y+int32(10), cnt_kills[i])
		x += NG_SPACINGX
		wi_drawPercent(x-pwidth, y+int32(10), cnt_items[i])
		x += NG_SPACINGX
		wi_drawPercent(x-pwidth, y+int32(10), cnt_secret[i])
		x += NG_SPACINGX
		if dofrags != 0 {
			wi_drawNum(x, y+int32(10), cnt_frags[i], -1)
		}
		y += WI_SPACINGY
	}
}

var sp_state int32

func wi_initStats() {
	var v1, v2, v3 int32
	state = StatCount
	acceleratestage = 0
	sp_state = 1
	v2 = -1
	cnt_secret[0] = v2
	v1 = v2
	cnt_items[0] = v1
	cnt_kills[0] = v1
	v3 = -1
	cnt_par = v3
	cnt_time = v3
	cnt_pause = TICRATE
	wi_initAnimatedBack()
}

func wi_updateStats() {
	var v1 int32
	wi_updateAnimatedBack()
	if acceleratestage != 0 && sp_state != 10 {
		acceleratestage = 0
		cnt_kills[0] = plrs[me].Fskills * 100 / wbs.Fmaxkills
		cnt_items[0] = plrs[me].Fsitems * 100 / wbs.Fmaxitems
		cnt_secret[0] = plrs[me].Fssecret * 100 / wbs.Fmaxsecret
		cnt_time = plrs[me].Fstime / TICRATE
		cnt_par = wbs.Fpartime / TICRATE
		s_StartSound(nil, int32(sfx_barexp))
		sp_state = 10
	}
	if sp_state == 2 {
		cnt_kills[0] += 2
		if bcnt&3 == 0 {
			s_StartSound(nil, int32(sfx_pistol))
		}
		if cnt_kills[0] >= plrs[me].Fskills*int32(100)/wbs.Fmaxkills {
			cnt_kills[0] = plrs[me].Fskills * 100 / wbs.Fmaxkills
			s_StartSound(nil, int32(sfx_barexp))
			sp_state++
		}
	} else {
		if sp_state == 4 {
			cnt_items[0] += 2
			if bcnt&3 == 0 {
				s_StartSound(nil, int32(sfx_pistol))
			}
			if cnt_items[0] >= plrs[me].Fsitems*int32(100)/wbs.Fmaxitems {
				cnt_items[0] = plrs[me].Fsitems * 100 / wbs.Fmaxitems
				s_StartSound(nil, int32(sfx_barexp))
				sp_state++
			}
		} else {
			if sp_state == 6 {
				cnt_secret[0] += 2
				if bcnt&3 == 0 {
					s_StartSound(nil, int32(sfx_pistol))
				}
				if cnt_secret[0] >= plrs[me].Fssecret*int32(100)/wbs.Fmaxsecret {
					cnt_secret[0] = plrs[me].Fssecret * 100 / wbs.Fmaxsecret
					s_StartSound(nil, int32(sfx_barexp))
					sp_state++
				}
			} else {
				if sp_state == 8 {
					if bcnt&3 == 0 {
						s_StartSound(nil, int32(sfx_pistol))
					}
					cnt_time += 3
					if cnt_time >= plrs[me].Fstime/TICRATE {
						cnt_time = plrs[me].Fstime / TICRATE
					}
					cnt_par += 3
					if cnt_par >= wbs.Fpartime/TICRATE {
						cnt_par = wbs.Fpartime / TICRATE
						if cnt_time >= plrs[me].Fstime/TICRATE {
							s_StartSound(nil, int32(sfx_barexp))
							sp_state++
						}
					}
				} else {
					if sp_state == 10 {
						if acceleratestage != 0 {
							s_StartSound(nil, int32(sfx_sgcock))
							if gamemode == commercial {
								wi_initNoState()
							} else {
								wi_initShowNextLoc()
							}
						}
					} else {
						if sp_state&1 != 0 {
							cnt_pause--
							v1 = cnt_pause
							if v1 == 0 {
								sp_state++
								cnt_pause = TICRATE
							}
						}
					}
				}
			}
		}
	}
}

func wi_drawStats() {
	var lh int32
	lh = 3 * int32(num[0].Fheight) / 2
	wi_slamBackground()
	// draw animated background
	wi_drawAnimatedBack()
	wi_drawLF()
	v_DrawPatch(SP_STATSX, SP_STATSY, kills)
	wi_drawPercent(SCREENWIDTH-SP_STATSX, SP_STATSY, cnt_kills[0])
	v_DrawPatch(SP_STATSX, SP_STATSY+lh, items)
	wi_drawPercent(SCREENWIDTH-SP_STATSX, SP_STATSY+lh, cnt_items[0])
	v_DrawPatch(SP_STATSX, SP_STATSY+2*lh, sp_secret)
	wi_drawPercent(SCREENWIDTH-SP_STATSX, SP_STATSY+2*lh, cnt_secret[0])
	v_DrawPatch(SP_TIMEX, SCREENHEIGHT-32, timepatch)
	wi_drawTime(SCREENWIDTH/2-SP_TIMEX, SCREENHEIGHT-32, cnt_time)
	if wbs.Fepsd < 3 {
		v_DrawPatch(SCREENWIDTH/2+SP_TIMEX, SCREENHEIGHT-32, par)
		wi_drawTime(SCREENWIDTH-SP_TIMEX, SCREENHEIGHT-32, cnt_par)
	}
}

func wi_checkForAccelerate() {
	// check for button presses to skip delays
	for i := 0; i < MAXPLAYERS; i++ {
		if playeringame[i] != 0 {
			player := &players[i]
			if int32(player.Fcmd.Fbuttons)&bt_ATTACK != 0 {
				if player.Fattackdown == 0 {
					acceleratestage = 1
				}
				player.Fattackdown = 1
			} else {
				player.Fattackdown = 0
			}
			if int32(player.Fcmd.Fbuttons)&bt_USE != 0 {
				if player.Fusedown == 0 {
					acceleratestage = 1
				}
				player.Fusedown = 1
			} else {
				player.Fusedown = 0
			}
		}
	}
}

// C documentation
//
//	// Updates stuff each tick
func wi_Ticker() {
	// counter for general background animation
	bcnt++
	if bcnt == 1 {
		// intermission music
		if gamemode == commercial {
			s_ChangeMusic(int32(mus_dm2int), 1)
		} else {
			s_ChangeMusic(int32(mus_inter), 1)
		}
	}
	wi_checkForAccelerate()
	switch state {
	case StatCount:
		if deathmatch != 0 {
			wi_updateDeathmatchStats()
		} else {
			if netgame != 0 {
				wi_updateNetgameStats()
			} else {
				wi_updateStats()
			}
		}
	case ShowNextLoc:
		wi_updateShowNextLoc()
	case NoState:
		wi_updateNoState()
		break
	}
}

// Common load/unload function.  Iterates over all the graphics
// lumps to be loaded/unloaded into memory.

func wi_loadUnloadData(callback func(string, **patch_t)) {
	if gamemode == commercial {
		for i := range NUMCMAPS {
			bp1 := fmt.Sprintf("CWILV%2.2d", i)
			callback(bp1, &lnames[i])
		}
	} else {
		for i := range NUMMAPS {
			bp1 := fmt.Sprintf("WILV%d%d", wbs.Fepsd, i)
			callback(bp1, &lnames[i])
		}
		// you are here
		callback("WIURH0", &yah[0])
		// you are here (alt.)
		callback("WIURH1", &yah[1])
		// splat
		callback("WISPLAT", &splat[0])
		if wbs.Fepsd < 3 {
			for j := range NUMANIMS[wbs.Fepsd] {
				a := &anims1[wbs.Fepsd][j]
				for i := range a.Fnanims {
					// MONDO HACK!
					if wbs.Fepsd != 1 || j != 8 {
						// animations
						bp1 := fmt.Sprintf("WIA%d%.2d%.2d", wbs.Fepsd, j, i)
						callback(bp1, &a.Fp[i])
					} else {
						// HACK ALERT!
						a.Fp[i] = anims1[1][4].Fp[i]
					}
				}
			}
		}
	}
	// More hacks on minus sign.
	callback("WIMINUS", &wiminus)
	for i := range 10 {
		// numbers 0-9
		bp1 := fmt.Sprintf("WINUM%d", i)
		callback(bp1, &num[i])
	}
	// percent sign
	callback("WIPCNT", &percent)
	// "finished"
	callback("WIF", &finished)
	// "entering"
	callback("WIENTER", &entering)
	// "kills"
	callback("WIOSTK", &kills)
	// "scrt"
	callback("WIOSTS", &secret)
	// "secret"
	callback("WISCRT2", &sp_secret)
	// french wad uses WIOBJ (?)
	if w_CheckNumForName("WIOBJ") >= 0 {
		// "items"
		if netgame != 0 && deathmatch == 0 {
			callback("WIOBJ", &items)
		} else {
			callback("WIOSTI", &items)
		}
	} else {
		callback("WIOSTI", &items)
	}
	// "frgs"
	callback("WIFRGS", &frags)
	// ":"
	callback("WICOLON", &colon)
	// "time"
	callback("WITIME", &timepatch)
	// "sucks"
	callback("WISUCKS", &sucks)
	// "par"
	callback("WIPAR", &par)
	// "killers" (vertical)
	callback("WIKILRS", &killers)
	// "victims" (horiz)
	callback("WIVCTMS", &victims)
	// "total"
	callback("WIMSTT", &total)
	for i := range MAXPLAYERS {
		// "1,2,3,4"
		bp1 := fmt.Sprintf("STPB%d", i)
		callback(bp1, &p[i])
		// "1,2,3,4"
		bp1 = fmt.Sprintf("WIBP%d", i+1)
		callback(bp1, &bp[i])
	}
	// Background image
	var bp1 string
	if gamemode == commercial {
		bp1 = "INTERPIC"
	} else {
		if gamemode == retail && wbs.Fepsd == 3 {
			bp1 = "INTERPIC"
		} else {
			bp1 = fmt.Sprintf("WIMAP%d", wbs.Fepsd)
		}
	}
	// Draw backdrop and save to a temporary buffer
	callback(bp1, &background)
}

func wi_loadCallback(name string, variable **patch_t) {
	*variable = w_CacheLumpNameT(name)
}

func wi_loadData() {
	if gamemode == commercial {
		NUMCMAPS = 32
		lnames = make([]*patch_t, NUMCMAPS)
	} else {
		lnames = make([]*patch_t, NUMMAPS)
	}
	wi_loadUnloadData(wi_loadCallback)
	// These two graphics are special cased because we're sharing
	// them with the status bar code
	// your face
	star = w_CacheLumpNameT("STFST01")
	// dead face
	bstar = w_CacheLumpNameT("STFDEAD0")
}

func wi_unloadCallback(name string, variable **patch_t) {
	w_ReleaseLumpName(name)
	*variable = nil
}

func wi_unloadData() {
	wi_loadUnloadData(wi_unloadCallback)
	// We do not free these lumps as they are shared with the status
	// bar code.
	// w_ReleaseLumpName("STFST01");
	// w_ReleaseLumpName("STFDEAD0");
}

func wi_Drawer() {
	switch state {
	case StatCount:
		if deathmatch != 0 {
			wi_drawDeathmatchStats()
		} else {
			if netgame != 0 {
				wi_drawNetgameStats()
			} else {
				wi_drawStats()
			}
		}
	case ShowNextLoc:
		wi_drawShowNextLoc()
	case NoState:
		wi_drawNoState()
		break
	}
}

func wi_initVariables(wbstartstruct *wbstartstruct_t) {
	var v1 int32
	wbs = wbstartstruct
	acceleratestage = 0
	v1 = 0
	bcnt = v1
	cnt = v1
	me = wbs.Fpnum
	plrs = wbs.Fplyr[:]
	if wbs.Fmaxkills == 0 {
		wbs.Fmaxkills = 1
	}
	if wbs.Fmaxitems == 0 {
		wbs.Fmaxitems = 1
	}
	if wbs.Fmaxsecret == 0 {
		wbs.Fmaxsecret = 1
	}
	if gamemode != retail {
		if wbs.Fepsd > 2 {
			wbs.Fepsd -= 3
		}
	}
}

func wi_Start(wbstartstruct *wbstartstruct_t) {
	wi_initVariables(wbstartstruct)
	wi_loadData()
	if deathmatch != 0 {
		wi_initDeathmatchStats()
	} else {
		if netgame != 0 {
			wi_initNetgameStats()
		} else {
			wi_initStats()
		}
	}
}

var open_wadfiles []*os.File

func getFileNumber(handle *os.File) int32 {
	for i := 0; i < len(open_wadfiles); i++ {
		if open_wadfiles[i] == handle {
			return int32(i)
		}
	}

	open_wadfiles = append(open_wadfiles, handle)
	return int32(len(open_wadfiles) - 1)
}

func checksumAddLump(sha hash.Hash, lump *lumpinfo_t) {
	sha1_UpdateString(sha, lump.Name())
	sha1_UpdateInt32(sha, uint32(getFileNumber(lump.Fwad_file)))
	sha1_UpdateInt32(sha, uint32(lump.Fposition))
	sha1_UpdateInt32(sha, uint32(lump.Fsize))
}

func w_Checksum(digest *sha1_digest_t) {
	sha := sha1.New()
	open_wadfiles = nil
	// Go through each entry in the WAD directory, adding information
	// about each entry to the SHA1 hash.
	for i := range numlumps {
		checksumAddLump(sha, &lumpinfo[i])
	}
	copy(digest[:], sha.Sum(nil))
}

func w_OpenFile(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error opening file %q: %v", path, err)
		return nil
	}
	return f
}

func w_Read(wad *os.File, offset uint32, buffer uintptr, buffer_len uint64) uint64 {
	buf := unsafe.Slice((*byte)(unsafe.Pointer(buffer)), buffer_len)
	n, err := wad.ReadAt(buf, int64(offset))
	if err != nil {
		log.Printf("Error reading from file: %v", err)
	}
	return uint64(n)
}

//
// This is used to get the local FILE:LINE info from CPP
// prior to really call the function in question.
//

// Parse the command line, merging WAD files that are sppecified.
// Returns true if at least one file was added.

func w_ParseCommandLine() boolean {
	var filename string
	var modifiedgame boolean
	var p, v1 int32
	modifiedgame = 0
	//!
	// @arg <files>
	// @vanilla
	//
	// Load the specified PWAD files.
	//
	p = m_CheckParmWithArgs("-file", 1)
	if p != 0 {
		// the parms after p are wadfile/lump names,
		// until end of parms or another - preceded parm
		modifiedgame = 1 // homebrew levels
		for {
			p++
			v1 = p
			if !(v1 != int32(len(myargs)) && myargs[p][0] != '-') {
				break
			}
			filename = d_TryFindWADByName(myargs[p])
			fprintf_ccgo(os.Stdout, " adding %s\n", filename)
			w_AddFile(filename)
		}
	}
	//    W_PrintDirectory();
	return modifiedgame
}

type wadinfo_t struct {
	Fidentification [4]byte
	Fnumlumps       int32
	Finfotableofs   int32
}

type filelump_t struct {
	Ffilepos int32
	Fsize    int32
	Fname    [8]byte
}

// Hash table for fast lookups

var lumphash []*lumpinfo_t

// Hash function used for lump names.

func w_LumpNameHash(s string) uint32 {
	// This is the djb2 string hash function, modded to work on strings
	// that have a maximum length of 8.
	result := uint32(5381)
	for i := 0; i < len(s) && i < 8; i++ {
		result = result<<5 ^ result ^ uint32(xtoupper(int32(s[i])))
	}
	return result
}

// C documentation
//
//	// Increase the size of the lumpinfo[] array to the specified size.
func extendLumpInfo(newnumlumps int32) {
	if newnumlumps >= int32(len(lumpinfo)) {
		// TODO: Should be lumpinfo = append(lumpinfo, lumpinfo_t{})
		panic("extendLumpInfo called with newnumlumps >= len(lumpinfo)")
	}

	numlumps = uint32(newnumlumps)
}

// LUMP BASED ROUTINES.
//
// Stop go garbage collecting these
var wad_files = map[uintptr]*os.File{}

//
// W_AddFile
// All files are optional, but at least one file must be
//  found (PWAD, if all required lumps are present).
// Files with a .wad extension are wadlink files
//  with multiple lumps.
// Other files are single lumps with the base filename
//  for the lump name.

func w_AddFile(filename string) *os.File {
	var fileinfo []filelump_t
	var wad_file *os.File
	var length, newnumlumps int32
	var startlump uint32
	// open the file and add to directory
	stat, err := os.Stat(filename)
	if err != nil {
		log.Printf("Error stating file %q: %v", filename, err)
		return nil
	}
	wad_file = w_OpenFile(filename)
	if wad_file == nil {
		fprintf_ccgo(os.Stdout, " couldn't open %s\n", filename)
		return nil
	}
	newnumlumps = int32(numlumps)
	if !strings.EqualFold(filepath.Ext(filename), ".wad") {
		// single lump file
		// fraggle: Swap the filepos and size here.  The WAD directory
		// parsing code expects a little-endian directory, so will swap
		// them back.  Effectively we're constructing a "fake WAD directory"
		// here, as it would appear on disk.
		fileinfo = make([]filelump_t, 1)
		fileinfo[0].Ffilepos = 0
		fileinfo[0].Fsize = int32(stat.Size())
		// Name the lump after the base of the filename (without the
		// extension).
		m_ExtractFileBase(filename, fileinfo[0].Fname[:])
		newnumlumps++
	} else {
		var wadinfo wadinfo_t
		// WAD file
		w_Read(wad_file, 0, (uintptr)(unsafe.Pointer(&wadinfo)), 12)
		if gostring_bytes(wadinfo.Fidentification[:]) != "IWAD" {
			// Homebrew levels?
			if gostring_bytes(wadinfo.Fidentification[:]) != "PWAD" {
				i_Error("Wad file %s doesn't have IWAD or PWAD id\n", filename)
			}
			// ???modifiedgame = true;
		}
		wadinfo.Fnumlumps = wadinfo.Fnumlumps
		wadinfo.Finfotableofs = wadinfo.Finfotableofs
		length = int32(uint64(wadinfo.Fnumlumps) * 16)
		fileinfo = make([]filelump_t, wadinfo.Fnumlumps)
		w_Read(wad_file, uint32(wadinfo.Finfotableofs), (uintptr)(unsafe.Pointer(&fileinfo[0])), uint64(length))
		newnumlumps += wadinfo.Fnumlumps
	}
	// Increase size of numlumps array to accomodate the new file.
	startlump = numlumps
	extendLumpInfo(newnumlumps)
	for i := startlump; i < numlumps; i++ {
		lump_p := &lumpinfo[i]
		lump_p.Fwad_file = wad_file
		lump_p.Fposition = fileinfo[i].Ffilepos
		lump_p.Fsize = fileinfo[i].Fsize
		lump_p.Fcache = nil
		lump_p.Fname = fileinfo[i].Fname
	}
	lumphash = nil
	wad_files[wad_file.Fd()] = wad_file
	return wad_file
}

//
// W_CheckNumForName
// Returns -1 if name not found.
//

func w_CheckNumForName(name string) int32 {
	// Do we have a hash table yet?
	if lumphash != nil {
		// We do! Excellent.
		hash := int32(w_LumpNameHash(name) % numlumps)
		for lump_p := lumphash[hash]; lump_p != nil; lump_p = lump_p.Fnext {
			if strings.EqualFold(lump_p.Name(), name) {
				return lumpIndex(lump_p)
			}
		}
	} else {
		// We don't have a hash table generate yet. Linear search :-(
		//
		// scan backwards so patch lump files take precedence
		for i := int32(numlumps - 1); i >= 0; i-- {
			if strings.EqualFold(lumpinfo[i].Name(), name) {
				return i
			}
		}
	}
	// TFB. Not found.
	return -1
}

// C documentation
//
//	//
//	// W_GetNumForName
//	// Calls w_CheckNumForName, but bombs out if not found.
//	//
func w_GetNumForName(name string) int32 {
	var i int32
	i = w_CheckNumForName(name)
	if i < 0 {
		i_Error("w_GetNumForName: %s not found!", name)
	}
	return i
}

// C documentation
//
//	//
//	// W_LumpLength
//	// Returns the buffer size needed to load the given lump.
//	//
func w_LumpLength(lump uint32) int32 {
	if lump >= numlumps {
		i_Error("w_LumpLength: %d >= numlumps", lump)
	}
	return lumpinfo[lump].Fsize
}

func w_ReadLumpBytes(lump uint32) []byte {
	if lump >= numlumps {
		i_Error("w_ReadLumpBytes: %d >= numlumps", lump)
	}
	l := &lumpinfo[lump]
	res := make([]byte, l.Fsize)
	if n, err := l.Fwad_file.ReadAt(res, int64(l.Fposition)); err != nil {
		log.Fatalf("w_ReadLumpBytes: error reading lump %d (%dB at %d): %v", lump, l.Fsize, l.Fposition, err)
	} else if n < int(l.Fsize) {
		log.Fatalf("w_ReadLumpBytes: only read %d of %d on lump %d", n, l.Fsize, lump)
	}
	return res
}

//
// w_CacheLumpNum
//
// Load a lump into memory and return a pointer to a buffer containing
// the lump data.
//

func w_CacheLumpNum(lumpnum int32) uintptr {
	if uint32(lumpnum) >= numlumps {
		i_Error("w_CacheLumpNum: %d >= numlumps", lumpnum)
	}
	lump := &lumpinfo[lumpnum]
	if lump.Fcache == nil {
		lump.Fcache = w_ReadLumpBytes(uint32(lumpnum))
	}
	return (uintptr)(unsafe.Pointer(&lump.Fcache[0]))
}
func w_CacheLumpNumBytes(lumpnum int32) []byte {
	if uint32(lumpnum) >= numlumps {
		i_Error("w_CacheLumpNum: %d >= numlumps", lumpnum)
	}
	lump := &lumpinfo[lumpnum]
	if lump.Fcache == nil {
		lump.Fcache = w_ReadLumpBytes(uint32(lumpnum))
	}
	return lump.Fcache
}
func w_CacheLumpNumT[T lumpType](lumpnum int32) T {
	var result uintptr
	result = w_CacheLumpNum(lumpnum)
	if result == 0 {
		panic("lump failure")
	}
	return (T)(unsafe.Pointer(result))
}

// C documentation
//
//	//
//	// W_CacheLumpName
//	//
func w_CacheLumpName(name string) uintptr {
	return w_CacheLumpNum(w_GetNumForName(name))
}

func w_CacheLumpNameBytes(name string) []byte {
	return w_CacheLumpNumBytes(w_GetNumForName(name))
}

func w_CacheLumpNameT[T lumpType](name string) T {
	var result uintptr
	result = w_CacheLumpName(name)
	if result == 0 {
		panic("lump failure")
	}
	return (T)(unsafe.Pointer(result))
}

//
// Release a lump back to the cache, so that it can be reused later
// without having to read from disk again, or alternatively, discarded
// if we run out of memory.
//
// Back in Vanilla Doom, this was just done using Z_ChangeTag
// directly, but now that we have WAD mmap, things are a bit more
// complicated ...
//

func w_ReleaseLumpNum(lumpnum int32) {
	if uint32(lumpnum) >= numlumps {
		i_Error("w_ReleaseLumpNum: %d >= numlumps", lumpnum)
	}
	// TODO/GORE: We don't do anything here. Lumps are just progressively cached & never released. It's a finite number
}

func w_ReleaseLumpName(name string) {
	w_ReleaseLumpNum(w_GetNumForName(name))
}

// Generate a hash table for fast lookups

func w_GenerateHashTable() {
	var hash uint32
	// Free the old hash table, if there is one
	lumphash = nil
	// Generate hash table
	if numlumps > 0 {
		lumphash = make([]*lumpinfo_t, numlumps)
		for i := range numlumps {
			hash = w_LumpNameHash(lumpinfo[i].Name()) % numlumps
			// Hook into the hash table
			lumpinfo[i].Fnext = lumphash[hash]
			lumphash[hash] = &lumpinfo[i]
		}
	}
	// All done!
}

// C documentation
//
//	// Lump names that are unique to particular game types. This lets us check
//	// the user is not trying to play with the wrong executable, eg.
//	// chocolate-doom -iwad hexen.wad.
var unique_lumps = [4]struct {
	Fmission  gamemission_t
	Flumpname string
}{
	0: {
		Flumpname: "POSSA1",
	},
	1: {
		Fmission:  heretic,
		Flumpname: "IMPXA1",
	},
	2: {
		Fmission:  hexen,
		Flumpname: "ETTNA1",
	},
	3: {
		Fmission:  strife,
		Flumpname: "AGRDA1",
	},
}

func w_CheckCorrectIWAD(mission gamemission_t) {
	for i := range len(unique_lumps) {
		if mission != unique_lumps[i].Fmission {
			lumpnum := w_CheckNumForName(unique_lumps[i].Flumpname)
			if lumpnum >= 0 {
				i_Error("\nYou are trying to use a %s IWAD file with the %s%s binary.\nThis isn't going to work.\nYou probably want to use the %s%s binary.", d_SuggestGameName(unique_lumps[i].Fmission, indetermined), "doomgeneric", d_GameMissionString(mission), "doomgeneric", d_GameMissionString(unique_lumps[i].Fmission))
			}
		}
	}
}

const MINFRAGMENT = 64
const ZONEID = 1919505

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Simple basic typedefs, isolated here to make it easier
//	 separating modules.
//

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Simple basic typedefs, isolated here to make it easier
//	 separating modules.
//

//
// ZONE MEMORY ALLOCATION
//
// There is never any space between memblocks,
//  and there will never be two contiguous free memblocks.
// The rover can be left pointing at a non-empty block.
//
// It is of no value to free a cachable block,
//  because it will get overwritten automatically if needed.
//

// C documentation
//
//	//
//	// Z_Init
//	//
func z_Init() {
	return
}

// Read data from the specified position in the file into the
// provided buffer.  Returns the number of bytes read.

func init() {
	vanilla_keyboard_mapping = 1
}

// Is the shift key currently down?

var shiftdown int32 = 0

// C documentation
//
//	// Lookup table for mapping ASCII characters to their equivalent when
//	// shift is pressed on an American layout keyboard:
var shiftxform = [128]uint8{
	1:   1,
	2:   2,
	3:   3,
	4:   4,
	5:   5,
	6:   6,
	7:   7,
	8:   8,
	9:   9,
	10:  10,
	11:  11,
	12:  12,
	13:  13,
	14:  14,
	15:  15,
	16:  16,
	17:  17,
	18:  18,
	19:  19,
	20:  20,
	21:  21,
	22:  22,
	23:  23,
	24:  24,
	25:  25,
	26:  26,
	27:  27,
	28:  28,
	29:  29,
	30:  30,
	31:  31,
	32:  ' ',
	33:  '!',
	34:  '"',
	35:  '#',
	36:  '$',
	37:  '%',
	38:  '&',
	39:  '"',
	40:  '(',
	41:  ')',
	42:  '*',
	43:  '+',
	44:  '<',
	45:  '_',
	46:  '>',
	47:  '?',
	48:  ')',
	49:  '!',
	50:  '@',
	51:  '#',
	52:  '$',
	53:  '%',
	54:  '^',
	55:  '&',
	56:  '*',
	57:  '(',
	58:  ':',
	59:  ':',
	60:  '<',
	61:  '+',
	62:  '>',
	63:  '?',
	64:  '@',
	65:  'A',
	66:  'B',
	67:  'C',
	68:  'D',
	69:  'E',
	70:  'F',
	71:  'G',
	72:  'H',
	73:  'I',
	74:  'J',
	75:  'K',
	76:  'L',
	77:  'M',
	78:  'N',
	79:  'O',
	80:  'P',
	81:  'Q',
	82:  'R',
	83:  'S',
	84:  'T',
	85:  'U',
	86:  'V',
	87:  'W',
	88:  'X',
	89:  'Y',
	90:  'Z',
	91:  '[',
	92:  '!',
	93:  ']',
	94:  '"',
	95:  '_',
	96:  '\'',
	97:  'A',
	98:  'B',
	99:  'C',
	100: 'D',
	101: 'E',
	102: 'F',
	103: 'G',
	104: 'H',
	105: 'I',
	106: 'J',
	107: 'K',
	108: 'L',
	109: 'M',
	110: 'N',
	111: 'O',
	112: 'P',
	113: 'Q',
	114: 'R',
	115: 'S',
	116: 'T',
	117: 'U',
	118: 'V',
	119: 'W',
	120: 'X',
	121: 'Y',
	122: 'Z',
	123: '{',
	124: '|',
	125: '}',
	126: '~',
	127: 127,
}

// Get the equivalent ASCII (Unicode?) character for a keypress.

func getTypedChar(key uint8) uint8 {
	// Is shift held down?  If so, perform a translation.
	if shiftdown > 0 {
		if key >= 0 && key < 128 {
			key = shiftxform[key]
		} else {
			key = 0
		}
	}
	return key
}

func updateShiftStatus(pressed int32, key uint8) {
	var change int32
	if pressed != 0 {
		change = 1
	} else {
		change = -1
	}
	if key == 0x80+0x36 {
		shiftdown += change
	}
}

type DoomEvent struct {
	Type  Evtype_t
	Key   uint8
	Mouse struct {
		Button1 bool
		Button2 bool
		XPos    float64 // from 0-1
		YPos    float64 // from 0-1
	}
}

var lastMouse DoomEvent

func i_GetEvent() {
	var newEvent event_t
	var event DoomEvent
	for dg_frontend.GetEvent(&event) {
		if event.Type == Ev_keydown || event.Type == Ev_keyup {
			pressed := int32(0)
			if event.Type == Ev_keydown {
				pressed = 1
			}
			updateShiftStatus(pressed, event.Key)
			// process event
			if event.Type == Ev_keydown {
				// data1 has the key pressed, data2 has the character
				// (shift-translated, etc)
				newEvent.Ftype1 = Ev_keydown
				newEvent.Fdata1 = int32(event.Key)
				newEvent.Fdata2 = int32(getTypedChar(event.Key))
				if newEvent.Fdata1 != 0 {
					d_PostEvent(&newEvent)
				}
			} else {
				newEvent.Ftype1 = Ev_keyup
				newEvent.Fdata1 = int32(event.Key)
				// data2 is just initialized to zero for ev_keyup.
				// For ev_keydown it's the shifted Unicode character
				// that was typed, but if something wants to detect
				// key releases it should do so based on data1
				// (key ID), not the printable char.
				newEvent.Fdata2 = 0
				if newEvent.Fdata1 != 0 {
					d_PostEvent(&newEvent)
				}
			}
		}
		if event.Type == Ev_mouse {
			if lastMouse.Type == 0 {
				lastMouse = event
			}
			newEvent.Ftype1 = Ev_mouse
			if event.Mouse.Button1 {
				newEvent.Fdata1 |= 1
			}
			if event.Mouse.Button2 {
				newEvent.Fdata1 |= 2
			}
			newEvent.Fdata2 = int32((event.Mouse.XPos - lastMouse.Mouse.XPos) * SCREENWIDTH * 100)
			newEvent.Fdata3 = int32((lastMouse.Mouse.YPos - event.Mouse.YPos) * SCREENHEIGHT * 100)
			if newEvent.Fdata2 < 5 && newEvent.Fdata2 > -5 &&
				newEvent.Fdata3 < 5 && newEvent.Fdata3 > -5 {
				// Ignore small mouse movements.
				continue
			}
			lastMouse = event
			d_PostEvent(&newEvent)
		}
	}
}

func i_InitInput() {
}

const INT_MAX19 = 2147483647

var colors [256]color.RGBA

func init() {
	mouse_acceleration = float32(2)
	mouse_threshold = 10
}

func i_InitGraphics() {
	/* Allocate screen to draw to */
	I_VideoBuffer = make([]byte, SCREENWIDTH*SCREENHEIGHT) // For DOOM to draw on
	i_InitInput()
}

func i_StartFrame() {
}

func i_StartTic() {
	i_GetEvent()
}

func i_UpdateNoBlit() {
}

//
// I_FinishUpdate
//

func i_FinishUpdate() {
	var line_in_pos = 0
	for y := SCREENHEIGHT - 1; y >= 0; y-- {
		for i := 0; i < SCREENWIDTH; i++ {
			inRaw := I_VideoBuffer[line_in_pos+i]
			col := colors[inRaw]
			pos := SCREENWIDTH*4*int(SCREENHEIGHT-y-1) + i*4
			DG_ScreenBuffer.Pix[pos] = col.R
			DG_ScreenBuffer.Pix[pos+1] = col.G
			DG_ScreenBuffer.Pix[pos+2] = col.B
			DG_ScreenBuffer.Pix[pos+3] = 0xff
		}
		line_in_pos += SCREENWIDTH
	}
	dg_frontend.DrawFrame(DG_ScreenBuffer)
}

// C documentation
//
//	//
//	// I_ReadScreen
//	//
func i_ReadScreen(scr []byte) {
	copy(scr, I_VideoBuffer)
}

//
// I_SetPalette
//

func i_SetPalette(palette []byte) {
	for i := range 256 {
		colors[i].R = palette[i*3]
		colors[i].G = palette[i*3+1]
		colors[i].B = palette[i*3+2]
	}
}

// Given an RGB value, find the closest matching palette index.

func i_GetPaletteIndex(r int32, g int32, b int32) (r1 int32) {
	var best, best_diff, diff int32
	fprintf_ccgo(os.Stdout, "i_GetPaletteIndex\n")
	best = 0
	best_diff = int32(INT_MAX19)
	for i := int32(0); i < 256; i++ {
		red := int32(colors[i].R)
		green := int32(colors[i].G)
		blue := int32(colors[i].B)
		diff = (r-red)*(r-red) + (g-green)*(g-green) + (b-blue)*(b-blue)
		if diff < best_diff {
			best = i
			best_diff = diff
		}
		if diff == 0 {
			break
		}
	}
	return best
}

func i_SetWindowTitle(title string) {
	dg_frontend.SetTitle(title)
}

func i_GraphicsCheckCommandLine() {
}

func i_SetGrabMouseCallback(func1 func() boolean) {
}

func i_EnableLoadingDisk() {
}

func i_BindVideoVariables() {
}

func i_DisplayFPSDots(dots_on boolean) {
}

func i_CheckIsScreensaver() {
}

func doomgeneric_Create(args []string) {
	// save arguments
	myargs = args
	m_FindResponseFile()

	DG_ScreenBuffer = image.NewRGBA(image.Rect(0, 0, SCREENWIDTH, SCREENHEIGHT))
	d_DoomMain()
}

func Run(fg DoomFrontend, args []string) {
	if dg_frontend != nil {
		log.Printf("Run called twice, ignoring second call")
	}
	dg_frontend = fg
	dg_exiting = false
	start_time = time.Now()

	// Convert command line arguments to C strings.
	args = append([]string{"doom"}, args...) // prepend "doom" as argv[0]
	doomgeneric_Create(args)
	for !dg_exiting {
		doomgeneric_Tick()
	}
	dg_frontend = nil
}

func Stop() {
	dg_exiting = true
}

var DG_ScreenBuffer *image.RGBA

var EpiDef menu_t

var EpisodeMenu [4]menuitem_t

// The screen buffer; this is modified to draw things to the screen

var I_VideoBuffer []byte

var LoadDef menu_t

var LoadMenu [6]menuitem_t

var MainDef menu_t

var MainMenu [6]menuitem_t

var NewDef menu_t

var NewGameMenu [5]menuitem_t

var OptionsDef menu_t

var OptionsMenu [8]menuitem_t

var ReadDef1 menu_t

var ReadDef2 menu_t

var ReadMenu1 [1]menuitem_t

var ReadMenu2 [1]menuitem_t

//
// Information about all the music
//

var S_music [68]musicinfo_t

//
// Information about all the sfx
//

var S_sfx [NUMSFX]sfxinfo_t

var SaveDef menu_t

// C documentation
//
//	//
//	// SAVE GAME MENU
//	//
var SaveMenu [6]menuitem_t

var SoundDef menu_t

var SoundMenu [4]menuitem_t

var TRACEANGLE angle_t

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// Data.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

//
// CEILINGS
//

var activeceilings [MAXCEILINGS]*ceiling_t

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// Data.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

var activeplats [MAXPLATS]*plat_t

// Used in the test suite to stop the demo running in the
// background, as it messes with screenshots
var dont_run_demo bool
var advancedemo boolean

var aimslope fixed_t

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// C documentation
//
//	//
//	// CHANGE THE TEXTURE OF A WALL SWITCH TO ITS OPPOSITE
//	//
var alphSwitchList [41]switchlist_t

var angleturn [3]fixed_t

//
// P_InitPicAnims
//

// C documentation
//
//	// Floor/ceiling animation sequences,
//	//  defined by first and last frame,
//	//  i.e. the flat (64x64 tile) name to
//	//  be used.
//	// The full animation sequence is given
//	//  using all the flats between the start
//	//  and end entry, in the order found in
//	//  the WAD file.
//	//
var animdefs [23]animdef_t

var anims [32]anim_t

var attackrange fixed_t

var automapactive boolean

var autostart boolean

var backsector *sector_t

var basecolfunc func()

var basexscale fixed_t

var baseyscale fixed_t

// C documentation
//
//	//
//	// SLIDE MOVE
//	// Allows the player to slide along any angled walls.
//	//
var bestslidefrac fixed_t

var bestslideline *line_t

// C documentation
//
//	// "BFG Edition" version of doom2.wad does not include TITLEPIC.
var bfgedition boolean

// C documentation
//
//	// for thing chains
var blocklinks []*mobj_t

var blockmap []int16

// offsets in blockmap are from here
var blockmaplump []int16

var bmapheight int32

// C documentation
//
//	// origin of block map
var bmaporgx fixed_t

var bmaporgy fixed_t

// C documentation
//
//	// BLOCKMAP
//	// Created from axis aligned bounding box
//	// of the map, a rectangular array of
//	// blocks of size ...
//	// Used to speed up collision detection
//	// by spatial subdivision in 2D.
//	//
//	// Blockmap size.
var bmapwidth int32

var bodyque [32]*mobj_t

var bodyqueslot int32

var bombdamage int32

// C documentation
//
//	//
//	// RADIUS ATTACK
//	//
var bombsource *mobj_t

var bombspot *mobj_t

var bottomfrac fixed_t

var bottomslope fixed_t

var bottomstep fixed_t

var bottomtexture int32

var braintargeton int32

var braintargets [32]*mobj_t

// C documentation
//
//	//
//	// P_BulletSlope
//	// Sets a slope so a near miss is at aproximately
//	// the height of the intended target
//	//
var bulletslope fixed_t

var buttonlist [MAXBUTTONS]button_t

var cacheddistance [200]fixed_t

var cachedheight [200]fixed_t

var cachedxstep [200]fixed_t

var cachedystep [200]fixed_t

var castattacking boolean

var castdeath boolean

var castframes int32

var castnum int32

var castonmelee int32

var castorder [18]castinfo_t

var caststate *state_t

var casttics int32

var ceilingclip [320]int16

// C documentation
//
//	// keep track of the line that lowers the ceiling,
//	// so missiles don't explode against sky hack walls
var ceilingline *line_t

var ceilingplane *visplane_t

var centerx int32

var centerxfrac fixed_t

var centery int32

var centeryfrac fixed_t

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

//
// Locally used constants, shortcuts.
//

var chat_macros [10]string

var chat_on boolean

var cheat_amap cheatseq_t

var cheat_ammo cheatseq_t

var cheat_ammonokey cheatseq_t

var cheat_choppers cheatseq_t

var cheat_clev cheatseq_t

var cheat_commercial_noclip cheatseq_t

var cheat_god cheatseq_t

var cheat_mus cheatseq_t

var cheat_mypos cheatseq_t

var cheat_noclip cheatseq_t

var cheat_player_arrow [16]mline_t

var cheat_powerup [7]cheatseq_t

// C documentation
//
//	//
//	// R_CheckBBox
//	// Checks BSP node/subtree bounding box.
//	// Returns true
//	//  if some part of the bbox might be visible.
//	//
var checkcoord [12][4]int32

var clipammo [4]int32

// C documentation
//
//	//
//	// precalculated math tables
//	//
var clipangle angle_t

var colfunc func()

var colormaps []lighttable_t

var columnofs [1120]int32

//
// This is used to get the local FILE:LINE info from CPP
// prior to really call the function in question.
//

//
// DEFAULTS
//

// Location where all configuration data is stored -
// default.cfg, savegames, etc.

var configdir string

var consistancy [4][128]uint8

var consoleplayer int32

// C documentation
//
//	//
//	// PIT_VileCheck
//	// Detect a corpse that could be raised.
//	//
var corpsehit *mobj_t

// C documentation
//
//	// DOOM II Par Times
var cpars [32]int32

// C documentation
//
//	//
//	// SECTOR HEIGHT CHANGING
//	// After modifying a sectors floor or ceiling height,
//	// call this routine to adjust the positions
//	// of all things that touch the sector.
//	//
//	// If anything doesn't fit anymore, true will be returned.
//	// If crunch is true, they will take damage
//	//  as they are being crushed.
//	// If Crunch is false, you should set the sector height back
//	//  the way it was and call p_ChangeSector again
//	//  to undo the changes.
//	//
var crushchange boolean

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

//#include "r_local.h"

var curline *seg_t

// C documentation
//
//	// current menudef
var currentMenu *menu_t

var d_episode int32

var d_map int32

// C documentation
//
//	//
//	// G_InitNew
//	// Can be called by the startup code or the menu task,
//	// consoleplayer, displayplayer, playeringame[] should be set.
//	//
var d_skill skill_t

// C documentation
//
//	//
//	// R_DrawColumn
//	// Source is the top of the column to scale.
//	//
var dc_colormap []lighttable_t

var dc_iscale fixed_t

// C documentation
//
//	// first pixel in a column (possibly virtual)
var dc_source uintptr

var dc_texturemid fixed_t

// C documentation
//
//	//
//	// R_DrawTranslatedColumn
//	// Used to draw player sprites
//	//  with the green colorramp mapped to others.
//	// Could be used with different translation
//	//  tables, e.g. the lighter colored version
//	//  of the BaronOfHell, the HellKnight, uses
//	//  identical sprites, kinda brightened up.
//	//
var dc_translation []byte

var dc_x int32

var dc_yh int32

var dc_yl int32

// Control whether if a mouse button is double clicked, it acts like
// "use" has been pressed

var dclick_use int32

var deathmatch int32

var deathmatch_pos int

// Maintain single and multi player starting spots.

var deathmatchstarts [10]mapthing_t

//
// G_PlayDemo
//

var defdemoname string

var demo_pos int

var demobuffer []byte

var demoname string

var demoplayback boolean

var demorecording boolean

// C documentation
//
//	//
//	//  DEMO LOOP
//	//
var demosequence int32

// C documentation
//
//	// Blocky mode, has default, 0 = high, 1 = normal
var detailLevel int32

// C documentation
//
//	// 0 = high, 1 = low
var detailshift int32

var devparm boolean

var diags [4]dirtype_t

var dirtybox box_t

var displayplayer int32

var distscale [320]fixed_t

var doom1_endmsg [NUM_QUITMESSAGES]string

var doom2_endmsg [NUM_QUITMESSAGES]string

var drawsegs [256]drawseg_t

var drone boolean

var ds_colormap []lighttable_t

var ds_index int

// C documentation
//
//	// start of a 64*64 tile image
var ds_source []byte

var ds_x1 int32

var ds_x2 int32

var ds_xfrac fixed_t

var ds_xstep fixed_t

// C documentation
//
//	//
//	// R_DrawSpan
//	// With DOOM style restrictions on view orientation,
//	//  the floors and ceilings consist of horizontal slices
//	//  or spans with constant z depth.
//	// However, rotation around the world z axis is possible,
//	//  thus this mapping, while simpler and faster than
//	//  perspective correct texture mapping, has to traverse
//	//  the texture at an angle in all but a few cases.
//	// In consequence, flats are not stored by column (like walls),
//	//  and the inner loop has to step in texture space u and v.
//	//
var ds_y int32

var ds_yfrac fixed_t

var ds_ystep fixed_t

var earlyout boolean

var endstring string

// C documentation
//
//	//
//	//      M_Episode
//	//
var epi int32

// C documentation
//
//	// bumped light from gun blasts
var extralight int32

var fastparm boolean

var finalecount uint32

var finaleflat string

// ?
//#include "doomstat.h"
//#include "r_local.h"
//#include "f_finale.h"

// C documentation
//
//	// Stage of animation:
var finalestage finalestage_t

var finaletext string

var finecosine []fixed_t

var finesine [10240]fixed_t

var finetangent [4096]fixed_t

var firstflat int32

var firstspritelump int32

var fixedcolormap []lighttable_t

// C documentation
//
//	// for global animation
var flattranslation []int32

// C documentation
//
//	// If "floatok" true, move would be ok
//	// if within "tmfloorz - tmceilingz".
var floatok boolean

// C documentation
//
//	//
//	// Clip values are the solid pixel bounding the range.
//	//  floorclip starts out SCREENHEIGHT
//	//  ceilingclip starts out -1
//	//
var floorclip [320]int16

var floorplane *visplane_t

var forwardmove [2]fixed_t

// C documentation

var frontsector *sector_t

var fuzzcolfunc func()

//
// Spectre/Invisibility.
//

var fuzzoffset [50]int32

var fuzzpos int32

var gameaction gameaction_t

var gamedescription string

var gameepisode int32

var gamemap int32

var gamemission gamemission_t

// C documentation
//
//	// Game Mode - identify IWAD as shareware, retail etc.
var gamemode gamemode_t

var gameskill skill_t

var gamestate gamestate_t

// The number of tics that have been run (using runTic) so far.

var gametic int32

var gameversion gameversion_t

var gammamsg [5]string

// C documentation
var hu_font [63]*patch_t

var inhelpscreens boolean

var intercept_pos int32

// C documentation
//
//	//
//	// INTERCEPT ROUTINES
//	//
var intercepts [189]intercept_t

var iquehead int32

var iquetail int32

var itemOn int16

// C documentation
//
//	//
//	// P_RemoveMobj
//	//
var itemrespawnque [128]mapthing_t

var itemrespawntime [128]int32

// location of IWAD and WAD files

var iwadfile string

//
// Joystick controls
//

var joybfire int32

var joybjump int32

var joybmenu int32

var joybnextweapon int32

var joybprevweapon int32

var joybspeed int32

var joybstrafe int32

var joybstrafeleft int32

var joybstraferight int32

var joybuse int32

var key_arti_all int32

var key_arti_blastradius int32

var key_arti_egg int32

var key_arti_health int32

var key_arti_invulnerability int32

var key_arti_poisonbag int32

var key_arti_teleport int32

var key_arti_teleportother int32

var key_demo_quit int32

var key_down int32

var key_fire int32

var key_flycenter int32

var key_flydown int32

// Heretic keyboard controls
var key_flyup int32

var key_invdrop int32

var key_invend int32

var key_invhome int32

var key_invkey int32

var key_invleft int32

var key_invpop int32

var key_invquery int32

var key_invright int32

var key_invuse int32

//
// Hexen key controls
//

var key_jump int32

var key_left int32

var key_lookcenter int32

var key_lookdown int32

var key_lookup int32

var key_map_clearmark int32

var key_map_east int32

var key_map_follow int32

var key_map_grid int32

var key_map_mark int32

var key_map_maxzoom int32

// Map control keys:

var key_map_north int32

var key_map_south int32

var key_map_toggle int32

var key_map_west int32

var key_map_zoomin int32

var key_map_zoomout int32

var key_menu_abort int32

// menu keys:

var key_menu_activate int32

var key_menu_back int32

var key_menu_confirm int32

var key_menu_decscreen int32

var key_menu_detail int32

var key_menu_down int32

var key_menu_endgame int32

var key_menu_forward int32

var key_menu_gamma int32

var key_menu_help int32

var key_menu_incscreen int32

var key_menu_left int32

var key_menu_load int32

var key_menu_messages int32

var key_menu_qload int32

var key_menu_qsave int32

var key_menu_quit int32

var key_menu_right int32

var key_menu_save int32

var key_menu_screenshot int32

var key_menu_up int32

var key_menu_volume int32

var key_message_refresh int32

var key_mission int32

// Multiplayer chat keys:

var key_multi_msg int32

var key_multi_msgplayer [8]int32

var key_nextweapon int32

var key_pause int32

var key_prevweapon int32

//
// Keyboard controls
//

var key_right int32

var key_speed int32

var key_spy int32

var key_strafe int32

var key_strafeleft int32

var key_straferight int32

var key_up int32

var key_use int32

var key_useartifact int32

//
// Strife key controls
//
// haleyjd 09/01/10
//

// Note: Strife also uses key_invleft, key_invright, key_jump, key_lookup, and
// key_lookdown, but with different default values.

var key_usehealth int32

// Weapon selection keys:

var key_weapon1 int32

var key_weapon2 int32

var key_weapon3 int32

var key_weapon4 int32

var key_weapon5 int32

var key_weapon6 int32

var key_weapon7 int32

var key_weapon8 int32

var la_damage int32

var lastanim *anim_t

var lastflat int32

var lastopening uintptr

var lastspritelump int32

// C documentation
//
//	//
//	// NetUpdate
//	// Builds ticcmds for console player,
//	// sends out a packet
//	//
var lasttime int32

var lastvisplane_index int

var levelTimeCount int32

// C documentation
//
//	//
//	// P_UpdateSpecials
//	// Animate planes, scroll walls, etc.
//	//
var levelTimer boolean

var leveltime int32

var linedef *line_t

var lines []line_t

// TODO: ANDRE/GORE: This is a hack to allow easy conversion of addresses into indexs
func lineIndex(l *line_t) int32 {
	idx := int32((uintptr(unsafe.Pointer(l)) - uintptr(unsafe.Pointer(&lines[0]))) / unsafe.Sizeof(line_t{}))
	if idx < 0 || idx >= int32(len(lines)) {
		log.Fatalf("lineIndex: line %p out of bounds, %d lines length %d", l, idx, len(lines))
	}
	return idx
}

var linespeciallist [64]*line_t

// C documentation
//
//	//
//	// P_LineAttack
//	//
var linetarget *mobj_t

var longtics boolean

var lowfloor fixed_t

var lowres_turn boolean

//
// GLOBALS
//

// Location of each lump on disk.

// TODO: GORE/ANDRE - once we've got Go memory management fully in place, this should
// become a dynamic array that can be resized. We keep it static at the moment
// so that addresses don't change for the Z_Change... functions
var lumpinfo [4096]lumpinfo_t

func lumpIndex(l *lumpinfo_t) int32 {
	idx := (uintptr(unsafe.Pointer(l)) - uintptr(unsafe.Pointer(&lumpinfo[0]))) / unsafe.Sizeof(lumpinfo_t{})
	if idx < 0 || idx >= uintptr(len(lumpinfo)) {
		log.Fatalf("lumpIndex: lump %p out of bounds, %d lumps length", l, len(lumpinfo))
	}
	return int32(idx)
}

// C documentation
//
//	// If true, the main game loop has started.
var main_loop_started boolean

//
// Builtin map names.
// The actual names can be found in DStrings.h.
//

var mapnames [45]string

// List of names for levels in commercial IWADs
// (doom2.wad, plutonia.wad, tnt.wad).  These are stored in a
// single large array; WADs like pl2.wad have a MAP33, and rely on
// the layout in the Vanilla executable, where it is possible to
// overflow the end of one array into the next.

var mapnames_commercial [96]string

var markceiling boolean

// C documentation
//
//	// False if the back side is the same plane.
var markfloor boolean

var maskedtexture boolean

var maskedtexturecol uintptr

// C documentation
//
//	// a weapon is found with two clip loads,
//	// a big item has five clip loads
var maxammo [4]int32

var maxframe int32

var mceilingclip uintptr

var menuactive boolean

var messageLastMenuActive int32

// C documentation
//
//	// timed message = no input from user
var messageNeedsInput boolean

var messageRoutine *func(int32)

// C documentation
//
//	// ...and here is the message string!
var messageString string

// C documentation
//
//	// 1 = message to be printed
var messageToPrint int32

var message_dontfuckwithme boolean

// C documentation
//
//	//
//	// R_DrawMaskedColumn
//	// Used for sprites and masked mid textures.
//	// Masked means: partly transparent, i.e. stored
//	//  in posts/runs of opaque pixels.
//	//
var mfloorclip uintptr

var midtexture int32

var mobjinfo [137]mobjinfo_t

// C documentation
//
//	// Set if homebrew PWAD stuff has been added.
var modifiedgame boolean

// C documentation
//
//	//
//	// defaulted values
//	//
var mouseSensitivity int32

// Mouse acceleration
//
// This emulates some of the behavior of DOS mouse drivers by increasing
// the speed when the mouse is moved fast.
//
// The mouse input values are input directly to the game, but when
// the values exceed the value of mouse_threshold, they are multiplied
// by mouse_acceleration to increase the speed.

var mouse_acceleration float32

var mouse_threshold int32

var mousebbackward int32

//
// Mouse controls
//

var mousebfire int32

var mousebforward int32

var mousebjump int32

var mousebnextweapon int32

var mousebprevweapon int32

var mousebstrafe int32

var mousebstrafeleft int32

var mousebstraferight int32

var mousebuse int32

// C documentation
//
//	// mouse values are used once
var mousex int32

var mousey int32

// Maximum volume of music.

var musicVolume int32

var myargs []string

// C documentation
//
//	// constant arrays
//	//  used for psprite clipping and initializing clipping
var negonearray [320]int16

/* Support signed or unsigned plain-char */

/* Implementation choices... */

/* Arbitrary numbers... */

/* POSIX/SUS requirements follow. These numbers come directly
 * from SUS and have nothing to do with the host system. */

/*---------------------------------------------------------------------*
 *  local definitions                                                  *
 *---------------------------------------------------------------------*/

/*---------------------------------------------------------------------*
 *  external declarations                                              *
 *---------------------------------------------------------------------*/

/*---------------------------------------------------------------------*
 *  public data                                                        *
 *---------------------------------------------------------------------*/

var net_client_connected boolean

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Main loop stuff.
//

var netcmds []ticcmd_t

var netdemo boolean

var netgame boolean

// C documentation
//
//	// newend is one past the last valid seg in `solidsegs`
var newend int

var nodes []node_t

var nodrawers boolean

var nofit boolean

var nomonsters boolean

var numbraintargets int32

var numflats int32

var numlines int32

//
// SPECIAL SPAWNING
//

// C documentation
//
//	//
//	// P_SpawnSpecials
//	// After the map has been loaded, scan for specials
//	//  that spawn thinkers
//	//
var numlinespecials int16

var numlumps uint32

var numnodes int32

var numsectors int32

var numsegs int32

var numsides int32

var numspechit int32

var numspritelumps int32

var numsprites int32

var numsubsectors int32

var numswitches int32

var numtextures int32

// C documentation
//
//	//
//	// MAP related Lookup tables.
//	// Store VERTEXES, LINEDEFS, SIDEDEFS, etc.
//	//
var numvertexes int32

// Amount to offset the timer for game sync.

var offsetms fixed_t

// Gamestate the last time g_Ticker was called.

var oldgamestate gamestate_t

// Index of the special effects (INVUL inverse) map.

//
// Movement.
//

// 16 pixels of bob

var onground boolean

var openbottom fixed_t

// C documentation
//
//	// ?
var openings [20480]int16

var openrange fixed_t

// C documentation
//
//	//
//	// P_LineOpening
//	// Sets opentop and openbottom to the window
//	// through a two sided line.
//	// OPTIMIZE: keep this precalculated
//	//
var opentop fixed_t

// C documentation
//
//	//
//	// p_NewChaseDir related LUT.
//	//
var opposite [9]dirtype_t

// C documentation
//
//	//
//	// R_NewVisSprite
//	//
var overflowsprite vissprite_t

var pagename string

var pagetic int32

// C documentation
//
//	// DOOM Par Times
var pars [4][10]int32

var paused boolean

var pixhigh fixed_t

var pixhighstep fixed_t

var pixlow fixed_t

var pixlowstep fixed_t

var planeheight fixed_t

// C documentation
//
//	//
//	// texture mapping
//	//
var planezlight [][]lighttable_t

// C documentation
//
//	//
//	// The vector graphics for the automap.
//	//  A line drawing of the player pointing right,
//	//   starting from the middle.
//	//
var player_arrow [7]mline_t

var player_names [4]string

var playeringame [4]boolean

var players [4]player_t

func playerIndex(p *player_t) int32 {
	idx := int32((uintptr(unsafe.Pointer(p)) - uintptr(unsafe.Pointer(&players[0]))) / unsafe.Sizeof(player_t{}))
	if idx < 0 || idx >= int32(len(players)) {
		log.Fatalf("playerIndex: player %p out of bounds, %d players length %d", p, idx, len(players))
	}
	return idx
}

var playerstarts [4]mapthing_t

var precache boolean

var prndindex int32

var projection fixed_t

var pspriteiscale fixed_t

// C documentation
//
//	//
//	// Sprite rotation 0 is facing the viewer,
//	//  rotation 1 is one angle turn CLOCKWISE around the axis.
//	// This is not the same as the angle,
//	//  which increases counter clockwise (protractor).
//	// There was a lot of stuff grabbed wrong, so I changed it...
//	//
var pspritescale fixed_t

// C documentation
//
//	// -1 = no quicksave slot picked!
var quickSaveSlot int32

// C documentation
//
//	//
//	// M_QuitDOOM
//	//
var quitsounds [8]int32

var quitsounds2 [8]int32

// C documentation
//
//	// REJECT
//	// For fast sight rejection.
//	// Speeds up enemy AI by skipping detailed
//	//  LineOf Sight calculation.
//	// Without special effect, this could be
//	//  used as a PVS lookup as well.
//	//
var rejectmatrix []byte

var respawnmonsters boolean

var respawnparm boolean

var rndindex int32

// C documentation
//
//	// angle to line origin
var rw_angle1 int32

var rw_bottomtexturemid fixed_t

var rw_centerangle angle_t

var rw_distance fixed_t

var rw_midtexturemid fixed_t

var rw_normalangle angle_t

var rw_offset fixed_t

var rw_scale fixed_t

var rw_scalestep fixed_t

var rw_stopx int32

var rw_toptexturemid fixed_t

// C documentation
//
//	//
//	// regular wall
//	//
var rw_x int32

var saveCharIndex int

// old save description before edit
var saveOldString string

var saveSlot int32

// C documentation
//
//	// we are going to be entering a savegame string
var saveStringEnter int32

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

var save_stream *os.File

var savegame_error boolean

// Location where savegames are stored

var savegamedir string

var savegamestrings [10]string

var savename string

var scaledviewwidth int32

var scalelight [16][48][]lighttable_t

var scalelightfixed [48][]lighttable_t

// C documentation
//
//	// temp for screenblocks (0-9)
var screenSize int32

var screenblocks int32

var screenheightarray [320]int16

// If true, game is running as a screensaver

var screensaver_mode boolean

// C documentation
//
//	//
//	// G_DoCompleted
//	//
var secretexit boolean

var sectors []sector_t

// TODO: ANDRE/GORE: Faster way do to pointer division to determine offset?
func sectorIndex(sector *sector_t) int32 {
	idx := int32((uintptr(unsafe.Pointer(sector)) - uintptr(unsafe.Pointer(&sectors[0]))) / unsafe.Sizeof(sector_t{}))
	if idx < 0 || idx >= int32(len(sectors)) {
		log.Fatalf("sectorIndex: sector %p out of bounds, %d sectors length %d", sector, idx, len(sectors))
	}
	return idx
}

var segs []seg_t

// OPTIMIZE: closed two sided lines as single sided

// C documentation
//
//	// True if any of the segs textures might be visible.
var segtextured boolean

var sendpause boolean

var sendsave boolean

var setblocks int32

var setdetail int32

// C documentation
//
//	//
//	// R_SetViewSize
//	// Do not really change anything here,
//	//  because it might be in the middle of a refresh.
//	// The change will take effect next refresh.
//	//
var setsizeneeded boolean

// Maximum volume of a sound effect.
// Internal default is max out of 0-15.

var sfxVolume int32

var shootthing *mobj_t

// C documentation
//
//	// Height if not aiming up or down
//	// ???: use slope for monsters?
var shootz fixed_t

// C documentation
//
//	// Show messages has default, 0 = off, 1 = on
var showMessages int32

var show_endoom int32

var sidedef *side_t

var sidemove [2]fixed_t

var sides []side_t

var sightcounts [2]int32

// State.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// C documentation
//
//	//
//	// P_CheckSight
//	//
var sightzstart fixed_t

var singledemo boolean

// When set to true, a single tic is run each time tryRunTics() is called.
// This is used for -timedemo mode.

var singletics boolean

var skullAnimCounter int16

// C documentation
//
//	// graphic name of skulls
//	// warning: initializer-string for array of chars is too long
var skullName [2]string

// C documentation
//
//	//
//	// sky mapping
//	//
var skyflatnum int32

var skytexture int32

var skytexturemid int32

var slidemo *mobj_t

// Maximum number of bytes to dedicate to allocated sound effects.
// (Default: 64MB)

var snd_cachesize int32

// Number of channels to use

var snd_channels int32

// Config variable that controls the sound buffer size.
// We default to 28ms (1000 / 35fps = 1 buffer per tic).

var snd_maxslicetime_ms int32

// External command to invoke to play back music.

var snd_musiccmd string

var snd_musicdevice int32

// Sound sample rate to use for digital output (Hz)

var snd_samplerate int32

var snd_sfxdevice int32

var solidsegs [32]cliprange_t

//
// ENEMY THINKING
// Enemies are allways spawned
// with targetplayer = -1, threshold = 0
// Most monsters are spawned unaware of all players,
// but some can be made preaware
//

//
// Called by p_NoiseAlert.
// Recursively traverse adjacent sectors,
// sound blocking lines cut off traversal.
//

var soundtarget *mobj_t

var spanfunc func()

// C documentation
//
//	//
//	// spanstart holds the start of a plane span
//	// initialized to 0 at start
//	//
var spanstart [200]int32

// keep track of special lines as they are hit,
// but don't process them until the move is proven valid

var spechit [20]*line_t

var spritelights [48][]lighttable_t

var spriteoffset []fixed_t

//
// INITIALIZATION FUNCTIONS
//

// C documentation
//
//	// variables used to look up
//	//  and range check thing_t sprites patches
var sprites []spritedef_t

var spritetopoffset []fixed_t

// C documentation
//
//	// needed for pre rendering
var spritewidth []fixed_t

var sprnames []string

var sprtemp [29]spriteframe_t

var sprtopscreen fixed_t

var spryscale fixed_t

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

//
// STATUS BAR DATA
//

// Palette indices.
// For damage/bonus red-/gold-shifts
// Radiation suit, green shift.

// N/256*100% probability
//  that the normal face state will change

// For Responder

// Location of status bar

// Should be set to patch width
//  for tall numbers later on

// Number of status faces.

// Location and size of statistics,
//  justified according to widget type.
// Problem is, within which space? STbar? Screen?
// Note: this could be read in by a lump.
//       Problem is, is the stuff rendered
//       into a buffer,
//       or into the frame buffer?

// AMMO number pos.

// HEALTH number pos.

// Weapon pos.

// Frags pos.

// ARMOR number pos.

// Key icon positions.

// Ammunition counter.

// Indicate maximum ammunition.
// Only needed because backpack exists.

// pistol

// shotgun

// chain gun

// missile launcher

// plasma gun

// bfg

// WPNS title

// DETH title

//Incoming messages window location
//UNUSED
// #define st_MSGTEXTX	   (viewwindowx)
// #define st_MSGTEXTY	   (viewwindowy+viewheight-18)
// Dimensions given in characters.
// Or shall I say, in lines?

// Width, in characters again.
// Height, in lines.

// C documentation
//
//	// graphics are drawn to a backing screen and blitted to the real screen
var st_backing_screen []byte

var startepisode int32

var startloadgame int32

var startmap int32

var startskill skill_t

var starttime int32

var states [967]state_t

func stateIndex(s *state_t) int32 {
	idx := int32((uintptr(unsafe.Pointer(s)) - uintptr(unsafe.Pointer(&states[0]))) / unsafe.Sizeof(state_t{}))
	if idx < 0 || idx >= int32(len(states)) {
		log.Fatalf("stateIndex: state %p out of bounds, %d states length %d", s, idx, len(states))
	}
	return idx
}

// C documentation
//
//	// Store demo, do not accept any inputs
var storedemo boolean

var strace divline_t

// C documentation
//
//	//
//	// Hack display negative frags.
//	//  Loads and store the stminus lump.
//	//
var sttminus *patch_t

var subsectors []subsector_t

var switchlist [100]int32

var t2x fixed_t

var t2y fixed_t

var tantoangle [2049]angle_t

// C documentation
//
//	//
//	//      M_QuickSave
//	//

// if true, load all graphics at start

var testcontrols boolean

var testcontrols_mousespeed int32

var texturecolumnlump [][]int16

var texturecolumnofs [][]uint16

var texturecomposite [][]byte

var texturecompositesize []int32

// C documentation
//
//	// needed for texture pegging
var textureheight []fixed_t

var textures []*texture_t

var textures_hashtable []*texture_t

var texturetranslation []int32

var texturewidthmask []int32

//
// THINKERS
// All thinkers should be allocated by Z_Malloc
// so they can be operated on uniformly.
// The actual structures will vary in size,
// but the first element must be thinker_t.
//

// C documentation
//
//	// Both the head and tail of the thinker list.
var thinkercap thinker_t

var thintriangle_guy [3]mline_t

// Reduce the bandwidth needed by sampling game input less and transmitting
// less.  If ticdup is 2, sample half normal, 3 = one third normal, etc.

var ticdup int32

// If non-zero, exit the level after this number of minutes.

var timelimit int32

var timingdemo boolean

// C documentation

//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Refresh/render internal state variables (global).
//

// Data.
//
// Copyright(C) 1993-1996 Id Software, Inc.
// Copyright(C) 2005-2014 Simon Howard
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// DESCRIPTION:
//	Created by the sound utility written by Dave Taylor.
//	Kept as a sample, DOOM2  sounds. Frozen.
//

// Spechit overrun magic value.
//
// This is the value used by PrBoom-plus.  I think the value below is
// actually better and works with more demos.  However, I think
// it's better for the spechits emulation to be compatible with
// PrBoom-plus, at least so that the big spechits emulation list
// on Doomworld can also be used with Chocolate Doom.

// This is from a post by myk on the Doomworld forums,
// outputted from entryway's spechit_magic generator for
// s205n546.lmp.  The _exact_ value of this isn't too
// important; as long as it is in the right general
// range, it will usually work.  Otherwise, we can use
// the generator (hacked doom2.exe) and provide it
// with -spechit.

//#define DEFAULT_SPECHIT_MAGIC 0x84f968e8

var tmbbox box_t

var tmceilingz fixed_t

var tmdropoffz fixed_t

var tmflags int32

var tmfloorz fixed_t

var tmthing *mobj_t

var tmx fixed_t

var tmxmove fixed_t

var tmy fixed_t

var tmymove fixed_t

var topfrac fixed_t

var topslope fixed_t

var topstep fixed_t

var toptexture int32

var totalitems int32

var totalkills int32

var totalsecret int32

var trace divline_t

var transcolfunc func()

var translationtables []byte

var turbodetected [4]boolean

// Gamma correction level to use

var usegamma int32

var usemouse int32

var usergame boolean

// C documentation
//
//	//
//	// USE LINES
//	//
var usething *mobj_t

// C documentation
//
//	// increment every time a check is made
var validcount int32

var vanilla_demo_limit int32

var vanilla_keyboard_mapping int32

var vanilla_savegame_limit int32

var vertexes []vertex_t

var viewactive boolean

var viewangle angle_t

// Fineangles in the SCREENWIDTH wide window.

var viewangleoffset uint32

// C documentation
//
//	// The viewangletox[viewangle + FINEANGLES/4] lookup
//	// maps the visible view angles to screen X coordinates,
//	// flattening the arc to a flat projection plane.
//	// There will be many angles mapped to the same X.
var viewangletox [4096]int32

var viewcos fixed_t

var viewheight int32

var viewplayer *player_t

var viewsin fixed_t

var viewwidth int32

var viewwindowx int32

var viewwindowy int32

var viewx fixed_t

var viewy fixed_t

var viewz fixed_t

var viletryx fixed_t

var viletryy fixed_t

//
// opening
//

// C documentation
//
//	// Here comes the obnoxious "visplane".
var visplanes [128]visplane_t

var vissprite_n int

// C documentation
//
//	//
//	// GAME FUNCTIONS
//	//
var vissprites [128]vissprite_t

// C documentation
//
//	//
//	// R_SortVisSprites
//	//
var vsprsortedhead vissprite_t

var walllights [48][]lighttable_t

// C documentation
//
//	//
//	// PSPRITE ACTIONS for waepons.
//	// This struct controls the weapon animations.
//	//
//	// Each entry is:
//	//   ammo/amunition type
//	//  upstate
//	//  downstate
//	// readystate
//	// atkstate, i.e. attack/fire/hit frame
//	// flashstate, muzzle flash
//	//
var weaponinfo [9]weaponinfo_t

var whichSkull int16

//
// D_Display
//  draw current display, possibly wiping it from the previous
//

// C documentation
//
//	// wipegamestate can be set to -1 to force a wipe on the next draw
var wipegamestate gamestate_t

var wminfo wbstartstruct_t

var worldbottom int32

var worldhigh int32

var worldlow int32

var worldtop int32

// C documentation
//
//	//
//	// P_Move
//	// Move in the current direction,
//	// returns false if the move is blocked.
//	//
var xspeed [8]fixed_t

// C documentation
//
//	// The xtoviewangleangle[] table maps a screen pixel
//	// to the lowest viewangle that maps back to x ranges
//	// from clipangle to -clipangle.
var xtoviewangle [321]angle_t

var ylookup [832]int32

var yslope [200]fixed_t

var yspeed [8]fixed_t

var zlight [16][128][]lighttable_t

func fprintf_ccgo(output io.Writer, str string, args ...any) {
	fmt.Fprintf(output, str, args...)
}

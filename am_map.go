package gore

import (
	"fmt"
	"os"
)

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
var grid int32 = 0

var finit_width int32 = SCREENWIDTH
var finit_height int32 = SCREENHEIGHT - 32

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
	var i, v4 int32
	v1 = INT_MAX1
	min_y = v1
	min_x = v1
	v2 = -INT_MAX1
	max_y = v2
	max_x = v2
	i = 0
	for {
		if i >= numvertexes {
			break
		}
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
		goto _3
	_3:
		;
		i++
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
	var i int32
	i = 0
	for {
		if i >= 10 {
			break
		}
		name := fmt.Sprintf("AMMNUM%d", i)
		marknums[i] = w_CacheLumpNameT(name)
		goto _1
	_1:
		;
		i++
	}
}

func am_unloadPics() {
	var i int32
	i = 0
	for {
		if i >= 10 {
			break
		}
		name := fmt.Sprintf("AMMNUM%d", i)
		w_ReleaseLumpName(name)
		goto _1
	_1:
		;
		i++
	}
}

func am_clearMarks() {
	var i int32
	i = 0
	for {
		if i >= AM_NUMMARKPOINTS {
			break
		}
		markpoints[i].Fx = -1
		goto _1
	_1:
		;
		i++
	} // means empty
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
	f_w = finit_width
	f_h = finit_height
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
													grid = boolint32(grid == 0)
													if grid != 0 {
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
	x = start
	for {
		if x >= end {
			break
		}
		bp.Fa.Fx = x
		bp.Fb.Fx = x
		am_drawMline(bp, color)
		goto _1
	_1:
		;
		x += MAPBLOCKUNITS << FRACBITS
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
	y = start
	for {
		if y >= end {
			break
		}
		bp.Fa.Fy = y
		bp.Fb.Fy = y
		am_drawMline(bp, color)
		goto _2
	_2:
		;
		y += MAPBLOCKUNITS << FRACBITS
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
	if grid != 0 {
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

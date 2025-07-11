package gore

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var storedemo boolean
var show_endoom int32

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

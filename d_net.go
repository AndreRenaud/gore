package gore

import (
	"fmt"
	"os"
)

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

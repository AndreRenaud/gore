package gore

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
	var i, newtics, nowtime int32
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
	i = 0
	for {
		if i >= newtics {
			break
		}
		if buildNewTic() == 0 {
			break
		}
		goto _1
	_1:
		;
		i++
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
	var i uint32
	var keyplayer int32
	keyplayer = -1
	frameon++
	// ideally maketic should be 1 - 3 tics above lowtic
	// if we are consistantly slower, speed up time
	i = 0
	for {
		if i >= NET_MAXPLAYERS {
			break
		}
		if local_playeringame[i] != 0 {
			keyplayer = int32(i)
			break
		}
		goto _1
	_1:
		;
		i++
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
	var i uint32
	var result boolean
	result = 0
	// If we are connected to a server, check if there are any players
	// in the game.
	if net_client_connected != 0 {
		i = 0
		for {
			if i >= NET_MAXPLAYERS {
				break
			}
			result = booluint32(result != 0 || local_playeringame[i] != 0)
			goto _1
		_1:
			;
			i++
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
	var i uint32
	i = 0
	for {
		if i >= NET_MAXPLAYERS {
			break
		}
		cmd := &set.Fcmds[i]
		cmd.Fchatchar = 0
		if int32(cmd.Fbuttons)&bt_SPECIAL != 0 {
			cmd.Fbuttons = 0
		}
		goto _1
	_1:
		;
		i++
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

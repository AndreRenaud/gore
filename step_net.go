package gore

// Multiplayer for the step API.
//
// Vanilla DOOM's netcode is lockstep: every node runs the whole simulation,
// and a tic only runs once every player's ticcmd for it has arrived. A step
// host does the same thing with StepNetTic, feeding the ticcmds of every
// player in the game. Start the engine with "-solo-net" (and optionally
// "-deathmatch") so the netgame rules apply: respawning instead of
// restarting the level, multiplayer-only things, and so on.
//
// Vanilla fixes the set of players when the game starts. Step hosts can
// add and remove players between tics instead. Membership changes are part
// of the tic, so every node applies them at the same moment.

import "fmt"

// NetTic is one tic of a multiplayer game.
type NetTic struct {
	// Cmds holds the input of each player in the game. Entries for players
	// who aren't in the game are ignored.
	Cmds [MAXPLAYERS]TicCmd
	// Leave and Join are bitmasks of players leaving and joining the game
	// before this tic runs. Leaves are applied first, so a slot can be
	// handed from one player to the next in a single tic.
	Leave, Join uint8
}

// stepView is the player the host wants to see (see StepSetView).
var stepView int32

// StepNetTic runs exactly one game tic of a multiplayer game.
func StepNetTic(t *NetTic) {
	stepSoundTic = gametic + 1
	for p := range int32(MAXPLAYERS) {
		if t.Leave&(1<<p) != 0 {
			stepLeave(p)
		}
	}
	var spawning []int32 // players who joined mid-level, and spawn during this tic
	for p := range int32(MAXPLAYERS) {
		if t.Join&(1<<p) != 0 && stepJoin(p) {
			spawning = append(spawning, p)
		}
	}
	stepFixView() // the console player may have just left
	if PlayersInGame() == 0 {
		// Like vanilla (see tryRunTics), don't run tics with nobody in the
		// game: the world waits for someone to join. (Monsters would also
		// look for a target forever, in p_LookForPlayers.)
		return
	}

	var cmds [NET_MAXPLAYERS]ticcmd_t
	for p := range MAXPLAYERS {
		c := t.Cmds[p]
		cmds[p] = ticcmd_t{Fforwardmove: c.Forward, Fsidemove: c.Side, Fangleturn: c.AngleTurn, Fbuttons: c.Buttons}
	}
	// Players come and go through Join and Leave, so runTic's check for
	// players who have quit never applies: everyone in the game is playing.
	runTic(cmds[:], playeringame[:])
	gametic++

	for _, p := range spawning {
		stepTeleportFog(players[p].Fmo)
	}
	stepFixView() // a joining player now has a body to look through
	stepUpdateSounds()
}

// stepJoin puts player p in the game. In a level they spawn during the next
// tic, the way a dead netgame player respawns (see g_DoReborn), and stepJoin
// returns true. Otherwise they spawn when the next level starts.
func stepJoin(p int32) bool {
	if playeringame[p] != 0 {
		return false
	}
	stepClearSlot(p)
	playeringame[p] = 1
	players[consoleplayer].Fmessage = fmt.Sprintf("%sjoined the game", player_names[p])
	if gamestate != gs_LEVEL {
		players[p].Fplayerstate = Pst_DEAD // g_DoLoadLevel turns this into a reborn
		return false
	}
	players[p].Fplayerstate = Pst_REBORN
	return true
}

// stepLeave takes player p out of the game. A living marine teleports out;
// a corpse stays where it fell.
func stepLeave(p int32) {
	if playeringame[p] == 0 {
		return
	}
	if mo := players[p].Fmo; mo != nil {
		mo.Fplayer = nil
		if gamestate == gs_LEVEL && players[p].Fplayerstate == Pst_LIVE {
			stepTeleportFog(mo)
			// p_RemoveMobj doesn't clear other mobjs' pointers to mo.
			// Monsters chasing it give up on targets that aren't
			// shootable, and look for someone else.
			mo.Fflags &^= mf_SHOOTABLE | mf_SOLID
			p_RemoveMobj(mo)
		}
	}
	stepClearSlot(p)
	playeringame[p] = 0
	players[consoleplayer].Fmessage = fmt.Sprintf("%sleft the game", player_names[p])
}

// stepClearSlot resets player slot p, so nothing carries over from its last
// occupant to the next.
func stepClearSlot(p int32) {
	players[p] = player_t{}
	for i := range players {
		players[i].Ffrags[p] = 0
	}
}

// stepTeleportFog spawns teleport fog, and its sound, where a marine joining
// or leaving the game appears or disappears.
func stepTeleportFog(mo *mobj_t) {
	fog := p_SpawnMobj(mo.Fx, mo.Fy, mo.Ffloorz, mt_TFOG)
	s_StartSound(&fog.degenmobj_t, int32(sfx_telept))
}

// stepCheckSpot reports whether a player joining mid-level can spawn at
// mthing. It is g_CheckSpot's test, but the player has no body (not even a
// corpse) to test the spot with, so it uses a stand-in the size of a marine.
func stepCheckSpot(playernum int32, mthing *mapthing_t) boolean {
	if mthing.Ftype1 == 0 {
		return 0 // the map has no such start
	}
	x := int32(mthing.Fx) << FRACBITS
	y := int32(mthing.Fy) << FRACBITS
	info := &mobjinfo[mt_PLAYER]
	standIn := &mobj_t{
		Finfo:   info,
		Fradius: info.Fradius,
		Fheight: info.Fheight,
		Fflags:  mf_SOLID | mf_SHOOTABLE, // not mf_PICKUP: a stand-in can't pick things up
		Fplayer: &players[playernum],
	}
	standIn.Fx, standIn.Fy = x, y
	return p_CheckPosition(standIn, x, y)
}

// StepSetView chooses whose eyes and status bar StepRender draws, and
// whose ears hear the sounds.
//
// This sets consoleplayer, which in vanilla DOOM only affects the user
// interface and sound, never the simulation. So nodes looking through
// different players' eyes stay in sync.
func StepSetView(p int32) {
	stepView = p
	stepFixView()
}

// stepFixView points the view (consoleplayer and displayplayer) at stepView.
// The renderer, status bar and sound code all expect the console player to
// have a body. So until stepView has one, the view stays where it is, or
// moves to the first player who has one.
func stepFixView() {
	switch {
	case stepHasBody(stepView):
		stepSetConsolePlayer(stepView)
	case stepHasBody(consoleplayer):
		// Keep looking through the current player for now.
	default:
		for p := range int32(MAXPLAYERS) {
			if stepHasBody(p) {
				stepSetConsolePlayer(p)
				return
			}
		}
	}
}

func stepHasBody(p int32) bool {
	return p >= 0 && p < MAXPLAYERS && playeringame[p] != 0 && players[p].Fmo != nil
}

func stepSetConsolePlayer(p int32) {
	if p == consoleplayer && p == displayplayer {
		return
	}
	consoleplayer, displayplayer = p, p
	if gamestate == gs_LEVEL {
		// Show the new player's status bar and messages.
		st_Start()
		hu_Start()
	}
}

// PlayersInGame returns a bitmask of the players in the game.
func PlayersInGame() uint8 {
	var mask uint8
	for p := range MAXPLAYERS {
		if playeringame[p] != 0 {
			mask |= 1 << p
		}
	}
	return mask
}

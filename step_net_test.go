package gore

import "testing"

// Bit masks for players in NetTic.Join, NetTic.Leave and PlayersInGame.
const (
	player0Mask = 1 << 0
	player1Mask = 1 << 1
	player2Mask = 1 << 2
)

// TestStepNet joins and removes players mid-level, saves with several
// players in the game, and waits once everyone has left.
func TestStepNet(t *testing.T) {
	stepInitForTest(t, "-solo-net")
	if netgame == 0 || PlayersInGame() != player0Mask {
		t.Fatalf("netgame=%d players=%b, want a netgame with player 0", netgame, PlayersInGame())
	}
	run := func(n int, tic NetTic) {
		for i := range n {
			for p := range tic.Cmds {
				tic.Cmds[p] = TicCmd{Forward: 25, AngleTurn: int16(300 * (p + 1))}
				if i%15 == 0 {
					tic.Cmds[p].Buttons = ButtonAttack
				}
			}
			StepNetTic(&tic)
			tic.Join, tic.Leave = 0, 0
		}
	}
	run(35, NetTic{})
	StepSetView(2) // not in the game yet: keep looking through player 0
	if consoleplayer != 0 {
		t.Fatalf("console player %d, want 0 until player 2 has a body", consoleplayer)
	}
	run(35, NetTic{Join: player1Mask | player2Mask})
	if PlayersInGame() != player0Mask|player1Mask|player2Mask || players[1].Fmo == nil || players[2].Fmo == nil {
		t.Fatalf("players=%b, want 0-2 in the game with bodies", PlayersInGame())
	}
	if consoleplayer != 2 || displayplayer != 2 {
		t.Fatalf("console player %d, display %d; want 2", consoleplayer, displayplayer)
	}
	StepRender()

	save, err := SaveGameBytes()
	if err != nil {
		t.Fatal(err)
	}
	before := Checksum()
	if err := LoadGameBytes(save); err != nil {
		t.Fatal(err)
	}
	if Checksum() != before || PlayersInGame() != player0Mask|player1Mask|player2Mask {
		t.Fatalf("after load: checksum %08x (want %08x), players %b", Checksum(), before, PlayersInGame())
	}

	// Player 2 leaves; the view falls back to someone still playing.
	body := players[2].Fmo
	run(10, NetTic{Leave: player2Mask})
	if PlayersInGame() != player0Mask|player1Mask || players[2].Fmo != nil || body.Fplayer != nil {
		t.Fatalf("players=%b after player 2 left", PlayersInGame())
	}
	if consoleplayer == 2 {
		t.Fatalf("still viewing player 2 after they left")
	}
	// The slot can be handed straight to someone else.
	run(10, NetTic{Leave: player1Mask, Join: player1Mask})
	if players[1].Fmo == nil || players[1].Fkillcount != 0 {
		t.Fatalf("player 1 didn't respawn fresh")
	}
	StepRender()
	if _, err := SaveGameBytes(); err != nil {
		t.Fatal(err)
	}

	// With nobody left, the world waits, and renders a black frame.
	tic := GetStats().Tic
	run(10, NetTic{Leave: player0Mask | player1Mask})
	if PlayersInGame() != 0 || GetStats().Tic != tic {
		t.Fatalf("players=%b, tic %d after everyone left; want none, tic %d", PlayersInGame(), GetStats().Tic, tic)
	}
	StepRender()
	for _, b := range FrameBuffer() {
		if b != 0 {
			t.Fatalf("frame isn't black with nobody in the game")
		}
	}
}

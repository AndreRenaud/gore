package gore

import (
	"os"
	"testing"
	"testing/fstest"
)

// stepInitForTest starts the engine in step mode on E1M1, or skips the test
// if there is no WAD. The WAD is doom1.wad, or $DOOM_WAD if set.
func stepInitForTest(t *testing.T, args ...string) {
	t.Helper()
	wadPath := os.Getenv("DOOM_WAD")
	if wadPath == "" {
		wadPath = "doom1.wad"
	}
	wad, err := os.ReadFile(wadPath)
	if err != nil {
		t.Skip("doom1.wad not present")
	}
	SetVirtualFileSystem(fstest.MapFS{"doom1.wad": {Data: wad}})
	StepInit(nil, append([]string{"-iwad", "doom1.wad", "-warp", "1", "1", "-skill", "3"}, args...))
}

func TestStep(t *testing.T) {
	stepInitForTest(t)
	for i := 0; i < 350; i++ {
		cmd := TicCmd{Forward: 25} // walk forward
		if i%70 < 20 {
			cmd.AngleTurn = 640 // and turn left at walking speed, 3.5 degrees a tic
		}
		if i%10 == 0 {
			cmd.Buttons = ButtonAttack
		}
		StepTic(cmd)
	}
	StepRender()
	s := GetStats()
	t.Logf("stats: %+v", s)
	if s.Tic != 350 || s.GameState != "level" {
		t.Fatalf("unexpected stats %+v", s)
	}
	nonzero := 0
	for _, b := range FrameBuffer() {
		if b != 0 {
			nonzero++
		}
	}
	if nonzero < 10000 {
		t.Fatalf("frame looks empty: %d nonzero pixels", nonzero)
	}
	save, err := SaveGameBytes()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("savegame: %d bytes", len(save))
	before := Checksum()
	if err := LoadGameBytes(save); err != nil {
		t.Fatal(err)
	}
	if GetStats().Tic != 350 {
		t.Fatalf("tic not restored")
	}
	if Checksum() != before {
		t.Fatalf("checksum %08x after load, want %08x", Checksum(), before)
	}
}

// TestSaveEveryTic saves after every tic of a few minutes of play, to
// exercise saving in as many different game states as possible.
func TestSaveEveryTic(t *testing.T) {
	stepInitForTest(t)
	saves := 0
	for i := 0; i < 3000; i++ {
		cmd := TicCmd{Forward: 50} // run forward
		if i%140 < 20 {
			cmd.AngleTurn = 1280 // and turn left at running speed, 7 degrees a tic
		}
		if i%9 == 0 {
			cmd.Buttons |= ButtonAttack
		}
		if i%20 == 0 {
			cmd.Buttons |= ButtonUse
		}
		StepTic(cmd)
		if CanSave() {
			if _, err := SaveGameBytes(); err != nil {
				t.Fatalf("tic %d: %v", i, err)
			}
			saves++
		}
	}
	t.Logf("%d saves, final stats %+v", saves, GetStats())
}

// TestStepSound checks that sounds are reported with the right tic, and that
// channels are freed when sounds finish, so sounds keep coming.
func TestStepSound(t *testing.T) {
	stepInitForTest(t)
	StepSoundEvents() // level start
	counts := map[string]int{}
	lastPistol := int32(0)
	for i := 0; i < 700; i++ {
		cmd := TicCmd{}
		if i%20 == 0 {
			cmd.Buttons = ButtonAttack
		}
		StepTic(cmd)
		for _, ev := range StepSoundEvents() {
			if ev.Tic != int32(i+1) {
				t.Fatalf("event tic %d, want %d: %+v", ev.Tic, i+1, ev)
			}
			counts[ev.Kind+":"+ev.Name]++
			if ev.Name == "pistol" {
				lastPistol = ev.Tic
			}
		}
	}
	t.Logf("sound events: %v", counts)
	// Every bullet fired must be heard, right up to the end. Before step
	// mode had a sound module, channels were never freed.
	fired := 50 - GetStats().Ammo
	if fired < 30 || counts["start:pistol"] != int(fired) || lastPistol < 680 {
		t.Fatalf("fired %d bullets but heard %d pistol shots (last at tic %d)", fired, counts["start:pistol"], lastPistol)
	}
}

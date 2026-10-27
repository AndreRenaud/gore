// stepper drives the engine with the step API: no clock and no keyboard.
// Instead, each tic gets a ticcmd from a fixed script: walk forward, turn
// left for a moment every two seconds, and fire once a second. The output
// depends only on the WAD and the flags, so running it twice prints the
// same checksums.
//
//	go run ./example/stepper -iwad doom1.wad -tics 700 -o frame.png
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/AndreRenaud/gore"
)

func main() {
	iwad := flag.String("iwad", "doom1.wad", "IWAD to play")
	tics := flag.Int("tics", 700, "number of tics to run (35 per second)")
	out := flag.String("o", "", "write the final frame to this PNG file")
	flag.Parse()

	// Start on E1M1, at skill 3 ("Hurt Me Plenty").
	gore.StepInit(nil, []string{"-iwad", *iwad, "-warp", "1", "1", "-skill", "3"})

	sounds := 0
	for i := range *tics {
		cmd := gore.TicCmd{Forward: 25} // walk forward
		if i%70 < 20 {
			cmd.AngleTurn = 640 // turn left at walking speed
		}
		// Fire once a second. A single shot of the pistol takes about half
		// a second, and presses while it is firing are ignored.
		if i%gore.TICRATE == 0 {
			cmd.Buttons = gore.ButtonAttack
		}
		gore.StepTic(cmd)
		sounds += len(gore.StepSoundEvents())
		if (i+1)%gore.TICRATE == 0 {
			s := gore.GetStats()
			fmt.Printf("tic %4d: health %3d ammo %3d kills %d/%d checksum %08x\n",
				s.Tic, s.Health, s.Ammo, s.Kills, s.TotalKills, s.Checksum)
		}
	}
	fmt.Printf("%d sound events\n", sounds)

	// A savegame restores the exact simulation state. Saving isn't possible
	// in every state (e.g. while dead), so skip the check if need be.
	if !gore.CanSave() {
		fmt.Println("can't save right now; skipping the save/load check")
	} else {
		before := gore.Checksum()
		save, err := gore.SaveGameBytes()
		if err != nil {
			log.Fatal(err)
		}
		if err := gore.LoadGameBytes(save); err != nil {
			log.Fatal(err)
		}
		if after := gore.Checksum(); after != before {
			log.Fatalf("checksum %08x after loading, want %08x", after, before)
		}
		fmt.Printf("saved and reloaded %d bytes; checksum unchanged\n", len(save))
	}

	// Optionally save a screenshot, to see where the script ended up.
	if *out != "" {
		gore.StepRender()
		if err := writePNG(*out, gore.FrameBuffer(), gore.Palette()); err != nil {
			log.Fatal(err)
		}
	}
}

// writePNG writes a palette-indexed frame as a PNG.
func writePNG(name string, frame, palette []byte) error {
	pal := make(color.Palette, 256)
	for i := range pal {
		pal[i] = color.RGBA{palette[i*3], palette[i*3+1], palette[i*3+2], 255}
	}
	img := image.NewPaletted(image.Rect(0, 0, gore.SCREENWIDTH, gore.SCREENHEIGHT), pal)
	copy(img.Pix, frame)
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

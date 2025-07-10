package gore

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	diff "github.com/olegfedoseev/image-diff"
)

type delayedEvent struct {
	ticks       int32 // How many game ticks before we trigger this event, since the last one
	event       DoomEvent
	autoRelease bool // If set, auto release 1 tick later
	callback    func(*doomTestHeadless)
}

type doomTestHeadless struct {
	t             *testing.T
	keys          []delayedEvent
	lastEventTick int32
	outputFile    io.WriteCloser
	lock          sync.Mutex
	lastImage     *image.RGBA
}

func (d *doomTestHeadless) Close() {
	if err := d.outputFile.Close(); err != nil {
		d.t.Errorf("Error closing output file: %v", err)
	}
}

type bufferedWriteCloser struct {
	*bufio.Writer
	io.Closer
}

func (b *bufferedWriteCloser) Close() error {
	if err := b.Writer.Flush(); err != nil {
		return fmt.Errorf("error flushing buffer: %w", err)
	}
	return b.Closer.Close()
}

func ffmpegSaver(filename string) (io.WriteCloser, error) {
	args := []string{
		"ffmpeg",
		"-y", // Overwrite output file if it exists
		"-loglevel", "error",
		"-hide_banner",
		"-f", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", SCREENWIDTH, SCREENHEIGHT),
		"-r", "29", // Frame rate - 35ms ticks?
		"-pix_fmt", "rgba",
		"-i", "-",
		"-crf", "27",
		"-preset", "veryfast",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-flush_packets", "1",
		"-movflags", "+faststart",
		filename,
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	// We don't want ffmpeg compression to slow down the game, so just use a lot of memory instead
	return &bufferedWriteCloser{
		Writer: bufio.NewWriterSize(stdin, 256*1024*1024),
		Closer: stdin,
	}, nil
}

func (d *doomTestHeadless) DrawFrame(frame *image.RGBA) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.outputFile == nil {
		var err error
		name := fmt.Sprintf("doom_test_%s.mp4", d.t.Name())
		d.outputFile, err = ffmpegSaver(name)
		if err != nil {
			d.t.Fatalf("Error starting ffmpeg: %v", err)
		}
		d.t.Logf("Saving output to %s", name)
	}
	d.outputFile.Write(frame.Pix)
	if d.lastImage == nil {
		d.lastImage = image.NewRGBA(frame.Rect)
	}
	draw.Draw(d.lastImage, d.lastImage.Rect, frame, frame.Bounds().Min, draw.Src)
}

func (d *doomTestHeadless) SetTitle(title string) {
	d.t.Logf("SetTitle called with: %s", title)
}

func (d *doomTestHeadless) GetScreen() *image.RGBA {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.lastImage == nil {
		return nil
	}
	// Return a copy of the last image
	screenCopy := image.NewRGBA(d.lastImage.Rect)
	draw.Draw(screenCopy, screenCopy.Rect, d.lastImage, d.lastImage.Bounds().Min, draw.Src)
	return screenCopy
}

func (d *doomTestHeadless) GetEvent(event *DoomEvent) bool {
	if len(d.keys) == 0 {
		return false
	}
	if d.lastEventTick == 0 {
		d.lastEventTick = I_GetTimeMS()
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	now := I_GetTimeMS()
	delta := now - d.lastEventTick
	if d.keys[0].ticks > delta {
		return false
	}
	retval := false
	if d.keys[0].callback != nil {
		callback := d.keys[0].callback
		d.lock.Unlock()
		callback(d)
		d.lock.Lock()
	}
	if d.keys[0].event.Key != 0 {
		*event = d.keys[0].event
		retval = true
	}
	//d.t.Logf("Key event: %#v, delta=%d (%d remaining)", *event, delta, len(d.keys)-1)
	if d.keys[0].autoRelease && d.keys[0].event.Type == Ev_keydown {
		// If it's a down & auto-release, just replace it with the up event 1-tick later
		d.keys[0].event.Type = Ev_keyup
		d.keys[0].autoRelease = false
		d.keys[0].ticks = 1
	} else {
		d.keys = d.keys[1:]
	}
	d.lastEventTick = now
	return retval
}

// InsertKey simulates an immediate key press and release event in the game.
func (d *doomTestHeadless) InsertKey(key uint8) {
	d.InsertKeySequence(key)
}

// Insert a series of key presses and releases, and wait for them to be processed.
func (d *doomTestHeadless) InsertKeySequence(keys ...uint8) {
	d.lock.Lock()
	for _, key := range keys {
		// Insert a key press and release for each key
		d.keys = append(d.keys, delayedEvent{
			event: DoomEvent{
				Type: Ev_keydown,
				Key:  key,
			},
			ticks:       1,
			autoRelease: true,
		})
	}
	d.lock.Unlock()
	// Wait for the last key event to be processed
	for {
		d.lock.Lock()
		inuse := len(d.keys) > 0
		d.lock.Unlock()
		time.Sleep(100 * time.Microsecond) // Wait a bit before checking again
		if !inuse {
			break
		}
	}
}

func (d *doomTestHeadless) InsertKeyChange(Key uint8, pressed bool) {
	d.lock.Lock()
	evType := Ev_keyup
	if pressed {
		evType = Ev_keydown
	}
	d.keys = append(d.keys, delayedEvent{
		event: DoomEvent{
			Type: evType,
			Key:  Key,
		},
		ticks: 0, // Insert immediately
	})
	d.lock.Unlock()
	// Wait for it to leave the queue
	for {
		d.lock.Lock()
		inuse := len(d.keys) > 0
		d.lock.Unlock()
		time.Sleep(100 * time.Microsecond) // Wait a bit before checking again
		if !inuse {
			break
		}
	}
}

// Run the demo at super speed to make sure it all goes ok
func TestDoomDemo(t *testing.T) {
	dg_run_full_speed = true
	game := &doomTestHeadless{
		t: t,
	}
	defer game.Close()
	go func() {
		time.Sleep(2 * time.Second)

		// Quit
		Stop()
	}()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func savePNG(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating PNG file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("error encoding PNG: %w", err)
	}
	return nil
}

func loadPNG(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening PNG file: %w", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding PNG: %w", err)
	}

	return img, nil
}

func TestLoadSave(t *testing.T) {
	dg_run_full_speed = true
	var imgPlayedGame, imgNewGame, imgLoadedGame *image.RGBA
	game := &doomTestHeadless{
		t: t,
		keys: []delayedEvent{
			// Start a new game
			{ticks: 50, event: DoomEvent{Type: Ev_keydown, Key: KEY_ESCAPE}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},

			// Run straight into the opposing wall - to make the screen as different as possible
			{ticks: 10, event: DoomEvent{Type: Ev_keydown, Key: KEY_UPARROW1}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: uint8(key_speed)}},
			{ticks: 5000, event: DoomEvent{Type: Ev_keyup, Key: KEY_UPARROW1}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: uint8(key_speed)}},

			// Grab a screenshot (but give it some time to slow down)
			{ticks: 500, callback: func(d *doomTestHeadless) { imgPlayedGame = d.GetScreen() }},

			// Go to the menu and save
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ESCAPE}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_DOWNARROW1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_DOWNARROW1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_DOWNARROW1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},

			// Clear the old name
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_BACKSPACE1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_BACKSPACE1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_BACKSPACE1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_BACKSPACE1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_BACKSPACE1}, autoRelease: true},

			// Enter a new name
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 't'}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'e'}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 's'}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 't'}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},

			// Start a new game
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ESCAPE}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_UPARROW1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_UPARROW1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_UPARROW1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},

			// Grab a screenshot of the new game
			{ticks: 500, callback: func(d *doomTestHeadless) { imgNewGame = d.GetScreen() }},

			// Load the saved game
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ESCAPE}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_DOWNARROW1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_DOWNARROW1}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}, autoRelease: true},

			// Grab a screenshot after loading the save
			{ticks: 500, callback: func(d *doomTestHeadless) { imgLoadedGame = d.GetScreen() }},

			// Compare all the screenshots - before saving should match post-load, and new game should be different
			{ticks: 10, callback: func(d *doomTestHeadless) {
				// Compare the screenshots
				diffImg, percent, err := diff.CompareImages(imgPlayedGame, imgLoadedGame)
				if err != nil {
					d.t.Errorf("save/load comparison failed: %v", err)
					return
				}
				d.t.Logf("Load game screenshot comparison: %f%% difference", percent)
				if percent > 2 { // Allow a small margin of error
					savePNG("doom_test_screenshot1.png", imgPlayedGame)
					savePNG("doom_test_screenshot2.png", imgLoadedGame)
					savePNG("doom_test_diff.png", diffImg)
					d.t.Errorf("Screenshots do not match after loading save: %f%% difference", percent)
				}

				diffImg, percent, err = diff.CompareImages(imgPlayedGame, imgNewGame)
				if err != nil {
					t.Errorf("new game comparison failed: %v", err)
				}
				t.Logf("New game screenshot comparison: %f%% difference", percent)
				if percent < 70 { // They should be different, so allow a very large margin of error
					savePNG("doom_test_screenshot1.png", imgPlayedGame)
					savePNG("doom_test_screenshot_new.png", imgNewGame)
					t.Errorf("New game screenshot matches the original: %f%% difference", percent)
				}
			}},
			{ticks: 1000, callback: func(d *doomTestHeadless) { Stop() }},
		},
	}
	defer game.Close()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func TestDoomRandom(t *testing.T) {
	dg_run_full_speed = true
	game := &doomTestHeadless{
		t: t,
	}
	defer game.Close()
	go func() {
		// Let things get settled
		time.Sleep(20 * time.Millisecond)
		// Start a game
		game.InsertKey(KEY_ESCAPE)     // Open menu
		game.InsertKey(KEY_ENTER)      // New Game
		game.InsertKey(KEY_ENTER)      // Knee-Deep in the Dead
		game.InsertKey(KEY_DOWNARROW1) // Ultra Violence
		game.InsertKey(KEY_ENTER)      // Start new game

		time.Sleep(10 * time.Millisecond)
		keys := []uint8{
			KEY_UPARROW1, KEY_UPARROW1, KEY_UPARROW1, KEY_UPARROW1, // Make forward movement more likely
			KEY_DOWNARROW1, KEY_LEFTARROW1,
			KEY_RIGHTARROW1, KEY_FIRE1, KEY_USE1,
		}
		// Press shift to run
		game.InsertKeyChange(0x80+0x36, true)
		// Do some random movement
		count := 5000
		for i := range count {
			key := keys[rand.Intn(len(keys))]
			game.InsertKeyChange(key, true)
			time.Sleep(1 * time.Millisecond)
			game.InsertKeyChange(key, false)
			if i%100 == 0 {
				t.Logf("%d/%d done", i, count)
			}
		}

		// Exit
		game.InsertKey(KEY_ESCAPE)   // Open menu
		game.InsertKey(KEY_UPARROW1) // Go to quit
		game.InsertKey(KEY_ENTER)    // Confirm quit
		game.InsertKey('y')          // Confirm exit
	}()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func compareScreen(game *doomTestHeadless, testdataPrefix string, percentOk float64) {
	screen := game.GetScreen()
	if screen == nil {
		game.t.Errorf("No screen captured for %s", filename)
		return
	}
	// Save the screenshot for debugging
	if err := savePNG(fmt.Sprintf("doom_test_%s.png", testdataPrefix), screen); err != nil {
		game.t.Errorf("Error saving screenshot: %v", err)
	}

	knownGood, err := loadPNG(fmt.Sprintf("testdata/good_doom_test_%s.png", testdataPrefix))
	if err != nil {
		game.t.Errorf("Error loading known good image: %v", err)
		return
	}

	diffImg, percent, err := diff.CompareImages(screen, knownGood)
	if err != nil {
		game.t.Errorf("Error comparing screenshot: %v", err)
		return
	}
	if percent > percentOk {
		game.t.Errorf("Screenshot %s does not match known good: %f%% difference (over %f%%)", testdataPrefix, percent, percentOk)
		savePNG(fmt.Sprintf("doom_test_%s_diff.png", testdataPrefix), diffImg)
	}
	game.t.Logf("Screenshot %s comparison: %f%% difference (allowed: %f%%)", testdataPrefix, percent, percentOk)
}

func TestDoomLevels(t *testing.T) {
	dg_run_full_speed = true
	var game *doomTestHeadless
	game = &doomTestHeadless{
		t: t,
		keys: []delayedEvent{
			{ticks: 1500, callback: func(d *doomTestHeadless) { compareScreen(d, "start", 2) }},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ESCAPE}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: KEY_ESCAPE}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: KEY_ENTER}},
		},
	}
	defer game.Close()
	for i := 1; i <= 9; i++ {
		game.keys = append(game.keys, []delayedEvent{
			{ticks: 100, event: DoomEvent{Type: Ev_keydown, Key: 'i'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'i'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'd'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'd'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'c'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'c'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'l'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'l'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'e'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'e'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'v'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'v'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: '1'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: '1'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: '0' + byte(i)}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: '0' + byte(i)}},
			{ticks: 2000, callback: func(d *doomTestHeadless) { compareScreen(d, fmt.Sprintf("e1m%d", i), 15) }},
		}...)
	}
	// Quit the game
	game.keys = append(game.keys, delayedEvent{ticks: 1000, callback: func(d *doomTestHeadless) { Stop() }})
	Run(game, []string{"-iwad", "doom1.wad"})
}

func TestDoomMap(t *testing.T) {
	dg_run_full_speed = true
	game := &doomTestHeadless{
		t: t,
	}
	defer game.Close()
	go func() {
		// Let things get settled
		time.Sleep(20 * time.Millisecond)
		// Start a game
		game.InsertKey(KEY_ESCAPE) // Open menu
		game.InsertKey(KEY_ENTER)
		game.InsertKey(KEY_ENTER)
		game.InsertKey(KEY_ENTER) // Start new game
		time.Sleep(10 * time.Millisecond)

		// Move a bit
		game.InsertKeyChange(KEY_UPARROW1, true) // Move up
		time.Sleep(10 * time.Millisecond)        // Move up for a bit
		game.InsertKeyChange(KEY_TAB, true)      // Open map
		time.Sleep(10 * time.Millisecond)
		game.InsertKeyChange(KEY_UPARROW1, false)  // Stop moving
		time.Sleep(10 * time.Millisecond)          // Wait a bit
		game.InsertKeyChange(KEY_LEFTARROW1, true) // Turn for a bit
		time.Sleep(10 * time.Millisecond)
		game.InsertKeyChange(KEY_LEFTARROW1, false) // Turn for a bit
		time.Sleep(10 * time.Millisecond)
		game.InsertKeyChange(KEY_TAB, false) // close map

		// Exit
		game.InsertKey(KEY_ESCAPE)   // Open menu
		game.InsertKey(KEY_UPARROW1) // Go to quit
		game.InsertKey(KEY_ENTER)    // Confirm quit
		game.InsertKey('y')          // Confirm exit
	}()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func TestWeapons(t *testing.T) {
	dg_run_full_speed = true
	game := &doomTestHeadless{
		t: t,
		keys: []delayedEvent{
			// Start a new game, and turn on all the weapons
			{ticks: 1500, event: DoomEvent{Type: Ev_keydown, Key: KEY_ESCAPE}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: KEY_ESCAPE}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: KEY_ENTER}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: KEY_ENTER}},
			{ticks: 100, event: DoomEvent{Type: Ev_keydown, Key: 'i'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'i'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'd'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'd'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'f'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'f'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keydown, Key: 'a'}},
			{ticks: 1, event: DoomEvent{Type: Ev_keyup, Key: 'a'}},
		},
	}
	// Cycle each weapon, get a screenshot to confirm it shows up, then fire it
	// BFG & Plasma gun aren't available in the shareware wad, so we only test 1-5
	// - Chainsaw, Pistol, Shotgun, Machine Gun, Rocket Launcher
	for i := byte('1'); i <= '5'; i++ {
		game.keys = append(game.keys, []delayedEvent{
			{ticks: 300, event: DoomEvent{Type: Ev_keydown, Key: i}},
			{ticks: 5, event: DoomEvent{Type: Ev_keyup, Key: i}},
			{ticks: 2000, callback: func(d *doomTestHeadless) {
				t.Logf("Enabled weapon %c", i)
				compareScreen(d, fmt.Sprintf("weapon_%c", i), 5)
			}},
			{ticks: 50, event: DoomEvent{Type: Ev_keydown, Key: KEY_FIRE1}},
			{ticks: 300, event: DoomEvent{Type: Ev_keyup, Key: KEY_FIRE1}},
			{ticks: 10, event: DoomEvent{}},
		}...)
	}
	// Quit the game
	game.keys = append(game.keys, delayedEvent{ticks: 1000, callback: func(d *doomTestHeadless) { Stop() }})
	defer game.Close()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func confirmMenu(t *testing.T, game *doomTestHeadless, name string) {
	time.Sleep(1 * time.Millisecond)
	screen := game.GetScreen()
	if screen == nil {
		t.Errorf("No screen captured for %s", name)
		return
	}
	// Save the screenshot for debugging
	//if err := savePNG(fmt.Sprintf("doom_test_menu_%s.png", name), screen); err != nil {
	//t.Errorf("Error saving menu screenshot: %v", err)
	//}

	knownGoodMenuImage, err := loadPNG(fmt.Sprintf("testdata/good_doom_test_menu_%s.png", name))
	if err != nil {
		t.Errorf("Error loading known good menu image for %s: %v", name, err)
		return
	}

	diff, percent, err := diff.CompareImages(screen, knownGoodMenuImage)

	if err != nil {
		t.Errorf("Error comparing menu screenshot for %s: %v", name, err)
	}
	if percent > 2 { // Allow a small margin of error
		t.Errorf("Menu screenshot for %s does not match known good: %f%% difference", name, percent)
		// Save the diff image for debugging
		savePNG(fmt.Sprintf("doom_test_menu_diff_%s.png", name), diff)
		savePNG(fmt.Sprintf("doom_test_menu_screenshot_%s.png", name), screen)
	}
}

// TestMenus walks through the menus and checks the screenshots
func TestMenus(t *testing.T) {
	dg_run_full_speed = true
	game := &doomTestHeadless{
		t: t,
	}
	defer game.Close()
	// Disable the demo playback, since it messes with the screenshots
	dont_run_demo = true

	go func() {
		// Wait for screen wipe
		time.Sleep(5 * time.Millisecond)
		for wipe_running != 0 {
			time.Sleep(1 * time.Millisecond)
		}
		time.Sleep(5 * time.Millisecond)

		game.InsertKey(KEY_ESCAPE) // Open menu
		confirmMenu(t, game, "main")

		// Go to the options menu
		game.InsertKey(KEY_DOWNARROW1)
		game.InsertKey(KEY_ENTER)
		confirmMenu(t, game, "options")

		// Go to the load menu
		game.InsertKey(KEY_ESCAPE)
		game.InsertKey(KEY_ESCAPE)
		game.InsertKey(KEY_DOWNARROW1)
		game.InsertKey(KEY_ENTER)
		confirmMenu(t, game, "load")

		// Quit
		Stop()
	}()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func TestForceLevel(t *testing.T) {
	dg_run_full_speed = true
	game := &doomTestHeadless{
		t: t,
		keys: []delayedEvent{
			// Start a new game
			{ticks: 1, callback: func(d *doomTestHeadless) {
				StartLevel(1, 6, sk_medium) // Start E1M1 on medium skill
			}},
			{ticks: 100, callback: func(d *doomTestHeadless) {
				// Wait for the level to load and take a screenshot
				compareScreen(d, "e1m6", 5)
			}},
			{ticks: 1000, callback: func(d *doomTestHeadless) {
				// Quit the game
				Stop()
			}},
		},
	}
	defer game.Close()
	Run(game, []string{"-iwad", "doom1.wad"})
}

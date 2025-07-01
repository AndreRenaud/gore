package gore

import (
	"bufio"
	"fmt"
	"image"
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
	ticks    int32 // How many game ticks before we trigger this event, since the last one
	event    DoomEvent
	callback func(*doomTestHeadless)
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
	copy(d.lastImage.Pix, frame.Pix)
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
	copy(screenCopy.Pix, d.lastImage.Pix)
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
		d.lock.Unlock()
		d.keys[0].callback(d)
		d.lock.Lock()
	}
	if d.keys[0].event.Key != 0 {
		*event = d.keys[0].event
		retval = true
	}
	//d.t.Logf("Key event: %#v, delta=%d (%d remaining)", *event, delta, len(d.keys)-1)
	d.keys = d.keys[1:]
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
			ticks: 1,
		},
			delayedEvent{
				event: DoomEvent{
					Type: Ev_keyup,
					Key:  key,
				},
				ticks: 1,
			},
		)
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
	dg_speed_ratio = 100.0
	game := &doomTestHeadless{
		t: t,
	}
	defer game.Close()
	go func() {
		time.Sleep(2 * time.Second)

		// Quit
		D_Endoom()
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
	dg_speed_ratio = 100.0
	game := &doomTestHeadless{
		t: t,
	}
	go func() {
		time.Sleep(20 * time.Millisecond) // Let things get settled
		// Start a new game
		game.InsertKeySequence(KEY_ESCAPE, KEY_ENTER, KEY_ENTER, KEY_ENTER)
		time.Sleep(5 * time.Millisecond) // Wait for the game to start
		// Move around in the game a bit
		game.InsertKeyChange(KEY_UPARROW1, true)
		time.Sleep(20 * time.Millisecond) // Move up for a bit
		game.InsertKeyChange(KEY_UPARROW1, false)
		game.InsertKeyChange(KEY_LEFTARROW1, true)
		time.Sleep(20 * time.Millisecond) // Move left for a bit
		game.InsertKeyChange(KEY_LEFTARROW1, false)
		time.Sleep(10 * time.Millisecond)
		// Grab a screenshot
		img1 := game.GetScreen()
		// Go to the menu and save
		game.InsertKeySequence(KEY_ESCAPE, KEY_DOWNARROW1, KEY_DOWNARROW1, KEY_DOWNARROW1, KEY_ENTER, KEY_ENTER)
		// Clear the old name
		game.InsertKeySequence(KEY_BACKSPACE1, KEY_BACKSPACE1, KEY_BACKSPACE1, KEY_BACKSPACE1, KEY_BACKSPACE1, KEY_BACKSPACE1)
		// Enter a new name
		game.InsertKeySequence('t', 'e', 's', 't', KEY_ENTER)
		// Start a new game
		game.InsertKeySequence(KEY_ESCAPE, KEY_UPARROW1, KEY_UPARROW1, KEY_UPARROW1, KEY_ENTER, KEY_ENTER, KEY_ENTER) // Open menu
		time.Sleep(10 * time.Millisecond)                                                                             // Wait for the game to start
		// New games must have a different screenshot
		imgNew := game.GetScreen()
		// Load the saved game
		game.InsertKeySequence(KEY_ESCAPE, KEY_DOWNARROW1, KEY_DOWNARROW1, KEY_ENTER, KEY_ENTER)
		time.Sleep(10 * time.Millisecond) // Wait for the game to load
		img2 := game.GetScreen()
		// Check if the images are the same
		diffImg, percent, err := diff.CompareImages(img1, img2)
		if err != nil {
			t.Errorf("save/load comparison failed: %v", err)
		}
		t.Logf("Load game screenshot comparison: %f%% difference", percent)
		if percent > 2 { // Allow a small margin of error
			savePNG("doom_test_screenshot1.png", img1)
			savePNG("doom_test_screenshot2.png", img2)
			savePNG("doom_test_diff.png", diffImg)
			t.Errorf("Screenshots do not match after loading save: %f%% difference", percent)
		}

		diffImg, percent, err = diff.CompareImages(img1, imgNew)
		if err != nil {
			t.Errorf("new game comparison failed: %v", err)
		}
		t.Logf("New game screenshot comparison: %f%% difference", percent)
		if percent < 50 { // They should be different, so allow a very large margin of error
			savePNG("doom_test_screenshot1.png", img1)
			savePNG("doom_test_screenshot_new.png", imgNew)
			t.Errorf("New game screenshot matches the original: %f%% difference", percent)
		}

		// Quit
		game.InsertKeySequence(KEY_ESCAPE, KEY_DOWNARROW1, KEY_DOWNARROW1, KEY_DOWNARROW1, KEY_ENTER, 'y')
	}()
	defer game.Close()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func TestDoomRandom(t *testing.T) {
	dg_speed_ratio = 200.0
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
		keys := []uint8{
			KEY_UPARROW1, KEY_DOWNARROW1, KEY_LEFTARROW1,
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
	dg_speed_ratio = 100.0
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
	game.keys = append(game.keys, delayedEvent{ticks: 1000, callback: func(d *doomTestHeadless) { D_Endoom() }})
	Run(game, []string{"-iwad", "doom1.wad"})
}

func TestDoomMap(t *testing.T) {
	dg_speed_ratio = 100.0
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
	dg_speed_ratio = 100.0
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
		game.InsertKeySequence('i', 'd', 'f', 'a') // Give all weapons

		// Test weapon switching - BFG & Plasma gun aren't available in the shareware wad
		for i := uint8(1); i <= 5; i++ {
			game.InsertKey('0' + i)
			time.Sleep(10 * time.Millisecond) // Wait to simulate weapon switch
			// Fire
			game.InsertKeyChange(KEY_FIRE1, true)
			time.Sleep(20 * time.Millisecond) // Wait to simulate firing
			game.InsertKeyChange(KEY_FIRE1, false)
			t.Logf("Switched to weapon %d", i)
		}

		// Exit
		game.InsertKey(KEY_ESCAPE)   // Open menu
		game.InsertKey(KEY_UPARROW1) // Go to quit
		game.InsertKey(KEY_ENTER)    // Confirm quit
		game.InsertKey('y')          // Confirm exit
	}()
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
	dg_speed_ratio = 100.0
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
		D_Endoom()
	}()
	Run(game, []string{"-iwad", "doom1.wad"})
}

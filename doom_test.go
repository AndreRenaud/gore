package gore

import (
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

type delayedKeyEvent struct {
	event DoomKeyEvent
	ticks int32 // How many game ticks before we trigger this event, since the last one
}

type doomTestHeadless struct {
	start         time.Time
	t             *testing.T
	keys          []delayedKeyEvent
	lastEventTick int32
	outputFile    io.WriteCloser
	lock          sync.Mutex
	lastImage     *image.RGBA
}

func (d *doomTestHeadless) Close() {
	if d.outputFile != nil {
		if err := d.outputFile.Close(); err != nil {
			d.t.Errorf("Error closing output file: %v", err)
		}
	}
}

/*
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
*/
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
	return stdin, nil
	// We don't want ffmpeg compression to slow down the game, so just use a lot of memory instead
	//return &bufferedWriteCloser{
	//Writer: bufio.NewWriterSize(stdin, 10*1024*1024),
	//Closer: stdin,
	//}, nil
}

func (d *doomTestHeadless) DrawFrame(frame *image.RGBA) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if true {
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
	}
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

func (d *doomTestHeadless) GetKey(event *DoomKeyEvent) bool {
	if len(d.keys) == 0 {
		return false
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	now := I_GetTimeMS()
	delta := now - d.lastEventTick
	if d.keys[0].ticks <= delta {
		*event = d.keys[0].event
		//d.t.Logf("Key event: %#v, delta=%d (%d remaining)", *event, delta, len(d.keys)-1)
		d.keys = d.keys[1:]
		d.lastEventTick = now
		return true
	}
	return false
}

// InsertKey simulates an immediate key press and release event in the game.
func (d *doomTestHeadless) InsertKey(key uint8) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.keys = append(d.keys, delayedKeyEvent{
		event: DoomKeyEvent{
			Pressed: true,
			Key:     key,
		},
		ticks: 1,
	},
		delayedKeyEvent{
			event: DoomKeyEvent{
				Pressed: false,
				Key:     key,
			},
			ticks: 1,
		},
	)
}

func (d *doomTestHeadless) InsertKeySequence(keys ...uint8) {
	d.lock.Lock()
	for _, key := range keys {
		// Insert a key press and release for each key
		d.keys = append(d.keys, delayedKeyEvent{
			event: DoomKeyEvent{
				Pressed: true,
				Key:     key,
			},
			ticks: 2,
		},
			delayedKeyEvent{
				event: DoomKeyEvent{
					Pressed: false,
					Key:     key,
				},
				ticks: 2,
			},
		)
	}
	d.lock.Unlock()
	// Wait for the last key event to be processed
	for {
		d.lock.Lock()
		inuse := len(d.keys) > 0
		d.lock.Unlock()
		time.Sleep(1 * time.Millisecond) // Wait a bit before checking again
		if !inuse {
			break
		}
	}
}

func (d *doomTestHeadless) InsertKeyChange(Key uint8, pressed bool) {
	d.lock.Lock()
	d.keys = append(d.keys, delayedKeyEvent{
		event: DoomKeyEvent{
			Pressed: pressed,
			Key:     Key,
		},
		ticks: 0, // Insert immediately
	})
	d.lock.Unlock()
	// Wait for it to leave the queue
	for {
		d.lock.Lock()
		inuse := len(d.keys) > 0
		d.lock.Unlock()
		time.Sleep(1 * time.Millisecond) // Wait a bit before checking again
		if !inuse {
			break
		}
	}
}

// Run the demo at super speed to make sure it all goes ok
func TestDoomDemo(t *testing.T) {
	dg_speed_ratio = 100.0 // Run at 50x speed
	game := &doomTestHeadless{
		t:     t,
		start: time.Now(),
		keys: []delayedKeyEvent{
			{DoomKeyEvent{Pressed: true, Key: KEY_ESCAPE}, 60_000},
			{DoomKeyEvent{Pressed: false, Key: KEY_ESCAPE}, 100},
			{DoomKeyEvent{Pressed: true, Key: KEY_UPARROW1}, 100},
			{DoomKeyEvent{Pressed: false, Key: KEY_UPARROW1}, 100},
			{DoomKeyEvent{Pressed: true, Key: KEY_ENTER}, 100},
			{DoomKeyEvent{Pressed: false, Key: KEY_ENTER}, 100},
			{DoomKeyEvent{Pressed: true, Key: 'y'}, 100},
			{DoomKeyEvent{Pressed: false, Key: 'y'}, 100},
		},
	}
	defer game.Close()
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

func TestLoadSave(t *testing.T) {
	dg_speed_ratio = 100.0 // Run at 50x speed
	game := &doomTestHeadless{
		t:     t,
		start: time.Now(),
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
		if percent > 10 { // Allow a small margin of error
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
	dg_speed_ratio = 100.0 // Run at 50x speed
	game := &doomTestHeadless{
		t:     t,
		start: time.Now(),
		keys:  nil,
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
		// Do 10 seconds of random movement
		for i := 0; i < 1000; i++ {
			change := rand.Intn(len(keys))
			key := keys[change]
			game.InsertKeyChange(key, true)
			time.Sleep(2 * time.Millisecond)
			game.InsertKeyChange(key, false)
			time.Sleep(1 * time.Millisecond)
		}

		// Exit
		game.InsertKey(KEY_ESCAPE)   // Open menu
		game.InsertKey(KEY_UPARROW1) // Go to quit
		game.InsertKey(KEY_ENTER)    // Confirm quit
		game.InsertKey('y')          // Confirm exit
	}()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func TestDoomLevels(t *testing.T) {

	dg_speed_ratio = 100.0 // Run at 50x speed
	game := &doomTestHeadless{
		t:     t,
		start: time.Now(),
		keys:  nil,
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

		// Go through levels E1M1 to E1M9 using the IDCLEV cheat
		for i := 1; i <= 9; i++ {
			sequence := []uint8{'i', 'd', 'c', 'l', 'e', 'v', '1', '0' + uint8(i)}
			game.InsertKeySequence(sequence...)
			time.Sleep(20 * time.Millisecond) // Wait for the level to load
			t.Logf("Completed level E1M%d", i)
		}

		// Exit
		game.InsertKey(KEY_ESCAPE)   // Open menu
		game.InsertKey(KEY_UPARROW1) // Go to quit
		game.InsertKey(KEY_ENTER)    // Confirm quit
		game.InsertKey('y')          // Confirm exit
	}()
	Run(game, []string{"-iwad", "doom1.wad"})
}

func TestDoomMap(t *testing.T) {
	dg_speed_ratio = 100.0 // Run at 50x speed
	game := &doomTestHeadless{
		t:     t,
		start: time.Now(),
		keys:  nil,
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
	dg_speed_ratio = 100.0 // Run at 50x speed
	game := &doomTestHeadless{
		t:     t,
		start: time.Now(),
		keys:  nil,
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

package gore

import (
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
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
}

func (d *doomTestHeadless) Close() {
	if err := d.outputFile.Close(); err != nil {
		d.t.Errorf("Error closing output file: %v", err)
	}
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
	return stdin, nil
}

func (d *doomTestHeadless) DrawFrame(frame *image.RGBA) {
	if d.outputFile == nil {
		var err error
		name := fmt.Sprintf("doom_test_%s.mp4", d.t.Name())
		d.outputFile, err = ffmpegSaver(name)
		if err != nil {
			d.t.Fatalf("Error starting ffmpeg: %v", err)
		}
	}
	d.outputFile.Write(frame.Pix)
}

func (d *doomTestHeadless) SetTitle(title string) {
	d.t.Logf("SetTitle called with: %s", title)
}

func (d *doomTestHeadless) GetKey(event *DoomKeyEvent) bool {
	if len(d.keys) == 0 {
		return false
	}
	now := I_GetTimeMS()
	delta := now - d.lastEventTick
	if d.keys[0].ticks < delta {
		*event = d.keys[0].event
		d.t.Logf("Key event: %#v, delta=%d (%d remaining)", *event, delta, len(d.keys)-1)
		d.keys = d.keys[1:]
		d.lastEventTick = now
		return true
	}
	return false
}

// Run the demo at super speed to make sure it all goes ok
func TestDoomDemo(t *testing.T) {
	dg_speed_ratio = 50.0 // Run at 50x speed
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
	Run(game, 1, 0)
}

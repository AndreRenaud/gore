package main

import (
	"image"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/AndreRenaud/gore"
)

type keyChange struct {
	Key   int
	State bool
}

var (
	streamer   MJPEGHandler
	keyChanges []keyChange
	keyLock    sync.Mutex
)

type webDoomFrontend struct {
}

func handleKey(key string, state string) error {
	keyVal, err := strconv.Atoi(key)
	if err != nil {
		return err
	}
	stateVal, err := strconv.Atoi(state)
	if err != nil {
		return err
	}
	stateBool := stateVal != 0

	keyLock.Lock()
	defer keyLock.Unlock()
	keyChanges = append(keyChanges, keyChange{
		Key:   keyVal,
		State: stateBool,
	})
	return nil
}

func (w *webDoomFrontend) DrawFrame(frame *image.RGBA) {
	if _, err := streamer.AddImage(frame); err != nil {
		log.Printf("Error adding image to MJPEG stream: %v\n", err)
	}
}

func (w *webDoomFrontend) GetEvent(event *gore.DoomEvent) bool {
	// This is a stub; actual key handling would depend on the platform and input system.
	keyLock.Lock()
	defer keyLock.Unlock()
	//log.Printf("DG_GetKey called with pressed: %d, doomKey: %d outstanding entries %d\n", pressed, doomKey, len(keyChanges))
	if len(keyChanges) > 0 {
		change := keyChanges[0]
		keyChanges = keyChanges[1:]
		log.Printf("Processing key change: key=%d, state=%t\n", change.Key, change.State)

		var thisDoomKey int32
		switch change.Key {
		case 38:
			thisDoomKey = gore.KEY_UPARROW1
		case 40:
			thisDoomKey = gore.KEY_DOWNARROW1
		case 37:
			thisDoomKey = gore.KEY_LEFTARROW1
		case 39:
			thisDoomKey = gore.KEY_RIGHTARROW1
		case 17: // Ctrl
			thisDoomKey = gore.KEY_FIRE1
		case 32:
			thisDoomKey = gore.KEY_USE1
		case 13:
			thisDoomKey = gore.KEY_ENTER
		case 27:
			thisDoomKey = gore.KEY_ESCAPE
		default:
			log.Printf("Unknown key %d, ignoring", change.Key)
			return false
		}

		t := gore.Ev_keyup
		if change.State {
			t = gore.Ev_keydown
		}
		event.Type = t
		event.Key = uint8(thisDoomKey)
		return true
	}
	return false
}

func (w *webDoomFrontend) SetTitle(title string) {
	log.Printf("DG_SetWindowTitle called with title: %s\n", title)
	// This is a stub; actual window title setting would depend on the platform and windowing system.
}

func main() {
	frontend := &webDoomFrontend{}

	log.Printf("DG_Init called\n")
	addr := ":8080"

	mux := http.NewServeMux()

	mux.Handle("GET /stream.mjpg", &streamer)
	mux.HandleFunc("POST /key/{key}/{state}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		state := r.PathValue("state")
		if err := handleKey(key, state); err != nil {
			http.Error(w, "Invalid key or state value", http.StatusBadRequest)
			log.Printf("Error handling key event: %v\n", err)
			return
		}
	})
	mux.Handle("GET /", http.FileServer(http.Dir("./static")))

	go func() {
		log.Printf("Starting HTTP server on %s\n", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("Failed to start HTTP server: %v\n", err)
		}
	}()

	// Called explicitly here to ensure `deadcode` check doesn't impact us
	gore.SetVirtualFileSystem(os.DirFS("."))

	defer gore.Stop()

	gore.Run(frontend, os.Args[1:])
}

package main

import (
	"flag"
	"image"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/AndreRenaud/gore"
)

type keyEvent struct {
	doomKey uint8
	down    bool
}

type Client struct {
	id       string
	pressed  map[uint8]bool // doomKey -> pressed
	lastSeen time.Time
}

var (
	streamer     MJPEGHandler
	clientsMu    sync.Mutex
	clients      = make(map[string]*Client)
	aggPressed   = make(map[uint8]bool) // aggregated pressed state across all clients
	eventQueue   []keyEvent             // events to feed to DOOM
	eventQueueMu sync.Mutex
)

// mapBrowserKeyToDoom maps JS keyCode values to DOOM key codes
func mapBrowserKeyToDoom(key int) uint8 {
	switch key {
	case 38:
		return gore.KEY_UPARROW1
	case 40:
		return gore.KEY_DOWNARROW1
	case 37:
		return gore.KEY_LEFTARROW1
	case 39:
		return gore.KEY_RIGHTARROW1
	case 17: // Ctrl
		return gore.KEY_FIRE1
	case 32: // Space
		return gore.KEY_USE1
	case 13: // Enter
		return gore.KEY_ENTER
	case 27: // Escape
		return gore.KEY_ESCAPE
	case 89: // Y
		return 'y'
	case 78: // N
		return 'n'
	case '0':
		return '0'
	case '1':
		return '1'
	case '2':
		return '2'
	case '3':
		return '3'
	case '4':
		return '4'
	case '5':
		return '5'
	case '6':
		return '6'
	case '7':
		return '7'
	case '8':
		return '8'
	case '9':
		return '9'
	default:
		log.Printf("Unsupported key code: %d", key)
		return 0
	}
}

// updateClientKey updates a client's key state and enqueues DOOM events if the aggregated state changed.
func updateClientKey(cid string, doomKey uint8, down bool) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	c := clients[cid]
	if c == nil {
		c = &Client{id: cid, pressed: make(map[uint8]bool), lastSeen: time.Now()}
		clients[cid] = c
	}
	c.lastSeen = time.Now()

	// Update client's pressed state for this key
	if down {
		c.pressed[doomKey] = true
	} else {
		delete(c.pressed, doomKey)
	}

	// Recompute aggregated state for this key
	was := aggPressed[doomKey]
	now := false
	for _, other := range clients {
		if other.pressed[doomKey] {
			now = true
			break
		}
	}
	if was != now {
		aggPressed[doomKey] = now
		queueEvent(doomKey, now)
	}
}

func queueEvent(doomKey uint8, down bool) {
	eventQueueMu.Lock()
	defer eventQueueMu.Unlock()
	eventQueue = append(eventQueue, keyEvent{doomKey: doomKey, down: down})
}

// removeClient removes a client and enqueues key-up events where aggregation changes to up
func removeClient(cid string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	c := clients[cid]
	if c == nil {
		return
	}
	delete(clients, cid)
	// For all keys that this client had pressed, recompute aggregation
	for dk := range c.pressed {
		was := aggPressed[dk]
		now := false
		for _, other := range clients {
			if other.pressed[dk] {
				now = true
				break
			}
		}
		if was != now {
			aggPressed[dk] = now
			queueEvent(dk, now)
		}
	}
}

// cleanupInactiveClients periodically removes clients that have not pinged recently.
func cleanupInactiveClients(maxAge time.Duration) {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		deadline := time.Now().Add(-maxAge)
		var toRemove []string
		clientsMu.Lock()
		for id, c := range clients {
			if c.lastSeen.Before(deadline) {
				toRemove = append(toRemove, id)
			}
		}
		clientsMu.Unlock()
		for _, id := range toRemove {
			log.Printf("Removing inactive client %s", id)
			removeClient(id)
		}
	}
}

type webDoomFrontend struct{}

func (w *webDoomFrontend) DrawFrame(frame *image.RGBA) {
	if _, err := streamer.AddImage(frame); err != nil {
		log.Printf("Error adding image to MJPEG stream: %v", err)
	}
}

func (w *webDoomFrontend) GetEvent(event *gore.DoomEvent) bool {
	eventQueueMu.Lock()
	defer eventQueueMu.Unlock()
	if len(eventQueue) == 0 {
		return false
	}
	ev := eventQueue[0]
	eventQueue = eventQueue[1:]
	if ev.down {
		event.Type = gore.Ev_keydown
	} else {
		event.Type = gore.Ev_keyup
	}
	event.Key = ev.doomKey
	return true
}

func (w *webDoomFrontend) SetTitle(title string)                        {}
func (w *webDoomFrontend) CacheSound(name string, data []byte)          {}
func (w *webDoomFrontend) PlaySound(name string, channel, vol, sep int) {}

func getClientID(r *http.Request) string {
	cid := r.URL.Query().Get("cid")
	if cid != "" {
		return cid
	}
	// fallback: remote address
	return r.RemoteAddr
}

func handleKey(w http.ResponseWriter, r *http.Request) {
	keyStr := r.PathValue("key")
	stateStr := r.PathValue("state")
	keyVal, err := strconv.Atoi(keyStr)
	if err != nil {
		http.Error(w, "bad key", http.StatusBadRequest)
		return
	}
	stateVal, err := strconv.Atoi(stateStr)
	if err != nil {
		http.Error(w, "bad state", http.StatusBadRequest)
		return
	}
	doomKey := mapBrowserKeyToDoom(keyVal)
	if doomKey == 0 {
		// ignore unsupported keys silently
		w.WriteHeader(http.StatusNoContent)
		return
	}
	cid := getClientID(r)
	updateClientKey(cid, doomKey, stateVal != 0)
	w.WriteHeader(http.StatusNoContent)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	cid := getClientID(r)
	clientsMu.Lock()
	c := clients[cid]
	if c == nil {
		c = &Client{id: cid, pressed: make(map[uint8]bool)}
		clients[cid] = c
	}
	c.lastSeen = time.Now()
	clientsMu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func handleDisconnect(w http.ResponseWriter, r *http.Request) {
	cid := getClientID(r)
	removeClient(cid)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	addr := flag.String("addr", ":8080", "Address to listen on for HTTP server")
	flag.Parse()
	frontend := &webDoomFrontend{}

	mux := http.NewServeMux()

	mux.Handle("GET /stream.mjpg", &streamer)
	mux.HandleFunc("POST /key/{key}/{state}", handleKey)
	mux.HandleFunc("GET /ping", handlePing)
	mux.HandleFunc("POST /disconnect", handleDisconnect)

	// Serve index.html from this example directory
	cwd, _ := os.Getwd()
	root := filepath.Join(cwd, "example", "1doom4all")
	mux.Handle("GET /", http.FileServer(http.Dir(root)))

	go func() {
		log.Printf("Starting HTTP server on %s", *addr)
		if err := http.ListenAndServe(*addr, mux); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Make wad files visible from repo root
	gore.SetVirtualFileSystem(os.DirFS("."))
	defer gore.Stop()

	// Cleanup inactive clients after 20 seconds of inactivity
	go cleanupInactiveClients(20 * time.Second)

	gore.EnableQuitting(false)

	gore.Run(frontend, os.Args[1:])
}

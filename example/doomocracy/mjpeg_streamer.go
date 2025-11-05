package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"sync"
)

type MJPEGHandler struct {
	mutex     sync.Mutex
	listeners []chan []byte
	// Store the last 3 frames to avoid allocations
	buffer     [3]bytes.Buffer
	nextBuffer int
}

// AddImage encodes the frame once and fan-outs to all listeners without re-encoding.
func (h *MJPEGHandler) AddImage(img image.Image) (int, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	// Shortcut - don't bother doing anything if nobody is listening
	if len(h.listeners) == 0 {
		return 0, nil
	}
	buf := &h.buffer[h.nextBuffer]
	h.nextBuffer = (h.nextBuffer + 1) % len(h.buffer)

	buf.Reset()
	options := &jpeg.Options{Quality: 90}
	if err := jpeg.Encode(buf, img, options); err != nil {
		return 0, err
	}

	deleted := 0
	for i, c := range h.listeners {
		// If the listener has not consumed the previous frame, drop it and close
		select {
		case c <- buf.Bytes():
			// frame sent successfully
		default:
			// listener is not keeping up, drop it
			log.Printf("Listener is not ready to receive a new frame")
			deleted++
			h.listeners[i] = nil
			close(c)
		}
	}
	// Only rebuild the listeners slice if needed, to avoid lots of allocations
	if deleted > 0 {
		var newListeners []chan []byte
		for _, c := range h.listeners {
			if c != nil {
				newListeners = append(newListeners, c)
			}
		}
		h.listeners = newListeners
	}
	return len(h.listeners), nil
}

func (h *MJPEGHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	boundary := "\r\n--frame\r\nContent-Type: image/jpeg\r\n\r\n"

	h.mutex.Lock()
	c := make(chan []byte, 2)
	h.listeners = append(h.listeners, c)
	h.mutex.Unlock()

	for {
		imgBuf, ok := <-c
		if !ok {
			break
		}
		if n, err := io.WriteString(w, boundary); err != nil || n != len(boundary) {
			return
		}
		if n, err := w.Write(imgBuf); err != nil || n != len(imgBuf) {
			return
		}
		if n, err := io.WriteString(w, "\r\n"); err != nil || n != 2 {
			return
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func (h *MJPEGHandler) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	for _, c := range h.listeners {
		close(c)
	}
	h.listeners = nil
}

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
}

// AddImage encodes the frame once and fan-outs to all listeners without re-encoding.
func (h *MJPEGHandler) AddImage(img image.Image) (int, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	// Don't bother doing anything if nobody is listening
	if len(h.listeners) == 0 {
		return 0, nil
	}

	var buf bytes.Buffer
	options := &jpeg.Options{Quality: 90}
	if err := jpeg.Encode(&buf, img, options); err != nil {
		return 0, err
	}

	newListeners := make([]chan []byte, 0, len(h.listeners))
	for _, c := range h.listeners {
		// If the listener has not consumed the previous frame, drop it and close
		if len(c) > 0 {
			log.Printf("Listener is not ready to receive a new frame")
			close(c)
			continue
		}
		c <- buf.Bytes()
		newListeners = append(newListeners, c)
	}
	h.listeners = newListeners
	return len(newListeners), nil
}

func (h *MJPEGHandler) HasClients() bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return len(h.listeners) > 0
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

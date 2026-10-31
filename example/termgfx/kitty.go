package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"image"
)

// writeKitty appends img to out as a kitty graphics protocol image, which the
// terminal stretches to cover cols x rows cells.
// See https://sw.kovidgoyal.net/kitty/graphics-protocol/
func writeKitty(out *bytes.Buffer, img *image.RGBA, cols, rows int) {
	var z bytes.Buffer
	zw, _ := zlib.NewWriterLevel(&z, zlib.BestSpeed)
	_, _ = zw.Write(img.Pix) // Doom's frame buffer has no padding, so this is exactly the image
	_ = zw.Close()
	data := base64.StdEncoding.EncodeToString(z.Bytes())

	// Reusing the image and placement ids replaces the previous frame, q=2
	// suppresses replies that would otherwise arrive as key presses, and C=1
	// leaves the cursor where it is so the image never scrolls the screen.
	b := img.Bounds()
	fmt.Fprintf(out, "\x1b_Ga=T,f=32,o=z,s=%d,v=%d,i=1,p=1,q=2,C=1,c=%d,r=%d", b.Dx(), b.Dy(), cols, rows)
	// The payload is sent in chunks of at most 4096 bytes
	for len(data) > 0 {
		chunk := data[:min(len(data), 4096)]
		data = data[len(chunk):]
		more := 0
		if len(data) > 0 {
			more = 1
		}
		fmt.Fprintf(out, ",m=%d;%s\x1b\\", more, chunk)
		if more == 1 {
			out.WriteString("\x1b_Gq=2")
		}
	}
}

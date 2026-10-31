package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"image"
)

// kittyEncoder converts frames to kitty graphics protocol escape sequences,
// reusing its buffers between frames.
// See https://sw.kovidgoyal.net/kitty/graphics-protocol/
type kittyEncoder struct {
	raw []byte
	z   bytes.Buffer
	zw  *zlib.Writer
	b64 []byte
}

// Encode appends img to out as a kitty image, scaled up by an integer factor.
// When cols and rows are non-zero the terminal stretches the image to cover
// that many cells, otherwise it is shown at its pixel size.
func (e *kittyEncoder) Encode(out *bytes.Buffer, img *image.RGBA, scale, cols, rows int) {
	b := img.Bounds()
	w, h := b.Dx()*scale, b.Dy()*scale

	// Packed 24-bit RGB, nearest neighbour scaled
	stride := w * 3
	if cap(e.raw) < stride*h {
		e.raw = make([]byte, stride*h)
	}
	e.raw = e.raw[:stride*h]
	for y := range h {
		dst := e.raw[y*stride : (y+1)*stride]
		if y%scale != 0 {
			copy(dst, e.raw[(y-1)*stride:y*stride])
			continue
		}
		src := img.Pix[img.PixOffset(b.Min.X, b.Min.Y+y/scale):]
		for x := range w {
			s := x / scale * 4
			copy(dst[x*3:x*3+3], src[s:s+3])
		}
	}

	e.z.Reset()
	if e.zw == nil {
		e.zw, _ = zlib.NewWriterLevel(&e.z, zlib.BestSpeed)
	} else {
		e.zw.Reset(&e.z)
	}
	_, _ = e.zw.Write(e.raw)
	_ = e.zw.Close()

	n := base64.StdEncoding.EncodedLen(e.z.Len())
	if cap(e.b64) < n {
		e.b64 = make([]byte, n)
	}
	e.b64 = e.b64[:n]
	base64.StdEncoding.Encode(e.b64, e.z.Bytes())

	// The payload is sent in chunks of at most 4096 bytes. Reusing the image
	// and placement ids replaces the previous frame, q=2 suppresses replies
	// that would otherwise arrive as key presses, and C=1 leaves the cursor
	// where it is so the image never scrolls the screen.
	const chunk = 4096
	for off := 0; off < len(e.b64); off += chunk {
		end := min(off+chunk, len(e.b64))
		more := 0
		if end < len(e.b64) {
			more = 1
		}
		if off == 0 {
			fmt.Fprintf(out, "\x1b_Ga=T,f=24,o=z,s=%d,v=%d,i=1,p=1,q=2,C=1", w, h)
			if cols > 0 && rows > 0 {
				fmt.Fprintf(out, ",c=%d,r=%d", cols, rows)
			}
			fmt.Fprintf(out, ",m=%d;", more)
		} else {
			fmt.Fprintf(out, "\x1b_Gq=2,m=%d;", more)
		}
		out.Write(e.b64[off:end])
		out.WriteString("\x1b\\")
	}
}

// kittyDeleteAll removes every kitty image, freeing their data.
const kittyDeleteAll = "\x1b_Ga=d,d=A,q=2\x1b\\"

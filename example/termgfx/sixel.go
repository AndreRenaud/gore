package main

import (
	"bytes"
	"fmt"
	"image"
)

// sixelEncoder converts frames to sixel images, reusing its buffers between
// frames.
type sixelEncoder struct {
	colours map[uint32]uint8
	indexed []uint8 // colour register of each source pixel
	bits    []byte  // sixels of the current band, a row for each colour
	used    []uint8 // colours in the current band
	inBand  [256]bool
}

// Encode appends img to out as a w x h sixel image, using nearest neighbour
// scaling.
func (e *sixelEncoder) Encode(out *bytes.Buffer, img *image.RGBA, w, h int) {
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	// DCS with transparent background (P2=1) and raster attributes giving
	// square pixels and the image size
	fmt.Fprintf(out, "\x1bP0;1;0q\"1;1;%d;%d", w, h)

	// Doom renders through a 256 entry palette, so the frame's colours fit in
	// the sixel colour registers without any loss.
	if e.colours == nil {
		e.colours = make(map[uint32]uint8)
	}
	clear(e.colours)
	e.indexed = e.indexed[:0]
	for y := range sh {
		row := img.Pix[img.PixOffset(b.Min.X, b.Min.Y+y):]
		for x := range sw {
			p := row[x*4:]
			key := uint32(p[0])<<16 | uint32(p[1])<<8 | uint32(p[2])
			c, ok := e.colours[key]
			if !ok {
				c = uint8(len(e.colours))
				e.colours[key] = c
				// sixel colour components are percentages
				fmt.Fprintf(out, "#%d;2;%d;%d;%d", c, pct(p[0]), pct(p[1]), pct(p[2]))
			}
			e.indexed = append(e.indexed, c)
		}
	}

	if len(e.bits) != 256*w {
		e.bits = make([]byte, 256*w)
	}
	// Each sixel character covers a column of 6 pixels
	for by := 0; by < h; by += 6 {
		for dy := 0; dy < 6 && by+dy < h; dy++ {
			src := e.indexed[(by+dy)*sh/h*sw:]
			for x := range w {
				c := src[x*sw/w]
				if !e.inBand[c] {
					e.inBand[c] = true
					e.used = append(e.used, c)
				}
				e.bits[int(c)*w+x] |= 1 << dy
			}
		}

		for i, c := range e.used {
			if i > 0 {
				// graphics carriage return, to overlay the next colour on this band
				out.WriteByte('$')
			}
			fmt.Fprintf(out, "#%d", c)
			row := e.bits[int(c)*w : int(c)*w+w]
			for x := 0; x < w; {
				n := 1
				for x+n < w && row[x+n] == row[x] {
					n++
				}
				if x+n < w || row[x] != 0 { // a trailing blank run can be left out
					writeRun(out, row[x], n)
				}
				x += n
			}
			clear(row)
			e.inBand[c] = false
		}
		e.used = e.used[:0]
		if by+6 < h {
			// graphics new line, move down to the next band
			out.WriteByte('-')
		}
	}
	out.WriteString("\x1b\\")
}

// writeRun emits n copies of the sixel with the given bits, using the
// repeat introducer once that is shorter.
func writeRun(out *bytes.Buffer, bits byte, n int) {
	if n > 3 {
		fmt.Fprintf(out, "!%d%c", n, '?'+bits)
		return
	}
	for range n {
		out.WriteByte('?' + bits)
	}
}

func pct(v uint8) int {
	return (int(v)*100 + 127) / 255
}

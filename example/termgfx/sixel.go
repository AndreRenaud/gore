package main

import (
	"bytes"
	"fmt"
	"image"
	"strconv"
)

// sixelEncoder converts frames to sixel escape sequences, reusing its
// buffers between frames.
type sixelEncoder struct {
	colours map[uint32]uint8
	palette [][3]uint8
	indexed []uint8 // palette index for each source pixel
	xmap    []int   // source column for each output column

	// Per-colour sixel bits for the band being encoded, along with the
	// range of columns that colour covers
	bands      [256][]byte
	minX, maxX [256]int
	used       []uint8
}

// Encode appends img to out as a w x h sixel image, using nearest neighbour
// scaling so integer scales stay pixel perfect.
func (e *sixelEncoder) Encode(out *bytes.Buffer, img *image.RGBA, w, h int) {
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()

	// Doom renders through a 256 entry palette, so a frame never has more
	// than 256 distinct colours and indexing them is lossless.
	if e.colours == nil {
		e.colours = make(map[uint32]uint8, 256)
	}
	clear(e.colours)
	e.palette = e.palette[:0]
	if cap(e.indexed) < sw*sh {
		e.indexed = make([]uint8, sw*sh)
	}
	e.indexed = e.indexed[:sw*sh]
	lastKey, lastIdx := uint32(1<<24), uint8(0)
	for y := range sh {
		row := img.Pix[img.PixOffset(b.Min.X, b.Min.Y+y):]
		for x := range sw {
			key := uint32(row[x*4])<<16 | uint32(row[x*4+1])<<8 | uint32(row[x*4+2])
			if key != lastKey {
				idx, ok := e.colours[key]
				if !ok {
					// Only reachable for non-Doom content; fold extras into the last register
					idx = 255
					if len(e.palette) < 256 {
						idx = uint8(len(e.palette))
						e.colours[key] = idx
						e.palette = append(e.palette, [3]uint8{row[x*4], row[x*4+1], row[x*4+2]})
					}
				}
				lastKey, lastIdx = key, idx
			}
			e.indexed[y*sw+x] = lastIdx
		}
	}

	if cap(e.xmap) < w {
		e.xmap = make([]int, w)
	}
	e.xmap = e.xmap[:w]
	for x := range w {
		e.xmap[x] = x * sw / w
	}

	// DCS with transparent background (P2=1) and raster attributes giving
	// square pixels and the image size
	fmt.Fprintf(out, "\x1bP0;1;0q\"1;1;%d;%d", w, h)
	for i, c := range e.palette {
		// sixel colour components are percentages
		fmt.Fprintf(out, "#%d;2;%d;%d;%d", i, pct(c[0]), pct(c[1]), pct(c[2]))
	}

	for i := range e.bands {
		if e.bands[i] != nil && len(e.bands[i]) != w {
			e.bands[i] = nil
		}
		e.minX[i], e.maxX[i] = w, -1
	}

	// Each sixel character covers a column of 6 pixels
	for by := 0; by < h; by += 6 {
		e.used = e.used[:0]
		for dy := 0; dy < 6 && by+dy < h; dy++ {
			sy := (by + dy) * sh / h
			src := e.indexed[sy*sw : sy*sw+sw]
			bit := byte(1) << dy
			for x, sx := range e.xmap {
				c := src[sx]
				row := e.bands[c]
				if row == nil {
					row = make([]byte, w)
					e.bands[c] = row
				}
				if e.maxX[c] < 0 {
					e.used = append(e.used, c)
				}
				e.minX[c] = min(e.minX[c], x)
				e.maxX[c] = max(e.maxX[c], x)
				row[x] |= bit
			}
		}

		for i, c := range e.used {
			if i > 0 {
				// graphics carriage return, to overlay the next colour on this band
				out.WriteByte('$')
			}
			out.WriteByte('#')
			out.WriteString(strconv.Itoa(int(c)))
			writeRun(out, 0, e.minX[c])
			row := e.bands[c][e.minX[c] : e.maxX[c]+1]
			for x := 0; x < len(row); {
				n := 1
				for x+n < len(row) && row[x+n] == row[x] {
					n++
				}
				writeRun(out, row[x], n)
				x += n
			}
			clear(row)
			e.minX[c], e.maxX[c] = w, -1
		}
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
	ch := '?' + bits
	if n > 3 {
		out.WriteByte('!')
		out.WriteString(strconv.Itoa(n))
		out.WriteByte(ch)
		return
	}
	for range n {
		out.WriteByte(ch)
	}
}

func pct(v uint8) int {
	return (int(v)*100 + 127) / 255
}

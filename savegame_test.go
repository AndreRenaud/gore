package gore

import (
	"bytes"
	"io"
	"testing"
)

// memSaveStream is an in-memory save_stream: reads and seeks come from the
// Reader, writes go to the Writer.
type memSaveStream struct {
	*bytes.Reader
	io.Writer
}

func (memSaveStream) Close() error { return nil }

// useSaveStream points save_stream at s for the rest of the test.
func useSaveStream(t *testing.T, s memSaveStream) {
	save_stream, savegame_error = s, 0
	t.Cleanup(func() { save_stream = nil })
}

// TestArchiveSpecialsWithRemovedThinker saves while a removed thinker is
// still in the thinker list. p_RemoveThinker only clears the function; the
// thinker is unlinked by the next p_RunThinkers, so this happens whenever a
// game is saved after something was removed during the previous tic.
func TestArchiveSpecialsWithRemovedThinker(t *testing.T) {
	savedCap, savedCeilings := thinkercap, activeceilings
	defer func() { thinkercap, activeceilings = savedCap, savedCeilings }()
	p_InitThinkers()
	activeceilings = [MAXCEILINGS]*ceiling_t{}

	var th thinker_t
	p_AddThinker(&th)
	p_RemoveThinker(&th)

	var buf bytes.Buffer
	useSaveStream(t, memSaveStream{bytes.NewReader(nil), &buf})
	p_ArchiveSpecials()
	if savegame_error != 0 || !bytes.Equal(buf.Bytes(), []byte{tc_endspecials}) {
		t.Fatalf("wrote %v (error %d), want just tc_endspecials", buf.Bytes(), savegame_error)
	}
}

// TestReadPointerSkip checks that saveg_readp_skip moves past exactly the 4
// bytes saveg_writep writes for a pointer.
func TestReadPointerSkip(t *testing.T) {
	r := bytes.NewReader([]byte{0xef, 0xbe, 0xad, 0xde, 0x2a, 0, 0, 0})
	useSaveStream(t, memSaveStream{r, io.Discard})
	saveg_readp_skip()
	if pos, _ := r.Seek(0, io.SeekCurrent); pos != 4 {
		t.Fatalf("offset %d after skipping a pointer, want 4", pos)
	}
	if v := saveg_read32(); v != 42 || savegame_error != 0 {
		t.Fatalf("next value %d (error %d), want 42", v, savegame_error)
	}
}

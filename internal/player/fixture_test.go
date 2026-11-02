package player

import (
	"testing"

	"netease-music-mygo/internal/lyrics"
)

// Setting fixture lyrics notifies listeners as a loaded track does.
func TestSetLyricsNotifies(t *testing.T) {
	p := New(func(func()) {})
	called := 0
	p.LyricsChanged = func() { called++ }
	p.SetLyrics([]lyrics.Line{{Text: "第一句"}})

	if called != 1 {
		t.Fatalf("lyrics changed %d times, want 1", called)
	}
	if lines, loaded := p.Lyrics(); !loaded || len(lines) != 1 {
		t.Fatalf("lyrics loaded %v, lines %v", loaded, lines)
	}
}

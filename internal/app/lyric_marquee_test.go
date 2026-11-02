package app

import (
	"strings"
	"testing"
	"time"
)

// A long lyric waits, moves while playing, and freezes when paused.
func TestLyricMarquee(t *testing.T) {
	var m lyricMarquee
	start := time.Unix(0, 0)
	if offset, after, moving := m.step(start, "很长的歌词", true, true, 100, 20); offset != 0 || after != lyricMarqueeDelay || moving {
		t.Fatalf("started at %v after %v moving %v", offset, after, moving)
	}
	if offset, after, moving := m.step(start.Add(lyricMarqueeDelay), "很长的歌词", true, true, 100, 20); offset != 0 || after != lyricMarqueeStep || !moving {
		t.Fatalf("waited at %v after %v moving %v", offset, after, moving)
	}
	offset, _, _ := m.step(start.Add(lyricMarqueeDelay+time.Second), "很长的歌词", true, true, 100, 20)
	paused, after, moving := m.step(start.Add(lyricMarqueeDelay+2*time.Second), "很长的歌词", true, false, 100, 20)
	if paused != offset || after != 0 || moving {
		t.Fatalf("paused at %v after %v moving %v, want %v", paused, after, moving, offset)
	}
	resumed, _, _ := m.step(start.Add(lyricMarqueeDelay+3*time.Second), "很长的歌词", true, true, 100, 20)
	if resumed != paused {
		t.Fatalf("resumed at %v, want %v", resumed, paused)
	}
	if offset, after, moving := m.step(start.Add(lyricMarqueeDelay+4*time.Second), "新的一句", true, true, 100, 20); offset != 0 || after != lyricMarqueeDelay || moving {
		t.Fatalf("changed at %v after %v moving %v", offset, after, moving)
	}
}

// A lyric that fits its viewport never starts a marquee.
func TestShortLyricMarqueeIsIdle(t *testing.T) {
	var m lyricMarquee
	offset, after, moving := m.step(time.Unix(0, 0), "短句", false, true, 100, 20)
	if offset != 0 || after != 0 || moving {
		t.Fatalf("short lyric at %v after %v moving %v", offset, after, moving)
	}
}

// The menu bar marquee always returns a rune-safe fixed-width window.
func TestMenuBarLyricMarquee(t *testing.T) {
	var m lyricMarquee
	start := time.Unix(0, 0)
	text := strings.Repeat("🎵", menuBarLyricRunes+1)
	first, after := m.menuBarLyricTitle(start, text, true)
	if len([]rune(first)) != menuBarLyricRunes || after != lyricMarqueeDelay {
		t.Fatalf("first title %q after %v", first, after)
	}
	m.menuBarLyricTitle(start.Add(lyricMarqueeDelay+time.Second), text, true)
	next, _ := m.menuBarLyricTitle(start.Add(lyricMarqueeDelay+2*time.Second), text, true)
	if next == first {
		t.Fatal("menu bar lyric did not scroll")
	}
}

// A short menu bar lyric is right-aligned against the app's tray icon.
func TestMenuBarLyricRightAlignment(t *testing.T) {
	var m lyricMarquee
	text := "短歌词"
	got, _ := m.menuBarLyricTitle(time.Unix(0, 0), text, true)
	want := strings.Repeat("\u2007", menuBarLyricRunes-len([]rune(text))) + text
	if got != want {
		t.Fatalf("title %q, want %q", got, want)
	}
}

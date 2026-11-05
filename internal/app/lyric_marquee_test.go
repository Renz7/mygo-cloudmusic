package app

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
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

// A long lyric stops once it has scrolled to its end, and stays there.
func TestLyricMarqueeStopsAtEnd(t *testing.T) {
	var m lyricMarquee
	start := time.Unix(0, 0)
	m.step(start, "很长的歌词", true, true, 100, 20)
	offset, after, moving := m.step(start.Add(lyricMarqueeDelay+6*time.Second), "很长的歌词", true, true, 100, 20)
	if offset != 100 || after != 0 || moving {
		t.Fatalf("ended at %v after %v moving %v", offset, after, moving)
	}
	offset, after, moving = m.step(start.Add(lyricMarqueeDelay+7*time.Second), "很长的歌词", true, true, 100, 20)
	if offset != 100 || after != 0 || moving {
		t.Fatalf("stayed at %v after %v moving %v", offset, after, moving)
	}
}

// The menu bar marquee scrolls a long lyric once, then stops at its end.
func TestMenuBarLyricMarquee(t *testing.T) {
	var m lyricMarquee
	start := time.Unix(0, 0)
	runes := []rune("开头" + strings.Repeat("词", menuBarLyricRunes) + "结尾")
	text := string(runes)
	first, after := m.menuBarLyricTitle(start, text, true)
	if first != string(runes[:menuBarLyricRunes]) || after != lyricMarqueeDelay {
		t.Fatalf("first title %q after %v", first, after)
	}
	end, after := m.menuBarLyricTitle(start.Add(lyricMarqueeDelay+time.Minute), text, true)
	want := string(runes[len(runes)-menuBarLyricRunes:])
	if end != want || after != 0 {
		t.Fatalf("end title %q after %v, want %q", end, after, want)
	}
	stayed, after := m.menuBarLyricTitle(start.Add(lyricMarqueeDelay+2*time.Minute), text, true)
	if stayed != want || after != 0 {
		t.Fatalf("stayed at %q after %v", stayed, after)
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

// An overflowing notch lyric scrolls only within a single visible line.
func TestNotchLyricMarqueeDoesNotWrap(t *testing.T) {
	n := &notch{a: newTestApp()}
	var marquee lyricMarquee
	tt := ui.NewTester(func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		n.lyricMarquee(c, &marquee, strings.Repeat("长歌词", 20)+"\n第二行", 100, 12, white)
	}, 100, 48)
	tt.Frame()
	for y := 20; y < tt.Image().Bounds().Dy(); y++ {
		for x := 0; x < tt.Image().Bounds().Dx(); x++ {
			if tt.Image().RGBAAt(x, y).A != 0 {
				t.Fatalf("lyric rendered below its single-line viewport at %d,%d", x, y)
			}
		}
	}
}

// An overflowing notch lyric starts left-aligned before it scrolls.
func TestNotchLyricMarqueeStartsLeftAligned(t *testing.T) {
	n := &notch{a: newTestApp()}
	var marquee lyricMarquee
	lyric := strings.Repeat("长歌词", 20)
	tt := ui.NewTester(func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		n.lyricMarquee(c, &marquee, lyric, 100, 12, white)
	}, 100, 48)
	tt.Frame()
	r, ok := tt.Find(lyric)
	if !ok {
		t.Fatal("did not show the overflowing lyric")
	}
	if r.X < -0.5 || r.X > 0.5 {
		t.Fatalf("overflowing lyric started at X=%v, want left-aligned near 0", r.X)
	}
}

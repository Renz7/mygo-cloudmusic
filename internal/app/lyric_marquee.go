package app

import (
	"strings"
	"time"
)

const (
	lyricMarqueeDelay = time.Second
	lyricMarqueeStep  = 150 * time.Millisecond
	lyricMarqueeSpeed = 30
)

// lyricMarquee keeps a scrolling lyric's position. Its units are points in the
// notch and runes in the menu bar. Overflowing lyrics scroll once to their
// end, then stop until the line changes.
type lyricMarquee struct {
	text     string
	overflow bool
	offset   float32
	delay    time.Duration
	last     time.Time
	paused   bool
}

func (m *lyricMarquee) reset() { *m = lyricMarquee{} }

// step moves an overflowing lyric while it plays, up to distance. It returns
// the next timer deadline during the initial pause, or the normal refresh
// interval while moving; once the end is reached it stays put.
func (m *lyricMarquee) step(now time.Time, text string, overflow, playing bool, distance, speed float32) (offset float32, after time.Duration, moving bool) {
	if text != m.text || overflow != m.overflow {
		m.text, m.overflow = text, overflow
		m.offset, m.delay, m.last = 0, lyricMarqueeDelay, now
		m.paused = false
	}
	if !overflow || distance <= 0 {
		return 0, 0, false
	}
	if m.last.IsZero() {
		m.last = now
	}
	if !playing {
		m.last, m.paused = now, true
		return m.offset, 0, false
	}
	if m.offset >= distance {
		m.offset = distance
		return m.offset, 0, false
	}
	if m.paused {
		m.last, m.paused = now, false
		if m.delay > 0 {
			return m.offset, m.delay, false
		}
		return m.offset, lyricMarqueeStep, true
	}
	elapsed := now.Sub(m.last)
	m.last = now
	if elapsed < 0 {
		elapsed = 0
	}
	if m.delay > 0 {
		if elapsed < m.delay {
			m.delay -= elapsed
			return m.offset, m.delay, false
		}
		elapsed -= m.delay
		m.delay = 0
	}
	m.offset += float32(elapsed.Seconds() * float64(speed))
	if m.offset >= distance {
		m.offset = distance
		return m.offset, 0, false
	}
	return m.offset, lyricMarqueeStep, true
}

// menuBarLyricTitle returns a fixed-rune title that scrolls a long lyric once
// to its end.
func (m *lyricMarquee) menuBarLyricTitle(now time.Time, text string, playing bool) (string, time.Duration) {
	runes := []rune(text)
	overflow := len(runes) > menuBarLyricRunes
	var max float32
	if overflow {
		max = float32(len(runes) - menuBarLyricRunes)
	}
	offset, after, _ := m.step(now, text, overflow, playing, max, 6)
	if !overflow {
		return padMenuBarLyric(runes), after
	}
	start := min(int(offset), len(runes)-menuBarLyricRunes)
	return string(runes[start : start+menuBarLyricRunes]), after
}

func padMenuBarLyric(runes []rune) string {
	return strings.Repeat("\u2007", menuBarLyricRunes-len(runes)) + string(runes)
}

func soonestAfter(a, b time.Duration) time.Duration {
	switch {
	case a <= 0:
		return b
	case b <= 0:
		return a
	default:
		return min(a, b)
	}
}

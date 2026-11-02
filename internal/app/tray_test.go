package app

import (
	"testing"
	"time"

	"netease-music-mygo/internal/lyrics"
	"netease-music-mygo/internal/netease"
)

// The menu bar lyric follows playback and explains when no line is available.
func TestMenuBarLyric(t *testing.T) {
	a := newTestApp()
	if text, after := a.menuBarLyric(); text != "未在播放" || after != 0 {
		t.Fatalf("idle lyric %q after %v", text, after)
	}

	a.pl.Cue([]netease.Track{{ID: 7, Name: "宝船", Duration: 200000}})
	if text, _ := a.menuBarLyric(); text != "正在加载歌词…" {
		t.Fatalf("unloaded lyric %q", text)
	}

	a.pl.SetLyrics(nil)
	if text, _ := a.menuBarLyric(); text != "暂无歌词" {
		t.Fatalf("empty lyric %q", text)
	}

	a.pl.SetLyrics([]lyrics.Line{
		{At: time.Second, End: 2 * time.Second, Text: "第一句"},
		{At: 2 * time.Second, End: 3 * time.Second, Text: "第二句"},
	})
	a.pl.Hold(time.Second)
	if text, after := a.menuBarLyric(); text != "第一句" || after != time.Second+20*time.Millisecond {
		t.Fatalf("current lyric %q after %v", text, after)
	}
}

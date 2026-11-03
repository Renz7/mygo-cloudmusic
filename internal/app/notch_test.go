package app

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/audio"
	"github.com/egoist/mygo/ui"

	"netease-music-mygo/internal/lyrics"
	"netease-music-mygo/internal/netease"
)

// The notch, turned off, shows itself as the first song plays, and only
// then; its button keeps it on.
func TestNotchGreetsOnce(t *testing.T) {
	a := newTestApp()
	a.settings = defaultSettings()
	n := &notch{a: a, w0: 200, h0: 32, winW: notchOpenW + 60, winH: 240}
	a.notch = n
	a.pl.Cue([]netease.Track{{ID: 7, Name: "宝船", Duration: 200000}})
	a.syncWindows()
	if n.hello {
		t.Fatal("greeted before the song played")
	}
	out := audio.NewPlayer(&silence{})
	out.SetVolume(0)
	a.pl.Attach(out)
	defer out.Close()
	a.syncWindows()
	if !n.hello || !a.settings.NotchGreeted {
		t.Fatal("did not greet as the song played")
	}

	tt := ui.NewTester(n.view, int(n.winW), int(n.winH))
	tt.Advance(time.Second)
	tt.Frame()
	if err := tt.Click("开启刘海播放器"); err != nil {
		t.Fatal(err)
	}
	if !a.settings.Notch || n.hello {
		t.Fatalf("kept: notch %v, greeting %v", a.settings.Notch, n.hello)
	}

	// Turned off again, it does not greet the next song.
	a.settings.Notch = false
	a.pl.Cue([]netease.Track{{ID: 8, Name: "下一首", Duration: 200000}})
	a.pl.Attach(out)
	a.syncWindows()
	if n.hello {
		t.Error("greeted twice")
	}
}

// The pointer resting on the greeting leaves it as it is, so its button can
// be clicked. It used to open the player, which hid the button, and leaving
// brought it back only to open again as the pointer came back to it.
func TestNotchGreetingKeptUnderPointer(t *testing.T) {
	a := newTestApp()
	a.settings = defaultSettings()
	n := &notch{a: a, real: true, w0: 185, h0: 32, winW: notchOpenW + 60, winH: 240}
	a.notch = n
	a.pl.Cue([]netease.Track{{ID: 7, Name: "宝船", Duration: 200000}})
	n.greet()
	if !n.hello {
		t.Fatal("did not greet")
	}

	defer func(f func() mygo.Point) { cursorPoint = f }(cursorPoint)
	cursorPoint = func() mygo.Point {
		return mygo.Point{X: int(n.screenX + n.winW/2 + 120), Y: int(n.screenY + n.h0 + 28)}
	}
	n.poll()
	n.enteredAt = n.lastInside.Add(-time.Second) // as if it had rested there
	n.poll()
	if n.open {
		t.Fatal("opened the player over the greeting")
	}

	tt := ui.NewTester(n.view, int(n.winW), int(n.winH))
	tt.Advance(time.Second)
	tt.Frame()
	if err := tt.Click("开启刘海播放器"); err != nil {
		t.Fatal(err)
	}
	if !a.settings.Notch || n.hello {
		t.Fatalf("kept: notch %v, greeting %v", a.settings.Notch, n.hello)
	}
}

// The notch shows its current lyric and hides it when its control is clicked.
func TestNotchShowsCurrentLyric(t *testing.T) {
	a := newTestApp()
	a.settings = defaultSettings()
	a.settings.Notch = true
	n := &notch{a: a, real: true, w0: 185, h0: 32, winW: notchOpenW + 60, winH: 240}
	a.notch = n
	a.pl.Cue([]netease.Track{{ID: 7, Name: "宝船", Duration: 200000}})
	a.pl.SetLyrics([]lyrics.Line{
		{At: time.Second, End: 2 * time.Second, Text: "第一句"},
		{At: 2 * time.Second, End: 3 * time.Second, Text: "第二句"},
	})
	a.pl.Hold(time.Second)

	if _, h := n.target(); h != n.h0+notchLyricH {
		t.Fatalf("compact height %v, want %v", h, n.h0+notchLyricH)
	}
	tt := ui.NewTester(n.view, int(n.winW), int(n.winH))
	tt.Frame()
	tt.Advance(time.Second)
	tt.Frame()
	if !tt.HasText("第一句") {
		t.Fatal("did not show the compact lyric")
	}

	n.open = true
	tt.Frame()
	tt.Advance(time.Second)
	tt.Frame()
	a.pl.Hold(2 * time.Second)
	tt.Frame()
	if !tt.HasText("第二句") || tt.HasText("第一句") {
		t.Fatal("did not update the expanded lyric")
	}

	defer func(apply func(*app)) { applyNotchSettings = apply }(applyNotchSettings)
	applyNotchSettings = func(*app) {}
	if err := tt.Click("切换刘海歌词"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.settings.NotchLyrics || tt.HasText("第二句") {
		t.Fatal("showed lyrics after menu bar lyrics were disabled")
	}

	a.settings.NotchLyrics = true
	a.pl.SetLyrics(nil)
	n.open = false
	tt.Frame()
	if _, h := n.target(); h != n.h0 {
		t.Fatalf("compact height without lyrics %v, want %v", h, n.h0)
	}
	if tt.HasText("第二句") {
		t.Fatal("showed a lyric after lyrics were removed")
	}
}

// A virtual notch keeps its lyric centered in the housing, not below it.
func TestNotchCentersLyricsWithoutPhysicalNotch(t *testing.T) {
	a := newTestApp()
	a.settings = defaultSettings()
	a.settings.Notch, a.settings.NotchVirtual = true, true
	n := &notch{a: a, h0: 32, w0: 200, winW: notchOpenW + 60, winH: 240}
	a.notch = n
	a.pl.Cue([]netease.Track{{ID: 7, Name: "宝船", Duration: 200000}})
	a.pl.SetLyrics([]lyrics.Line{{At: time.Second, End: 2 * time.Second, Text: "第一句"}})
	a.pl.Hold(time.Second)

	if _, h := n.target(); h != n.h0 {
		t.Fatalf("virtual notch height %v, want %v", h, n.h0)
	}
	tt := ui.NewTester(n.view, int(n.winW), int(n.winH))
	tt.Frame()
	tt.Advance(time.Second)
	tt.Frame()
	r, ok := tt.Find("第一句")
	if !ok {
		t.Fatal("did not show the virtual notch lyric")
	}
	if r.Y < 0 || r.Y+r.H > n.h0 {
		t.Fatalf("virtual notch lyric bounds %v, outside height %v", r, n.h0)
	}
}

// The notch queue changes tracks only when a queue entry is double-clicked.
func TestNotchQueueDoubleClickChangesTrack(t *testing.T) {
	a := newTestApp()
	a.settings = defaultSettings()
	a.settings.Notch = true
	n := &notch{a: a, real: true, w0: 185, h0: 32, winW: notchOpenW + 60, winH: notchQueueH + 64, open: true}
	a.notch = n
	a.pl.Cue([]netease.Track{
		{ID: 7, Name: "第一首", Duration: 200000},
		{ID: 8, Name: "下一首", Duration: 200000},
	})

	tt := ui.NewTester(n.view, int(n.winW), int(n.winH))
	tt.Advance(time.Second)
	tt.Frame()
	if err := tt.Click("打开播放队列"); err != nil {
		t.Fatal(err)
	}
	if !n.queueOpen {
		t.Fatal("did not open the queue")
	}
	tt.Advance(time.Second)
	tt.Frame()
	if err := tt.Click("下一首"); err != nil {
		t.Fatal(err)
	}
	if got := a.pl.Current(); got == nil || got.ID != 7 {
		t.Fatalf("single click changed to %#v", got)
	}
	if err := tt.Click("下一首"); err != nil {
		t.Fatal(err)
	}
	if got := a.pl.Current(); got == nil || got.ID != 8 {
		t.Fatalf("double click changed to %#v, want track 8", got)
	}
}

// Dragging the compact notch snaps it to the display under the pointer.
func TestNotchDragSnapsToDisplay(t *testing.T) {
	displays := notchDisplays
	defer func() { notchDisplays = displays }()
	notchDisplays = func() []mygo.Display {
		return []mygo.Display{
			{ID: 1, Bounds: mygo.Rectangle{X: 0, Y: 0, Width: 1000, Height: 800}},
			{ID: 2, Bounds: mygo.Rectangle{X: 1000, Y: 20, Width: 1200, Height: 900}},
		}
	}
	a := newTestApp()
	a.settings = defaultSettings()
	a.settings.Notch = true
	n := &notch{a: a, real: true, w0: 200, h0: 32, winW: notchOpenW + 60, winH: notchQueueH + 64, displayID: 1}
	n.place()

	n.dragStartedAt(mygo.Point{X: int(n.screenX + n.winW/2), Y: int(n.screenY + n.h0/2)})
	if !n.dragging {
		t.Fatal("did not start dragging the compact notch")
	}
	n.dragEndedAt(mygo.Point{X: 1600, Y: 100})
	if n.dragging || n.displayID != 2 {
		t.Fatalf("dragging %v, display %d, want display 2", n.dragging, n.displayID)
	}
	if want := float32(1000 + (1200-int(n.winW))/2); n.screenX != want || n.screenY != 20 {
		t.Fatalf("placed at %v,%v, want %v,20", n.screenX, n.screenY, want)
	}
}

// A drag outside the compact housing or every display leaves the notch where it was.
func TestNotchDragKeepsDisplayWithoutTarget(t *testing.T) {
	displays := notchDisplays
	defer func() { notchDisplays = displays }()
	notchDisplays = func() []mygo.Display {
		return []mygo.Display{{ID: 1, Bounds: mygo.Rectangle{X: 0, Y: 0, Width: 1000, Height: 800}}}
	}
	a := newTestApp()
	a.settings = defaultSettings()
	a.settings.Notch = true
	n := &notch{a: a, real: true, w0: 200, h0: 32, winW: notchOpenW + 60, winH: notchQueueH + 64, displayID: 1}
	n.place()
	x, y := n.screenX, n.screenY

	n.dragStartedAt(mygo.Point{X: int(x + n.winW/2), Y: int(y + n.h0 + 1)})
	if n.dragging {
		t.Fatal("started dragging below the compact housing")
	}
	n.dragStartedAt(mygo.Point{X: int(x + n.winW/2), Y: int(y + n.h0/2)})
	n.dragEndedAt(mygo.Point{X: 1200, Y: 100})
	if n.displayID != 1 || n.screenX != x || n.screenY != y {
		t.Fatalf("moved to display %d at %v,%v", n.displayID, n.screenX, n.screenY)
	}
}

// With NM_FRAMES set, saves the notch as it greets, growing and grown, and
// the player bar holding a song a login picked up.
func TestFirstUseFrames(t *testing.T) {
	dir := os.Getenv("NM_FRAMES")
	if dir == "" {
		t.Skip()
	}
	save := func(name string, tt *ui.Tester) {
		f, _ := os.Create(filepath.Join(dir, name+".png"))
		png.Encode(f, tt.Image())
		f.Close()
	}
	a := newTestApp()
	a.settings = defaultSettings()
	n := &notch{a: a, real: true, w0: 185, h0: 32, winW: notchOpenW + 60, winH: 240}
	a.notch = n
	a.pl.Cue([]netease.Track{{ID: 7, Name: "NAINAINAI", Artists: []netease.Artist{{Name: "ATARASHII GAKKO!"}}, Duration: 193877}})
	main := ui.NewTester(a.view, 1200, 800) // for the app's colors
	main.Frame()
	tt := ui.NewTester(n.view, int(n.winW), int(n.winH))
	tt.Frame()
	n.greet()
	for _, ms := range []int{120, 600} {
		tt.Advance(time.Duration(ms) * time.Millisecond)
		tt.Frame()
		save(fmt.Sprintf("notch-hello-%d", ms), tt)
	}

	a.picked = pickedUp{id: 7, heard: "3分钟前在手机上听过"}
	main.Advance(time.Second)
	main.Frame()
	save("pickup-bar", main)
}

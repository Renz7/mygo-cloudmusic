package app

import (
	_ "embed"
	"log"
	"runtime"
	"time"

	"github.com/egoist/mygo"

	"netease-music-mygo/internal/lyrics"
)

// The app's icon in the menu bar on macOS, as the settings choose
// (appIcons), and always in the notification area on Windows and Linux.
// While it shows, closing the window leaves the music playing there, until
// the listener quits; the icon opens the window again.

// trayTemplate is the record and tonearm of the app's icon, black on
// transparent for macOS to tint: 15.6 points tall in the menu bar's 18, at
// 2×, as icon/tray draws it.
//
//go:embed tray/tray-template.png
var trayTemplate []byte

// trayColor is the app's icon, for the notification area.
//
//go:embed tray/tray.png
var trayColor []byte

// appIcons are where the app's icon shows on macOS, with their names: the
// Dock and the menu bar, the Dock alone, or the menu bar alone.
var appIcons = []struct{ id, label string }{
	{"dock-and-menubar", "程序坞与菜单栏"},
	{"dock", "仅程序坞"},
	{"menubar", "仅菜单栏"},
}

// showsTray tells whether the app's icon shows in the menu bar or the
// notification area.
func showsTray(s settings) bool {
	return runtime.GOOS != "darwin" || s.AppIcon != "dock"
}

// activationPolicy is macOS's for the settings: no Dock icon for the menu
// bar alone, unless a window is open, so the listener can switch to it
// with Command-Tab and the app's menu shows.
func activationPolicy(s settings, windowOpen bool) mygo.ActivationPolicy {
	if s.AppIcon == "menubar" && !windowOpen {
		return mygo.ActivationPolicyAccessory
	}
	return mygo.ActivationPolicyRegular
}

// windowOpen tells whether the main window or the settings window is
// open: shown or minimized, as the notch player doesn't count.
func (a *app) windowOpen() bool {
	if a.settingsWin != nil {
		return true
	}
	return a.win != nil && (a.win.IsVisible() || a.win.IsMinimized())
}

// syncDock shows or hides the Dock icon, as the settings and the open
// windows say.
func (a *app) syncDock() {
	if runtime.GOOS == "darwin" {
		mygo.App.SetActivationPolicy(activationPolicy(a.settings, a.windowOpen()))
	}
}

// changeAppIcon shows the app's icon where the settings now say, as the
// listener chose another place for it.
func (a *app) changeAppIcon() {
	a.syncDock()
	a.applyTray()
}

// applyTray adds or removes the icon in the menu bar or notification area,
// as the settings say.
func (a *app) applyTray() {
	switch show := showsTray(a.settings); {
	case show && a.tray == nil:
		icon, template := trayColor, false
		if runtime.GOOS == "darwin" {
			icon, template = trayTemplate, true
		}
		t, err := mygo.NewTray(mygo.TrayOptions{ID: "cloudmusic", Icon: icon, IconIsTemplate: template, ToolTip: mygo.App.Name(), Menu: a.trayMenu()})
		if err != nil {
			log.Println("tray:", err)
			return
		}
		// Windows' left click, where the right one opens the menu.
		t.OnClick(a.showMain)
		a.tray = t
	case !show && a.tray != nil:
		a.tray.Destroy()
		a.tray = nil
	}
	a.syncMenuBarLyrics()
}

const menuBarLyricRunes = 24

// syncMenuBarLyrics keeps the current lyric in its own status item, created
// after the app's one so it sits to that icon's left.
func (a *app) syncMenuBarLyrics() {
	a.lyricsGen++
	gen := a.lyricsGen
	if runtime.GOOS != "darwin" || !a.settings.MenuBarLyrics || a.tray == nil {
		a.menuLyrics.reset()
		if a.lyricsTray != nil {
			a.lyricsTray.Destroy()
			a.lyricsTray = nil
		}
		return
	}

	text, lyricAfter := a.menuBarLyric()
	title, marqueeAfter := a.menuLyrics.menuBarLyricTitle(time.Now(), text, a.pl.Playing())
	after := soonestAfter(lyricAfter, marqueeAfter)
	if a.lyricsTray == nil {
		t, err := mygo.NewTray(mygo.TrayOptions{
			ID:      "cloudmusic-lyrics",
			Title:   title,
			ToolTip: text,
			Menu:    a.trayMenu(),
		})
		if err != nil {
			log.Println("menu bar lyrics:", err)
			return
		}
		a.lyricsTray = t
	} else {
		a.lyricsTray.SetTitle(title)
		a.lyricsTray.SetToolTip(text)
	}
	if after > 0 {
		time.AfterFunc(after, func() {
			mygo.RunOnMain(func() {
				if a.lyricsGen == gen {
					a.syncMenuBarLyrics()
				}
			})
		})
	}
}

// menuBarLyric returns the status item's full text and when its next line
// begins, if the player is moving.
func (a *app) menuBarLyric() (string, time.Duration) {
	pl := a.pl
	if pl.Current() == nil {
		return "未在播放", 0
	}
	lines, loaded := pl.Lyrics()
	if !loaded {
		return "正在加载歌词…", 0
	}
	if len(lines) == 0 {
		return "暂无歌词", 0
	}
	pos := pl.Position()
	cur := lyrics.Index(lines, pos)
	var after time.Duration
	if pl.Playing() && cur+1 < len(lines) {
		after = lines[cur+1].At - pos + 20*time.Millisecond
	}
	if cur < 0 || lines[cur].Text == "" {
		return "暂无歌词", after
	}
	return lines[cur].Text, after
}

// trayMenu is the icon's menu: the window, the player's controls,
// settings, updates, and quitting.
func (a *app) trayMenu() *mygo.Menu {
	run := func(fn func()) func(*mygo.MenuItem, *mygo.Window) {
		return func(*mygo.MenuItem, *mygo.Window) {
			fn()
			a.redraw()
		}
	}
	// In front, as from the menu bar alone, where the app may not be.
	front := func(fn func()) func() {
		return func() {
			mygo.App.Focus()
			fn()
		}
	}
	name := mygo.App.Name()
	return mygo.NewMenu([]*mygo.MenuItem{
		{Label: "打开" + name, Click: run(a.showMain)},
		mygo.Separator(),
		{Label: "播放/暂停", Click: run(func() { a.pl.TogglePlay() })},
		{Label: "下一首", Click: run(func() { a.pl.Next() })},
		{Label: "上一首", Click: run(func() { a.pl.Prev() })},
		{Label: "喜欢/取消喜欢", Click: run(a.likeCurrent)},
		mygo.Separator(),
		{Label: "设置…", Click: run(front(a.openSettings))},
		{Label: "检查更新…", Click: run(front(checkForUpdates))},
		mygo.Separator(),
		{Label: "退出" + name, Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	})
}

//go:build !darwin

package app

func notchSize(screen int) (w, h float64, ok bool) { return 0, 0, false }

func raiseAboveMenuBar(handle uintptr) {}

func pinToTop(handle uintptr, screen int, w, h float64) {}

func windowFrame(handle uintptr) (x, y, w, h float64) { return }

func menuBarHeight(screen int) float64 { return 0 }

func watchPointer(move, down, up func()) (stop func(), ok bool) { return func() {}, false }

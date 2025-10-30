package main

//import (
//  "fmt"
//  "log"
//  "strconv"
//
//  "fyne.io/fyne/v2"
//  "fyne.io/fyne/v2/app"
//)

func main() {
    BuildAndRun()
}

//func main() {
//	defer func() {
//		if r := recover(); r != nil {
//			log.Printf("[panic] main: %v", r)
//		}
//	}()
//
//	logFile, err := initLogger()
//	if err != nil {
//		fmt.Println("logging init error:", err)
//	} else {
//		defer logFile.Close()
//	}
//
//	a := app.New()
//	w := a.NewWindow("Sudoku")
//
//	state := NewAppState(a, w)
//	state.buildBoard()
//	ui := state.buildContent()
//	w.SetContent(ui)
//
//	// keyboard input
//	w.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
//		defer func() { if r := recover(); r != nil { log.Printf("[panic] key handler: %v", r) } }()
//		switch ev.Name {
//		case fyne.KeyH:
//		    state.onHint(); return
//		case fyne.KeyN:
//		    state.onNotesToggle(); return
//		case fyne.KeyP:
//		    state.onPause(); return
//		}
//		if state.paused || state.selected == nil || state.selected.fixed {
//			return
//		}
//		switch ev.Name {
//		case fyne.KeyDelete, fyne.KeyBackspace, fyne.Key0:
//			state.setCell(state.selected, 0, false)
//		default:
//			if k := string(ev.Name); len(k) == 1 {
//				if v, e := strconv.Atoi(k); e == nil && v >= 1 && v <= 9 {
//					state.applyNumber(v)
//					return
//				}
//			}
//		}
//		state.updatePalette()
//		state.refreshHighlights()
//		state.checkComplete()
//	})
//
//	w.Resize(fyne.NewSize(state.cellPx*size+340, state.cellPx*size+420))
//	state.newPuzzle(32)
//	w.ShowAndRun()
//}
//
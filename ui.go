package main

import (
	"image/color"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

/* ===========================
   UI data structures
   =========================== */

type UICell struct {
	row, col int
	value    int
	fixed    bool

	bg  *canvas.Rectangle // white cell background + border
	txt *canvas.Text      // big digit

	// pencil notes: 1..9 (index 0 unused). These are canvas.Text
	// placed in ABSOLUTE board coordinates and added *above* the cell widget.
	nT [10]*canvas.Text
}

type tapCell struct {
	widget.BaseWidget
	cell  *UICell
	onTap func(*UICell)
}

func newTapCell(c *UICell, onTap func(*UICell)) *tapCell {
	t := &tapCell{cell: c, onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tapCell) CreateRenderer() fyne.WidgetRenderer {
	// Only background + big value here.
	// Notes are added to the board grid as separate canvas objects (above).
	return widget.NewSimpleRenderer(container.NewMax(t.cell.bg, t.cell.txt))
}

func (t *tapCell) Tapped(*fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap(t.cell)
	}
}
func (t *tapCell) TappedSecondary(*fyne.PointEvent) {}

/* ===========================
   App State (UI)
   =========================== */

type AppState struct {
	app fyne.App
	win fyne.Window

	// core board widgets
	grid        *fyne.Container // absolute board layer: cells + note texts
	boardView   *fyne.Container // grid + thick lines
	pausedLayer *fyne.Container // overlay

	// data
	cells    [][]*UICell
	selected *UICell

	// timer
	timerLabel *widget.Label
	btnPause   *widget.Button
	seconds    int
	ticker     *time.Ticker
	tickerQuit chan struct{}

	// errors & hints
	errorsLabel *widget.Label
	errorsCount int
	btnHint     *widget.Button
	btnNotes    *widget.Button
	notesOn     bool
	hintsLeft   int

	// difficulty & palette
	diff        *widget.Select
	paletteBox  *fyne.Container
	paletteBtns []*widget.Button // [1..9]

	// leaderboard
	lbView LabelLike
	lb     Leaderboard

	// theme
	themeSelect *widget.Select

	// rendering constants
	cellPx    float32
	textPt    float32
	notePt    float32
	boardSize fyne.Size

	// colors (re-applied in setTheme)
	bgWhite    color.NRGBA
	hlRowCol   color.NRGBA
	hlSame     color.NRGBA
	hlSelected color.NRGBA
	wrongFlash color.NRGBA
	userColor  color.NRGBA
	givenColor color.NRGBA
	noteColor  color.NRGBA

	// current solution/puzzle
	solution Grid
}

// LabelLike lets us swap a label for richer views later.
type LabelLike interface {
	SetText(string)
}

/* ===========================
   Construction & top-level
   =========================== */

func NewAppState(a fyne.App, w fyne.Window) *AppState {
	s := &AppState{
		app:         a,
		win:         w,
		cellPx:      48,
		textPt:      22,
		notePt:      12,
		paletteBtns: make([]*widget.Button, 10), // 1..9 used
		lbView:      widget.NewLabel(""),
	}
	s.boardSize = fyne.NewSize(s.cellPx*size, s.cellPx*size)

	// defaults before theme selection
	s.bgWhite = color.NRGBA{255, 255, 255, 255}
	s.hlRowCol = color.NRGBA{220, 230, 245, 255}
	s.hlSame = color.NRGBA{220, 240, 228, 255}
	s.hlSelected = color.NRGBA{0, 114, 206, 80}
	s.wrongFlash = color.NRGBA{230, 90, 90, 150}
	s.userColor = color.NRGBA{25, 25, 25, 255}
	s.givenColor = color.NRGBA{20, 20, 20, 255}
	s.noteColor = color.NRGBA{50, 50, 60, 255}

	return s
}

// Build creates the whole UI tree and returns it.
func (s *AppState) Build() fyne.CanvasObject {
	s.lb = loadLB() // from leaderboard.go

	s.buildBoard()
	top := s.makeTopBar()
	palette := s.makePaletteRow()

	// surfaces
	topBar := s.bar(color.NRGBA{30, 30, 35, 255}, top)
	paletteBar := s.bar(color.NRGBA{38, 38, 45, 255}, palette)

	centered := container.NewStack(container.NewCenter(s.boardView), s.pausedLayer)
	boardPad := s.surface(color.NRGBA{245, 245, 248, 255},
		container.NewVBox(layout.NewSpacer(), container.NewCenter(centered), layout.NewSpacer()))

	side := container.NewVBox(s.lbView.(fyne.CanvasObject))

	return container.NewBorder(
		container.NewVBox(topBar, paletteBar),
		side, nil, nil,
		container.NewPadded(boardPad),
	)
}

/* ===========================
   Bars & helpers
   =========================== */

func (s *AppState) bar(bg color.NRGBA, inner *fyne.Container) *fyne.Container {
	r := canvas.NewRectangle(bg)
	r.SetMinSize(fyne.NewSize(0, 44))
	return container.NewMax(r, container.NewPadded(inner))
}
func (s *AppState) surface(bg color.NRGBA, inner *fyne.Container) *fyne.Container {
	r := canvas.NewRectangle(bg)
	return container.NewMax(r, container.NewPadded(inner))
}

func (s *AppState) makeTopBar() *fyne.Container {
	// difficulty
	s.diff = widget.NewSelect([]string{"Easy", "Medium", "Hard"}, nil)
	s.diff.SetSelected("Medium")
	s.diff.OnChanged = func(_ string) {
		s.updateLBView()
	}

	// theme select
	s.themeSelect = widget.NewSelect([]string{"Light", "Medium", "Dark"}, nil)
	s.themeSelect.SetSelected("Medium")
	s.themeSelect.OnChanged = func(v string) { s.setTheme(v) }

	// timer + pause
	s.timerLabel = widget.NewLabel("00:00")
	s.btnPause = widget.NewButton("Pause", func() { s.onPause() })
	s.btnPause.Importance = widget.HighImportance

	// errors
	s.errorsLabel = widget.NewLabel("Errors: 0")

	// hint & notes
	s.hintsLeft = 3
	s.btnHint = widget.NewButton("Hint (3)", func() { s.onHint() })
	s.btnHint.Importance = widget.HighImportance

	s.btnNotes = widget.NewButton("Notes: OFF", func() { s.onNotesToggle() })
	s.btnNotes.Importance = widget.MediumImportance

	// new puzzle
	btnNew := widget.NewButton("New", func() { s.onNewGame() })
	btnNew.Importance = widget.HighImportance

	return container.NewHBox(
		widget.NewLabel("Difficulty:"), s.diff,
		widget.NewSeparator(),
		widget.NewLabel("Theme:"), s.themeSelect,
		layout.NewSpacer(),
		widget.NewLabel("Time:"), s.timerLabel, s.btnPause,
		widget.NewSeparator(),
		s.errorsLabel,
		layout.NewSpacer(),
		s.btnHint,
		widget.NewSeparator(),
		s.btnNotes,
		widget.NewSeparator(),
		btnNew,
	)
}

func (s *AppState) makePaletteRow() *fyne.Container {
	s.paletteBox = container.NewHBox()
	for v := 1; v <= 9; v++ {
		num := v
		btn := widget.NewButton(strconv.Itoa(v), func() { s.applyNumber(num) })
		btn.Importance = widget.MediumImportance
		s.paletteBtns[v] = btn
		s.paletteBox.Add(btn)
	}
	s.updatePalette()
	return s.paletteBox
}

/* ===========================
   Board construction
   =========================== */

func (s *AppState) buildBoard() {
	s.grid = container.NewWithoutLayout()
	s.grid.Resize(s.boardSize)
	s.cells = make([][]*UICell, size)

	// cells + notes
	for r := 0; r < size; r++ {
		row := make([]*UICell, size)
		for c := 0; c < size; c++ {
			bg := canvas.NewRectangle(s.bgWhite)
			bg.StrokeColor = color.NRGBA{0, 0, 0, 255}
			bg.StrokeWidth = 1

			txt := canvas.NewText("", s.userColor)
			txt.TextSize = s.textPt
			txt.Alignment = fyne.TextAlignCenter

			cl := &UICell{row: r, col: c, bg: bg, txt: txt}

			// absolute placement
			x0 := float32(c) * s.cellPx
			y0 := float32(r) * s.cellPx
			wc := s.cellPx
			hc := s.cellPx

			tc := newTapCell(cl, func(sel *UICell) { s.selected = sel; s.refreshHighlights() })
			tc.Resize(fyne.NewSize(wc, hc))
			tc.Move(fyne.NewPos(x0, y0))

			bg.Resize(fyne.NewSize(wc, hc))
			bg.Move(fyne.NewPos(x0, y0))
			txt.Resize(fyne.NewSize(wc, hc))
			txt.Move(fyne.NewPos(x0, y0))

			s.grid.Add(tc) // add cell first (under notes)

			// notes (absolute board coordinates) — clockwise + center
			nw := wc / 3
			nh := hc / 3
			cx := x0 + wc/2 - nw/2
			cy := y0 + hc/2 - nh/2

			makeNote := func(i int, X, Y float32) {
				t := canvas.NewText("", s.noteColor)
				t.TextSize = s.notePt
				t.Alignment = fyne.TextAlignCenter
				t.TextStyle = fyne.TextStyle{Monospace: true}
				t.Resize(fyne.NewSize(nw, nh))
				t.Move(fyne.NewPos(X, Y))
				cl.nT[i] = t
				s.grid.Add(t) // above cell
			}
			makeNote(1, x0, y0)
			makeNote(2, x0+wc/2-nw/2, y0)
			makeNote(3, x0+wc-nw, y0)
			makeNote(4, x0, y0+hc/2-nh/2)
			makeNote(5, cx, cy)
			makeNote(6, x0+wc-nw, y0+hc/2-nh/2)
			makeNote(7, x0, y0+hc-nh)
			makeNote(8, x0+wc/2-nw/2, y0+hc-nh)
			makeNote(9, x0+wc-nw, y0+hc-nh)

			row[c] = cl
		}
		s.cells[r] = row
	}

	// thick 3×3 lines
	lines := container.NewWithoutLayout()
	makeV := func(x float32, thick bool) {
		ln := canvas.NewLine(color.NRGBA{0, 0, 0, 255})
		if thick {
			ln.StrokeWidth = 4
		} else {
			ln.StrokeWidth = 2
		}
		ln.Position1 = fyne.NewPos(x, 0)
		ln.Position2 = fyne.NewPos(x, s.boardSize.Height)
		lines.Add(ln)
	}
	makeH := func(y float32, thick bool) {
		ln := canvas.NewLine(color.NRGBA{0, 0, 0, 255})
		if thick {
			ln.StrokeWidth = 4
		} else {
			ln.StrokeWidth = 2
		}
		ln.Position1 = fyne.NewPos(0, y)
		ln.Position2 = fyne.NewPos(s.boardSize.Width, y)
		lines.Add(ln)
	}
	makeV(0, true)
	makeV(s.boardSize.Width, true)
	makeH(0, true)
	makeH(s.boardSize.Height, true)
	for i := 1; i <= 2; i++ {
		x := float32(3*i) * s.cellPx
		y := float32(3*i) * s.cellPx
		makeV(x, true)
		makeH(y, true)
	}

	spacer := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	spacer.SetMinSize(s.boardSize)
	s.boardView = container.NewMax(spacer, s.grid, lines)
	s.boardView.Resize(s.boardSize)

	// paused overlay
	overlay := canvas.NewText("PAUSED", color.NRGBA{50, 50, 50, 255})
	overlay.TextSize = 28
	overlay.Alignment = fyne.TextAlignCenter
	s.pausedLayer = container.NewCenter(overlay)
	s.pausedLayer.Hide()
}

/* ===========================
   Theme switching
   =========================== */

func (s *AppState) setTheme(name string) {
	switch name {
	case "Dark":
		s.app.Settings().SetTheme(theme.DarkTheme())
		// dark-ish UI; keep cells white for note legibility
		s.bgWhite = color.NRGBA{255, 255, 255, 255}
		s.hlRowCol = color.NRGBA{45, 55, 70, 255}
		s.hlSame = color.NRGBA{48, 70, 55, 255}
		s.hlSelected = color.NRGBA{0, 114, 206, 90}
		s.wrongFlash = color.NRGBA{180, 60, 60, 160}
		s.userColor = color.NRGBA{30, 30, 35, 255}
		s.givenColor = color.NRGBA{10, 10, 10, 255}
		s.noteColor = color.NRGBA{40, 40, 45, 255}
	case "Light":
		s.app.Settings().SetTheme(theme.LightTheme())
		s.bgWhite = color.NRGBA{255, 255, 255, 255}
		s.hlRowCol = color.NRGBA{230, 242, 255, 255}
		s.hlSame = color.NRGBA{225, 245, 230, 255}
		s.hlSelected = color.NRGBA{0, 114, 206, 70}
		s.wrongFlash = color.NRGBA{255, 120, 120, 160}
		s.userColor = color.NRGBA{0, 0, 0, 255}
		s.givenColor = color.NRGBA{30, 30, 30, 255}
		s.noteColor = color.NRGBA{50, 50, 60, 255}
	default: // Medium
		// If you added NewMediumTheme() in theme_custom.go — use it:
		if nt := NewMediumTheme(); nt != nil {
			s.app.Settings().SetTheme(nt)
		} else {
			s.app.Settings().SetTheme(theme.LightTheme())
		}
		s.bgWhite = color.NRGBA{255, 255, 255, 255}
		s.hlRowCol = color.NRGBA{220, 230, 245, 255}
		s.hlSame = color.NRGBA{220, 240, 228, 255}
		s.hlSelected = color.NRGBA{0, 114, 206, 80}
		s.wrongFlash = color.NRGBA{230, 90, 90, 150}
		s.userColor = color.NRGBA{25, 25, 25, 255}
		s.givenColor = color.NRGBA{20, 20, 20, 255}
		s.noteColor = color.NRGBA{50, 50, 60, 255}
	}

	// Re-apply colors on existing objects
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			cl := s.cells[r][c]
			if cl == nil {
				continue
			}
			cl.bg.FillColor = color.NRGBA{255, 255, 255, 255}
			cl.bg.StrokeColor = color.NRGBA{0, 0, 0, 255}
			cl.bg.Refresh()

			if cl.fixed {
				cl.txt.Color = s.givenColor
			} else {
				cl.txt.Color = s.userColor
			}
			cl.txt.Refresh()

			for i := 1; i <= 9; i++ {
				if cl.nT[i] != nil {
					cl.nT[i].Color = s.noteColor
					cl.nT[i].Refresh()
				}
			}
		}
	}
	canvas.Refresh(s.grid)
	s.refreshHighlights()
	s.updatePalette()
}

/* ===========================
   Game control hooks
   =========================== */

// main.go should call this after creating the window
func (s *AppState) Start() {
	s.setTheme("Medium") // default
	s.onNewGame()
	s.startTimer()
}

func (s *AppState) onNewGame() {
	// map difficulty to clues
	clues := 32
	switch s.diff.Selected {
	case "Easy":
		clues = 40
	case "Hard":
		clues = 25
	}
	// use optimized puzzle generator
	p, solved := MakeUniquePuzzleFast(clues)
	s.solution = solved
	s.errorsCount = 0
	s.errorsLabel.SetText("Errors: 0")
	s.hintsLeft = 3
	s.btnHint.SetText("Hint (3)")

	// populate cells
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			v := p[r][c]
			if v != 0 {
				s.setCell(s.cells[r][c], v, true)
			} else {
				s.setCell(s.cells[r][c], 0, false)
			}
			// clear notes for new game
			for i := 1; i <= 9; i++ {
				if s.cells[r][c].nT[i] != nil {
					s.cells[r][c].nT[i].Text = ""
					s.cells[r][c].nT[i].Refresh()
				}
			}
		}
	}
	s.selected = nil
	s.refreshHighlights()
	s.updatePalette()
	// reset timer
	s.stopTimer()
	s.seconds = 0
	s.timerLabel.SetText("00:00")
	s.startTimer()
	s.updateLBView()
}

// called by number palette or keyboard
func (s *AppState) applyNumber(v int) {
	if s.selected == nil || s.selected.fixed {
		return
	}
	r, c := s.selected.row, s.selected.col
	if v != s.solution[r][c] {
		s.errorsCount++
		s.errorsLabel.SetText("Errors: " + strconv.Itoa(s.errorsCount))
		s.flashWrong(s.selected)
		s.setCell(s.selected, v, false) // keep wrong entry; delete with 0/backspace
	} else {
		s.setCell(s.selected, v, false)
	}
	s.updatePalette()
	s.refreshHighlights()
	s.checkComplete()
}

func (s *AppState) onPause() {
	if s.pausedLayer.Visible() {
		s.pausedLayer.Hide()
		s.btnPause.SetText("Pause")
		s.startTimer()
	} else {
		s.pausedLayer.Show()
		s.btnPause.SetText("Resume")
		s.stopTimer()
	}
}

func (s *AppState) onNotesToggle() {
	s.notesOn = !s.notesOn
	if s.notesOn {
		s.btnNotes.SetText("Notes: ON")
	} else {
		s.btnNotes.SetText("Notes: OFF")
	}
}

func (s *AppState) onHint() {
	if s.hintsLeft <= 0 {
		dialog.ShowInformation("Hint", "No hints left.", s.win)
		return
	}
	// Build current grid
	var g Grid
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			g[r][c] = s.cells[r][c].value
		}
	}
	if rr, cc, vv, reason, ok := FindHint(g); ok {
		s.selected = s.cells[rr][cc]
		s.refreshHighlights()
		s.hintsLeft--
		s.btnHint.SetText("Hint (" + strconv.Itoa(s.hintsLeft) + ")")
		dialog.ShowInformation("Hint", reason+"\nSuggested value: "+strconv.Itoa(vv), s.win)
		return
	}
	// fallback: reveal one correct cell randomly
	type coord struct{ r, c int }
	var empties []coord
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if s.cells[r][c].value == 0 {
				empties = append(empties, coord{r, c})
			}
		}
	}
	if len(empties) > 0 {
		idx := rand.Intn(len(empties))
		r, c := empties[idx].r, empties[idx].c
		s.selected = s.cells[r][c]
		s.refreshHighlights()
		answer := s.solution[r][c]
		s.hintsLeft--
		s.btnHint.SetText("Hint (" + strconv.Itoa(s.hintsLeft) + ")")
		dialog.ShowInformation("Hint",
			"Consider row "+strconv.Itoa(r+1)+", col "+strconv.Itoa(c+1)+
				".\nThe correct number here is "+strconv.Itoa(answer)+".", s.win)
		return
	}
	dialog.ShowInformation("Hint", "Board already complete.", s.win)
}

/* ===========================
   Render/utility
   =========================== */

func (s *AppState) setCell(cl *UICell, v int, given bool) {
	cl.value = v
	cl.fixed = given
	if v == 0 {
		cl.txt.Text = ""
	} else {
		cl.txt.Text = strconv.Itoa(v)
	}
	if given {
		cl.txt.Color = s.givenColor
		cl.txt.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		cl.txt.Color = s.userColor
		cl.txt.TextStyle = fyne.TextStyle{Bold: false}
	}
	cl.txt.Refresh()
}

func (s *AppState) refreshHighlights() {
	for _, row := range s.cells {
		for _, cl := range row {
			cl.bg.FillColor = s.bgWhite
			cl.bg.StrokeColor = color.NRGBA{0, 0, 0, 255}
			cl.bg.StrokeWidth = 1
		}
	}
	if s.selected != nil {
		for i := 0; i < size; i++ {
			s.cells[s.selected.row][i].bg.FillColor = s.hlRowCol
			s.cells[i][s.selected.col].bg.FillColor = s.hlRowCol
		}
		for _, row := range s.cells {
			for _, cl := range row {
				if cl.value != 0 && s.selected.value != 0 && cl.value == s.selected.value {
					cl.bg.FillColor = s.hlSame
				}
			}
		}
		s.selected.bg.FillColor = s.hlSelected
		s.selected.bg.StrokeColor = color.NRGBA{0, 114, 206, 255}
		s.selected.bg.StrokeWidth = 3
	}
	canvas.Refresh(s.grid)
}

func (s *AppState) flashWrong(cl *UICell) {
	cl.bg.FillColor = s.wrongFlash
	canvas.Refresh(s.grid)
	time.AfterFunc(250*time.Millisecond, func() { s.refreshHighlights() })
}

func (s *AppState) updatePalette() {
	// remaining per digit
	counts := make([]int, 10)
	for v := 1; v <= 9; v++ {
		counts[v] = 9
	}
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			v := s.cells[r][c].value
			if v >= 1 && v <= 9 {
				counts[v]--
			}
		}
	}
	for v := 1; v <= 9; v++ {
		if s.paletteBtns[v] == nil {
			continue
		}
		btn := s.paletteBtns[v]
		btn.SetText(strconv.Itoa(v) + " (" + strconv.Itoa(counts[v]) + ")")
		btn.Enable()
		if counts[v] <= 0 {
			btn.Disable()
		}
	}
}

func (s *AppState) updateNoteTexts(cl *UICell) {
	if cl == nil {
		return
	}
	if !s.notesOn {
		for i := 1; i <= 9; i++ {
			if cl.nT[i] != nil {
				cl.nT[i].Text = ""
				cl.nT[i].Refresh()
			}
		}
		canvas.Refresh(s.grid)
		return
	}
	// (Optional) auto-candidate logic could go here. For manual toggle via number keys,
	// main.go can call this after user toggles a note.
	canvas.Refresh(s.grid)
}

func (s *AppState) checkComplete() {
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if s.cells[r][c].value != s.solution[r][c] {
				return
			}
		}
	}
	// solved
	s.stopTimer()
	addScore(s.lb, s.diff.Selected, Score{Seconds: s.seconds, Errors: s.errorsCount, FinishedAt: time.Now()})
	saveLB(s.lb)
	s.updateLBView()
	dialog.ShowInformation("Solved!", "Time: "+s.timerLabel.Text+"\nErrors: "+strconv.Itoa(s.errorsCount), s.win)
}

func (s *AppState) updateLBView() {
	arr := s.lb[s.diff.Selected]
	if len(arr) == 0 {
		s.lbView.SetText("Leaderboard (" + s.diff.Selected + "): —")
		return
	}
	out := "Leaderboard (" + s.diff.Selected + "):\n"
	for i, v := range arr {
		min := v.Seconds / 60
		sec := v.Seconds % 60
		out += strconv.Itoa(i+1) + ". " + strconv.Itoa(min) + "m " + strconv.Itoa(sec) + "s"
		if v.Errors > 0 {
			out += " (" + strconv.Itoa(v.Errors) + " errs)"
		}
		out += "\n"
	}
	s.lbView.SetText(out)
}

/* ===========================
   Timer
   =========================== */

func (s *AppState) startTimer() {
	s.stopTimer()
	s.ticker = time.NewTicker(time.Second)
	s.tickerQuit = make(chan struct{})
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.seconds++
				min := s.seconds / 60
				sec := s.seconds % 60
				s.timerLabel.SetText(
					strconv.Itoa(min/10) + strconv.Itoa(min%10) + ":" +
						strconv.Itoa(sec/10) + strconv.Itoa(sec%10))
			case <-s.tickerQuit:
				return
			}
		}
	}()
}

func (s *AppState) stopTimer() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.tickerQuit != nil {
		select {
		case <-s.tickerQuit:
		default:
			close(s.tickerQuit)
		}
	}
}

/* ===========================
   main wiring helper
   =========================== */

func BuildAndRun() {
	a := app.New()
	w := a.NewWindow("Sudoku")

	state := NewAppState(a, w)
	content := state.Build()
	w.SetContent(content)

	// keyboard input
	w.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		switch ev.Name {
		case fyne.KeyH:
			state.onHint()
			return
		case fyne.KeyN:
			state.onNotesToggle()
			return
		case fyne.KeyP:
			state.onPause()
			return
		}
		// number entry
		if state.selected != nil && !state.selected.fixed {
			switch ev.Name {
			case fyne.KeyDelete, fyne.KeyBackspace, fyne.Key0:
				state.setCell(state.selected, 0, false)
				state.updatePalette()
				state.refreshHighlights()
				return
			default:
				k := string(ev.Name)
				if len(k) == 1 {
					if v, err := strconv.Atoi(k); err == nil && v >= 1 && v <= 9 {
						if state.notesOn {
							// toggle note text
							t := state.selected.nT[v]
							if t != nil {
								if t.Text == "" {
									t.Text = strconv.Itoa(v)
								} else {
									t.Text = ""
								}
								t.Refresh()
								canvas.Refresh(state.grid)
							}
						} else {
							state.applyNumber(v)
						}
					}
				}
			}
		}
	})

	// default theme + start
	state.setTheme("Medium")
	state.Start()

	// reasonable default size
	w.Resize(fyne.NewSize(state.cellPx*size+320, state.cellPx*size+320))
	w.ShowAndRun()
}
